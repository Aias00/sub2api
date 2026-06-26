import { describe, expect, it } from 'vitest'

import {
  buildPricingPurchaseRoute,
  comparePricingCatalogItems,
  formatPricingCredits,
  formatPricingCurrency,
  resolvePricingBuyButton,
  resolvePricingPlanSourceLabel,
  resolvePricingPlanValidity,
  resolvePricingPromptsPath,
  resolvePricingPurchasePath,
  resolvePricingShellGroupLabel,
  resolvePricingShellLabel,
} from '../pricingRuntime'
import type { PricingCopy, PricingShellConfig } from '@/utils/pricingShell'
import type { SubscriptionPlan } from '@/types/payment'

const copy: PricingCopy = {
  prompts: 'Prompts',
  eyebrow: 'Eyebrow',
  title: 'Title',
  description: 'Description',
  catalogStatus: 'Catalog',
  rechargeProducts: 'Recharge',
  subscriptionPlans: 'Plans',
  recharge: 'Recharge tab',
  subscription: 'Subscription tab',
  buy: 'Buy',
  rechargeCta: 'Recharge CTA',
  subscriptionCta: 'Subscription CTA',
  loadFailed: 'Load failed',
  emptyRecharge: 'Empty recharge',
  emptyPlans: 'Empty plans',
  recommended: 'Recommended',
  creditedBalance: 'Credited balance',
  rate: 'Rate',
  quota: 'Quota',
  unlimited: 'Unlimited',
  day: 'day',
  days: 'days',
  month: 'month',
}

describe('pricingRuntime', () => {
  it('sorts and formats values safely', () => {
    expect(comparePricingCatalogItems({ sort_order: 1 } as any, { sort_order: 3 } as any)).toBeLessThan(0)
    expect(formatPricingCurrency(12, '$')).toBe('$12')
    expect(formatPricingCurrency(12.5, '$')).toBe('$12.50')
    expect(formatPricingCredits(100)).toBe('100')
    expect(formatPricingCredits(100.5)).toBe('100.50')
  })

  it('resolves shell labels, button title, and paths', () => {
    const shellConfig: PricingShellConfig = {
      labels: copy,
      button: { title: 'Configured buy' },
      defaults: {
        promptsPath: '/prompts',
        purchasePath: '/purchase',
      },
      groups: [
        { name: 'one-time', title: 'Configured one time' },
      ],
    }

    expect(resolvePricingShellGroupLabel(shellConfig, 'one-time')).toBe('Configured one time')
    expect(resolvePricingShellLabel(shellConfig, 'catalogStatus')).toBe('Catalog')
    expect(resolvePricingBuyButton(shellConfig, copy)).toBe('Configured buy')
    expect(resolvePricingPromptsPath(shellConfig)).toBe('/prompts')
    expect(resolvePricingPurchasePath(shellConfig)).toBe('/purchase')
    expect(buildPricingPurchaseRoute('/purchase', 'subscription', '10')).toEqual({
      path: '/purchase',
      query: { tab: 'subscription', group: '10' },
    })
  })

  it('resolves subscription plan source and validity without synthesis', () => {
    const plan = {
      group_display_label: '',
      group_name: 'GPT',
      group_platform: 'openai',
      validity_days: 30,
      validity_unit: 'day',
    } as SubscriptionPlan

    expect(resolvePricingPlanSourceLabel(plan)).toBe('GPT')
    expect(resolvePricingPlanValidity(plan, { day: 'day', days: 'days', month: 'month' })).toBe('/30days')
    expect(resolvePricingPlanValidity({ ...plan, validity_days: 0, validity_unit: 'month' } as SubscriptionPlan, {
      day: 'day',
      days: 'days',
      month: 'month',
    })).toBe('')
  })
})
