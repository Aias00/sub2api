# FAQ

## What does the group multiplier mean?

The multiplier is part of the balance calculation policy. It can reflect upstream model price, channel cost, and service strategy.

```text
charged amount = base model cost × group multiplier
```

Actual groups and models depend on the console configuration.

## Where can I inspect actual charges?

Use **Usage Records** for request-level details. The dashboard is for summary trends; usage records are better for debugging individual requests.

## Which group should I choose?

Start with the default or recommended group. Switch only when you need a specific model, cost profile, or stability target.

## Can I create multiple API keys?

Yes. Use separate keys for local development, automation scripts, and shared clients. This makes usage tracking and key revocation easier.

## How is reliability handled?

cloudbase uses channels, groups, and usage visibility to reduce the impact of a single upstream issue. Your own clients should still use timeouts, retries, and fallbacks.

## How do I purchase credit or subscriptions?

Open [Purchase](https://cloudbase.eu.org/purchase). The page displays products configured and enabled by administrators.

## Are invoices supported?

Invoice support depends on the current site operation. Contact the administrator with your order number, amount, and invoice details.

## What should I provide when login or payment fails?

Provide your account email, page URL, operation time, browser console error, and payment error if available. Do not repeatedly submit payment orders before the order status is clear.
