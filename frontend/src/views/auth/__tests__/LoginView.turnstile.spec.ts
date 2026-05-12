import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import LoginView from '../LoginView.vue'

const loginMock = vi.fn()
const showSuccessMock = vi.fn()
const showErrorMock = vi.fn()
const showWarningMock = vi.fn()
const pushMock = vi.fn()
const getPublicSettingsMock = vi.fn()

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: pushMock,
    currentRoute: {
      value: {
        query: {},
      },
    },
  }),
  RouterLink: {
    name: 'RouterLink',
    props: ['to'],
    template: '<a :href="to"><slot /></a>',
  },
}))

vi.mock('@/stores', () => ({
  useAuthStore: () => ({
    login: loginMock,
  }),
  useAppStore: () => ({
    showSuccess: showSuccessMock,
    showError: showErrorMock,
    showWarning: showWarningMock,
  }),
}))

vi.mock('@/api/auth', () => ({
  getPublicSettings: (...args: any[]) => getPublicSettingsMock(...args),
  isTotp2FARequired: (response: any) => response?.requires_2fa === true,
  isWeChatWebOAuthEnabled: () => false,
}))

describe('LoginView turnstile optimization', () => {
  beforeEach(() => {
    loginMock.mockReset()
    showSuccessMock.mockReset()
    showErrorMock.mockReset()
    showWarningMock.mockReset()
    pushMock.mockReset()
    sessionStorage.clear()
    getPublicSettingsMock.mockResolvedValue({
      turnstile_enabled: true,
      turnstile_site_key: 'site-key',
      linuxdo_oauth_enabled: false,
      wechat_oauth_enabled: false,
      backend_mode_enabled: false,
      oidc_oauth_enabled: false,
      oidc_oauth_provider_name: 'OIDC',
      github_oauth_enabled: false,
      google_oauth_enabled: false,
      password_reset_enabled: true,
      login_agreement_enabled: false,
      login_agreement_mode: 'modal',
      login_agreement_updated_at: '',
      login_agreement_revision: '',
      login_agreement_documents: [],
    })
  })

  it('does not render turnstile or send a turnstile token on password login', async () => {
    loginMock.mockResolvedValue({
      access_token: 'token',
      token_type: 'Bearer',
      user: {
        id: 1,
        email: 'user@example.com',
        username: 'user',
        role: 'user',
        balance: 0,
        concurrency: 5,
        status: 'active',
        allowed_groups: null,
        created_at: '',
        updated_at: '',
      },
    })

    const wrapper = mount(LoginView, {
      global: {
        stubs: {
          AuthLayout: { template: '<div><slot /><slot name="footer" /></div>' },
          LoginAgreementPrompt: true,
          TotpLoginModal: true,
          Icon: true,
          EmailOAuthButtons: true,
          LinuxDoOAuthSection: true,
          WechatOAuthSection: true,
          OidcOAuthSection: true,
          TurnstileWidget: { template: '<div data-testid="turnstile-widget" />' },
        },
      },
    })

    await flushPromises()

    expect(wrapper.find('[data-testid="turnstile-widget"]').exists()).toBe(false)

    await wrapper.get('#email').setValue('user@example.com')
    await wrapper.get('#password').setValue('Aizazadi2024!')
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(loginMock).toHaveBeenCalledWith({
      email: 'user@example.com',
      password: 'Aizazadi2024!',
    })
    expect(pushMock).toHaveBeenCalledWith('/dashboard')
  })
})
