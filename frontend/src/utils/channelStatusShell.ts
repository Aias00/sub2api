export type MonitorWindowKey = '7d' | '15d' | '30d'
export type OverallStatusKey = 'operational' | 'degraded'
export type DetailColumnKey =
  | 'model'
  | 'latestStatus'
  | 'latestLatency'
  | 'availability7d'
  | 'availability15d'
  | 'availability30d'
  | 'avgLatency7d'

export interface ChannelStatusShellLabels {
  refreshTitle: string
  detailTitle: string
  loadError: string
  detailLoadError: string
  latency: string
  ping: string
  availabilityPrefix: string
  extraModelsCount: string
  emptyTitle: string
  emptyDescription: string
  closeDetail: string
  windowTab: Record<MonitorWindowKey, string>
  overall: Record<OverallStatusKey, string>
  detailColumns: Record<DetailColumnKey, string>
}

export interface ChannelStatusShellDefaults {
  refreshIntervalSeconds: number
}

export interface ChannelStatusShellConfig {
  labels: ChannelStatusShellLabels
  defaults: ChannelStatusShellDefaults
}

type ParsedChannelStatusShellLabels = Partial<
  Omit<ChannelStatusShellLabels, 'windowTab' | 'overall' | 'detailColumns'>
> & {
  windowTab?: Partial<Record<MonitorWindowKey, string>>
  overall?: Partial<Record<OverallStatusKey, string>>
  detailColumns?: Partial<Record<DetailColumnKey, string>>
}

type ParsedChannelStatusShellConfig = {
  labels?: ParsedChannelStatusShellLabels
  defaults?: Partial<ChannelStatusShellDefaults>
}

export const DEFAULT_CHANNEL_STATUS_REFRESH_INTERVAL_SECONDS = 60

const flatLabelKeys = [
  'refreshTitle',
  'detailTitle',
  'loadError',
  'detailLoadError',
  'latency',
  'ping',
  'availabilityPrefix',
  'extraModelsCount',
  'emptyTitle',
  'emptyDescription',
  'closeDetail',
] as const

export function resolveChannelStatusShellLabels(
  raw: string | undefined,
  runtimeLocale: string,
): ChannelStatusShellLabels {
  return resolveChannelStatusShellConfig(raw, runtimeLocale).labels
}

export function resolveChannelStatusShellConfig(
  raw: string | undefined,
  runtimeLocale: string,
): ChannelStatusShellConfig {
  const parsed = readChannelStatusShellOverrides(raw, runtimeLocale)
  const empty = emptyChannelStatusShellLabels()
  return {
    labels: {
      ...empty,
      ...parsed.labels,
      windowTab: {
        ...empty.windowTab,
        ...parsed.labels?.windowTab,
      },
      overall: {
        ...empty.overall,
        ...parsed.labels?.overall,
      },
      detailColumns: {
        ...empty.detailColumns,
        ...parsed.labels?.detailColumns,
      },
    },
    defaults: {
      refreshIntervalSeconds:
        parsed.defaults?.refreshIntervalSeconds || DEFAULT_CHANNEL_STATUS_REFRESH_INTERVAL_SECONDS,
    },
  }
}

function emptyChannelStatusShellLabels(): ChannelStatusShellLabels {
  return {
    refreshTitle: '',
    detailTitle: '',
    loadError: '',
    detailLoadError: '',
    latency: '',
    ping: '',
    availabilityPrefix: '',
    extraModelsCount: '',
    emptyTitle: '',
    emptyDescription: '',
    closeDetail: '',
    windowTab: {
      '7d': '',
      '15d': '',
      '30d': '',
    },
    overall: {
      operational: '',
      degraded: '',
    },
    detailColumns: {
      model: '',
      latestStatus: '',
      latestLatency: '',
      availability7d: '',
      availability15d: '',
      availability30d: '',
      avgLatency7d: '',
    },
  }
}

function readChannelStatusShellOverrides(
  raw: string | undefined,
  runtimeLocale: string,
): ParsedChannelStatusShellConfig {
  if (!raw) return {}
  try {
    const parsed = JSON.parse(raw) as Record<
      string,
      { labels?: Record<string, unknown>; defaults?: Record<string, unknown> } | undefined
    >
    const normalizedLocale = runtimeLocale.toLowerCase()
    const localeKeys = [normalizedLocale, normalizedLocale.split('-')[0], 'en', 'zh']
    for (const key of localeKeys) {
      const localized = parsed[key]
      if (!localized) continue
      const labels = localized.labels || {}
      const result: ParsedChannelStatusShellLabels = {}
      for (const labelKey of flatLabelKeys) {
        if (typeof labels[labelKey] === 'string') {
          result[labelKey] = labels[labelKey]
        }
      }
      result.windowTab = readStringMap(labels.windowTab, ['7d', '15d', '30d'] as const)
      result.overall = readStringMap(labels.overall, ['operational', 'degraded'] as const)
      result.detailColumns = readStringMap(
        labels.detailColumns,
        [
          'model',
          'latestStatus',
          'latestLatency',
          'availability7d',
          'availability15d',
          'availability30d',
          'avgLatency7d',
        ] as const,
      )
      return {
        labels: result,
        defaults: readChannelStatusShellDefaults(localized.defaults),
      }
    }
  } catch {
    return {}
  }
  return {}
}

function readChannelStatusShellDefaults(value: unknown): Partial<ChannelStatusShellDefaults> {
  if (!value || typeof value !== 'object') return {}
  const source = value as Record<string, unknown>
  return {
    refreshIntervalSeconds: readRefreshIntervalSeconds(source.refreshIntervalSeconds),
  }
}

function readRefreshIntervalSeconds(value: unknown): number | undefined {
  const seconds = Number(value)
  if (!Number.isFinite(seconds)) return undefined
  const normalized = Math.floor(seconds)
  if (normalized < 15 || normalized > 3600) return undefined
  return normalized
}

function readStringMap<T extends string>(value: unknown, allowedKeys: readonly T[]): Partial<Record<T, string>> {
  if (!value || typeof value !== 'object') return {}
  const source = value as Record<string, unknown>
  const result: Partial<Record<T, string>> = {}
  for (const key of allowedKeys) {
    if (typeof source[key] === 'string') {
      result[key] = source[key]
    }
  }
  return result
}
