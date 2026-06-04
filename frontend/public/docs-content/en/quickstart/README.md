# Quickstart

This section is for users who already have a cloudbase account, available balance or subscription, and an API key, and now want to connect a specific client.

## Who This Is For

- You already signed in to cloudbase
- You already created an API key, or are about to create one
- You want to connect an existing tool to the shared gateway instead of maintaining separate upstream configs

## Before You Start

Before following a client-specific guide, confirm these three items:

1. Your account already has available balance or an active subscription
2. You created a dedicated API key for the client you are about to configure
3. You checked the `API Guide` page for the actual Base URL and recommended models on this site

## Recommended Reading Order

Pick the guide that matches the tool you are using:

- [Claude Code Quickstart](#/quickstart/claude-code)
- [Codex Quickstart](#/quickstart/codex)
- [Gemini CLI Quickstart](#/quickstart/gemini-cli)
- [OpenClaw Quickstart](#/quickstart/openclaw)
- [Hermes Quickstart](#/quickstart/hermes)
- [Cherry Studio Quickstart](#/quickstart/cherry-studio)
- [Claude Desktop Third-party Provider](#/quickstart/claude-desktop)
- [GPT-Image-2 Guide](#/quickstart/gpt-image-2)

## How To Choose The Right Guide

- `Claude Code / Codex / Gemini CLI`
  Best for terminal-based and local project workflows
- `Cherry Studio / Claude Desktop / OpenClaw / Hermes`
  Best for desktop clients, graphical usage, and multi-model switching
- `GPT-Image-2`
  Best for image generation and editing workflows

## What To Check After Setup

After completing any quickstart, run one minimal validation:

1. Send the simplest possible test request
2. Confirm the client is no longer using an older config
3. Return to cloudbase and verify:
   - the request appears in usage records
   - it hits the expected model or group
   - usage and cost start accumulating as expected

## If Something Breaks

Prepare the following before debugging:

- client name and version
- the API key you used for that client
- the Base URL shown on the `API Guide` page
- terminal output or screenshots of the error
- the approximate time of the request

Do not share the full API key.
