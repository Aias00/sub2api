import { resolveShellLabelOverrides } from './shellLabelOverrides'

export const usageShellLabelKeys = [
  'totalRequests',
  'inSelectedRange',
  'totalTokens',
  'in',
  'out',
  'totalCost',
  'actualCost',
  'standardCost',
  'avgDuration',
  'perRequest',
  'apiKeyFilter',
  'allApiKeys',
  'timeRange',
  'refresh',
  'reset',
  'exportCsv',
  'exporting',
  'model',
  'reasoningEffort',
  'endpoint',
  'type',
  'billingMode',
  'tokens',
  'cost',
  'firstToken',
  'duration',
  'time',
  'userAgent',
  'noRecords',
  'rate',
  'original',
  'billed',
  'failedToLoad',
  'noDataToExport',
  'preparingExport',
  'exportSuccess',
  'exportFailed',
] as const

export type UsageShellLabelKey = typeof usageShellLabelKeys[number]
export type UsageShellLabels = Partial<Record<UsageShellLabelKey, string>>

export type UsageShellDefaults = {
  dateRangeDays: number
  apiKeyPageSize: number
  exportPageSize: number
}

export type UsageShellConfig = {
  labels: UsageShellLabels
  defaults: UsageShellDefaults
}

export const DEFAULT_USAGE_EXPORT_PAGE_SIZE = 100
export const DEFAULT_USAGE_DATE_RANGE_DAYS = 7
export const DEFAULT_USAGE_API_KEY_PAGE_SIZE = 100

export function resolveUsageShellLabels(raw: string | undefined, runtimeLocale: string): UsageShellLabels {
  return resolveUsageShellConfig(raw, runtimeLocale).labels
}

export function resolveUsageShellConfig(raw: string | undefined, runtimeLocale: string): UsageShellConfig {
  const localized = pickLocalizedShellConfig(raw, runtimeLocale)
  return {
    labels: resolveShellLabelOverrides(raw, runtimeLocale, usageShellLabelKeys),
    defaults: {
      dateRangeDays: readPositiveIntegerDefault(
        localized?.defaults,
        'dateRangeDays',
        DEFAULT_USAGE_DATE_RANGE_DAYS,
        366,
      ),
      apiKeyPageSize: readPositiveIntegerDefault(
        localized?.defaults,
        'apiKeyPageSize',
        DEFAULT_USAGE_API_KEY_PAGE_SIZE,
        1000,
      ),
      exportPageSize: readExportPageSize(localized?.defaults),
    },
  }
}

export function renderUsageShellText(labels: UsageShellLabels, key: UsageShellLabelKey): string {
  return labels[key] || ''
}

function pickLocalizedShellConfig(raw: string | undefined, runtimeLocale: string): Record<string, unknown> | null {
  if (!raw?.trim()) return null
  try {
    const parsed = JSON.parse(raw) as unknown
    if (!isRecord(parsed)) return null
    const normalizedLocale = runtimeLocale.toLowerCase()
    const baseLocale = normalizedLocale.split('-')[0]
    const localized = parsed[normalizedLocale] ?? parsed[baseLocale] ?? parsed.en ?? parsed.zh ?? parsed
    return isRecord(localized) ? localized : null
  } catch {
    return null
  }
}

function readExportPageSize(value: unknown): number {
  return readPositiveIntegerDefault(value, 'exportPageSize', DEFAULT_USAGE_EXPORT_PAGE_SIZE, 1000)
}

function readPositiveIntegerDefault(
  value: unknown,
  key: string,
  fallback: number,
  max: number,
): number {
  if (!isRecord(value)) return fallback
  const normalized = Number(value[key])
  if (!Number.isInteger(normalized) || normalized <= 0 || normalized > max) {
    return fallback
  }
  return normalized
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}
