-- 160_wechat_export_global_content.sql
-- Normalize WeChat export storage so public accounts and articles are stored
-- once globally, while user-specific visibility is represented by bindings.

CREATE TABLE IF NOT EXISTS wechat_public_accounts (
    id BIGSERIAL PRIMARY KEY,
    fakeid TEXT NOT NULL UNIQUE,
    nickname TEXT NOT NULL DEFAULT '',
    alias TEXT NOT NULL DEFAULT '',
    avatar TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_wechat_public_accounts_fakeid
    ON wechat_public_accounts(fakeid);

CREATE TABLE IF NOT EXISTS wechat_account_bindings (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id BIGINT NOT NULL REFERENCES wechat_public_accounts(id) ON DELETE CASCADE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_synced_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, account_id)
);

CREATE INDEX IF NOT EXISTS idx_wechat_account_bindings_user_id
    ON wechat_account_bindings(user_id);

CREATE INDEX IF NOT EXISTS idx_wechat_account_bindings_account_id
    ON wechat_account_bindings(account_id);

CREATE TABLE IF NOT EXISTS wechat_public_articles (
    id BIGSERIAL PRIMARY KEY,
    account_id BIGINT REFERENCES wechat_public_accounts(id) ON DELETE SET NULL,
    title TEXT NOT NULL DEFAULT '',
    author TEXT NOT NULL DEFAULT '',
    link TEXT NOT NULL UNIQUE,
    cover TEXT NOT NULL DEFAULT '',
    digest TEXT NOT NULL DEFAULT '',
    publish_at TIMESTAMPTZ,
    is_original BOOLEAN NOT NULL DEFAULT FALSE,
    is_pay_subscribe BOOLEAN NOT NULL DEFAULT FALSE,
    content_status TEXT NOT NULL DEFAULT 'pending',
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_wechat_public_articles_account_id
    ON wechat_public_articles(account_id);

CREATE INDEX IF NOT EXISTS idx_wechat_public_articles_publish_at
    ON wechat_public_articles(publish_at DESC NULLS LAST);

CREATE TABLE IF NOT EXISTS wechat_article_bindings (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    article_id BIGINT NOT NULL REFERENCES wechat_public_articles(id) ON DELETE CASCADE,
    source_type TEXT NOT NULL DEFAULT 'direct_link',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, article_id)
);

CREATE INDEX IF NOT EXISTS idx_wechat_article_bindings_user_id
    ON wechat_article_bindings(user_id);

CREATE INDEX IF NOT EXISTS idx_wechat_article_bindings_article_id
    ON wechat_article_bindings(article_id);

INSERT INTO wechat_public_accounts (
    fakeid, nickname, alias, avatar, description, created_at, updated_at
)
SELECT DISTINCT ON (fakeid)
    fakeid,
    nickname,
    alias,
    avatar,
    description,
    MIN(created_at) OVER (PARTITION BY fakeid),
    MAX(updated_at) OVER (PARTITION BY fakeid)
FROM wechat_accounts
WHERE fakeid <> ''
ORDER BY fakeid, updated_at DESC, id DESC
ON CONFLICT (fakeid) DO UPDATE SET
    nickname = COALESCE(NULLIF(EXCLUDED.nickname, ''), wechat_public_accounts.nickname),
    alias = COALESCE(NULLIF(EXCLUDED.alias, ''), wechat_public_accounts.alias),
    avatar = COALESCE(NULLIF(EXCLUDED.avatar, ''), wechat_public_accounts.avatar),
    description = COALESCE(NULLIF(EXCLUDED.description, ''), wechat_public_accounts.description),
    updated_at = GREATEST(wechat_public_accounts.updated_at, EXCLUDED.updated_at);

INSERT INTO wechat_account_bindings (
    user_id, account_id, is_active, last_synced_at, created_at, updated_at
)
SELECT
    old.user_id,
    acct.id,
    old.is_active,
    old.last_synced_at,
    old.created_at,
    old.updated_at
FROM wechat_accounts old
JOIN wechat_public_accounts acct ON acct.fakeid = old.fakeid
ON CONFLICT (user_id, account_id) DO UPDATE SET
    is_active = EXCLUDED.is_active,
    last_synced_at = COALESCE(EXCLUDED.last_synced_at, wechat_account_bindings.last_synced_at),
    updated_at = GREATEST(wechat_account_bindings.updated_at, EXCLUDED.updated_at);

WITH article_accounts AS (
    SELECT
        article.id AS old_article_id,
        acct.id AS account_id
    FROM wechat_articles article
    LEFT JOIN wechat_public_accounts acct ON acct.fakeid = article.account_fakeid
)
INSERT INTO wechat_public_articles (
    account_id, title, author, link, cover, digest, publish_at, is_original,
    is_pay_subscribe, content_status, metadata_json, created_at, updated_at
)
SELECT DISTINCT ON (article.link)
    article_accounts.account_id,
    article.title,
    article.author,
    article.link,
    article.cover,
    article.digest,
    article.publish_at,
    article.is_original,
    article.is_pay_subscribe,
    article.content_status,
    article.metadata_json,
    MIN(article.created_at) OVER (PARTITION BY article.link),
    MAX(article.updated_at) OVER (PARTITION BY article.link)
FROM wechat_articles article
LEFT JOIN article_accounts ON article_accounts.old_article_id = article.id
WHERE article.link <> ''
ORDER BY article.link, article.updated_at DESC, article.id DESC
ON CONFLICT (link) DO UPDATE SET
    account_id = COALESCE(EXCLUDED.account_id, wechat_public_articles.account_id),
    title = COALESCE(NULLIF(EXCLUDED.title, ''), wechat_public_articles.title),
    author = COALESCE(NULLIF(EXCLUDED.author, ''), wechat_public_articles.author),
    cover = COALESCE(NULLIF(EXCLUDED.cover, ''), wechat_public_articles.cover),
    digest = COALESCE(NULLIF(EXCLUDED.digest, ''), wechat_public_articles.digest),
    publish_at = COALESCE(EXCLUDED.publish_at, wechat_public_articles.publish_at),
    is_original = EXCLUDED.is_original,
    is_pay_subscribe = EXCLUDED.is_pay_subscribe,
    content_status = COALESCE(NULLIF(EXCLUDED.content_status, ''), wechat_public_articles.content_status),
    metadata_json = CASE
        WHEN wechat_public_articles.metadata_json = '{}'::jsonb OR wechat_public_articles.metadata_json IS NULL
        THEN EXCLUDED.metadata_json
        WHEN EXCLUDED.metadata_json = '{}'::jsonb OR EXCLUDED.metadata_json IS NULL
        THEN wechat_public_articles.metadata_json
        ELSE wechat_public_articles.metadata_json || EXCLUDED.metadata_json
    END,
    updated_at = GREATEST(wechat_public_articles.updated_at, EXCLUDED.updated_at);

INSERT INTO wechat_article_bindings (
    user_id, article_id, source_type, created_at, updated_at
)
SELECT
    old.user_id,
    article.id,
    old.source_type,
    old.created_at,
    old.updated_at
FROM wechat_articles old
JOIN wechat_public_articles article ON article.link = old.link
ON CONFLICT (user_id, article_id) DO UPDATE SET
    source_type = COALESCE(NULLIF(EXCLUDED.source_type, ''), wechat_article_bindings.source_type),
    updated_at = GREATEST(wechat_article_bindings.updated_at, EXCLUDED.updated_at);
