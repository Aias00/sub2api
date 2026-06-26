import { describe, expect, it } from 'vitest'

import {
  buildUsageTableQueryParams,
  formatUsageLocalDate,
  resolveUsageDefaultDateRange,
} from '../usageRuntime'

describe('usageRuntime', () => {
  it('formats usage local dates and default date ranges', () => {
    const now = new Date(2026, 5, 20, 12, 0, 0)
    expect(formatUsageLocalDate(now)).toBe('2026-06-20')
    expect(resolveUsageDefaultDateRange(14, now)).toEqual({
      startDate: '2026-06-07',
      endDate: '2026-06-20',
    })
  })

  it('builds usage query params from filters and sort state', () => {
    expect(
      buildUsageTableQueryParams(
        2,
        50,
        { api_key_id: 9, start_date: '2026-06-01', end_date: '2026-06-20' },
        'created_at',
        'desc',
      ),
    ).toEqual({
      page: 2,
      page_size: 50,
      api_key_id: 9,
      start_date: '2026-06-01',
      end_date: '2026-06-20',
      sort_by: 'created_at',
      sort_order: 'desc',
    })
  })
})
