import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { readFileSync } from 'node:fs'

import ApiTestView from '../ApiTestView.vue'

const apiTestViewSource = readFileSync('src/views/user/ApiTestView.vue', 'utf8')

const { keyList, usageQuery, copyToClipboard, appStoreState } = vi.hoisted(() => ({
  keyList: vi.fn(),
  usageQuery: vi.fn(),
  copyToClipboard: vi.fn(),
  appStoreState: {
    cachedPublicSettings: null as null | {
      api_base_url?: string
      auth_shell_config?: string
      api_guide_shell_config?: string
      api_test_shell_config?: string
    },
    showError: vi.fn(),
  },
}))

vi.mock('@/api', () => ({
  keysAPI: {
    list: keyList,
  },
  usageAPI: {
    query: usageQuery,
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
  'apiTest.badge': 'Live Request',
  'apiTest.title': 'API Test',
  'apiTest.description': 'Test description',
  'apiTest.openGuide': 'Open Guide',
  'apiTest.send': 'Send',
  'apiTest.sending': 'Sending',
  'apiTest.keySelector': 'Key',
  'apiTest.noSelection': 'No selection',
  'apiTest.noGroupAssigned': 'No group',
  'apiTest.protocol': 'Protocol',
  'apiTest.model': 'Model',
  'apiTest.loading': 'Loading',
  'apiTest.noOptionsFound': 'No options',
  'apiTest.stream': 'Stream',
  'apiTest.requestMeta': 'Request meta',
  'apiTest.noKeysTitle': 'No keys',
  'apiTest.noKeysDescription': 'No keys desc',
  'apiTest.manageKeys': 'Manage keys',
  'apiTest.modelPlaceholder': 'Model placeholder',
  'apiTest.modelSearchPlaceholder': 'Search model',
  'apiTest.modelHint': 'Model hint',
  'apiTest.customModel': 'Custom model',
  'apiTest.customModelHint': 'Custom model hint',
  'apiTest.customModelOption': 'Manual model',
  'apiTest.customModelOptionHint': 'Manual model hint',
  'apiTest.prompt': 'Prompt',
  'apiTest.promptHint': 'Prompt hint',
  'apiTest.promptPlaceholder': 'Prompt placeholder',
  'apiTest.streamHint': 'Stream hint',
  'apiTest.unassignedTitle': 'Unassigned',
  'apiTest.unassignedDescription': 'Unassigned desc',
  'apiTest.liveBillingTitle': 'Real request',
  'apiTest.liveBillingDescription': 'Real request desc',
  'apiTest.copyCurl': 'Copy curl',
  'apiTest.platform': 'Platform',
  'apiTest.headerMode': 'Header',
  'apiTest.requestPreview': 'Request preview',
  'apiTest.copyRequest': 'Copy request',
  'apiTest.responsePreview': 'Response preview',
  'apiTest.statusCode': 'Status',
  'apiTest.duration': 'Duration',
  'apiTest.copyResponse': 'Copy response',
  'apiTest.responseSummary': 'Summary',
  'apiTest.usageRecordTitle': 'Usage sync',
  'apiTest.openUsage': 'Open usage',
  'apiTest.rawResponse': 'Raw response',
  'apiTest.responsePending': 'Pending',
  'apiTest.notReady': 'Not ready',
  'apiTest.unknownError': 'Unknown error',
  'keys.failedToLoad': 'Failed to load',
  'gateway.platforms.openai': 'OpenAI',
  'gateway.variants.openaiChat.label': 'OpenAI Chat',
  'gateway.variants.openaiChat.description': 'OpenAI chat compatible endpoint',
  'gateway.variants.openaiResponses.label': 'OpenAI Responses',
  'gateway.variants.openaiResponses.description': 'OpenAI responses compatible endpoint',
}

function buildAPITestShellConfig(
  overrides: Record<string, string> = {},
  gatewayVariants: Record<string, unknown> = {},
  defaults: Record<string, unknown> = {},
): string {
  const labels = Object.fromEntries(
    Object.entries(messages)
      .filter(([key]) => key.startsWith('apiTest.'))
      .map(([key, value]) => [key.replace('apiTest.', ''), value])
  )

  return JSON.stringify({
    zh: {
      labels: {
        ...labels,
        loadKeysFailed: 'Failed to load',
        ...overrides,
      },
      gatewayVariants,
      defaults,
    },
  })
}

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, values?: Record<string, unknown>) => {
        let text = messages[key] ?? key
        if (values) {
          for (const [name, value] of Object.entries(values)) {
            text = text.split(`{${name}}`).join(String(value))
          }
        }
        return text
      },
      locale: { value: 'zh-CN' },
    }),
  }
})

