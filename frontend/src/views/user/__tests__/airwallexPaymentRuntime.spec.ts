import { describe, expect, it } from 'vitest'

import { PAYMENT_RECOVERY_STORAGE_KEY } from '@/components/payment/paymentFlow'
import {
  buildAirwallexSuccessUrl,
  readAirwallexQueryString,
  restoreAirwallexPaymentSnapshot,
} from '../airwallexPaymentRuntime'

describe('airwallexPaymentRuntime', () => {
  it('reads query strings and builds success url', () => {
    const query = {
      order_id: '42',
      out_trade_no: 'trade-1',
      resume_token: 'resume-1',
    }
    expect(readAirwallexQueryString(query as any, 'order_id')).toBe('42')
    expect(
      buildAirwallexSuccessUrl(
        '/payment/result',
        query as any,
        {
          orderId: 99,
          outTradeNo: 'ignored',
          resumeToken: 'ignored',
        } as any,
        'https://example.com',
      ),
    ).toBe('https://example.com/payment/result?order_id=42&out_trade_no=trade-1&resume_token=resume-1')
  })

  it('restores only valid airwallex snapshots', () => {
    const storage = window.localStorage
    storage.setItem(PAYMENT_RECOVERY_STORAGE_KEY, JSON.stringify({
      orderId: 42,
      amount: 88,
      qrCode: '',
      expiresAt: '2099-01-01T00:10:00.000Z',
      paymentType: 'airwallex',
      payUrl: 'https://pay.example.com/session/42',
      outTradeNo: 'trade-1',
      clientSecret: 'secret',
      intentId: 'intent',
      currency: 'USD',
      countryCode: 'US',
      paymentEnv: 'demo',
      payAmount: 88,
      orderType: 'balance',
      paymentMode: 'redirect',
      resumeToken: 'resume-1',
      createdAt: Date.UTC(2099, 0, 1, 0, 0, 0),
    }))

    expect(
      restoreAirwallexPaymentSnapshot(storage, {
        order_id: '42',
        out_trade_no: 'trade-1',
        resume_token: 'resume-1',
      } as any),
    ).not.toBeNull()
  })
})
