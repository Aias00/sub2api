import { describe, expect, it } from 'vitest'

import {
  buildUserOrdersStatusFilters,
  buildUserOrdersTableLabels,
  canUserOrderRequestRefund,
} from '../userOrdersRuntime'

describe('userOrdersRuntime', () => {
  const paymentText = (key: any) => `label:${key}`

  it('builds status filters and table labels from shell text', () => {
    expect(buildUserOrdersStatusFilters(paymentText)).toEqual([
      { value: '', label: 'label:all' },
      { value: 'PENDING', label: 'label:pending' },
      { value: 'COMPLETED', label: 'label:completed' },
      { value: 'FAILED', label: 'label:failed' },
      { value: 'REFUNDED', label: 'label:refunded' },
    ])
    expect(buildUserOrdersTableLabels(paymentText).orderId).toBe('label:orderId')
    expect(buildUserOrdersTableLabels(paymentText).statusRefundFailed).toBe('label:statusRefundFailed')
  })

  it('checks refund eligibility from status and provider allowlist', () => {
    expect(
      canUserOrderRequestRefund(
        { status: 'COMPLETED', provider_instance_id: 'stripe-main' } as any,
        new Set(['stripe-main']),
      ),
    ).toBe(true)
    expect(
      canUserOrderRequestRefund(
        { status: 'PENDING', provider_instance_id: 'stripe-main' } as any,
        new Set(['stripe-main']),
      ),
    ).toBe(false)
  })
})
