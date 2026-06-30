CREATE TABLE IF NOT EXISTS wechat_export_task_logs (
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT NOT NULL REFERENCES wechat_export_tasks(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT '',
    message TEXT NOT NULL DEFAULT '',
    meta_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_wechat_export_task_logs_task_id
    ON wechat_export_task_logs(task_id, created_at ASC, id ASC);

CREATE INDEX IF NOT EXISTS idx_wechat_export_task_logs_user_id
    ON wechat_export_task_logs(user_id);
