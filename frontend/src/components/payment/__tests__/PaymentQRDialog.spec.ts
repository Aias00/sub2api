import { readFileSync } from 'node:fs'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, shallowMount } from '@vue/test-utils'
import PaymentQRDialog from '../PaymentQRDialog.vue'

const cancelOrder = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())
const pollOrderStatus = vi.hoisted(() => vi.fn())

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
    verifyOrder: vi.fn(),
  },
}))

vi.mock('@/stores/payment', () => ({
  usePaymentStore: () => ({
    pollOrderStatus,
  }),
}))

const paymentQRDialogSource = readFileSync('src/components/payment/PaymentQRDialog.vue', 'utf8')

function mountDialog(labels: Record<string, string> = {}, props: Record<string, unknown> = {}) {
  return shallowMount(PaymentQRDialog, {
    props: {
      show: true,
      orderId: 77,
      qrCode: '',
      expiresAt: '2099-01-01T00:10:00.000Z',
      paymentType: 'alipay',
      payUrl: 'https://pay.example.com',
      labels,
      ...props,
    },
    global: {
      stubs: {
        BaseDialog: {
          props: ['title'],
          template: '<section><h1>{{ title }}</h1><slot /><footer><slot name="footer" /></footer></section>',
        },
        Icon: true,
      },
    },
  })
}

describe('PaymentQRDialog', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    cancelOrder.mockReset()
    showError.mockReset()
    pollOrderStatus.mockReset()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('使用外部传入的 payment shell labels 渲染弹窗文案', () => {
    const wrapper = mountDialog({
      payInNewWindow: '配置新窗口标题',
      payInNewWindowHint: '配置新窗口提示',
      openPayWindow: '配置重新打开',
      waitingPayment: '配置等待',
      cancelOrder: '配置取消订单',
    })

    expect(wrapper.text()).toContain('配置新窗口标题')
    expect(wrapper.text()).toContain('配置新窗口提示')
    expect(wrapper.text()).toContain('配置重新打开')
    expect(wrapper.text()).toContain('配置等待')
    expect(wrapper.text()).toContain('配置取消订单')
  })

  it('取消失败时使用外部传入的错误兜底文案', async () => {
    cancelOrder.mockRejectedValue({})
    const wrapper = mountDialog({
      errorFallback: '配置支付错误',
      cancelOrder: '配置取消订单',
    })

    const cancelButton = wrapper.findAll('button').find(button => button.text() === '配置取消订单')
    expect(cancelButton).toBeTruthy()
    await cancelButton!.trigger('click')
    await flushPromises()

    expect(cancelOrder).toHaveBeenCalledWith(77)
    expect(showError).toHaveBeenCalledWith('配置支付错误')
  })

  it('expires immediately instead of synthesizing a local timeout when expiresAt is missing', async () => {
    const wrapper = mountDialog({
      expired: '配置已过期',
    }, {
      expiresAt: '',
    })

    await vi.advanceTimersByTimeAsync(3000)

    expect(wrapper.text()).toContain('配置已过期')
    expect(pollOrderStatus).not.toHaveBeenCalled()
  })

  it('uses the provided polling interval for status checks', async () => {
    mountDialog({}, {
      pollIntervalMs: 1234,
    })

    await vi.advanceTimersByTimeAsync(1233)
    expect(pollOrderStatus).not.toHaveBeenCalled()

    await vi.advanceTimersByTimeAsync(1)
    expect(pollOrderStatus).toHaveBeenCalledWith(77)
  })

  it('does not carry local QR dialog i18n fallback maps in the component', () => {
    expect(paymentQRDialogSource).not.toContain('const paymentQRDialogLabelKeys')
    expect(paymentQRDialogSource).toContain('renderPaymentQRDialogText')
    expect(paymentQRDialogSource).toContain('pollIntervalMs?: number')
    expect(paymentQRDialogSource).toContain('DEFAULT_PAYMENT_STATUS_POLL_INTERVAL_MS')
    expect(paymentQRDialogSource).not.toContain('paymentQRDialogFallbackKeys')
    expect(paymentQRDialogSource).not.toContain('payment.qr.payInNewWindow')
    expect(paymentQRDialogSource).not.toContain('payment.result.backToRecharge')
    expect(paymentQRDialogSource).not.toContain('return props.labels?.[key] || key')
    expect(paymentQRDialogSource).not.toContain('30 * 60')
    expect(paymentQRDialogSource).not.toContain('setInterval(pollStatus, 3000)')
  })
})
