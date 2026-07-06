import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import PublicDarkHeader from '../PublicDarkHeader.vue'

const authStoreState = vi.hoisted(() => ({
  isAuthenticated: true,
  isAdmin: false,
  logout: vi.fn(),
  user: {
    username: 'Ada',
    email: 'ada@example.com',
    avatar_url: 'https://static.example.com/avatar.png',
  },
}))

const routerPush = vi.hoisted(() => vi.fn())

const appStoreState = vi.hoisted(() => ({
  siteName: 'Cloudbase',
  siteLogo: '',
  cachedPublicSettings: {
    site_name: 'cloudbase',
    site_logo: '',
    doc_url: '/docs',
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

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({
      push: routerPush,
    }),
  }
})

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key === 'common.login' ? 'Log in' : key,
    }),
  }
})

describe('PublicDarkHeader', () => {
  beforeEach(() => {
    localStorage.clear()
    document.documentElement.removeAttribute('data-public-theme')
    document.documentElement.classList.remove('dark')
    authStoreState.isAuthenticated = true
    authStoreState.isAdmin = false
    authStoreState.logout.mockReset()
    appStoreState.siteName = 'Cloudbase'
    appStoreState.siteLogo = ''
    appStoreState.cachedPublicSettings = {
      site_name: 'cloudbase',
      site_logo: '',
      doc_url: '/docs',
    }
    routerPush.mockReset()
  })

  it('renders shared brand, avatar menu, and extra actions for authenticated users', async () => {
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
          DocsLink: { props: ['docUrl'], template: '<a href="/docs"><slot /></a>' },
        },
      },
    })

    expect(wrapper.find('nav').classes()).toContain('max-w-7xl')
    expect(wrapper.find('a[href="/home"]').text()).toContain('cloudbase')
    expect(wrapper.find('a[href="/docs"]').text()).toContain('nav.docs')
    expect(wrapper.find('a.public-dark-header__account').exists()).toBe(false)
    expect(wrapper.find('img[alt="Ada"]').attributes('src')).toBe('https://static.example.com/avatar.png')
    expect(wrapper.find('a[href="/prompts"]').exists()).toBe(true)
    expect(wrapper.find('button.public-dark-header__theme-toggle').exists()).toBe(true)

    await wrapper.find('button.public-dark-header__avatar').trigger('click')

    expect(wrapper.find('.public-dark-header__dropdown').exists()).toBe(true)
    expect(wrapper.find('.public-dark-header__dropdown a[href="/dashboard"]').text()).toContain('nav.dashboard')
    expect(wrapper.find('.public-dark-header__dropdown a[href="/app/tasks"]').text()).toContain('nav.myTasks')

    await wrapper.find('.public-dark-header__dropdown-item-danger').trigger('click')

    expect(authStoreState.logout).toHaveBeenCalledTimes(1)
    expect(routerPush).toHaveBeenCalledWith('/login')
  })

  it('keeps the login account action for guests', () => {
    authStoreState.isAuthenticated = false
    const wrapper = mount(PublicDarkHeader, {
      global: {
        stubs: {
          RouterLink: {
            props: ['to'],
            template: '<a :href="typeof to === \'string\' ? to : to.path"><slot /></a>',
          },
          LocaleSwitcher: { template: '<div>locale</div>' },
          DocsLink: { props: ['docUrl'], template: '<a href="/docs"><slot /></a>' },
        },
      },
    })

    expect(wrapper.find('a.public-dark-header__account[href="/login"]').text()).toContain('Log in')
    expect(wrapper.find('button.public-dark-header__avatar').exists()).toBe(false)
  })

  it('allows pages to align the header to their content grid', () => {
    const wrapper = mount(PublicDarkHeader, {
      props: {
        containerClass: 'max-w-6xl',
      },
      global: {
        stubs: {
          RouterLink: {
            props: ['to'],
            template: '<a :href="typeof to === \'string\' ? to : to.path"><slot /></a>',
          },
          LocaleSwitcher: { template: '<div>locale</div>' },
          DocsLink: { props: ['docUrl'], template: '<a href="/docs"><slot /></a>' },
        },
      },
    })

    expect(wrapper.find('nav').classes()).toContain('max-w-6xl')
    expect(wrapper.find('nav').classes()).not.toContain('max-w-7xl')
  })

  it('falls back to cloudbase instead of the legacy project name before settings load', () => {
    appStoreState.siteName = ''
    appStoreState.cachedPublicSettings = null
    const wrapper = mount(PublicDarkHeader, {
      global: {
        stubs: {
          RouterLink: {
            props: ['to'],
            template: '<a :href="typeof to === \'string\' ? to : to.path"><slot /></a>',
          },
          LocaleSwitcher: { template: '<div>locale</div>' },
          DocsLink: { props: ['docUrl'], template: '<a href="/docs"><slot /></a>' },
        },
      },
    })

    expect(wrapper.find('a[href="/home"]').text()).toContain('cloudbase')
    expect(wrapper.find('a[href="/home"]').text()).not.toContain('Cloudbase')
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
          DocsLink: { props: ['docUrl'], template: '<a href="/docs"><slot /></a>' },
        },
      },
    })

    const toggle = wrapper.find('button.public-dark-header__theme-toggle')
    expect(toggle.text()).toBe('')
    expect(toggle.attributes('aria-label')).toBe('nav.lightMode')
    expect(document.documentElement.dataset.publicTheme).toBe('dark')
    expect(document.documentElement.classList.contains('dark')).toBe(true)

    await toggle.trigger('click')

    expect(document.documentElement.dataset.publicTheme).toBe('light')
    expect(document.documentElement.classList.contains('dark')).toBe(false)
    expect(localStorage.getItem('public-theme')).toBe('light')
    expect(toggle.text()).toBe('')
    expect(toggle.attributes('aria-label')).toBe('nav.darkMode')
  })
})
