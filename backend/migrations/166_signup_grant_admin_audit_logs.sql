-- Audit trail for signup grant operations performed by administrators.

CREATE TABLE IF NOT EXISTS signup_grant_admin_audit_logs (
    id BIGSERIAL PRIMARY KEY,
    operation TEXT NOT NULL,
    target_user_id BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    subject_type TEXT NOT NULL DEFAULT '',
    subject_hash TEXT NOT NULL DEFAULT '',
    action TEXT NOT NULL DEFAULT '',
    amount DECIMAL(20,8) NOT NULL DEFAULT 0,
    reason TEXT NOT NULL DEFAULT '',
    admin_id BIGINT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_signup_grant_admin_audit_logs_created_at
    ON signup_grant_admin_audit_logs(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_signup_grant_admin_audit_logs_admin_created_at
    ON signup_grant_admin_audit_logs(admin_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_signup_grant_admin_audit_logs_user_created_at
    ON signup_grant_admin_audit_logs(target_user_id, created_at DESC);
