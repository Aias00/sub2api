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
        'ai-hot-jiqizhixin',
        'rss-generic',
        '机器之心',
        'AI Hot trusted Chinese AI news source.',
        TRUE,
        'https://www.jiqizhixin.com',
        '["https://www.jiqizhixin.com/feed"]'::jsonb,
        '{"category":"news","category_label":"AI 资讯","ai_only":true,"lang":"zh","priority":1,"source_weight":10,"source_handle":"jiqizhixin.com"}'::jsonb,
        100
    ),
    (
        'ai-hot-qbitai',
        'rss-generic',
        '量子位',
        'AI Hot trusted Chinese AI news source.',
        TRUE,
        'https://www.qbitai.com',
        '["https://www.qbitai.com/feed"]'::jsonb,
        '{"category":"news","category_label":"AI 资讯","ai_only":true,"lang":"zh","priority":1,"source_weight":8,"source_handle":"qbitai.com"}'::jsonb,
        110
    ),
    (
        'ai-hot-xinzhiyuan',
        'rss-generic',
        '新智元',
        'AI Hot trusted Chinese AI news source.',
        TRUE,
        'https://www.xinzhiyuan.com',
        '["https://www.xinzhiyuan.com/feed"]'::jsonb,
        '{"category":"news","category_label":"AI 资讯","ai_only":true,"lang":"zh","priority":1,"source_weight":8,"source_handle":"xinzhiyuan.com"}'::jsonb,
        120
    ),
    (
        'ai-hot-techcrunch-ai',
        'rss-generic',
        'TechCrunch AI',
        'AI Hot trusted English AI news source.',
        TRUE,
        'https://techcrunch.com',
        '["https://techcrunch.com/category/artificial-intelligence/feed/"]'::jsonb,
        '{"category":"news","category_label":"AI 资讯","ai_only":true,"lang":"en","priority":1,"source_weight":10,"source_handle":"techcrunch.com"}'::jsonb,
        130
    ),
    (
        'ai-hot-the-verge-ai',
        'rss-generic',
        'The Verge AI',
        'AI Hot trusted English AI news source.',
        TRUE,
        'https://www.theverge.com',
        '["https://www.theverge.com/rss/ai-artificial-intelligence/index.xml"]'::jsonb,
        '{"category":"news","category_label":"AI 资讯","ai_only":true,"lang":"en","priority":1,"source_weight":8,"source_handle":"theverge.com"}'::jsonb,
        140
    ),
    (
        'ai-hot-mit-tech-review',
        'rss-generic',
        'MIT Tech Review',
        'AI Hot filtered technology news source.',
        TRUE,
        'https://www.technologyreview.com',
        '["https://www.technologyreview.com/feed/"]'::jsonb,
        '{"category":"news","category_label":"AI 资讯","ai_only":false,"lang":"en","priority":2,"source_weight":8,"source_handle":"technologyreview.com"}'::jsonb,
        150
    ),
    (
        'ai-hot-ars-technica-ai',
        'rss-generic',
        'Ars Technica AI',
        'AI Hot filtered technology news source.',
        TRUE,
        'https://arstechnica.com',
        '["https://feeds.arstechnica.com/arstechnica/technology-lab"]'::jsonb,
        '{"category":"news","category_label":"AI 资讯","ai_only":false,"lang":"en","priority":2,"source_weight":6,"source_handle":"arstechnica.com"}'::jsonb,
        160
    ),
    (
        'ai-hot-venturebeat-ai',
        'rss-generic',
        'VentureBeat AI',
        'AI Hot trusted English AI news source.',
        TRUE,
        'https://venturebeat.com',
        '["https://venturebeat.com/category/ai/feed/"]'::jsonb,
        '{"category":"news","category_label":"AI 资讯","ai_only":true,"lang":"en","priority":2,"source_weight":7,"source_handle":"venturebeat.com"}'::jsonb,
        170
    ),
    (
        'ai-hot-36kr',
        'rss-generic',
        '36氪',
        'AI Hot filtered Chinese technology news source.',
        TRUE,
        'https://36kr.com',
        '["https://36kr.com/feed"]'::jsonb,
        '{"category":"news","category_label":"AI 资讯","ai_only":false,"lang":"zh","priority":4,"source_weight":8,"source_handle":"36kr.com"}'::jsonb,
        180
    ),
    (
        'ai-hot-infoq-ai',
        'rss-generic',
        'InfoQ AI',
        'AI Hot trusted Chinese AI news source.',
        TRUE,
        'https://www.infoq.cn',
        '["https://www.infoq.cn/feed"]'::jsonb,
        '{"category":"news","category_label":"AI 资讯","ai_only":true,"lang":"zh","priority":3,"source_weight":7,"source_handle":"infoq.cn"}'::jsonb,
        190
    ),
    (
        'ai-hot-linux-do',
        'rss-generic',
        'LINUX DO',
        'AI Hot filtered community news source.',
        TRUE,
        'https://linux.do',
        '["https://linux.do/c/news/34.rss"]'::jsonb,
        '{"category":"news","category_label":"AI 资讯","ai_only":false,"lang":"zh","priority":2,"source_weight":5,"source_handle":"linux.do"}'::jsonb,
        200
    ),
    (
        'ai-hot-hacker-news-ai',
        'rss-generic',
        'Hacker News AI',
        'AI Hot noisy upstream, seeded disabled by default.',
        FALSE,
        'https://hnrss.org',
        '["https://hnrss.org/newest?q=AI+OR+LLM+OR+GPT+OR+%22artificial+intelligence%22+OR+%22machine+learning%22&points=100"]'::jsonb,
        '{"category":"news","category_label":"AI 资讯","ai_only":true,"lang":"en","priority":2,"source_weight":0,"source_handle":"hnrss.org","noisy":true}'::jsonb,
        210
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
