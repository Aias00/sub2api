# Available Groups

The Available Groups page shows the groups your account can actually use right now.  
It is not a system-wide group list. It is the result of your current permissions, subscriptions, and admin grants.

## Why this page matters

Many users assume “missing models” means a client issue. In practice, the more common causes are:

- the account does not currently have access to that group
- the key is bound to a different group
- a subscription-based group is not active yet

Checking Available Groups first usually saves time.

## What kinds of groups appear here

### Public groups

These are typically:

- open to all users
- suitable as the default entry point
- used for common models or baseline access

### Exclusive / subscription groups

These usually come from:

- direct admin grants
- a subscription plan
- higher-tier or more specialized access

You only see the groups your current account can use. Hidden or unrelated groups are not exposed here.

## What a group controls

A group can affect:

- platform type, such as Claude, GPT, or Gemini
- rate multiplier
- quota limits
- certain extra capabilities, such as image-related support

## How this relates to API keys

When you create an API key, the key is bound to one of these groups.

That means:

- **Available Groups** tells you what you are allowed to choose
- **API Keys** tells you what you actually selected

If a key cannot access the model you expect, check:

1. which group the key is currently bound to
2. whether that group is still present in your available list

## When to revisit this page

- right after purchasing a subscription or credits
- after an admin says access has been granted
- when a tool suddenly stops showing a previously available model
- when you want different keys for different workflows

## A practical usage pattern

If you are unsure how to organize keys:

- bind one key to the safest public group
- bind another key to a higher-tier or subscription-only group

That keeps daily usage separate from expensive or specialized requests.
