import { resolveShellLabelOverrides } from './shellLabelOverrides'

export const customPageLabelKeys = [
  'tocTitle',
  'tocToggle',
  'notFoundTitle',
  'notFoundDesc',
  'notConfiguredTitle',
  'notConfiguredDesc',
  'openInNewTab',
  'markdownNotFound',
  'markdownLoadFailed',
  'copyCode',
  'copyCodeSuccess',
  'copyCodeFailed',
] as const

export type CustomPageLabelKey = typeof customPageLabelKeys[number]
export type CustomPageShellLabels = Partial<Record<CustomPageLabelKey, string>>

export function resolveCustomPageShellLabels(raw: string | undefined, runtimeLocale: string): CustomPageShellLabels {
  return resolveShellLabelOverrides(raw, runtimeLocale, customPageLabelKeys)
}

export function renderCustomPageShellText(labels: CustomPageShellLabels, key: CustomPageLabelKey): string {
  return labels[key] || ''
}
