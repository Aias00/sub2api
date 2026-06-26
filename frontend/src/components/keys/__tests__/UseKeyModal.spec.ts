import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import type { ApiKeysShellLabels } from '@/utils/apiKeysShell'

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: vi.fn().mockResolvedValue(true)
  })
}))

import UseKeyModal from '../UseKeyModal.vue'

const shellLabels: ApiKeysShellLabels = {
  cancel: 'Configured Close',
  useKeyModalAntigravityClaudeNote: 'Configured Antigravity Claude note',
  useKeyModalAntigravityDescription: 'Configured Antigravity description',
  useKeyModalAntigravityGeminiNote: 'Configured Antigravity Gemini note',
  useKeyModalCliClaudeCode: 'Claude Code',
  useKeyModalCliCodexCli: 'Codex CLI',
  useKeyModalCliCodexCliWs: 'Codex CLI (WebSocket)',
  useKeyModalCliGeminiCli: 'Gemini CLI',
  useKeyModalCliOpencode: 'OpenCode',
  useKeyModalCopied: 'Copied',
  useKeyModalCopy: 'Copy',
  useKeyModalDescription: 'Configured description',
  useKeyModalGeminiDescription: 'Configured Gemini description',
  useKeyModalGeminiModelComment: 'Configured Gemini model comment',
  useKeyModalGeminiNote: 'Configured Gemini note',
  useKeyModalNoGroupDescription: 'Configured no group description',
  useKeyModalNoGroupTitle: 'Configured no group title',
  useKeyModalNote: 'Configured note',
  useKeyModalOpenAIConfigTomlHint: 'Configured config hint',
  useKeyModalOpenAIDescription: 'Configured OpenAI description',
  useKeyModalOpenAINote: 'Configured OpenAI note',
  useKeyModalOpenAINoteWindows: 'Configured OpenAI Windows note',
  useKeyModalOpencodeHint: 'Configured OpenCode hint',
  useKeyModalTitle: 'Configured Use Key',
}

describe('UseKeyModal', () => {
  it('uses configured shell labels for modal chrome and warnings', () => {
    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKey: 'sk-test',
        baseUrl: 'https://example.com/v1',
        platform: null,
        shellLabels,
      },
      global: {
        stubs: {
          BaseDialog: {
            props: ['title'],
            template: '<div><h1>{{ title }}</h1><slot /><slot name="footer" /></div>'
          },
          Icon: {
            template: '<span />'
          }
        }
      }
    })

    expect(wrapper.text()).toContain('Configured Use Key')
    expect(wrapper.text()).toContain('Configured no group title')
    expect(wrapper.text()).toContain('Configured no group description')
    expect(wrapper.text()).toContain('Configured Close')
  })

  it('renders GPT-5.5 and goals feature in OpenAI Codex config', () => {
    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKey: 'sk-test',
        baseUrl: 'https://example.com/v1',
        platform: 'openai',
        shellLabels,
      },
      global: {
        stubs: {
          BaseDialog: {
            template: '<div><slot /><slot name="footer" /></div>'
          },
          Icon: {
            template: '<span />'
          }
        }
      }
    })

    const codeBlocks = wrapper.findAll('pre code').map((code) => code.text())
    const configToml = codeBlocks.find((content) => content.includes('model_provider = "OpenAI"'))

    expect(configToml).toBeDefined()
    expect(configToml).toContain('model = "gpt-5.5"')
    expect(configToml).toContain('review_model = "gpt-5.5"')
    expect(configToml).not.toContain('model = "gpt-5.4"')
    expect(configToml).not.toContain('model_context_window')
    expect(configToml).not.toContain('model_auto_compact_token_limit')
    expect(configToml).toContain('[features]\ngoals = true')
  })

  it('renders GPT-5.5 and goals feature in OpenAI Codex WebSocket config', async () => {
    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKey: 'sk-test',
        baseUrl: 'https://example.com/v1',
        platform: 'openai',
        shellLabels,
      },
      global: {
        stubs: {
          BaseDialog: {
            template: '<div><slot /><slot name="footer" /></div>'
          },
          Icon: {
            template: '<span />'
          }
        }
      }
    })

    const wsTab = wrapper.findAll('button').find((button) =>
      button.text().includes('Codex CLI (WebSocket)')
    )

    expect(wsTab).toBeDefined()
    await wsTab!.trigger('click')
    await nextTick()

    const codeBlocks = wrapper.findAll('pre code').map((code) => code.text())
    const configToml = codeBlocks.find((content) => content.includes('supports_websockets = true'))

    expect(configToml).toBeDefined()
    expect(configToml).toContain('model = "gpt-5.5"')
    expect(configToml).toContain('review_model = "gpt-5.5"')
    expect(configToml).not.toContain('model = "gpt-5.4"')
    expect(configToml).not.toContain('model_context_window')
    expect(configToml).not.toContain('model_auto_compact_token_limit')
    expect(configToml).toContain('[features]\nresponses_websockets_v2 = true\ngoals = true')
  })

  it('renders GPT-5.4 mini entry in OpenCode config', async () => {
    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKey: 'sk-test',
        baseUrl: 'https://example.com/v1',
        platform: 'openai',
        shellLabels,
      },
      global: {
        stubs: {
          BaseDialog: {
            template: '<div><slot /><slot name="footer" /></div>'
          },
          Icon: {
            template: '<span />'
          }
        }
      }
    })

    const opencodeTab = wrapper.findAll('button').find((button) =>
      button.text().includes('OpenCode')
    )

    expect(opencodeTab).toBeDefined()
    await opencodeTab!.trigger('click')
    await nextTick()

    const codeBlock = wrapper.find('pre code')
    expect(codeBlock.exists()).toBe(true)
    expect(codeBlock.text()).toContain('"name": "GPT-5.4 Mini"')
    expect(codeBlock.text()).not.toContain('"name": "GPT-5.4 Nano"')
  })
})
