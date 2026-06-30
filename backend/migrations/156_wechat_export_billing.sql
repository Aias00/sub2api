-- 156_wechat_export_billing.sql
-- Add billing support to wechat_export: cost reservation at task creation,
-- adjustment at completion, refund on failure/cancellation, and usage records.

-- Add billing columns to wechat_export_tasks
ALTER TABLE wechat_export_tasks
    ADD COLUMN IF NOT EXISTS cost_estimate NUMERIC(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS balance_snapshot NUMERIC(20,8) NOT NULL DEFAULT 0;

-- Usage records for billing audit trail
CREATE TABLE IF NOT EXISTS wechat_export_usage_records (
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT NOT NULL REFERENCES wechat_export_tasks(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    article_count INTEGER NOT NULL DEFAULT 0,
    format_count INTEGER NOT NULL DEFAULT 0,
    include_engagement BOOLEAN NOT NULL DEFAULT FALSE,
    reserved_cost NUMERIC(20,8) NOT NULL DEFAULT 0,
    actual_cost NUMERIC(20,8) NOT NULL DEFAULT 0,
    balance_snapshot NUMERIC(20,8) NOT NULL DEFAULT 0,
    billing_status TEXT NOT NULL DEFAULT 'settled',
    metadata_json JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(task_id)
);

CREATE INDEX IF NOT EXISTS idx_wechat_export_usage_records_user_id ON wechat_export_usage_records(user_id);
CREATE INDEX IF NOT EXISTS idx_wechat_export_usage_records_billing_status ON wechat_export_usage_records(billing_status);
