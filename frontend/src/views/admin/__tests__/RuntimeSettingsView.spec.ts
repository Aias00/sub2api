import { readFileSync } from 'node:fs'
import { describe, expect, it, vi, beforeEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'

import RuntimeSettingsView from '../RuntimeSettingsView.vue'

const runtimeSettingsViewSource = readFileSync('src/views/admin/RuntimeSettingsView.vue', 'utf8')

const {
  getSettings,
  updateSettings,
  fetchPublicSettings,
  showError,
  showSuccess,
} = vi.hoisted(() => ({
  getSettings: vi.fn(),
  updateSettings: vi.fn(),
  fetchPublicSettings: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))

vi.mock('@/api/admin', () => ({
  default: {
    settings: {
      getSettings,
      updateSettings,
    },
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    fetchPublicSettings,
    showError,
    showSuccess,
  }),
}))

vi.mock('@/utils/apiError', () => ({
  extractApiErrorMessage: (_error: unknown, fallback: string) => fallback,
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
      tm: (key: string) => key,
    }),
  }
})

const IconStub = defineComponent({
  props: {
    name: { type: String, required: true },
  },
  setup(props) {
    return () => h('span', { 'data-icon': props.name })
  },
})

const ToggleStub = defineComponent({
  props: {
    modelValue: { type: Boolean, required: true },
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    return () =>
      h('button', {
        type: 'button',
        'data-toggle': String(props.modelValue),
        onClick: () => emit('update:modelValue', !props.modelValue),
      })
  },
})

// Stub LocaleEnvelopeEditor since CodeMirror won't work in jsdom
const LocaleEnvelopeEditorStub = defineComponent({
  props: {
    modelValue: { type: String, default: '' },
    label: { type: String, default: '' },
    hint: { type: String, default: '' },
    error: { type: String, default: '' },
    height: { type: String, default: '300px' },
    disabled: { type: Boolean, default: false },
    localeKeys: { type: Array, default: () => ['en', 'zh'] },
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    return () =>
      h('div', { 'data-locale-envelope-editor': '', 'data-label': props.label }, [
        h('textarea', {
          value: props.modelValue,
          'data-testid': 'locale-envelope-editor',
          onInput: (e: Event) => emit('update:modelValue', (e.target as HTMLTextAreaElement).value),
        }),
      ])
  },
})

// Stub ImagePromptFilterConfigEditor since it uses child components that may need stubbing
const ImagePromptFilterConfigEditorStub = defineComponent({
  props: {
    modelValue: { type: String, default: '' },
    label: { type: String, default: '' },
    hint: { type: String, default: '' },
    error: { type: String, default: '' },
    disabled: { type: Boolean, default: false },
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    return () =>
      h('div', { 'data-image-prompt-filter-editor': '', 'data-label': props.label }, [
        h('textarea', {
          value: props.modelValue,
          'data-testid': 'image-prompt-filter-editor',
          onInput: (e: Event) => emit('update:modelValue', (e.target as HTMLTextAreaElement).value),
        }),
      ])
  },
})

function mountView() {
  return mount(RuntimeSettingsView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        Icon: IconStub,
        Toggle: ToggleStub,
        LocaleEnvelopeEditor: LocaleEnvelopeEditorStub,
        ImagePromptFilterConfigEditor: ImagePromptFilterConfigEditorStub,
      },
    },
  })
}

