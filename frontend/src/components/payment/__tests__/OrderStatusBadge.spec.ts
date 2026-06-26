import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'
import { shallowMount } from '@vue/test-utils'
import OrderStatusBadge from '../OrderStatusBadge.vue'

const orderStatusBadgeSource = readFileSync('src/components/payment/OrderStatusBadge.vue', 'utf8')

describe('OrderStatusBadge', () => {
  it('renders status labels from external shell labels', () => {
    const wrapper = shallowMount(OrderStatusBadge, {
      props: {
        status: 'COMPLETED',
        labels: {
          COMPLETED: '配置已完成',
        },
      },
    })

    expect(wrapper.text()).toContain('配置已完成')
  })

  it('falls back to the stable status code when no label is provided', () => {
    const wrapper = shallowMount(OrderStatusBadge, {
      props: {
        status: 'REFUND_REQUESTED',
      },
    })

    expect(wrapper.text()).toContain('REFUND_REQUESTED')
  })

  it('does not carry local order status i18n fallbacks in the component', () => {
    expect(orderStatusBadgeSource).not.toContain('payment.status.pending')
    expect(orderStatusBadgeSource).not.toContain('payment.status.completed')
    expect(orderStatusBadgeSource).not.toContain('useI18n')
  })
})
