import { describe, expect, it } from 'vitest'

import {
  formatPaymentQrCountdown,
  isPaymentQrCompleted,
  isPaymentQrTerminal,
  resolvePaymentQrRouteState,
  resolvePaymentQrSecondsUntil,
} from '../paymentQrRuntime'

describe('paymentQrRuntime', () => {
  it('resolves route state and countdown formatting', () => {
    expect(resolvePaymentQrRouteState({
      order_id: '7',
      qr: 'weixin://pay/test',
      pay_url: 'https://pay.example.com',
      payment_type: 'wxpay',
      expires_at: '2099-01-01T00:10:00.000Z',
    } as any)).toEqual({
      orderId: 7,
      qrUrl: 'weixin://pay/test',
      payUrl: 'https://pay.example.com',
      paymentType: 'wxpay',
      expiresAt: '2099-01-01T00:10:00.000Z',
    })
    expect(formatPaymentQrCountdown(65)).toBe('01:05')
  })

  it('computes seconds until and terminal/completed states', () => {
    const now = Date.parse('2099-01-01T00:00:00.000Z')
    expect(resolvePaymentQrSecondsUntil('2099-01-01T00:10:00.000Z', now)).toBe(600)
    expect(isPaymentQrCompleted('PAID')).toBe(true)
    expect(isPaymentQrTerminal('FAILED')).toBe(true)
  })
})
