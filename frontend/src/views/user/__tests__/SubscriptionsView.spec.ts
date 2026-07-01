import { readFileSync } from 'node:fs'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, shallowMount } from '@vue/test-utils'
import SubscriptionsView from '../SubscriptionsView.vue'
import type { Group, UserSubscription } from '@/types'

const routerPush = vi.hoisted(() => vi.fn())
const appStoreState = vi.hoisted(() => ({
  cachedPublicSettings: null as null | {
    auth_shell_config?: string
    payment_shell_config?: string
    pricing_currency_symbol?: string
  },
  showError: vi.fn(),
}))
const getMySubscriptions = vi.hoisted(() => vi.fn())

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
      t: (key: string, params?: Record<string, unknown>) => {
        if (params) {
          return Object.entries(params).reduce(
            (result, [name, value]) => result.replaceAll(`{${name}}`, String(value)),
            key,
          )
        }
        return key
      },
      locale: { value: 'zh-CN' },
    }),
  }
})

vi.mock('@/stores/app', () => ({
  useAppStore: () => appStoreState,
}))

vi.mock('@/api/subscriptions', () => ({
  default: {
    getMySubscriptions,
  },
}))

const subscriptionsViewSource = readFileSync('src/views/user/SubscriptionsView.vue', 'utf8')

function groupFactory(overrides: Partial<Group> = {}): Group {
  return {
    id: 12,
    name: 'OpenAI Group',
    description: 'Group description',
    platform: 'openai',
    rate_multiplier: 1,
    is_exclusive: false,
    status: 'active',
    subscription_type: 'subscription',
    daily_limit_usd: null,
    weekly_limit_usd: null,
    monthly_limit_usd: null,
    allow_image_generation: false,
    image_rate_independent: false,
    image_rate_multiplier: 1,
    image_price_1k: null,
    image_price_2k: null,
    image_price_4k: null,
    claude_code_only: false,
    fallback_group_id: null,
    fallback_group_id_on_invalid_request: null,
    require_oauth_only: false,
    require_privacy_set: false,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    ...overrides,
  }
}

function subscriptionFactory(overrides: Partial<UserSubscription> = {}): UserSubscription {
  return {
    id: 34,
    user_id: 5,
    group_id: 12,
    status: 'active',
    starts_at: '2026-01-01T00:00:00Z',
    daily_usage_usd: 0,
    weekly_usage_usd: 0,
    monthly_usage_usd: 0,
    daily_window_start: null,
    weekly_window_start: null,
    monthly_window_start: null,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    expires_at: null,
    group: groupFactory(),
    ...overrides,
  }
}

function mountView() {
  return shallowMount(SubscriptionsView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        Icon: true,
      },
    },
  })
}

