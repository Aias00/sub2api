# Recommended Global Rules

Use this template in global instructions, project `AGENTS.md`, or a similar configuration supported by your AI coding tool.

## Purpose

Rules should make the tool inspect context, keep changes small, and verify results. They should not be a long list of temporary task details.

## Template

```text
You are working in an existing codebase.

Before editing:
- Inspect the relevant files and existing patterns.
- State a short plan for non-trivial changes.
- Prefer small, reversible diffs.

While editing:
- Do not rewrite unrelated code.
- Do not expose secrets or API keys.
- Reuse existing utilities and conventions.
- Keep behavior unchanged unless the task explicitly asks for behavior changes.

Before finishing:
- Run the smallest meaningful verification.
- Report changed files, verification result, and remaining risks.
- If verification cannot run, explain the blocker clearly.
```

## cloudbase add-on

```text
For cloudbase integrations:
- Use the console API Guide as the source of truth for Base URL and models.
- Never paste full API keys into issue comments, screenshots, or commits.
- Use separate keys for separate tools.
- Check usage records when debugging unexpected cost or failures.
```
