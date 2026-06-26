import { describe, expect, it } from 'vitest'

import {
  formatSubscriptionDailyUsageWindow,
  formatSubscriptionExpirationDate,
  resolveSubscriptionExpirationClass,
  resolveSubscriptionProgressBarClass,
  resolveSubscriptionProgressWidth,
  resolveSubscriptionStatusText,
} from '../subscriptionsRuntime'

describe('subscriptionsRuntime', () => {
  const text = (key: any, params?: Record<string, string | number>) =>
    params?.days !== undefined
      ? `${key}:${params.days}`
      : params?.time !== undefined
        ? `${key}:${params.time}`
        : `label:${key}`

  it('resolves status and progress presentation', () => {
    expect(resolveSubscriptionStatusText('active' as any, text)).toBe('label:subscriptionStatusActive')
    expect(resolveSubscriptionProgressWidth(9, 10)).toBe('90%')
    expect(resolveSubscriptionProgressBarClass(9, 10)).toContain('red')
  })

  it('formats expiration and quota window text', () => {
    const future = new Date(Date.now() + 2 * 24 * 60 * 60 * 1000).toISOString()
    expect(formatSubscriptionExpirationDate(future, text)).toContain('subscriptionDaysRemaining')
    expect(resolveSubscriptionExpirationClass(future)).toBe('text-red-600 dark:text-red-400')

    expect(
      formatSubscriptionDailyUsageWindow(
        {
          subscription_type: 'subscription',
          daily_window_start: new Date(Date.now() - 1 * 60 * 60 * 1000).toISOString(),
        } as any,
        text,
      ),
    ).toContain('subscriptionResetIn')
  })
})
