# Introduction

Welcome to **cloudbase**. It is a unified workspace for AI development tools such as Claude Code, Codex, Gemini CLI, Cherry Studio, and Claude Desktop.

Console entry: [Home](/home) / [Dashboard](/dashboard)

## What is cloudbase?

cloudbase is not just “another proxy URL.” It is meant to centralize the parts developers repeatedly need:

- one account for balance, subscriptions, orders, and API keys
- one gateway entry for multiple clients and workflows
- group-based control over models, channels, pricing, and permissions
- dashboard and usage pages for cost, model usage, and failures

If you currently switch between multiple clients, maintain scattered keys, or need clearer cost visibility, cloudbase works more like a shared AI access workspace.

## Who it is for

- **Solo developers** who want one place to manage Claude Code, Codex, and Gemini CLI access
- **Small teams** that need groups, subscriptions, permissions, and cost visibility
- **Integrators** who want one stable gateway instead of wiring every tool to separate upstreams

## Core capabilities

### Business capability entries

The homepage highlights product-facing workflows such as WeChat export, hot topic tracking, image prompts, and the image workspace. These pages cover content collection, topic discovery, prompt reuse, and image production.

If you want to understand what each page does and how they connect into a workflow, start with the [Business Capabilities Guide](business-capabilities).

### Unified API keys

Create API keys in the console and reuse them across tools. A practical pattern is:

- one key for local development
- one key for automation
- one key for shared team tooling

That makes revocation, usage review, and cost attribution much easier.

### Group-driven model access

cloudbase does not expose every model to every user by default. Groups define:

- which models a key can access
- which platform capabilities are available
- what rate multiplier and quota rules apply

Regular users can inspect their current **available groups** directly in the console.

### API Guide and quick setup

The **API Guide** page generates the active:

- Base URL
- auth header
- suggested model
- usage examples

Always prefer the values shown in the console over older copied snippets.

### Balance, subscriptions, and orders

cloudbase supports both balance products and subscription plans. What a user can buy depends on the current site configuration.

### Usage and cost visibility

Use the dashboard, usage records, and key usage pages to inspect:

- token trends
- model consumption
- failed requests
- balance changes

## Recommended onboarding path

### 1. Create an account

Open [Register](/register) and create an account with email or an enabled third-party sign-in method.

### 2. Add credit or subscribe

Open [Purchase](/purchase) and select one of the currently available products or plans.

### 3. Check your available groups

Before configuring tools, open **Available Groups** and confirm what your account can actually use.
This helps explain why some models appear and others do not.

### 4. Create an API key

Open **API Keys** and create a key for the tool or workflow you want to connect.

### 5. Open the API Guide

Use the selected key’s gateway instructions as the source of truth for client configuration.

### 6. Pick the matching quickstart

- [Claude Code Quickstart](quickstart/claude-code)
- [Codex Quickstart](quickstart/codex)
- [Gemini CLI Quickstart](quickstart/gemini-cli)
- [OpenClaw Quickstart](quickstart/openclaw)
- [Hermes Quickstart](quickstart/hermes)
- [Cherry Studio Quickstart](quickstart/cherry-studio)
- [Claude Desktop Third-party Provider](quickstart/claude-desktop)
- [GPT-Image-2 Guide](quickstart/gpt-image-2)

### 7. Use business capabilities

If your goal is content collection, topic planning, prompt reuse, or image production instead of API integration, start with the [Business Capabilities Guide](business-capabilities).

## Console concepts worth learning first

- [Business Capabilities](business-capabilities)
- [API Keys](console/api-keys)
- [Available Groups](console/available-groups)
- [Model Plaza](console/models-plaza)

## Getting help

When asking for help, prepare:

- the tool name and version
- the Base URL shown in the console
- terminal logs or screenshots
- the model name you tried
- the approximate request time

Never expose a full API key in messages or screenshots.
