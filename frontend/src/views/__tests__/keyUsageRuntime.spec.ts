import { describe, expect, it } from 'vitest'

import {
  buildKeyUsageDateParams,
  formatKeyUsageResetTime,
  resolveKeyUsageStatusInfo,
} from '../keyUsageRuntime'

describe('keyUsageRuntime', () => {
  it('builds date params from configured range and custom range', () => {
    const now = new Date(Date.UTC(2026, 5, 20, 12, 0, 0))
    expect(buildKeyUsageDateParams({
      range: '7d',
      dailyUsageDays: 90,
      customStartDate: '',
      customEndDate: '',
      timezone: 'UTC',
      baseDate: now,
    })).toContain('start_date=2026-06-14')

    expect(buildKeyUsageDateParams({
      range: 'custom',
      dailyUsageDays: 30,
      customStartDate: '2026-06-01',
      customEndDate: '2026-06-20',
      timezone: 'UTC',
      baseDate: now,
    })).toContain('start_date=2026-06-01')
  })

  it('resolves status info and reset time text', () => {
    expect(resolveKeyUsageStatusInfo(
      { mode: 'quota_limited', isValid: true, status: 'active' },
      'Wallet',
      'Quota Mode',
    )).toEqual({
      label: 'Quota Mode',
      statusText: 'Active',
      isActive: true,
    })

    expect(formatKeyUsageResetTime(null, new Date(), 'Now')).toBe('')
  })
})
