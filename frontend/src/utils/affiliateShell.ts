import { resolveShellLabelOverrides } from './shellLabelOverrides'

export const affiliateLabelKeys = [
  'rebateRate',
  'rebateRateHint',
  'invitedUsers',
  'availableQuota',
  'totalQuota',
  'frozenQuota',
  'title',
  'description',
  'yourCode',
  'copyCode',
  'inviteLink',
  'copyLink',
  'tipsTitle',
  'tipShare',
  'tipRebate',
  'tipTransfer',
  'tipFreeze',
  'transferTitle',
  'transferDescription',
  'transferButton',
  'transferring',
  'transferEmpty',
  'transferSuccess',
  'inviteesTitle',
  'inviteesEmpty',
  'emailColumn',
  'usernameColumn',
  'rebateColumn',
  'joinedAtColumn',
  'rebatesTitle',
  'rebatesEmpty',
  'inviteeColumn',
  'orderAmountColumn',
  'payAmountColumn',
  'rebateAmountColumn',
  'paymentTypeColumn',
  'orderStatusColumn',
  'createdAtColumn',
  'transfersTitle',
  'transfersEmpty',
  'amountColumn',
  'balanceAfterColumn',
  'availableQuotaAfterColumn',
  'frozenQuotaAfterColumn',
  'historyQuotaAfterColumn',
  'transferredAtColumn',
  'codeCopied',
  'linkCopied',
  'loadFailed',
  'transferFailed',
] as const

export type AffiliateLabelKey = typeof affiliateLabelKeys[number]
export type AffiliateShellLabels = Partial<Record<AffiliateLabelKey, string>>

export function resolveAffiliateShellLabels(raw: string | undefined, runtimeLocale: string): AffiliateShellLabels {
  return resolveShellLabelOverrides(raw, runtimeLocale, affiliateLabelKeys)
}

export function renderAffiliateShellText(
  labels: AffiliateShellLabels,
  key: AffiliateLabelKey,
  values?: Record<string, string | number>,
): string {
  let text = labels[key] || ''
  if (!values) return text
  for (const [name, value] of Object.entries(values)) {
    text = text.split(`{${name}}`).join(String(value))
  }
  return text
}
