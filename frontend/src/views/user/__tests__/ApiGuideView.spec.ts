import { describe, expect, it, vi, beforeEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { readFileSync } from 'node:fs'

import ApiGuideView from '../ApiGuideView.vue'

const apiGuideViewSource = readFileSync('src/views/user/ApiGuideView.vue', 'utf8')

const { list, copyToClipboard, appStoreState } = vi.hoisted(() => ({
  list: vi.fn(),
  copyToClipboard: vi.fn(),
  appStoreState: {
    cachedPublicSettings: null as null | {
      api_base_url?: string
      auth_shell_config?: string
      api_guide_shell_config?: string
    },
    showError: vi.fn(),
  },
}))

vi.mock('@/api', () => ({
  keysAPI: {
    list,
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => appStoreState,
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({ copyToClipboard }),
}))

vi.mock('vue-router', () => ({
  useRoute: () => ({ query: {} }),
}))

const messages: Record<string, string> = {
  'common.status': 'Status',
  'apiGuide.badge': 'API Guide',
  'apiGuide.title': 'Gateway API Guide',
  'apiGuide.description': 'Guide description',
  'apiGuide.openTester': 'Open Tester',
  'apiGuide.manageKeys': 'Manage API Keys',
  'apiGuide.baseUrl': 'Base URL',
  'apiGuide.currentKey': 'Current Key',
  'apiGuide.noSelection': 'No selection',
  'apiGuide.selectKeyHint': 'Select key',
  'apiGuide.supportedEndpoints': 'Supported Endpoints',
  'apiGuide.noGroupAssigned': 'No group assigned',
  'apiGuide.noKeysTitle': 'No API Keys',
  'apiGuide.noKeysDescription': 'Create an API key',
  'apiGuide.keySelector': 'API Key',
  'apiGuide.keySelectorHint': 'Choose a key',
  'apiGuide.unassignedTitle': 'Unassigned',
  'apiGuide.unassignedDescription': 'Assign a group',
  'apiGuide.keySummary': 'Key Summary',
  'apiGuide.groupName': 'Group Name',
  'apiGuide.platform': 'Platform',
  'apiGuide.status': 'Status',
  'apiGuide.authHeaderTitle': 'Auth Header',
  'apiGuide.authHeaderDescription': 'Auth header description',
  'apiGuide.noEndpointVariants': 'No endpoints',
  'apiGuide.stream': 'Stream',
  'apiGuide.testThisVariant': 'Test this endpoint',
  'apiGuide.endpoint': 'Endpoint',
  'apiGuide.protocol': 'Protocol',
  'apiGuide.defaultModel': 'Default Model',
  'apiGuide.headerMode': 'Header Mode',
  'apiGuide.curlExample': 'curl Example',
  'apiGuide.copyCurl': 'Copy curl',
  'apiGuide.copyCurlSuccess': 'curl copied',
  'apiGuide.defaultPrompt': 'Hello',
  'keys.failedToLoad': 'Failed to load',
  'gateway.platforms.openai': 'OpenAI',
  'gateway.protocols.openai': 'OpenAI',
  'gateway.headerModes.bearer': 'Bearer',
  'gateway.variants.openaiChat.label': 'OpenAI Chat',
  'gateway.variants.openaiChat.description': 'OpenAI chat compatible endpoint',
  'gateway.variants.openaiResponses.label': 'OpenAI Responses',
  'gateway.variants.openaiResponses.description': 'OpenAI responses compatible endpoint',
}

function buildAPIGuideShellConfig(
  overrides: Record<string, string> = {},
  gatewayVariants: Record<string, unknown> = {},
  defaults: Record<string, unknown> = {},
): string {
  const labels = Object.fromEntries(
    Object.entries(messages)
      .filter(([key]) => key.startsWith('apiGuide.'))
      .map(([key, value]) => [key.replace('apiGuide.', ''), value])
  )

  return JSON.stringify({
    zh: {
      labels: {
        ...labels,
        loadKeysFailed: 'Failed to load',
        ...overrides,
      },
      defaults,
      gatewayVariants,
    },
  })
}

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => messages[key] ?? key,
      locale: { value: 'zh-CN' },
    }),
  }
})

const AppLayoutStub = { template: '<div><slot /></div>' }
const SelectStub = {
  props: ['modelValue', 'options', 'placeholder'],
  template: '<div><span>{{ placeholder }}</span><span v-for="option in options" :key="option.value">{{ option.label }}{{ option.description }}</span></div>',
}

