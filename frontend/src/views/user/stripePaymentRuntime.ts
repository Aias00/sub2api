import { formatPaymentAmount, normalizePaymentCurrency } from '@/components/payment/currency'
import {
  PAYMENT_RECOVERY_STORAGE_KEY,
  readPaymentRecoverySnapshot,
} from '@/components/payment/paymentFlow'
import type { LocationQuery } from 'vue-router'

export function resolveStripePaymentRouteState(query: LocationQuery) {
  return {
    orderId: Number(query.order_id),
    clientSecret: String(query.client_secret || ''),
    method: String(query.method || ''),
    resumeToken: typeof query.resume_token === 'string' ? query.resume_token : undefined,
  }
}

export function formatStripeGatewayAmount(value: number, currency: string, localeCode?: string) {
  return formatPaymentAmount(value, currency, localeCode)
}

export function restoreStripePaymentCurrency(
  storage: Storage | null,
  resumeToken: string | undefined,
  orderId: number,
): string {
  if (!storage) return ''
  const restored = readPaymentRecoverySnapshot(
    storage.getItem(PAYMENT_RECOVERY_STORAGE_KEY),
    { resumeToken },
  )
  if (restored?.orderId === orderId) {
    return normalizePaymentCurrency(restored.currency)
  }
  return ''
}

export function buildStripePaymentResultRoute(paymentResultPath: string, orderId: unknown) {
  return {
    path: paymentResultPath,
    query: { order_id: String(orderId || ''), status: 'success' },
  }
}

export function buildStripePaymentResultReturnURL(
  paymentResultPath: string,
  orderId: unknown,
  origin: string,
) {
  const target = new URL(paymentResultPath, origin)
  target.searchParams.set('order_id', String(orderId || ''))
  target.searchParams.set('status', 'success')
  return target.toString()
}
