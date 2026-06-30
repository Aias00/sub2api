import { readFileSync } from 'node:fs'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import PricingView from '../PricingView.vue'

const pricingViewSource = readFileSync('src/views/public/PricingView.vue', 'utf8')

const fetchPublicSettings = vi.hoisted(() => vi.fn())
const getPublicCatalog = vi.hoisted(() => vi.fn())
const currentLocale = vi.hoisted(() => ({ value: 'en' }))
const configuredPricingShellConfig = vi.hoisted(() => JSON.stringify({
  en: {
    button: {
      title: 'Configured buy',
    },
    defaults: {
      promptsPath: '/configured-prompts',
      purchasePath: '/configured-purchase',
    },
    groups: [
      { name: 'one-time', title: 'Configured recharge group' },
      { name: 'subscription', title: 'Configured subscription group' },
    ],
    labels: {
      eyebrow: 'Plans',
      title: 'Configurable Pricing',
      description: 'Configured from shell settings.',
      catalogStatus: 'Catalog',
      prompts: 'Configured cases',
      rechargeCta: 'Configured top up',
      emptyRecharge: 'No top ups configured',
    },
  },
}))

const appStoreState = vi.hoisted(() => ({
  publicSettingsLoaded: true,
  siteName: 'Sub2API',
  siteLogo: '',
  cachedPublicSettings: {
    site_name: 'Sub2API',
    site_logo: '',
    pricing_currency_symbol: '$',
    pricing_shell_config: configuredPricingShellConfig,
  },
  fetchPublicSettings,
}))

vi.mock('@/stores', () => ({
  useAppStore: () => appStoreState,
}))

vi.mock('@/i18n', () => ({
  getLocale: () => currentLocale.value,
}))

vi.mock('@/api/payment', () => ({
  paymentAPI: {
    getPublicCatalog,
  },
}))

vi.mock('@/components/layout/PublicDarkHeader.vue', () => ({
  default: {
    template: '<header data-public-dark-header><slot name="actions" /></header>',
  },
}))

