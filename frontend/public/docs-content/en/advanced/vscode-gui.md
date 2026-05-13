# VS Code GUI Workflow

VS Code works well with Codex, Claude Code, and other AI coding tools. Its GUI makes file browsing, diff review, and test execution easier.

## Recommended workflow

1. Open the project in VS Code.
2. Start the AI tool from the integrated terminal.
3. Ask the tool to inspect the project and propose a plan.
4. Review diffs in Source Control.
5. Run tests or builds.
6. Commit only after verification.

## Review checklist

- No unrelated files changed.
- No API keys or secrets leaked.
- Required behavior was not removed.
- UI-only changes did not break interaction.

Ask the AI to handle one clear problem at a time.
