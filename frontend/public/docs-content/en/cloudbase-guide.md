# cloudbase Guide

This guide is for users who are new to AI coding tools and API gateways. Follow the account, credit, API key, and tool configuration steps first; you can learn the deeper concepts later.

## Why cloudbase exists

Claude Code, Codex, Cherry Studio, Claude Desktop, and similar tools can all connect to large models. Without a gateway, you often have to manage accounts, keys, network settings, credit, and billing separately.

cloudbase centralizes that work:

- Manage API keys in one console.
- Track balance and spend in one place.
- Configure model channels and groups centrally.
- Provide one gateway endpoint for multiple tools.

## What the tools are

### ChatGPT and Claude

Conversation products for Q&A, writing, analysis, and code assistance.

### Codex and Claude Code

Developer workflow tools that can inspect files, run commands, edit code, and iterate inside a project.

### Cherry Studio and Claude Desktop

Desktop clients for managing models, conversations, knowledge, and tool integrations.

## What is a terminal?

A terminal is where you run commands. When a guide shows `node -v`, `npm -v`, or `codex`, run it in Terminal, PowerShell, or another shell.

## What is an API?

An API is a programmatic entry point. In cloudbase, your client usually needs:

- Base URL from the API Guide page.
- API key from the API Keys page.
- Model name allowed by the selected group.

## Start using cloudbase

1. Register from [Register](/register).
2. Add credit or subscribe from [Purchase](/purchase).
3. Create an API key from **API Keys**.
4. Open **API Guide** and copy the generated values.
5. Configure your client and send a small test request.

## Troubleshooting

Check these before retrying expensive requests:

1. API key was copied completely.
2. Base URL comes from the current cloudbase console.
3. Balance or subscription is available.
4. The selected group exposes the model.
5. Local proxy or network settings are not blocking the request.
6. Logs show whether the status is `401`, `403`, `404`, `429`, or `5xx`.

## Security

Never commit API keys, never share full keys in screenshots, and create separate keys for separate tools. Revoke and rotate a key immediately if it may have leaked.
