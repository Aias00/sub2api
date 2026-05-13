-- Track welcome email delivery and opt-out state for non-transactional emails.
ALTER TABLE users
	ADD COLUMN IF NOT EXISTS welcome_email_sent_at TIMESTAMPTZ,
	ADD COLUMN IF NOT EXISTS marketing_emails_unsubscribed_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_users_marketing_emails_unsubscribed_at
	ON users (marketing_emails_unsubscribed_at)
	WHERE marketing_emails_unsubscribed_at IS NOT NULL;
