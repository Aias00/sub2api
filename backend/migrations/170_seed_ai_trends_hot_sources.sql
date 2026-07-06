INSERT INTO hot_sources (
    source_id,
    adapter_kind,
    title,
    description,
    enabled,
    base_url,
    seed_urls_json,
    config_json,
    sort_order
) VALUES
    (
        'ai-hot-hacker-news-api',
        'ai-trends',
        'Hacker News AI API',
        'AI Hot native Hacker News API collector.',
        TRUE,
        'https://news.ycombinator.com',
        '[]'::jsonb,
        '{"category":"news","category_label":"AI 资讯","ai_only":true,"lang":"en","priority":2,"source_weight":5,"source_handle":"news.ycombinator.com","sources":["hackernews"],"limit_per_source":10}'::jsonb,
        205
    ),
    (
        'ai-hot-devto-api',
        'ai-trends',
        'Dev.to AI API',
        'AI Hot native Dev.to API collector.',
        TRUE,
        'https://dev.to',
        '[]'::jsonb,
        '{"category":"news","category_label":"AI 资讯","ai_only":true,"lang":"en","priority":2,"source_weight":5,"source_handle":"dev.to","sources":["devto"],"limit_per_source":10}'::jsonb,
        206
    )
ON CONFLICT (source_id) DO UPDATE SET
    adapter_kind = EXCLUDED.adapter_kind,
    title = EXCLUDED.title,
    description = EXCLUDED.description,
    base_url = EXCLUDED.base_url,
    seed_urls_json = EXCLUDED.seed_urls_json,
    config_json = EXCLUDED.config_json,
    sort_order = EXCLUDED.sort_order,
    updated_at = NOW();