const AppLayoutStub = { template: '<div><slot /></div>' }
const SelectStub = {
  props: ['modelValue', 'options', 'placeholder', 'searchPlaceholder', 'emptyText'],
  template: '<div><span>{{ placeholder }}</span><span>{{ searchPlaceholder }}</span><span>{{ emptyText }}</span><span v-for="option in options" :key="option.value">{{ option.label }}{{ option.description }}</span></div>',
}
const InputStub = {
  props: ['label', 'placeholder', 'hint'],
  template: '<div>{{ label }}{{ placeholder }}{{ hint }}</div>',
}
const TextAreaStub = {
  props: ['modelValue', 'label', 'placeholder', 'hint'],
  template: '<div>{{ label }}{{ placeholder }}{{ hint }}{{ modelValue }}</div>',
}

describe('ApiTestView', () => {
  beforeEach(() => {
    keyList.mockReset()
    usageQuery.mockReset()
    copyToClipboard.mockReset()
    appStoreState.showError.mockReset()
    appStoreState.cachedPublicSettings = null
    vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify({ data: [] }), { status: 200 })))
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('renders stable shell labels from public settings', async () => {
    appStoreState.cachedPublicSettings = {
      api_base_url: 'https://api.example.com',
      api_test_shell_config: buildAPITestShellConfig({
        badge: '配置 Live',
        title: '配置测试页',
        description: '配置描述',
        openGuide: '配置说明',
        send: '配置发送',
        keySelector: '配置密钥',
        protocol: '配置协议',
        model: '配置模型',
        loading: '配置加载中',
        noOptionsFound: '配置无选项',
        requestMeta: '配置请求信息',
        modelSearchPlaceholder: '配置搜索模型',
        modelHint: '配置模型提示',
        prompt: '配置提示词',
        promptPlaceholder: '配置提示词占位',
        promptHint: '配置提示词说明',
        stream: '配置流式',
        streamHint: '配置流式说明',
        liveBillingTitle: '配置真实请求',
        liveBillingDescription: '配置真实请求说明',
        copyCurl: '配置复制 curl',
        platform: '配置平台',
        headerMode: '配置 Header',
        requestPreview: '配置请求预览',
        copyRequest: '配置复制请求',
        responsePreview: '配置响应',
        statusCode: '配置状态',
        duration: '配置耗时',
        copyResponse: '配置复制响应',
        rawResponse: '配置原始响应',
        responsePending: '配置等待响应',
      }),
    }
    keyList.mockResolvedValue({
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

    const wrapper = mount(ApiTestView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
          EmptyState: true,
          Input: InputStub,
          LoadingSpinner: true,
          Select: SelectStub,
          TextArea: TextAreaStub,
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
    expect(text).toContain('配置 Live')
    expect(text).toContain('配置测试页')
    expect(text).toContain('配置描述')
    expect(text).toContain('配置说明')
    expect(text).toContain('配置发送')
    expect(text).toContain('配置密钥')
    expect(text).toContain('配置协议')
    expect(text).toContain('配置模型')
    expect(text).toContain('配置无选项')
    expect(text).toContain('配置请求信息')
    expect(text).toContain('配置搜索模型')
    expect(text).toContain('配置模型提示')
    expect(text).toContain('配置提示词')
    expect(text).toContain('配置提示词占位')
    expect(text).toContain('配置提示词说明')
    expect(text).toContain('配置流式')
    expect(text).toContain('配置流式说明')
    expect(text).toContain('配置真实请求')
    expect(text).toContain('配置真实请求说明')
    expect(text).toContain('配置复制 curl')
    expect(text).toContain('配置平台')
    expect(text).toContain('配置 Header')
    expect(text).toContain('配置请求预览')
    expect(text).toContain('配置复制请求')
    expect(text).toContain('配置响应')
    expect(text).toContain('配置状态')
    expect(text).toContain('配置耗时')
    expect(text).toContain('配置复制响应')
    expect(text).toContain('配置等待响应')
  })

  it('uses configured gateway default models from public API guide settings', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response('', { status: 500 })))
    appStoreState.cachedPublicSettings = {
      api_base_url: 'https://api.example.com',
      api_guide_shell_config: buildAPITestShellConfig({}, {
        openaiChat: {
          defaultModel: 'configured-gpt-model',
          fallbackModels: ['configured-gpt-model', 'configured-gpt-mini'],
        },
      }),
      api_test_shell_config: buildAPITestShellConfig(),
    }
    keyList.mockResolvedValue({
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

    const wrapper = mount(ApiTestView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
          EmptyState: true,
          Input: InputStub,
          LoadingSpinner: true,
          Select: SelectStub,
          TextArea: TextAreaStub,
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
    expect(text).toContain('configured-gpt-model')
    expect(text).toContain('configured-gpt-mini')
    expect(text).not.toContain('gpt-4.1')
  })

  it('uses configured API test default prompt from public settings', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response('', { status: 500 })))
    appStoreState.cachedPublicSettings = {
      api_base_url: 'https://api.example.com',
      api_guide_shell_config: buildAPITestShellConfig({}, {
        openaiChat: {
          defaultModel: 'configured-gpt-model',
          fallbackModels: ['configured-gpt-model'],
        },
      }),
      api_test_shell_config: buildAPITestShellConfig({}, {}, {
        defaultPrompt: 'Configured prompt from public settings',
        maxTokens: 512,
        apiKeyPageSize: 55,
      }),
    }
    keyList.mockResolvedValue({
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

    const wrapper = mount(ApiTestView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
          EmptyState: true,
          Input: InputStub,
          LoadingSpinner: true,
          Select: SelectStub,
          TextArea: TextAreaStub,
          Icon: true,
          RouterLink: {
            props: ['to'],
            template: '<a><slot /></a>',
          },
        },
      },
    })

    await flushPromises()

    expect(keyList).toHaveBeenCalledWith(1, 55)
    const text = wrapper.text()
    expect(text).toContain('Configured prompt from public settings')
    expect(text).toContain('"max_tokens": 512')
    expect(text).not.toContain('请简短介绍一下你当前命中的模型和主要能力。')
  })

  it('uses API test defaults for usage record sync page size', async () => {
    appStoreState.cachedPublicSettings = {
      api_base_url: 'https://api.example.com',
      api_test_shell_config: buildAPITestShellConfig({}, {}, {
        usageSyncPageSize: 6,
      }),
    }
    keyList.mockResolvedValue({
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
    usageQuery.mockResolvedValue({
      items: [
        {
          request_id: 'req-gateway-test',
          api_key_id: 7,
          created_at: new Date().toISOString(),
          actual_cost: 0.01,
          input_tokens: 1,
          output_tokens: 2,
          cache_creation_tokens: 0,
          cache_read_tokens: 0,
        },
      ],
    })

    const wrapper = mount(ApiTestView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
          EmptyState: true,
          Input: InputStub,
          LoadingSpinner: true,
          Select: SelectStub,
          TextArea: TextAreaStub,
          Icon: true,
          RouterLink: {
            props: ['to'],
            template: '<a><slot /></a>',
          },
        },
      },
    })

    await flushPromises()
    await (wrapper.vm as any).$?.setupState.runTest()

    expect(usageQuery).toHaveBeenCalledWith(expect.objectContaining({
      page: 1,
      page_size: 6,
      api_key_id: 7,
    }))
  })

  it('does not keep API test shell i18n fallback keys in the view bootstrap layer', () => {
    expect(apiTestViewSource).not.toContain('apiTestFallbackKeys')
    expect(apiTestViewSource).not.toContain('apiTestLabels.value[key] || key')
    expect(apiTestViewSource).not.toContain('apiTest.title')
    expect(apiTestViewSource).not.toContain('keys.failedToLoad')
    expect(apiTestViewSource).not.toContain("t('common.loading')")
    expect(apiTestViewSource).not.toContain("t('common.noOptionsFound')")
    expect(apiTestViewSource).not.toContain("t('common.unknownError')")
    expect(apiTestViewSource).not.toContain('const apiTestLabelKeys')
    expect(apiTestViewSource).not.toContain('resolveShellLabelOverrides(')
    expect(apiTestViewSource).toContain('resolveAPITestShellLabels')
    expect(apiTestViewSource).toContain('resolveAPITestShellDefaults')
    expect(apiTestViewSource).toContain('renderAPITestShellText')
    expect(apiTestViewSource).toContain('useAuthRouteDefaults')
    expect(apiTestViewSource).toContain(':to="apiTestDefaults.guidePath"')
    expect(apiTestViewSource).toContain(':action-to="authRouteDefaults.apiKeysPath"')
    expect(apiTestViewSource).toContain('path: authRouteDefaults.value.usagePath')
    expect(apiTestViewSource).not.toContain('DEFAULT_GATEWAY_TEST_PROMPT')
    expect(apiTestViewSource).toContain('maxTokens: apiTestDefaults.value.maxTokens')
    expect(apiTestViewSource).toContain('apiTestDefaults.value.apiKeyPageSize')
    expect(apiTestViewSource).toContain('apiTestDefaults.value.usageSyncPageSize')
    expect(apiTestViewSource).not.toContain('keysAPI.list(1, 100)')
    expect(apiTestViewSource).not.toContain('page_size: 10')
    expect(apiTestViewSource).not.toContain('action-to="/keys"')
    expect(apiTestViewSource).not.toContain('to="/gateway-guide"')
    expect(apiTestViewSource).not.toContain("path: '/usage'")
  })

  it('does not keep the legacy API test locale section in frontend bundles or route meta', () => {
    const zhLocaleSource = readFileSync('src/i18n/locales/zh.ts', 'utf8')
    const enLocaleSource = readFileSync('src/i18n/locales/en.ts', 'utf8')
    const routerSource = readFileSync('src/router/index.ts', 'utf8')

    for (const source of [zhLocaleSource, enLocaleSource]) {
      expect(source).not.toContain('\n  apiTest: {')
    }
    expect(routerSource).not.toContain("titleKey: 'apiTest.title'")
    expect(routerSource).not.toContain("descriptionKey: 'apiTest.description'")
  })
})
