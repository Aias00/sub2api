import { resolvePaymentShellLabels } from './paymentShell'

export const userOrdersLabelKeys = [
  'refresh',
  'all',
  'pending',
  'completed',
  'failed',
  'refunded',
  'backToRecharge',
  'cancelOrder',
  'requestRefund',
  'confirmCancel',
  'cancel',
  'processing',
  'orderId',
  'orderNo',
  'amount',
  'payAmount',
  'paymentMethod',
  'status',
  'createdAt',
  'actions',
  'fee',
  'creditedAmount',
  'refundReason',
  'refundReasonPlaceholder',
  'cancelSuccess',
  'refundSuccess',
  'errorFallback',
  'methodAlipay',
  'methodWxpay',
  'methodStripe',
  'methodAirwallex',
  'statusPending',
  'statusPaid',
  'statusRecharging',
  'statusCompleted',
  'statusExpired',
  'statusCancelled',
  'statusFailed',
  'statusRefundRequested',
  'statusRefunding',
  'statusRefunded',
  'statusPartiallyRefunded',
  'statusRefundFailed',
] as const

export type UserOrdersLabelKey = typeof userOrdersLabelKeys[number]
export type UserOrdersShellLabels = Partial<Record<UserOrdersLabelKey, string>>

export function resolveUserOrdersShellLabels(raw: string | undefined, runtimeLocale: string): UserOrdersShellLabels {
  return resolvePaymentShellLabels(raw, runtimeLocale, userOrdersLabelKeys)
}

export function renderUserOrdersShellText(labels: UserOrdersShellLabels, key: UserOrdersLabelKey): string {
  return labels[key] || ''
}
