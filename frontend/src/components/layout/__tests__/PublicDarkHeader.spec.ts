import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import PublicDarkHeader from '../PublicDarkHeader.vue'

const authStoreState = vi.hoisted(() => ({
  isAuthenticated: true,
  isAdmin: false,
  user: {
    username: 'Ada',
    email: 'ada@example.com',
    avatar_url: 'https://static.example.com/avatar.png',
  },
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

vi.mock('@/composables/useAuthRouteDefaults', () => ({
  useAuthRouteDefaults: () => ({
    authRouteDefaults: {
      homePath: '/home',
      loginPath: '/login',
      value: {
        homePath: '/home',
        loginPath: '/login',
      },
    },
    resolveHomePath: () => '/dashboard',
  }),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

describe('PublicDarkHeader', () => {
  beforeEach(() => {
    localStorage.clear()
    document.documentElement.removeAttribute('data-public-theme')
  })

  it('renders shared brand, account action, avatar, and extra actions', () => {
    const wrapper = mount(PublicDarkHeader, {
      props: {
        accountLabel: '去控制台',
      },
      slots: {
        actions: '<a href="/prompts">提示词</a>',
      },
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

    expect(wrapper.find('a[href="/home"]').text()).toContain('cloudbase')
    expect(wrapper.find('a[href="/dashboard"]').text()).toContain('去控制台')
    expect(wrapper.find('img[alt="Ada"]').attributes('src')).toBe('https://static.example.com/avatar.png')
    expect(wrapper.find('a[href="/prompts"]').exists()).toBe(true)
    expect(wrapper.find('button.public-dark-header__theme-toggle').exists()).toBe(true)
  })

  it('toggles the public theme template from the shared header', async () => {
    localStorage.setItem('public-theme', 'dark')
    const wrapper = mount(PublicDarkHeader, {
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

    const toggle = wrapper.find('button.public-dark-header__theme-toggle')
    expect(toggle.text()).toContain('Light')
    expect(document.documentElement.dataset.publicTheme).toBe('dark')

    await toggle.trigger('click')

    expect(document.documentElement.dataset.publicTheme).toBe('light')
    expect(localStorage.getItem('public-theme')).toBe('light')
    expect(toggle.text()).toContain('Dark')
  })
})
