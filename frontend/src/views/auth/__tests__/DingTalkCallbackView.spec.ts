import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import DingTalkCallbackView from '../DingTalkCallbackView.vue'

const replace = vi.fn()
const showSuccess = vi.fn()
const showError = vi.fn()
const setToken = vi.fn()
const setPendingAuthSession = vi.fn()
const clearPendingAuthSession = vi.fn()
const exchangePendingOAuthCompletion = vi.fn()
const getPublicSettings = vi.fn()
const apiClientPost = vi.fn()
const login2FA = vi.fn()

vi.mock('vue-router', () => ({
  useRoute: () => ({
    query: {}
  }),
  useRouter: () => ({
    replace
  })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
      te: () => false
    })
  }
})

vi.mock('@/stores', () => ({
  useAuthStore: () => ({
    setToken,
    setPendingAuthSession,
    clearPendingAuthSession
  }),
  useAppStore: () => ({
    showSuccess,
    showError
  })
}))

vi.mock('@/api/client', () => ({
  apiClient: {
    post: (...args: any[]) => apiClientPost(...args)
  }
}))

vi.mock('@/api/auth', async () => {
  const actual = await vi.importActual<typeof import('@/api/auth')>('@/api/auth')
  return {
    ...actual,
    exchangePendingOAuthCompletion: (...args: any[]) => exchangePendingOAuthCompletion(...args),
    getPublicSettings: (...args: any[]) => getPublicSettings(...args),
    login2FA: (...args: any[]) => login2FA(...args)
  }
})

describe('DingTalkCallbackView', () => {
  beforeEach(() => {
    replace.mockReset()
    showSuccess.mockReset()
    showError.mockReset()
    setToken.mockReset()
    setPendingAuthSession.mockReset()
    clearPendingAuthSession.mockReset()
    exchangePendingOAuthCompletion.mockReset()
    getPublicSettings.mockReset()
    apiClientPost.mockReset()
    login2FA.mockReset()
    window.location.hash = ''
    localStorage.clear()
    sessionStorage.clear()
    getPublicSettings.mockResolvedValue({
      auth_shell_config: JSON.stringify({
        en: {
          defaults: {
            defaultRedirectPath: '/configured-dashboard',
            bindRedirectPath: '/configured-profile'
          }
        }
      })
    })
  })

  it('uses auth shell redirect defaults for current pending-session callbacks without redirect params', async () => {
    window.location.hash = '#access_token=legacy-access-token&token_type=Bearer'
    exchangePendingOAuthCompletion.mockResolvedValue({
      access_token: 'current-access-token',
      token_type: 'Bearer'
    })
    setToken.mockResolvedValue({})

    mount(DingTalkCallbackView, {
      global: {
        stubs: {
          AuthLayout: { template: '<div><slot /></div>' },
          Icon: true,
          RouterLink: { template: '<a><slot /></a>' },
          transition: false
        }
      }
    })

    await flushPromises()

    expect(exchangePendingOAuthCompletion).toHaveBeenCalledTimes(1)
    expect(setToken).toHaveBeenCalledWith('current-access-token')
    expect(replace).toHaveBeenCalledWith('/configured-dashboard')
  })

  it('uses auth shell bind redirect defaults when bind completion has no redirect', async () => {
    exchangePendingOAuthCompletion.mockResolvedValue({})

    mount(DingTalkCallbackView, {
      global: {
        stubs: {
          AuthLayout: { template: '<div><slot /></div>' },
          Icon: true,
          RouterLink: { template: '<a><slot /></a>' },
          transition: false
        }
      }
    })

    await flushPromises()

    expect(setToken).not.toHaveBeenCalled()
    expect(clearPendingAuthSession).toHaveBeenCalledTimes(1)
    expect(showSuccess).toHaveBeenCalledWith('profile.authBindings.bindSuccess')
    expect(replace).toHaveBeenCalledWith('/configured-profile')
  })
})
