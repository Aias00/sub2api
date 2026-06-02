import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import AvailableGroupsView from '../AvailableGroupsView.vue'

const getAvailable = vi.hoisted(() => vi.fn())
const getUserGroupRates = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, string | number>) => {
        if (params?.amount !== undefined) return `${key}:${params.amount}`
        return key
      },
    }),
  }
})

vi.mock('@/api/groups', () => ({
  default: {
    getAvailable,
    getUserGroupRates,
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
  }),
}))

describe('AvailableGroupsView', () => {
  beforeEach(() => {
    getAvailable.mockReset().mockResolvedValue([
      {
        id: 1,
        name: 'Claude Public',
        description: 'Public Claude access',
        platform: 'anthropic',
        rate_multiplier: 1,
        rpm_limit: 0,
        is_exclusive: false,
        status: 'active',
        subscription_type: 'standard',
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
        created_at: '',
        updated_at: '',
      },
      {
        id: 2,
        name: 'GPT Pro Subscription',
        description: 'Subscription-only GPT access',
        platform: 'openai',
        rate_multiplier: 1.5,
        rpm_limit: 0,
        is_exclusive: false,
        status: 'active',
        subscription_type: 'subscription',
        daily_limit_usd: 20,
        weekly_limit_usd: null,
        monthly_limit_usd: null,
        allow_image_generation: true,
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
        created_at: '',
        updated_at: '',
      },
    ])
    getUserGroupRates.mockReset().mockResolvedValue({ 2: 2 })
    showError.mockReset()
  })

  it('renders public and member-only groups separately and supports search', async () => {
    const wrapper = mount(AvailableGroupsView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          EmptyState: { template: '<div><slot name="icon" />empty</div>' },
          GroupBadge: {
            props: ['name'],
            template: '<div>{{ name }}</div>',
          },
          Icon: { template: '<i />' },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('availableGroups.sections.public.title')
    expect(wrapper.text()).toContain('availableGroups.sections.member.title')
    expect(wrapper.text()).toContain('Claude Public')
    expect(wrapper.text()).toContain('GPT Pro Subscription')
    expect(getAvailable).toHaveBeenCalledTimes(1)
    expect(getUserGroupRates).toHaveBeenCalledTimes(1)

    await wrapper.get('input[type="text"]').setValue('gpt')
    expect(wrapper.text()).toContain('GPT Pro Subscription')
    expect(wrapper.text()).not.toContain('Claude Public')
  })
})
