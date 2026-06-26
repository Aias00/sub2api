import { readFileSync } from 'node:fs'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, shallowMount } from '@vue/test-utils'
import PaymentQRCodeView from '../PaymentQRCodeView.vue'

const routeState = vi.hoisted(() => ({
  query: {} as Record<string, unknown>,
}))
const routerPush = vi.hoisted(() => vi.fn())
const appStoreState = vi.hoisted(() => ({
  cachedPublicSettings: null as null | { auth_shell_config?: string, payment_shell_config?: string },
  showError: vi.fn(),
}))
const pollOrderStatus = vi.hoisted(() => vi.fn())
const cancelOrder = vi.hoisted(() => vi.fn())
const toCanvas = vi.hoisted(() => vi.fn())

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

vi.mock('@/stores/payment', () => ({
  usePaymentStore: () => ({
    pollOrderStatus,
  }),
}))

vi.mock('@/api/payment', () => ({
  paymentAPI: {
    cancelOrder,
  },
}))

vi.mock('qrcode', () => ({
  default: {
    toCanvas,
  },
}))

const paymentQRCodeViewSource = readFileSync('src/views/user/PaymentQRCodeView.vue', 'utf8')

function mountView() {
  return shallowMount(PaymentQRCodeView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
      },
    },
  })
}

describe('PaymentQRCodeView', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.spyOn(HTMLCanvasElement.prototype, 'getContext').mockReturnValue({
      fillStyle: '',
      beginPath: vi.fn(),
      moveTo: vi.fn(),
      arcTo: vi.fn(),
      fill: vi.fn(),
      drawImage: vi.fn(),
    } as unknown as CanvasRenderingContext2D)
    routeState.query = {
      order_id: '7',
      qr: 'weixin://pay/test',
      payment_type: 'wxpay',
      expires_at: '2099-01-01T00:10:00.000Z',
    }
    routerPush.mockReset()
    appStoreState.showError.mockReset()
    appStoreState.cachedPublicSettings = {
      payment_shell_config: JSON.stringify({
        zh: {
          labels: {
            scanWxpay: '配置微信扫码',
            scanWxpayHint: '配置微信提示',
            expiresIn: '配置剩余时间',
            waitingPayment: '配置等待支付',
            cancelOrder: '配置取消订单',
            expired: '配置已过期',
            errorFallback: '配置支付错误',
          },
          defaults: {
            paymentStatusPollIntervalMs: 1234,
          },
        },
      }),
    }
    pollOrderStatus.mockReset()
    cancelOrder.mockReset()
    toCanvas.mockReset().mockResolvedValue(undefined)
  })

  afterEach(() => {
    vi.restoreAllMocks()
    vi.useRealTimers()
  })

  it('优先使用 public settings 中的二维码页文案', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('配置微信扫码')
    expect(wrapper.text()).toContain('配置微信提示')
    expect(wrapper.text()).toContain('配置剩余时间')
    expect(wrapper.text()).toContain('配置等待支付')
    expect(wrapper.text()).toContain('配置取消订单')
  })

  it('取消失败时使用 public settings 中的错误兜底文案', async () => {
    cancelOrder.mockRejectedValue({})
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('button.btn.btn-secondary').trigger('click')
    await flushPromises()

    expect(cancelOrder).toHaveBeenCalledWith(7)
    expect(appStoreState.showError).toHaveBeenCalledWith('配置支付错误')
  })

  it('expires immediately instead of synthesizing a local timeout when expires_at is missing', async () => {
    routeState.query = {
      order_id: '7',
      qr: 'weixin://pay/test',
      payment_type: 'wxpay',
    }

    const wrapper = mountView()
    await flushPromises()
    await vi.advanceTimersByTimeAsync(3000)

    expect(wrapper.text()).toContain('配置已过期')
    expect(pollOrderStatus).not.toHaveBeenCalled()
  })

  it('uses public payment shell polling interval for status checks', async () => {
    mountView()
    await flushPromises()

    await vi.advanceTimersByTimeAsync(1233)
    expect(pollOrderStatus).not.toHaveBeenCalled()

    await vi.advanceTimersByTimeAsync(1)
    expect(pollOrderStatus).toHaveBeenCalledWith(7)
  })

  it('does not carry local QR-page i18n fallback maps in the view', () => {
    expect(paymentQRCodeViewSource).not.toContain('const paymentQRLabelKeys')
    expect(paymentQRCodeViewSource).not.toContain('resolvePaymentShellLabels(')
    expect(paymentQRCodeViewSource).toContain('resolvePaymentQRLabels')
    expect(paymentQRCodeViewSource).toContain('resolvePaymentStatusPollingDefaults')
    expect(paymentQRCodeViewSource).toContain('renderPaymentQRText')
    expect(paymentQRCodeViewSource).toContain("from './paymentQrRuntime'")
    expect(paymentQRCodeViewSource).toContain('resolvePaymentQrRouteState')
    expect(paymentQRCodeViewSource).toContain('resolvePaymentQrSecondsUntil')
    expect(paymentQRCodeViewSource).toContain('useAuthRouteDefaults')
    expect(paymentQRCodeViewSource).toContain('router.push(authRouteDefaults.purchasePath)')
    expect(paymentQRCodeViewSource).toContain('router.push(authRouteDefaults.value.purchasePath)')
    expect(paymentQRCodeViewSource).toContain('authRouteDefaults.value.paymentResultPath')
    expect(paymentQRCodeViewSource).not.toContain("router.push('/purchase')")
    expect(paymentQRCodeViewSource).not.toContain("path: '/payment/result'")
    expect(paymentQRCodeViewSource).not.toContain('paymentQRFallbackKeys')
    expect(paymentQRCodeViewSource).not.toContain('paymentQRLabels.value[key] || key')
    expect(paymentQRCodeViewSource).not.toContain('payment.qr.scanWxpay')
    expect(paymentQRCodeViewSource).not.toContain('payment.result.backToRecharge')
    expect(paymentQRCodeViewSource).not.toContain('30 * 60')
    expect(paymentQRCodeViewSource).not.toContain('setInterval(pollStatus, 3000)')
  })
})
