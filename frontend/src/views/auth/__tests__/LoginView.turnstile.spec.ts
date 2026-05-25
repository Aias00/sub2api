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

describe('LoginView turnstile', () => {
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

  it('renders turnstile and sends the token on password login', async () => {
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
          TurnstileWidget: { template: `<button data-testid="turnstile-widget" @click="$emit('verify', 'turnstile-token')">verify</button>` },
        },
      },
    })

    await flushPromises()

    expect(wrapper.find('[data-testid="turnstile-widget"]').exists()).toBe(true)
    await wrapper.get('[data-testid="turnstile-widget"]').trigger('click')
    await flushPromises()

    await wrapper.get('#email').setValue('user@example.com')
    await wrapper.get('#password').setValue('Aizazadi2024!')
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(loginMock).toHaveBeenCalledWith({
      email: 'user@example.com',
      password: 'Aizazadi2024!',
      turnstile_token: 'turnstile-token',
    })
    expect(pushMock).toHaveBeenCalledWith('/dashboard')
  })

  it('re-opens the agreement gate when switching to a different email in the same browser', async () => {
    getPublicSettingsMock.mockResolvedValue({
      turnstile_enabled: false,
      linuxdo_oauth_enabled: false,
      wechat_oauth_enabled: false,
      backend_mode_enabled: false,
      oidc_oauth_enabled: false,
      oidc_oauth_provider_name: 'OIDC',
      github_oauth_enabled: false,
      google_oauth_enabled: false,
      password_reset_enabled: true,
      login_agreement_enabled: true,
      login_agreement_mode: 'modal',
      login_agreement_updated_at: '2026-05-22',
      login_agreement_revision: 'rev-login-1',
      login_agreement_documents: [
        { id: 'terms', title: '服务条款', content_md: '# 条款' },
      ],
    })
    const wrapper = mount(LoginView, {
      global: {
        stubs: {
          AuthLayout: { template: '<div><slot /><slot name="footer" /></div>' },
          RouterLink: { template: '<a><slot /></a>' },
          LoginAgreementPrompt: {
            name: 'LoginAgreementPrompt',
            props: ['accepted', 'visible'],
            template: '<div data-testid="agreement" :data-accepted="String(accepted)" :data-visible="String(visible)" />',
          },
          TotpLoginModal: true,
          Icon: true,
          EmailOAuthButtons: true,
          LinuxDoOAuthSection: true,
          WechatOAuthSection: true,
          OidcOAuthSection: true,
        },
      },
    })

    await flushPromises()
    const agreement = wrapper.get('[data-testid="agreement"]')
    const submit = wrapper.get('button[type="submit"]')

    await wrapper.get('#email').setValue('first@example.com')
    await flushPromises()
    await wrapper.findComponent({ name: 'LoginAgreementPrompt' }).vm.$emit('accept')
    await flushPromises()
    expect(submit.attributes('disabled')).toBeUndefined()
    expect(agreement.attributes('data-visible')).toBe('false')

    await wrapper.get('#email').setValue('second@example.com')
    await flushPromises()
    expect(submit.attributes('disabled')).toBeDefined()
    expect(agreement.attributes('data-visible')).toBe('true')
  })
})
