UPDATE hot_items AS item
SET
    source_name = COALESCE(NULLIF(source.title, ''), NULLIF(item.source_name, ''), item.source_id),
    body = COALESCE(NULLIF(item.body, ''), NULLIF(item.summary, ''), item.title),
    reason = COALESCE(NULLIF(item.reason, ''), item.metrics_json ->> 'reason', 'RSS 采集'),
    badge = COALESCE(
        NULLIF(item.badge, ''),
        item.metrics_json ->> 'source_category_label',
        CASE WHEN source.config_json ->> 'category' = 'official' THEN '官方信源' ELSE '博客文章' END
    ),
    score = COALESCE(
        NULLIF(item.score, ''),
        item.metrics_json ->> 'hot_score',
        CASE WHEN source.config_json ->> 'category' = 'official' THEN '70' ELSE '60' END
    )
FROM hot_sources AS source
WHERE item.source_id = source.source_id;
