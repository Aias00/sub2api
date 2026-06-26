import { resolveLocalizedShellLabels } from '@/utils/localizedShell'

export const keyUsageShellLabelKeys = [
  'apply',
  'allRightsReserved',
  'avgDuration',
  'cacheCreationTokens',
  'cacheWriteTokens',
  'cacheReadTokens',
  'cost',
  'dailyDetail',
  'date',
  'dateRange',
  'dateRange30d',
  'dateRange7d',
  'dateRange90d',
  'dateRangeCustom',
  'dateRangeToday',
  'daysLeft',
  'detailInfo',
  'docs',
  'enterApiKey',
  'expiresAt',
  'inputTokens',
  'limit5h',
  'limit7d',
  'limitDaily',
  'limitMonthly',
  'limitWeekly',
  'model',
  'modelStats',
  'noDailyUsage',
  'outputTokens',
  'placeholder',
  'privacyNote',
  'query',
  'queryFailed',
  'queryFailedRetry',
  'querySuccess',
  'querying',
  'quotaMode',
  'remainingQuota',
  'requests',
  'resetNow',
  'rpmTpm',
  'subscriptionExpires',
  'subscriptionType',
  'subtitle',
  'title',
  'todayCacheCreation',
  'todayCacheRead',
  'todayCost',
  'todayExpires',
  'todayInputTokens',
  'todayOutputTokens',
  'todayRequests',
  'todayTokens',
  'tokenStats',
  'totalCacheCreation',
  'totalCacheRead',
  'totalCost',
  'totalInputTokens',
  'totalOutputTokens',
  'totalQuota',
  'totalRequests',
  'totalTokens',
  'totalTokensLabel',
  'used',
  'usedQuota',
  'walletBalance',
] as const

export type KeyUsageShellLabelKey = typeof keyUsageShellLabelKeys[number]
export type KeyUsageShellLabels = Record<KeyUsageShellLabelKey, string>

export type KeyUsageShellDefaults = {
  defaultDateRange: 'today' | '7d' | '30d'
  dailyUsageDays: 7 | 30 | 90
}

export type KeyUsageShellConfig = {
  labels: KeyUsageShellLabels
  defaults: KeyUsageShellDefaults
}

export const DEFAULT_KEY_USAGE_DATE_RANGE = 'today'
export const DEFAULT_KEY_USAGE_DAILY_USAGE_DAYS = 30
const allowedDateRanges = new Set<KeyUsageShellDefaults['defaultDateRange']>(['today', '7d', '30d'])
const allowedDailyUsageDays = new Set<KeyUsageShellDefaults['dailyUsageDays']>([7, 30, 90])

export function resolveKeyUsageShellLabels(raw: string | undefined, runtimeLocale: 'zh' | 'en'): KeyUsageShellLabels {
  return resolveKeyUsageShellConfig(raw, runtimeLocale).labels
}

export function resolveKeyUsageShellConfig(raw: string | undefined, runtimeLocale: 'zh' | 'en'): KeyUsageShellConfig {
  const localized = pickLocalizedShellConfig(raw, runtimeLocale)
  const defaults = isRecord(localized?.defaults) ? localized.defaults : null
  return {
    labels: resolveLocalizedShellLabels(raw, runtimeLocale, keyUsageShellLabelKeys),
    defaults: {
      defaultDateRange: readDefaultDateRange(defaults?.defaultDateRange),
      dailyUsageDays: readDailyUsageDays(defaults?.dailyUsageDays),
    },
  }
}

function readDefaultDateRange(value: unknown): KeyUsageShellDefaults['defaultDateRange'] {
  return typeof value === 'string' && allowedDateRanges.has(value as KeyUsageShellDefaults['defaultDateRange'])
    ? value as KeyUsageShellDefaults['defaultDateRange']
    : DEFAULT_KEY_USAGE_DATE_RANGE
}

function readDailyUsageDays(value: unknown): KeyUsageShellDefaults['dailyUsageDays'] {
  return typeof value === 'number' && allowedDailyUsageDays.has(value as KeyUsageShellDefaults['dailyUsageDays'])
    ? value as KeyUsageShellDefaults['dailyUsageDays']
    : DEFAULT_KEY_USAGE_DAILY_USAGE_DAYS
}

function pickLocalizedShellConfig(raw: string | undefined, runtimeLocale: 'zh' | 'en'): Record<string, unknown> | null {
  if (!raw?.trim()) return null
  try {
    const parsed = JSON.parse(raw) as unknown
    if (!isRecord(parsed)) return null
    const localized = parsed[runtimeLocale] ?? parsed.en ?? parsed.zh ?? parsed
    return isRecord(localized) ? localized : null
  } catch {
    return null
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

export function renderKeyUsageShellText(
  labels: KeyUsageShellLabels,
  key: KeyUsageShellLabelKey,
  params: Record<string, string | number> = {},
): string {
  let value = labels[key] || ''
  for (const [paramKey, paramValue] of Object.entries(params)) {
    value = value.split(`{${paramKey}}`).join(String(paramValue))
  }
  return value
}
