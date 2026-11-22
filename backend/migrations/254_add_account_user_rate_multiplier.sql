-- 线路（账号）级本站倍率：填写后该线路上的请求按此倍率向用户扣费；
-- NULL 表示沿用分组倍率（既有行为）。与 rate_multiplier（上游倍率 / 上游成本口径）相互独立。
ALTER TABLE accounts
    ADD COLUMN IF NOT EXISTS user_rate_multiplier DECIMAL(10,4) NULL;

COMMENT ON COLUMN accounts.user_rate_multiplier IS
    'Per-account downstream (site) rate multiplier used to charge users; NULL falls back to the group rate multiplier';
