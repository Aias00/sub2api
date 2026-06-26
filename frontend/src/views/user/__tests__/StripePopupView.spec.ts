import { readFileSync } from 'node:fs'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { shallowMount } from '@vue/test-utils'
import StripePopupView from '../StripePopupView.vue'

const routeState = vi.hoisted(() => ({
  query: {} as Record<string, unknown>,
}))
const appStoreState = vi.hoisted(() => ({
  cachedPublicSettings: null as null | { auth_shell_config?: string, payment_shell_config?: string },
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => routeState,
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

const stripePopupViewSource = readFileSync(
  'src/views/user/StripePopupView.vue',
  'utf8',
)

function mountPopup() {
  return shallowMount(StripePopupView)
}

describe('StripePopupView', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    routeState.query = {
      order_id: '99',
      method: 'wechat_pay',
      amount: '10',
      currency: 'CNY',
    }
    appStoreState.cachedPublicSettings = null
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('优先使用 public settings 中的 Stripe popup 文案', async () => {
    appStoreState.cachedPublicSettings = {
      payment_shell_config: JSON.stringify({
        zh: {
          labels: {
            orderId: '配置订单',
            close: '配置关闭',
            stripePopupRedirecting: '配置跳转中',
            stripePopupTimeout: '配置等待超时',
          },
        },
      }),
    }

    const wrapper = mountPopup()

    expect(wrapper.text()).toContain('配置跳转中')
    expect(wrapper.text()).toContain('配置订单: 99')

    vi.advanceTimersByTime(15000)
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('配置等待超时')
    expect(wrapper.text()).toContain('配置关闭')
  })

  it('uses Stripe popup init timeout from public payment shell settings', async () => {
    appStoreState.cachedPublicSettings = {
      payment_shell_config: JSON.stringify({
        zh: {
          labels: {
            stripePopupRedirecting: '配置跳转中',
            stripePopupTimeout: '配置等待超时',
          },
          defaults: {
            stripePopupInitTimeoutMs: 1234,
          },
        },
      }),
    }

    const wrapper = mountPopup()

    expect(wrapper.text()).toContain('配置跳转中')
    vi.advanceTimersByTime(1233)
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).not.toContain('配置等待超时')

    vi.advanceTimersByTime(1)
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toContain('配置等待超时')
  })

  it('does not carry local Stripe popup i18n fallback maps in the view', () => {
    expect(stripePopupViewSource).not.toContain('const stripePopupLabelKeys')
    expect(stripePopupViewSource).not.toContain('resolvePaymentShellLabels(')
    expect(stripePopupViewSource).toContain('resolveStripePopupLabels')
    expect(stripePopupViewSource).toContain('resolveStripePaymentRuntimeDefaults')
    expect(stripePopupViewSource).toContain('renderStripePopupText')
    expect(stripePopupViewSource).toContain("from './stripePopupRuntime'")
    expect(stripePopupViewSource).toContain('resolveStripePopupRouteState')
    expect(stripePopupViewSource).toContain('buildStripePopupPaymentResultReturnUrl')
    expect(stripePopupViewSource).not.toContain('stripePopupFallbackKeys')
    expect(stripePopupViewSource).not.toContain('stripePopupLabels.value[key] || key')
    expect(stripePopupViewSource).not.toContain('payment.stripePopup.redirecting')
    expect(stripePopupViewSource).not.toContain('common.close')
    expect(stripePopupViewSource).not.toContain("route.query.currency || 'CNY'")
    expect(stripePopupViewSource).not.toContain('METHOD_COLORS')
    expect(stripePopupViewSource).not.toContain('DEFAULT_METHOD_COLOR')
    expect(stripePopupViewSource).not.toContain("route.query.method || 'alipay'")
    expect(stripePopupViewSource).toContain('authRouteDefaults.value.paymentResultPath')
    expect(stripePopupViewSource).not.toContain("window.location.origin + '/payment/result")
    expect(stripePopupViewSource).not.toContain('}, 3000)')
    expect(stripePopupViewSource).not.toContain('}, 2000)')
    expect(stripePopupViewSource).not.toContain('}, 15000)')
  })
})
