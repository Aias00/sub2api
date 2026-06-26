import { describe, expect, it } from 'vitest'

import { PAYMENT_RECOVERY_STORAGE_KEY } from '@/components/payment/paymentFlow'
import {
  buildStripePaymentResultReturnURL,
  buildStripePaymentResultRoute,
  formatStripeGatewayAmount,
  resolveStripePaymentRouteState,
  restoreStripePaymentCurrency,
} from '../stripePaymentRuntime'

describe('stripePaymentRuntime', () => {
  it('resolves route state and formats gateway amount', () => {
    expect(resolveStripePaymentRouteState({
      order_id: '42',
      client_secret: 'secret',
      method: 'wechat_pay',
      resume_token: 'resume-1',
    } as any)).toEqual({
      orderId: 42,
      clientSecret: 'secret',
      method: 'wechat_pay',
      resumeToken: 'resume-1',
    })
    expect(formatStripeGatewayAmount(103, 'HKD', 'zh-CN')).toBe('$103.00')
  })

  it('restores currency and builds success routes', () => {
    const storage = window.localStorage
    storage.setItem(PAYMENT_RECOVERY_STORAGE_KEY, JSON.stringify({
      orderId: 42,
      amount: 88,
      qrCode: '',
      expiresAt: '2099-01-01T00:10:00.000Z',
      paymentType: 'stripe',
      payUrl: 'https://pay.example.com/session/42',
      outTradeNo: 'trade-1',
      clientSecret: 'secret',
      intentId: 'intent',
      currency: 'HKD',
      countryCode: 'HK',
      paymentEnv: 'demo',
      payAmount: 88,
      orderType: 'balance',
      paymentMode: 'redirect',
      resumeToken: 'resume-1',
      createdAt: Date.UTC(2099, 0, 1, 0, 0, 0),
    }))

    expect(restoreStripePaymentCurrency(storage, 'resume-1', 42)).toBe('HKD')
    expect(buildStripePaymentResultRoute('/payment/result', '42')).toEqual({
      path: '/payment/result',
      query: { order_id: '42', status: 'success' },
    })
    expect(
      buildStripePaymentResultReturnURL('/payment/result', '42', 'https://example.com'),
    ).toBe('https://example.com/payment/result?order_id=42&status=success')
  })
})
