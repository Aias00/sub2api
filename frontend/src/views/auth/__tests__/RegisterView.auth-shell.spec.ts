import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import RegisterView from '../RegisterView.vue'

const registerViewSource = readFileSync(resolve(process.cwd(), 'src/views/auth/RegisterView.vue'), 'utf8')

const getPublicSettingsMock = vi.fn()
const registerMock = vi.fn()
const pushMock = vi.fn()
const showSuccessMock = vi.fn()
const showErrorMock = vi.fn()
const showWarningMock = vi.fn()

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
  useRouter: () => ({ push: pushMock }),
  useRoute: () => ({ query: {} }),
  RouterLink: {
    name: 'RouterLink',
    props: ['to'],
    template: '<a :href="to"><slot /></a>',
  },
}))

vi.mock('@/stores', () => ({
  useAuthStore: () => ({
    register: registerMock,
    login: vi.fn(),
  }),
  useAppStore: () => ({
    showError: showErrorMock,
    showWarning: showWarningMock,
    showSuccess: showSuccessMock,
  }),
}))

vi.mock('@/api/auth', () => ({
  getPublicSettings: (...args: unknown[]) => getPublicSettingsMock(...args),
  validatePromoCode: vi.fn(),
  validateInvitationCode: vi.fn(),
}))

