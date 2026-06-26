import { describe, expect, it } from 'vitest'

import {
  buildStripePopupPaymentResultReturnUrl,
  formatStripePopupDisplayAmount,
  resolveStripePopupRouteState,
} from '../stripePopupRuntime'

describe('stripePopupRuntime', () => {
  it('normalizes route state safely', () => {
    expect(
      resolveStripePopupRouteState({
        order_id: 42,
        method: 'wechat_pay',
        amount: '10.5',
        currency: 'usd',
      }),
    ).toEqual({
      orderId: '42',
      method: 'wechat_pay',
      amount: '10.5',
      currency: 'USD',
    })
  })

  it('formats display amount and return url', () => {
    expect(formatStripePopupDisplayAmount('10.5', 'USD')).toBe('$10.50')
    expect(
      buildStripePopupPaymentResultReturnUrl('/payment/result', '42', 'https://example.com'),
    ).toBe('https://example.com/payment/result?order_id=42&status=success')
  })
})
