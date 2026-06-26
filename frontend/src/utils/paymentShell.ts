export function resolvePaymentShellLabels<K extends string>(
  raw: string | undefined,
  runtimeLocale: string,
  allowedKeys: readonly K[],
): Partial<Record<K, string>> {
  if (!raw) return {}
  try {
    const parsed = JSON.parse(raw) as unknown
    if (!isRecord(parsed)) return {}

    const normalizedLocale = runtimeLocale.toLowerCase()
    const localeKeys = [normalizedLocale, normalizedLocale.split('-')[0], 'en', 'zh']
    for (const key of localeKeys) {
      const localized = parsed[key]
      if (!isRecord(localized) || !isRecord(localized.labels)) continue
      return pickPaymentShellLabels(localized.labels, allowedKeys)
    }
  } catch {
    return {}
  }
  return {}
}

export function interpolatePaymentShellLabel(
  label: string,
  params?: Record<string, string | number>,
): string {
  if (!params) return label
  return label.replace(/\{(\w+)\}/g, (_match, key: string) =>
    Object.prototype.hasOwnProperty.call(params, key) ? String(params[key]) : `{${key}}`,
  )
}

export const paymentQRDialogLabelKeys = [
  'payInNewWindow',
  'payInNewWindowHint',
  'scanAlipay',
  'scanAlipayHint',
  'scanWxpay',
  'scanWxpayHint',
  'scanToPay',
  'openPayWindow',
  'expired',
  'expiresIn',
  'waitingPayment',
  'cancelOrder',
  'success',
  'errorFallback',
  'orderId',
  'amount',
  'payAmount',
  'processing',
  'confirm',
  'backToRecharge',
] as const

export type PaymentQRDialogLabelKey = typeof paymentQRDialogLabelKeys[number]
export type PaymentQRDialogLabels = Partial<Record<PaymentQRDialogLabelKey, string>>

export function resolvePaymentQRDialogLabels(raw: string | undefined, runtimeLocale: string): PaymentQRDialogLabels {
  return resolvePaymentShellLabels(raw, runtimeLocale, paymentQRDialogLabelKeys)
}

export function renderPaymentQRDialogText(labels: PaymentQRDialogLabels | undefined, key: PaymentQRDialogLabelKey): string {
  return labels?.[key] || ''
}

export const stripeInlineLabelKeys = [
  'actualPay',
  'orderId',
  'amount',
  'payAmount',
  'confirm',
  'success',
  'processing',
  'backToRecharge',
  'cancelOrder',
  'failed',
  'errorFallback',
  'stripePay',
  'stripeLoadFailed',
] as const

export type StripeInlineLabelKey = typeof stripeInlineLabelKeys[number]
export type StripeInlineLabels = Partial<Record<StripeInlineLabelKey, string>>

export function resolveStripeInlineLabels(raw: string | undefined, runtimeLocale: string): StripeInlineLabels {
  return resolvePaymentShellLabels(raw, runtimeLocale, stripeInlineLabelKeys)
}

export function renderStripeInlineText(labels: StripeInlineLabels | undefined, key: StripeInlineLabelKey): string {
  return labels?.[key] || ''
}

