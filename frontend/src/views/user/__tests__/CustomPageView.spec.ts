import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { readFileSync } from 'node:fs'
import CustomPageView from '../CustomPageView.vue'

const customPageViewSource = readFileSync('src/views/user/CustomPageView.vue', 'utf8')

const routeState = vi.hoisted(() => ({
  params: { id: 'missing' as string },
}))

const appStoreState = vi.hoisted(() => ({
  publicSettingsLoaded: true,
  cachedPublicSettings: {
    custom_page_shell_config: '',
    custom_menu_items: [] as Array<{ id: string; url: string; page_slug?: string }>,
  },
  fetchPublicSettings: vi.fn(),
}))

const authStoreState = vi.hoisted(() => ({
  isAdmin: false,
  user: { id: 1 },
  token: 'token',
}))

const adminSettingsStoreState = vi.hoisted(() => ({
  customMenuItems: [] as Array<{ id: string; url: string; page_slug?: string }>,
}))

vi.mock('vue-router', () => ({
  useRoute: () => routeState,
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
  useAppStore: () => appStoreState,
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => authStoreState,
}))

vi.mock('@/stores/adminSettings', () => ({
  useAdminSettingsStore: () => adminSettingsStoreState,
}))

vi.mock('@/utils/embedded-url', () => ({
  buildEmbeddedUrl: (url: string) => url,
  detectTheme: () => 'light',
}))

describe('CustomPageView', () => {
  beforeEach(() => {
    routeState.params.id = 'missing'
    appStoreState.publicSettingsLoaded = true
    appStoreState.fetchPublicSettings.mockReset()
    appStoreState.cachedPublicSettings = {
      custom_page_shell_config: JSON.stringify({
        zh: {
          labels: {
            notFoundTitle: '配置不存在标题',
            notFoundDesc: '配置不存在描述',
            notConfiguredTitle: '配置未配置标题',
            notConfiguredDesc: '配置未配置描述',
            openInNewTab: '配置新窗口',
          },
        },
      }),
      custom_menu_items: [],
    }
    authStoreState.isAdmin = false
    adminSettingsStoreState.customMenuItems = []
  })

  it('renders missing custom-page state from Sub2API shell labels', () => {
    const wrapper = mount(CustomPageView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          Icon: { template: '<i />' },
        },
      },
    })

    expect(wrapper.text()).toContain('配置不存在标题')
    expect(wrapper.text()).toContain('配置不存在描述')
  })

  it('renders invalid and iframe custom-page actions from Sub2API shell labels', async () => {
    routeState.params.id = 'bad-url'
    appStoreState.cachedPublicSettings.custom_menu_items = [
      { id: 'bad-url', url: 'ftp://example.com' },
    ]
    const invalidWrapper = mount(CustomPageView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          Icon: { template: '<i />' },
        },
      },
    })

    expect(invalidWrapper.text()).toContain('配置未配置标题')
    expect(invalidWrapper.text()).toContain('配置未配置描述')

    routeState.params.id = 'docs'
    appStoreState.cachedPublicSettings.custom_menu_items = [
      { id: 'docs', url: 'https://example.com/docs' },
    ]
    const validWrapper = mount(CustomPageView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          Icon: { template: '<i />' },
        },
      },
    })

    expect(validWrapper.text()).toContain('配置新窗口')
    expect(validWrapper.find('a').attributes('href')).toBe('https://example.com/docs')
  })

  it('does not keep custom-page shell i18n fallback keys in the view bootstrap layer', () => {
    expect(customPageViewSource).not.toContain('customPageFallbackKeys')
    expect(customPageViewSource).not.toContain('customPageShellLabels.value[key] || key')
    expect(customPageViewSource).not.toContain('customPage.notFoundTitle')
    expect(customPageViewSource).not.toContain('common.copy')
    expect(customPageViewSource).toContain("from './customPageRuntime'")
    expect(customPageViewSource).toContain('resolveCustomPageMenuItem')
    expect(customPageViewSource).toContain('resolveCustomPageMarkdownSlug')
    expect(customPageViewSource).toContain('buildCustomPageImageUrl')
    expect(customPageViewSource).not.toContain('const customPageLabelKeys')
    expect(customPageViewSource).not.toContain('resolveShellLabelOverrides(')
    expect(customPageViewSource).toContain('resolveCustomPageShellLabels')
    expect(customPageViewSource).toContain('renderCustomPageShellText')
  })
})