describe('RuntimeSettingsView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    getSettings.mockResolvedValue({
      site_name: 'Cloudbase',
      app_name: 'Web',
      app_url: 'https://web.example.com',
      home_business_shell_config: '{"en":{"labels":{"heroTitle":"Business Home"}}}',
      prompt_cases_title: 'Prompt Cases',
      prompt_catalog_shell_config: '{"en":{"labels":{"caseTitle":"Prompt Cases"}}}',
      pricing_title: 'Legacy Pricing',
      pricing_description: 'Legacy pricing copy',
      pricing_shell_config: '{"en":{"labels":{"title":"Pricing"}}}',
      payment_shell_config: '{"en":{"labels":{"createOrder":"Create"}}}',
      pricing_currency_symbol: '$',
      credits_title: 'Legacy Credits',
      credits_description: 'Legacy credits copy',
      credits_purchase_label: 'Legacy Purchase',
      credits_balance_label: 'Legacy balance: {balance}',
      credits_shell_config: '{"en":{"labels":{"title":"Credits"}}}',
      credits_per_balance: '12',
      email_auth_visible: true,
      google_auth_visible: false,
      github_auth_visible: true,
      public_integrations_enabled: false,
      crisp_enabled: false,
      tawk_enabled: true,
    })
    updateSettings.mockImplementation(async (payload) => payload)
    fetchPublicSettings.mockResolvedValue({})
  })

  it('saves only public runtime settings and refreshes public settings cache', async () => {
    const wrapper = mountView()
    await flushPromises()

    const saveButton = wrapper.findAll('button').find((button) => button.text().includes('common.save'))
    expect(saveButton).toBeTruthy()
    await saveButton!.trigger('click')
    await flushPromises()

    expect(updateSettings).toHaveBeenCalledTimes(1)
    const payload = updateSettings.mock.calls[0][0]
    expect(payload).toMatchObject({
      app_name: 'Web',
      app_url: 'https://web.example.com',
      home_business_shell_config: '{"en":{"labels":{"heroTitle":"Business Home"}}}',
      prompt_catalog_shell_config: '{"en":{"labels":{"caseTitle":"Prompt Cases"}}}',
      pricing_shell_config: '{"en":{"labels":{"title":"Pricing"}}}',
      payment_shell_config: '{"en":{"labels":{"createOrder":"Create"}}}',
      pricing_currency_symbol: '$',
      credits_shell_config: '{"en":{"labels":{"title":"Credits"}}}',
      credits_per_balance: '12',
      google_analytics_id: '',
      public_integrations_enabled: false,
      email_auth_visible: true,
      google_auth_visible: false,
      github_auth_visible: true,
      crisp_enabled: false,
      crisp_website_id: '',
      tawk_enabled: true,
    })
    expect(payload).not.toHaveProperty('touch_app_name')
    expect(payload).not.toHaveProperty('touch_email_auth_visible')
    expect(payload).not.toHaveProperty('site_name')
    expect(payload).not.toHaveProperty('payment_enabled')
    expect(payload).not.toHaveProperty('prompt_cases_title')
    expect(payload).not.toHaveProperty('prompt_cases_description')
    expect(payload).not.toHaveProperty('prompt_templates_title')
    expect(payload).not.toHaveProperty('prompt_templates_description')
    expect(payload).not.toHaveProperty('pricing_title')
    expect(payload).not.toHaveProperty('pricing_description')
    expect(payload).not.toHaveProperty('credits_title')
    expect(payload).not.toHaveProperty('credits_description')
    expect(payload).not.toHaveProperty('credits_purchase_label')
    expect(payload).not.toHaveProperty('credits_balance_label')
    expect(fetchPublicSettings).toHaveBeenCalledWith(true)
    expect(showSuccess).toHaveBeenCalledWith('admin.settings.runtime.saveSuccess')
  })

  it('does not keep a local pricing currency default in the runtime settings form', () => {
    expect(runtimeSettingsViewSource).not.toContain("pricing_currency_symbol: '¥'")
  })

  it('does not keep a local credits conversion default in the runtime settings form', () => {
    expect(runtimeSettingsViewSource).not.toContain("credits_per_balance: '10'")
  })

  it('exposes the business home runtime config field in the runtime settings form', () => {
    expect(runtimeSettingsViewSource).toContain('form.home_business_shell_config')
  })

  it('renders shell config fields using LocaleEnvelopeEditor instead of raw textarea', async () => {
    const wrapper = mountView()
    await flushPromises()

    // Shell config fields should be rendered by LocaleEnvelopeEditor stub
    const editors = wrapper.findAll('[data-locale-envelope-editor]')
    expect(editors.length).toBeGreaterThan(0)

    // The image_prompt_filter_config should use ImagePromptFilterConfigEditor
    const filterEditor = wrapper.find('[data-image-prompt-filter-editor]')
    expect(filterEditor.exists()).toBe(true)

    // Raw textareas with shell config placeholders should no longer exist
    expect(wrapper.find('textarea[placeholder*="welcomeBack"]').exists()).toBe(false)
  })

  it('renders a format JSON button', async () => {
    const wrapper = mountView()
    await flushPromises()

    const formatButton = wrapper.findAll('button').find((b) => b.text().includes('formatAllJson'))
    expect(formatButton).toBeTruthy()
  })

  it('does not expose standalone prompt catalog heading fields in the runtime settings form', () => {
    expect(runtimeSettingsViewSource).not.toContain('form.prompt_cases_title')
    expect(runtimeSettingsViewSource).not.toContain('form.prompt_cases_description')
    expect(runtimeSettingsViewSource).not.toContain('form.prompt_templates_title')
    expect(runtimeSettingsViewSource).not.toContain('form.prompt_templates_description')
  })

  it('does not expose standalone pricing copy fields in the runtime settings form', () => {
    expect(runtimeSettingsViewSource).not.toContain('form.pricing_title')
    expect(runtimeSettingsViewSource).not.toContain('form.pricing_description')
  })

  it('does not expose standalone credits copy fields in the runtime settings form', () => {
    expect(runtimeSettingsViewSource).not.toContain('form.credits_title')
    expect(runtimeSettingsViewSource).not.toContain('form.credits_description')
    expect(runtimeSettingsViewSource).not.toContain('form.credits_purchase_label')
    expect(runtimeSettingsViewSource).not.toContain('form.credits_balance_label')
  })
})
