import { resolveShellLabelOverrides } from './shellLabelOverrides'

export const availableGroupsLabelKeys = [
  'title',
  'description',
  'total',
  'public',
  'memberOnly',
  'searchPlaceholder',
  'emptyTitle',
  'emptyDescription',
  'emptyFilteredDescription',
  'publicTitle',
  'publicDescription',
  'memberTitle',
  'memberDescription',
  'publicBadge',
  'subscriptionBadge',
  'exclusiveBadge',
  'standardBadge',
  'imageEnabledBadge',
  'rate',
  'quota',
  'dailyLimit',
  'weeklyLimit',
  'monthlyLimit',
  'unlimited',
  'loadFailed',
] as const

export type AvailableGroupsLabelKey = typeof availableGroupsLabelKeys[number]
export type AvailableGroupsShellLabels = Partial<Record<AvailableGroupsLabelKey, string>>

export function resolveAvailableGroupsShellLabels(raw: string | undefined, runtimeLocale: string): AvailableGroupsShellLabels {
  return resolveShellLabelOverrides(raw, runtimeLocale, availableGroupsLabelKeys)
}

export function renderAvailableGroupsShellText(
  labels: AvailableGroupsShellLabels,
  key: AvailableGroupsLabelKey,
  values?: Record<string, string | number>,
): string {
  let text = labels[key] || ''
  if (!values) return text
  for (const [name, value] of Object.entries(values)) {
    text = text.split(`{${name}}`).join(String(value))
  }
  return text
}