export const orderTableLabelKeys = [
  'orderId',
  'orderNo',
  'payAmount',
  'paymentMethod',
  'status',
  'createdAt',
  'actions',
  'fee',
  'creditedAmount',
  'user',
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

export type OrderTableLabelKey = typeof orderTableLabelKeys[number]
export type OrderTableLabels = Partial<Record<OrderTableLabelKey, string>>

export function resolveOrderTableLabels(raw: string | undefined, runtimeLocale: string): OrderTableLabels {
  return resolvePaymentShellLabels(raw, runtimeLocale, orderTableLabelKeys)
}

export function renderOrderTableText(labels: OrderTableLabels | undefined, key: OrderTableLabelKey): string {
  return labels?.[key] || ''
}

export const paymentMethodLabelKeys = [
  'alipay',
  'wxpay',
  'stripe',
  'airwallex',
] as const

export type PaymentMethodLabelKey = typeof paymentMethodLabelKeys[number]
export type PaymentMethodLabels = Partial<Record<PaymentMethodLabelKey, string>>

export const stripePopupLabelKeys = [
  'orderId',
  'close',
  'success',
  'failed',
  'stripeLoadFailed',
  'stripeMissingParams',
  'stripePopupLoadingQr',
  'stripePopupRedirecting',
  'stripePopupTimeout',
] as const

export type StripePopupLabelKey = typeof stripePopupLabelKeys[number]
export type StripePopupLabels = Partial<Record<StripePopupLabelKey, string>>

export function resolveStripePopupLabels(raw: string | undefined, runtimeLocale: string): StripePopupLabels {
  return resolvePaymentShellLabels(raw, runtimeLocale, stripePopupLabelKeys)
}

export function renderStripePopupText(labels: StripePopupLabels | undefined, key: StripePopupLabelKey): string {
  return labels?.[key] || ''
}

export const airwallexPaymentLabelKeys = [
  'airwallexLoadFailed',
  'airwallexMissingParams',
  'backToRecharge',
  'payInNewWindowHint',
] as const

export type AirwallexPaymentLabelKey = typeof airwallexPaymentLabelKeys[number]
export type AirwallexPaymentLabels = Partial<Record<AirwallexPaymentLabelKey, string>>

export function resolveAirwallexPaymentLabels(raw: string | undefined, runtimeLocale: string): AirwallexPaymentLabels {
  return resolvePaymentShellLabels(raw, runtimeLocale, airwallexPaymentLabelKeys)
}

export function renderAirwallexPaymentText(
  labels: AirwallexPaymentLabels | undefined,
  key: AirwallexPaymentLabelKey,
): string {
  return labels?.[key] || ''
}

export const paymentQRLabelKeys = [
  'payInNewWindow',
  'payInNewWindowHint',
  'scanAlipay',
  'scanAlipayHint',
  'scanWxpay',
  'scanWxpayHint',
  'scanToPay',
  'openPayWindow',
  'expired',
  'expiresIn',
  'waitingPayment',
  'cancelOrder',
  'processing',
  'backToRecharge',
  'errorFallback',
] as const

export type PaymentQRLabelKey = typeof paymentQRLabelKeys[number]
export type PaymentQRLabels = Partial<Record<PaymentQRLabelKey, string>>

export function resolvePaymentQRLabels(raw: string | undefined, runtimeLocale: string): PaymentQRLabels {
  return resolvePaymentShellLabels(raw, runtimeLocale, paymentQRLabelKeys)
}

export function renderPaymentQRText(labels: PaymentQRLabels | undefined, key: PaymentQRLabelKey): string {
  return labels?.[key] || ''
}

export const paymentResultLabelKeys = [
  'success',
  'processing',
  'failed',
  'processingHint',
  'backToRecharge',
  'viewOrders',
  'orderId',
  'orderNo',
  'baseAmount',
  'fee',
  'payAmount',
  'creditedAmount',
  'paymentMethod',
  'status',
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

export type PaymentResultLabelKey = typeof paymentResultLabelKeys[number]
export type PaymentResultLabels = Partial<Record<PaymentResultLabelKey, string>>

export function resolvePaymentResultLabels(raw: string | undefined, runtimeLocale: string): PaymentResultLabels {
  return resolvePaymentShellLabels(raw, runtimeLocale, paymentResultLabelKeys)
}

export function renderPaymentResultText(labels: PaymentResultLabels | undefined, key: PaymentResultLabelKey): string {
  return labels?.[key] || ''
}

export type PaymentResultDefaults = {
  refreshIntervalMs: number
  maxRefreshAttempts: number
}

export const DEFAULT_PAYMENT_RESULT_REFRESH_INTERVAL_MS = 2000
export const DEFAULT_PAYMENT_RESULT_MAX_REFRESH_ATTEMPTS = 15
export const DEFAULT_PAYMENT_STATUS_POLL_INTERVAL_MS = 3000
export const DEFAULT_PAYMENT_VERIFY_RETRY_INTERVAL_MS = 15000
export const DEFAULT_PAYMENT_VERIFY_RETRY_MAX_ATTEMPTS = 6
export const DEFAULT_STRIPE_CLOSE_DELAY_MS = 2000
export const DEFAULT_STRIPE_POPUP_INIT_TIMEOUT_MS = 15000

export function resolvePaymentResultDefaults(
  raw: string | undefined,
  runtimeLocale: string,
): PaymentResultDefaults {
  const defaults = pickPaymentShellDefaults(raw, runtimeLocale)
  return {
    refreshIntervalMs: readPositiveIntegerDefault(
      defaults?.paymentResultRefreshIntervalMs,
      DEFAULT_PAYMENT_RESULT_REFRESH_INTERVAL_MS,
      60_000,
    ),
    maxRefreshAttempts: readPositiveIntegerDefault(
      defaults?.paymentResultMaxRefreshAttempts,
      DEFAULT_PAYMENT_RESULT_MAX_REFRESH_ATTEMPTS,
      100,
    ),
  }
}

export type PaymentStatusPollingDefaults = {
  pollIntervalMs: number
}

export function resolvePaymentStatusPollingDefaults(
  raw: string | undefined,
  runtimeLocale: string,
): PaymentStatusPollingDefaults {
  const defaults = pickPaymentShellDefaults(raw, runtimeLocale)
  return {
    pollIntervalMs: readPositiveIntegerDefault(
      defaults?.paymentStatusPollIntervalMs,
      DEFAULT_PAYMENT_STATUS_POLL_INTERVAL_MS,
      60_000,
    ),
  }
}

export type PaymentVerifyRetryDefaults = {
  verifyRetryIntervalMs: number
  verifyRetryMaxAttempts: number
}

export function resolvePaymentVerifyRetryDefaults(
  raw: string | undefined,
  runtimeLocale: string,
): PaymentVerifyRetryDefaults {
  const defaults = pickPaymentShellDefaults(raw, runtimeLocale)
  return {
    verifyRetryIntervalMs: readPositiveIntegerDefault(
      defaults?.paymentVerifyRetryIntervalMs,
      DEFAULT_PAYMENT_VERIFY_RETRY_INTERVAL_MS,
      120_000,
    ),
    verifyRetryMaxAttempts: readPositiveIntegerDefault(
      defaults?.paymentVerifyRetryMaxAttempts,
      DEFAULT_PAYMENT_VERIFY_RETRY_MAX_ATTEMPTS,
      100,
    ),
  }
}

export type StripePaymentRuntimeDefaults = {
  pollIntervalMs: number
  closeDelayMs: number
  popupInitTimeoutMs: number
}

export function resolveStripePaymentRuntimeDefaults(
  raw: string | undefined,
  runtimeLocale: string,
): StripePaymentRuntimeDefaults {
  const defaults = pickPaymentShellDefaults(raw, runtimeLocale)
  return {
    pollIntervalMs: readPositiveIntegerDefault(
      defaults?.stripePollIntervalMs,
      DEFAULT_PAYMENT_STATUS_POLL_INTERVAL_MS,
      60_000,
    ),
    closeDelayMs: readPositiveIntegerDefault(
      defaults?.stripeCloseDelayMs,
      DEFAULT_STRIPE_CLOSE_DELAY_MS,
      60_000,
    ),
    popupInitTimeoutMs: readPositiveIntegerDefault(
      defaults?.stripePopupInitTimeoutMs,
      DEFAULT_STRIPE_POPUP_INIT_TIMEOUT_MS,
      120_000,
    ),
  }
}

export const wechatPaymentCallbackLabelKeys = [
  'wechatPaymentCallbackTitle',
  'wechatPaymentCallbackProcessing',
  'wechatPaymentCallbackBackToPayment',
  'wechatPaymentCallbackMissingResumeToken',
] as const

export type WechatPaymentCallbackLabelKey = typeof wechatPaymentCallbackLabelKeys[number]
export type WechatPaymentCallbackLabels = Partial<Record<WechatPaymentCallbackLabelKey, string>>

export function resolveWechatPaymentCallbackLabels(
  raw: string | undefined,
  runtimeLocale: string,
): WechatPaymentCallbackLabels {
  return resolvePaymentShellLabels(raw, runtimeLocale, wechatPaymentCallbackLabelKeys)
}

export function renderWechatPaymentCallbackText(
  labels: WechatPaymentCallbackLabels | undefined,
  key: WechatPaymentCallbackLabelKey,
): string {
  return labels?.[key] || ''
}

export const subscriptionLabelKeys = [
  'renewNow',
  'subscriptionNoActive',
  'subscriptionNoActiveDesc',
  'subscriptionExpires',
  'subscriptionNoExpiration',
  'subscriptionStatusActive',
  'subscriptionStatusExpired',
  'subscriptionStatusRevoked',
  'subscriptionDaily',
  'subscriptionWeekly',
  'subscriptionMonthly',
  'subscriptionUnlimited',
  'subscriptionUnlimitedDesc',
  'subscriptionDaysRemaining',
  'subscriptionResetIn',
  'subscriptionQuotaEndsIn',
  'subscriptionWindowNotActive',
  'subscriptionToday',
  'subscriptionTomorrow',
  'subscriptionFailedToLoad',
] as const

export type SubscriptionLabelKey = typeof subscriptionLabelKeys[number]
export type SubscriptionLabels = Partial<Record<SubscriptionLabelKey, string>>

export function resolveSubscriptionLabels(raw: string | undefined, runtimeLocale: string): SubscriptionLabels {
  return resolvePaymentShellLabels(raw, runtimeLocale, subscriptionLabelKeys)
}

export function renderSubscriptionText(
  labels: SubscriptionLabels | undefined,
  key: SubscriptionLabelKey,
  params?: Record<string, string | number>,
): string {
  const label = labels?.[key] || key
  return interpolatePaymentShellLabel(label, params)
}

export const stripePaymentLabelKeys = [
  'actualPay',
  'scanWxpay',
  'scanWxpayHint',
  'waitingPayment',
  'payInNewWindowHint',
  'success',
  'failed',
  'backToRecharge',
  'processing',
  'stripePay',
  'stripeLoadFailed',
  'stripeMissingParams',
  'stripeNotConfigured',
  'stripeSuccessProcessing',
] as const

export type StripePaymentLabelKey = typeof stripePaymentLabelKeys[number]
export type StripePaymentLabels = Partial<Record<StripePaymentLabelKey, string>>

export function resolveStripePaymentLabels(raw: string | undefined, runtimeLocale: string): StripePaymentLabels {
  return resolvePaymentShellLabels(raw, runtimeLocale, stripePaymentLabelKeys)
}

export function renderStripePaymentText(labels: StripePaymentLabels | undefined, key: StripePaymentLabelKey): string {
  return labels?.[key] || ''
}

export const paymentStatusPanelLabelKeys = [
  'success',
  'subscriptionSuccess',
  'orderId',
  'orderNo',
  'amount',
  'payAmount',
  'confirm',
  'cancelled',
  'cancelledDesc',
  'expired',
  'expiredDesc',
  'scanAlipay',
  'scanAlipayHint',
  'scanWxpay',
  'scanWxpayHint',
  'scanToPay',
  'openPayWindow',
  'expiresIn',
  'waitingPayment',
  'processing',
  'cancelOrder',
  'payInNewWindowHint',
  'errorFallback',
] as const

export type PaymentStatusPanelLabelKey = typeof paymentStatusPanelLabelKeys[number]
export type PaymentStatusPanelLabels = Partial<Record<PaymentStatusPanelLabelKey, string>>

export function renderPaymentStatusPanelText(
  labels: PaymentStatusPanelLabels | undefined,
  key: PaymentStatusPanelLabelKey,
): string {
  return labels?.[key] || ''
}

export const subscriptionPlanCardLabelKeys = [
  'rate',
  'dailyLimit',
  'weeklyLimit',
  'monthlyLimit',
  'quota',
  'unlimited',
  'models',
  'subscribeNow',
  'renewNow',
  'perMonth',
  'perYear',
  'days',
] as const

export type SubscriptionPlanCardLabelKey = typeof subscriptionPlanCardLabelKeys[number]
export type SubscriptionPlanCardLabels = Partial<Record<SubscriptionPlanCardLabelKey, string>>

export const paymentViewLabelKeys = [
  'tabTopUp',
  'tabSubscribe',
  'rechargeAccount',
  'currentBalance',
  'notAvailable',
  'noRechargeProducts',
  'rechargeProductRecommended',
  'rechargeProductCreditLine',
  'rechargeProductCta',
  'paymentMethod',
  'methodAlipay',
  'methodWxpay',
  'methodStripe',
  'methodAirwallex',
  'success',
  'subscriptionSuccess',
  'orderId',
  'orderNo',
  'amount',
  'payAmount',
  'confirm',
  'cancelled',
  'cancelledDesc',
  'expired',
  'expiredDesc',
  'scanAlipay',
  'scanAlipayHint',
  'scanWxpay',
  'scanWxpayHint',
  'scanToPay',
  'openPayWindow',
  'expiresIn',
  'waitingPayment',
  'cancelOrder',
  'payInNewWindowHint',
  'paymentAmount',
  'fee',
  'actualPay',
  'creditedBalance',
  'rechargeRatePreview',
  'processing',
  'createOrder',
  'cancel',
  'selectAmountFirst',
  'amountNoMethod',
  'amountTooLow',
  'amountTooHigh',
  'amountLabel',
  'noPlans',
  'activeSubscription',
  'selectPlan',
  'groupFallback',
  'daysRemaining',
  'noExpiration',
  'activeStatus',
  'rate',
  'dailyLimit',
  'weeklyLimit',
  'monthlyLimit',
  'quota',
  'unlimited',
  'models',
  'subscribeNow',
  'renewNow',
  'perMonth',
  'perYear',
  'days',
  'failed',
  'errorFallback',
  'tooManyPending',
  'cancelRateLimited',
  'mobilePaymentFallbackToQr',
] as const

export type PaymentViewLabelKey = typeof paymentViewLabelKeys[number]
export type PaymentViewLabels = Partial<Record<PaymentViewLabelKey, string>>

export function resolvePaymentViewLabels(raw: string | undefined, runtimeLocale: string): PaymentViewLabels {
  return resolvePaymentShellLabels(raw, runtimeLocale, paymentViewLabelKeys)
}

export function renderPaymentViewText(
  labels: PaymentViewLabels | undefined,
  key: PaymentViewLabelKey,
  params?: Record<string, string | number>,
): string {
  const label = labels?.[key] || key
  return interpolatePaymentShellLabel(label, params)
}

function pickPaymentShellLabels<K extends string>(
  labels: Record<string, unknown>,
  allowedKeys: readonly K[],
): Partial<Record<K, string>> {
  const allowed = new Set<string>(allowedKeys)
  const result: Partial<Record<K, string>> = {}
  for (const [labelKey, value] of Object.entries(labels)) {
    if (typeof value === 'string' && allowed.has(labelKey)) {
      result[labelKey as K] = value
    }
  }
  return result
}

function pickPaymentShellDefaults(raw: string | undefined, runtimeLocale: string): Record<string, unknown> | null {
  if (!raw) return null
  try {
    const parsed = JSON.parse(raw) as unknown
    if (!isRecord(parsed)) return null

    const normalizedLocale = runtimeLocale.toLowerCase()
    const localeKeys = [normalizedLocale, normalizedLocale.split('-')[0], 'en', 'zh']
    for (const key of localeKeys) {
      const localized = parsed[key]
      if (isRecord(localized) && isRecord(localized.defaults)) {
        return localized.defaults
      }
    }
  } catch {
    return null
  }
  return null
}

function readPositiveIntegerDefault(value: unknown, fallback: number, max: number): number {
  const normalized = Number(value)
  if (!Number.isInteger(normalized) || normalized <= 0 || normalized > max) {
    return fallback
  }
  return normalized
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}
