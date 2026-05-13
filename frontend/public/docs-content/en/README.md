# cloudbase Guide

cloudbase is a unified gateway for AI development tools. It lets you connect Claude Code, Codex, Cherry Studio, Claude Desktop, and similar clients through one platform account, one billing system, and managed API keys.

Console entry: [Home](/home) / [Dashboard](/dashboard)

## What is cloudbase?

cloudbase puts multiple upstream AI services behind a controlled gateway.

- Manage balance, subscriptions, orders, and API keys in one account.
- Use one gateway endpoint for multiple developer tools.
- Route models through groups with different channel and pricing policies.
- Track requests, token usage, spend, and latency from the dashboard.

## Core Features

### Unified API Keys

Create an API key in the console and use it in Claude Code, Codex, Cherry Studio, Claude Desktop, or any client that supports a custom provider.

### Gateway Guide

The Gateway Guide page generates the Base URL, authentication header, and examples for the selected key and group. Prefer the values shown in the console over hard-coded examples.

### Balance and Subscriptions

The purchase page shows only products configured and enabled by administrators.

### Usage Visibility

Use the dashboard and usage records to understand cost, model usage, and failed requests.

## Quick Start

### 1. Create an account

Open [Register](/register) and create an account with email or an enabled third-party login provider.

### 2. Add credit or subscribe

Open [Purchase](/purchase) and select an available product.

### 3. Create an API key

Open **API Keys** and create a key. Use separate keys for separate tools whenever possible.

### 4. Configure your tool

Choose the matching quickstart:

- [Claude Code Quickstart](quickstart/claude-code)
- [Codex Quickstart](quickstart/codex)
- [Cherry Studio Quickstart](quickstart/cherry-studio)
- [Claude Desktop Third-party Provider](quickstart/claude-desktop)

## Getting Help

When asking for help, provide the tool name, tool version, Base URL shown in the console, error logs, model name, and request time. Never expose a full API key in screenshots or messages.
