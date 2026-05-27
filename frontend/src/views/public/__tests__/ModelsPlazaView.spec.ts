import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import ModelsPlazaView from '../ModelsPlazaView.vue'

const authStoreState = vi.hoisted(() => ({
  isAuthenticated: false,
  isAdmin: false,
  checkAuth: vi.fn(),
}))

const copyToClipboard = vi.hoisted(() => vi.fn())
const fetchPublicSettings = vi.hoisted(() => vi.fn())

const appStoreState = vi.hoisted(() => ({
  publicSettingsLoaded: true,
  siteName: 'Sub2API',
  siteLogo: '',
  docUrl: '/docs',
  cachedPublicSettings: {
    site_name: 'cloudbase',
    site_logo: '',
    doc_url: '/docs',
    model_plaza_items: [
      {
        id: 'claude-opus-4-6',
        provider: 'anthropic',
        title: 'Claude Opus 4.6',
        badge: '旗舰',
        description: '复杂推理与长上下文处理',
        capability_tags: ['复杂推理', '代码审查'],
        model_ids: ['claude-opus-4-6'],
        input_price: '¥2.0000 / 1M Tokens',
        output_price: '¥10.0000 / 1M Tokens',
        cache_read_price: '',
        cache_write_price: '',
        billing_badge: '按量计费',
        visible: true,
        sort_order: 10,
      },
      {
        id: 'gpt-5-3-codex',
        provider: 'openai',
        title: 'GPT-5.3 Codex',
        badge: '编码',
        description: '高频代码生成与 agent 调用',
        capability_tags: ['代码生成', 'Agent 调用'],
        model_ids: ['gpt-5.3-codex'],
        input_price: '¥1.2000 / 1M Tokens',
        output_price: '¥6.0000 / 1M Tokens',
        cache_read_price: '',
        cache_write_price: '',
        billing_badge: '按量计费',
        visible: true,
        sort_order: 15,
      },
      {
        id: 'hidden-model',
        provider: 'openai',
        title: 'Hidden',
        badge: '',
        description: '',
        capability_tags: [],
        model_ids: [],
        input_price: '',
        output_price: '',
        cache_read_price: '',
        cache_write_price: '',
        billing_badge: '',
        visible: false,
        sort_order: 20,
      },
    ],
  },
  fetchPublicSettings,
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

vi.mock('@/stores', () => ({
  useAuthStore: () => authStoreState,
  useAppStore: () => appStoreState,
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard,
  }),
}))

describe('ModelsPlazaView', () => {
  beforeEach(() => {
    authStoreState.isAuthenticated = false
    authStoreState.isAdmin = false
    authStoreState.checkAuth.mockReset()
    copyToClipboard.mockReset()
    fetchPublicSettings.mockReset()
    appStoreState.publicSettingsLoaded = true
  })

  it('renders visible plaza cards from public settings', async () => {
    const wrapper = mount(ModelsPlazaView, {
      global: {
        stubs: {
          RouterLink: {
            props: ['to'],
            template: '<a :href="typeof to === \'string\' ? to : to.path"><slot /></a>',
          },
          DocsLink: { template: '<a><slot /></a>' },
          LocaleSwitcher: { template: '<div>locale</div>' },
          Icon: { template: '<i />' },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('Claude Opus 4.6')
    expect(wrapper.text()).toContain('GPT-5.3 Codex')
    expect(wrapper.text()).toContain('复杂推理')
    expect(wrapper.text()).toContain('¥2.0000 / 1M Tokens')
    expect(wrapper.text()).not.toContain('Hidden')
    expect(authStoreState.checkAuth).toHaveBeenCalledTimes(1)
    expect(fetchPublicSettings).not.toHaveBeenCalled()
  })

  it('filters cards by group and search query', async () => {
    const wrapper = mount(ModelsPlazaView, {
      global: {
        stubs: {
          RouterLink: {
            props: ['to'],
            template: '<a :href="typeof to === \'string\' ? to : to.path"><slot /></a>',
          },
          DocsLink: { template: '<a><slot /></a>' },
          LocaleSwitcher: { template: '<div>locale</div>' },
          Icon: { template: '<i />' },
        },
      },
    })

    await flushPromises()

    const groupButtons = wrapper.findAll('button').filter((button) => button.text().includes('GPT'))
    expect(groupButtons).toHaveLength(1)
    await groupButtons[0].trigger('click')
    expect(wrapper.text()).toContain('GPT-5.3 Codex')
    expect(wrapper.text()).not.toContain('Claude Opus 4.6')

    const searchInput = wrapper.get('input[type="search"]')
    await searchInput.setValue('agent')
    expect(wrapper.text()).toContain('GPT-5.3 Codex')

    await searchInput.setValue('不存在的关键词')
    expect(wrapper.text()).toContain('没有匹配的模型卡片')
  })
})
