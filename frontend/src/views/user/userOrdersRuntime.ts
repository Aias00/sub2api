import type { PaymentOrder } from '@/types/payment'
import type { UserOrdersLabelKey } from '@/utils/userOrdersShell'

export type UserOrdersTextGetter = (key: UserOrdersLabelKey) => string

export function buildUserOrdersStatusFilters(paymentText: UserOrdersTextGetter) {
  return [
    { value: '', label: paymentText('all') },
    { value: 'PENDING', label: paymentText('pending') },
    { value: 'COMPLETED', label: paymentText('completed') },
    { value: 'FAILED', label: paymentText('failed') },
    { value: 'REFUNDED', label: paymentText('refunded') },
  ]
}

export function buildUserOrdersTableLabels(paymentText: UserOrdersTextGetter) {
  return {
    orderId: paymentText('orderId'),
    orderNo: paymentText('orderNo'),
    payAmount: paymentText('payAmount'),
    paymentMethod: paymentText('paymentMethod'),
    status: paymentText('status'),
    createdAt: paymentText('createdAt'),
    actions: paymentText('actions'),
    fee: paymentText('fee'),
    creditedAmount: paymentText('creditedAmount'),
    methodAlipay: paymentText('methodAlipay'),
    methodWxpay: paymentText('methodWxpay'),
    methodStripe: paymentText('methodStripe'),
    methodAirwallex: paymentText('methodAirwallex'),
    statusPending: paymentText('statusPending'),
    statusPaid: paymentText('statusPaid'),
    statusRecharging: paymentText('statusRecharging'),
    statusCompleted: paymentText('statusCompleted'),
    statusExpired: paymentText('statusExpired'),
    statusCancelled: paymentText('statusCancelled'),
    statusFailed: paymentText('statusFailed'),
    statusRefundRequested: paymentText('statusRefundRequested'),
    statusRefunding: paymentText('statusRefunding'),
    statusRefunded: paymentText('statusRefunded'),
    statusPartiallyRefunded: paymentText('statusPartiallyRefunded'),
    statusRefundFailed: paymentText('statusRefundFailed'),
  }
}

export function canUserOrderRequestRefund(
  order: Pick<PaymentOrder, 'status' | 'provider_instance_id'>,
  refundEligibleProviders: Set<string>,
): boolean {
  if (order.status !== 'COMPLETED') return false
  if (!order.provider_instance_id) return false
  return refundEligibleProviders.has(order.provider_instance_id)
}