describe('SubscriptionsView', () => {
  beforeEach(() => {
    routerPush.mockReset()
    appStoreState.showError.mockReset()
    appStoreState.cachedPublicSettings = {
      auth_shell_config: JSON.stringify({
        zh: {
          defaults: {
            purchasePath: '/configured-purchase',
          },
        },
      }),
      pricing_currency_symbol: '€',
      payment_shell_config: JSON.stringify({
        zh: {
          labels: {
            subscriptionNoActive: '配置暂无订阅',
            subscriptionNoActiveDesc: '配置暂无订阅说明',
            subscriptionStatusActive: '配置有效',
            renewNow: '配置续费',
            subscriptionExpires: '配置到期',
            subscriptionNoExpiration: '配置无到期',
            subscriptionUnlimited: '配置无限制',
            subscriptionUnlimitedDesc: '配置无限制说明',
          },
        },
      }),
    }
    getMySubscriptions.mockReset()
  })

  it('空状态优先使用 public settings 中的订阅页文案', async () => {
    getMySubscriptions.mockResolvedValue([])

    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('配置暂无订阅')
    expect(wrapper.text()).toContain('配置暂无订阅说明')
  })

  it('空状态在未配置 payment shell 时使用中文默认文案', async () => {
    appStoreState.cachedPublicSettings = {
      auth_shell_config: JSON.stringify({
        zh: {
          defaults: {
            purchasePath: '/configured-purchase',
          },
        },
      }),
      pricing_currency_symbol: '€',
    }
    getMySubscriptions.mockResolvedValue([])

    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('暂无有效订阅')
    expect(wrapper.text()).toContain('没有有效订阅')
    expect(wrapper.text()).not.toContain('subscriptionNoActive')
    expect(wrapper.text()).not.toContain('subscriptionNoActiveDesc')
  })

  it('订阅卡片优先使用 public settings 中的订阅页文案', async () => {
    getMySubscriptions.mockResolvedValue([
      subscriptionFactory({
        daily_usage_usd: 1.25,
        weekly_usage_usd: 2.5,
        monthly_usage_usd: 3.75,
        group: groupFactory({
          daily_limit_usd: 10,
          weekly_limit_usd: 20,
          monthly_limit_usd: 30,
        }),
      }),
      subscriptionFactory({
        id: 35,
        group_id: 13,
        group: groupFactory({ id: 13, name: 'Unlimited group' }),
      }),
    ])

    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('配置有效')
    expect(wrapper.text()).toContain('配置续费')
    expect(wrapper.text()).toContain('配置到期')
    expect(wrapper.text()).toContain('配置无到期')
    expect(wrapper.text()).toContain('配置无限制')
    expect(wrapper.text()).toContain('配置无限制说明')
    expect(wrapper.text()).toContain('€1.25 / €10.00')
    expect(wrapper.text()).toContain('€2.50 / €20.00')
    expect(wrapper.text()).toContain('€3.75 / €30.00')
    expect(wrapper.text()).not.toContain('$1.25')
  })

  it('uses auth route defaults for subscription renewal navigation', async () => {
    getMySubscriptions.mockResolvedValue([subscriptionFactory({ group_id: 19 })])

    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('button.rounded-lg').trigger('click')

    expect(routerPush).toHaveBeenCalledWith({
      path: '/configured-purchase',
      query: { tab: 'subscription', group: '19' },
    })
  })

  it('does not carry local subscriptions-page i18n fallback maps in the view', () => {
    expect(subscriptionsViewSource).not.toContain('const subscriptionLabelKeys')
    expect(subscriptionsViewSource).not.toContain('resolvePaymentShellLabels(')
    expect(subscriptionsViewSource).toContain('resolveSubscriptionLabels')
    expect(subscriptionsViewSource).toContain('renderSubscriptionText')
    expect(subscriptionsViewSource).toContain("from './subscriptionsRuntime'")
    expect(subscriptionsViewSource).toContain('resolveSubscriptionStatusText')
    expect(subscriptionsViewSource).toContain('formatSubscriptionExpirationDate')
    expect(subscriptionsViewSource).not.toContain('subscriptionFallbackKeys')
    expect(subscriptionsViewSource).not.toContain('payment.renewNow')
    expect(subscriptionsViewSource).not.toContain('userSubscriptions.noActiveSubscriptions')
    expect(subscriptionsViewSource).not.toContain('${{ (subscription.daily_usage_usd')
    expect(subscriptionsViewSource).not.toContain('${{ (subscription.weekly_usage_usd')
    expect(subscriptionsViewSource).not.toContain('${{ (subscription.monthly_usage_usd')
    expect(subscriptionsViewSource).toContain('pricing_currency_symbol')
    expect(subscriptionsViewSource).toContain('formatPublicMoneyAmount')
    expect(subscriptionsViewSource).toContain('useAuthRouteDefaults')
    expect(subscriptionsViewSource).toContain('path: authRouteDefaults.purchasePath')
    expect(subscriptionsViewSource).not.toContain("path: '/purchase'")
  })
})
