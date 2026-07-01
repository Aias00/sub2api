import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'

import {
  DEFAULT_PAYMENT_RESULT_MAX_REFRESH_ATTEMPTS,
  DEFAULT_PAYMENT_RESULT_REFRESH_INTERVAL_MS,
  DEFAULT_PAYMENT_STATUS_POLL_INTERVAL_MS,
  DEFAULT_STRIPE_CLOSE_DELAY_MS,
  DEFAULT_STRIPE_POPUP_INIT_TIMEOUT_MS,
  interpolatePaymentShellLabel,
  paymentMethodLabelKeys,
  paymentStatusPanelLabelKeys,
  renderAirwallexPaymentText,
  renderOrderTableText,
  renderPaymentQRDialogText,
  renderPaymentQRText,
  renderPaymentResultText,
  renderPaymentStatusPanelText,
  renderPaymentViewText,
  renderStripeInlineText,
  renderStripePaymentText,
  renderStripePopupText,
  renderSubscriptionText,
  renderWechatPaymentCallbackText,
  resolveAirwallexPaymentLabels,
  resolveOrderTableLabels,
  resolvePaymentQRDialogLabels,
  resolvePaymentQRLabels,
  resolvePaymentResultDefaults,
  resolvePaymentResultLabels,
  resolvePaymentShellLabels,
  resolvePaymentStatusPollingDefaults,
  resolvePaymentViewLabels,
  resolveStripeInlineLabels,
  resolveStripePaymentLabels,
  resolveStripePaymentRuntimeDefaults,
  resolveStripePopupLabels,
  resolveSubscriptionLabels,
  resolveWechatPaymentCallbackLabels,
  subscriptionPlanCardLabelKeys,
} from '../paymentShell'

