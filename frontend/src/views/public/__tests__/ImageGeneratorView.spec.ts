import { readFileSync } from 'node:fs'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import ImageGeneratorView from '../ImageGeneratorView.vue'

const imageGeneratorViewSource = readFileSync('src/views/public/ImageGeneratorView.vue', 'utf8')
const fetchPublicSettings = vi.hoisted(() => vi.fn())
const currentLocale = vi.hoisted(() => ({ value: 'en' }))
const configuredWorkspaceShellConfig = vi.hoisted(() => JSON.stringify({
  en: {
    defaults: {
      catalogPath: '/configured-prompts',
      maxPromptLength: 12,
    },
    catalogLabel: 'Configured catalog',
    eyebrow: 'Configured workspace',
    title: 'Configured Image Workbench',
    heroDescription: 'Configured hero copy from public settings.',
    promptLabel: 'Configured prompt',
    promptPlaceholder: 'Configured placeholder',
    clearLabel: 'Configured clear',
    copyPromptLabel: 'Configured copy',
    workspaceTitle: 'Configured status',
    workspaceDescription: 'Configured workspace description.',
    workspaceStatus: 'Configured workspace status.',
    backToCatalogLabel: 'Configured back',
  },
}))

const appStoreState = vi.hoisted(() => ({
  publicSettingsLoaded: true,
  siteName: 'Sub2API',
  siteLogo: '',
  cachedPublicSettings: {
    site_name: 'Sub2API',
    site_logo: '',
    workspace_shell_config: configuredWorkspaceShellConfig,
  },
  fetchPublicSettings,
  showError: vi.fn(),
}))

vi.mock('@/stores', () => ({
  useAppStore: () => appStoreState,
}))

vi.mock('@/i18n', () => ({
  getLocale: () => currentLocale.value,
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: vi.fn(),
  }),
}))

describe('ImageGeneratorView', () => {
  beforeEach(() => {
    currentLocale.value = 'en'
    fetchPublicSettings.mockReset()
    appStoreState.publicSettingsLoaded = true
    appStoreState.cachedPublicSettings = {
      site_name: 'Sub2API',
      site_logo: '',
      workspace_shell_config: configuredWorkspaceShellConfig,
    }
    appStoreState.showError.mockReset()
  })

  it('renders workspace shell copy from public settings', () => {
    const wrapper = mount(ImageGeneratorView, {
      global: {
        stubs: {
          RouterLink: {
            props: ['to'],
            template: '<a :href="typeof to === \'string\' ? to : to.path"><slot /></a>',
          },
          LocaleSwitcher: { template: '<div>locale</div>' },
        },
      },
    })

    expect(wrapper.text()).toContain('Configured catalog')
    expect(wrapper.text()).toContain('Configured workspace')
    expect(wrapper.text()).toContain('Configured Image Workbench')
    expect(wrapper.text()).toContain('Configured hero copy from public settings.')
    expect(wrapper.text()).toContain('Configured prompt')
    expect(wrapper.find('textarea').attributes('placeholder')).toBe('Configured placeholder')
    expect(wrapper.text()).toContain('Configured clear')
    expect(wrapper.text()).toContain('Configured copy')
    expect(wrapper.text()).toContain('Configured status')
    expect(wrapper.text()).toContain('Configured workspace description.')
    expect(wrapper.text()).toContain('Configured workspace status.')
    expect(wrapper.text()).toContain('Configured back')
    expect(wrapper.text()).toContain('0 / 12')
    const catalogLinks = wrapper.findAll('a[href="/configured-prompts"]')
    expect(catalogLinks).toHaveLength(2)
  })

  it('does not embed default workspace shell copy in the Vue view', () => {
    expect(imageGeneratorViewSource).toContain('useAuthRouteDefaults')
    expect(imageGeneratorViewSource).toContain(':to="authRouteDefaults.homePath"')
    expect(imageGeneratorViewSource).toContain('resolveWorkspaceShellDefaults')
    expect(imageGeneratorViewSource).toContain(':to="catalogPath"')
    expect(imageGeneratorViewSource).toContain('workspaceShellDefaults.value.maxPromptLength')
    expect(imageGeneratorViewSource).toContain("from './imageGeneratorRuntime'")
    expect(imageGeneratorViewSource).toContain('applyImageGeneratorDraft')
    expect(imageGeneratorViewSource).toContain('resolveImageGeneratorCatalogPath')
    expect(imageGeneratorViewSource).not.toContain('MAX_PROMPT_LENGTH')
    expect(imageGeneratorViewSource).not.toContain('to="/home"')
    expect(imageGeneratorViewSource).not.toContain('to="/prompts"')
    expect(imageGeneratorViewSource).not.toContain("|| '/prompts'")
    expect(imageGeneratorViewSource).not.toContain('const EMPTY_WORKSPACE_SHELL')
    expect(imageGeneratorViewSource).not.toContain('EMPTY_WORKSPACE_SHELL')
    expect(imageGeneratorViewSource).not.toContain('DEFAULT_WORKSPACE_SHELL')
    expect(imageGeneratorViewSource).not.toContain('function formatTemplate')
    expect(imageGeneratorViewSource).toContain('formatWorkspaceShellTemplate')
    expect(imageGeneratorViewSource).not.toMatch(/resolveWorkspaceShellConfig\([^\n]*,[^\n]*,[^\n]*\)/)
    expect(imageGeneratorViewSource).not.toContain("catalogLabel: '提示词案例'")
    expect(imageGeneratorViewSource).not.toContain("eyebrow: '提示词工作台'")
    expect(imageGeneratorViewSource).not.toContain("title: 'AI 生图工作区'")
    expect(imageGeneratorViewSource).not.toContain("promptPlaceholder: '输入或从案例库导入提示词'")
    expect(imageGeneratorViewSource).not.toContain("copyPromptLabel: '复制提示词'")
    expect(imageGeneratorViewSource).not.toContain("backToCatalogLabel: '返回案例库'")
    expect(imageGeneratorViewSource).not.toContain("catalogLabel: 'Prompt catalog'")
    expect(imageGeneratorViewSource).not.toContain("eyebrow: 'Prompt Workspace'")
    expect(imageGeneratorViewSource).not.toContain("title: 'AI Image Workspace'")
    expect(imageGeneratorViewSource).not.toContain("promptPlaceholder: 'Enter a prompt or import one from the catalog'")
    expect(imageGeneratorViewSource).not.toContain("copyPromptLabel: 'Copy prompt'")
    expect(imageGeneratorViewSource).not.toContain("backToCatalogLabel: 'Back to catalog'")
  })
})
