ALTER TABLE users
    ADD COLUMN IF NOT EXISTS login_agreement_accepted_revision text NOT NULL DEFAULT '';

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS login_agreement_accepted_at timestamptz NULL;
