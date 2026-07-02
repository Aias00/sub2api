import { readFileSync } from 'node:fs'
import { describe, expect, it, vi } from 'vitest'
import { shallowMount } from '@vue/test-utils'
import OrderTable from '../OrderTable.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      locale: { value: 'zh-CN' },
    }),
  }
})

const orderTableSource = readFileSync('src/components/payment/OrderTable.vue', 'utf8')

function mountTable(orderOverrides: Record<string, unknown> = {}) {
  return shallowMount(OrderTable, {
    props: {
      orders: [{
        id: 1,
        user_id: 2,
        amount: 10,
        pay_amount: 11,
        fee_rate: 10,
        payment_type: 'stripe',
        out_trade_no: 'sub2_order_1',
        status: 'COMPLETED',
        order_type: 'balance',
        created_at: '2026-06-19T00:00:00Z',
        expires_at: '2026-06-19T00:30:00Z',
        refund_amount: 0,
        currency: 'CNY',
        ...orderOverrides,
      }],
      loading: false,
      showUser: true,
      labels: {
        orderId: '配置订单 ID',
        orderNo: '配置订单号',
        user: '配置用户',
        payAmount: '配置支付金额',
        paymentMethod: '配置支付方式',
        status: '配置状态',
        createdAt: '配置创建时间',
        actions: '配置操作',
        fee: '配置手续费',
        creditedAmount: '配置到账',
        methodStripe: '配置 Stripe',
        statusCompleted: '配置已完成',
      },
    },
    global: {
      stubs: {
        DataTable: {
          props: ['columns', 'data', 'loading'],
          template: `
            <div>
              <span v-for="column in columns" :key="column.key">{{ column.label }}</span>
              <slot name="cell-user_email" :value="data[0].user_email" :row="data[0]" />
              <slot name="cell-payment_type" :value="data[0].payment_type" :row="data[0]" />
              <slot name="cell-pay_amount" :value="data[0].pay_amount" :row="data[0]" />
              <slot name="cell-status" :value="data[0].status" :row="data[0]" />
            </div>
          `,
        },
        OrderStatusBadge: {
          props: ['status', 'labels'],
          template: '<span>{{ labels[status] }}</span>',
        },
      },
    },
  })
}

describe('OrderTable', () => {
  it('renders labels and known payment method names from external shell labels', () => {
    const wrapper = mountTable()

    expect(wrapper.text()).toContain('配置订单 ID')
    expect(wrapper.text()).toContain('配置订单号')
    expect(wrapper.text()).toContain('配置用户')
    expect(wrapper.text()).toContain('配置支付金额')
    expect(wrapper.text()).toContain('配置支付方式')
    expect(wrapper.text()).toContain('配置状态')
    expect(wrapper.text()).toContain('配置创建时间')
    expect(wrapper.text()).toContain('配置操作')
    expect(wrapper.text()).toContain('配置 Stripe')
    expect(wrapper.text()).toContain('配置已完成')
    expect(wrapper.text()).toContain('配置手续费')
    expect(wrapper.text()).toContain('配置到账')
  })

  it('does not synthesize a user display value from the local user id', () => {
    const wrapper = mountTable({
      user_id: 42,
      user_email: '',
      user_name: '',
    })

    expect(wrapper.text()).not.toContain('#42')
  })

  it('uses explicit user email or name when Cloudbase provides it', () => {
    expect(mountTable({ user_email: 'billing@example.com', user_name: 'Configured User' }).text()).toContain('billing@example.com')
    expect(mountTable({ user_email: '', user_name: 'Configured User' }).text()).toContain('Configured User')
  })

  it('does not carry local order-table payment i18n fallback maps in the component', () => {
    expect(orderTableSource).not.toContain('const orderTableLabelKeys')
    expect(orderTableSource).toContain('renderOrderTableText')
    expect(orderTableSource).not.toContain('orderTableFallbackKeys')
    expect(orderTableSource).not.toContain('payment.orders.orderId')
    expect(orderTableSource).not.toContain('payment.methods.')
    expect(orderTableSource).not.toContain('return props.labels?.[key] || key')
    expect(orderTableSource).not.toContain("'$' + row.amount.toFixed(2)")
    expect(orderTableSource).toContain('formatOrderCreditedAmount(row)')
    expect(orderTableSource).not.toContain("value || row.user_name || '#' + row.user_id")
    expect(orderTableSource).not.toContain("'#' + row.user_id")
  })
})
