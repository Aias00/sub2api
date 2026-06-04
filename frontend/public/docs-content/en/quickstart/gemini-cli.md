# Gemini CLI Quickstart

If you want Gemini CLI to use cloudbase instead of a separate direct upstream setup, use the following flow.

## Best fit

This guide is useful when:

- you already use Gemini CLI
- you want usage, balance, and group control to stay inside cloudbase
- you want Gemini CLI to follow the same console workflow as Claude Code and Codex

## Before you start

- you have a cloudbase account
- your account has balance or an active subscription
- you have already created an API key
- your account actually has access to a Gemini-capable group

If you are not sure about the last point, check [Available Groups](#/console/available-groups) first.

## Setup steps

1. Sign in to the cloudbase console
2. Open **API Keys**
3. Choose or create a key for Gemini CLI
4. Open **API Guide**
5. Copy the Base URL, auth details, and model name shown there

A common environment-based setup looks like this:

```bash
export GEMINI_API_KEY="your cloudbase API key"
export GEMINI_BASE_URL="use the value shown in the console"
```

If your Gemini CLI distribution reads from a config file instead of these variables, copy the same values into the actual config location used by that client.

## First verification

Start with a low-cost prompt, for example:

```text
Explain this repository in one sentence.
```

If that works, move on to more complex tasks.

## Common issues

### Gemini models do not appear

This is usually not a CLI bug. More often:

- the account does not have access to a Gemini group
- the key is bound to a different group

Check [Available Groups](#/console/available-groups).

### Requests still go to an older endpoint

The client is probably reading an old config file or stale environment values.  
Restart the terminal and the client, then test again.

### Cost does not match expectation

Inspect:

- usage records
- key usage
- the multiplier of the group bound to that key
