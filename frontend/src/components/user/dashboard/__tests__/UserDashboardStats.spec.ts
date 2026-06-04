import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import UserDashboardStats from '../UserDashboardStats.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  const messages: Record<string, string> = {
    'common.active': 'active',
    'common.available': 'available',
    'common.total': 'total',
    'dashboard.actual': 'actual',
    'dashboard.apiKeys': 'API keys',
    'dashboard.averageTime': 'average',
    'dashboard.avgResponse': 'Average response',
    'dashboard.balance': 'Balance',
    'dashboard.cacheRead': 'Cache read',
    'dashboard.cacheWrite': 'Cache write',
    'dashboard.input': 'Input',
    'dashboard.output': 'Output',
    'dashboard.performance': 'Performance',
    'dashboard.platformBreakdown': 'Platform breakdown',
    'dashboard.platformCount': '{count} platforms',
    'dashboard.platformOther': 'Other',
    'dashboard.requests': 'Requests',
    'dashboard.standard': 'standard',
    'dashboard.todayCost': 'Today cost',
    'dashboard.todayRequests': 'Today requests',
    'dashboard.todayTokens': 'Today tokens',
    'dashboard.tokens': 'Tokens',
    'dashboard.totalTokens': 'Total tokens',
  }

  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, string | number>) => {
        if (key === 'dashboard.platformCount') return `${params?.count ?? 0} platforms`
        return messages[key] ?? key
      },
      te: (key: string) => key in messages,
      locale: { value: 'en' },
    }),
  }
})

describe('UserDashboardStats', () => {
  it('renders when dashboard usage payload has missing numeric fields', () => {
    const wrapper = mount(UserDashboardStats, {
      props: {
        balance: 0,
        isSimple: false,
        platformQuotas: [],
        stats: {
          total_api_keys: 1,
          active_api_keys: 1,
          total_requests: undefined,
          total_tokens: undefined,
          total_actual_cost: undefined,
          today_actual_cost: undefined,
          by_platform: [
            {
              platform: 'anthropic',
              total_actual_cost: undefined,
              today_actual_cost: undefined,
              total_requests: undefined,
              total_tokens: undefined,
            },
          ],
        } as any,
      },
      global: {
        stubs: {
          Icon: { template: '<i />' },
        },
      },
    })

    expect(wrapper.text()).toContain('Platform breakdown')
    expect(wrapper.text()).toContain('$0.0000')
  })
})
