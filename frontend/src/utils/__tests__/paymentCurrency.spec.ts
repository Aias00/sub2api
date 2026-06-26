import { describe, expect, it } from 'vitest'
import { formatOrderCreditedAmount, formatOrderPayAmount, formatPaymentCurrencyAmount, formatPublicMoneyAmount, resolvePaymentCurrencyPrefix } from '../paymentCurrency'
import type { PaymentOrder } from '@/types/payment'

const baseOrder: PaymentOrder = {
  id: 1,
  user_id: 2,
  amount: 12,
  pay_amount: 12,
  currency: 'CNY',
  fee_rate: 0,
  payment_type: 'alipay',
  out_trade_no: 'order-1',
  status: 'PENDING',
  order_type: 'balance',
  created_at: '2026-06-19T00:00:00Z',
  expires_at: '2026-06-19T01:00:00Z',
  refund_amount: 0,
}

describe('payment currency formatting', () => {
  it('formats known, unknown, and missing payment currencies', () => {
    expect(formatPaymentCurrencyAmount(12, 'CNY')).toBe('¥12.00')
    expect(formatPaymentCurrencyAmount(12, 'USD')).toBe('$12.00')
    expect(formatPaymentCurrencyAmount(12, 'KWD')).toBe('KWD 12.00')
    expect(formatPaymentCurrencyAmount(12, '')).toBe('12.00')
  })

  it('formats order pay amount from order currency', () => {
    expect(formatOrderPayAmount(baseOrder)).toBe('¥12.00')
    expect(formatOrderPayAmount({ ...baseOrder, currency: 'USD' })).toBe('$12.00')
  })

  it('does not apply fiat currency prefixes to balance credited amounts', () => {
    expect(formatOrderCreditedAmount(baseOrder)).toBe('12.00')
    expect(formatOrderCreditedAmount({ ...baseOrder, order_type: 'subscription' })).toBe('¥12.00')
  })

  it('resolves prefixes independently for inputs', () => {
    expect(resolvePaymentCurrencyPrefix('CNY')).toBe('¥')
    expect(resolvePaymentCurrencyPrefix('KWD')).toBe('KWD ')
    expect(resolvePaymentCurrencyPrefix(undefined)).toBe('')
  })

  it('formats public money from runtime settings prefix without local fallback', () => {
    expect(formatPublicMoneyAmount(12.345, '€', 2)).toBe('€12.35')
    expect(formatPublicMoneyAmount(12.345, '', 4)).toBe('12.3450')
    expect(formatPublicMoneyAmount(undefined, '$', 2)).toBe('$0.00')
    expect(formatPublicMoneyAmount(12.3, 'USD ', 2)).toBe('USD 12.30')
  })
})
