-- Admin audit visibility: allow administrators to inspect raw signup grant
-- identifiers while keeping hash columns for matching and rate limiting.

ALTER TABLE signup_grant_claims
    ADD COLUMN IF NOT EXISTS email TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS email_domain TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS ip_address TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS provider_subject TEXT NOT NULL DEFAULT '';

ALTER TABLE signup_grant_risk_overrides
    ADD COLUMN IF NOT EXISTS subject_value TEXT NOT NULL DEFAULT '';

ALTER TABLE signup_grant_admin_audit_logs
    ADD COLUMN IF NOT EXISTS subject_value TEXT NOT NULL DEFAULT '';
