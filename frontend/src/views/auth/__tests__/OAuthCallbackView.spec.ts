import { flushPromises, mount } from '@vue/test-utils'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import OAuthCallbackView from '@/views/auth/OAuthCallbackView.vue'

const oauthCallbackSource = readFileSync(resolve(process.cwd(), 'src/views/auth/OAuthCallbackView.vue'), 'utf8')

const {
  routeState,
  locationState,
  routerReplaceMock,
  showErrorMock,
  showSuccessMock,
  setTokenMock,
  copyToClipboardMock,
  exchangePendingOAuthCompletionMock,
  getPublicSettingsMock,
  apiPostMock,
} = vi.hoisted(() => ({
  routeState: {
    path: '/auth/callback',
    query: {} as Record<string, unknown>,
  },
  locationState: {
    current: {
      href: 'http://localhost/auth/callback',
      hash: '',
    } as { href: string; hash: string },
  },
  routerReplaceMock: vi.fn(),
  showErrorMock: vi.fn(),
  showSuccessMock: vi.fn(),
  setTokenMock: vi.fn(),
  copyToClipboardMock: vi.fn(),
  exchangePendingOAuthCompletionMock: vi.fn(),
  getPublicSettingsMock: vi.fn(),
  apiPostMock: vi.fn(),
}))

vi.mock('vue-router', () => ({
  useRoute: () => routeState,
  useRouter: () => ({
    replace: (...args: any[]) => routerReplaceMock(...args),
  }),
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key,
    locale: { value: 'en' },
  }),
}))

vi.mock('@/stores', () => ({
  useAuthStore: () => ({
    setToken: (...args: any[]) => setTokenMock(...args),
  }),
  useAppStore: () => ({
    showError: (...args: any[]) => showErrorMock(...args),
    showSuccess: (...args: any[]) => showSuccessMock(...args),
  }),
}))

vi.mock('@/api/client', () => ({
  apiClient: {
    post: (...args: any[]) => apiPostMock(...args),
  },
  buildApiUrl: (path: string, settings?: { api_base_url?: string | null } | null) => {
    const configured = settings?.api_base_url
      || window.__APP_CONFIG__?.api_base_url
      || '/api/v1'
    return `${configured.replace(/\/+$/, '')}${path.startsWith('/') ? path : `/${path}`}`
  },
}))

vi.mock('@/api/auth', async () => {
  const actual = await vi.importActual<typeof import('@/api/auth')>('@/api/auth')
  return {
    ...actual,
    exchangePendingOAuthCompletion: (...args: any[]) => exchangePendingOAuthCompletionMock(...args),
    getPublicSettings: (...args: any[]) => getPublicSettingsMock(...args),
    persistOAuthTokenContext: vi.fn(),
  }
})

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: (...args: any[]) => copyToClipboardMock(...args),
  }),
}))

