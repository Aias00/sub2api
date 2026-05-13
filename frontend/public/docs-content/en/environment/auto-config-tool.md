# Auto Configuration Tool

Most custom-provider clients need three values:

- Base URL from the cloudbase Gateway Guide.
- API Key from the API Keys page.
- Model name allowed by the selected group.

## Recommended flow

1. Log in to the [cloudbase dashboard](/dashboard).
2. Create or select an API key.
3. Open **Gateway Guide**.
4. Copy the generated Base URL and authentication values.
5. Paste them into your target client.
6. Send a small test request.

## Template

```text
Provider Name: cloudbase
Base URL: use the value shown in Gateway Guide
API Key: sk-********
Model: choose a model enabled by the current group
```

## When to reconfigure

Reconfigure when you rotate the key, change groups, upgrade a client that resets provider settings, or move to a new machine.
