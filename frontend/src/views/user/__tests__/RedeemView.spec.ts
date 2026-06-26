import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { readFileSync } from 'node:fs'
import RedeemView from '../RedeemView.vue'

const redeemViewSource = readFileSync('src/views/user/RedeemView.vue', 'utf8')

const redeem = vi.hoisted(() => vi.fn())
const getHistory = vi.hoisted(() => vi.fn())
const getPublicSettings = vi.hoisted(() => vi.fn())
const refreshUser = vi.hoisted(() => vi.fn())
const fetchActiveSubscriptions = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())
const showSuccess = vi.hoisted(() => vi.fn())
const showWarning = vi.hoisted(() => vi.fn())

const appStoreState = vi.hoisted(() => ({
  cachedPublicSettings: {
    pricing_currency_symbol: '€',
    redeem_shell_config: JSON.stringify({
      en: {
        labels: {
          currentBalance: 'Configured balance',
          concurrency: 'Configured concurrency',
          requests: 'configured requests',
          redeemCodeLabel: 'Configured code label',
          redeemCodePlaceholder: 'Configured placeholder',
          redeemCodeHint: 'Configured hint',
          redeemButton: 'Configured redeem',
          redeeming: 'Configured redeeming',
          redeemSuccess: 'Configured success title',
          redeemFailed: 'Configured failed title',
          added: 'Configured added',
          concurrentRequests: 'configured concurrent',
          subscriptionAssigned: 'Configured subscription',
          subscriptionDays: '{days} configured days',
          newBalance: 'Configured new balance',
          newConcurrency: 'Configured new concurrency',
          aboutCodes: 'Configured about',
          codeRule1: 'Configured rule one',
          codeRule2: 'Configured rule two',
          codeRule3: 'Configured support',
          codeRule4: 'Configured rule four',
          recentActivity: 'Configured recent',
          historyWillAppear: 'Configured empty history',
          adminAdjustment: 'Configured admin adjustment',
          balanceAddedRedeem: 'Configured balance redeem',
          codeRedeemSuccess: 'Configured toast success',
          failedToRedeem: 'Configured failed fallback',
        },
      },
    }),
  },
  showError,
  showSuccess,
  showWarning,
}))

const authStoreState = vi.hoisted(() => ({
  user: {
    balance: 3.5,
    concurrency: 7,
  },
  refreshUser,
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, string | number>) => {
        if (!params) return key
        return Object.entries(params).reduce(
          (text, [name, value]) => text.split(`{${name}}`).join(String(value)),
          key,
        )
      },
      locale: { value: 'en-US' },
    }),
  }
})

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => authStoreState,
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => appStoreState,
}))

vi.mock('@/stores/subscriptions', () => ({
  useSubscriptionStore: () => ({
    fetchActiveSubscriptions,
  }),
}))

vi.mock('@/api', () => ({
  redeemAPI: {
    redeem,
    getHistory,
  },
  authAPI: {
    getPublicSettings,
  },
}))

