import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'
import { shallowMount } from '@vue/test-utils'
import PaymentMethodSelector from '../PaymentMethodSelector.vue'

const paymentMethodSelectorSource = readFileSync('src/components/payment/PaymentMethodSelector.vue', 'utf8')

function mountSelector() {
  return shallowMount(PaymentMethodSelector, {
    props: {
      methods: [
        { type: 'stripe', fee_rate: 3, available: true },
        { type: 'alipay', fee_rate: 0, available: true },
        { type: 'wxpay_direct', fee_rate: 0, available: true },
      ],
      selected: 'stripe',
      label: '配置支付方式',
      feeLabel: '配置手续费',
      methodLabels: {
        stripe: '配置 Stripe',
        alipay: '配置支付宝',
        wxpay: '配置微信支付',
      },
    },
  })
}

describe('PaymentMethodSelector', () => {
  it('renders labels and method names from external shell labels', () => {
    const wrapper = mountSelector()

    expect(wrapper.text()).toContain('配置支付方式')
    expect(wrapper.text()).toContain('配置手续费 3%')
    expect(wrapper.text()).toContain('配置 Stripe')
    expect(wrapper.text()).toContain('配置支付宝')
    expect(wrapper.text()).toContain('配置微信支付')
  })

  it('does not synthesize an Alipay icon for unknown payment methods', () => {
    const wrapper = shallowMount(PaymentMethodSelector, {
      props: {
        methods: [
          { type: 'custom_provider', fee_rate: 0, available: true },
        ],
        selected: 'custom_provider',
      },
    })

    expect(wrapper.text()).toContain('custom_provider')
    expect(wrapper.find('img').exists()).toBe(false)
  })

  it('does not carry local payment-method i18n fallbacks in the component', () => {
    expect(paymentMethodSelectorSource).not.toContain("t('payment.paymentMethod')")
    expect(paymentMethodSelectorSource).not.toContain("t('payment.fee')")
    expect(paymentMethodSelectorSource).not.toContain('payment.methods.')
    expect(paymentMethodSelectorSource).not.toContain('useI18n')
    expect(paymentMethodSelectorSource).not.toContain("label || 'paymentMethod'")
    expect(paymentMethodSelectorSource).not.toContain("feeLabel || 'fee'")
    expect(paymentMethodSelectorSource).toContain('PaymentMethodLabels')
    expect(paymentMethodSelectorSource).not.toContain('methodLabels?: Record<string, string>')
    expect(paymentMethodSelectorSource).not.toContain('METHOD_ICONS[type] || alipayIcon')
  })
})
