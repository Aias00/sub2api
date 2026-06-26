import type { LocationQuery } from 'vue-router'

export function resolvePaymentQrRouteState(query: LocationQuery) {
  return {
    orderId: Number(query.order_id) || 0,
    qrUrl: String(query.qr || ''),
    payUrl: String(query.pay_url || ''),
    paymentType: String(query.payment_type || ''),
    expiresAt: String(query.expires_at || ''),
  }
}

export function formatPaymentQrCountdown(seconds: number): string {
  const minutes = Math.floor(seconds / 60)
  const remainder = seconds % 60
  return `${minutes.toString().padStart(2, '0')}:${remainder.toString().padStart(2, '0')}`
}

export function resolvePaymentQrSecondsUntil(expiresAtStr: string, now = Date.now()): number {
  if (!expiresAtStr) return 0
  const expiresAt = Date.parse(expiresAtStr)
  if (!Number.isFinite(expiresAt)) return 0
  return Math.floor((expiresAt - now) / 1000)
}

export function isPaymentQrCompleted(status: string): boolean {
  return status === 'COMPLETED' || status === 'PAID'
}

export function isPaymentQrTerminal(status: string): boolean {
  return status === 'EXPIRED' || status === 'CANCELLED' || status === 'FAILED'
}
