-- Keep signup grant admin queries aligned with the claims table timestamp
-- contract. Older installations had only created_at, while the admin API
-- exposes updated_at for list and summary views.

ALTER TABLE signup_grant_claims
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ;

UPDATE signup_grant_claims
SET updated_at = created_at
WHERE updated_at IS NULL;

ALTER TABLE signup_grant_claims
    ALTER COLUMN updated_at SET DEFAULT NOW(),
    ALTER COLUMN updated_at SET NOT NULL;
