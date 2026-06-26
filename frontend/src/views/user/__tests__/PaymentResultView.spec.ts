import { readFileSync } from 'node:fs'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

const routeState = vi.hoisted(() => ({
  query: {} as Record<string, unknown>,
}))

const routerPush = vi.hoisted(() => vi.fn())
const pollOrderStatus = vi.hoisted(() => vi.fn())
const verifyOrder = vi.hoisted(() => vi.fn())
const verifyOrderPublic = vi.hoisted(() => vi.fn())
const resolveOrderPublicByResumeToken = vi.hoisted(() => vi.fn())
const appStoreState = vi.hoisted(() => ({
  cachedPublicSettings: null as null | { payment_shell_config?: string },
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => routeState,
    useRouter: () => ({ push: routerPush }),
  }
})

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
      locale: { value: 'en-US' },
    }),
  }
})

vi.mock('@/stores/payment', () => ({
  usePaymentStore: () => ({
    pollOrderStatus,
  }),
}))

vi.mock('@/stores', () => ({
  useAppStore: () => appStoreState,
}))

vi.mock('@/api/payment', () => ({
  paymentAPI: {
    verifyOrder,
    verifyOrderPublic,
    resolveOrderPublicByResumeToken,
  },
}))

import PaymentResultView from '../PaymentResultView.vue'
import { PAYMENT_RECOVERY_STORAGE_KEY } from '@/components/payment/paymentFlow'
import { formatPaymentAmount } from '@/components/payment/currency'

const paymentResultViewSource = readFileSync(
  'src/views/user/PaymentResultView.vue',
  'utf8',
)

function buildPaymentResultShellConfig(
  overrides: Record<string, string> = {},
  defaults: Record<string, unknown> = {},
) {
  return JSON.stringify({
    en: {
      labels: {
        success: 'Payment successful',
        processing: 'Payment processing',
        failed: 'Payment failed',
        processingHint: 'Payment is still being confirmed',
        backToRecharge: 'Back to recharge',
        viewOrders: 'View orders',
        orderId: 'Order ID',
        orderNo: 'Order number',
        baseAmount: 'Base amount',
        fee: 'Fee',
        payAmount: 'Paid amount',
        creditedAmount: 'Credited amount',
        paymentMethod: 'Payment method',
        status: 'Status',
        methodAlipay: 'Alipay',
        methodWxpay: 'WeChat Pay',
        methodStripe: 'Stripe',
        methodAirwallex: 'Airwallex',
        statusPending: 'Pending',
        statusPaid: 'Paid',
        statusRecharging: 'Recharging',
        statusCompleted: 'Completed',
        statusExpired: 'Expired',
        statusCancelled: 'Cancelled',
        statusFailed: 'Failed',
        statusRefundRequested: 'Refund requested',
        statusRefunding: 'Refunding',
        statusRefunded: 'Refunded',
        statusPartiallyRefunded: 'Partially refunded',
        statusRefundFailed: 'Refund failed',
        ...overrides,
      },
      defaults,
    },
  })
}

const orderFactory = (status: string) => ({
  id: 42,
  user_id: 9,
  amount: 88,
  pay_amount: 88,
  fee_rate: 0,
  payment_type: 'alipay',
  out_trade_no: 'sub2_20260420abcd1234',
  status,
  order_type: 'balance',
  created_at: '2026-04-20T12:00:00Z',
  expires_at: '2026-04-20T12:30:00Z',
  refund_amount: 0,
})

const recoverySnapshotFactory = (resumeToken: string) => ({
  orderId: 42,
  amount: 88,
  qrCode: '',
  expiresAt: '2099-01-01T00:10:00.000Z',
  paymentType: 'alipay',
  payUrl: 'https://pay.example.com/session/42',
  outTradeNo: 'sub2_20260420abcd1234',
  clientSecret: '',
  intentId: '',
  currency: '',
  countryCode: '',
  paymentEnv: '',
  payAmount: 88,
  orderType: 'balance',
  paymentMode: 'popup',
  resumeToken,
  createdAt: Date.UTC(2099, 0, 1, 0, 0, 0),
})