describe('ApiGuideView', () => {
  beforeEach(() => {
    list.mockReset()
    copyToClipboard.mockReset()
    appStoreState.showError.mockReset()
    appStoreState.cachedPublicSettings = null
  })

  it('renders stable shell labels from public settings', async () => {
    appStoreState.cachedPublicSettings = {
      api_base_url: 'https://api.example.com',
      api_guide_shell_config: buildAPIGuideShellConfig({
        badge: '配置 API',
        title: '配置调用说明',
        description: '配置描述',
        openTester: '配置测试',
        manageKeys: '配置密钥',
        baseUrl: '配置 Base',
        currentKey: '配置当前密钥',
        supportedEndpoints: '配置端点',
        keySelector: '配置选择密钥',
        keySelectorHint: '配置选择说明',
        keySummary: '配置密钥摘要',
        groupName: '配置分组',
        platform: '配置平台',
        status: '配置状态',
        authHeaderTitle: '配置鉴权',
        authHeaderDescription: '配置鉴权说明',
        stream: '配置流式',
        testThisVariant: '配置测试端点',
        endpoint: '配置 Endpoint',
        protocol: '配置协议',
        defaultModel: '配置默认模型',
        headerMode: '配置 Header',
        curlExample: '配置 curl',
        copyCurl: '配置复制',
      }),
    }
    list.mockResolvedValue({
      items: [
        {
          id: 7,
          user_id: 1,
          key: 'sk-test-1234567890',
          name: 'demo-key',
          group_id: 2,
          status: 'active',
          ip_whitelist: [],
          ip_blacklist: [],
          last_used_at: null,
          quota: 0,
          quota_used: 0,
          expires_at: null,
          created_at: '2026-01-01T00:00:00Z',
          updated_at: '2026-01-01T00:00:00Z',
          group: {
            id: 2,
            name: 'openai-group',
            platform: 'openai',
            allow_messages_dispatch: false,
          },
          rate_limit_5h: 0,
          rate_limit_1d: 0,
          rate_limit_7d: 0,
          usage_5h: 0,
          usage_1d: 0,
          usage_7d: 0,
          window_5h_start: null,
          window_1d_start: null,
          window_7d_start: null,
          reset_5h_at: null,
          reset_1d_at: null,
          reset_7d_at: null,
        },
      ],
      total: 1,
      page: 1,
      page_size: 100,
      pages: 1,
    })

    const wrapper = mount(ApiGuideView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
          EmptyState: true,
          LoadingSpinner: true,
          Select: SelectStub,
          Icon: true,
          RouterLink: {
            props: ['to'],
            template: '<a><slot /></a>',
          },
        },
      },
    })

    await flushPromises()

    const text = wrapper.text()
    expect(text).toContain('配置 API')
    expect(text).toContain('配置调用说明')
    expect(text).toContain('配置描述')
    expect(text).toContain('配置测试')
    expect(text).toContain('配置密钥')
    expect(text).toContain('配置 Base')
    expect(text).toContain('https://api.example.com')
    expect(text).toContain('配置当前密钥')
    expect(text).toContain('配置端点')
    expect(text).toContain('配置选择密钥')
    expect(text).toContain('配置选择说明')
    expect(text).toContain('配置密钥摘要')
    expect(text).toContain('配置分组')
    expect(text).toContain('配置平台')
    expect(text).toContain('配置状态')
    expect(text).toContain('配置鉴权')
    expect(text).toContain('配置鉴权说明')
    expect(text).toContain('配置流式')
    expect(text).toContain('配置测试端点')
    expect(text).toContain('配置 Endpoint')
    expect(text).toContain('配置协议')
    expect(text).toContain('配置默认模型')
    expect(text).toContain('配置 Header')
    expect(text).toContain('配置 curl')
    expect(text).toContain('配置复制')
  })

  it('uses configured gateway default models from public API guide settings', async () => {
    appStoreState.cachedPublicSettings = {
      api_base_url: 'https://api.example.com',
      api_guide_shell_config: buildAPIGuideShellConfig({}, {
        openaiChat: {
          defaultModel: 'configured-gpt-model',
        },
        openaiResponses: {
          defaultModel: 'configured-response-model',
        },
      }),
    }
    list.mockResolvedValue({
      items: [
        {
          id: 7,
          user_id: 1,
          key: 'sk-test-1234567890',
          name: 'demo-key',
          group_id: 2,
          status: 'active',
          ip_whitelist: [],
          ip_blacklist: [],
          last_used_at: null,
          quota: 0,
          quota_used: 0,
          expires_at: null,
          created_at: '2026-01-01T00:00:00Z',
          updated_at: '2026-01-01T00:00:00Z',
          group: {
            id: 2,
            name: 'openai-group',
            platform: 'openai',
            allow_messages_dispatch: false,
          },
          rate_limit_5h: 0,
          rate_limit_1d: 0,
          rate_limit_7d: 0,
          usage_5h: 0,
          usage_1d: 0,
          usage_7d: 0,
          window_5h_start: null,
          window_1d_start: null,
          window_7d_start: null,
          reset_5h_at: null,
          reset_1d_at: null,
          reset_7d_at: null,
        },
      ],
      total: 1,
      page: 1,
      page_size: 100,
      pages: 1,
    })

    const wrapper = mount(ApiGuideView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
          EmptyState: true,
          LoadingSpinner: true,
          Select: SelectStub,
          Icon: true,
          RouterLink: {
            props: ['to'],
            template: '<a><slot /></a>',
          },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('configured-gpt-model')
    expect(wrapper.text()).toContain('configured-response-model')
    expect(wrapper.text()).not.toContain('gpt-4.1')
  })

  it('uses API guide defaults for curl prompt and max tokens', async () => {
    appStoreState.cachedPublicSettings = {
      api_base_url: 'https://api.example.com',
      api_guide_shell_config: buildAPIGuideShellConfig({}, {
        openaiChat: {
          defaultModel: 'configured-gpt-model',
        },
      }, {
        defaultPrompt: 'Configured guide prompt',
        maxTokens: 777,
        apiKeyPageSize: 44,
      }),
    }
    list.mockResolvedValue({
      items: [
        {
          id: 7,
          user_id: 1,
          key: 'sk-test-1234567890',
          name: 'demo-key',
          group_id: 2,
          status: 'active',
          ip_whitelist: [],
          ip_blacklist: [],
          last_used_at: null,
          quota: 0,
          quota_used: 0,
          expires_at: null,
          created_at: '2026-01-01T00:00:00Z',
          updated_at: '2026-01-01T00:00:00Z',
          group: {
            id: 2,
            name: 'openai-group',
            platform: 'openai',
            allow_messages_dispatch: false,
          },
          rate_limit_5h: 0,
          rate_limit_1d: 0,
          rate_limit_7d: 0,
          usage_5h: 0,
          usage_1d: 0,
          usage_7d: 0,
          window_5h_start: null,
          window_1d_start: null,
          window_7d_start: null,
          reset_5h_at: null,
          reset_1d_at: null,
          reset_7d_at: null,
        },
      ],
      total: 1,
      page: 1,
      page_size: 100,
      pages: 1,
    })

    const wrapper = mount(ApiGuideView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
          EmptyState: true,
          LoadingSpinner: true,
          Select: SelectStub,
          Icon: true,
          RouterLink: {
            props: ['to'],
            template: '<a><slot /></a>',
          },
        },
      },
    })

    await flushPromises()

    expect(list).toHaveBeenCalledWith(1, 44)
    expect(wrapper.text()).toContain('Configured guide prompt')
    expect(wrapper.text()).toContain('"max_tokens": 777')
    expect(wrapper.text()).toContain('configured-gpt-model')
    expect(apiGuideViewSource).toContain('apiGuideDefaults.value.defaultPrompt')
    expect(apiGuideViewSource).toContain('apiGuideDefaults.value.maxTokens')
    expect(apiGuideViewSource).toContain('apiGuideDefaults.value.apiKeyPageSize')
    expect(apiGuideViewSource).not.toContain('keysAPI.list(1, 100)')
    expect(apiGuideViewSource).not.toContain("apiGuideText('defaultPrompt')")
  })

  it('does not keep API guide shell i18n fallback keys in the view bootstrap layer', () => {
    expect(apiGuideViewSource).not.toContain('apiGuideFallbackKeys')
    expect(apiGuideViewSource).not.toContain('apiGuide.title')
    expect(apiGuideViewSource).not.toContain('apiTest.stream')
    expect(apiGuideViewSource).not.toContain('keys.failedToLoad')
    expect(apiGuideViewSource).not.toContain("t('common.status')")
    expect(apiGuideViewSource).not.toContain('apiGuideLabels.value[key] || key')
    expect(apiGuideViewSource).not.toContain('const apiGuideLabelKeys')
    expect(apiGuideViewSource).not.toContain('resolveShellLabelOverrides(')
    expect(apiGuideViewSource).toContain('resolveAPIGuideShellLabels')
    expect(apiGuideViewSource).toContain('resolveAPIGuideShellDefaults')
    expect(apiGuideViewSource).toContain('renderAPIGuideShellText')
    expect(apiGuideViewSource).toContain("from './apiGuideRuntime'")
    expect(apiGuideViewSource).toContain('buildApiGuideKeyOptions')
    expect(apiGuideViewSource).toContain('resolveApiGuideAuthHeaderPreview')
    expect(apiGuideViewSource).toContain('useAuthRouteDefaults')
    expect(apiGuideViewSource).toContain(':to="apiGuideDefaults.testPath"')
    expect(apiGuideViewSource).toContain('path: apiGuideDefaults.testPath')
    expect(apiGuideViewSource).toContain(':to="authRouteDefaults.apiKeysPath"')
    expect(apiGuideViewSource).toContain(':action-to="authRouteDefaults.apiKeysPath"')
    expect(apiGuideViewSource).not.toContain('to="/keys"')
    expect(apiGuideViewSource).not.toContain('action-to="/keys"')
    expect(apiGuideViewSource).not.toContain('to="/gateway-test"')
    expect(apiGuideViewSource).not.toContain("path: '/gateway-test'")
  })

  it('does not keep the legacy API guide locale section in frontend bundles or route meta', () => {
    const zhLocaleSource = readFileSync('src/i18n/locales/zh.ts', 'utf8')
    const enLocaleSource = readFileSync('src/i18n/locales/en.ts', 'utf8')
    const routerSource = readFileSync('src/router/index.ts', 'utf8')

    for (const source of [zhLocaleSource, enLocaleSource]) {
      expect(source).not.toContain('\n  apiGuide: {')
    }
    expect(routerSource).not.toContain("titleKey: 'apiGuide.title'")
    expect(routerSource).not.toContain("descriptionKey: 'apiGuide.description'")
  })
})
