-- Migration: 137_channel_monitor_health_state
-- Adds operational health metadata to channel monitor checks.

ALTER TABLE channel_monitors
    ADD COLUMN IF NOT EXISTS auto_disabled BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS auto_disabled_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS auto_disabled_reason VARCHAR(500) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS auto_recovered_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS last_health_status VARCHAR(20) NOT NULL DEFAULT 'unknown';

CREATE INDEX IF NOT EXISTS idx_channel_monitors_auto_disabled_health
    ON channel_monitors (auto_disabled, last_health_status);

ALTER TABLE channel_monitor_histories
    ADD COLUMN IF NOT EXISTS error_category VARCHAR(32) NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_channel_monitor_histories_monitor_error_category_checked
    ON channel_monitor_histories (monitor_id, error_category, checked_at DESC);
