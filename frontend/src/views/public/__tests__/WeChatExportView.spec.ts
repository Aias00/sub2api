import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import WeChatExportView from '../WeChatExportView.vue'

const authStoreState = vi.hoisted(() => ({
  isAuthenticated: false,
  isAdmin: false,
}))

const appStoreState = vi.hoisted(() => ({
  siteName: 'Sub2API',
  siteLogo: '',
  cachedPublicSettings: {
    site_name: 'cloudbase',
    site_logo: '',
  },
}))

vi.mock('@/stores', () => ({
  useAuthStore: () => authStoreState,
  useAppStore: () => appStoreState,
}))

vi.mock('@/api/wechat-export', () => ({
  listWeChatArticles: vi.fn(async () => ({ items: [], total: 0, page: 1, page_size: 50, pages: 1 })),
  listWeChatExportTasks: vi.fn(async () => ({ items: [], total: 0, page: 1, page_size: 20, pages: 1 })),
  importWeChatArticleLink: vi.fn(),
  createWeChatExportTask: vi.fn(),
  listWeChatExportArtifacts: vi.fn(async () => []),
}))

describe('WeChatExportView', () => {
  beforeEach(() => {
    authStoreState.isAuthenticated = true
    authStoreState.isAdmin = false
  })

  it('renders the WeChat export workspace controls', () => {
    const wrapper = mount(WeChatExportView, {
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

    expect(wrapper.text()).toContain('WeChat export')
    expect(wrapper.text()).toContain('Article intake')
    expect(wrapper.text()).toContain('Create task')
    expect(wrapper.text()).toContain('Tasks and artifacts')
    expect(wrapper.find('input[type="url"]').exists()).toBe(true)
    expect(wrapper.findAll('input[type="checkbox"]').length).toBeGreaterThanOrEqual(4)
  })

  it('prompts anonymous visitors to log in', () => {
    authStoreState.isAuthenticated = false
    const wrapper = mount(WeChatExportView, {
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

    expect(wrapper.text()).toContain('Log in before using WeChat export')
  })
})
