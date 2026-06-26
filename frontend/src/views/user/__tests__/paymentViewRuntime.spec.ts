import { describe, expect, it } from 'vitest'

import {
  buildPaymentResultRedirectQuery,
  buildWechatPaymentAuthorizeUrl,
  createEmptyPaymentRecoveryState,
} from '../paymentViewRuntime'

describe('paymentViewRuntime', () => {
  it('creates an empty payment recovery state and result redirect query', () => {
    expect(createEmptyPaymentRecoveryState()).toEqual({
      orderId: 0,
      amount: 0,
      qrCode: '',
      expiresAt: '',
      paymentType: '',
      payUrl: '',
      outTradeNo: '',
      clientSecret: '',
      intentId: '',
      currency: '',
      countryCode: '',
      paymentEnv: '',
      payAmount: 0,
      orderType: '',
      paymentMode: '',
      resumeToken: '',
      createdAt: 0,
    })
    expect(
      buildPaymentResultRedirectQuery({
        orderId: 123,
        outTradeNo: 'sub2_jsapi_123',
        resumeToken: 'resume-123',
      } as any),
    ).toEqual({
      order_id: '123',
      out_trade_no: 'sub2_jsapi_123',
      resume_token: 'resume-123',
    })
  })

  it('builds wechat authorize redirect with normalized payment context', () => {
    const url = buildWechatPaymentAuthorizeUrl(
      '/api/v1/auth/oauth/wechat/payment/start?payment_type=wxpay',
      '/purchase',
      {
        paymentType: 'wxpay_direct',
        orderType: 'subscription',
        planId: 7,
        orderAmount: 128,
      },
      'http://localhost',
    )

    expect(new URL(url, 'http://localhost').searchParams.get('redirect')).toBe(
      '/purchase?payment_type=wxpay&order_type=subscription&plan_id=7&amount=128',
    )
  })
})
