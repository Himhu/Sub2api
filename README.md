<div align="center">

<img src="assets/logo.svg" alt="Sub2API Logo" width="128" />

# Sub2API (single-admin self-hosted edition)

English | [中文](README_CN.md)

</div>

A private relay gateway trimmed down from [Wei-Shaw/sub2api](https://github.com/Wei-Shaw/sub2api). It has exactly one admin account. Registration, top-ups, subscriptions, redeem codes, referrals and other multi-user/SaaS features have been removed, including their code and database tables.

Typical use: put a gateway in front of upstream accounts (Anthropic, OpenAI, etc. API keys) and let downstream sites such as NewAPI call it with a sub2api API key.

## Features

- **Upstream accounts**: multiple platforms, isolated by group; proxies, model mapping, error-code policies.
- **Scheduling**: sticky session → priority → lowest cost (with TTFT guard) → load → LRU. See below.
- **Billing**: token-level usage records. An account's "billing rate multiplier" is the upstream cost; the "site rate multiplier" is what downstream is charged (falls back to the group rate when unset).
- **Ops monitoring**: error logs, latency, time to first token, alert rules.
- **Admin UI**: embedded in the binary (`-tags embed`).

## Scheduling: lowest cost + TTFT guard

Within the same priority, accounts are tiered by billing rate multiplier, cheapest first. A cheaper account is demoted when its time-to-first-token EWMA is slower than the fastest account by more than both `cost_ttft_guard_abs_ms` and `cost_ttft_guard_ratio`. Accounts without samples are neutral.

Applies to the Anthropic, OpenAI (and compatible platforms) and Gemini lines. TTFT stats live in memory and reset on restart. The OpenAI advanced scheduler (`openai_advanced_scheduler_enabled`) does not use this logic.

```yaml
gateway:
  scheduling:
    prefer_lower_cost: true        # default false
    cost_ttft_guard_abs_ms: 800    # absolute threshold, <=0 disables the guard
    cost_ttft_guard_ratio: 0.5     # relative threshold
```

Environment variables: `GATEWAY_SCHEDULING_PREFER_LOWER_COST`, etc.

## Concurrency

Set the admin user's concurrency to `0` for no limit; only each upstream account's own limit applies. Every downstream request uses the single admin identity, so do not give it a small value.

## Deployment

Use Docker Compose with PostgreSQL and Redis; see [deploy/](deploy/) and `deploy/config.example.yaml`.

Build a single binary with the embedded frontend:

```bash
cd frontend && pnpm install && pnpm run build      # outputs to backend/internal/web/dist
cd ../backend && CGO_ENABLED=0 go build -tags embed -trimpath -o sub2api ./cmd/server
```

Database migrations run automatically on startup (`backend/migrations/`). Back up the database before upgrading.

### Nginx reverse proxy

When used with Codex CLI, add this to the Nginx `http` block:

```nginx
underscores_in_headers on;
```

Nginx drops headers containing underscores (such as `session_id`) by default, which breaks sticky sessions.

## Codex Fast/Flex policy

`Settings -> Gateway -> OpenAI Fast/Flex policy` only rewrites `service_tier` in the request body; it does not change the Codex client's model catalog:

- `pass`: keep the client's `service_tier`; `fast` is normalized to `priority`.
- `filter`: remove `service_tier`.
- `block`: reject matching requests.
- `force_priority`: force `priority`, billed at Priority prices. `all` only matches an explicit tier; to upgrade requests that omit `service_tier`, add a `service_tier=missing + force_priority` rule.

Client side, in `~/.codex/config.toml`:

```toml
service_tier = "fast"

[features]
fast_mode = true
```

## Development

```bash
cd backend
go test -tags=unit ./...          # unit tests
go vet -tags=integration ./...    # compile-check integration tests
cd ../frontend && pnpm test
```

## License

Licensed under the [GNU Lesser General Public License v3.0](LICENSE) (or later). Upstream project Copyright (c) 2026 Wesley Liddick.
