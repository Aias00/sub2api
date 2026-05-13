# 推荐全局规则

下面是一份适合 AI 编码工具的全局规则模板。你可以放到工具支持的全局指令、项目 `AGENTS.md` 或类似配置中，再按团队习惯调整。

## 设计思想

规则不是为了限制 AI 输出，而是为了让它在修改代码前先理解上下文、控制改动范围，并在完成后给出可验证结果。

## 推荐规则

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

## 适合 cloudbase 的补充规则

```text
For cloudbase integrations:
- Use the console Gateway Guide as the source of truth for Base URL and models.
- Never paste full API keys into issue comments, screenshots, or commits.
- Use separate keys for separate tools.
- Check usage records when debugging unexpected cost or failures.
```

## 使用建议

全局规则保持短而稳定，项目规则再补充具体命令、目录和测试方式。不要把临时需求写进全局规则，否则后续任务容易被旧指令干扰。
