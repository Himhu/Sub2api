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
	// 成本比较只在同一优先级内进行：高优先级的账号即使更贵，也不能被低优先级的便宜账号挤掉。
	accounts = filterByMinPriority(accounts)
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

// sharedAccountTTFTStats 由 Anthropic/Gemini 网关和 OpenAI 网关共用，
// 账号 ID 全局唯一，任一网关记录的首字耗时都能被另一侧调度看到。
var sharedAccountTTFTStats = newAccountTTFTStats()

// orderByLowestCost 把候选重排成「优先级 → 成本档（含首字守门）→ 原有顺序」，
// 供按列表顺序逐个尝试抢槽的调度路径使用（OpenAI 网关）。
// 不要求输入已按优先级排好：按优先级数值从小到大逐组处理，组内反复取出
// 当前最优成本档并保持其内部原有顺序，直到组内账号全部排完。
func orderByLowestCost(accounts []accountWithLoad, stats *accountTTFTStats, cfg config.GatewaySchedulingConfig) []accountWithLoad {
	if len(accounts) <= 1 {
		return accounts
	}
	priorities := make([]int, 0, len(accounts))
	seen := make(map[int]struct{}, len(accounts))
	for _, acc := range accounts {
		if _, ok := seen[acc.account.Priority]; ok {
			continue
		}
		seen[acc.account.Priority] = struct{}{}
		priorities = append(priorities, acc.account.Priority)
	}
	sort.Ints(priorities)

	ordered := make([]accountWithLoad, 0, len(accounts))
	for _, priority := range priorities {
		remaining := make([]accountWithLoad, 0, len(accounts))
		for _, acc := range accounts {
			if acc.account.Priority == priority {
				remaining = append(remaining, acc)
			}
		}
		for len(remaining) > 0 {
			tier := filterByLowestCost(remaining, stats, cfg)
			if len(tier) == 0 || len(tier) == len(remaining) {
				ordered = append(ordered, remaining...)
				break
			}
			picked := make(map[int64]struct{}, len(tier))
			for _, acc := range tier {
				picked[acc.account.ID] = struct{}{}
			}
			next := make([]accountWithLoad, 0, len(remaining)-len(tier))
			for _, acc := range remaining {
				if _, ok := picked[acc.account.ID]; ok {
					ordered = append(ordered, acc)
				} else {
					next = append(next, acc)
				}
			}
			remaining = next
		}
	}
	return ordered
}

// orderAccountsByLowestCost 是 orderByLowestCost 的 []*Account 版本，
// 用于没有负载信息的兜底路径（负载查询失败 / 排队等待）。
func orderAccountsByLowestCost(accounts []*Account, stats *accountTTFTStats, cfg config.GatewaySchedulingConfig) []*Account {
	if len(accounts) <= 1 {
		return accounts
	}
	wrapped := make([]accountWithLoad, 0, len(accounts))
	for _, acc := range accounts {
		if acc == nil {
			continue
		}
		wrapped = append(wrapped, accountWithLoad{account: acc, loadInfo: &AccountLoadInfo{AccountID: acc.ID}})
	}
	wrapped = orderByLowestCost(wrapped, stats, cfg)
	out := make([]*Account, 0, len(wrapped))
	for _, acc := range wrapped {
		out = append(out, acc.account)
	}
	return out
}
