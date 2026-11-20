package adminaccess

import (
	"slices"
	"strings"
)

type Permission struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Category string `json:"category"`
	Scope    string `json:"scope"`
}

type permissionDefinition struct {
	Permission
	Routes []string
}

// 接口按方法和 Gin 完整路径精确登记；新增接口不会继承同目录权限。
// 全站权限与账户规则独立，选中分组绝不会隐式授予全站操作。
var definitions = []permissionDefinition{
	{Permission{"accounts.identity.read", "读取账户邮箱、套餐及创建时间（不含令牌）", "账户", "account"}, []string{"GET /accounts/:id"}},
	{Permission{"accounts.usage.read", "读取及主动刷新授权账户额度与快照", "账户", "account"}, []string{"GET /accounts/:id/usage", "POST /accounts/usage/batch"}},
	{Permission{"accounts.stats.read", "读取账户用量与费用统计", "账户", "account"}, []string{"GET /accounts/:id/today-stats", "POST /accounts/today-stats/batch", "GET /accounts/:id/stats"}},
	{Permission{"openai.quota.read", "查询 OpenAI 上游额度与重置次数", "OpenAI", "account"}, []string{"GET /openai/accounts/:id/quota", "POST /openai/accounts/:id/quota/refresh"}},
	{Permission{"openai.quota.reset", "消耗 OpenAI 重置次数恢复额度", "OpenAI", "account"}, []string{"POST /openai/accounts/:id/reset-quota"}},
	{Permission{"openai.referrals.read", "查询 OpenAI 邀请资格与次数", "OpenAI", "account"}, []string{"POST /openai/accounts/:id/referrals/refresh"}},
	{Permission{"openai.referrals.invite", "发送 OpenAI 邀请邮件", "OpenAI", "account"}, []string{"POST /openai/accounts/:id/referrals/invite"}},
	{Permission{"accounts.options.write", "写入 OpenAI OAuth 身份与连接选项", "账户", "account"}, []string{"POST /accounts"}},
	{Permission{"accounts.rate.write", "设置账户计费倍率", "账户", "account"}, []string{"POST /accounts", "PUT /accounts/:id"}},
	{Permission{"groups.list.read", "读取所选分组基本信息（名称、平台、状态）", "分组", "group"}, []string{"GET /groups", "GET /groups/all", "GET /groups/:id"}},
	{Permission{"groups.capacity.read", "读取所选分组容量（整个分组，不按账户类型过滤）", "分组", "group"}, []string{"GET /groups/capacity-summary"}},
	{Permission{"openai.tokens.refresh", "刷新调用方提供的 OpenAI 令牌（返回令牌，可使用所选代理）", "OAuth", "global"}, []string{"POST /openai/refresh-token"}},
	{Permission{"accounts.read", "查看账户（脱敏）", "账户", "account"}, []string{"GET /accounts", "GET /accounts/:id", "GET /accounts/:id/models"}},
	{Permission{"accounts.create", "创建账户", "账户", "account"}, []string{"POST /accounts", "POST /accounts/import/codex-session"}},
	{Permission{"accounts.update", "编辑名称与备注", "账户", "account"}, []string{"PUT /accounts/:id"}},
	{Permission{"accounts.credentials", "写入凭证（不包含读取）", "账户", "account"}, []string{"PUT /accounts/:id"}},
	{Permission{"accounts.models.write", "编辑账户白名单模型 ID 与映射", "账户", "account"}, []string{"PUT /accounts/:id", "POST /accounts/bulk-update"}},
	{Permission{"accounts.network", "修改接入地址、代理与连接选项", "账户", "account"}, []string{"PUT /accounts/:id"}},
	{Permission{"accounts.status", "启用、停用与调度开关", "账户", "account"}, []string{"PUT /accounts/:id", "POST /accounts/:id/schedulable"}},
	{Permission{"accounts.schedule", "修改并发、优先级与有效期", "账户", "account"}, []string{"PUT /accounts/:id"}},
	{Permission{"accounts.refresh", "刷新 OAuth 凭证", "账户", "account"}, []string{"POST /accounts/:id/refresh"}},
	{Permission{"accounts.recover", "清除账户错误与限流状态", "账户", "account"}, []string{"POST /accounts/:id/clear-error"}},
	{Permission{"accounts.delete", "删除账户（影响共享分组）", "账户", "account"}, []string{"DELETE /accounts/:id"}},
	{Permission{"accounts.groups", "调整授权分组关联", "账户", "account"}, []string{"PUT /accounts/:id"}},
	{Permission{"groups.models.read", "读取所选分组模型白名单与候选 ID", "分组模型", "group"}, []string{"GET /groups", "GET /groups/all", "GET /groups/:id", "GET /groups/:id/model-allowlist-candidates"}},
	{Permission{"groups.models.write", "编辑所选分组模型白名单（影响整个分组）", "分组模型", "group"}, []string{"PUT /groups/:id"}},
	{Permission{"proxies.read", "读取代理 IP 候选（分页，不含认证凭证）", "代理", "global"}, []string{"GET /proxies", "GET /proxies/all", "GET /proxies/:id"}},
	{Permission{"models.defaults.read", "读取默认白名单模型 ID（含默认映射）", "模型", "global"}, []string{"GET /accounts/antigravity/default-model-mapping", "GET /groups/:id/model-allowlist-candidates"}},
	{Permission{"dashboard.read", "读取全站经营报表", "报表", "global"}, []string{"GET /dashboard/snapshot-v2", "GET /dashboard/stats", "GET /dashboard/realtime", "GET /dashboard/trend", "GET /dashboard/models", "GET /dashboard/groups", "GET /dashboard/api-keys-trend", "GET /dashboard/users-trend", "GET /dashboard/users-ranking", "POST /dashboard/users-usage", "POST /dashboard/api-keys-usage", "GET /dashboard/user-breakdown"}},
	{Permission{"groups.read", "读取所有分组配置", "分组", "global"}, []string{"GET /groups", "GET /groups/all", "GET /groups/:id", "GET /groups/:id/rate-multipliers", "GET /groups/:id/composite-routes", "GET /groups/:id/model-allowlist-candidates"}},
	{Permission{"usage.read", "读取全站用量明细及用户用量", "报表", "global"}, []string{"GET /usage", "GET /usage/stats"}},
	{Permission{"ops.read", "读取全站运维指标", "监控", "global"}, []string{"GET /ops/concurrency", "GET /ops/user-concurrency", "GET /ops/account-availability", "GET /ops/realtime-traffic", "GET /ops/ingress-rejections", "GET /ops/ingress-rejections/health", "GET /ops/auth-cache-invalidation/health", "GET /ops/dashboard/snapshot-v2", "GET /ops/dashboard/overview", "GET /ops/dashboard/throughput-trend", "GET /ops/dashboard/latency-histogram", "GET /ops/dashboard/error-trend", "GET /ops/dashboard/error-distribution", "GET /ops/dashboard/openai-token-stats"}},
	{Permission{"ops.errors.read", "读取全站请求错误详情（敏感）", "监控", "global"}, []string{"GET /ops/errors", "GET /ops/errors/:id", "GET /ops/request-errors", "GET /ops/request-errors/:id", "GET /ops/request-errors/:id/upstream-errors", "GET /ops/upstream-errors", "GET /ops/upstream-errors/:id", "GET /ops/requests"}},
	{Permission{"ops.errors.resolve", "标记请求错误处理状态", "监控", "global"}, []string{"PUT /ops/errors/:id/resolve", "PUT /ops/request-errors/:id/resolve", "PUT /ops/upstream-errors/:id/resolve"}},
	{Permission{"ops.alerts.read", "读取全站告警规则与事件", "监控", "global"}, []string{"GET /ops/alert-rules", "GET /ops/alert-events", "GET /ops/alert-events/:id"}},
	{Permission{"ops.alerts.write", "修改告警规则、状态与静默", "监控", "global"}, []string{"POST /ops/alert-rules", "PUT /ops/alert-rules/:id", "DELETE /ops/alert-rules/:id", "PUT /ops/alert-events/:id/status", "POST /ops/alert-silences"}},
}

func PermissionCatalog() []Permission {
	result := make([]Permission, 0, len(definitions))
	for _, definition := range definitions {
		result = append(result, definition.Permission)
	}
	return result
}

// RouteAllowed 只检查动作准入；账户资源及字段必须再由专用处理器检查。
func RouteAllowed(policy Policy, method, path string) bool {
	if policy.FullAccess {
		return strings.HasPrefix(path, "/api/v1/admin/") && !strings.HasPrefix(path, "/api/v1/admin/settings/admin-api-key")
	}
	target := method + " " + path
	for _, definition := range definitions {
		matched := false
		for _, route := range definition.Routes {
			// route 的方法前缀与资源路径固定，不做前缀通配或按 HTTP 方法推断读写。
			for i, ch := range route {
				if ch == ' ' {
					matched = target == route[:i]+" /api/v1/admin"+route[i+1:]
					break
				}
			}
			if matched {
				break
			}
		}
		if !matched {
			continue
		}
		if definition.Scope == "global" && slices.Contains(policy.GlobalPermissions, definition.ID) {
			return true
		}
		if (definition.Scope == "account" || definition.Scope == "group") && len(policy.RulesFor(definition.ID)) > 0 {
			return true
		}
	}
	return false
}
