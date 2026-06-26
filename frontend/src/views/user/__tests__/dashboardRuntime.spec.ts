import { describe, expect, it } from 'vitest'

import {
  formatDashboardLocalDate,
  resolveDashboardEndDate,
  resolveDashboardStartDate,
  selectDashboardRecentUsage,
} from '../dashboardRuntime'

describe('dashboardRuntime', () => {
  it('formats dashboard dates and resolves default ranges', () => {
    const now = new Date('2026-06-25T12:00:00.000Z')
    expect(formatDashboardLocalDate(now)).toBe('2026-06-25')
    expect(resolveDashboardEndDate(now)).toBe('2026-06-25')
    expect(resolveDashboardStartDate(7, now)).toBe('2026-06-19')
  })

  it('selects recent usage by configured limit', () => {
    const items = [{ id: 1 }, { id: 2 }, { id: 3 }] as any[]
    expect(selectDashboardRecentUsage(items as any, 2)).toEqual([{ id: 1 }, { id: 2 }])
  })
})
