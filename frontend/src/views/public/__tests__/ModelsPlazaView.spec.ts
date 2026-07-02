import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { readFileSync } from 'node:fs'
import ModelsPlazaView from '../ModelsPlazaView.vue'

const modelsPlazaViewSource = readFileSync('src/views/public/ModelsPlazaView.vue', 'utf8')

const authStoreState = vi.hoisted(() => ({
  isAuthenticated: false,
  isAdmin: false,
  checkAuth: vi.fn(),
}))

const copyToClipboard = vi.hoisted(() => vi.fn())
const fetchPublicSettings = vi.hoisted(() => vi.fn())
const currentLocale = vi.hoisted(() => ({ value: 'zh-CN' }))

const appStoreState = vi.hoisted(() => ({
  publicSettingsLoaded: true,
  siteName: 'Cloudbase',
  siteLogo: '',
  docUrl: '/docs',
  cachedPublicSettings: {
    site_name: 'cloudbase',
    site_logo: '',
    doc_url: '/docs',
    model_plaza_shell_config: JSON.stringify({
      zh: {
        labels: {
          badge: '模型展示',
          title: '可售模型',
          description: '由 public settings 管理展示文案。',
          quickFind: '配置搜索',
          inputPrice: '输入价',
          emptyFilteredTitle: '没有匹配的模型卡片',
        },
      },
    }),
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
      locale: currentLocale,
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
    currentLocale.value = 'zh-CN'
    appStoreState.cachedPublicSettings.model_plaza_shell_config = JSON.stringify({
      zh: {
        labels: {
          badge: '模型展示',
          title: '可售模型',
          description: '由 public settings 管理展示文案。',
          quickFind: '配置搜索',
          inputPrice: '输入价',
          emptyFilteredTitle: '没有匹配的模型卡片',
        },
      },
    })
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
    expect(wrapper.text()).toContain('模型展示')
    expect(wrapper.text()).toContain('可售模型')
    expect(wrapper.text()).toContain('配置搜索')
    expect(wrapper.text()).toContain('输入价 ¥2.0000 / 1M Tokens')
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

  it('does not keep locale-specific model plaza fallback copy in the view bootstrap layer', () => {
    expect(modelsPlazaViewSource).toContain('PublicDarkHeader')
    expect(modelsPlazaViewSource).not.toContain('DocsLink')
    expect(modelsPlazaViewSource).not.toContain('copy.viewDocs')
    expect(modelsPlazaViewSource).not.toContain('useAuthRouteDefaults')
    expect(modelsPlazaViewSource).not.toContain(':to="authRouteDefaults.homePath"')
    expect(modelsPlazaViewSource).not.toContain('to="/home"')
    expect(modelsPlazaViewSource).not.toContain("isAuthenticated ? dashboardPath : '/login'")
    expect(modelsPlazaViewSource).not.toContain("authStore.isAdmin ? '/admin/dashboard' : '/dashboard'")
    expect(modelsPlazaViewSource).not.toContain('EMPTY_MODELS_PLAZA_COPY')
    expect(modelsPlazaViewSource).not.toContain('DEFAULT_MODELS_PLAZA_COPY')
    expect(modelsPlazaViewSource).not.toContain("badge: '模型广场'")
    expect(modelsPlazaViewSource).not.toContain("title: '公开模型目录'")
    expect(modelsPlazaViewSource).not.toContain("searchPlaceholder: '搜索模型、能力或标签'")
    expect(modelsPlazaViewSource).not.toContain("badge: 'Model Plaza'")
    expect(modelsPlazaViewSource).not.toContain("title: 'Public Model Catalog'")
    expect(modelsPlazaViewSource).not.toContain("searchPlaceholder: 'Search models, capabilities, or tags'")
    expect(modelsPlazaViewSource).not.toContain("const activeGroup = ref('all')")
    expect(modelsPlazaViewSource).not.toContain("return normalized || 'other'")
    expect(modelsPlazaViewSource).not.toContain("|| 'M'")
    expect(modelsPlazaViewSource).not.toContain('type ModelsPlazaCopy')
    expect(modelsPlazaViewSource).not.toContain('const modelsPlazaCopyKeys')
    expect(modelsPlazaViewSource).not.toContain('function formatTemplate')
    expect(modelsPlazaViewSource).toContain("from './modelsPlazaRuntime'")
    expect(modelsPlazaViewSource).toContain('resolveVisibleModelsPlazaItems')
    expect(modelsPlazaViewSource).toContain('resolveModelsPlazaGroupOptions')
    expect(modelsPlazaViewSource).toContain('filterModelsPlazaItems')
    expect(modelsPlazaViewSource).toContain('resolveModelPlazaProviderInitial')
    expect(modelsPlazaViewSource).toContain('resolveModelsPlazaCopy')
  })

  it('does not keep the legacy model plaza locale section in frontend bundles', () => {
    const zhLocaleSource = readFileSync('src/i18n/locales/zh.ts', 'utf8')
    const enLocaleSource = readFileSync('src/i18n/locales/en.ts', 'utf8')
    for (const source of [zhLocaleSource, enLocaleSource]) {
      expect(source).not.toContain('\n  modelsPlaza: {')
    }
  })
})
