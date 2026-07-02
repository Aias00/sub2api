-- Split user balance into paid and gift buckets while keeping users.balance as
-- the backward-compatible total balance.

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS paid_balance DECIMAL(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS gift_balance DECIMAL(20,8) NOT NULL DEFAULT 0;

UPDATE users
SET paid_balance = balance,
    gift_balance = 0
WHERE paid_balance = 0
  AND gift_balance = 0
  AND balance <> 0;

COMMENT ON COLUMN users.paid_balance IS 'Paid/recharge balance component. users.balance remains the total.';
COMMENT ON COLUMN users.gift_balance IS 'Gift/promotional balance component. Restricted to low-cost eligible usage.';

ALTER TABLE image_workspace_tasks
    ADD COLUMN IF NOT EXISTS reserved_paid_balance NUMERIC(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS reserved_gift_balance NUMERIC(20,8) NOT NULL DEFAULT 0;

ALTER TABLE wechat_export_tasks
    ADD COLUMN IF NOT EXISTS reserved_paid_balance NUMERIC(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS reserved_gift_balance NUMERIC(20,8) NOT NULL DEFAULT 0;

COMMENT ON COLUMN image_workspace_tasks.reserved_paid_balance IS 'Paid bucket amount reserved or spent by this task.';
COMMENT ON COLUMN image_workspace_tasks.reserved_gift_balance IS 'Gift bucket amount reserved or spent by this task.';
COMMENT ON COLUMN wechat_export_tasks.reserved_paid_balance IS 'Paid bucket amount reserved or spent by this task.';
COMMENT ON COLUMN wechat_export_tasks.reserved_gift_balance IS 'Gift bucket amount reserved or spent by this task.';

CREATE TABLE IF NOT EXISTS signup_grant_risk_overrides (
    id BIGSERIAL PRIMARY KEY,
    subject_type TEXT NOT NULL,
    subject_hash TEXT NOT NULL,
    action TEXT NOT NULL,
    reason TEXT NOT NULL DEFAULT '',
    created_by BIGINT NULL,
    expires_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT signup_grant_risk_overrides_subject_type_check
        CHECK (subject_type IN ('email', 'email_domain', 'ip', 'oauth_identity', 'device')),
    CONSTRAINT signup_grant_risk_overrides_action_check
        CHECK (action IN ('allow', 'block'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_signup_grant_risk_overrides_subject
    ON signup_grant_risk_overrides(subject_type, subject_hash);

CREATE INDEX IF NOT EXISTS idx_signup_grant_claims_created_at
    ON signup_grant_claims(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_signup_grant_claims_decision_created_at
    ON signup_grant_claims(decision, created_at DESC);
