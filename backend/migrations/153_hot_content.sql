CREATE TABLE IF NOT EXISTS hot_sources (
    id BIGSERIAL PRIMARY KEY,
    source_id TEXT NOT NULL UNIQUE,
    adapter_kind TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    base_url TEXT NOT NULL DEFAULT '',
    seed_urls_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    config_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_hot_sources_enabled_adapter
    ON hot_sources(enabled, adapter_kind, sort_order, id);

CREATE TABLE IF NOT EXISTS hot_runs (
    id BIGSERIAL PRIMARY KEY,
    run_id TEXT NOT NULL UNIQUE,
    source_id TEXT NOT NULL REFERENCES hot_sources(source_id) ON DELETE CASCADE,
    status TEXT NOT NULL,
    trigger_type TEXT NOT NULL DEFAULT '',
    request_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    summary_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    error_message TEXT NOT NULL DEFAULT '',
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_hot_runs_source_created
    ON hot_runs(source_id, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_hot_runs_status_created
    ON hot_runs(status, created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS hot_checkpoints (
    id BIGSERIAL PRIMARY KEY,
    source_id TEXT NOT NULL UNIQUE REFERENCES hot_sources(source_id) ON DELETE CASCADE,
    checkpoint_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS hot_items (
    id BIGSERIAL PRIMARY KEY,
    source_id TEXT NOT NULL REFERENCES hot_sources(source_id) ON DELETE CASCADE,
    external_id TEXT NOT NULL,
    canonical_url TEXT NOT NULL DEFAULT '',
    title TEXT NOT NULL,
    summary TEXT NOT NULL DEFAULT '',
    body TEXT NOT NULL DEFAULT '',
    quoted TEXT NOT NULL DEFAULT '',
    reason TEXT NOT NULL DEFAULT '',
    published_at TIMESTAMPTZ,
    author TEXT NOT NULL DEFAULT '',
    source_name TEXT NOT NULL DEFAULT '',
    source_handle TEXT NOT NULL DEFAULT '',
    badge TEXT NOT NULL DEFAULT '',
    score TEXT NOT NULL DEFAULT '',
    content_type TEXT NOT NULL DEFAULT 'article',
    tags_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    metrics_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    raw_ref_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    content_hash TEXT NOT NULL DEFAULT '',
    has_media BOOLEAN NOT NULL DEFAULT FALSE,
    status TEXT NOT NULL DEFAULT 'published',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(source_id, external_id)
);

CREATE INDEX IF NOT EXISTS idx_hot_items_published_at
    ON hot_items(published_at DESC NULLS LAST, id DESC);

CREATE INDEX IF NOT EXISTS idx_hot_items_source_published
    ON hot_items(source_id, published_at DESC NULLS LAST, id DESC);

CREATE INDEX IF NOT EXISTS idx_hot_items_status_published
    ON hot_items(status, published_at DESC NULLS LAST, id DESC);

CREATE TABLE IF NOT EXISTS hot_item_media (
    id BIGSERIAL PRIMARY KEY,
    hot_item_id BIGINT NOT NULL REFERENCES hot_items(id) ON DELETE CASCADE,
    media_type TEXT NOT NULL DEFAULT '',
    original_url TEXT NOT NULL DEFAULT '',
    cover_url TEXT NOT NULL DEFAULT '',
    thumb_url TEXT NOT NULL DEFAULT '',
    mime_type TEXT NOT NULL DEFAULT '',
    width INTEGER NOT NULL DEFAULT 0,
    height INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_hot_item_media_item
    ON hot_item_media(hot_item_id, id);
