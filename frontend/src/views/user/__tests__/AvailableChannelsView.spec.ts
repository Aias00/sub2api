import { readFileSync } from 'node:fs'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import AvailableChannelsView from '../AvailableChannelsView.vue'

const availableChannelsViewSource = readFileSync('src/views/user/AvailableChannelsView.vue', 'utf8')
const availableChannelsTableSource = readFileSync('src/components/channels/AvailableChannelsTable.vue', 'utf8')
const groupBadgeSource = readFileSync('src/components/common/GroupBadge.vue', 'utf8')
const groupOptionItemSource = readFileSync('src/components/common/GroupOptionItem.vue', 'utf8')
const getAvailable = vi.hoisted(() => vi.fn())
const getUserGroupRates = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())
const appStoreState = vi.hoisted(() => ({
  cachedPublicSettings: null as null | { available_channels_shell_config?: string },
  showError,
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
      locale: { value: 'zh-CN' },
    }),
  }
})

vi.mock('@/api/channels', () => ({
  default: {
    getAvailable,
  },
}))

vi.mock('@/api/groups', () => ({
  default: {
    getUserGroupRates,
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => appStoreState,
}))

describe('AvailableChannelsView', () => {
  beforeEach(() => {
    getAvailable.mockReset().mockResolvedValue([
      {
        name: 'Claude Channel',
        description: 'Anthropic models',
        platforms: [
          {
            platform: 'anthropic',
            groups: [
              {
                id: 1,
                name: 'Claude Public',
                platform: 'anthropic',
                subscription_type: 'standard',
                rate_multiplier: 1,
                is_exclusive: false,
              },
              {
                id: 2,
                name: 'Claude Exclusive',
                platform: 'anthropic',
                subscription_type: 'subscription',
                rate_multiplier: 1.2,
                is_exclusive: true,
              },
            ],
            supported_models: [
              {
                name: 'claude-opus',
                platform: 'anthropic',
                pricing: null,
              },
            ],
          },
        ],
      },
      {
        name: 'GPT Channel',
        description: 'OpenAI models',
        platforms: [
          {
            platform: 'openai',
            groups: [],
            supported_models: [
              {
                name: 'gpt-5',
                platform: 'openai',
                pricing: null,
              },
            ],
          },
        ],
      },
    ])
    getUserGroupRates.mockReset().mockResolvedValue({ 2: 1.5 })
    showError.mockReset()
    appStoreState.cachedPublicSettings = null
  })

  it('passes Cloudbase shell labels into the available channels table and search', async () => {
    appStoreState.cachedPublicSettings = {
      available_channels_shell_config: JSON.stringify({
        zh: {
          labels: {
            searchPlaceholder: '配置搜索渠道',
            refreshTitle: '配置刷新',
            noPricing: '配置无定价',
            noModels: '配置无模型',
            empty: '配置空态',
            exclusive: '配置专属',
            exclusiveTooltip: '配置专属提示',
            public: '配置公开',
            publicTooltip: '配置公开提示',
            pricing: {
              billingMode: '配置计费模式',
              billingModeImage: '配置按图片',
              billingModePerRequest: '配置按次',
              billingModeToken: '配置按 Token',
              cacheReadPrice: '配置缓存读取',
              cacheWritePrice: '配置缓存写入',
              imageOutputPrice: '配置图片输出',
              inputPrice: '配置输入',
              intervals: '配置阶梯定价',
              outputPrice: '配置输出',
              perRequestPrice: '配置每次请求',
              unitPerMillion: '配置每百万',
              unitPerRequest: '配置每次',
            },
            columns: {
              name: '配置渠道名',
              description: '配置描述',
              platform: '配置平台',
              groups: '配置分组',
              supportedModels: '配置支持模型',
            },
          },
        },
      }),
    }

    const wrapper = mount(AvailableChannelsView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          Icon: { template: '<i />' },
          AvailableChannelsTable: {
            props: [
              'columns',
              'rows',
              'noPricingLabel',
              'noModelsLabel',
              'emptyLabel',
              'exclusiveLabel',
              'exclusiveTooltipLabel',
              'publicLabel',
              'publicTooltipLabel',
              'pricingLabels',
            ],
            template: `
              <div>
                <span>{{ columns.name }}</span>
                <span>{{ columns.description }}</span>
                <span>{{ columns.platform }}</span>
                <span>{{ columns.groups }}</span>
                <span>{{ columns.supportedModels }}</span>
                <span>{{ noPricingLabel }}</span>
                <span>{{ noModelsLabel }}</span>
                <span>{{ emptyLabel }}</span>
                <span>{{ exclusiveLabel }}</span>
                <span>{{ exclusiveTooltipLabel }}</span>
                <span>{{ publicLabel }}</span>
                <span>{{ publicTooltipLabel }}</span>
                <span>{{ pricingLabels.billingMode }}</span>
                <span>{{ pricingLabels.billingModeToken }}</span>
                <span>{{ pricingLabels.inputPrice }}</span>
                <span>{{ pricingLabels.unitPerMillion }}</span>
                <span v-for="row in rows" :key="row.name">{{ row.name }}</span>
              </div>
            `,
          },
        },
      },
    })

    await flushPromises()

    expect(wrapper.find('input[type="text"]').attributes('placeholder')).toBe('配置搜索渠道')
    expect(wrapper.find('button').attributes('title')).toBe('配置刷新')
    expect(wrapper.text()).toContain('配置渠道名')
    expect(wrapper.text()).toContain('配置描述')
    expect(wrapper.text()).toContain('配置平台')
    expect(wrapper.text()).toContain('配置分组')
    expect(wrapper.text()).toContain('配置支持模型')
    expect(wrapper.text()).toContain('配置无定价')
    expect(wrapper.text()).toContain('配置无模型')
    expect(wrapper.text()).toContain('配置空态')
    expect(wrapper.text()).toContain('配置专属')
    expect(wrapper.text()).toContain('配置专属提示')
    expect(wrapper.text()).toContain('配置公开')
    expect(wrapper.text()).toContain('配置公开提示')
    expect(wrapper.text()).toContain('配置计费模式')
    expect(wrapper.text()).toContain('配置按 Token')
    expect(wrapper.text()).toContain('配置输入')
    expect(wrapper.text()).toContain('配置每百万')
    expect(wrapper.text()).toContain('Claude Channel')
    expect(wrapper.text()).toContain('GPT Channel')

    await wrapper.get('input[type="text"]').setValue('gpt')
    expect(wrapper.text()).not.toContain('Claude Channel')
    expect(wrapper.text()).toContain('GPT Channel')
  })

  it('does not embed default available-channels shell copy in the Vue view', () => {
    expect(availableChannelsViewSource).not.toContain('defaultAvailableChannelsLabels')
    expect(availableChannelsViewSource).not.toContain("searchPlaceholder: '搜索渠道或模型...'")
    expect(availableChannelsViewSource).not.toContain("searchPlaceholder: 'Search channels or models...'")
    expect(availableChannelsViewSource).not.toMatch(/resolveAvailableChannelsShellLabels\([^\n]*,[^\n]*,[^\n]*,[^\n]*\)/)
    expect(availableChannelsViewSource).toContain("from './availableChannelsRuntime'")
    expect(availableChannelsViewSource).toContain('buildAvailableChannelsColumnLabels')
    expect(availableChannelsViewSource).toContain('buildAvailableChannelsPricingLabels')
    expect(availableChannelsViewSource).toContain('filterAvailableChannelsByQuery')
    expect(availableChannelsViewSource).not.toContain('const availableChannelsLabelKeys')
    expect(availableChannelsViewSource).toContain('resolveConfiguredAvailableChannelsShellLabels')
  })

  it('does not keep the legacy available channels locale section in frontend bundles or route meta', () => {
    const zhLocaleSource = readFileSync('src/i18n/locales/zh.ts', 'utf8')
    const enLocaleSource = readFileSync('src/i18n/locales/en.ts', 'utf8')
    const routerSource = readFileSync('src/router/index.ts', 'utf8')

    for (const source of [zhLocaleSource, enLocaleSource]) {
      expect(source).not.toContain('\n  availableChannels: {')
    }
    expect(routerSource).not.toContain("titleKey: 'availableChannels.title'")
    expect(routerSource).not.toContain("descriptionKey: 'availableChannels.description'")
    expect(availableChannelsViewSource).not.toContain('availableChannels.pricing')
  })

  it('does not synthesize missing group subscription types as standard in available-channel UI', () => {
    expect(availableChannelsTableSource).not.toContain("g.subscription_type || 'standard'")
    expect(availableChannelsTableSource).not.toContain('g.subscription_type || "standard"')
    expect(availableChannelsTableSource).toContain(':subscription-type="g.subscription_type"')
    expect(groupBadgeSource).not.toContain("subscriptionType: 'standard'")
    expect(groupBadgeSource).not.toContain('subscriptionType: "standard"')
    expect(groupOptionItemSource).not.toContain("subscriptionType: 'standard'")
    expect(groupOptionItemSource).not.toContain('subscriptionType: "standard"')
  })
})