describe('PaymentResultView', () => {
  beforeEach(() => {
    routeState.query = {}
    routerPush.mockReset()
    pollOrderStatus.mockReset()
    verifyOrder.mockReset()
    verifyOrderPublic.mockReset()
    resolveOrderPublicByResumeToken.mockReset()
    appStoreState.cachedPublicSettings = {
      payment_shell_config: buildPaymentResultShellConfig(),
    }
    window.localStorage.clear()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('renders a pending state instead of a failure state when the restored order is still pending', async () => {
    routeState.query = {
      resume_token: 'resume-42',
      order_id: '999',
      status: 'success',
    }
    window.localStorage.setItem(PAYMENT_RECOVERY_STORAGE_KEY, JSON.stringify({
      orderId: 42,
      amount: 88,
      qrCode: '',
      expiresAt: '2099-01-01T00:10:00.000Z',
      paymentType: 'alipay',
      payUrl: 'https://pay.example.com/session/42',
      outTradeNo: 'sub2_20260420abcd1234',
      clientSecret: '',
      intentId: '',
      currency: '',
      countryCode: '',
      paymentEnv: '',
      payAmount: 88,
      orderType: 'balance',
      paymentMode: 'redirect',
      resumeToken: 'resume-42',
      createdAt: Date.UTC(2099, 0, 1, 0, 0, 0),
    }))
    resolveOrderPublicByResumeToken.mockResolvedValue({
      data: orderFactory('PENDING'),
    })

    const wrapper = mount(PaymentResultView, {
      global: {
        stubs: {
          OrderStatusBadge: true,
        },
      },
    })

    await flushPromises()

    expect(resolveOrderPublicByResumeToken).toHaveBeenCalledWith('resume-42')
    expect(pollOrderStatus).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('Payment processing')
    expect(wrapper.text()).not.toContain('Payment successful')
    expect(wrapper.text()).not.toContain('Payment failed')
  })

  it('prefers the public resume-token result over a stale restored DB snapshot', async () => {
    routeState.query = {
      resume_token: 'resume-authoritative',
      order_id: '42',
      status: 'success',
    }
    window.localStorage.setItem(PAYMENT_RECOVERY_STORAGE_KEY, JSON.stringify({
      orderId: 42,
      amount: 88,
      qrCode: '',
      expiresAt: '2099-01-01T00:10:00.000Z',
      paymentType: 'alipay',
      payUrl: 'https://pay.example.com/session/42',
      outTradeNo: 'sub2_20260420abcd1234',
      clientSecret: '',
      intentId: '',
      currency: '',
      countryCode: '',
      paymentEnv: '',
      payAmount: 88,
      orderType: 'balance',
      paymentMode: 'popup',
      resumeToken: 'resume-authoritative',
      createdAt: Date.UTC(2099, 0, 1, 0, 0, 0),
    }))
    resolveOrderPublicByResumeToken.mockResolvedValue({
      data: {
        ...orderFactory('PAID'),
        amount: 100,
        pay_amount: 103,
        fee_rate: 3,
      },
    })

    const wrapper = mount(PaymentResultView, {
      global: {
        stubs: {
          OrderStatusBadge: true,
        },
      },
    })

    await flushPromises()

    expect(pollOrderStatus).not.toHaveBeenCalled()
    expect(resolveOrderPublicByResumeToken).toHaveBeenCalledWith('resume-authoritative')
    expect(wrapper.text()).toContain('Payment successful')
    expect(wrapper.text()).toContain('103.00')
    expect(wrapper.text()).toContain('100.00')
    expect(window.localStorage.getItem(PAYMENT_RECOVERY_STORAGE_KEY)).toBeNull()
  })

  it('refreshes a pending resume-token result until the order becomes paid', async () => {
    vi.useFakeTimers()
    appStoreState.cachedPublicSettings = {
      payment_shell_config: buildPaymentResultShellConfig({}, {
        paymentResultRefreshIntervalMs: 1234,
        paymentResultMaxRefreshAttempts: 3,
      }),
    }
    routeState.query = {
      resume_token: 'resume-77',
    }
    window.localStorage.setItem(
      PAYMENT_RECOVERY_STORAGE_KEY,
      JSON.stringify(recoverySnapshotFactory('resume-77')),
    )
    resolveOrderPublicByResumeToken
      .mockResolvedValueOnce({
        data: orderFactory('PENDING'),
      })
      .mockResolvedValueOnce({
        data: orderFactory('PAID'),
      })

    const wrapper = mount(PaymentResultView, {
      global: {
        stubs: {
          OrderStatusBadge: true,
        },
      },
    })

    await flushPromises()

    expect(resolveOrderPublicByResumeToken).toHaveBeenCalledTimes(1)
    expect(wrapper.text()).toContain('Payment processing')
    expect(window.localStorage.getItem(PAYMENT_RECOVERY_STORAGE_KEY)).not.toBeNull()

    await vi.advanceTimersByTimeAsync(1234)
    await flushPromises()

    expect(resolveOrderPublicByResumeToken).toHaveBeenCalledTimes(2)
    expect(wrapper.text()).toContain('Payment successful')
    expect(wrapper.text()).not.toContain('Payment failed')
    expect(window.localStorage.getItem(PAYMENT_RECOVERY_STORAGE_KEY)).toBeNull()
  })

  it('stops refreshing pending orders after the configured max attempts', async () => {
    vi.useFakeTimers()
    appStoreState.cachedPublicSettings = {
      payment_shell_config: buildPaymentResultShellConfig({}, {
        paymentResultRefreshIntervalMs: 50,
        paymentResultMaxRefreshAttempts: 1,
      }),
    }
    routeState.query = {
      resume_token: 'resume-limited',
    }
    window.localStorage.setItem(
      PAYMENT_RECOVERY_STORAGE_KEY,
      JSON.stringify(recoverySnapshotFactory('resume-limited')),
    )
    resolveOrderPublicByResumeToken.mockResolvedValue({
      data: orderFactory('PENDING'),
    })

    mount(PaymentResultView, {
      global: {
        stubs: {
          OrderStatusBadge: true,
        },
      },
    })

    await flushPromises()
    expect(resolveOrderPublicByResumeToken).toHaveBeenCalledTimes(1)

    await vi.advanceTimersByTimeAsync(50)
    await flushPromises()
    expect(resolveOrderPublicByResumeToken).toHaveBeenCalledTimes(2)

    await vi.advanceTimersByTimeAsync(500)
    await flushPromises()
    expect(resolveOrderPublicByResumeToken).toHaveBeenCalledTimes(2)
  })

  it('falls back to order_id polling when resume-token recovery fails', async () => {
    routeState.query = {
      resume_token: 'resume-fail',
      order_id: '77',
    }
    window.localStorage.setItem(
      PAYMENT_RECOVERY_STORAGE_KEY,
      JSON.stringify({
        ...recoverySnapshotFactory('resume-fail'),
        orderId: 42,
      }),
    )
    resolveOrderPublicByResumeToken.mockRejectedValueOnce(new Error('resume failed'))
    pollOrderStatus.mockResolvedValueOnce({
      ...orderFactory('PAID'),
      id: 77,
    })

    const wrapper = mount(PaymentResultView, {
      global: {
        stubs: {
          OrderStatusBadge: true,
        },
      },
    })

    await flushPromises()

    expect(resolveOrderPublicByResumeToken).toHaveBeenCalledWith('resume-fail')
    expect(pollOrderStatus).toHaveBeenCalledWith(77)
    expect(verifyOrderPublic).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('Payment successful')
    expect(window.localStorage.getItem(PAYMENT_RECOVERY_STORAGE_KEY)).toBeNull()
  })

  it('falls back to public out_trade_no verification when resume_token recovery fails in legacy return flows', async () => {
    routeState.query = {
      resume_token: 'resume-fail',
      out_trade_no: 'legacy-should-not-run',
      trade_status: 'TRADE_SUCCESS',
    }
    resolveOrderPublicByResumeToken.mockRejectedValueOnce(new Error('resume failed'))
    verifyOrderPublic.mockResolvedValueOnce({
      data: {
        ...orderFactory('PAID'),
        out_trade_no: 'legacy-should-not-run',
      },
    })

    const wrapper = mount(PaymentResultView, {
      global: {
        stubs: {
          OrderStatusBadge: true,
        },
      },
    })

    await flushPromises()

    expect(resolveOrderPublicByResumeToken).toHaveBeenCalledWith('resume-fail')
    expect(verifyOrderPublic).toHaveBeenCalledWith('legacy-should-not-run')
    expect(pollOrderStatus).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('Payment successful')
  })

  it('ignores a stale global recovery snapshot when legacy return markers do not identify the order', async () => {
    routeState.query = {
      trade_status: 'TRADE_SUCCESS',
    }
    window.localStorage.setItem(
      PAYMENT_RECOVERY_STORAGE_KEY,
      JSON.stringify(recoverySnapshotFactory('resume-stale')),
    )

    const wrapper = mount(PaymentResultView, {
      global: {
        stubs: {
          OrderStatusBadge: true,
        },
      },
    })

    await flushPromises()

    expect(resolveOrderPublicByResumeToken).not.toHaveBeenCalled()
    expect(verifyOrderPublic).not.toHaveBeenCalled()
    expect(pollOrderStatus).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('Payment failed')
    expect(wrapper.text()).not.toContain('sub2_20260420abcd1234')
  })

  it('uses public out_trade_no verification when no signed resume context is available', async () => {
    routeState.query = {
      out_trade_no: 'legacy-123',
      trade_status: 'TRADE_SUCCESS',
    }
    verifyOrder.mockRejectedValue(new Error('auth required'))
    verifyOrderPublic.mockResolvedValue({
      data: orderFactory('PAID'),
    })

    const wrapper = mount(PaymentResultView, {
      global: {
        stubs: {
          OrderStatusBadge: true,
        },
      },
    })

    await flushPromises()

    expect(verifyOrder).toHaveBeenCalledWith('legacy-123')
    expect(verifyOrderPublic).toHaveBeenCalledWith('legacy-123')
    expect(pollOrderStatus).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('Payment successful')
  })

  it('prefers authenticated order verification before falling back to public lookup', async () => {
    routeState.query = {
      out_trade_no: 'auth-verify-123',
      trade_status: 'TRADE_SUCCESS',
    }
    verifyOrder.mockResolvedValue({
      data: orderFactory('COMPLETED'),
    })

    const wrapper = mount(PaymentResultView, {
      global: {
        stubs: {
          OrderStatusBadge: true,
        },
      },
    })

    await flushPromises()

    expect(verifyOrder).toHaveBeenCalledWith('auth-verify-123')
    expect(verifyOrderPublic).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('Payment successful')
  })

  it('does not use public out_trade_no verification for bare order numbers without legacy return markers', async () => {
    routeState.query = {
      out_trade_no: 'legacy-bare',
    }

    mount(PaymentResultView, {
      global: {
        stubs: {
          OrderStatusBadge: true,
        },
      },
    })

    await flushPromises()

    expect(verifyOrderPublic).not.toHaveBeenCalled()
  })

  it('resolves order by resume token when local recovery snapshot is missing', async () => {
    routeState.query = {
      resume_token: 'resume-77',
    }
    resolveOrderPublicByResumeToken.mockResolvedValue({
      data: orderFactory('PAID'),
    })

    const wrapper = mount(PaymentResultView, {
      global: {
        stubs: {
          OrderStatusBadge: true,
        },
      },
    })

    await flushPromises()

    expect(resolveOrderPublicByResumeToken).toHaveBeenCalledWith('resume-77')
    expect(wrapper.text()).toContain('Payment successful')
  })

  it('uses the currency returned by the order API when rendering amounts', async () => {
    routeState.query = {
      resume_token: 'resume-hkd',
    }
    resolveOrderPublicByResumeToken.mockResolvedValue({
      data: {
        ...orderFactory('PAID'),
        currency: 'HKD',
        amount: 100,
        pay_amount: 103,
        fee_rate: 3,
      },
    })

    const wrapper = mount(PaymentResultView, {
      global: {
        stubs: {
          OrderStatusBadge: true,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain(formatPaymentAmount(103, 'HKD'))
  })

  it('normalizes aliased payment methods before rendering the label', async () => {
    routeState.query = {
      resume_token: 'resume-88',
    }
    resolveOrderPublicByResumeToken.mockResolvedValueOnce({
      data: {
        ...orderFactory('PAID'),
        payment_type: 'alipay_direct',
      },
    })

    const wrapper = mount(PaymentResultView, {
      global: {
        stubs: {
          OrderStatusBadge: true,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('Alipay')
    expect(wrapper.text()).not.toContain('alipay_direct')
  })

  it('renders payment result labels from public payment shell config when available', async () => {
    appStoreState.cachedPublicSettings = {
      payment_shell_config: buildPaymentResultShellConfig({
        success: 'Configured success',
        orderId: 'Configured order',
        baseAmount: 'Configured base',
        payAmount: 'Configured paid',
        paymentMethod: 'Configured method',
        status: 'Configured status',
        backToRecharge: 'Configured recharge',
        viewOrders: 'Configured orders',
        methodAlipay: 'Configured Alipay',
      }),
    }
    routeState.query = {
      resume_token: 'resume-configured',
    }
    resolveOrderPublicByResumeToken.mockResolvedValueOnce({
      data: orderFactory('PAID'),
    })

    const wrapper = mount(PaymentResultView, {
      global: {
        stubs: {
          OrderStatusBadge: true,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('Configured success')
    expect(wrapper.text()).toContain('Configured order')
    expect(wrapper.text()).toContain('Configured base')
    expect(wrapper.text()).toContain('Configured paid')
    expect(wrapper.text()).toContain('Configured method')
    expect(wrapper.text()).toContain('Configured status')
    expect(wrapper.text()).toContain('Configured recharge')
    expect(wrapper.text()).toContain('Configured orders')
    expect(wrapper.text()).toContain('Configured Alipay')
  })

  it('does not carry local payment-result i18n fallback maps in the view', () => {
    expect(paymentResultViewSource).not.toContain('const paymentResultLabelKeys')
    expect(paymentResultViewSource).not.toContain('resolvePaymentShellLabels(')
    expect(paymentResultViewSource).toContain('resolvePaymentResultLabels')
    expect(paymentResultViewSource).toContain('resolvePaymentResultDefaults')
    expect(paymentResultViewSource).toContain('renderPaymentResultText')
    expect(paymentResultViewSource).toContain("from './paymentResultRuntime'")
    expect(paymentResultViewSource).toContain('calculatePaymentBaseAmount')
    expect(paymentResultViewSource).toContain('restorePaymentRecoverySnapshot')
    expect(paymentResultViewSource).toContain('paymentResultDefaults.value.refreshIntervalMs')
    expect(paymentResultViewSource).toContain('paymentResultDefaults.value.maxRefreshAttempts')
    expect(paymentResultViewSource).not.toContain('STATUS_REFRESH_INTERVAL_MS')
    expect(paymentResultViewSource).not.toContain('STATUS_REFRESH_MAX_ATTEMPTS')
    expect(paymentResultViewSource).not.toContain('paymentResultFallbackKeys')
    expect(paymentResultViewSource).not.toContain('paymentResultLabels.value[key] || key')
    expect(paymentResultViewSource).not.toContain('payment.result.success')
    expect(paymentResultViewSource).not.toContain('payment.methods.alipay')
    expect(paymentResultViewSource).not.toContain('paymentMethodI18nKey')
    expect(paymentResultViewSource).not.toContain("'$' + order.amount.toFixed(2)")
    expect(paymentResultViewSource).not.toContain("currency = ref('CNY')")
    expect(paymentResultViewSource).not.toContain('currency = ref(normalizePaymentCurrency())')
  })

  it('uses auth route defaults for payment-result navigation actions', () => {
    expect(paymentResultViewSource).toContain('useAuthRouteDefaults')
    expect(paymentResultViewSource).toContain('router.push(authRouteDefaults.purchasePath)')
    expect(paymentResultViewSource).toContain('router.push(authRouteDefaults.ordersPath)')
    expect(paymentResultViewSource).not.toContain("router.push('/purchase')")
    expect(paymentResultViewSource).not.toContain("router.push('/orders')")
  })
})
