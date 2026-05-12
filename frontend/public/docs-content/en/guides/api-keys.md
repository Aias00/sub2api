# API Keys

An API key is the smallest authorization unit for client access to the platform.

## Creation Guidelines

- Use one key for one project or environment.
- Do not share the same production key across independent systems.
- Use meaningful names so log review and key rotation are easier later.

## Why group binding matters

What a key can actually call depends not only on the key itself, but also on the group bound to it.

Groups affect:

- Available model scope
- Routing strategy
- Price multiplier
- RPM and other rate-limit settings

## Rotation Strategy

Rotate keys proactively in these cases:

- A key may have leaked.
- A project is being handed over.
- A test environment changes.
- A third-party vendor changes.

Typical rotation flow:

1. Create a new key.
2. Gradually switch callers to the new key.
3. Confirm the new key is stable.
4. Revoke the old key.

## Security Recommendations

### Do not hardcode keys in frontend source code

Browser apps, public repositories, and mobile app packages are not suitable places for production keys.

### Use separate keys for separate callers

This lets you:

- Disable one problematic caller independently.
- Locate abnormal traffic sources quickly.
- Track real costs per project.

### Review usage records regularly

If request volume, token usage, or model distribution suddenly changes, first check whether the caller source changed.
