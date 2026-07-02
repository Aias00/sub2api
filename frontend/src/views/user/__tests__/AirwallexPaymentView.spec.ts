import { readFileSync } from 'node:fs'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, shallowMount } from '@vue/test-utils'
import AirwallexPaymentView from '../AirwallexPaymentView.vue'
import {
  PAYMENT_RECOVERY_STORAGE_KEY,
  type PaymentRecoverySnapshot,
} from '@/components/payment/paymentFlow'

const routeState = vi.hoisted(() => ({
  query: {} as Record<string, unknown>,
}))
const routerPush = vi.hoisted(() => vi.fn())
const airwallexInit = vi.hoisted(() => vi.fn())
const redirectToCheckout = vi.hoisted(() => vi.fn())
const appStoreState = vi.hoisted(() => ({
  cachedPublicSettings: null as null | {
    auth_shell_config?: string
    payment_shell_config?: string
  },
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

vi.mock('@/stores', () => ({
  useAppStore: () => appStoreState,
}))

vi.mock('@airwallex/components-sdk', () => ({
  init: airwallexInit,
}))

const airwallexPaymentViewSource = readFileSync(
  'src/views/user/AirwallexPaymentView.vue',
  'utf8',
)

function buildAirwallexPaymentShellConfig(overrides: Record<string, string> = {}) {
  return JSON.stringify({
    zh: {
      labels: {
        airwallexLoadFailed: 'Airwallex 加载失败',
        airwallexMissingParams: 'Airwallex 参数缺失',
        backToRecharge: '返回充值',
        payInNewWindowHint: '正在打开支付页面',
        ...overrides,
      },
    },
  })
}

function buildAuthShellConfig(paymentResultPath: string) {
  return JSON.stringify({
    zh: {
      defaults: {
        paymentResultPath,
      },
    },
  })
}

function airwallexSnapshot(overrides: Partial<PaymentRecoverySnapshot> = {}): PaymentRecoverySnapshot {
  return {
    orderId: 101,
    amount: 88,
    qrCode: '',
    expiresAt: '2099-01-01T00:10:00.000Z',
    paymentType: 'airwallex',
    payUrl: '/payment/airwallex?order_id=101&out_trade_no=sub2_awx_101&resume_token=resume-awx',
    outTradeNo: 'sub2_awx_101',
    clientSecret: 'awx_client_secret',
    intentId: 'int_awx_101',
    currency: 'CNY',
    countryCode: 'CN',
    paymentEnv: 'demo',
    payAmount: 88,
    orderType: 'balance',
    paymentMode: '',
    resumeToken: 'resume-awx',
    createdAt: Date.UTC(2099, 0, 1, 0, 0, 0),
    ...overrides,
  }
}

function mountView() {
  return shallowMount(AirwallexPaymentView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        Icon: true,
      },
    },
  })
}

