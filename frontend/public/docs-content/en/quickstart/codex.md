# Codex Quickstart

Codex is useful for repository reading, code edits, test runs, and pre-commit checks. With cloudbase, you can keep usage and billing centralized.

## Requirements

- Codex or a Codex-compatible client installed.
- cloudbase API key.
- Base URL copied from **Gateway Guide**.

## Configure

```text
Base URL: value from cloudbase Gateway Guide
API Key: your cloudbase API key
Model: model enabled by the current group
```

Store secrets in private local configuration or environment variables. Do not commit keys to the repository.

## Usage tips

- Validate with a small project first.
- Ask Codex to inspect the project and propose a plan before large edits.
- Confirm commands before running destructive operations.
- Ask for test results and changed files at the end.