describe('PricingView', () => {
  beforeEach(() => {
    fetchPublicSettings.mockReset()
    getPublicCatalog.mockReset()
    currentLocale.value = 'en'
    appStoreState.publicSettingsLoaded = true
    appStoreState.cachedPublicSettings = {
      site_name: 'Sub2API',
      site_logo: '',
      pricing_currency_symbol: '$',
      pricing_shell_config: configuredPricingShellConfig,
    }
    getPublicCatalog.mockResolvedValue({
      data: {
        recharge_products: [
          {
            id: 'starter',
            name: 'Starter',
            amount: 12,
            credited_amount: 120,
            sort_order: 1,
          },
        ],
        plans: [
          {
            id: 1,
            group_id: 10,
            group_platform: 'openai',
            group_name: 'GPT',
            group_display_label: 'Configured Platform',
            quota_label: '$250',
            name: 'Pro Plan',
            description: 'Plan from catalog',
            price: 99,
            validity_days: 30,
            validity_unit: 'day',
            features: [],
            for_sale: true,
            sort_order: 2,
          },
        ],
      },
    })
  })

  it('renders pricing shell labels from public settings', async () => {
    const wrapper = mount(PricingView, {
      global: {
        stubs: {
          RouterLink: {
            props: ['to'],
            template: '<a :href="typeof to === \'string\' ? to : to.path"><slot /></a>',
          },
          LocaleSwitcher: { template: '<div>locale</div>' },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('Plans')
    expect(wrapper.text()).toContain('Configurable Pricing')
    expect(wrapper.text()).toContain('Configured from shell settings.')
    expect(wrapper.text()).toContain('Catalog')
    expect(wrapper.text()).toContain('Configured cases')
    expect(wrapper.text()).toContain('Configured recharge group')
    expect(wrapper.text()).toContain('Configured subscription group')
    expect(wrapper.text()).toContain('Configured top up')
    expect(wrapper.text()).toContain('Configured buy')
    expect(wrapper.find('a[href="/configured-prompts"]').exists()).toBe(true)
    expect(wrapper.find('a[href="/configured-purchase"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('$12')
    await wrapper.findAll('button').find((button) => button.text().includes('Configured subscription group'))?.trigger('click')
    expect(wrapper.text()).toContain('Configured Platform')
    expect(wrapper.text()).toContain('Pro Plan')
    expect(wrapper.text()).toContain('$250')
    expect(getPublicCatalog).toHaveBeenCalledTimes(1)
    expect(fetchPublicSettings).not.toHaveBeenCalled()
  })

  it('does not synthesize a one-month validity fallback when catalog days are missing', async () => {
    getPublicCatalog.mockResolvedValueOnce({
      data: {
        recharge_products: [],
        plans: [
          {
            id: 2,
            group_id: 20,
            group_platform: 'openai',
            group_name: 'GPT',
            name: 'No Duration Plan',
            price: 25,
            validity_days: 0,
            validity_unit: 'month',
            features: [],
            for_sale: true,
            sort_order: 1,
          },
        ],
      },
    })

    const wrapper = mount(PricingView, {
      global: {
        stubs: {
          RouterLink: {
            props: ['to'],
            template: '<a :href="typeof to === \'string\' ? to : to.path"><slot /></a>',
          },
          LocaleSwitcher: { template: '<div>locale</div>' },
        },
      },
    })

    await flushPromises()
    await wrapper.findAll('button').find((button) => button.text().includes('Configured subscription group'))?.trigger('click')

    expect(wrapper.text()).toContain('No Duration Plan')
    expect(wrapper.text()).not.toContain('/1')
  })

  it('does not synthesize product and plan metadata when catalog fields are missing', async () => {
    getPublicCatalog.mockResolvedValueOnce({
      data: {
        recharge_products: [
          {
            id: 'missing-name',
            name: '',
            amount: 10,
            credited_amount: 100,
            sort_order: 1,
          },
        ],
        plans: [
          {
            id: 3,
            group_id: 30,
            group_platform: '',
            group_name: '',
            group_display_label: '',
            quota_label: '',
            name: 'Sparse Plan',
            price: 25,
            validity_days: 0,
            validity_unit: '',
            rate_multiplier: null,
            features: [],
            for_sale: true,
            sort_order: 1,
          },
        ],
      },
    })

    const wrapper = mount(PricingView, {
      global: {
        stubs: {
          RouterLink: {
            props: ['to'],
            template: '<a :href="typeof to === \'string\' ? to : to.path"><slot /></a>',
          },
          LocaleSwitcher: { template: '<div>locale</div>' },
        },
      },
    })

    await flushPromises()
    expect(wrapper.text()).not.toContain('missing-name')
    await wrapper.findAll('button').find((button) => button.text().includes('Configured subscription group'))?.trigger('click')

    expect(wrapper.text()).toContain('Sparse Plan')
    expect(wrapper.text()).not.toContain('x1')
    expect(wrapper.text()).not.toContain('Unlimited')
  })

  it('does not embed default pricing shell copy in the Vue view', () => {
    expect(pricingViewSource).not.toContain('EMPTY_PRICING_COPY')
    expect(pricingViewSource).not.toContain('DEFAULT_PRICING_COPY')
    expect(pricingViewSource).not.toContain("title: '价格与套餐'")
    expect(pricingViewSource).not.toContain("catalogStatus: '目录状态'")
    expect(pricingViewSource).not.toContain("prompts: '提示词案例'")
    expect(pricingViewSource).not.toContain("rechargeCta: '购买充值包'")
    expect(pricingViewSource).not.toContain("title: 'Pricing'")
    expect(pricingViewSource).not.toContain("catalogStatus: 'Catalog status'")
    expect(pricingViewSource).not.toContain("prompts: 'Prompt cases'")
    expect(pricingViewSource).not.toContain("rechargeCta: 'Buy balance'")
    expect(pricingViewSource).not.toContain('function shellLabel(key: keyof PricingShellConfig[\'labels\'], fallback')
    expect(pricingViewSource).not.toContain('function shellGroupLabel(name: string, fallback')
    expect(pricingViewSource).not.toContain("pricing_currency_symbol?.trim() || '¥'")
    expect(pricingViewSource).not.toContain('days || 1')
    expect(pricingViewSource).not.toContain('pricing_title')
    expect(pricingViewSource).not.toContain('pricing_description')
    expect(pricingViewSource).not.toContain('product.name || product.id')
    expect(pricingViewSource).not.toContain('plan.group_display_label || plan.group_name || plan.group_platform || siteName')
    expect(pricingViewSource).not.toContain('plan.rate_multiplier ?? 1')
    expect(pricingViewSource).not.toContain('plan.quota_label || unlimitedLabel')
    expect(pricingViewSource).not.toContain('to="/prompts"')
    expect(pricingViewSource).toContain('PublicDarkHeader')
    expect(pricingViewSource).not.toContain('useAuthRouteDefaults')
    expect(pricingViewSource).not.toContain(':to="authRouteDefaults.homePath"')
    expect(pricingViewSource).not.toContain('to="/home"')
    expect(pricingViewSource).not.toContain("path: '/purchase'")
    expect(pricingViewSource).not.toContain("'/purchase?tab=recharge'")
    expect(pricingViewSource).toContain("from './pricingRuntime'")
    expect(pricingViewSource).toContain('resolvePricingPromptsPath')
    expect(pricingViewSource).toContain('resolvePricingPurchasePath')
    expect(pricingViewSource).toContain('buildPricingPurchaseRoute')
  })
})
