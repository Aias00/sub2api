CREATE TABLE IF NOT EXISTS image_workspace_usage_records (
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT NOT NULL REFERENCES image_workspace_tasks(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider TEXT NOT NULL DEFAULT '',
    model TEXT NOT NULL DEFAULT '',
    size TEXT NOT NULL DEFAULT '',
    quality TEXT NOT NULL DEFAULT '',
    image_count INTEGER NOT NULL DEFAULT 0,
    reserved_cost NUMERIC(20,8) NOT NULL DEFAULT 0,
    actual_cost NUMERIC(20,8) NOT NULL DEFAULT 0,
    balance_snapshot NUMERIC(20,8) NOT NULL DEFAULT 0,
    billing_status TEXT NOT NULL DEFAULT 'settled',
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_image_workspace_usage_records_task_id
    ON image_workspace_usage_records(task_id);

CREATE INDEX IF NOT EXISTS idx_image_workspace_usage_records_user_id_created
    ON image_workspace_usage_records(user_id, created_at DESC, id DESC);
