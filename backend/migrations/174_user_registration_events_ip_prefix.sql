-- Add an IPv6-prefix-normalized column to user_registration_events so that
-- "same IP signup" aggregation survives IPv6 privacy-extension (RFC 4941)
-- interface-id rotation. ip_address keeps the full address for audit; ip_prefix
-- holds the /64 prefix for IPv6 (or the full IPv4 address) and is what aggregate
-- queries (admin insights / profile summary same-IP counts) match against.
--
-- The prefix length is configured by signup_grant_risk_control_ipv6_prefix_bits
-- (default 64). The backfill below uses a fixed /64 for IPv6 so historical rows
-- aggregate consistently with the default deployment; rows written after this
-- migration are populated by the application using the configured prefix length.
--
-- Backfill uses a PL/pgSQL DO block with per-row exception handling because
-- ip_address may contain empty strings or unparseable values that would abort a
-- set-based UPDATE with an inet cast error. host(network(set_masklen(v, 64)))
-- yields the compressed /64 prefix string (e.g. 2408:8215:5413:4700::).
-- IPv4 addresses are stored verbatim (no truncation). IPv4-mapped IPv6 (::ffff::)
-- is treated as IPv6 by family() and truncated to :: — this differs slightly from
-- the application path which passes mapped addresses through as IPv4; mapped
-- addresses are not expected in real registration IPs, so the impact is negligible.

ALTER TABLE user_registration_events
    ADD COLUMN IF NOT EXISTS ip_prefix TEXT NOT NULL DEFAULT '';

DO $$
DECLARE
  r RECORD;
  v inet;
BEGIN
  FOR r IN SELECT id, ip_address FROM user_registration_events WHERE ip_prefix = '' LOOP
    BEGIN
      v := r.ip_address::inet;
      IF family(v) = 6 THEN
        UPDATE user_registration_events SET ip_prefix = host(network(set_masklen(v, 64))) WHERE id = r.id;
      ELSE
        UPDATE user_registration_events SET ip_prefix = host(v) WHERE id = r.id;
      END IF;
    EXCEPTION WHEN OTHERS THEN
      -- unparseable ip_address: leave ip_prefix as ''
      NULL;
    END;
  END LOOP;
END $$;

CREATE INDEX IF NOT EXISTS idx_user_registration_events_prefix_created
    ON user_registration_events(ip_prefix, created_at DESC)
    WHERE ip_prefix <> '';
