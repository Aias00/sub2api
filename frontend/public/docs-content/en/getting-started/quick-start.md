# Quick Start

If you only want to make the first successful request, follow this sequence.

## Step 1: Sign in and check account status

- Confirm that you can sign in normally.
- If the administrator enabled email verification or 2FA, finish the security check first.
- After signing in, check your dashboard balance and subscription status so quota issues do not interrupt debugging later.

## Step 2: Create an API key

Open the `API Keys` page and create a new key.

Recommendations:

- Use one key per project or environment so it is easy to trace and revoke.
- Include the environment or purpose in the name, such as `web-prod` or `agent-dev`.

## Step 3: Confirm group binding

The platform's available capabilities depend on the group currently bound to the key.

- If you cannot see the expected model, check whether the group is correct.
- If the administrator limits selectable groups, regular users can only choose from the allowed scope.

## Step 4: Copy gateway call parameters

Open the `Gateway Guide` page and select the key you just created.

You will see:

- Gateway URL
- Authentication header
- curl example
- Call snippets generated for the current group

## Step 5: Send the first request

Use curl or Postman first, then integrate it into production code.

```bash
curl https://your-domain.example/v1/chat/completions \
  -H "Authorization: Bearer sk-xxx" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [
      {"role": "user", "content": "Hello"}
    ]
  }'
```

## What to check first when a request fails

### 401 or 403

- The key is invalid.
- The key is disabled.
- The current group does not have permission to call the target model.

### Balance or plan errors

- Balance is insufficient.
- Subscription quota is exhausted.
- Payment has not been confirmed yet.

### Model or routing errors

- The current group is not bound to the target model.
- The upstream channel is unhealthy.
- The group routing strategy differs from expectation.
