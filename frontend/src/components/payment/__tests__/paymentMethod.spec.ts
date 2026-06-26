import { describe, expect, it } from 'vitest'
import {
  getStripePaymentMethodColor,
  normalizeStripePaymentMethod,
  resolveVisiblePaymentMethod,
} from '../paymentMethod'

describe('payment method normalization', () => {
  it('normalizes only explicit visible payment method data from Sub2API', () => {
    expect(resolveVisiblePaymentMethod('wxpay_direct')).toBe('wxpay')
    expect(resolveVisiblePaymentMethod('')).toBe('')
    expect(resolveVisiblePaymentMethod(null)).toBe('')
    expect(resolveVisiblePaymentMethod('unknown')).toBe('')
  })

  it('normalizes only explicit Stripe popup methods while keeping neutral unknown color', () => {
    expect(normalizeStripePaymentMethod('wechat_pay')).toBe('wechat_pay')
    expect(normalizeStripePaymentMethod('alipay')).toBe('alipay')
    expect(normalizeStripePaymentMethod('')).toBe('')
    expect(normalizeStripePaymentMethod('card')).toBe('')
    expect(getStripePaymentMethodColor('alipay')).toBe('#00AEEF')
    expect(getStripePaymentMethodColor('wechat_pay')).toBe('#07C160')
    expect(getStripePaymentMethodColor('card')).toBe('#635bff')
  })
})
