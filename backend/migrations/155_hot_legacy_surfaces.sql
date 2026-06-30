CREATE TABLE IF NOT EXISTS hot_media_assets (
    id BIGSERIAL PRIMARY KEY,
    asset_id TEXT NOT NULL UNIQUE,
    source_kind TEXT NOT NULL DEFAULT '',
    source_origin_url TEXT NOT NULL DEFAULT '',
    storage_key_original TEXT NOT NULL DEFAULT '',
    storage_key_cover TEXT NOT NULL DEFAULT '',
    storage_key_thumb TEXT NOT NULL DEFAULT '',
    original_url TEXT NOT NULL DEFAULT '',
    cover_url TEXT NOT NULL DEFAULT '',
    thumb_url TEXT NOT NULL DEFAULT '',
    mime_type TEXT NOT NULL DEFAULT '',
    width INTEGER NOT NULL DEFAULT 0,
    height INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'pending',
    content_hash TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_hot_media_assets_status
    ON hot_media_assets(status, created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS hot_run_events (
    id BIGSERIAL PRIMARY KEY,
    legacy_id BIGINT NOT NULL,
    run_id TEXT NOT NULL REFERENCES hot_runs(run_id) ON DELETE CASCADE,
    node TEXT NOT NULL DEFAULT '',
    message TEXT NOT NULL DEFAULT '',
    payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(run_id, legacy_id)
);

CREATE INDEX IF NOT EXISTS idx_hot_run_events_run_created
    ON hot_run_events(run_id, created_at ASC, id ASC);

CREATE TABLE IF NOT EXISTS hot_feed_items (
    id BIGSERIAL PRIMARY KEY,
    legacy_id BIGINT NOT NULL UNIQUE,
    day_label TEXT NOT NULL DEFAULT '',
    time_label TEXT NOT NULL DEFAULT '',
    source TEXT NOT NULL DEFAULT '',
    handle TEXT NOT NULL DEFAULT '',
    badge TEXT NOT NULL DEFAULT '',
    score TEXT NOT NULL DEFAULT '',
    title TEXT NOT NULL DEFAULT '',
    body TEXT NOT NULL DEFAULT '',
    quoted TEXT NOT NULL DEFAULT '',
    reason TEXT NOT NULL DEFAULT '',
    tags_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    has_media BOOLEAN NOT NULL DEFAULT FALSE,
    link TEXT NOT NULL DEFAULT '',
    sort_index INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_hot_feed_items_day_sort
    ON hot_feed_items(day_label DESC, sort_index ASC, id ASC);

CREATE TABLE IF NOT EXISTS hot_daily_issues (
    issue_date TEXT PRIMARY KEY,
    month_label TEXT NOT NULL DEFAULT '',
    headline TEXT NOT NULL DEFAULT '',
    event_count INTEGER NOT NULL DEFAULT 0,
    volume TEXT NOT NULL DEFAULT '',
    issue_title TEXT NOT NULL DEFAULT '',
    masthead_date TEXT NOT NULL DEFAULT '',
    tagline TEXT NOT NULL DEFAULT '',
    is_latest BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_hot_daily_issues_latest
    ON hot_daily_issues(is_latest DESC, issue_date DESC);

CREATE TABLE IF NOT EXISTS hot_daily_sections (
    issue_date TEXT NOT NULL REFERENCES hot_daily_issues(issue_date) ON DELETE CASCADE,
    section_no TEXT NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    english TEXT NOT NULL DEFAULT '',
    count_label TEXT NOT NULL DEFAULT '',
    sort_index INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (issue_date, section_no)
);

CREATE TABLE IF NOT EXISTS hot_daily_stories (
    issue_date TEXT NOT NULL,
    section_no TEXT NOT NULL,
    sort_index INTEGER NOT NULL DEFAULT 0,
    title TEXT NOT NULL DEFAULT '',
    href TEXT NOT NULL DEFAULT '',
    source_role TEXT NOT NULL DEFAULT '',
    source TEXT NOT NULL DEFAULT '',
    summary TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (issue_date, section_no, sort_index),
    FOREIGN KEY (issue_date, section_no)
        REFERENCES hot_daily_sections(issue_date, section_no)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS hot_mp_entries (
    id BIGSERIAL PRIMARY KEY,
    legacy_id BIGINT NOT NULL UNIQUE,
    published_date TEXT NOT NULL DEFAULT '',
    title TEXT NOT NULL DEFAULT '',
    account TEXT NOT NULL DEFAULT '',
    href TEXT NOT NULL DEFAULT '',
    account_href TEXT NOT NULL DEFAULT '',
    badge TEXT NOT NULL DEFAULT '',
    reads TEXT NOT NULL DEFAULT '',
    likes TEXT NOT NULL DEFAULT '',
    shares TEXT NOT NULL DEFAULT '',
    outlier TEXT NOT NULL DEFAULT '',
    sort_index INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_hot_mp_entries_date_sort
    ON hot_mp_entries(published_date DESC, sort_index ASC, id ASC);
