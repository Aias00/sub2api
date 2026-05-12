# Gateway Calls

The unified gateway is the platform's core external capability.

## What the Gateway Guide provides

After you select an API key on the `Gateway Guide` page, the platform generates these values based on the group currently bound to that key:

- Unified gateway URL
- Recommended authentication header
- curl example
- Integration notes for the target model platform

## Minimal debugging loop

Check these items in order:

1. Whether the key is valid
2. Whether the key is bound to the correct group
3. Whether the group contains the target model or upstream channel
4. Whether balance, subscription quota, or rate limits were triggered

## Common headers

```http
Authorization: Bearer sk-xxxx
Content-Type: application/json
```

Some upstreams or gateway modes may require additional headers. Always follow the content generated for the current key on the Gateway Guide page.

## Recommended integration flow

### Start with curl

curl removes many variables from frontend SDKs, proxy layers, and CORS.

### Then integrate an SDK

After curl works, connect an OpenAI-compatible SDK, frontend app, or agent framework.

## Troubleshooting

### Model does not exist

- The current group does not include the model.
- Model naming does not match the platform mapping.
- The upstream account or channel is not active.

### Request succeeds but output is unexpected

- The actual channel hit is not the one you expected.
- Group multiplier, scheduling, or fallback model rules took effect.
- The upstream response was transformed by the compatibility layer.

### Intermittent failures

- Upstream channel instability
- Account rate limiting
- A route becoming temporarily unavailable

For these issues, check `Channel Status`, `Usage Records`, and administrator-side channel monitoring together.
