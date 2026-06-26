import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { readFileSync } from 'node:fs'
import AvailableGroupsView from '../AvailableGroupsView.vue'

const availableGroupsViewSource = readFileSync('src/views/user/AvailableGroupsView.vue', 'utf8')

const getAvailable = vi.hoisted(() => vi.fn())
const getUserGroupRates = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())
const appStoreState = vi.hoisted(() => ({
  cachedPublicSettings: null as null | { available_groups_shell_config?: string },
  showError,
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, string | number>) => {
        if (params?.amount !== undefined) return `${key}:${params.amount}`
        return key
      },
      locale: { value: 'zh-CN' },
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
  useAppStore: () => appStoreState,
}))

function buildAvailableGroupsShellConfig(overrides: Record<string, string> = {}): string {
  return JSON.stringify({
    zh: {
      labels: {
        title: '可用分组',
        description: '查看当前账号可见的模型分组、倍率、额度和订阅访问要求。',
        total: '总分组',
        public: '公开分组',
        memberOnly: '会员专属',
        searchPlaceholder: '搜索分组名称、描述、平台或订阅类型',
        emptyTitle: '没有可用分组',
        emptyDescription: '当前还没有可展示的分组。',
        emptyFilteredDescription: '没有匹配当前搜索条件的分组。',
        publicTitle: '公开分组',
        publicDescription: '这些分组对当前账号可直接使用。',
        memberTitle: '会员或专属分组',
        memberDescription: '这些分组需要订阅、权限或专属配置。',
        publicBadge: '公开',
        subscriptionBadge: '订阅',
        exclusiveBadge: '专属',
        standardBadge: '标准',
        imageEnabledBadge: '支持生图',
        rate: '倍率',
        quota: '额度',
        dailyLimit: '每日 ${amount}',
        weeklyLimit: '每周 ${amount}',
        monthlyLimit: '每月 ${amount}',
        unlimited: '不限',
        loadFailed: '加载失败',
        ...overrides,
      },
    },
  })
}

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
    ])
    getUserGroupRates.mockReset().mockResolvedValue({ 2: 2 })
    showError.mockReset()
    appStoreState.cachedPublicSettings = null
  })

  it('renders public and member-only groups separately and supports search', async () => {
    appStoreState.cachedPublicSettings = {
      available_groups_shell_config: buildAvailableGroupsShellConfig({
        title: '配置可用分组',
        description: '配置描述',
        total: '配置总数',
        public: '配置公开',
        memberOnly: '配置会员',
        searchPlaceholder: '配置搜索',
        publicTitle: '配置公开分组',
        publicDescription: '配置公开描述',
        memberTitle: '配置会员分组',
        memberDescription: '配置会员描述',
        publicBadge: '配置公开徽标',
        subscriptionBadge: '配置订阅徽标',
        standardBadge: '配置标准徽标',
        imageEnabledBadge: '配置生图徽标',
        rate: '配置倍率',
        quota: '配置额度',
        dailyLimit: '配置每日 {amount}',
        unlimited: '配置不限',
      }),
    }
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

    expect(wrapper.text()).toContain('配置可用分组')
    expect(wrapper.text()).toContain('配置描述')
    expect(wrapper.text()).toContain('配置总数')
    expect(wrapper.text()).toContain('配置公开')
    expect(wrapper.text()).toContain('配置会员')
    expect(wrapper.find('input[type="text"]').attributes('placeholder')).toBe('配置搜索')
    expect(wrapper.text()).toContain('配置公开分组')
    expect(wrapper.text()).toContain('配置公开描述')
    expect(wrapper.text()).toContain('配置会员分组')
    expect(wrapper.text()).toContain('配置会员描述')
    expect(wrapper.text()).toContain('配置公开徽标')
    expect(wrapper.text()).toContain('配置订阅徽标')
    expect(wrapper.text()).toContain('配置标准徽标')
    expect(wrapper.text()).toContain('配置生图徽标')
    expect(wrapper.text()).toContain('配置倍率')
    expect(wrapper.text()).toContain('配置额度')
    expect(wrapper.text()).toContain('配置每日 20.00')
    expect(wrapper.text()).toContain('配置不限')
    expect(wrapper.text()).toContain('Claude Public')
    expect(wrapper.text()).toContain('GPT Pro Subscription')
    expect(getAvailable).toHaveBeenCalledTimes(1)
    expect(getUserGroupRates).toHaveBeenCalledTimes(1)

    await wrapper.get('input[type="text"]').setValue('gpt')
    expect(wrapper.text()).toContain('GPT Pro Subscription')
    expect(wrapper.text()).not.toContain('Claude Public')
  })

  it('does not keep available-groups shell i18n fallback keys in the view bootstrap layer', () => {
    expect(availableGroupsViewSource).not.toContain('availableGroupsFallbackKeys')
    expect(availableGroupsViewSource).not.toContain('availableGroupsLabels.value[key] || key')
    expect(availableGroupsViewSource).not.toContain('availableGroups.title')
    expect(availableGroupsViewSource).not.toContain('availableGroups.limits.daily')
    expect(availableGroupsViewSource).toContain("from './availableGroupsRuntime'")
    expect(availableGroupsViewSource).toContain('filterAvailableGroupsByQuery')
    expect(availableGroupsViewSource).toContain('resolveAvailableGroupQuotaSummary')
    expect(availableGroupsViewSource).not.toContain('const availableGroupsLabelKeys')
    expect(availableGroupsViewSource).not.toContain('resolveShellLabelOverrides(')
    expect(availableGroupsViewSource).toContain('resolveAvailableGroupsShellLabels')
    expect(availableGroupsViewSource).toContain('renderAvailableGroupsShellText')
  })

  it('does not keep the legacy available-groups locale section in frontend bundles or route meta', () => {
    const zhLocaleSource = readFileSync('src/i18n/locales/zh.ts', 'utf8')
    const enLocaleSource = readFileSync('src/i18n/locales/en.ts', 'utf8')
    const routerSource = readFileSync('src/router/index.ts', 'utf8')

    for (const source of [zhLocaleSource, enLocaleSource]) {
      expect(source).not.toContain('\n  availableGroups: {')
    }
    expect(routerSource).not.toContain("titleKey: 'availableGroups.title'")
    expect(routerSource).not.toContain("descriptionKey: 'availableGroups.description'")
  })
})
