-- Signup grant risk control records bonus eligibility decisions without storing raw email/IP.

CREATE TABLE IF NOT EXISTS signup_grant_claims (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    email_hash VARCHAR(64) NOT NULL DEFAULT '',
    email_domain_hash VARCHAR(64) NOT NULL DEFAULT '',
    ip_hash VARCHAR(64) NOT NULL DEFAULT '',
    user_agent_hash VARCHAR(64) NOT NULL DEFAULT '',
    signup_source VARCHAR(32) NOT NULL DEFAULT '',
    provider_type VARCHAR(32) NOT NULL DEFAULT '',
    provider_subject_hash VARCHAR(64) NOT NULL DEFAULT '',
    decision VARCHAR(20) NOT NULL CHECK (decision IN ('allowed', 'blocked')),
    reason TEXT NOT NULL DEFAULT '',
    grant_balance DECIMAL(20,8) NOT NULL DEFAULT 0,
    grant_metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_signup_grant_claims_email_allowed
    ON signup_grant_claims (email_hash, created_at DESC)
    WHERE decision = 'allowed';

CREATE INDEX IF NOT EXISTS idx_signup_grant_claims_domain_allowed
    ON signup_grant_claims (email_domain_hash, created_at DESC)
    WHERE decision = 'allowed' AND email_domain_hash <> '';

CREATE INDEX IF NOT EXISTS idx_signup_grant_claims_ip_allowed
    ON signup_grant_claims (ip_hash, created_at DESC)
    WHERE decision = 'allowed' AND ip_hash <> '';

CREATE INDEX IF NOT EXISTS idx_signup_grant_claims_user_created_at
    ON signup_grant_claims (user_id, created_at DESC);

INSERT INTO settings (key, value, updated_at) VALUES
    ('signup_grant_risk_control_enabled', 'false', NOW()),
    ('signup_grant_risk_control_email_limit', '1', NOW()),
    ('signup_grant_risk_control_ip_daily_limit', '3', NOW()),
    ('signup_grant_risk_control_domain_daily_limit', '10', NOW())
ON CONFLICT (key) DO NOTHING;
