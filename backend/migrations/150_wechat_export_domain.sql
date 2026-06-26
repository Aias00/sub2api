CREATE TABLE IF NOT EXISTS wechat_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending',
    login_token TEXT NOT NULL DEFAULT '',
    cookies_encrypted TEXT NOT NULL DEFAULT '',
    login_account_name TEXT NOT NULL DEFAULT '',
    last_validated_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_wechat_sessions_user_id
    ON wechat_sessions(user_id);

CREATE INDEX IF NOT EXISTS idx_wechat_sessions_status
    ON wechat_sessions(status);

CREATE INDEX IF NOT EXISTS idx_wechat_sessions_expires_at
    ON wechat_sessions(expires_at);

CREATE TABLE IF NOT EXISTS wechat_accounts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    fakeid TEXT NOT NULL,
    nickname TEXT NOT NULL DEFAULT '',
    alias TEXT NOT NULL DEFAULT '',
    avatar TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_synced_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, fakeid)
);

CREATE INDEX IF NOT EXISTS idx_wechat_accounts_user_id
    ON wechat_accounts(user_id);

CREATE INDEX IF NOT EXISTS idx_wechat_accounts_fakeid
    ON wechat_accounts(fakeid);

CREATE TABLE IF NOT EXISTS wechat_articles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_fakeid TEXT NOT NULL DEFAULT '',
    source_type TEXT NOT NULL DEFAULT 'direct_link',
    title TEXT NOT NULL DEFAULT '',
    author TEXT NOT NULL DEFAULT '',
    link TEXT NOT NULL,
    cover TEXT NOT NULL DEFAULT '',
    digest TEXT NOT NULL DEFAULT '',
    publish_at TIMESTAMPTZ,
    is_original BOOLEAN NOT NULL DEFAULT FALSE,
    is_pay_subscribe BOOLEAN NOT NULL DEFAULT FALSE,
    content_status TEXT NOT NULL DEFAULT 'pending',
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, link)
);

CREATE INDEX IF NOT EXISTS idx_wechat_articles_user_id
    ON wechat_articles(user_id);

CREATE INDEX IF NOT EXISTS idx_wechat_articles_fakeid
    ON wechat_articles(account_fakeid);

CREATE INDEX IF NOT EXISTS idx_wechat_articles_source_type
    ON wechat_articles(source_type);

CREATE INDEX IF NOT EXISTS idx_wechat_articles_publish_at
    ON wechat_articles(publish_at DESC NULLS LAST);

CREATE TABLE IF NOT EXISTS wechat_export_tasks (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'queued',
    selected_article_count INTEGER NOT NULL DEFAULT 0,
    successful_article_count INTEGER NOT NULL DEFAULT 0,
    failed_article_count INTEGER NOT NULL DEFAULT 0,
    formats_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    include_engagement BOOLEAN NOT NULL DEFAULT FALSE,
    payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    result_manifest_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    error_message TEXT NOT NULL DEFAULT '',
    worker_lease_until TIMESTAMPTZ,
    retention_days INTEGER NOT NULL DEFAULT 7,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_wechat_export_tasks_user_id
    ON wechat_export_tasks(user_id);

CREATE INDEX IF NOT EXISTS idx_wechat_export_tasks_status
    ON wechat_export_tasks(status);

CREATE INDEX IF NOT EXISTS idx_wechat_export_tasks_worker_lease_until
    ON wechat_export_tasks(worker_lease_until);

CREATE TABLE IF NOT EXISTS wechat_export_artifacts (
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT NOT NULL REFERENCES wechat_export_tasks(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    format TEXT NOT NULL,
    storage_provider TEXT NOT NULL DEFAULT '',
    storage_key TEXT NOT NULL DEFAULT '',
    download_url TEXT NOT NULL DEFAULT '',
    file_name TEXT NOT NULL DEFAULT '',
    file_size BIGINT NOT NULL DEFAULT 0,
    checksum TEXT NOT NULL DEFAULT '',
    expires_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_wechat_export_artifacts_task_id
    ON wechat_export_artifacts(task_id);

CREATE INDEX IF NOT EXISTS idx_wechat_export_artifacts_user_id
    ON wechat_export_artifacts(user_id);

CREATE INDEX IF NOT EXISTS idx_wechat_export_artifacts_expires_at
    ON wechat_export_artifacts(expires_at);
