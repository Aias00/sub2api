import { formatDateOnly } from '@/utils/format'
import { getRemainingDurationParts, isOneTimeDailyQuota, type RemainingDurationParts } from '@/utils/subscriptionQuota'
import type { UserSubscription } from '@/types'
import type { SubscriptionLabelKey } from '@/utils/paymentShell'

export type SubscriptionTextGetter = (
  key: SubscriptionLabelKey,
  params?: Record<string, string | number>,
) => string

export function resolveSubscriptionProgressWidth(used: number | undefined, limit: number | null | undefined): string {
  if (!limit || limit === 0) return '0%'
  const percentage = Math.min(((used || 0) / limit) * 100, 100)
  return `${percentage}%`
}

export function resolveSubscriptionProgressBarClass(used: number | undefined, limit: number | null | undefined): string {
  if (!limit || limit === 0) return 'bg-gray-400'
  const percentage = ((used || 0) / limit) * 100
  if (percentage >= 90) return 'bg-red-500'
  if (percentage >= 70) return 'bg-orange-500'
  return 'bg-green-500'
}

export function formatSubscriptionDurationParts(parts: RemainingDurationParts): string {
  if (parts.days > 0) {
    return `${parts.days}d ${parts.hours}h`
  }

  if (parts.hours > 0) {
    return `${parts.hours}h ${parts.minutes}m`
  }

  return `${parts.minutes}m`
}

export function formatSubscriptionResetTime(
  windowStart: string | null,
  windowHours: number,
  paymentText: SubscriptionTextGetter,
): string {
  if (!windowStart) return paymentText('subscriptionWindowNotActive')

  const start = new Date(windowStart)
  const end = new Date(start.getTime() + windowHours * 60 * 60 * 1000)
  const parts = getRemainingDurationParts(end)

  return parts ? formatSubscriptionDurationParts(parts) : paymentText('subscriptionWindowNotActive')
}

export function formatSubscriptionDailyUsageWindow(
  subscription: UserSubscription,
  paymentText: SubscriptionTextGetter,
): string {
  if (isOneTimeDailyQuota(subscription) && subscription.expires_at) {
    const parts = getRemainingDurationParts(subscription.expires_at)
    if (!parts) return paymentText('subscriptionWindowNotActive')
    return paymentText('subscriptionQuotaEndsIn', { time: formatSubscriptionDurationParts(parts) })
  }

  return paymentText('subscriptionResetIn', {
    time: formatSubscriptionResetTime(subscription.daily_window_start, 24, paymentText),
  })
}

export function formatSubscriptionExpirationDate(expiresAt: string, paymentText: SubscriptionTextGetter): string {
  const now = new Date()
  const expires = new Date(expiresAt)
  const diff = expires.getTime() - now.getTime()
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))

  if (days < 0) {
    return paymentText('subscriptionStatusExpired')
  }

  const dateStr = formatDateOnly(expires)

  if (days === 0) {
    return `${dateStr} (${paymentText('subscriptionToday')})`
  }
  if (days === 1) {
    return `${dateStr} (${paymentText('subscriptionTomorrow')})`
  }

  return paymentText('subscriptionDaysRemaining', { days }) + ` (${dateStr})`
}

export function resolveSubscriptionExpirationClass(expiresAt: string): string {
  const now = new Date()
  const expires = new Date(expiresAt)
  const diff = expires.getTime() - now.getTime()
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))

  if (days <= 0) return 'text-red-600 dark:text-red-400 font-medium'
  if (days <= 3) return 'text-red-600 dark:text-red-400'
  if (days <= 7) return 'text-orange-600 dark:text-orange-400'
  return 'text-gray-700 dark:text-gray-300'
}

export function resolveSubscriptionStatusText(
  status: UserSubscription['status'],
  paymentText: SubscriptionTextGetter,
): string {
  if (status === 'active') return paymentText('subscriptionStatusActive')
  if (status === 'expired') return paymentText('subscriptionStatusExpired')
  return paymentText('subscriptionStatusRevoked')
}
