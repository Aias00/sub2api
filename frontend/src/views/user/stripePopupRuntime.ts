import { formatPaymentAmount, normalizePaymentCurrency } from '@/components/payment/currency'
import { normalizeStripePaymentMethod } from '@/components/payment/paymentMethod'

export function resolveStripePopupRouteState(query: Record<string, unknown>) {
  return {
    orderId: String(query.order_id || ''),
    method: normalizeStripePaymentMethod(String(query.method || '')),
    amount: String(query.amount || ''),
    currency: normalizePaymentCurrency(String(query.currency || '')),
  }
}

export function formatStripePopupDisplayAmount(amount: string, currency: string) {
  const numericAmount = Number.parseFloat(amount)
  if (!Number.isFinite(numericAmount)) return amount
  return formatPaymentAmount(numericAmount, currency)
}

export function buildStripePopupPaymentResultReturnUrl(
  paymentResultPath: string,
  orderId: string,
  origin: string,
) {
  const target = new URL(paymentResultPath, origin)
  target.searchParams.set('order_id', orderId)
  target.searchParams.set('status', 'success')
  return target.toString()
}
