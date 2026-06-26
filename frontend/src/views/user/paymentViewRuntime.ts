import type { PaymentRecoverySnapshot } from '@/components/payment/paymentFlow'
import type { OrderType } from '@/types/payment'
import { normalizeVisibleMethod } from '@/components/payment/paymentFlow'

export function createEmptyPaymentRecoveryState(): PaymentRecoverySnapshot {
  return {
    orderId: 0,
    amount: 0,
    qrCode: '',
    expiresAt: '',
    paymentType: '',
    payUrl: '',
    outTradeNo: '',
    clientSecret: '',
    intentId: '',
    currency: '',
    countryCode: '',
    paymentEnv: '',
    payAmount: 0,
    orderType: '',
    paymentMode: '',
    resumeToken: '',
    createdAt: 0,
  }
}

export function buildPaymentResultRedirectQuery(state: Pick<PaymentRecoverySnapshot, 'orderId' | 'outTradeNo' | 'resumeToken'>) {
  const query: Record<string, string | undefined> = {}
  if (state.orderId > 0) query.order_id = String(state.orderId)
  if (state.outTradeNo) query.out_trade_no = state.outTradeNo
  if (state.resumeToken) query.resume_token = state.resumeToken
  return query
}

export function buildWechatPaymentAuthorizeUrl(
  authorizeUrl: string,
  purchasePath: string,
  context: { paymentType: string; orderType: OrderType; planId?: number; orderAmount: number },
  origin: string,
): string {
  const normalizedUrl = authorizeUrl.trim()
  if (!normalizedUrl) {
    return normalizedUrl
  }

  try {
    const targetUrl = new URL(normalizedUrl, origin)
    const redirectPath = targetUrl.searchParams.get('redirect') || purchasePath
    const redirectUrl = new URL(redirectPath, origin)
    const paymentType = context.paymentType.trim()
      ? normalizeVisibleMethod(context.paymentType) || context.paymentType.trim()
      : ''

    if (paymentType) redirectUrl.searchParams.set('payment_type', paymentType)
    else redirectUrl.searchParams.delete('payment_type')

    redirectUrl.searchParams.set('order_type', context.orderType)

    if (context.planId) redirectUrl.searchParams.set('plan_id', String(context.planId))
    else redirectUrl.searchParams.delete('plan_id')

    if (context.orderAmount > 0) redirectUrl.searchParams.set('amount', String(context.orderAmount))
    else redirectUrl.searchParams.delete('amount')

    targetUrl.searchParams.set('redirect', `${redirectUrl.pathname}${redirectUrl.search}`)
    return targetUrl.toString()
  } catch {
    return normalizedUrl
  }
}
