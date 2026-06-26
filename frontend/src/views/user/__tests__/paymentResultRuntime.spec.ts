import { describe, expect, it } from 'vitest'

import { PAYMENT_RECOVERY_STORAGE_KEY } from '@/components/payment/paymentFlow'
import {
  applyResolvedPaymentOrder,
  calculatePaymentBaseAmount,
  calculatePaymentFeeAmount,
  clearPaymentRecoverySnapshotForTerminalStatus,
  isPaymentResultPending,
  isPaymentResultSuccess,
  normalizePaymentResultStatus,
  readPaymentResultQueryString,
  restorePaymentRecoverySnapshot,
} from '../paymentResultRuntime'

describe('paymentResultRuntime', () => {
  it('calculates payment base and fee amounts', () => {
    expect(calculatePaymentBaseAmount({ pay_amount: 103, fee_rate: 3 } as any)).toBe(100)
    expect(calculatePaymentFeeAmount({ pay_amount: 103, fee_rate: 3 } as any)).toBe(3)
  })

  it('normalizes statuses and resolves success/pending flags', () => {
    expect(normalizePaymentResultStatus(' paid ')).toBe('PAID')
    expect(isPaymentResultSuccess('paid')).toBe(true)
    expect(isPaymentResultPending('processing')).toBe(true)
  })

  it('reads query values, applies resolved order currency, and restores snapshots', () => {
    expect(readPaymentResultQueryString({ order_id: ['42'] } as any, 'order_id')).toBe('42')
    expect(applyResolvedPaymentOrder({ currency: 'usd' } as any, '')).toEqual({
      order: { currency: 'usd' },
      currency: 'USD',
    })

    const storage = window.localStorage
    storage.setItem(PAYMENT_RECOVERY_STORAGE_KEY, JSON.stringify({
      orderId: 42,
      amount: 88,
      qrCode: '',
      expiresAt: '2099-01-01T00:10:00.000Z',
      paymentType: 'alipay',
      payUrl: 'https://pay.example.com/session/42',
      outTradeNo: 'order-no',
      clientSecret: '',
      intentId: '',
      currency: '',
      countryCode: '',
      paymentEnv: '',
      payAmount: 88,
      orderType: 'balance',
      paymentMode: 'popup',
      resumeToken: 'resume-token',
      createdAt: Date.UTC(2099, 0, 1, 0, 0, 0),
    }))
    expect(restorePaymentRecoverySnapshot(storage, {
      resumeToken: 'resume-token',
      routeOrderId: 0,
      routeOutTradeNo: '',
    })).not.toBeNull()

    clearPaymentRecoverySnapshotForTerminalStatus(storage, 'PAID')
    expect(storage.getItem(PAYMENT_RECOVERY_STORAGE_KEY)).toBeNull()
  })
})
