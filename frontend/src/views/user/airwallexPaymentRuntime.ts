import {
  PAYMENT_RECOVERY_STORAGE_KEY,
  readPaymentRecoverySnapshot,
  type PaymentRecoverySnapshot,
} from '@/components/payment/paymentFlow'
import type { LocationQuery } from 'vue-router'

export function readAirwallexQueryString(query: LocationQuery, key: string): string {
  const value = query[key]
  if (Array.isArray(value)) return value[0] || ''
  return typeof value === 'string' ? value : ''
}

export function buildAirwallexSuccessUrl(
  paymentResultPath: string,
  query: LocationQuery,
  snapshot: PaymentRecoverySnapshot,
  origin: string,
): string {
  const url = new URL(paymentResultPath, origin)
  const orderId = readAirwallexQueryString(query, 'order_id')
  const outTradeNo = readAirwallexQueryString(query, 'out_trade_no')
  const resumeToken = readAirwallexQueryString(query, 'resume_token')

  if (orderId || snapshot.orderId > 0) url.searchParams.set('order_id', orderId || String(snapshot.orderId))
  if (outTradeNo || snapshot.outTradeNo) url.searchParams.set('out_trade_no', outTradeNo || snapshot.outTradeNo)
  if (resumeToken || snapshot.resumeToken) url.searchParams.set('resume_token', resumeToken || snapshot.resumeToken)
  return url.toString()
}

export function restoreAirwallexPaymentSnapshot(
  storage: Storage | null,
  query: LocationQuery,
): PaymentRecoverySnapshot | null {
  if (!storage) {
    return null
  }

  const orderId = Number(readAirwallexQueryString(query, 'order_id')) || 0
  const outTradeNo = readAirwallexQueryString(query, 'out_trade_no')
  const resumeToken = readAirwallexQueryString(query, 'resume_token')
  const snapshot = readPaymentRecoverySnapshot(
    storage.getItem(PAYMENT_RECOVERY_STORAGE_KEY),
    resumeToken ? { resumeToken } : {},
  )

  if (!snapshot || snapshot.paymentType !== 'airwallex') {
    return null
  }
  if (orderId > 0 && snapshot.orderId !== orderId) {
    return null
  }
  if (outTradeNo && snapshot.outTradeNo !== outTradeNo) {
    return null
  }
  if (!snapshot.intentId || !snapshot.clientSecret) {
    return null
  }
  return snapshot
}
