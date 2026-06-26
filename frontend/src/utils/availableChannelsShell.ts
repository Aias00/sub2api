export const availableChannelsLabelKeys = [
  'searchPlaceholder',
  'refreshTitle',
  'noPricing',
  'noModels',
  'empty',
  'loadError',
  'exclusive',
  'exclusiveTooltip',
  'public',
  'publicTooltip',
  'columnName',
  'columnDescription',
  'columnPlatform',
  'columnGroups',
  'columnSupportedModels',
  'pricingBillingMode',
  'pricingBillingModeImage',
  'pricingBillingModePerRequest',
  'pricingBillingModeToken',
  'pricingCacheReadPrice',
  'pricingCacheWritePrice',
  'pricingImageOutputPrice',
  'pricingInputPrice',
  'pricingIntervals',
  'pricingOutputPrice',
  'pricingPerRequestPrice',
  'pricingUnitPerMillion',
  'pricingUnitPerRequest',
] as const

export type AvailableChannelsLabelKey = typeof availableChannelsLabelKeys[number]
export type AvailableChannelsShellLabels = Record<AvailableChannelsLabelKey, string>

export function resolveConfiguredAvailableChannelsShellLabels(
  raw: string | undefined,
  runtimeLocale: string,
): AvailableChannelsShellLabels {
  return resolveAvailableChannelsShellLabels(raw, runtimeLocale, availableChannelsLabelKeys)
}

export function resolveAvailableChannelsShellLabels<K extends string>(
  raw: string | undefined,
  runtimeLocale: string,
  allowedKeys: readonly K[],
): Record<K, string> {
  const emptyLabels = createEmptyLabels(allowedKeys)
  if (!raw) return emptyLabels
  try {
    const parsed = JSON.parse(raw) as Record<string, { labels?: Record<string, unknown> } | undefined>
    const normalizedLocale = runtimeLocale.toLowerCase()
    const localeKeys = [normalizedLocale, normalizedLocale.split('-')[0], 'en', 'zh']
    for (const key of localeKeys) {
      const labels = parsed[key]?.labels
      if (!labels) continue
      const result: Record<K, string> = { ...emptyLabels }
      for (const labelKey of allowedKeys) {
        const value = labels[labelKey]
        if (typeof value === 'string') {
          result[labelKey] = value
        }
      }
      applyColumnLabels(result, labels)
      applyPricingLabels(result, labels)
      return result
    }
  } catch {
    return emptyLabels
  }
  return emptyLabels
}

function applyPricingLabels<K extends string>(result: Record<K, string>, labels: Record<string, unknown>) {
  const pricing = labels.pricing
  if (!pricing || typeof pricing !== 'object') return
  const pricingValues = pricing as Record<string, unknown>
  const pricingKeyMap = {
    billingMode: 'pricingBillingMode',
    billingModeImage: 'pricingBillingModeImage',
    billingModePerRequest: 'pricingBillingModePerRequest',
    billingModeToken: 'pricingBillingModeToken',
    cacheReadPrice: 'pricingCacheReadPrice',
    cacheWritePrice: 'pricingCacheWritePrice',
    imageOutputPrice: 'pricingImageOutputPrice',
    inputPrice: 'pricingInputPrice',
    intervals: 'pricingIntervals',
    outputPrice: 'pricingOutputPrice',
    perRequestPrice: 'pricingPerRequestPrice',
    unitPerMillion: 'pricingUnitPerMillion',
    unitPerRequest: 'pricingUnitPerRequest',
  } as const
  for (const [sourceKey, targetKey] of Object.entries(pricingKeyMap)) {
    const value = pricingValues[sourceKey]
    if (typeof value === 'string' && targetKey in result) {
      result[targetKey as K] = value
    }
  }
}

function createEmptyLabels<K extends string>(allowedKeys: readonly K[]): Record<K, string> {
  return allowedKeys.reduce(
    (labels, key) => {
      labels[key] = ''
      return labels
    },
    {} as Record<K, string>,
  )
}

function applyColumnLabels<K extends string>(result: Record<K, string>, labels: Record<string, unknown>) {
  const columns = labels.columns
  if (!columns || typeof columns !== 'object') return
  const columnValues = columns as Record<string, unknown>
  const columnKeyMap = {
    name: 'columnName',
    description: 'columnDescription',
    platform: 'columnPlatform',
    groups: 'columnGroups',
    supportedModels: 'columnSupportedModels',
  } as const
  for (const [sourceKey, targetKey] of Object.entries(columnKeyMap)) {
    const value = columnValues[sourceKey]
    if (typeof value === 'string' && targetKey in result) {
      result[targetKey as K] = value
    }
  }
}
