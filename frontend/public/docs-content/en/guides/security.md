# Account and Security

Account security settings are mainly managed on the profile page.

## Password Management

Recommendations:

- Use a dedicated password and do not reuse passwords from other sites.
- After changing your password, clear cached credentials from old clients.
- If you suspect account leakage, change the password first, then rotate API keys.

## Two-factor authentication (2FA)

After 2FA is enabled, sign-in requires an authenticator code in addition to the password.

### Before enabling it

- Confirm that the server has a stable `TOTP_ENCRYPTION_KEY` configured.
- Confirm that the authenticator app time is synchronized.
- Know how recovery or rebinding will be handled.

### If verification codes keep failing

Check these first:

- Whether phone time is synchronized automatically.
- Whether the server was restarted recently with a different TOTP encryption key.
- Whether the current binding is an invalid legacy record.

### Rebinding

If the old binding is invalid, the safest flow is usually:

1. Ask an administrator to reset the current account's 2FA status.
2. Re-enter the binding flow.
3. Scan the new QR code.

## Profile Maintenance

The profile page can usually maintain:

- Display name
- Email binding
- Avatar
- Balance notification email

If a third-party identity source is not enabled, its card is hidden instead of showing an unusable empty state.
