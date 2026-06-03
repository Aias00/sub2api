# API Keys

API keys are one of the most important objects in cloudbase. In most cases, what you configure into a client is not your account password, but a key created in the console.

## When to create a new key

Prefer splitting keys by purpose instead of reusing one everywhere:

- one key for local development
- one key for automation
- one key for shared team tooling
- one key for temporary testing

Benefits:

- you can disable one bad key without touching everything
- usage attribution becomes clearer
- one client leak does not force a full rotation across all tools

## The most important fields when creating a key

### Name

Use a name that reflects the purpose, for example:

- `claude-code-local`
- `codex-repo-bot`
- `team-shared-cherry`

### Group

The selected group determines:

- which models this key can access
- which platform it uses
- what multiplier and quota rules apply

If you are not sure which one to pick, check [Available Groups](available-groups) first.

### Expiration / quota / IP restrictions

These fields are useful when:

- giving a temporary key to an external collaborator
- assigning a budget to an automation flow
- allowing usage only from a known machine or network

## What to do right after creating a key

After creating a key, go to:

1. **Gateway Guide** to copy the Base URL, auth header, and examples
2. **Available Groups** to confirm the capability scope behind the key
3. **Usage records** to verify that traffic is actually going through cloudbase

## Practical recommendations

### Do not reuse one key for too many tools

Otherwise it becomes hard to answer:

- which client is consuming the balance
- which tool is misconfigured
- which usage belongs to which workflow

### Save the key immediately after creation

Many systems only show the full key once. Do not rely on browser history to recover it later.

### Never commit keys into a repository

Avoid putting them into:

- source code
- screenshots
- shared documents

## What to check when something breaks

- whether the key is disabled
- whether the key is bound to the intended group
- whether the client is really using this key
- whether the Base URL comes from the current gateway guide
- whether the request is still going to an older provider
