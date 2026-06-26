import { readFileSync } from 'node:fs'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, shallowMount } from '@vue/test-utils'

const routeState = vi.hoisted(() => ({
  query: {} as Record<string, unknown>,
}))
const routerPush = vi.hoisted(() => vi.fn())
const getOrder = vi.hoisted(() => vi.fn())
const paymentStore = vi.hoisted(() => ({
  config: { stripe_publishable_key: 'pk_test' } as { stripe_publishable_key?: string },
  fetchConfig: vi.fn(),
  pollOrderStatus: vi.fn(),
}))
const appStoreState = vi.hoisted(() => ({
  cachedPublicSettings: null as null | { auth_shell_config?: string, payment_shell_config?: string },
}))
const loadStripe = vi.hoisted(() => vi.fn())
const stripeElements = vi.hoisted(() => ({
  create: vi.fn(),
}))
const stripePaymentElement = vi.hoisted(() => ({
  mount: vi.fn(),
  on: vi.fn(),
}))
const stripeInstance = vi.hoisted(() => ({
  elements: vi.fn(),
  confirmPayment: vi.fn(),
  confirmAlipayPayment: vi.fn(),
  confirmWechatPayPayment: vi.fn(),
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
      locale: { value: 'zh-CN' },
    }),
  }
})

vi.mock('@/stores/payment', () => ({
  usePaymentStore: () => paymentStore,
}))

vi.mock('@/stores', () => ({
  useAppStore: () => appStoreState,
}))

vi.mock('@/api/payment', () => ({
  paymentAPI: {
    getOrder,
  },
}))

vi.mock('@stripe/stripe-js', () => ({
  loadStripe,
}))

import StripePaymentView from '../StripePaymentView.vue'
import { formatPaymentAmount } from '@/components/payment/currency'
import type { PaymentOrder } from '@/types/payment'

const stripePaymentViewSource = readFileSync('src/views/user/StripePaymentView.vue', 'utf8')

function orderFactory(overrides: Partial<PaymentOrder> = {}): PaymentOrder {
  return {
    id: 42,
    user_id: 7,
    amount: 100,
    pay_amount: 103,
    currency: 'CNY',
    fee_rate: 0.03,
    payment_type: 'stripe',
    out_trade_no: 'sub2_stripe_42',
    status: 'PENDING',
    order_type: 'balance',
    created_at: '2026-04-20T12:00:00Z',
    expires_at: '2026-04-20T12:30:00Z',
    refund_amount: 0,
    ...overrides,
  }
}

function mountView() {
  return shallowMount(StripePaymentView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        Icon: true,
      },
    },
  })
}

