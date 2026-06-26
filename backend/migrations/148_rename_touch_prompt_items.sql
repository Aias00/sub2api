DO $$
BEGIN
    IF to_regclass('public.prompt_catalog_items') IS NULL
       AND to_regclass('public.touch_prompt_items') IS NOT NULL THEN
        ALTER TABLE touch_prompt_items RENAME TO prompt_catalog_items;
    END IF;
END $$;

DROP INDEX IF EXISTS idx_touch_prompt_items_status_source_type;
DROP INDEX IF EXISTS idx_touch_prompt_items_category;
DROP INDEX IF EXISTS idx_touch_prompt_items_source_project;
DROP INDEX IF EXISTS idx_touch_prompt_items_source_url;
DROP INDEX IF EXISTS idx_touch_prompt_items_imported_at;

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
