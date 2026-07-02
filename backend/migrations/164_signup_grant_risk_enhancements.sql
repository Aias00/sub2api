-- Extend signup grant risk controls with OAuth identity, device, and domain-quality settings.

CREATE INDEX IF NOT EXISTS idx_signup_grant_claims_provider_subject_allowed
    ON signup_grant_claims (provider_type, provider_subject_hash, created_at DESC)
    WHERE decision = 'allowed' AND provider_subject_hash <> '';

CREATE INDEX IF NOT EXISTS idx_signup_grant_claims_device_allowed
    ON signup_grant_claims (user_agent_hash, created_at DESC)
    WHERE decision = 'allowed' AND user_agent_hash <> '';

INSERT INTO settings (key, value, description, created_at, updated_at) VALUES
    ('signup_grant_risk_control_oauth_identity_enabled', 'true', '是否启用 OAuth identity 唯一领取注册赠送', NOW(), NOW()),
    ('signup_grant_risk_control_device_enabled', 'true', '是否启用同设备/浏览器 24 小时领取注册赠送限制', NOW(), NOW()),
    ('signup_grant_risk_control_device_daily_limit', '2', '同设备/浏览器 24 小时可领取注册赠送次数，0 表示不限', NOW(), NOW()),
    ('signup_grant_risk_control_free_domain_daily_limit', '5', '免费邮箱域名 24 小时可领取注册赠送次数，0 表示使用普通域名上限', NOW(), NOW()),
    ('signup_grant_risk_control_blocked_email_domains', '', '禁止领取注册赠送的邮箱域名列表，逗号或换行分隔', NOW(), NOW()),
    ('signup_grant_risk_control_free_email_domains', 'gmail.com,googlemail.com,outlook.com,hotmail.com,live.com,icloud.com,yahoo.com,qq.com,163.com,126.com,foxmail.com', '免费邮箱域名列表，逗号或换行分隔', NOW(), NOW()),
    ('signup_grant_risk_control_trusted_email_domains', '', '可信企业邮箱域名列表，逗号或换行分隔；命中后跳过邮箱域名频控', NOW(), NOW())
ON CONFLICT (key) DO NOTHING;