describe('AirwallexPaymentView', () => {
  beforeEach(() => {
    routeState.query = {}
    routerPush.mockReset()
    appStoreState.cachedPublicSettings = {
      payment_shell_config: buildAirwallexPaymentShellConfig(),
    }
    airwallexInit.mockReset().mockResolvedValue({
      payments: {
        redirectToCheckout,
      },
    })
    redirectToCheckout.mockReset()
    window.localStorage.clear()
  })

  it('从本地恢复快照读取支付参数，避免在 URL 中暴露 client_secret', async () => {
    routeState.query = {
      order_id: '101',
      out_trade_no: 'sub2_awx_101',
      resume_token: 'resume-awx',
    }
    window.localStorage.setItem(
      PAYMENT_RECOVERY_STORAGE_KEY,
      JSON.stringify(airwallexSnapshot()),
    )

    mountView()
    await flushPromises()
    await flushPromises()

    expect(airwallexInit).toHaveBeenCalledWith({
      env: 'demo',
      enabledElements: ['payments'],
      locale: 'zh',
    })
    expect(redirectToCheckout).toHaveBeenCalledWith(expect.objectContaining({
      intent_id: 'int_awx_101',
      client_secret: 'awx_client_secret',
      currency: 'CNY',
      country_code: 'CN',
    }))

    const checkoutOptions = redirectToCheckout.mock.calls[0][0]
    const successUrl = new URL(checkoutOptions.successUrl)
    expect(successUrl.pathname).toBe('/payment/result')
    expect(successUrl.searchParams.get('order_id')).toBe('101')
    expect(successUrl.searchParams.get('out_trade_no')).toBe('sub2_awx_101')
    expect(successUrl.searchParams.get('resume_token')).toBe('resume-awx')
  })

  it('uses the Cloudbase auth shell payment result path for Airwallex success URLs', async () => {
    appStoreState.cachedPublicSettings = {
      auth_shell_config: buildAuthShellConfig('/configured-payment-result'),
      payment_shell_config: buildAirwallexPaymentShellConfig(),
    }
    routeState.query = {
      order_id: '101',
      out_trade_no: 'sub2_awx_101',
      resume_token: 'resume-awx',
    }
    window.localStorage.setItem(
      PAYMENT_RECOVERY_STORAGE_KEY,
      JSON.stringify(airwallexSnapshot()),
    )

    mountView()
    await flushPromises()
    await flushPromises()

    const checkoutOptions = redirectToCheckout.mock.calls[0][0]
    const successUrl = new URL(checkoutOptions.successUrl)
    expect(successUrl.pathname).toBe('/configured-payment-result')
    expect(successUrl.searchParams.get('order_id')).toBe('101')
    expect(successUrl.searchParams.get('out_trade_no')).toBe('sub2_awx_101')
    expect(successUrl.searchParams.get('resume_token')).toBe('resume-awx')
  })

  it('缺失币种和国家码时不在前端合成 CNY/CN', async () => {
    routeState.query = {
      order_id: '101',
      out_trade_no: 'sub2_awx_101',
      resume_token: 'resume-awx',
    }
    window.localStorage.setItem(
      PAYMENT_RECOVERY_STORAGE_KEY,
      JSON.stringify(airwallexSnapshot({
        currency: '',
        countryCode: '',
      })),
    )

    mountView()
    await flushPromises()
    await flushPromises()

    expect(redirectToCheckout).toHaveBeenCalledWith(expect.objectContaining({
      currency: '',
      country_code: '',
    }))
  })

  it('拒绝只从 URL query 读取 Airwallex 支付密钥', async () => {
    routeState.query = {
      order_id: '101',
      intent_id: 'int_from_query',
      client_secret: 'secret_from_query',
    }

    const wrapper = mountView()
    await flushPromises()

    expect(airwallexInit).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('Airwallex 参数缺失')
  })

  it('优先使用 public settings 中的 Airwallex 承载页文案', async () => {
    appStoreState.cachedPublicSettings = {
      payment_shell_config: buildAirwallexPaymentShellConfig({
        airwallexLoadFailed: '配置 Airwallex 失败',
        airwallexMissingParams: '配置缺参',
        backToRecharge: '配置返回',
      }),
    }

    const wrapper = mountView()
    await flushPromises()

    expect(airwallexInit).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('配置 Airwallex 失败')
    expect(wrapper.text()).toContain('配置缺参')
    expect(wrapper.text()).toContain('配置返回')
  })

  it('does not carry local Airwallex payment i18n fallback maps in the view', () => {
    expect(airwallexPaymentViewSource).not.toContain('const airwallexPaymentLabelKeys')
    expect(airwallexPaymentViewSource).not.toContain('resolvePaymentShellLabels(')
    expect(airwallexPaymentViewSource).toContain('resolveAirwallexPaymentLabels')
    expect(airwallexPaymentViewSource).toContain('renderAirwallexPaymentText')
    expect(airwallexPaymentViewSource).toContain("from './airwallexPaymentRuntime'")
    expect(airwallexPaymentViewSource).toContain('buildAirwallexSuccessUrl')
    expect(airwallexPaymentViewSource).toContain('restoreAirwallexPaymentSnapshot')
    expect(airwallexPaymentViewSource).toContain('useAuthRouteDefaults')
    expect(airwallexPaymentViewSource).toContain('authRouteDefaults.value.paymentResultPath')
    expect(airwallexPaymentViewSource).toContain('router.push(authRouteDefaults.purchasePath)')
    expect(airwallexPaymentViewSource).not.toContain("router.push('/purchase')")
    expect(airwallexPaymentViewSource).not.toContain("new URL('/payment/result'")
    expect(airwallexPaymentViewSource).not.toContain('airwallexPaymentFallbackKeys')
    expect(airwallexPaymentViewSource).not.toContain('airwallexPaymentLabels.value[key] || key')
    expect(airwallexPaymentViewSource).not.toContain('payment.airwallexMissingParams')
    expect(airwallexPaymentViewSource).not.toContain('payment.result.backToRecharge')
    expect(airwallexPaymentViewSource).not.toContain("snapshot.currency || 'CNY'")
    expect(airwallexPaymentViewSource).not.toContain("snapshot.countryCode || 'CN'")
  })
})
