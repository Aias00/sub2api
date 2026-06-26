export const dashboardShellLabelKeys = [
  'balance',
  'available',
  'apiKeys',
  'active',
  'todayRequests',
  'total',
  'todayCost',
  'actual',
  'standard',
  'todayTokens',
  'totalTokens',
  'input',
  'output',
  'cacheWrite',
  'cacheRead',
  'performance',
  'avgResponse',
  'averageTime',
  'platformBreakdown',
  'platformCount',
  'platformOther',
  'requests',
  'tokens',
  'platformQuotaTitle',
  'platformQuotaDaily',
  'platformQuotaWeekly',
  'platformQuotaMonthly',
  'platformQuotaDisabled',
  'platformQuotaResetsAt',
  'recentUsage',
  'last7Days',
  'noUsageRecords',
  'startUsingApi',
  'viewAllUsage',
  'timeRange',
  'refresh',
  'granularity',
  'day',
  'hour',
  'modelDistribution',
  'noDataAvailable',
  'model',
  'quickActions',
  'createApiKey',
  'generateNewKey',
  'viewUsage',
  'checkDetailedLogs',
  'redeemCode',
  'addBalanceWithCode',
] as const

export type DashboardShellLabelKey = typeof dashboardShellLabelKeys[number]
export type DashboardShellLabels = Partial<Record<DashboardShellLabelKey, string>>
export type DashboardQuickActionDefaults = {
  createApiKeyPath: string
  usagePath: string
  redeemPath: string
}
export type DashboardShellConfig = {
  labels: DashboardShellLabels
  defaults: {
    quickActions: DashboardQuickActionDefaults
    dateRangeDays: number
    defaultGranularity: string
    recentUsageLimit: number
  }
}

export const DEFAULT_DASHBOARD_DATE_RANGE_DAYS = 7
export const DEFAULT_DASHBOARD_GRANULARITY = 'day'
export const DEFAULT_DASHBOARD_RECENT_USAGE_LIMIT = 5

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value && typeof value === 'object' && !Array.isArray(value))
}

function readPath(value: unknown, fallback: string): string {
  if (typeof value !== 'string') return fallback
  const trimmed = value.trim()
  if (!trimmed.startsWith('/') || trimmed.startsWith('//') || trimmed.includes('\\')) return fallback
  return trimmed
}

function readDashboardQuickActionDefaults(
  localized: Record<string, unknown>,
  fallback: DashboardQuickActionDefaults,
): DashboardQuickActionDefaults {
  const defaults = isRecord(localized.defaults) ? localized.defaults : {}
  const quickActions = isRecord(defaults.quickActions) ? defaults.quickActions : {}

  return {
    createApiKeyPath: readPath(quickActions.createApiKeyPath, fallback.createApiKeyPath),
    usagePath: readPath(quickActions.usagePath, fallback.usagePath),
    redeemPath: readPath(quickActions.redeemPath, fallback.redeemPath),
  }
}

function readPositiveInteger(value: unknown, fallback: number): number {
  const normalized = Number(value)
  if (!Number.isInteger(normalized) || normalized <= 0) return fallback
  return normalized
}

function readDashboardGranularity(value: unknown, fallback: string): string {
  if (typeof value !== 'string') return fallback
  const normalized = value.trim()
  return ['hour', 'day'].includes(normalized) ? normalized : fallback
}

function readDashboardLabels(localized: Record<string, unknown>): DashboardShellLabels {
  if (!isRecord(localized.labels)) return {}

  const labels: DashboardShellLabels = {}
  for (const labelKey of dashboardShellLabelKeys) {
    const value = localized.labels[labelKey]
    if (typeof value === 'string') {
      labels[labelKey] = value
    }
  }
  return labels
}

export function resolveDashboardShellConfig(
  raw: string | undefined,
  runtimeLocale: string,
  quickActionFallbacks: DashboardQuickActionDefaults,
): DashboardShellConfig {
  const fallbackConfig: DashboardShellConfig = {
    labels: {},
    defaults: {
      quickActions: quickActionFallbacks,
      dateRangeDays: DEFAULT_DASHBOARD_DATE_RANGE_DAYS,
      defaultGranularity: DEFAULT_DASHBOARD_GRANULARITY,
      recentUsageLimit: DEFAULT_DASHBOARD_RECENT_USAGE_LIMIT,
    },
  }
  if (!raw?.trim()) return fallbackConfig

  try {
    const parsed = JSON.parse(raw) as unknown
    if (!isRecord(parsed)) return fallbackConfig

    const normalizedLocale = runtimeLocale.toLowerCase()
    const localeKeys = [normalizedLocale, normalizedLocale.split('-')[0], 'en', 'zh']
    for (const localeKey of localeKeys) {
      const scoped = parsed[localeKey]
      if (!isRecord(scoped)) continue

      return {
        labels: readDashboardLabels(scoped),
        defaults: {
          quickActions: readDashboardQuickActionDefaults(scoped, quickActionFallbacks),
          dateRangeDays: readPositiveInteger(
            isRecord(scoped.defaults) ? scoped.defaults.dateRangeDays : undefined,
            DEFAULT_DASHBOARD_DATE_RANGE_DAYS,
          ),
          defaultGranularity: readDashboardGranularity(
            isRecord(scoped.defaults) ? scoped.defaults.defaultGranularity : undefined,
            DEFAULT_DASHBOARD_GRANULARITY,
          ),
          recentUsageLimit: readPositiveInteger(
            isRecord(scoped.defaults) ? scoped.defaults.recentUsageLimit : undefined,
            DEFAULT_DASHBOARD_RECENT_USAGE_LIMIT,
          ),
        },
      }
    }
  } catch {
    return fallbackConfig
  }

  return fallbackConfig
}

export function resolveDashboardShellLabels(raw: string | undefined, runtimeLocale: string): DashboardShellLabels {
  return resolveDashboardShellConfig(raw, runtimeLocale, {
    createApiKeyPath: '',
    usagePath: '',
    redeemPath: '',
  }).labels
}

export function interpolateDashboardLabel(label: string, params?: Record<string, string | number>): string {
  if (!params) return label
  return label.replace(/\{(\w+)\}/g, (_match, key: string) =>
    Object.prototype.hasOwnProperty.call(params, key) ? String(params[key]) : `{${key}}`,
  )
}
