import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import LoginView from '../LoginView.vue'

const loginViewSource = readFileSync(resolve(process.cwd(), 'src/views/auth/LoginView.vue'), 'utf8')

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
      locale: { value: 'en' },
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
      auth_shell_config: JSON.stringify({
        en: {
          defaults: {
            defaultRedirectPath: '/configured-dashboard',
            registerPath: '/configured-register',
            forgotPasswordPath: '/configured-forgot-password',
          },
        },
      }),
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
          RouterLink: {
            name: 'RouterLink',
            props: ['to'],
            template: '<a :href="to"><slot /></a>',
          },
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
    expect(pushMock).toHaveBeenCalledWith('/configured-dashboard')
  })

  it('renders login shell labels from public settings', async () => {
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
      login_agreement_enabled: false,
      login_agreement_mode: 'modal',
      login_agreement_updated_at: '',
      login_agreement_revision: '',
      login_agreement_documents: [],
      auth_shell_config: JSON.stringify({
        en: {
          labels: {
            welcomeBack: 'Configured Login Title',
            signInToAccount: 'Configured login subtitle',
            emailPlaceholder: 'configured-email',
            passwordPlaceholder: 'configured-password',
            signIn: 'Configured Sign In',
            dontHaveAccount: 'Configured no account',
            signUp: 'Configured Register',
          },
          defaults: {
            registerPath: '/configured-register',
            forgotPasswordPath: '/configured-forgot-password',
          },
        },
      }),
    })

    const wrapper = mount(LoginView, {
      global: {
        stubs: {
          AuthLayout: { template: '<div><slot /><slot name="footer" /></div>' },
          RouterLink: {
            name: 'RouterLink',
            props: ['to'],
            template: '<a :href="to"><slot /></a>',
          },
          LoginAgreementPrompt: true,
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

    const text = wrapper.text()
    expect(text).toContain('Configured Login Title')
    expect(text).toContain('Configured login subtitle')
    expect(text).toContain('Configured Sign In')
    expect(text).toContain('Configured no account')
    expect(text).toContain('Configured Register')
    expect(wrapper.get('#email').attributes('placeholder')).toBe('configured-email')
    expect(wrapper.get('#password').attributes('placeholder')).toBe('configured-password')
    expect(wrapper.find('a[href="/configured-register"]').exists()).toBe(true)
    expect(wrapper.find('a[href="/configured-forgot-password"]').exists()).toBe(true)
  })

  it('does not keep unsupported OIDC login state', () => {
    expect(loginViewSource).not.toContain("ref<string>('OIDC')")
    expect(loginViewSource).not.toContain("|| 'OIDC'")
    expect(loginViewSource).not.toContain('oidcOAuthProviderName')
    expect(loginViewSource).not.toContain('OidcOAuthSection')
  })

  it('uses auth shell redirect defaults for password login', () => {
    expect(loginViewSource).toContain('useAuthShellText')
    expect(loginViewSource).toContain('applyAuthShellConfig(settings.auth_shell_config)')
    expect(loginViewSource).toContain('defaultRedirectPath.value')
    expect(loginViewSource).toContain('resolveRouteAuthRedirect(router.currentRoute.value.query.redirect, defaultRedirectPath.value)')
    expect(loginViewSource).not.toContain('to="/register"')
    expect(loginViewSource).not.toContain('to="/forgot-password"')
  })

  it('keeps login auth shell labels on the shared typed schema', () => {
    expect(loginViewSource).toContain('useAuthShellText')
    expect(loginViewSource).toContain('const { authText, authShellLabels, authRouteDefaults, defaultRedirectPath, applyAuthShellConfig } = useAuthShellText()')
    expect(loginViewSource).not.toContain('ref<Record<string, string>>({})')
    expect(loginViewSource).not.toContain('function authText(key: string')
  })

  it('does not auto-open the agreement modal when switching emails', async () => {
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
    expect(submit.attributes('disabled')).toBeUndefined()
    expect(agreement.attributes('data-visible')).toBe('false')
  })

  it('opens the agreement modal on login requirement and retries automatically after accept', async () => {
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

    loginMock
      .mockRejectedValueOnce({
        response: {
          data: {
            reason: 'LOGIN_AGREEMENT_REQUIRED',
            metadata: {
              agreement_revision: 'rev-login-2',
            },
          },
        },
      })
      .mockResolvedValueOnce({
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
    await wrapper.get('#email').setValue('user@example.com')
    await wrapper.get('#password').setValue('Aizazadi2024!')
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(loginMock).toHaveBeenCalledTimes(1)
    expect(wrapper.get('[data-testid="agreement"]').attributes('data-visible')).toBe('true')

    await wrapper.findComponent({ name: 'LoginAgreementPrompt' }).vm.$emit('accept')
    await flushPromises()

    expect(loginMock).toHaveBeenCalledTimes(2)
    expect(loginMock).toHaveBeenLastCalledWith({
      email: 'user@example.com',
      password: 'Aizazadi2024!',
      turnstile_token: undefined,
      agreement_accepted: true,
      agreement_revision: 'rev-login-2',
    })
    expect(pushMock).toHaveBeenCalledWith('/dashboard')
  })
})
