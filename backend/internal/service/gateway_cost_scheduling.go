package service

import (
	"math"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

// accountTTFTStats 在内存中维护每个账号的首字耗时 EWMA（指数平滑均值），
// 供成本优先调度做「首字守门」。进程重启后从零开始学习；没有样本的账号
// 视为中性（不惩罚）。
type accountTTFTStats struct {
	accounts sync.Map // accountID -> *atomic.Uint64 (float64 bits, NaN = 无样本)
}

const accountTTFTEWMAAlpha = 0.2

func newAccountTTFTStats() *accountTTFTStats {
	return &accountTTFTStats{}
}

// report 记录一次请求的首字耗时；非流式请求没有首字时间时传 nil，不更新。
func (s *accountTTFTStats) report(accountID int64, firstTokenMs *int) {
	if s == nil || accountID <= 0 || firstTokenMs == nil || *firstTokenMs <= 0 {
		return
	}
	sample := float64(*firstTokenMs)
	value, ok := s.accounts.Load(accountID)
	if !ok {
		fresh := &atomic.Uint64{}
		fresh.Store(math.Float64bits(math.NaN()))
		value, _ = s.accounts.LoadOrStore(accountID, fresh)
	}
	target, _ := value.(*atomic.Uint64)
	if target == nil {
		return
	}
	for {
		oldBits := target.Load()
		oldValue := math.Float64frombits(oldBits)
		newValue := sample
		if !math.IsNaN(oldValue) {
			newValue = accountTTFTEWMAAlpha*sample + (1-accountTTFTEWMAAlpha)*oldValue
		}
		if target.CompareAndSwap(oldBits, math.Float64bits(newValue)) {
			return
		}
	}
}

// snapshot 返回账号当前的首字 EWMA（毫秒）；没有样本时 ok=false。
func (s *accountTTFTStats) snapshot(accountID int64) (ttftMs float64, ok bool) {
	if s == nil || accountID <= 0 {
		return 0, false
	}
	value, loaded := s.accounts.Load(accountID)
	if !loaded {
		return 0, false
	}
	target, _ := value.(*atomic.Uint64)
	if target == nil {
		return 0, false
	}
	ttftMs = math.Float64frombits(target.Load())
	if math.IsNaN(ttftMs) {
		return 0, false
	}
	return ttftMs, true
}

// forget 删除账号的统计（账号被删除/停用时调用，避免 map 无限增长）。
func (s *accountTTFTStats) forget(accountID int64) {
	if s == nil {
		return
	}
	s.accounts.Delete(accountID)
}

// accountSchedulingCostRate 返回账号用于成本分档的计费倍率；未设置视为 1。
func accountSchedulingCostRate(account *Account) float64 {
	if account == nil || account.RateMultiplier == nil || *account.RateMultiplier < 0 {
		return 1
	}
	return *account.RateMultiplier
}

// ttftGuardTripped 判断一条线路的首字耗时是否比最快线路慢到需要降档：
// 必须同时超过绝对阈值和相对阈值。absMS<=0 表示守门关闭。
func ttftGuardTripped(ttftMs, fastestMs float64, cfg config.GatewaySchedulingConfig) bool {
	if cfg.CostTTFTGuardAbsMS <= 0 {
		return false
	}
	diff := ttftMs - fastestMs
	if diff <= float64(cfg.CostTTFTGuardAbsMS) {
		return false
	}
	ratio := cfg.CostTTFTGuardRatio
	if ratio < 0 {
		ratio = 0
	}
	if fastestMs <= 0 {
		return true
	}
	return diff > fastestMs*ratio
}

// filterByLowestCost 从同优先级候选中挑出「成本最低且首字耗时未被守门拦下」的档。
//
// 算法：
//  1. 按账号计费倍率从低到高分档（倍率相同的账号在同一档）。
//  2. 取本轮候选中首字 EWMA 最快的值作为基准。
//  3. 从最便宜的档开始，剔除其中被守门拦下（比基准慢超过阈值）的账号；
//     若该档剩下账号则返回它们，否则看下一档。
//  4. 所有档都被拦下（极端情况）时，返回原集合，退回后续的负载率/LRU 选择。
//
// 没有首字样本的账号视为中性，永不被拦下。
func filterByLowestCost(accounts []accountWithLoad, stats *accountTTFTStats, cfg config.GatewaySchedulingConfig) []accountWithLoad {
	if len(accounts) <= 1 {
		return accounts
	}

	type sample struct {
		acc     accountWithLoad
		rate    float64
		ttft    float64
		hasTTFT bool
	}
	samples := make([]sample, 0, len(accounts))
	fastest := math.Inf(1)
	for _, acc := range accounts {
		s := sample{acc: acc, rate: accountSchedulingCostRate(acc.account)}
		if stats != nil {
			s.ttft, s.hasTTFT = stats.snapshot(acc.account.ID)
			if s.hasTTFT && s.ttft < fastest {
				fastest = s.ttft
			}
		}
		samples = append(samples, s)
	}
	sort.SliceStable(samples, func(i, j int) bool { return samples[i].rate < samples[j].rate })

	hasBaseline := !math.IsInf(fastest, 1)
	for i := 0; i < len(samples); {
		j := i + 1
		for j < len(samples) && samples[j].rate == samples[i].rate {
			j++
		}
		tier := make([]accountWithLoad, 0, j-i)
		for _, s := range samples[i:j] {
			if hasBaseline && s.hasTTFT && ttftGuardTripped(s.ttft, fastest, cfg) {
				continue
			}
			tier = append(tier, s.acc)
		}
		if len(tier) > 0 {
			return tier
		}
		i = j
	}
	return accounts
}
