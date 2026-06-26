import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import AuthPopupView from '@/views/auth/AuthPopupView.vue'

const { getPublicSettingsMock, routeState, routerReplaceMock } = vi.hoisted(() => ({
  getPublicSettingsMock: vi.fn(),
  routeState: {
    query: {} as Record<string, unknown>,
  },
  routerReplaceMock: vi.fn(),
}))

vi.mock('vue-router', () => ({
  useRoute: () => routeState,
  useRouter: () => ({
    replace: (...args: any[]) => routerReplaceMock(...args),
  }),
}))

vi.mock('@/api/auth', () => ({
  getPublicSettings: (...args: any[]) => getPublicSettingsMock(...args),
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    locale: 'zh-CN',
  }),
}))

describe('AuthPopupView', () => {
  beforeEach(() => {
    routeState.query = {}
    getPublicSettingsMock.mockReset()
    getPublicSettingsMock.mockResolvedValue({
      auth_shell_config: JSON.stringify({
        zh: {
          defaults: {
            loginPath: '/configured-login',
          },
          labels: {
            oauthCallbackHint: '配置 OAuth 跳转提示',
          },
        },
      }),
    })
    routerReplaceMock.mockReset()
  })

  it('renders popup copy from auth shell settings', async () => {
    const wrapper = mount(AuthPopupView)

    await flushPromises()

    expect(wrapper.text()).toContain('配置 OAuth 跳转提示')
  })

  it('forwards safe callbackUrl as login redirect', async () => {
    routeState.query = {
      callbackUrl: '/settings/credits',
    }

    mount(AuthPopupView)
    await flushPromises()

    expect(routerReplaceMock).toHaveBeenCalledWith({
      path: '/configured-login',
      query: { redirect: '/settings/credits' },
    })
  })

  it('drops external callbackUrl values', async () => {
    routeState.query = {
      callbackUrl: 'https://evil.test/callback',
    }

    mount(AuthPopupView)
    await flushPromises()

    expect(routerReplaceMock).toHaveBeenCalledWith({
      path: '/configured-login',
      query: undefined,
    })
  })
})
