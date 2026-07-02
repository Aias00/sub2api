-- Separate Touch-created users from native Cloudbase identities.
-- Touch users are keyed by auth_identities(provider_type='touch') instead of email.

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_signup_source_check;

ALTER TABLE users
    ADD CONSTRAINT users_signup_source_check
    CHECK (signup_source IN ('email', 'linuxdo', 'wechat', 'oidc', 'github', 'google', 'dingtalk', 'touch'));

ALTER TABLE auth_identities
    DROP CONSTRAINT IF EXISTS auth_identities_provider_type_check;

ALTER TABLE auth_identities
    ADD CONSTRAINT auth_identities_provider_type_check
    CHECK (provider_type IN ('email', 'linuxdo', 'wechat', 'oidc', 'github', 'google', 'dingtalk', 'touch'));

ALTER TABLE auth_identity_channels
    DROP CONSTRAINT IF EXISTS auth_identity_channels_provider_type_check;

ALTER TABLE auth_identity_channels
    ADD CONSTRAINT auth_identity_channels_provider_type_check
    CHECK (provider_type IN ('email', 'linuxdo', 'wechat', 'oidc', 'github', 'google', 'dingtalk', 'touch'));

ALTER TABLE pending_auth_sessions
    DROP CONSTRAINT IF EXISTS pending_auth_sessions_provider_type_check;

ALTER TABLE pending_auth_sessions
    ADD CONSTRAINT pending_auth_sessions_provider_type_check
    CHECK (provider_type IN ('email', 'linuxdo', 'wechat', 'oidc', 'github', 'google', 'dingtalk', 'touch'));

ALTER TABLE user_provider_default_grants
    DROP CONSTRAINT IF EXISTS user_provider_default_grants_provider_type_check;

ALTER TABLE user_provider_default_grants
    ADD CONSTRAINT user_provider_default_grants_provider_type_check
    CHECK (provider_type IN ('email', 'linuxdo', 'wechat', 'oidc', 'github', 'google', 'dingtalk', 'touch'));

DROP INDEX IF EXISTS users_email_unique_active;

CREATE UNIQUE INDEX IF NOT EXISTS users_email_unique_active_non_touch
    ON users(email)
    WHERE deleted_at IS NULL AND signup_source <> 'touch';

DELETE FROM auth_identities ai
USING users u
WHERE ai.user_id = u.id
  AND u.signup_source = 'touch'
  AND ai.provider_type = 'email';

INSERT INTO auth_identities (
    user_id,
    provider_type,
    provider_key,
    provider_subject,
    verified_at,
    metadata,
    created_at,
    updated_at
)
SELECT
    u.id,
    'touch',
    'touch',
    SUBSTRING(u.notes FROM 'touch_user_id=([^[:space:],;]+)'),
    NOW(),
    jsonb_build_object(
        'source', 'touch_migration',
        'email', LOWER(BTRIM(u.email))
    ),
    NOW(),
    NOW()
FROM users u
WHERE u.deleted_at IS NULL
  AND u.signup_source = 'touch'
  AND SUBSTRING(u.notes FROM 'touch_user_id=([^[:space:],;]+)') IS NOT NULL
  AND SUBSTRING(u.notes FROM 'touch_user_id=([^[:space:],;]+)') <> ''
ON CONFLICT (provider_type, provider_key, provider_subject) DO NOTHING;
