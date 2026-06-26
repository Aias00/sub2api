export type KeyUsageDateRangeKey = 'today' | '7d' | '30d' | 'custom'

export function formatKeyUsageDate(date: Date): string {
  return date.toISOString().split('T')[0]
}

export function buildKeyUsageDateParams(input: {
  range: KeyUsageDateRangeKey
  dailyUsageDays: number
  customStartDate: string
  customEndDate: string
  timezone: string
  baseDate?: Date
}): string {
  const now = input.baseDate ?? new Date()
  const params = new URLSearchParams()

  if (input.range === 'custom') {
    if (input.customStartDate && input.customEndDate) {
      params.set('start_date', input.customStartDate)
      params.set('end_date', input.customEndDate)
    }
  } else {
    const end = formatKeyUsageDate(now)
    let start: string
    switch (input.range) {
      case 'today':
        start = end
        break
      case '7d':
        start = formatKeyUsageDate(new Date(now.getTime() - (7 - 1) * 86400000))
        break
      case '30d':
        start = formatKeyUsageDate(new Date(now.getTime() - (30 - 1) * 86400000))
        break
      default:
        start = formatKeyUsageDate(new Date(now.getTime() - (input.dailyUsageDays - 1) * 86400000))
    }
    params.set('start_date', start)
    params.set('end_date', end)
  }

  params.set('days', String(input.dailyUsageDays))
  params.set('timezone', input.timezone)
  return params.toString()
}

export function resolveKeyUsageStatusInfo(
  data: any,
  walletBalanceLabel: string,
  quotaModeLabel: string,
) {
  if (!data) return null

  if (data.mode === 'quota_limited') {
    const isValid = data.isValid !== false
    const statusMap: Record<string, string> = {
      active: 'Active',
      quota_exhausted: 'Quota Exhausted',
      expired: 'Expired',
    }
    return {
      label: quotaModeLabel,
      statusText: statusMap[data.status] || data.status || 'Unknown',
      isActive: isValid && data.status === 'active',
    }
  }

  return {
    label: data.planName || walletBalanceLabel,
    statusText: 'Active',
    isActive: true,
  }
}

export function formatKeyUsageResetTime(
  resetAt: string | null | undefined,
  now: Date,
  resetNowLabel: string,
): string {
  if (!resetAt) return ''
  const diff = new Date(resetAt).getTime() - now.getTime()
  if (diff <= 0) return resetNowLabel
  const days = Math.floor(diff / 86400000)
  const hours = Math.floor((diff % 86400000) / 3600000)
  const mins = Math.floor((diff % 3600000) / 60000)
  if (days > 0) return `${days}d ${hours}h`
  if (hours > 0) return `${hours}h ${mins}m`
  return `${mins}m`
}
