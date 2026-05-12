# Billing and Subscriptions

Recharge and subscriptions are two parallel capabilities.

## Recharge

Recharge is better for pay-as-you-go usage.

The platform supports fixed recharge product cards configured by administrators:

- One configured product creates one visible card.
- A product can be marked as recommended.
- A product can include feature descriptions and credited amount.

## Subscriptions

Subscriptions are better when you want fixed quota, fixed periods, or plan-level constraints.

Whether subscription plans appear, and which plans appear, is also controlled by administrator configuration.

## Which one should you choose?

### Choose recharge

- Usage volume is unstable.
- You prefer on-demand payment.
- You do not need a recurring plan.

### Choose subscriptions

- Team budget is fixed.
- Monthly usage is stable.
- You want quota management through plans.

## Payment completed but credits did not arrive

Check these items first:

1. Whether the order is still pending or processing
2. Whether the payment webhook succeeded
3. Whether the page is still showing cached old state

Refresh `My Orders` and `Billing/Subscriptions`, then check whether balance or subscription status changed.

## Related administrator settings

Administrators can control:

- Minimum and maximum payment amount
- Visible payment methods
- Recharge product list
- Subscription plan list
- Order timeout

If the expected payment card is not visible, it is usually because the corresponding product or plan is not configured in the admin backend.