describe('RedeemView', () => {
  beforeEach(() => {
    redeem.mockReset()
    getHistory.mockReset().mockResolvedValue([
      {
        id: 1,
        type: 'balance',
        value: 2,
        used_at: '2026-06-19T00:00:00Z',
        code: 'ABCDEFGH1234',
      },
    ])
    getPublicSettings.mockReset().mockResolvedValue({ contact_info: 'support@example.com' })
    refreshUser.mockReset().mockResolvedValue(undefined)
    fetchActiveSubscriptions.mockReset().mockResolvedValue(undefined)
    showError.mockReset()
    showSuccess.mockReset()
    showWarning.mockReset()
  })

  it('renders redeem shell labels from public settings', async () => {
    const wrapper = mount(RedeemView, {
      global: {
        stubs: {
          AppLayout: { template: '<main><slot /></main>' },
          Icon: { template: '<i />' },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('Configured balance')
    expect(wrapper.text()).toContain('€3.50')
    expect(wrapper.text()).not.toContain('$3.50')
    expect(wrapper.text()).toContain('Configured concurrency: 7 configured requests')
    expect(wrapper.text()).toContain('Configured code label')
    expect(wrapper.get('input#code').attributes('placeholder')).toBe('Configured placeholder')
    expect(wrapper.text()).toContain('Configured hint')
    expect(wrapper.text()).toContain('Configured redeem')
    expect(wrapper.text()).toContain('Configured about')
    expect(wrapper.text()).toContain('Configured rule one')
    expect(wrapper.text()).toContain('Configured support')
    expect(wrapper.text()).toContain('support@example.com')
    expect(wrapper.text()).toContain('Configured recent')
    expect(wrapper.text()).toContain('Configured balance redeem')
  })

  it('uses configured labels for redeem success flow', async () => {
    redeem.mockResolvedValue({
      message: 'server message',
      type: 'subscription',
      value: 30,
      group_name: 'VIP',
      validity_days: 30,
      new_balance: 5,
      new_concurrency: 9,
    })

    const wrapper = mount(RedeemView, {
      global: {
        stubs: {
          AppLayout: { template: '<main><slot /></main>' },
          Icon: { template: '<i />' },
        },
      },
    })

    await flushPromises()
    await wrapper.get('input#code').setValue('CODE123')
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(redeem).toHaveBeenCalledWith('CODE123')
    expect(wrapper.text()).toContain('Configured success title')
    expect(wrapper.text()).toContain('Configured subscription')
    expect(wrapper.text()).toContain('30 configured days')
    expect(wrapper.text()).toContain('Configured new balance')
    expect(wrapper.text()).toContain('€5.00')
    expect(wrapper.text()).toContain('Configured new concurrency')
    expect(showSuccess).toHaveBeenCalledWith('Configured toast success')
    expect(refreshUser).toHaveBeenCalled()
    expect(fetchActiveSubscriptions).toHaveBeenCalledWith(true)
  })

  it('does not keep redeem shell i18n fallback keys in the view bootstrap layer', () => {
    expect(redeemViewSource).not.toContain('redeemFallbackKeys')
    expect(redeemViewSource).not.toContain('redeemLabels.value[key] || key')
    expect(redeemViewSource).not.toContain('redeem.currentBalance')
    expect(redeemViewSource).not.toContain('common.unknown')
    expect(redeemViewSource).not.toContain('${{ user?.balance')
    expect(redeemViewSource).not.toContain('${{ redeemResult.value')
    expect(redeemViewSource).not.toContain('${{ redeemResult.new_balance')
    expect(redeemViewSource).toContain("from './redeemRuntime'")
    expect(redeemViewSource).toContain('resolveRedeemHistoryItemTitle')
    expect(redeemViewSource).toContain('formatRedeemHistoryValue')
    expect(redeemViewSource).toContain('pricing_currency_symbol')
    expect(redeemViewSource).toContain('formatPublicMoneyAmount')
    expect(redeemViewSource).not.toContain('const redeemLabelKeys')
    expect(redeemViewSource).not.toContain('resolveShellLabelOverrides(')
    expect(redeemViewSource).toContain('resolveRedeemShellLabels')
    expect(redeemViewSource).toContain('renderRedeemShellText')
  })

  it('does not keep the legacy redeem locale section in frontend bundles or route meta', () => {
    const zhLocaleSource = readFileSync('src/i18n/locales/zh.ts', 'utf8')
    const enLocaleSource = readFileSync('src/i18n/locales/en.ts', 'utf8')
    const routerSource = readFileSync('src/router/index.ts', 'utf8')

    for (const source of [zhLocaleSource, enLocaleSource]) {
      expect(source).not.toContain('\n  redeem: {')
    }
    expect(routerSource).not.toContain("titleKey: 'redeem.title'")
    expect(routerSource).not.toContain("descriptionKey: 'redeem.description'")
  })
})
