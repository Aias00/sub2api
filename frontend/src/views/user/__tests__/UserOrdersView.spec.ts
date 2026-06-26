import { readFileSync } from 'node:fs'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, shallowMount } from '@vue/test-utils'
import UserOrdersView from '../UserOrdersView.vue'

const routerPush = vi.hoisted(() => vi.fn())
const appStoreState = vi.hoisted(() => ({
  cachedPublicSettings: null as null | { payment_shell_config?: string },
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))
const getMyOrders = vi.hoisted(() => vi.fn())
const getRefundEligibleProviders = vi.hoisted(() => vi.fn())
const cancelOrder = vi.hoisted(() => vi.fn())
const requestRefund = vi.hoisted(() => vi.fn())

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
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

vi.mock('@/api/payment', () => ({
  paymentAPI: {
    getMyOrders,
    getRefundEligibleProviders,
    cancelOrder,
    requestRefund,
  },
}))

const userOrdersViewSource = readFileSync('src/views/user/UserOrdersView.vue', 'utf8')

function mountView() {
  return shallowMount(UserOrdersView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        Select: {
          props: ['options'],
          template: '<div><span v-for="option in options" :key="option.value">{{ option.label }}</span></div>',
        },
        OrderTable: {
          props: ['orders', 'labels'],
          template: '<div>{{ labels.orderId }} {{ labels.createdAt }} {{ labels.actions }}<slot v-for="order in orders" name="actions" :row="order" /></div>',
        },
        Pagination: true,
        BaseDialog: {
          props: ['title'],
          template: '<section><h2>{{ title }}</h2><slot /><footer><slot name="footer" /></footer></section>',
        },
        Icon: true,
      },
    },
  })
}

describe('UserOrdersView', () => {
  beforeEach(() => {
    routerPush.mockReset()
    appStoreState.showError.mockReset()
    appStoreState.showSuccess.mockReset()
    appStoreState.cachedPublicSettings = {
      payment_shell_config: JSON.stringify({
        zh: {
          labels: {
            all: '配置全部',
            pending: '配置待支付',
            completed: '配置已完成',
            failed: '配置失败',
            refunded: '配置已退款',
            backToRecharge: '配置返回充值',
            orderId: '配置订单 ID',
            createdAt: '配置创建时间',
            actions: '配置操作',
            methodStripe: '配置 Stripe',
            statusPending: '配置待支付',
            statusCompleted: '配置已完成',
            cancelSuccess: '配置取消成功',
            refundSuccess: '配置退款成功',
          },
        },
      }),
    }
    getMyOrders.mockReset().mockResolvedValue({ data: { items: [], total: 0 } })
    getRefundEligibleProviders.mockReset().mockResolvedValue({ data: { provider_instance_ids: [] } })
    cancelOrder.mockReset().mockResolvedValue({ data: {} })
    requestRefund.mockReset().mockResolvedValue({ data: {} })
  })

  it('优先使用 public settings 中的订单页和表格文案', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('配置全部')
    expect(wrapper.text()).toContain('配置待支付')
    expect(wrapper.text()).toContain('配置已完成')
    expect(wrapper.text()).toContain('配置失败')
    expect(wrapper.text()).toContain('配置已退款')
    expect(wrapper.text()).toContain('配置返回充值')
    expect(wrapper.text()).toContain('配置订单 ID')
    expect(wrapper.text()).toContain('配置创建时间')
    expect(wrapper.text()).toContain('配置操作')
  })

  it('使用 public settings 中的订单操作反馈文案', async () => {
    getMyOrders.mockResolvedValue({
      data: {
        items: [
          {
            id: 1,
            status: 'PENDING',
            amount: 10,
            pay_amount: 10,
            payment_type: 'stripe',
            provider_instance_id: 'stripe-main',
            created_at: '2026-06-19T00:00:00Z',
          },
          {
            id: 2,
            status: 'COMPLETED',
            amount: 20,
            pay_amount: 20,
            payment_type: 'stripe',
            provider_instance_id: 'stripe-main',
            created_at: '2026-06-19T00:00:00Z',
          },
        ],
        total: 2,
      },
    })
    getRefundEligibleProviders.mockResolvedValue({ data: { provider_instance_ids: ['stripe-main'] } })
    const wrapper = mountView()
    await flushPromises()

    const setupState = (wrapper.vm as any).$?.setupState
    setupState.handleCancel(1)
    await setupState.confirmCancel()
    await flushPromises()

    expect(cancelOrder).toHaveBeenCalledWith(1)
    expect(appStoreState.showSuccess).toHaveBeenCalledWith('配置取消成功')

    setupState.openRefundDialog({
      id: 2,
      status: 'COMPLETED',
      amount: 20,
      pay_amount: 20,
      payment_type: 'stripe',
      provider_instance_id: 'stripe-main',
      created_at: '2026-06-19T00:00:00Z',
    })
    setupState.refundReason = 'refund reason'
    await setupState.confirmRefund()
    await flushPromises()

    expect(requestRefund).toHaveBeenCalledWith(2, { reason: 'refund reason' })
    expect(appStoreState.showSuccess).toHaveBeenCalledWith('配置退款成功')
  })

  it('does not carry local orders-page i18n fallback maps in the view', () => {
    expect(userOrdersViewSource).not.toContain('userOrdersFallbackKeys')
    expect(userOrdersViewSource).not.toContain('userOrdersLabels.value[key] || key')
    expect(userOrdersViewSource).not.toContain('payment.orders.orderId')
    expect(userOrdersViewSource).not.toContain('payment.result.backToRecharge')
    expect(userOrdersViewSource).toContain("from './userOrdersRuntime'")
    expect(userOrdersViewSource).toContain('buildUserOrdersStatusFilters')
    expect(userOrdersViewSource).toContain('buildUserOrdersTableLabels')
    expect(userOrdersViewSource).toContain('canUserOrderRequestRefund')
    expect(userOrdersViewSource).toContain('useAuthRouteDefaults')
    expect(userOrdersViewSource).toContain('router.push(authRouteDefaults.purchasePath)')
    expect(userOrdersViewSource).not.toContain("router.push('/purchase')")
    expect(userOrdersViewSource).not.toContain('${{ refundTarget.amount.toFixed(2) }}')
    expect(userOrdersViewSource).not.toContain('const userOrdersLabelKeys')
    expect(userOrdersViewSource).not.toContain('resolvePaymentShellLabels(')
    expect(userOrdersViewSource).toContain('resolveUserOrdersShellLabels')
    expect(userOrdersViewSource).toContain('renderUserOrdersShellText')
    expect(userOrdersViewSource).toContain('formatOrderCreditedAmount(refundTarget)')
  })
})
