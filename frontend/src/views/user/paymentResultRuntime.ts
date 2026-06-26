import { normalizePaymentCurrency } from '@/components/payment/currency'
import {
  PAYMENT_RECOVERY_STORAGE_KEY,
  clearPaymentRecoverySnapshot,
  readPaymentRecoverySnapshot,
} from '@/components/payment/paymentFlow'
import type { PaymentOrder } from '@/types/payment'
import type { LocationQuery } from 'vue-router'

const SUCCESS_STATUSES = new Set(['COMPLETED', 'PAID', 'RECHARGING'])
const PENDING_STATUSES = new Set(['PENDING', 'CREATED', 'WAITING', 'PROCESSING'])

export function calculatePaymentBaseAmount(order: Pick<PaymentOrder, 'pay_amount' | 'fee_rate'> | null): number {
  if (!order) return 0
  const feeRate = Number(order.fee_rate) || 0
  if (feeRate <= 0) return order.pay_amount ?? 0
  return Math.round((order.pay_amount / (1 + feeRate / 100)) * 100) / 100
}

export function calculatePaymentFeeAmount(order: Pick<PaymentOrder, 'pay_amount' | 'fee_rate'> | null): number {
  if (!order) return 0
  const feeRate = Number(order.fee_rate) || 0
  if (feeRate <= 0) return 0
  return Math.round((order.pay_amount - calculatePaymentBaseAmount(order)) * 100) / 100
}

export function normalizePaymentResultStatus(status: string | null | undefined): string {
  return String(status || '').trim().toUpperCase()
}

export function isPaymentResultSuccess(status: string | null | undefined): boolean {
  return SUCCESS_STATUSES.has(normalizePaymentResultStatus(status))
}

export function isPaymentResultPending(status: string | null | undefined): boolean {
  return PENDING_STATUSES.has(normalizePaymentResultStatus(status))
}

export function readPaymentResultQueryString(query: LocationQuery, key: string): string {
  const value = query[key]
  if (Array.isArray(value)) {
    return typeof value[0] === 'string' ? value[0] : ''
  }
  return typeof value === 'string' ? value : ''
}

export function applyResolvedPaymentOrder(
  order: PaymentOrder | null,
  currentCurrency: string,
): { order: PaymentOrder | null; currency: string } {
  if (order?.currency) {
    return {
      order,
      currency: normalizePaymentCurrency(order.currency),
    }
  }
  return { order, currency: currentCurrency }
}

export function restorePaymentRecoverySnapshot(
  storage: Storage | null,
  context: {
    resumeToken: string
    routeOrderId: number
    routeOutTradeNo: string
  },
) {
  if (!storage) return null
  const rawSnapshot = storage.getItem(PAYMENT_RECOVERY_STORAGE_KEY)
  if (!rawSnapshot) return null

  if (context.resumeToken) {
    return readPaymentRecoverySnapshot(rawSnapshot, {
      resumeToken: context.resumeToken,
    })
  }

  if (!context.routeOrderId && !context.routeOutTradeNo) {
    return null
  }

  const restored = readPaymentRecoverySnapshot(rawSnapshot)
  if (!restored) {
    return null
  }

  if (context.routeOrderId > 0 && restored.orderId !== context.routeOrderId) {
    return null
  }

  if (context.routeOutTradeNo && restored.outTradeNo !== context.routeOutTradeNo) {
    return null
  }

  return restored
}

export function clearPaymentRecoverySnapshotForTerminalStatus(
  storage: Storage | null,
  status: string | null | undefined,
) {
  if (!storage || !status) return
  if (!isPaymentResultPending(status)) {
    clearPaymentRecoverySnapshot(storage, PAYMENT_RECOVERY_STORAGE_KEY)
  }
}
