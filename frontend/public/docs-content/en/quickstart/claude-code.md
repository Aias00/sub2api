# Claude Code Quickstart

This page explains how to connect Claude Code to cloudbase. Claude Code configuration may change between versions, so use official installation instructions and use the cloudbase console as the source of truth for endpoint values.

## Requirements

- cloudbase account.
- Available balance or subscription.
- API key.
- Node.js installed locally.

## Configure

1. Open the cloudbase console.
2. Go to **API Guide**.
3. Select the API key for Claude Code.
4. Copy Base URL, API key, and model name.
5. Put those values into the configuration required by your Claude Code version.

Common environment-variable shape:

```bash
export ANTHROPIC_BASE_URL="value from API Guide"
export ANTHROPIC_API_KEY="your cloudbase API key"
```

If your client uses an OpenAI-compatible interface, use the matching OpenAI Base URL and key fields.

## Verify

Ask a short question first. If it succeeds, continue with larger coding tasks.

## Troubleshooting

Authentication errors usually mean the API key is wrong or disabled. Model errors usually mean the selected group does not expose that model.
