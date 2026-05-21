import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent } from 'vue'
import AccountTestModal from '../AccountTestModal.vue'

const { getAvailableModelsMock, copyToClipboardMock } = vi.hoisted(() => ({
  getAvailableModelsMock: vi.fn(),
  copyToClipboardMock: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      getAvailableModels: getAvailableModelsMock
    }
  }
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: copyToClipboardMock
  })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

const BaseDialogStub = defineComponent({
  name: 'BaseDialog',
  props: { show: { type: Boolean, default: false } },
  template: '<div v-if="show"><slot /><slot name="footer" /></div>'
})

const SelectStub = defineComponent({
  name: 'SelectStub',
  props: {
    modelValue: { type: [String, Number, Boolean, null], default: '' },
    options: { type: Array, default: () => [] },
    valueKey: { type: String, default: 'value' },
    labelKey: { type: String, default: 'label' }
  },
  emits: ['update:modelValue'],
  template: `
    <select
      v-bind="$attrs"
      :value="modelValue"
      @change="$emit('update:modelValue', $event.target.value)"
    >
      <option
        v-for="option in options"
        :key="option[valueKey]"
        :value="option[valueKey]"
      >
        {{ option[labelKey] }}
      </option>
    </select>
  `
})

const TextAreaStub = defineComponent({
  name: 'TextArea',
  props: {
    modelValue: { type: String, default: '' }
  },
  emits: ['update:modelValue'],
  template: `
    <textarea
      v-bind="$attrs"
      :value="modelValue"
      @input="$emit('update:modelValue', $event.target.value)"
    />
  `
})

function buildAccount() {
  return {
    id: 1,
    name: 'OpenAI OAuth',
    platform: 'openai',
    type: 'oauth',
    status: 'active',
    credentials: {},
    extra: {},
    concurrency: 1,
    priority: 1,
    proxy_id: null,
    auto_pause_on_expired: false
  } as any
}

