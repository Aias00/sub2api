import { resolveShellLabelOverrides } from './shellLabelOverrides'

export const redeemLabelKeys = [
  'currentBalance',
  'concurrency',
  'requests',
  'redeemCodeLabel',
  'redeemCodePlaceholder',
  'redeemCodeHint',
  'redeemButton',
  'redeeming',
  'redeemSuccess',
  'redeemFailed',
  'added',
  'concurrentRequests',
  'subscriptionAssigned',
  'subscriptionDays',
  'newBalance',
  'newConcurrency',
  'aboutCodes',
  'codeRule1',
  'codeRule2',
  'codeRule3',
  'codeRule4',
  'recentActivity',
  'historyWillAppear',
  'adminAdjustment',
  'balanceAddedRedeem',
  'balanceAddedAffiliate',
  'balanceAddedAdmin',
  'balanceDeductedAdmin',
  'concurrencyAddedRedeem',
  'concurrencyAddedAdmin',
  'concurrencyReducedAdmin',
  'days',
  'pleaseEnterCode',
  'subscriptionRefreshFailed',
  'codeRedeemSuccess',
  'failedToRedeem',
  'unknown',
] as const

export type RedeemLabelKey = typeof redeemLabelKeys[number]
export type RedeemShellLabels = Partial<Record<RedeemLabelKey, string>>

export function resolveRedeemShellLabels(raw: string | undefined, runtimeLocale: string): RedeemShellLabels {
  return resolveShellLabelOverrides(raw, runtimeLocale, redeemLabelKeys)
}

export function renderRedeemShellText(
  labels: RedeemShellLabels,
  key: RedeemLabelKey,
  values?: Record<string, string | number>,
): string {
  let text = labels[key] || ''
  if (!values) return text
  for (const [name, value] of Object.entries(values)) {
    text = text.split(`{${name}}`).join(String(value))
  }
  return text
}
