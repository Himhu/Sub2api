-- 单管理员自用版：清理已删除功能遗留、且不再被任何代码读写的设置项和数据表。
--
-- 设置项：第三方登录（邮箱 / LinuxDo / OIDC / 微信）注册默认值、支付方式可见性、
-- 支付宝深链、微信公众号/开放平台开关、第三方注册强制邮箱、订阅到期提醒。
--
-- 数据表：
--   billing_usage_entries          早期计费一致性表，已被 usage_billing_dedup 取代
--   deleted_api_key_audits         已删 Key 审计，写入逻辑已移除
--   orphan_allowed_groups_audit    迁移 013 的一次性审计记录
--   groups_video_price_backup_220  迁移 220 的一次性备份
--
-- 注意：本迁移不可逆。升级已有部署前请先备份数据库。

DELETE FROM settings
WHERE key LIKE 'auth\_source\_default\_%' ESCAPE '\'
   OR key LIKE 'payment\_visible\_method\_%' ESCAPE '\'
   OR key IN (
       'ALIPAY_MOBILE_PRECREATE_DEEP_LINK',
       'wechat_connect_mp_enabled',
       'wechat_connect_open_enabled',
       'force_email_on_third_party_signup',
       'subscription_expiry_notify_enabled'
   );

DROP TABLE IF EXISTS billing_usage_entries;
DROP TABLE IF EXISTS deleted_api_key_audits;
DROP TABLE IF EXISTS orphan_allowed_groups_audit;
DROP TABLE IF EXISTS groups_video_price_backup_220;
