import { readFileSync } from 'node:fs'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import CreditsView from '../CreditsView.vue'

const creditsViewSource = readFileSync('src/views/user/CreditsView.vue', 'utf8')
const fetchPublicSettings = vi.hoisted(() => vi.fn())
const refreshUser = vi.hoisted(() => vi.fn())
const currentLocale = vi.hoisted(() => ({ value: 'en' }))
const configuredCreditsShellConfig = vi.hoisted(() => JSON.stringify({
  en: {
    defaults: {
      purchasePath: '/configured-purchase',
      ordersPath: '/configured-orders',
    },
    labels: {
      eyebrow: 'Configured eyebrow',
      title: 'Wallet overview',
      description: 'Configured balance copy.',
      purchase: 'Buy balance',
      orders: 'Billing history',
      credits: 'Configured balance',
      cloudbaseBalance: 'Configured balance',
      balanceLabel: 'Configured balance: {balance}',
      actionsTitle: 'Configured actions',
      actionsDescription: 'Configured action description.',
      recharge: 'Configured recharge',
      viewOrders: 'Configured orders',
    },
    actions: {
      title: 'Configured action block',
      description: 'Configured action block description.',
    },
    buttons: {
      recharge: 'Configured recharge button',
      orders: 'Configured orders button',
    },
    conversion: 'Configured conversion: {creditsPerBalance} to 1.',
  },
}))

const appStoreState = vi.hoisted(() => ({
  publicSettingsLoaded: true,
  cachedPublicSettings: {
    credits_shell_config: configuredCreditsShellConfig,
    credits_per_balance: '12',
    pricing_currency_symbol: '€',
  },
  fetchPublicSettings,
}))

const authStoreState = vi.hoisted(() => ({
  user: {
    balance: 2.5,
  },
  refreshUser,
}))

vi.mock('@/stores', () => ({
  useAppStore: () => appStoreState,
  useAuthStore: () => authStoreState,
}))

vi.mock('@/i18n', () => ({
  getLocale: () => currentLocale.value,
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key,
  }),
}))

describe('CreditsView', () => {
  beforeEach(() => {
    fetchPublicSettings.mockReset()
    refreshUser.mockReset()
    refreshUser.mockResolvedValue(undefined)
    currentLocale.value = 'en'
    appStoreState.publicSettingsLoaded = true
    appStoreState.cachedPublicSettings = {
      credits_shell_config: configuredCreditsShellConfig,
      credits_per_balance: '12',
      pricing_currency_symbol: '€',
    }
  })

  it('renders balance shell labels from public settings', async () => {
    const wrapper = mount(CreditsView, {
      global: {
        stubs: {
          AppLayout: { template: '<main><slot /></main>' },
          RouterLink: {
            props: ['to'],
            template: '<a :href="typeof to === \'string\' ? to : to.path"><slot /></a>',
          },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('Configured eyebrow')
    expect(wrapper.text()).toContain('Wallet overview')
    expect(wrapper.text()).toContain('Configured balance copy.')
    expect(wrapper.text()).toContain('Buy balance')
    expect(wrapper.text()).toContain('Billing history')
    expect(wrapper.text()).toContain('Configured balance')
    expect(wrapper.text()).toContain('2.50')
    expect(wrapper.text()).toContain('Configured balance')
    expect(wrapper.text()).toContain('€2.50')
    expect(wrapper.text()).not.toContain('$2.50')
    expect(wrapper.text()).toContain('Configured balance: 2.50')
    expect(wrapper.text()).toContain('Configured conversion: 1 to 1.')
    expect(wrapper.text()).toContain('Configured action block')
    expect(wrapper.text()).toContain('Configured action block description.')
    expect(wrapper.text()).toContain('Configured recharge button')
    expect(wrapper.text()).toContain('Configured orders button')
    expect(wrapper.find('a[href="/configured-purchase"]').exists()).toBe(true)
    expect(wrapper.find('a[href="/configured-orders"]').exists()).toBe(true)
    expect(fetchPublicSettings).not.toHaveBeenCalled()
    expect(refreshUser).toHaveBeenCalledTimes(1)
  })

  it('does not embed default credits shell copy in the Vue view', () => {
    expect(creditsViewSource).not.toContain('EMPTY_CREDITS_COPY')
    expect(creditsViewSource).not.toContain('DEFAULT_CREDITS_COPY')
    expect(creditsViewSource).not.toContain("title: '余额'")
    expect(creditsViewSource).not.toContain("purchase: '购买余额'")
    expect(creditsViewSource).not.toContain("orders: '订单记录'")
    expect(creditsViewSource).not.toContain("recharge: '去充值'")
    expect(creditsViewSource).not.toContain("viewOrders: '查看订单'")
    expect(creditsViewSource).not.toContain("title: 'Balance'")
    expect(creditsViewSource).not.toContain("purchase: 'Recharge balance'")
    expect(creditsViewSource).not.toContain("orders: 'Orders'")
    expect(creditsViewSource).not.toContain("recharge: 'Recharge'")
    expect(creditsViewSource).not.toContain("viewOrders: 'View orders'")
    expect(creditsViewSource).not.toContain('parsed : 10')
    expect(creditsViewSource).not.toContain('${{ formattedBalance }}')
    expect(creditsViewSource).not.toContain('credits_title')
    expect(creditsViewSource).not.toContain('credits_description')
    expect(creditsViewSource).not.toContain('credits_purchase_label')
    expect(creditsViewSource).not.toContain('credits_balance_label')
    expect(creditsViewSource).not.toContain('to="/purchase"')
    expect(creditsViewSource).not.toContain('to="/purchase?tab=recharge"')
    expect(creditsViewSource).not.toContain('to="/orders"')
    expect(creditsViewSource).toContain("from './creditsRuntime'")
    expect(creditsViewSource).toContain('resolveCreditsPurchasePath')
    expect(creditsViewSource).toContain('resolveCreditsOrdersPath')
    expect(creditsViewSource).toContain('buildCreditsPurchaseRoute')
    expect(creditsViewSource).toContain('pricing_currency_symbol')
    expect(creditsViewSource).toContain('formatPublicMoneyAmount')
  })
})
