import type { LocationQuery, LocationQueryRaw } from 'vue-router'
import { resolveVisiblePaymentMethod } from '@/components/payment/paymentMethod'

export interface ParsedWechatResumeRoute {
  orderAmount: number
  orderType: 'balance' | 'subscription'
  paymentType: string
  planId?: number
  wechatResumeToken?: string
}

function readQueryString(query: LocationQuery, key: string): string {
  const value = query[key]
  if (Array.isArray(value)) {
    return typeof value[0] === 'string' ? value[0] : ''
  }
  return typeof value === 'string' ? value : ''
}

export function hasWechatResumeQuery(query: LocationQuery): boolean {
  if (readQueryString(query, 'wechat_resume') === '1') {
    return true
  }
  return readQueryString(query, 'wechat_resume_token') !== ''
}

export function parseWechatResumeRoute(
  query: LocationQuery,
): ParsedWechatResumeRoute | null {
  if (!hasWechatResumeQuery(query)) {
    return null
  }

  const wechatResumeToken = readQueryString(query, 'wechat_resume_token')
  const paymentType = resolveVisiblePaymentMethod(readQueryString(query, 'payment_type'))
  const planId = Number.parseInt(readQueryString(query, 'plan_id'), 10)
  const hasPlanId = Number.isFinite(planId) && planId > 0
  const orderType = readQueryString(query, 'order_type') === 'subscription' || hasPlanId
    ? 'subscription'
    : 'balance'

  if (!wechatResumeToken) {
    return null
  }

  return {
    wechatResumeToken,
    paymentType,
    orderType,
    orderAmount: 0,
    planId: hasPlanId ? planId : undefined,
  }
}

export function stripWechatResumeQuery(query: LocationQuery): LocationQueryRaw {
  const nextQuery: LocationQueryRaw = { ...query }
  delete nextQuery.wechat_resume
  delete nextQuery.wechat_resume_token
  delete nextQuery.openid
  delete nextQuery.state
  delete nextQuery.scope
  delete nextQuery.payment_type
  delete nextQuery.amount
  delete nextQuery.order_type
  delete nextQuery.plan_id
  return nextQuery
}
