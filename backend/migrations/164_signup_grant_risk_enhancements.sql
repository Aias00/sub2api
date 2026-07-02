-- Extend signup grant risk controls with OAuth identity, device, and domain-quality settings.

CREATE INDEX IF NOT EXISTS idx_signup_grant_claims_provider_subject_allowed
    ON signup_grant_claims (provider_type, provider_subject_hash, created_at DESC)
    WHERE decision = 'allowed' AND provider_subject_hash <> '';

CREATE INDEX IF NOT EXISTS idx_signup_grant_claims_device_allowed
    ON signup_grant_claims (user_agent_hash, created_at DESC)
    WHERE decision = 'allowed' AND user_agent_hash <> '';

INSERT INTO settings (key, value, updated_at) VALUES
    ('signup_grant_risk_control_oauth_identity_enabled', 'true', NOW()),
    ('signup_grant_risk_control_device_enabled', 'true', NOW()),
    ('signup_grant_risk_control_device_daily_limit', '2', NOW()),
    ('signup_grant_risk_control_free_domain_daily_limit', '5', NOW()),
    ('signup_grant_risk_control_blocked_email_domains', '', NOW()),
    ('signup_grant_risk_control_free_email_domains', 'gmail.com,googlemail.com,outlook.com,hotmail.com,live.com,icloud.com,yahoo.com,qq.com,163.com,126.com,foxmail.com', NOW()),
    ('signup_grant_risk_control_trusted_email_domains', '', NOW())
ON CONFLICT (key) DO NOTHING;