describe('paymentShell', () => {
  const allowedKeys = ['success', 'orderId', 'amountTooLow'] as const

  it('resolves locale-scoped payment shell labels and filters unknown keys', () => {
    const labels = resolvePaymentShellLabels(
      JSON.stringify({
        zh: {
          labels: {
            success: '配置成功',
            orderId: '配置订单',
            ignored: '不应出现',
          },
        },
      }),
      'zh-CN',
      allowedKeys,
    )

    expect(labels.success).toBe('配置成功')
    expect(labels.orderId).toBe('配置订单')
    expect(labels).not.toHaveProperty('ignored')
  })

  it('falls back to English labels when the runtime locale is unavailable', () => {
    const labels = resolvePaymentShellLabels(
      JSON.stringify({
        en: {
          labels: {
            success: 'Configured success',
          },
        },
      }),
      'fr',
      allowedKeys,
    )

    expect(labels.success).toBe('Configured success')
  })

  it('returns an empty label map for invalid payment shell config', () => {
    expect(resolvePaymentShellLabels('{bad json', 'zh', allowedKeys)).toEqual({})
  })

  it('interpolates known parameters while preserving unknown placeholders', () => {
    expect(interpolatePaymentShellLabel('Amount must be at least {min}, got {actual}', { min: 10 })).toBe(
      'Amount must be at least 10, got {actual}',
    )
  })

  it('resolves and renders QR dialog labels through the payment shell contract', () => {
    const labels = resolvePaymentQRDialogLabels(JSON.stringify({
      zh: {
        labels: {
          payInNewWindow: '配置新窗口支付',
          ignored: '不应出现',
        },
      },
    }), 'zh-CN')

    expect(renderPaymentQRDialogText(labels, 'payInNewWindow')).toBe('配置新窗口支付')
    expect(labels).not.toHaveProperty('ignored')
  })

  it('resolves and renders Stripe inline labels through the payment shell contract', () => {
    const labels = resolveStripeInlineLabels(JSON.stringify({
      en: {
        labels: {
          stripePay: 'Configured Stripe pay',
          ignored: 'should not appear',
        },
      },
    }), 'en-US')

    expect(renderStripeInlineText(labels, 'stripePay')).toBe('Configured Stripe pay')
    expect(labels).not.toHaveProperty('ignored')
  })

  it('resolves and renders order table labels through the payment shell contract', () => {
    const labels = resolveOrderTableLabels(JSON.stringify({
      zh: {
        labels: {
          methodStripe: '配置 Stripe',
          ignored: '不应出现',
        },
      },
    }), 'zh-CN')

    expect(renderOrderTableText(labels, 'methodStripe')).toBe('配置 Stripe')
    expect(labels).not.toHaveProperty('ignored')
  })

  it('resolves and renders Stripe popup labels through the payment shell contract', () => {
    const labels = resolveStripePopupLabels(JSON.stringify({
      zh: {
        labels: {
          stripePopupRedirecting: '配置跳转中',
          ignored: '不应出现',
        },
      },
    }), 'zh-CN')

    expect(renderStripePopupText(labels, 'stripePopupRedirecting')).toBe('配置跳转中')
    expect(labels).not.toHaveProperty('ignored')
  })

  it('resolves and renders Airwallex payment labels through the payment shell contract', () => {
    const labels = resolveAirwallexPaymentLabels(JSON.stringify({
      zh: {
        labels: {
          airwallexMissingParams: '配置缺参',
          ignored: '不应出现',
        },
      },
    }), 'zh-CN')

    expect(renderAirwallexPaymentText(labels, 'airwallexMissingParams')).toBe('配置缺参')
    expect(labels).not.toHaveProperty('ignored')
  })

  it('resolves and renders QR page labels through the payment shell contract', () => {
    const labels = resolvePaymentQRLabels(JSON.stringify({
      zh: {
        labels: {
          scanWxpay: '配置微信扫码',
          ignored: '不应出现',
        },
      },
    }), 'zh-CN')

    expect(renderPaymentQRText(labels, 'scanWxpay')).toBe('配置微信扫码')
    expect(labels).not.toHaveProperty('ignored')
  })

  it('resolves and renders payment result labels through the payment shell contract', () => {
    const labels = resolvePaymentResultLabels(JSON.stringify({
      en: {
        labels: {
          processingHint: 'Configured processing hint',
          ignored: 'should not appear',
        },
      },
    }), 'en-US')

    expect(renderPaymentResultText(labels, 'processingHint')).toBe('Configured processing hint')
    expect(labels).not.toHaveProperty('ignored')
  })

  it('resolves payment result polling defaults through the payment shell contract', () => {
    expect(resolvePaymentResultDefaults(JSON.stringify({
      en: {
        defaults: {
          paymentResultRefreshIntervalMs: 1234,
          paymentResultMaxRefreshAttempts: 4,
        },
      },
    }), 'en-US')).toEqual({
      refreshIntervalMs: 1234,
      maxRefreshAttempts: 4,
    })

    expect(resolvePaymentResultDefaults(undefined, 'en-US')).toEqual({
      refreshIntervalMs: DEFAULT_PAYMENT_RESULT_REFRESH_INTERVAL_MS,
      maxRefreshAttempts: DEFAULT_PAYMENT_RESULT_MAX_REFRESH_ATTEMPTS,
    })
    expect(resolvePaymentResultDefaults(JSON.stringify({
      en: {
        defaults: {
          paymentResultRefreshIntervalMs: 0,
          paymentResultMaxRefreshAttempts: 101,
        },
      },
    }), 'en-US')).toEqual({
      refreshIntervalMs: DEFAULT_PAYMENT_RESULT_REFRESH_INTERVAL_MS,
      maxRefreshAttempts: DEFAULT_PAYMENT_RESULT_MAX_REFRESH_ATTEMPTS,
    })
  })

  it('resolves shared payment status polling defaults through the payment shell contract', () => {
    expect(resolvePaymentStatusPollingDefaults(JSON.stringify({
      zh: {
        defaults: {
          paymentStatusPollIntervalMs: 4321,
        },
      },
    }), 'zh-CN')).toEqual({
      pollIntervalMs: 4321,
    })

    expect(resolvePaymentStatusPollingDefaults(undefined, 'zh-CN')).toEqual({
      pollIntervalMs: DEFAULT_PAYMENT_STATUS_POLL_INTERVAL_MS,
    })
    expect(resolvePaymentStatusPollingDefaults(JSON.stringify({
      zh: {
        defaults: {
          paymentStatusPollIntervalMs: 0,
        },
      },
    }), 'zh-CN')).toEqual({
      pollIntervalMs: DEFAULT_PAYMENT_STATUS_POLL_INTERVAL_MS,
    })
  })

  it('resolves Stripe runtime defaults through the payment shell contract', () => {
    expect(resolveStripePaymentRuntimeDefaults(JSON.stringify({
      en: {
        defaults: {
          stripePollIntervalMs: 1111,
          stripeCloseDelayMs: 2222,
          stripePopupInitTimeoutMs: 3333,
        },
      },
    }), 'en-US')).toEqual({
      pollIntervalMs: 1111,
      closeDelayMs: 2222,
      popupInitTimeoutMs: 3333,
    })

    expect(resolveStripePaymentRuntimeDefaults(undefined, 'en-US')).toEqual({
      pollIntervalMs: DEFAULT_PAYMENT_STATUS_POLL_INTERVAL_MS,
      closeDelayMs: DEFAULT_STRIPE_CLOSE_DELAY_MS,
      popupInitTimeoutMs: DEFAULT_STRIPE_POPUP_INIT_TIMEOUT_MS,
    })
    expect(resolveStripePaymentRuntimeDefaults(JSON.stringify({
      en: {
        defaults: {
          stripePollIntervalMs: 0,
          stripeCloseDelayMs: 0,
          stripePopupInitTimeoutMs: 200000,
        },
      },
    }), 'en-US')).toEqual({
      pollIntervalMs: DEFAULT_PAYMENT_STATUS_POLL_INTERVAL_MS,
      closeDelayMs: DEFAULT_STRIPE_CLOSE_DELAY_MS,
      popupInitTimeoutMs: DEFAULT_STRIPE_POPUP_INIT_TIMEOUT_MS,
    })
  })

  it('resolves and renders WeChat payment callback labels through the payment shell contract', () => {
    const labels = resolveWechatPaymentCallbackLabels(JSON.stringify({
      zh: {
        labels: {
          wechatPaymentCallbackTitle: '配置微信支付回调',
          ignored: '不应出现',
        },
      },
    }), 'zh-CN')

    expect(renderWechatPaymentCallbackText(labels, 'wechatPaymentCallbackTitle')).toBe('配置微信支付回调')
    expect(labels).not.toHaveProperty('ignored')
  })

  it('resolves and renders subscription labels through the payment shell contract', () => {
    const labels = resolveSubscriptionLabels(JSON.stringify({
      zh: {
        labels: {
          subscriptionResetIn: '配置 {time} 后重置',
          ignored: '不应出现',
        },
      },
    }), 'zh-CN')

    expect(renderSubscriptionText(labels, 'subscriptionResetIn', { time: '2h' })).toBe('配置 2h 后重置')
    expect(labels).not.toHaveProperty('ignored')
  })

  it('falls back to localized subscription labels when shell config is absent', () => {
    const zhLabels = resolveSubscriptionLabels(undefined, 'zh-CN')
    expect(renderSubscriptionText(zhLabels, 'subscriptionNoActive')).toBe('暂无有效订阅')
    expect(renderSubscriptionText(zhLabels, 'subscriptionNoActiveDesc')).toContain('没有有效订阅')
    expect(renderSubscriptionText(zhLabels, 'renewNow')).toBe('续费')
    expect(renderSubscriptionText(zhLabels, 'subscriptionStatusActive')).toBe('有效')

    const enLabels = resolveSubscriptionLabels(undefined, 'en-US')
    expect(renderSubscriptionText(enLabels, 'subscriptionNoActive')).toBe('No Active Subscriptions')
    expect(renderSubscriptionText(enLabels, 'renewNow')).toBe('Renew')
  })

  it('resolves and renders Stripe payment labels through the payment shell contract', () => {
    const labels = resolveStripePaymentLabels(JSON.stringify({
      zh: {
        labels: {
          stripeNotConfigured: '配置 Stripe 未配置',
          ignored: '不应出现',
        },
      },
    }), 'zh-CN')

    expect(renderStripePaymentText(labels, 'stripeNotConfigured')).toBe('配置 Stripe 未配置')
    expect(labels).not.toHaveProperty('ignored')
  })

  it('resolves and renders payment view labels through the payment shell contract', () => {
    const labels = resolvePaymentViewLabels(JSON.stringify({
      zh: {
        labels: {
          amountTooLow: '最低 {min}',
          ignored: '不应出现',
        },
      },
    }), 'zh-CN')

    expect(renderPaymentViewText(labels, 'amountTooLow', { min: 5 })).toBe('最低 5')
    expect(labels).not.toHaveProperty('ignored')
  })

  it('falls back to localized payment view labels when shell config is absent', () => {
    const zhLabels = resolvePaymentViewLabels(undefined, 'zh-CN')
    expect(renderPaymentViewText(zhLabels, 'tabTopUp')).toBe('充值')
    expect(renderPaymentViewText(zhLabels, 'tabSubscribe')).toBe('订阅')
    expect(renderPaymentViewText(zhLabels, 'rechargeAccount')).toBe('充值账户')
    expect(renderPaymentViewText(zhLabels, 'currentBalance')).toBe('当前余额')
    expect(renderPaymentViewText(zhLabels, 'paymentMethod')).toBe('支付方式')
    expect(renderPaymentViewText(zhLabels, 'methodStripe')).toBe('Stripe')
    expect(renderPaymentViewText(zhLabels, 'rechargeProductCta')).toBe('立即充值')

    const enLabels = resolvePaymentViewLabels(undefined, 'en-US')
    expect(renderPaymentViewText(enLabels, 'tabTopUp')).toBe('Top Up')
    expect(renderPaymentViewText(enLabels, 'paymentMethod')).toBe('Payment Method')
  })

  it('centralizes payment component label schemas', () => {
    expect(paymentStatusPanelLabelKeys).toContain('waitingPayment')
    expect(paymentStatusPanelLabelKeys).toContain('subscriptionSuccess')
    expect(renderPaymentStatusPanelText({ waitingPayment: 'Configured waiting' }, 'waitingPayment')).toBe(
      'Configured waiting',
    )
    expect(subscriptionPlanCardLabelKeys).toContain('subscribeNow')
    expect(subscriptionPlanCardLabelKeys).toContain('perYear')
    expect(paymentMethodLabelKeys).toEqual(['alipay', 'wxpay', 'stripe', 'airwallex'])
  })

  it('keeps payment components from owning local label interfaces', () => {
    const componentFiles = [
      'src/components/payment/OrderTable.vue',
      'src/components/payment/PaymentQRDialog.vue',
      'src/components/payment/PaymentStatusPanel.vue',
      'src/components/payment/StripePaymentInline.vue',
      'src/components/payment/SubscriptionPlanCard.vue',
    ]

    for (const file of componentFiles) {
      const source = readFileSync(file, 'utf8')
      expect(source).not.toMatch(/interface\s+\w*Labels/)
      expect(source).not.toMatch(/type\s+\w*LabelKey\s*=/)
      expect(source).not.toMatch(/Partial<Record<\w*LabelKey,\s*string>>/)
      expect(source).toContain("from '@/utils/paymentShell'")
    }
  })

  it('keeps payment method selector labels typed through the payment shell contract', () => {
    const selectorSource = readFileSync('src/components/payment/PaymentMethodSelector.vue', 'utf8')
    const paymentViewSource = readFileSync('src/views/user/PaymentView.vue', 'utf8')

    expect(selectorSource).toContain('type PaymentMethodLabels')
    expect(selectorSource).toContain('methodLabels?: PaymentMethodLabels')
    expect(selectorSource).not.toContain('methodLabels?: Record<string, string>')
    expect(paymentViewSource).toContain('type PaymentMethodLabels')
    expect(paymentViewSource).toContain('computed<PaymentMethodLabels>')
  })
})
