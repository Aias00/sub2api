ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_signup_source_check;

UPDATE users
SET signup_source = 'email'
WHERE signup_source NOT IN ('email', 'github', 'google');

DROP INDEX IF EXISTS users_email_unique_active_non_touch;

CREATE UNIQUE INDEX IF NOT EXISTS users_email_unique_active
    ON users(email)
    WHERE deleted_at IS NULL;

ALTER TABLE users
    ADD CONSTRAINT users_signup_source_check
    CHECK (signup_source IN ('email', 'github', 'google'));

ALTER TABLE auth_identities
    DROP CONSTRAINT IF EXISTS auth_identities_provider_type_check;

DELETE FROM auth_identity_channels
WHERE provider_type NOT IN ('email', 'github', 'google');

DELETE FROM auth_identities
WHERE provider_type NOT IN ('email', 'github', 'google');

ALTER TABLE auth_identities
    ADD CONSTRAINT auth_identities_provider_type_check
    CHECK (provider_type IN ('email', 'github', 'google'));

ALTER TABLE auth_identity_channels
    DROP CONSTRAINT IF EXISTS auth_identity_channels_provider_type_check;

ALTER TABLE auth_identity_channels
    ADD CONSTRAINT auth_identity_channels_provider_type_check
    CHECK (provider_type IN ('email', 'github', 'google'));

ALTER TABLE pending_auth_sessions
    DROP CONSTRAINT IF EXISTS pending_auth_sessions_provider_type_check;

DELETE FROM pending_auth_sessions
WHERE provider_type NOT IN ('email', 'github', 'google');

ALTER TABLE pending_auth_sessions
    ADD CONSTRAINT pending_auth_sessions_provider_type_check
    CHECK (provider_type IN ('email', 'github', 'google'));

ALTER TABLE user_provider_default_grants
    DROP CONSTRAINT IF EXISTS user_provider_default_grants_provider_type_check;

DELETE FROM user_provider_default_grants
WHERE provider_type NOT IN ('email', 'github', 'google');

ALTER TABLE user_provider_default_grants
    ADD CONSTRAINT user_provider_default_grants_provider_type_check
    CHECK (provider_type IN ('email', 'github', 'google'));
