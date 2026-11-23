<div align="center">

<img src="assets/logo.svg" alt="Sub2API Logo" width="128" />

# Sub2API（单管理员自用版）

[English](README.md) | 中文

</div>

基于 [Wei-Shaw/sub2api](https://github.com/Wei-Shaw/sub2api) 精简的私有中转网关。只有一个管理员账号，不提供注册、充值、订阅、兑换码、邀请返利等多用户/SaaS 功能，相关代码和数据表已删除。

典型用法：在上游账号（Anthropic、OpenAI 等 API Key）前面挂一层网关，由 NewAPI 等下游站点通过 sub2api 的 API Key 调用。

## 功能

- **上游账号管理**：多平台账号，按分组隔离；支持代理、模型映射、错误码策略。
- **调度**：粘性会话 → 优先级 → 成本优先（含首字守门）→ 负载 → LRU，见下文。
- **计费**：Token 级用量记录。账号的「账号计费倍率」表示上游成本，「本站倍率」表示向下游收费的倍率（未设置时按分组倍率）。
- **运维监控**：错误日志、延迟、首字耗时、告警规则。
- **管理后台**：Web 界面，内置在二进制中（`-tags embed`）。

## 调度：成本优先 + 首字守门

在同一优先级内，按「账号计费倍率」从低到高分档，优先使用最便宜的一档。如果某个便宜账号的首字耗时（EWMA）比最快账号同时慢出 `cost_ttft_guard_abs_ms` 毫秒和 `cost_ttft_guard_ratio` 倍，就暂时降档，避免为了省钱牺牲体验。没有首字样本的账号视为中性。

适用于 Anthropic、OpenAI（含兼容平台）和 Gemini 线路。首字统计保存在内存中，进程重启后重新学习。OpenAI 高级调度器（`openai_advanced_scheduler_enabled`）开启时不走这套逻辑。

```yaml
gateway:
  scheduling:
    prefer_lower_cost: true        # 默认 false
    cost_ttft_guard_abs_ms: 800    # 绝对阈值，<=0 关闭守门
    cost_ttft_guard_ratio: 0.5     # 相对阈值
```

环境变量形式：`GATEWAY_SCHEDULING_PREFER_LOWER_COST` 等。

## 并发

管理员用户的并发设为 `0` 表示不限制，只受各上游账号自身的并发上限约束。所有下游请求都使用同一个管理员身份，不要给它设置较小的值。

## 部署

使用 Docker Compose，包含 PostgreSQL 和 Redis，参见 [deploy/](deploy/) 目录和 `deploy/config.example.yaml`。

构建带前端的单文件二进制：

```bash
cd frontend && pnpm install && pnpm run build      # 输出到 backend/internal/web/dist
cd ../backend && CGO_ENABLED=0 go build -tags embed -trimpath -o sub2api ./cmd/server
```

数据库迁移在启动时自动执行（`backend/migrations/`）。升级前请先备份数据库。

### Nginx 反向代理

搭配 Codex CLI 使用时，需要在 Nginx `http` 块中添加：

```nginx
underscores_in_headers on;
```

Nginx 默认会丢弃名称中含下划线的请求头（如 `session_id`），导致粘性会话失效。

## Codex Fast/Flex 策略

后台 `系统设置 -> 网关服务 -> OpenAI Fast/Flex 策略` 只处理请求体中的 `service_tier`，不会修改 Codex 客户端的模型目录：

- `pass`：保留客户端的 `service_tier`，`fast` 规范为 `priority`。
- `filter`：移除 `service_tier`。
- `block`：拒绝匹配的请求。
- `force_priority`：强制设为 `priority`，按 Priority 价格计费。`all` 只匹配显式存在的 tier；要让省略 `service_tier` 的请求也升级，需要新增 `service_tier=missing + force_priority` 规则。

客户端可在 `~/.codex/config.toml` 中指定：

```toml
service_tier = "fast"

[features]
fast_mode = true
```

## 开发

```bash
cd backend
go test -tags=unit ./...          # 单元测试
go vet -tags=integration ./...    # 集成测试编译检查
cd ../frontend && pnpm test
```

## 许可证

基于 [GNU 宽通用公共许可证 v3.0](LICENSE)（或更高版本）授权。上游项目 Copyright (c) 2026 Wesley Liddick。
