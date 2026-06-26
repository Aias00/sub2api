import { describe, expect, it, vi } from 'vitest'

import {
  asAffiliateOrderStatus,
  changeAffiliatePage,
  changeAffiliatePageSize,
  formatAffiliateCount,
  formatAffiliateNullableCurrency,
  formatAffiliatePaymentType,
} from '../affiliateRuntime'

describe('affiliateRuntime', () => {
  it('formats count, nullable currency, and payment type', () => {
    expect(formatAffiliateCount(1234)).toBe('1,234')
    expect(formatAffiliateNullableCurrency(12.5, (value) => `$${value.toFixed(2)}`)).toBe('$12.50')
    expect(formatAffiliateNullableCurrency(undefined, (value) => `$${value.toFixed(2)}`)).toBe('-')
    expect(formatAffiliatePaymentType('stripe', () => true, (key) => `translated:${key}`)).toBe('translated:payment.methods.stripe')
    expect(formatAffiliatePaymentType('unknown', () => false, (key) => key)).toBe('unknown')
  })

  it('passes through order status and pagination callbacks', () => {
    expect(asAffiliateOrderStatus('paid')).toBe('paid')

    const reload = vi.fn()
    let page = 0
    let size = 0
    let resetCount = 0
    changeAffiliatePage(2, (next) => { page = next }, reload)
    changeAffiliatePageSize(20, (next) => { size = next }, () => { resetCount += 1 }, reload)

    expect(page).toBe(2)
    expect(size).toBe(20)
    expect(resetCount).toBe(1)
    expect(reload).toHaveBeenCalledTimes(2)
  })
})
