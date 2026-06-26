import { readFileSync } from 'node:fs'
import { describe, expect, it, vi } from 'vitest'
import { shallowMount } from '@vue/test-utils'
import RechargeProductCard from '../RechargeProductCard.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      locale: { value: 'zh-CN' },
    }),
  }
})

const rechargeProductCardSource = readFileSync('src/components/payment/RechargeProductCard.vue', 'utf8')

function mountCard() {
  return shallowMount(RechargeProductCard, {
    props: {
      product: {
        id: 'starter',
        name: 'Starter',
        description: 'Starter product',
        amount: 10,
        credited_amount: 100,
        recommended: true,
        features: ['Feature one'],
        is_active: true,
        sort_order: 1,
      },
      labels: {
        recommended: '配置推荐',
        creditLine: '配置到账 {amount}',
        cta: '配置购买',
      },
    },
  })
}

describe('RechargeProductCard', () => {
  it('renders recharge product labels from external shell labels', () => {
    const wrapper = mountCard()

    expect(wrapper.text()).toContain('配置推荐')
    expect(wrapper.text()).toContain('配置到账 100.00')
    expect(wrapper.text()).toContain('配置购买')
  })

  it('does not carry local recharge product i18n fallbacks in the component', () => {
    expect(rechargeProductCardSource).not.toContain('payment.rechargeProducts.recommended')
    expect(rechargeProductCardSource).not.toContain('payment.rechargeProducts.creditLine')
    expect(rechargeProductCardSource).not.toContain('payment.rechargeProducts.cta')
    expect(rechargeProductCardSource).not.toContain('rechargeProductRecommended')
    expect(rechargeProductCardSource).not.toContain('rechargeProductCreditLine')
    expect(rechargeProductCardSource).not.toContain('rechargeProductCta')
  })
})
