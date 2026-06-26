import type { RedeemHistoryItem } from '@/api'
import type { RedeemLabelKey } from '@/utils/redeemShell'

export type RedeemTextGetter = (
  key: RedeemLabelKey,
  values?: Record<string, string | number>,
) => string

export function isRedeemBalanceType(type: string) {
  return type === 'balance' || type === 'admin_balance'
}

export function isRedeemSubscriptionType(type: string) {
  return type === 'subscription'
}

export function isRedeemAdminAdjustment(type: string) {
  return type === 'admin_balance' || type === 'admin_concurrency'
}

export function resolveRedeemHistoryItemTitle(
  item: RedeemHistoryItem,
  redeemText: RedeemTextGetter,
) {
  if (item.type === 'balance') {
    return redeemText('balanceAddedRedeem')
  } else if (item.type === 'admin_balance') {
    return item.value >= 0 ? redeemText('balanceAddedAdmin') : redeemText('balanceDeductedAdmin')
  } else if (item.type === 'concurrency') {
    return redeemText('concurrencyAddedRedeem')
  } else if (item.type === 'admin_concurrency') {
    return item.value >= 0 ? redeemText('concurrencyAddedAdmin') : redeemText('concurrencyReducedAdmin')
  } else if (item.type === 'subscription') {
    return redeemText('subscriptionAssigned')
  }
  return redeemText('unknown')
}

export function formatRedeemHistoryValue(
  item: RedeemHistoryItem,
  redeemText: RedeemTextGetter,
  formatCurrency: (value: number) => string,
) {
  if (isRedeemBalanceType(item.type)) {
    const sign = item.value >= 0 ? '+' : ''
    return `${sign}${formatCurrency(item.value)}`
  } else if (isRedeemSubscriptionType(item.type)) {
    const days = item.validity_days || Math.round(item.value)
    const groupName = item.group?.name || ''
    return groupName ? `${days}${redeemText('days')} - ${groupName}` : `${days}${redeemText('days')}`
  } else {
    const sign = item.value >= 0 ? '+' : ''
    return `${sign}${item.value} ${redeemText('requests')}`
  }
}
