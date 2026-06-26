import { describe, expect, it } from 'vitest'
import {
  formatPaymentAmount,
  normalizePaymentCountryCode,
  normalizePaymentCurrency,
} from '../currency'

describe('payment currency normalization', () => {
  it('normalizes only explicit payment currency data from Sub2API', () => {
    expect(normalizePaymentCurrency(' usd ')).toBe('USD')
    expect(normalizePaymentCurrency('')).toBe('')
    expect(normalizePaymentCurrency(null)).toBe('')
    expect(normalizePaymentCurrency('invalid')).toBe('')
  })

  it('normalizes only explicit payment country data from Sub2API', () => {
    expect(normalizePaymentCountryCode(' hk ')).toBe('HK')
    expect(normalizePaymentCountryCode('')).toBe('')
    expect(normalizePaymentCountryCode(null)).toBe('')
    expect(normalizePaymentCountryCode('invalid')).toBe('')
  })
})

describe('formatPaymentAmount', () => {
  it('does not synthesize a currency prefix when currency data is missing', () => {
    expect(formatPaymentAmount(100, '', 'en-US')).toBe('100.00')
    expect(formatPaymentAmount(100, null, 'en-US')).toBe('100.00')
    expect(formatPaymentAmount(100, 'invalid', 'en-US')).toBe('100.00')
  })

  it('uses the currency default fraction digits', () => {
    expect(formatPaymentAmount(100, 'JPY', 'en-US')).not.toContain('.00')
    expect(formatPaymentAmount(100, 'KRW', 'en-US')).not.toContain('.00')
    expect(formatPaymentAmount(100, 'HKD', 'en-US')).toContain('.00')
  })
})
