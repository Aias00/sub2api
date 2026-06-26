import { readFileSync } from 'node:fs'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

const pollOrderStatus = vi.hoisted(() => vi.fn())
const cancelOrder = vi.hoisted(() => vi.fn())
const verifyOrder = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())
const toCanvas = vi.hoisted(() => vi.fn())

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

vi.mock('@/stores/payment', () => ({
  usePaymentStore: () => ({
    pollOrderStatus,
  }),
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    showError,
  }),
}))

vi.mock('@/api/payment', () => ({
  paymentAPI: {
    cancelOrder,
    verifyOrder,
  },
}))

vi.mock('qrcode', () => ({
  default: {
    toCanvas,
  },
}))

import PaymentStatusPanel from '../PaymentStatusPanel.vue'

const paymentStatusPanelSource = readFileSync('src/components/payment/PaymentStatusPanel.vue', 'utf8')

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
  expires_at: '2099-01-01T12:30:00Z',
  refund_amount: 0,
})

describe('PaymentStatusPanel', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    pollOrderStatus.mockReset()
    cancelOrder.mockReset()
    verifyOrder.mockReset()
    showError.mockReset()
    toCanvas.mockReset().mockResolvedValue(undefined)
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('treats RECHARGING as a successful terminal state', async () => {
    pollOrderStatus.mockResolvedValue(orderFactory('RECHARGING'))

    const wrapper = mount(PaymentStatusPanel, {
      props: {
        orderId: 42,
        qrCode: 'https://pay.example.com/qr/42',
        expiresAt: '2099-01-01T12:30:00Z',
        paymentType: 'alipay',
        orderType: 'balance',
        labels: {
          success: 'Configured success',
        },
      },
      global: {
        stubs: {
          Icon: true,
        },
      },
    })

    await flushPromises()
    await vi.advanceTimersByTimeAsync(3000)
    await flushPromises()

    expect(pollOrderStatus).toHaveBeenCalledWith(42)
    expect(wrapper.text()).toContain('Configured success')
    expect(wrapper.emitted('success')).toHaveLength(1)
  })

  it('uses the provided polling interval for status checks', async () => {
    pollOrderStatus.mockResolvedValue(orderFactory('PENDING'))

    mount(PaymentStatusPanel, {
      props: {
        orderId: 42,
        qrCode: 'https://pay.example.com/qr/42',
        expiresAt: '2099-01-01T12:30:00Z',
        paymentType: 'alipay',
        orderType: 'balance',
        pollIntervalMs: 1234,
      },
      global: {
        stubs: {
          Icon: true,
        },
      },
    })

    await flushPromises()
    await vi.advanceTimersByTimeAsync(1233)
    expect(pollOrderStatus).not.toHaveBeenCalled()

    await vi.advanceTimersByTimeAsync(1)
    expect(pollOrderStatus).toHaveBeenCalledWith(42)
  })

  it('shows reopen button in QR mode when payUrl is also available', async () => {
    const openSpy = vi.spyOn(window, 'open').mockReturnValue({ closed: false } as Window)

    const wrapper = mount(PaymentStatusPanel, {
      props: {
        orderId: 42,
        qrCode: 'https://pay.example.com/qr/42',
        payUrl: 'https://pay.example.com/session/42',
        expiresAt: '2099-01-01T12:30:00Z',
        paymentType: 'alipay',
        orderType: 'balance',
        labels: {
          openPayWindow: 'Configured reopen',
        },
      },
      global: {
        stubs: {
          Icon: true,
        },
      },
    })

    await flushPromises()
    expect(wrapper.text()).toContain('Configured reopen')

    await wrapper.get('button.btn.btn-secondary.text-sm').trigger('click')
    expect(openSpy).toHaveBeenCalledWith(
      'https://pay.example.com/session/42',
      'paymentPopup',
      expect.any(String),
    )

    openSpy.mockRestore()
  })

  it('renders configured shell labels when provided by the parent payment page', async () => {
    const wrapper = mount(PaymentStatusPanel, {
      props: {
        orderId: 42,
        qrCode: 'https://pay.example.com/qr/42',
        payUrl: 'https://pay.example.com/session/42',
        expiresAt: '2099-01-01T12:30:00Z',
        paymentType: 'alipay',
        orderType: 'balance',
        labels: {
          scanAlipay: 'Configured Alipay title',
          scanAlipayHint: 'Configured Alipay hint',
          openPayWindow: 'Configured reopen',
          expiresIn: 'Configured expires',
          waitingPayment: 'Configured waiting',
          cancelOrder: 'Configured cancel',
          errorFallback: 'Configured payment error',
        },
      },
      global: {
        stubs: {
          Icon: true,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('Configured Alipay title')
    expect(wrapper.text()).toContain('Configured Alipay hint')
    expect(wrapper.text()).toContain('Configured reopen')
    expect(wrapper.text()).toContain('Configured expires')
    expect(wrapper.text()).toContain('Configured waiting')
    expect(wrapper.text()).toContain('Configured cancel')
  })

  it('uses configured shell fallback when cancel fails', async () => {
    cancelOrder.mockRejectedValue({})

    const wrapper = mount(PaymentStatusPanel, {
      props: {
        orderId: 42,
        qrCode: 'https://pay.example.com/qr/42',
        payUrl: 'https://pay.example.com/session/42',
        expiresAt: '2099-01-01T12:30:00Z',
        paymentType: 'alipay',
        orderType: 'balance',
        labels: {
          cancelOrder: 'Configured cancel',
          errorFallback: 'Configured payment error',
        },
      },
      global: {
        stubs: {
          Icon: true,
        },
      },
    })

    await flushPromises()
    await wrapper.get('button.btn.btn-secondary.w-full').trigger('click')
    await flushPromises()

    expect(cancelOrder).toHaveBeenCalledWith(42)
    expect(showError).toHaveBeenCalledWith('Configured payment error')
  })

  it('expires immediately instead of synthesizing a local timeout when expiresAt is missing', async () => {
    const wrapper = mount(PaymentStatusPanel, {
      props: {
        orderId: 42,
        qrCode: 'https://pay.example.com/qr/42',
        expiresAt: '',
        paymentType: 'alipay',
        orderType: 'balance',
        labels: {
          expired: 'Configured expired',
        },
      },
      global: {
        stubs: {
          Icon: true,
        },
      },
    })

    await flushPromises()
    await vi.advanceTimersByTimeAsync(3000)
    await flushPromises()

    expect(wrapper.text()).toContain('Configured expired')
    expect(pollOrderStatus).not.toHaveBeenCalled()
    expect(wrapper.emitted('settled')).toEqual([['expired']])
  })

  it('actively verifies a stuck pending order and settles it when upstream confirms payment', async () => {
    pollOrderStatus.mockResolvedValue(orderFactory('PENDING'))
    verifyOrder.mockResolvedValue({
      data: orderFactory('COMPLETED'),
    })

    const wrapper = mount(PaymentStatusPanel, {
      props: {
        orderId: 42,
        qrCode: 'https://pay.example.com/qr/42',
        expiresAt: '2099-01-01T12:30:00Z',
        paymentType: 'wxpay',
        orderType: 'balance',
        labels: {
          success: 'Configured success',
        },
      },
      global: {
        stubs: {
          Icon: true,
        },
      },
    })

    await flushPromises()
    await vi.advanceTimersByTimeAsync(3000)
    await flushPromises()

    expect(pollOrderStatus).toHaveBeenCalledWith(42)
    expect(verifyOrder).toHaveBeenCalledWith('sub2_20260420abcd1234')
    expect(wrapper.text()).toContain('Configured success')
    expect(wrapper.emitted('success')).toHaveLength(1)
  })

  it('does not carry local status-panel i18n fallback maps in the component', () => {
    expect(paymentStatusPanelSource).not.toContain('panelFallbackKeys')
    expect(paymentStatusPanelSource).not.toContain('payment.result.success')
    expect(paymentStatusPanelSource).not.toContain('payment.qr.openPayWindow')
    expect(paymentStatusPanelSource).not.toContain('return props.labels?.[key] || key')
    expect(paymentStatusPanelSource).not.toContain("'$' + paidOrder.amount.toFixed(2)")
    expect(paymentStatusPanelSource).toContain('formatOrderCreditedAmount(paidOrder)')
    expect(paymentStatusPanelSource).not.toContain('30 * 60')
    expect(paymentStatusPanelSource).toContain('pollIntervalMs?: number')
    expect(paymentStatusPanelSource).toContain('DEFAULT_PAYMENT_STATUS_POLL_INTERVAL_MS')
    expect(paymentStatusPanelSource).not.toContain('setInterval(pollStatus, 3000)')
  })
})
