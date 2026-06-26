CREATE TABLE IF NOT EXISTS prompt_catalog_items (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    prompt TEXT NOT NULL,
    prompt_preview TEXT NOT NULL DEFAULT '',
    category TEXT NOT NULL DEFAULT 'general',
    tags JSONB NOT NULL DEFAULT '[]'::jsonb,
    model_tags JSONB NOT NULL DEFAULT '[]'::jsonb,
    source_url TEXT,
    image_url TEXT,
    image_urls JSONB NOT NULL DEFAULT '[]'::jsonb,
    source_project TEXT NOT NULL DEFAULT 'manual',
    source_type TEXT NOT NULL DEFAULT 'case',
    source_label TEXT,
    github_url TEXT,
    featured BOOLEAN NOT NULL DEFAULT FALSE,
    styles JSONB NOT NULL DEFAULT '[]'::jsonb,
    scenes JSONB NOT NULL DEFAULT '[]'::jsonb,
    image_original_url TEXT,
    image_preview_url TEXT,
    image_thumb_url TEXT,
    import_source TEXT NOT NULL DEFAULT 'catalog',
    raw_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    status TEXT NOT NULL DEFAULT 'published',
    imported_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_prompt_catalog_items_status_source_type
    ON prompt_catalog_items(status, source_type)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_prompt_catalog_items_category
    ON prompt_catalog_items(category)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_prompt_catalog_items_source_project
    ON prompt_catalog_items(source_project)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_prompt_catalog_items_source_url
    ON prompt_catalog_items(source_url)
    WHERE source_url IS NOT NULL AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_prompt_catalog_items_imported_at
    ON prompt_catalog_items(imported_at DESC NULLS LAST, title ASC)
    WHERE deleted_at IS NULL;
