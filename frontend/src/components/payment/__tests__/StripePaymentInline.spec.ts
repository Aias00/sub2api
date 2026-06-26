import { readFileSync } from 'node:fs'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, shallowMount } from '@vue/test-utils'
import StripePaymentInline from '../StripePaymentInline.vue'

const loadStripe = vi.hoisted(() => vi.fn())
const cancelOrder = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({
      resolve: vi.fn((route: { path: string }) => ({ href: route.path })),
    }),
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
  useAppStore: () => ({
    showError,
  }),
}))

vi.mock('@/api/payment', () => ({
  paymentAPI: {
    cancelOrder,
  },
}))

vi.mock('@stripe/stripe-js', () => ({
  loadStripe,
}))

const stripePaymentInlineSource = readFileSync('src/components/payment/StripePaymentInline.vue', 'utf8')

function mountInline(labels: Record<string, string> = {}) {
  return shallowMount(StripePaymentInline, {
    props: {
      orderId: 88,
      amount: 20,
      clientSecret: 'pi_secret_88',
      publishableKey: 'pk_test',
      payAmount: 21,
      labels,
    },
    global: {
      stubs: {
        Icon: true,
      },
    },
  })
}

describe('StripePaymentInline', () => {
  beforeEach(() => {
    loadStripe.mockReset().mockResolvedValue(null)
    cancelOrder.mockReset()
    showError.mockReset()
  })

  it('使用外部传入的 payment shell labels 渲染错误操作文案', async () => {
    const wrapper = mountInline({
      stripeLoadFailed: '配置 Stripe 加载失败',
      backToRecharge: '配置返回充值',
    })

    await flushPromises()

    expect(wrapper.text()).toContain('配置 Stripe 加载失败')
    expect(wrapper.text()).toContain('配置返回充值')
  })

  it('取消失败时使用外部传入的错误兜底文案', async () => {
    loadStripe.mockResolvedValue({
      elements: vi.fn(() => ({
        create: vi.fn(() => ({
          mount: vi.fn(),
          on: vi.fn(),
        })),
      })),
    })
    cancelOrder.mockRejectedValue({})
    const wrapper = mountInline({
      errorFallback: '配置支付错误',
      cancelOrder: '配置取消订单',
    })

    await flushPromises()
    await wrapper.get('button.btn.btn-secondary').trigger('click')
    await flushPromises()

    expect(cancelOrder).toHaveBeenCalledWith(88)
    expect(showError).toHaveBeenCalledWith('配置支付错误')
  })

  it('does not carry local Stripe inline i18n fallback maps in the component', () => {
    expect(stripePaymentInlineSource).not.toContain('const stripeInlineLabelKeys')
    expect(stripePaymentInlineSource).toContain('renderStripeInlineText')
    expect(stripePaymentInlineSource).not.toContain('stripeInlineFallbackKeys')
    expect(stripePaymentInlineSource).not.toContain('payment.stripePay')
    expect(stripePaymentInlineSource).not.toContain('payment.result.backToRecharge')
    expect(stripePaymentInlineSource).toContain('authRouteDefaults.value.paymentResultPath')
    expect(stripePaymentInlineSource).not.toContain("window.location.origin + '/payment/result")
    expect(stripePaymentInlineSource).not.toContain('return props.labels?.[key] || key')
  })
})
