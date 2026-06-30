-- Legacy Hot presentation/import tables are kept for schema compatibility only.
-- The live collector writes current content to hot_items and diagnostics to
-- hot_runs/hot_checkpoints/hot_run_events.
TRUNCATE TABLE
    hot_item_media,
    hot_media_assets,
    hot_feed_items,
    hot_daily_stories,
    hot_daily_sections,
    hot_daily_issues,
    hot_mp_entries
RESTART IDENTITY;
