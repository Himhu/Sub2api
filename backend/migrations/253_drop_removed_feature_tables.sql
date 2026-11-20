-- 单管理员自用版：删除已移除功能遗留的数据表。
--
-- 对应功能（支付充值、订阅套餐、兑换码、优惠码、邀请返利、公告、第三方登录身份、
-- 待定登录会话、用户自定义属性）的代码与 ORM 实体均已删除，这些表不再被任何代码读写。
--
-- 注意：本迁移不可逆，会连同表内数据一并删除。升级已有部署前请先备份数据库。
-- CASCADE 仅用于移除其他保留表上指向这些表的外键约束
-- （例如 usage_logs.subscription_id -> user_subscriptions），不会删除保留表的数据或列。

DROP TABLE IF EXISTS announcement_reads CASCADE;
DROP TABLE IF EXISTS announcements CASCADE;

DROP TABLE IF EXISTS payment_audit_logs CASCADE;
DROP TABLE IF EXISTS user_affiliate_ledger CASCADE;
DROP TABLE IF EXISTS user_affiliates CASCADE;
DROP TABLE IF EXISTS payment_orders CASCADE;
DROP TABLE IF EXISTS payment_provider_instances CASCADE;
DROP TABLE IF EXISTS subscription_plans CASCADE;

DROP TABLE IF EXISTS promo_code_usages CASCADE;
DROP TABLE IF EXISTS promo_codes CASCADE;
DROP TABLE IF EXISTS redeem_codes CASCADE;

DROP TABLE IF EXISTS user_subscriptions CASCADE;

DROP TABLE IF EXISTS identity_adoption_decisions CASCADE;
DROP TABLE IF EXISTS auth_identity_channels CASCADE;
DROP TABLE IF EXISTS pending_auth_sessions CASCADE;
DROP TABLE IF EXISTS auth_identities CASCADE;
DROP TABLE IF EXISTS auth_identity_migration_reports CASCADE;
DROP TABLE IF EXISTS user_provider_default_grants CASCADE;

DROP TABLE IF EXISTS user_attribute_values CASCADE;
DROP TABLE IF EXISTS user_attribute_definitions CASCADE;
