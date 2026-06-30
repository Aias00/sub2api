CREATE TABLE IF NOT EXISTS image_workspace_tasks (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'queued',
    prompt TEXT NOT NULL,
    negative_prompt TEXT NOT NULL DEFAULT '',
    model TEXT NOT NULL DEFAULT '',
    provider TEXT NOT NULL DEFAULT 'openai',
    size TEXT NOT NULL DEFAULT '1024x1024',
    quality TEXT NOT NULL DEFAULT 'standard',
    style TEXT NOT NULL DEFAULT '',
    seed BIGINT,
    batch_size INTEGER NOT NULL DEFAULT 1,
    template_id BIGINT,
    worker_lease_until TIMESTAMPTZ,
    cost_estimate NUMERIC(20,8) NOT NULL DEFAULT 0,
    balance_snapshot NUMERIC(20,8) NOT NULL DEFAULT 0,
    error_message TEXT NOT NULL DEFAULT '',
    result_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_image_workspace_tasks_user_id
    ON image_workspace_tasks(user_id, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_image_workspace_tasks_status
    ON image_workspace_tasks(status);

CREATE INDEX IF NOT EXISTS idx_image_workspace_tasks_worker_claim
    ON image_workspace_tasks(status, worker_lease_until, created_at, id);

CREATE TABLE IF NOT EXISTS image_workspace_artifacts (
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT NOT NULL REFERENCES image_workspace_tasks(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    storage_provider TEXT NOT NULL DEFAULT '',
    storage_key TEXT NOT NULL DEFAULT '',
    image_url TEXT NOT NULL DEFAULT '',
    prompt TEXT NOT NULL DEFAULT '',
    mime_type TEXT NOT NULL DEFAULT '',
    width INTEGER NOT NULL DEFAULT 0,
    height INTEGER NOT NULL DEFAULT 0,
    file_size BIGINT NOT NULL DEFAULT 0,
    checksum TEXT NOT NULL DEFAULT '',
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_image_workspace_artifacts_task_id
    ON image_workspace_artifacts(task_id);

CREATE INDEX IF NOT EXISTS idx_image_workspace_artifacts_user_id
    ON image_workspace_artifacts(user_id, created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS image_workspace_templates (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    prompt TEXT NOT NULL,
    negative_prompt TEXT NOT NULL DEFAULT '',
    model TEXT NOT NULL DEFAULT '',
    size TEXT NOT NULL DEFAULT '1024x1024',
    quality TEXT NOT NULL DEFAULT 'standard',
    style TEXT NOT NULL DEFAULT '',
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_image_workspace_templates_user_id
    ON image_workspace_templates(user_id, updated_at DESC, id DESC);
