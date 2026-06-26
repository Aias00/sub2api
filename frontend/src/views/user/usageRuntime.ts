import type { UsageQueryParams } from '@/types'

export type UsageTableQueryParams = UsageQueryParams & {
  sort_by?: string
  sort_order?: 'asc' | 'desc'
}

export function formatUsageLocalDate(date: Date): string {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

export function resolveUsageDefaultDateRange(dateRangeDays: number, baseDate = new Date()) {
  const start = new Date(baseDate)
  start.setDate(start.getDate() - (dateRangeDays - 1))
  return {
    startDate: formatUsageLocalDate(start),
    endDate: formatUsageLocalDate(baseDate),
  }
}

export function buildUsageTableQueryParams(
  page: number,
  pageSize: number,
  filters: UsageQueryParams,
  sortBy: string,
  sortOrder: 'asc' | 'desc',
): UsageTableQueryParams {
  return {
    page,
    page_size: pageSize,
    ...filters,
    sort_by: sortBy,
    sort_order: sortOrder,
  }
}