describe('StripePaymentView', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    routeState.query = {
      order_id: '42',
      client_secret: 'pi_secret_42',
      method: '',
    }
    routerPush.mockReset()
    getOrder.mockReset()
    paymentStore.config = { stripe_publishable_key: 'pk_test' }
    appStoreState.cachedPublicSettings = null
    paymentStore.fetchConfig.mockReset().mockResolvedValue(undefined)
    paymentStore.pollOrderStatus.mockReset()
    loadStripe.mockReset().mockResolvedValue(stripeInstance)
    stripeElements.create.mockReset().mockReturnValue(stripePaymentElement)
    stripePaymentElement.mount.mockReset()
    stripePaymentElement.on.mockReset().mockImplementation((event: string, callback: () => void) => {
      if (event === 'ready') callback()
    })
    stripeInstance.elements.mockReset().mockReturnValue(stripeElements)
    stripeInstance.confirmPayment.mockReset()
    stripeInstance.confirmAlipayPayment.mockReset()
    stripeInstance.confirmWechatPayPayment.mockReset()
    Object.defineProperty(window, 'matchMedia', {
      configurable: true,
      value: vi.fn().mockReturnValue({ matches: false }),
    })
    window.localStorage.clear()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('本地恢复快照缺失时使用订单接口返回的 Stripe 币种展示金额', async () => {
    getOrder.mockResolvedValue({
      data: orderFactory({ currency: 'HKD', pay_amount: 103 }),
    })

    const wrapper = mountView()
    await flushPromises()
    await flushPromises()

    expect(getOrder).toHaveBeenCalledWith(42)
    expect(loadStripe).toHaveBeenCalledWith('pk_test')
    expect(wrapper.text()).toContain(formatPaymentAmount(103, 'HKD', 'zh-CN'))
  })

  it('优先使用 public settings 中的 Stripe 承载页文案', async () => {
    appStoreState.cachedPublicSettings = {
      payment_shell_config: JSON.stringify({
        zh: {
          labels: {
            actualPay: '配置实付',
            stripePay: '配置支付',
            backToRecharge: '配置返回',
          },
        },
      }),
    }
    getOrder.mockResolvedValue({
      data: orderFactory(),
    })

    const wrapper = mountView()
    await flushPromises()
    await flushPromises()

    expect(wrapper.text()).toContain('配置实付')
    expect(wrapper.text()).toContain('配置支付')
    expect(wrapper.text()).toContain('配置返回')
  })

  it('uses Stripe runtime defaults from public payment shell settings', async () => {
    routeState.query = {
      order_id: '42',
      client_secret: 'pi_secret_42',
      method: 'wechat_pay',
    }
    appStoreState.cachedPublicSettings = {
      payment_shell_config: JSON.stringify({
        zh: {
          labels: {
            scanWxpay: '配置微信扫码',
            scanWxpayHint: '配置微信提示',
            success: '配置成功',
          },
          defaults: {
            stripePollIntervalMs: 1234,
            stripeCloseDelayMs: 2345,
          },
        },
      }),
    }
    getOrder.mockResolvedValue({
      data: orderFactory(),
    })
    stripeInstance.confirmWechatPayPayment.mockResolvedValue({
      paymentIntent: {
        status: 'requires_action',
        next_action: {
          wechat_pay_display_qr_code: {
            image_data_url: 'data:image/png;base64,qr',
          },
        },
      },
    })
    paymentStore.pollOrderStatus.mockResolvedValue(orderFactory({ status: 'PAID' }))

    const wrapper = mountView()
    await flushPromises()
    await flushPromises()

    expect(wrapper.text()).toContain('配置微信扫码')
    await vi.advanceTimersByTimeAsync(1233)
    expect(paymentStore.pollOrderStatus).not.toHaveBeenCalled()

    await vi.advanceTimersByTimeAsync(1)
    await flushPromises()
    expect(paymentStore.pollOrderStatus).toHaveBeenCalledWith(42)
    expect(wrapper.text()).toContain('配置成功')

    expect(routerPush).not.toHaveBeenCalled()
    await vi.advanceTimersByTimeAsync(2345)
    expect(routerPush).toHaveBeenCalledWith({
      path: '/payment/result',
      query: { order_id: '42', status: 'success' },
    })
  })

  it('does not carry local Stripe carrier-page i18n fallback maps in the view', () => {
    expect(stripePaymentViewSource).not.toContain('const stripePaymentLabelKeys')
    expect(stripePaymentViewSource).not.toContain('resolvePaymentShellLabels(')
    expect(stripePaymentViewSource).toContain('resolveStripePaymentLabels')
    expect(stripePaymentViewSource).toContain('resolveStripePaymentRuntimeDefaults')
    expect(stripePaymentViewSource).toContain('renderStripePaymentText')
    expect(stripePaymentViewSource).toContain("from './stripePaymentRuntime'")
    expect(stripePaymentViewSource).toContain('resolveStripePaymentRouteState')
    expect(stripePaymentViewSource).toContain('buildStripePaymentResultReturnURL')
    expect(stripePaymentViewSource).toContain('useAuthRouteDefaults')
    expect(stripePaymentViewSource).toContain('router.push(authRouteDefaults.purchasePath)')
    expect(stripePaymentViewSource).toContain('authRouteDefaults.value.paymentResultPath')
    expect(stripePaymentViewSource).not.toContain("router.push('/purchase')")
    expect(stripePaymentViewSource).not.toContain("path: '/payment/result'")
    expect(stripePaymentViewSource).not.toContain("window.location.origin + '/payment/result")
    expect(stripePaymentViewSource).not.toContain('stripePaymentFallbackKeys')
    expect(stripePaymentViewSource).not.toContain('stripePaymentLabels.value[key] || key')
    expect(stripePaymentViewSource).not.toContain('payment.stripePay')
    expect(stripePaymentViewSource).not.toContain('payment.result.backToRecharge')
    expect(stripePaymentViewSource).not.toContain("currency = ref('CNY')")
    expect(stripePaymentViewSource).not.toContain('currency = ref(normalizePaymentCurrency())')
    expect(stripePaymentViewSource).not.toContain('}, 3000)')
    expect(stripePaymentViewSource).not.toContain('}, 2000)')
  })
})
