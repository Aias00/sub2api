-- Persist the request context that created a user account.
-- This is separate from signup_grant_claims because registration diagnostics
-- must exist even when signup grants or risk controls are disabled.

CREATE TABLE IF NOT EXISTS user_registration_events (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email TEXT NOT NULL DEFAULT '',
    signup_source VARCHAR(32) NOT NULL DEFAULT '',
    provider_type VARCHAR(32) NOT NULL DEFAULT '',
    provider_subject TEXT NOT NULL DEFAULT '',
    ip_address TEXT NOT NULL DEFAULT '',
    user_agent TEXT NOT NULL DEFAULT '',
    accept_language TEXT NOT NULL DEFAULT '',
    device_fingerprint TEXT NOT NULL DEFAULT '',
    headers_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_registration_events_user_id
    ON user_registration_events(user_id);

CREATE INDEX IF NOT EXISTS idx_user_registration_events_ip_created
    ON user_registration_events(ip_address, created_at DESC)
    WHERE ip_address <> '';

CREATE INDEX IF NOT EXISTS idx_user_registration_events_source_created
    ON user_registration_events(signup_source, created_at DESC);