describe('OAuthCallbackView', () => {
  beforeEach(() => {
    routeState.path = '/auth/callback'
    routeState.query = {}
    locationState.current = {
      href: 'http://localhost/auth/callback',
      hash: '',
    }
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: locationState.current,
    })
    routerReplaceMock.mockReset()
    showErrorMock.mockReset()
    showSuccessMock.mockReset()
    setTokenMock.mockReset()
    copyToClipboardMock.mockReset()
    exchangePendingOAuthCompletionMock.mockReset()
    getPublicSettingsMock.mockReset()
    getPublicSettingsMock.mockResolvedValue({
      auth_shell_config: JSON.stringify({
        en: {
          defaults: {
            defaultRedirectPath: '/configured-dashboard',
            loginPath: '/configured-login',
          },
          labels: {
            backToLogin: 'Configured back to login',
            emailLabel: 'Configured email label',
            passwordLabel: 'Configured password label',
            createPasswordPlaceholder: 'configured-password-placeholder',
            confirmPassword: 'Configured confirm password',
            confirmPasswordPlaceholder: 'configured-confirm-placeholder',
            invitationCodeLabel: 'Configured invitation label',
            invitationCodePlaceholder: 'configured-invitation-placeholder',
            optional: 'Configured optional',
            oauthCallbackCode: 'Configured code',
            oauthCallbackFullUrl: 'Configured full URL',
            oauthCallbackHint: 'Configured callback hint',
            oauthCallbackInvalidHint: 'Configured invalid hint',
            oauthCallbackInvalidTitle: 'Configured invalid title',
            oauthCallbackPasswordOptionalHint: 'Configured optional password for {providerName}',
            oauthCallbackRegistrationHint: 'Configured registration hint',
            oauthCallbackRegistrationInvitationRequired: 'Configured invitation required for {providerName}',
            oauthCallbackState: 'Configured state',
            oauthCallbackSubmitRegistration: 'Configured complete registration',
            oauthCallbackTitle: 'Configured OAuth callback',
            processing: 'Configured processing',
          },
        },
      }),
    })
    apiPostMock.mockReset()
    window.sessionStorage.clear()
    delete window.__APP_CONFIG__
  })

  it('renders localized callback copy actions', async () => {
    routeState.query = {
      code: 'oauth-code',
      state: 'oauth-state',
    }

    const wrapper = mount(OAuthCallbackView)
    await flushPromises()

    expect(wrapper.text()).toContain('Configured OAuth callback')
    expect(wrapper.text()).toContain('Configured callback hint')
    expect(wrapper.text()).toContain('Configured code')
    expect(wrapper.text()).toContain('Configured state')
    expect(wrapper.text()).toContain('Configured full URL')
    expect(wrapper.text()).toContain('common.copy')
    expect(wrapper.find('input[value="oauth-code"]').exists()).toBe(true)
    expect(wrapper.find('input[value="oauth-state"]').exists()).toBe(true)
  })

  it('sends callback errors to toast instead of rendering inline red text', () => {
    routeState.query = {
      error: 'oauth failed',
    }

    const wrapper = mount(OAuthCallbackView)

    expect(showErrorMock).toHaveBeenCalledWith('oauth failed')
    expect(wrapper.text()).not.toContain('oauth failed')
    expect(wrapper.find('.bg-red-50').exists()).toBe(false)
  })

  it('does not render manual copy fields for direct email oauth callback visits', async () => {
    routeState.path = '/auth/oauth/callback'
    exchangePendingOAuthCompletionMock.mockRejectedValue(new Error('pending session not found'))

    const wrapper = mount(OAuthCallbackView)
    await vi.dynamicImportSettled()

    expect(exchangePendingOAuthCompletionMock).toHaveBeenCalledTimes(1)
    expect(wrapper.text()).toContain('Configured invalid title')
    expect(wrapper.text()).toContain('Configured invalid hint')
    expect(wrapper.text()).toContain('Configured back to login')
    expect(wrapper.find('input[readonly]').exists()).toBe(false)
  })

  it('forwards frontend email oauth provider callbacks back to the backend callback endpoint', async () => {
    routeState.path = '/auth/oauth/callback'
    routeState.query = {
      code: 'provider-code',
      state: 'provider-state',
    }
    window.sessionStorage.setItem('email_oauth_pending_provider', 'google')
    window.__APP_CONFIG__ = {
      api_base_url: 'https://api.example.com/api/v1/',
    }

    mount(OAuthCallbackView)
    await vi.dynamicImportSettled()

    expect(locationState.current.href).toBe(
      'https://api.example.com/api/v1/auth/oauth/google/callback?code=provider-code&state=provider-state'
    )
    expect(exchangePendingOAuthCompletionMock).not.toHaveBeenCalled()
  })

  it('uses auth shell redirect defaults for direct token callbacks without redirect param', async () => {
    window.location.hash = '#access_token=token-direct&token_type=Bearer'

    mount(OAuthCallbackView)
    await vi.dynamicImportSettled()

    expect(setTokenMock).toHaveBeenCalledWith('token-direct')
    expect(routerReplaceMock).toHaveBeenCalledWith('/configured-dashboard')
  })

  it('submits stored affiliate code when completing invited email oauth registration', async () => {
    routeState.path = '/auth/oauth/callback'
    exchangePendingOAuthCompletionMock.mockResolvedValue({
      error: 'invitation_required',
      provider: 'google',
      redirect: '/dashboard',
      resolved_email: 'pending@example.com',
      invitation_required: true,
    })
    apiPostMock.mockResolvedValue({
      data: {
        access_token: 'token-1',
      },
    })
    window.sessionStorage.setItem('oauth_aff_code', 'AFF456')

    const wrapper = mount(OAuthCallbackView)
    await vi.dynamicImportSettled()
    const passwordInputs = wrapper.findAll('input[type="password"]')
    await passwordInputs[0].setValue('secret-123')
    await passwordInputs[1].setValue('secret-123')
    const invitationInput = wrapper.find('input[type="text"]')
    await invitationInput.setValue('INVITE456')
    await wrapper.findAll('button').at(0)?.trigger('click')

    expect(apiPostMock).toHaveBeenCalledWith('/auth/oauth/google/complete-registration', {
      password: 'secret-123',
      invitation_code: 'INVITE456',
      aff_code: 'AFF456',
    })
    expect(setTokenMock).toHaveBeenCalledWith('token-1')
  })

  it('completes email oauth registration with readonly email and without posting email', async () => {
    routeState.path = '/auth/oauth/callback'
    exchangePendingOAuthCompletionMock.mockResolvedValue({
      error: 'registration_completion_required',
      provider: 'github',
      redirect: '/dashboard',
      resolved_email: 'verified@example.com',
      invitation_required: false,
    })
    apiPostMock.mockResolvedValue({
      data: {
        access_token: 'token-2',
      },
    })

    const wrapper = mount(OAuthCallbackView)
    await vi.dynamicImportSettled()

    const emailInput = wrapper.find('input[type="email"]')
    expect(emailInput.exists()).toBe(true)
    expect((emailInput.element as HTMLInputElement).value).toBe('verified@example.com')
    expect(emailInput.attributes('readonly')).toBeDefined()
    expect(emailInput.attributes('disabled')).toBeDefined()

    const passwordInputs = wrapper.findAll('input[type="password"]')
    await passwordInputs[0].setValue('secret-456')
    await passwordInputs[1].setValue('secret-456')
    await wrapper.findAll('button').at(0)?.trigger('click')

    expect(apiPostMock).toHaveBeenCalledWith('/auth/oauth/github/complete-registration', {
      password: 'secret-456',
    })
    expect(apiPostMock.mock.calls[0][1]).not.toHaveProperty('email')
    expect(setTokenMock).toHaveBeenCalledWith('token-2')
  })

  it('allows google oauth registration completion without setting a password', async () => {
    routeState.path = '/auth/oauth/callback'
    exchangePendingOAuthCompletionMock.mockResolvedValue({
      error: 'registration_completion_required',
      provider: 'google',
      resolved_email: 'google-only@example.com',
      invitation_required: false,
    })
    apiPostMock.mockResolvedValue({
      data: {
        access_token: 'token-3',
      },
    })

    const wrapper = mount(OAuthCallbackView)
    await vi.dynamicImportSettled()

    const submitButton = wrapper.findAll('button').at(0)
    expect(submitButton?.attributes('disabled')).toBeUndefined()
    await submitButton?.trigger('click')

    expect(apiPostMock).toHaveBeenCalledWith('/auth/oauth/google/complete-registration', {})
    expect(setTokenMock).toHaveBeenCalledWith('token-3')
    expect(routerReplaceMock).toHaveBeenCalledWith('/configured-dashboard')
  })

  it('renders optional password labels from auth shell public settings', async () => {
    routeState.path = '/auth/oauth/callback'
    exchangePendingOAuthCompletionMock.mockResolvedValue({
      error: 'registration_completion_required',
      provider: 'google',
      redirect: '/dashboard',
      resolved_email: 'google-only@example.com',
      invitation_required: false,
    })

    const wrapper = mount(OAuthCallbackView)
    await vi.dynamicImportSettled()

    expect(wrapper.text()).toContain('(Configured optional)')
    expect(wrapper.text()).toContain('Configured optional password for Google')
    expect(getPublicSettingsMock).toHaveBeenCalledTimes(1)
  })

  it('keeps callback shell status labels on auth shell settings', () => {
    expect(oauthCallbackSource).toContain('loadAuthShellConfig')
    expect(oauthCallbackSource).toContain('defaultRedirectPath')
    expect(oauthCallbackSource).toContain('authRouteDefaults.loginPath')
    expect(oauthCallbackSource).toContain('sanitizeAuthRedirectPath(redirect, defaultRedirectPath.value)')
    expect(oauthCallbackSource).toContain("authText('optional')")
    expect(oauthCallbackSource).toContain("authText('processing')")
    expect(oauthCallbackSource).toContain("authText('oauthCallbackTitle')")
    expect(oauthCallbackSource).toContain("authText('oauthCallbackHint')")
    expect(oauthCallbackSource).toContain("authText('oauthCallbackSubmitRegistration')")
    expect(oauthCallbackSource).not.toContain("t('common.optional')")
    expect(oauthCallbackSource).not.toContain("t('common.processing')")
    expect(oauthCallbackSource).not.toContain("t('auth.oauth.callbackTitle')")
    expect(oauthCallbackSource).not.toContain("t('auth.oauth.callbackHint')")
    expect(oauthCallbackSource).not.toContain("t('auth.oidc.completeRegistration')")
    expect(oauthCallbackSource).not.toContain("router.replace('/login')")
  })
})
