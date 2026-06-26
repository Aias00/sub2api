import { describe, expect, it } from 'vitest'

import {
  formatRedeemHistoryValue,
  isRedeemAdminAdjustment,
  isRedeemBalanceType,
  isRedeemSubscriptionType,
  resolveRedeemHistoryItemTitle,
} from '../redeemRuntime'

describe('redeemRuntime', () => {
  const text = (key: any) => `label:${key}`
  const formatCurrency = (value: number) => `$${value.toFixed(2)}`

  it('classifies redeem record types', () => {
    expect(isRedeemBalanceType('balance')).toBe(true)
    expect(isRedeemSubscriptionType('subscription')).toBe(true)
    expect(isRedeemAdminAdjustment('admin_concurrency')).toBe(true)
  })

  it('resolves history titles and values', () => {
    expect(
      resolveRedeemHistoryItemTitle({ type: 'balance', value: 1 } as any, text),
    ).toBe('label:balanceAddedRedeem')
    expect(
      formatRedeemHistoryValue(
        { type: 'balance', value: 2 } as any,
        text,
        formatCurrency,
      ),
    ).toBe('+$2.00')
    expect(
      formatRedeemHistoryValue(
        { type: 'subscription', value: 30, validity_days: 30, group: { name: 'VIP' } } as any,
        text,
        formatCurrency,
      ),
    ).toContain('VIP')
  })
})
