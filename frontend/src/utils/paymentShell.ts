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
  return {
    ...defaultSubscriptionLabels(runtimeLocale),
    ...resolvePaymentShellLabels(raw, runtimeLocale, subscriptionLabelKeys),
  }
}

export function renderSubscriptionText(
  labels: SubscriptionLabels | undefined,
  key: SubscriptionLabelKey,
  params?: Record<string, string | number>,
): string {
  const label = labels?.[key] || key
  return interpolatePaymentShellLabel(label, params)
}

function defaultSubscriptionLabels(runtimeLocale: string): SubscriptionLabels {
  if (runtimeLocale.toLowerCase().startsWith('zh')) {
    return {
      renewNow: '续费',
      subscriptionNoActive: '暂无有效订阅',
      subscriptionNoActiveDesc: '您当前没有有效订阅，可前往充值/订阅页面购买套餐。',
      subscriptionExpires: '到期时间',
      subscriptionNoExpiration: '无到期时间',
      subscriptionStatusActive: '有效',
      subscriptionStatusExpired: '已过期',
      subscriptionStatusRevoked: '已撤销',
      subscriptionDaily: '每日',
      subscriptionWeekly: '每周',
      subscriptionMonthly: '每月',
      subscriptionUnlimited: '无限制',
      subscriptionUnlimitedDesc: '该订阅无用量限制',
      subscriptionDaysRemaining: '剩余 {days} 天',
      subscriptionResetIn: '{time} 后重置',
      subscriptionQuotaEndsIn: '额度将在 {time} 后重置',
      subscriptionWindowNotActive: '等待首次使用',
      subscriptionToday: '今天',
      subscriptionTomorrow: '明天',
      subscriptionFailedToLoad: '加载订阅失败',
    }
  }

  return {
    renewNow: 'Renew',
    subscriptionNoActive: 'No Active Subscriptions',
    subscriptionNoActiveDesc: 'You do not have an active subscription. Go to Recharge / Subscription to purchase a plan.',
    subscriptionExpires: 'Expires',
    subscriptionNoExpiration: 'No expiration',
    subscriptionStatusActive: 'Active',
    subscriptionStatusExpired: 'Expired',
    subscriptionStatusRevoked: 'Revoked',
    subscriptionDaily: 'Daily',
    subscriptionWeekly: 'Weekly',
    subscriptionMonthly: 'Monthly',
    subscriptionUnlimited: 'Unlimited',
    subscriptionUnlimitedDesc: 'No usage limits on this subscription',
    subscriptionDaysRemaining: '{days} days remaining',
    subscriptionResetIn: 'Resets in {time}',
    subscriptionQuotaEndsIn: 'Quota resets in {time}',
    subscriptionWindowNotActive: 'Awaiting first use',
    subscriptionToday: 'Today',
    subscriptionTomorrow: 'Tomorrow',
    subscriptionFailedToLoad: 'Failed to load subscriptions',
  }
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
  return {
    ...defaultPaymentViewLabels(runtimeLocale),
    ...resolvePaymentShellLabels(raw, runtimeLocale, paymentViewLabelKeys),
  }
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

function defaultPaymentViewLabels(runtimeLocale: string): PaymentViewLabels {
  if (runtimeLocale.toLowerCase().startsWith('zh')) {
    return {
      tabTopUp: '充值',
      tabSubscribe: '订阅',
      rechargeAccount: '充值账户',
      currentBalance: '当前余额',
      notAvailable: '支付暂不可用',
      noRechargeProducts: '暂无可用充值商品',
      rechargeProductRecommended: '推荐',
      rechargeProductCreditLine: '获得 ${amount} 余额',
      rechargeProductCta: '立即充值',
      paymentMethod: '支付方式',
      methodAlipay: '支付宝',
      methodWxpay: '微信支付',
      methodStripe: 'Stripe',
      methodAirwallex: 'Airwallex',
      success: '支付成功',
      subscriptionSuccess: '订阅成功',
      orderId: '订单 ID',
      orderNo: '订单号',
      amount: '金额',
      payAmount: '支付金额',
      confirm: '确认',
      cancelled: '已取消',
      cancelledDesc: '您已取消本次支付',
      expired: '订单已过期',
      expiredDesc: '订单已超时，请重新创建订单',
      scanAlipay: '支付宝扫码支付',
      scanAlipayHint: '请使用手机打开支付宝，扫描二维码完成支付',
      scanWxpay: '微信扫码支付',
      scanWxpayHint: '请使用手机打开微信，扫描二维码完成支付',
      scanToPay: '请扫码支付',
      openPayWindow: '重新打开支付页面',
      expiresIn: '剩余支付时间',
      waitingPayment: '等待支付...',
      cancelOrder: '取消订单',
      payInNewWindowHint: '支付页面已在新窗口打开，请在新窗口中完成支付后返回此页面',
      paymentAmount: '支付金额',
      fee: '手续费',
      actualPay: '实付金额',
      creditedBalance: '到账余额',
      rechargeRatePreview: '当前倍率：1 CNY = {usd} USD',
      processing: '处理中...',
      createOrder: '创建订单',
      cancel: '取消',
      selectAmountFirst: '请选择充值商品',
      amountNoMethod: '请选择支付方式',
      amountTooLow: '最低 {min}',
      amountTooHigh: '最高 {max}',
      amountLabel: '金额',
      noPlans: '暂无可用订阅套餐',
      activeSubscription: '当前订阅',
      selectPlan: '选择套餐',
      groupFallback: '分组 #{id}',
      daysRemaining: '剩余 {days} 天',
      noExpiration: '永不过期',
      activeStatus: '生效中',
      rate: '倍率',
      dailyLimit: '每日额度',
      weeklyLimit: '每周额度',
      monthlyLimit: '每月额度',
      quota: '额度',
      unlimited: '不限量',
      models: '模型',
      subscribeNow: '立即开通',
      renewNow: '续费',
      perMonth: '月',
      perYear: '年',
      days: '天',
      failed: '支付失败',
      errorFallback: '支付请求失败，请稍后重试',
      tooManyPending: '待支付订单过多，请先完成或取消已有订单',
      cancelRateLimited: '取消过于频繁，请稍后再试',
      mobilePaymentFallbackToQr: '当前环境无法直接拉起支付，已切换为扫码支付。',
    }
  }

  return {
    tabTopUp: 'Top Up',
    tabSubscribe: 'Subscribe',
    rechargeAccount: 'Recharge Account',
    currentBalance: 'Current Balance',
    notAvailable: 'Payment is not available',
    noRechargeProducts: 'No recharge products available',
    rechargeProductRecommended: 'Recommended',
    rechargeProductCreditLine: 'Get ${amount} balance',
    rechargeProductCta: 'Top up now',
    paymentMethod: 'Payment Method',
    methodAlipay: 'Alipay',
    methodWxpay: 'WeChat Pay',
    methodStripe: 'Stripe',
    methodAirwallex: 'Airwallex',
    success: 'Payment Successful',
    subscriptionSuccess: 'Subscription Successful',
    orderId: 'Order ID',
    orderNo: 'Order No.',
    amount: 'Amount',
    payAmount: 'Payment Amount',
    confirm: 'Confirm',
    cancelled: 'Cancelled',
    cancelledDesc: 'You cancelled this payment.',
    expired: 'Order Expired',
    expiredDesc: 'This order has expired. Please create a new one.',
    scanAlipay: 'Alipay QR Payment',
    scanAlipayHint: 'Open Alipay on your phone and scan the QR code to pay',
    scanWxpay: 'WeChat QR Payment',
    scanWxpayHint: 'Open WeChat on your phone and scan the QR code to pay',
    scanToPay: 'Scan to Pay',
    openPayWindow: 'Reopen Payment Page',
    expiresIn: 'Expires in',
    waitingPayment: 'Waiting for payment...',
    cancelOrder: 'Cancel Order',
    payInNewWindowHint: 'The payment page has opened in a new window. Please complete the payment there and return to this page.',
    paymentAmount: 'Payment Amount',
    fee: 'Fee',
    actualPay: 'Actual Pay',
    creditedBalance: 'Credited Balance',
    rechargeRatePreview: 'Current rate: 1 CNY = {usd} USD',
    processing: 'Processing...',
    createOrder: 'Create Order',
    cancel: 'Cancel',
    selectAmountFirst: 'Select a recharge product',
    amountNoMethod: 'Select a payment method',
    amountTooLow: 'Minimum {min}',
    amountTooHigh: 'Maximum {max}',
    amountLabel: 'Amount',
    noPlans: 'No subscription plans available',
    activeSubscription: 'Active Subscription',
    selectPlan: 'Select Plan',
    groupFallback: 'Group #{id}',
    daysRemaining: '{days} days remaining',
    noExpiration: 'No expiration',
    activeStatus: 'Active',
    rate: 'Rate',
    dailyLimit: 'Daily Limit',
    weeklyLimit: 'Weekly Limit',
    monthlyLimit: 'Monthly Limit',
    quota: 'Quota',
    unlimited: 'Unlimited',
    models: 'Models',
    subscribeNow: 'Subscribe Now',
    renewNow: 'Renew',
    perMonth: 'month',
    perYear: 'year',
    days: 'days',
    failed: 'Payment Failed',
    errorFallback: 'Payment request failed. Please try again later.',
    tooManyPending: 'Too many pending orders. Please complete or cancel an existing order first.',
    cancelRateLimited: 'Cancellation is too frequent. Please try again later.',
    mobilePaymentFallbackToQr: 'Direct payment is unavailable in this environment. QR payment is shown instead.',
  }
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
