-- Add a non-sequential public user identifier for API/UI exposure.
-- The internal numeric id remains the primary key for relations and auth.

ALTER TABLE users
  ADD COLUMN IF NOT EXISTS public_id VARCHAR(32);

WITH users_to_fill AS (
  SELECT
    id,
    GREATEST(
      0,
      (EXTRACT(EPOCH FROM (COALESCE(created_at, NOW()) - TIMESTAMPTZ '2024-01-01 00:00:00+00')) * 1000)::BIGINT
    ) AS base_millis
  FROM users
  WHERE public_id IS NULL OR public_id = ''
),
ordered_users AS (
  SELECT
    id,
    base_millis,
    ROW_NUMBER() OVER (PARTITION BY base_millis ORDER BY id) - 1 AS seq_index
  FROM users_to_fill
),
generated_users AS (
  SELECT
    id,
    'u_' || (
      (((base_millis + (seq_index >> 12)) << 22)
        | (1::BIGINT << 12)
        | (seq_index & 4095))
    )::TEXT AS generated_public_id
  FROM ordered_users
)
UPDATE users AS u
SET public_id = generated_users.generated_public_id
FROM generated_users
WHERE u.id = generated_users.id;

UPDATE users
SET public_id = 'u_' || public_id
WHERE public_id IS NOT NULL
  AND public_id <> ''
  AND public_id !~ '^u_';

ALTER TABLE users
  ALTER COLUMN public_id SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS users_public_id_key
  ON users (public_id);
