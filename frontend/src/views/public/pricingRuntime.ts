import type { RechargeProduct, SubscriptionPlan } from '@/types/payment'
import type { PricingCopy, PricingShellConfig } from '@/utils/pricingShell'

export function comparePricingCatalogItems(a: RechargeProduct | SubscriptionPlan, b: RechargeProduct | SubscriptionPlan) {
  return (a.sort_order ?? 0) - (b.sort_order ?? 0)
}

export function resolvePricingShellGroupLabel(shellConfig: PricingShellConfig, name: string): string {
  const group = shellConfig.groups?.find((item) => item.name === name)
  return group?.title?.trim() || ''
}

export function resolvePricingShellLabel(shellConfig: PricingShellConfig, key: keyof PricingShellConfig['labels']): string {
  return shellConfig.labels?.[key]?.trim() || ''
}

export function resolvePricingBuyButton(shellConfig: PricingShellConfig, copy: PricingCopy): string {
  return shellConfig.button?.title?.trim() || copy.buy
}

export function resolvePricingPromptsPath(shellConfig: PricingShellConfig): string {
  return shellConfig.defaults?.promptsPath?.trim() || ''
}

export function resolvePricingPurchasePath(shellConfig: PricingShellConfig): string {
  return shellConfig.defaults?.purchasePath?.trim() || ''
}

export function buildPricingPurchaseRoute(path: string, tab: 'recharge' | 'subscription', group?: string) {
  if (!path) return null
  return {
    path,
    query: {
      tab,
      ...(group ? { group } : {}),
    },
  }
}

export function formatPricingCurrency(value: number | undefined, currencySymbol: string) {
  const amount = Number.isFinite(value) ? Number(value) : 0
  return `${currencySymbol}${amount.toFixed(amount % 1 === 0 ? 0 : 2)}`
}

export function formatPricingCredits(value: number | undefined) {
  const amount = Number.isFinite(value) ? Number(value) : 0
  return amount.toFixed(amount % 1 === 0 ? 0 : 2)
}

export function resolvePricingPlanSourceLabel(plan: SubscriptionPlan): string {
  return plan.group_display_label || plan.group_name || plan.group_platform || ''
}

export function resolvePricingPlanValidity(
  plan: SubscriptionPlan,
  labels: { day: string; days: string; month: string },
): string {
  const days = plan.validity_days ?? 0
  if (days <= 0) return ''
  if (plan.validity_unit === 'month') {
    return `/${days}${labels.month}`
  }
  if (plan.validity_unit === 'day') {
    return `/${days}${days === 1 ? labels.day : labels.days}`
  }
  return ''
}