describe('AccountTestModal', () => {
  const originalFetch = global.fetch
  const encoder = new TextEncoder()

  function createStreamResponse(lines: string[]) {
    const chunks = lines.map((line) => encoder.encode(line))
    let index = 0

    return {
      ok: true,
      body: {
        getReader: () => ({
          read: vi.fn().mockImplementation(async () => {
            if (index < chunks.length) {
              return { done: false, value: chunks[index++] }
            }
            return { done: true, value: undefined }
          })
        })
      }
    } as Response
  }

  beforeEach(() => {
    getAvailableModelsMock.mockReset()
    copyToClipboardMock.mockReset()
    getAvailableModelsMock.mockResolvedValue([
      { id: 'gpt-5.4', display_name: 'GPT-5.4' }
    ])
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      body: {
        getReader: () => ({
          read: vi.fn().mockResolvedValue({ done: true, value: undefined })
        })
      }
    } as any)
    localStorage.setItem('auth_token', 'test-token')
  })

  afterEach(() => {
    global.fetch = originalFetch
    localStorage.clear()
  })

  it('posts compact mode for OpenAI compact probe', async () => {
    const wrapper = mount(AccountTestModal, {
      props: {
        show: true,
        account: buildAccount()
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Select: SelectStub,
          TextArea: TextAreaStub,
          Icon: true
        }
      }
    })

    await flushPromises()
    ;(wrapper.vm as any).selectedModelId = 'gpt-5.4'
    ;(wrapper.vm as any).testMode = 'compact'
    await (wrapper.vm as any).startTest()
    await flushPromises()

    expect(global.fetch).toHaveBeenCalledTimes(1)
    const [, options] = (global.fetch as any).mock.calls[0]
    expect(JSON.parse(options.body)).toMatchObject({
      model_id: 'gpt-5.4',
      mode: 'compact'
    })
  })

  it('copies only the response body from the output panel', async () => {
    global.fetch = vi.fn().mockResolvedValue(
      createStreamResponse([
        'data: {"type":"test_start","model":"gpt-5.4"}\n',
        'data: {"type":"content","text":"Hey! "}\n',
        'data: {"type":"content","text":"What are you working on?"}\n',
        'data: {"type":"test_complete","success":true}\n'
      ])
    ) as any

    const wrapper = mount(AccountTestModal, {
      props: {
        show: true,
        account: buildAccount()
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Select: SelectStub,
          TextArea: TextAreaStub,
          Icon: true
        }
      }
    })

    await flushPromises()
    ;(wrapper.vm as any).selectedModelId = 'gpt-5.4'
    await (wrapper.vm as any).startTest()
    await flushPromises()
    await flushPromises()

    const copyButton = wrapper.find('button[title="admin.accounts.copyOutput"]')
    expect(copyButton.exists()).toBe(true)

    await copyButton.trigger('click')

    expect(copyToClipboardMock).toHaveBeenCalledWith(
      'Hey! What are you working on?',
      'admin.accounts.outputCopied'
    )
  })

  it('copies the error response when the test fails', async () => {
    global.fetch = vi.fn().mockResolvedValue(
      createStreamResponse([
        'data: {"type":"test_start","model":"gpt-5.4"}\n',
        'data: {"type":"error","error":"API returned 403: {\\\"error\\\":{\\\"message\\\":\\\"tampered\\\"}}"}\n'
      ])
    ) as any

    const wrapper = mount(AccountTestModal, {
      props: {
        show: true,
        account: buildAccount()
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Select: SelectStub,
          TextArea: TextAreaStub,
          Icon: true
        }
      }
    })

    await flushPromises()
    ;(wrapper.vm as any).selectedModelId = 'gpt-5.4'
    await (wrapper.vm as any).startTest()
    await flushPromises()
    await flushPromises()

    const copyButton = wrapper.find('button[title="admin.accounts.copyOutput"]')
    expect(copyButton.exists()).toBe(true)

    await copyButton.trigger('click')

    expect(copyToClipboardMock).toHaveBeenCalledWith(
      'API returned 403: {"error":{"message":"tampered"}}',
      'admin.accounts.outputCopied'
    )
  })

  it('passes claude_code mode and custom test content for Claude compatible accounts', async () => {
    global.fetch = vi.fn().mockResolvedValue(
      createStreamResponse([
        'data: {"type":"test_start","model":"claude-opus-4-7"}\n',
        'data: {"type":"test_complete","success":true}\n'
      ])
    ) as any

    const wrapper = mount(AccountTestModal, {
      props: {
        show: true,
        account: {
          id: 8,
          name: 'Claude Native',
          platform: 'anthropic',
          type: 'apikey',
          status: 'active',
          credentials: {},
          extra: {},
          concurrency: 1,
          priority: 1,
          proxy_id: null,
          auto_pause_on_expired: false
        }
      } as any,
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Select: SelectStub,
          TextArea: TextAreaStub,
          Icon: true
        }
      }
    })

    getAvailableModelsMock.mockResolvedValueOnce([
      { id: 'claude-opus-4-7', display_name: 'Claude Opus 4.7' }
    ])

    await flushPromises()
    const promptInput = wrapper.find('textarea')
    expect(promptInput.exists()).toBe(true)
    await promptInput.setValue('summarize the latest request')

    ;(wrapper.vm as any).selectedModelId = 'claude-opus-4-7'
    ;(wrapper.vm as any).testMode = 'claude_code'

    await (wrapper.vm as any).startTest()
    await flushPromises()
    await flushPromises()

    expect(global.fetch).toHaveBeenCalledTimes(1)
    const [, options] = (global.fetch as any).mock.calls[0]
    expect(JSON.parse(options.body)).toEqual({
      model_id: 'claude-opus-4-7',
      prompt: 'summarize the latest request',
      mode: 'claude_code'
    })
  })

  it('renders upstream request and response debug details', async () => {
    global.fetch = vi.fn().mockResolvedValue(
      createStreamResponse([
        'data: {"type":"request_debug","text":"[Upstream Request]\\nPOST https://example.com/v1/messages\\nHeaders:\\nAuthorization: <redacted>"}\n',
        'data: {"type":"response_debug","text":"[Upstream Response]\\nStatus: 200\\nBody:\\n[streamed response body shown below]"}\n',
        'data: {"type":"test_start","model":"claude-opus-4-7"}\n',
        'data: {"type":"test_complete","success":true}\n'
      ])
    ) as any

    const wrapper = mount(AccountTestModal, {
      props: {
        show: true,
        account: {
          id: 9,
          name: 'Claude Native',
          platform: 'anthropic',
          type: 'apikey',
          status: 'active',
          credentials: {},
          extra: {},
          concurrency: 1,
          priority: 1,
          proxy_id: null,
          auto_pause_on_expired: false
        }
      } as any,
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Select: SelectStub,
          TextArea: TextAreaStub,
          Icon: true
        }
      }
    })

    getAvailableModelsMock.mockResolvedValueOnce([
      { id: 'claude-opus-4-7', display_name: 'Claude Opus 4.7' }
    ])

    await flushPromises()
    ;(wrapper.vm as any).selectedModelId = 'claude-opus-4-7'
    await (wrapper.vm as any).startTest()
    await flushPromises()
    await flushPromises()

    expect(wrapper.text()).toContain('admin.accounts.rawUpstreamRequest')
    expect(wrapper.text()).toContain('POST https://example.com/v1/messages')
    expect(wrapper.text()).toContain('admin.accounts.rawUpstreamResponse')
    expect(wrapper.text()).toContain('[streamed response body shown below]')
  })
})
