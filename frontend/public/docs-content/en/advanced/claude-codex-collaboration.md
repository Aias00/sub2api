# Claude Code and Codex Collaboration

Claude Code and Codex can both work on code, but they are useful in different parts of the workflow. cloudbase lets you route both through the same account while keeping usage visible.

## Suggested split

- Claude Code: long-context reasoning, requirement analysis, complex refactor discussion.
- Codex: local repository edits, test runs, CI fixes, and small scoped changes.

## Flow

1. Use one tool to analyze and plan.
2. Use another tool for a scoped implementation.
3. Return to the first tool for boundary review.
4. Run tests and inspect diffs before committing.

## Key management

Create separate keys such as `claude-code-local`, `codex-local`, and `vscode-helper` so usage records remain easy to read.