describe('RegisterView auth shell', () => {
  beforeEach(() => {
    getPublicSettingsMock.mockReset()
    registerMock.mockReset()
    pushMock.mockReset()
    showSuccessMock.mockReset()
    showErrorMock.mockReset()
    showWarningMock.mockReset()
  })

  it('renders registration shell labels from public settings', async () => {
    getPublicSettingsMock.mockResolvedValue({
      registration_enabled: true,
      email_verify_enabled: false,
      promo_code_enabled: true,
      invitation_code_enabled: false,
      turnstile_enabled: false,
      turnstile_site_key: '',
      password_min_length: 8,
      site_name: 'Cloudbase',
      wechat_oauth_enabled: false,
      github_oauth_enabled: false,
      google_oauth_enabled: false,
      registration_email_suffix_whitelist: [],
      login_agreement_enabled: false,
      login_agreement_mode: 'modal',
      login_agreement_updated_at: '',
      login_agreement_revision: '',
      login_agreement_documents: [],
      auth_shell_config: JSON.stringify({
        en: {
          defaults: {
            defaultRedirectPath: '/configured-dashboard',
            loginPath: '/configured-login',
          },
          labels: {
            createAccount: 'Configured Create Account',
            signUpToStart: 'Configured start for {siteName}',
            emailPlaceholder: 'configured-register-email',
            createPasswordPlaceholder: 'configured-register-password',
            passwordHint: 'Configured minimum {count}',
            optional: 'Configured optional',
            alreadyHaveAccount: 'Configured existing account',
            signIn: 'Configured Login',
          },
        },
      }),
    })

    const wrapper = mount(RegisterView, {
      global: {
        stubs: {
          AuthLayout: { template: '<div><slot /><slot name="footer" /></div>' },
          RouterLink: { template: '<a><slot /></a>' },
          LoginAgreementPrompt: true,
          Icon: true,
          EmailOAuthButtons: true,
          TurnstileWidget: true,
        },
      },
    })

    await flushPromises()

    const text = wrapper.text()
    expect(text).toContain('Configured Create Account')
    expect(text).toContain('Configured start for Cloudbase')
    expect(text).toContain('Configured minimum 8')
    expect(text).toContain('(Configured optional)')
    expect(text).toContain('Configured existing account')
    expect(text).toContain('Configured Login')
    expect(wrapper.get('#email').attributes('placeholder')).toBe('configured-register-email')
    expect(wrapper.get('#password').attributes('placeholder')).toBe('configured-register-password')
  })

  it('uses auth shell redirect defaults after registration', async () => {
    getPublicSettingsMock.mockResolvedValue({
      registration_enabled: true,
      email_verify_enabled: false,
      promo_code_enabled: false,
      invitation_code_enabled: false,
      turnstile_enabled: false,
      turnstile_site_key: '',
      password_min_length: 8,
      site_name: 'Cloudbase',
      wechat_oauth_enabled: false,
      github_oauth_enabled: false,
      google_oauth_enabled: false,
      registration_email_suffix_whitelist: [],
      login_agreement_enabled: false,
      login_agreement_mode: 'modal',
      login_agreement_updated_at: '',
      login_agreement_revision: '',
      login_agreement_documents: [],
      auth_shell_config: JSON.stringify({
        en: {
          defaults: {
            defaultRedirectPath: '/configured-dashboard',
            emailVerifyPath: '/configured-email-verify',
          },
          labels: {
            createAccount: 'Create',
            emailPlaceholder: 'email',
            createPasswordPlaceholder: 'password',
          },
        },
      }),
    })
    registerMock.mockResolvedValue({})

    const wrapper = mount(RegisterView, {
      global: {
        stubs: {
          AuthLayout: { template: '<div><slot /><slot name="footer" /></div>' },
          RouterLink: { template: '<a><slot /></a>' },
          LoginAgreementPrompt: true,
          Icon: true,
          EmailOAuthButtons: true,
          TurnstileWidget: true,
        },
      },
    })

    await flushPromises()
    await wrapper.get('#email').setValue('user@example.com')
    await wrapper.get('#password').setValue('Aizazadi2024!')
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(registerMock).toHaveBeenCalledWith(expect.objectContaining({
      email: 'user@example.com',
      password: 'Aizazadi2024!',
    }))
    expect(pushMock).toHaveBeenCalledWith('/configured-dashboard')
  })

  it('uses auth shell route defaults for login and email verification navigation', async () => {
    getPublicSettingsMock.mockResolvedValue({
      registration_enabled: true,
      email_verify_enabled: true,
      promo_code_enabled: false,
      invitation_code_enabled: false,
      turnstile_enabled: false,
      turnstile_site_key: '',
      password_min_length: 8,
      site_name: 'Cloudbase',
      wechat_oauth_enabled: false,
      github_oauth_enabled: false,
      google_oauth_enabled: false,
      registration_email_suffix_whitelist: [],
      login_agreement_enabled: false,
      login_agreement_mode: 'modal',
      login_agreement_updated_at: '',
      login_agreement_revision: '',
      login_agreement_documents: [],
      auth_shell_config: JSON.stringify({
        en: {
          defaults: {
            loginPath: '/configured-login',
            emailVerifyPath: '/configured-email-verify',
          },
        },
      }),
    })

    const wrapper = mount(RegisterView, {
      global: {
        stubs: {
          AuthLayout: { template: '<div><slot /><slot name="footer" /></div>' },
          RouterLink: {
            name: 'RouterLink',
            props: ['to'],
            template: '<a :href="to"><slot /></a>',
          },
          LoginAgreementPrompt: true,
          Icon: true,
          EmailOAuthButtons: true,
          TurnstileWidget: true,
        },
      },
    })

    await flushPromises()
    expect(wrapper.find('a[href="/configured-login"]').exists()).toBe(true)

    registerMock.mockResolvedValue({})
    await wrapper.get('#email').setValue('user@example.com')
    await wrapper.get('#password').setValue('Aizazadi2024!')
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(pushMock).toHaveBeenCalledWith('/configured-email-verify')
  })

  it('does not keep unsupported OIDC login state', () => {
    expect(registerViewSource).not.toContain("ref<string>('OIDC')")
    expect(registerViewSource).not.toContain("|| 'OIDC'")
    expect(registerViewSource).not.toContain('oidcOAuthProviderName')
    expect(registerViewSource).not.toContain('OidcOAuthSection')
  })

  it('keeps registration auth shell labels on the shared typed schema', () => {
    expect(registerViewSource).toContain('useAuthShellText')
    expect(registerViewSource).toContain('const { authText, authShellLabels, authRouteDefaults, defaultRedirectPath, applyAuthShellConfig } = useAuthShellText()')
    expect(registerViewSource).toContain('applyAuthShellConfig(settings.auth_shell_config)')
    expect(registerViewSource).toContain('defaultRedirectPath.value')
    expect(registerViewSource).not.toContain("to=\"/login\"")
    expect(registerViewSource).not.toContain("router.push('/email-verify')")
    expect(registerViewSource).not.toContain('ref<Record<string, string>>({})')
    expect(registerViewSource).not.toContain('function authText(key: string')
  })
})
