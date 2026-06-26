import type { UsageLog } from '@/types'

export function formatDashboardLocalDate(date: Date): string {
  return date.toISOString().split('T')[0]
}

export function resolveDashboardStartDate(dateRangeDays: number, now = new Date()): string {
  return formatDashboardLocalDate(new Date(now.getTime() - (dateRangeDays - 1) * 86400000))
}

export function resolveDashboardEndDate(now = new Date()): string {
  return formatDashboardLocalDate(now)
}

export function selectDashboardRecentUsage(items: UsageLog[], limit: number): UsageLog[] {
  return items.slice(0, limit)
}
