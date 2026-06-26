import type { Group, SubscriptionType } from '@/types'
import type { AvailableGroupsLabelKey } from '@/utils/availableGroupsShell'

export type AvailableGroupsTextGetter = (
  key: AvailableGroupsLabelKey,
  values?: Record<string, string | number>,
) => string

export function filterAvailableGroupsByQuery(groups: Group[], query: string): Group[] {
  const normalized = query.trim().toLowerCase()
  if (!normalized) return groups
  return groups.filter((group) => {
    const haystack = [group.name, group.description || '', group.platform, group.subscription_type]
      .join(' ')
      .toLowerCase()
    return haystack.includes(normalized)
  })
}

export function resolvePublicAvailableGroups(groups: Group[]): Group[] {
  return groups.filter((group) => !group.is_exclusive && group.subscription_type === 'standard')
}

export function resolveMemberAvailableGroups(groups: Group[]): Group[] {
  return groups.filter((group) => group.is_exclusive || group.subscription_type === 'subscription')
}

export function resolveAvailableGroupSubscriptionLabel(
  type: SubscriptionType,
  availableGroupsText: AvailableGroupsTextGetter,
): string {
  return type === 'subscription'
    ? availableGroupsText('subscriptionBadge')
    : availableGroupsText('standardBadge')
}

export function resolveAvailableGroupQuotaSummary(
  group: Group,
  availableGroupsText: AvailableGroupsTextGetter,
): string {
  const limits: string[] = []
  if (group.daily_limit_usd != null) limits.push(availableGroupsText('dailyLimit', { amount: group.daily_limit_usd.toFixed(2) }))
  if (group.weekly_limit_usd != null) limits.push(availableGroupsText('weeklyLimit', { amount: group.weekly_limit_usd.toFixed(2) }))
  if (group.monthly_limit_usd != null) limits.push(availableGroupsText('monthlyLimit', { amount: group.monthly_limit_usd.toFixed(2) }))
  return limits.length > 0 ? limits.join(' / ') : availableGroupsText('unlimited')
}
