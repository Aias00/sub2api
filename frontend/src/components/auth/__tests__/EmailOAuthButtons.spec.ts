import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import EmailOAuthButtons from '@/components/auth/EmailOAuthButtons.vue'
import type { AuthShellLabels } from '@/utils/authShell'

const routeState = vi.hoisted(() => ({
  query: {} as Record<string, unknown>,
}))

const locationState = vi.hoisted(() => ({
  current: { href: 'http://localhost/register?aff=AFF123' } as { href: string },
}))

vi.mock('vue-router', () => ({
  useRoute: () => routeState,
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, string>) => {
        if (key === 'auth.emailOAuth.signIn') {
          return `使用 ${params?.providerName ?? ''} 登录`
        }
        return key
      },
    }),
  }
})

const shellLabels: AuthShellLabels = {
  oauthAlternativeMethods: 'Configured OAuth divider',
  signInWithProvider: 'Configured {providerName} sign in',
}

describe('EmailOAuthButtons', () => {
  beforeEach(() => {
    routeState.query = { redirect: '/billing?plan=pro', aff: 'AFF123' }
    locationState.current = { href: 'http://localhost/register?aff=AFF123' }
    delete window.__APP_CONFIG__
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: locationState.current,
    })
    window.localStorage.clear()
    window.sessionStorage.clear()
  })

  it('passes the affiliate code to the email oauth start URL', async () => {
    const wrapper = mount(EmailOAuthButtons, {
      props: {
        githubEnabled: true,
        googleEnabled: false,
        shellLabels,
      },
      global: {
        stubs: {
          GitHubMark: true,
          GoogleMark: true,
        },
      },
    })

    await wrapper.get('button').trigger('click')

    expect(locationState.current.href).toBe(
      '/api/v1/auth/oauth/github/start?redirect=%2Fbilling%3Fplan%3Dpro&aff_code=AFF123'
    )
    expect(window.sessionStorage.getItem('oauth_aff_code')).toBe('AFF123')
    expect(window.sessionStorage.getItem('email_oauth_pending_provider')).toBe('github')
  })

  it('passes turnstile and agreement revision to the email oauth start URL', async () => {
    window.__APP_CONFIG__ = { api_base_url: 'https://runtime.example.com/api/v1/' } as typeof window.__APP_CONFIG__

    const wrapper = mount(EmailOAuthButtons, {
      props: {
        githubEnabled: true,
        googleEnabled: false,
        shellLabels,
        agreementRevision: 'rev-1',
        turnstileToken: 'ts-token-123',
      },
      global: {
        stubs: {
          GitHubMark: true,
          GoogleMark: true,
        },
      },
    })

    await wrapper.get('button').trigger('click')

    expect(locationState.current.href).toBe(
      'https://runtime.example.com/api/v1/auth/oauth/github/start?redirect=%2Fbilling%3Fplan%3Dpro&aff_code=AFF123&agreement_revision=rev-1&turnstile_token=ts-token-123'
    )
  })

  it('uses a full-width descriptive button when only GitHub is enabled', () => {
    const wrapper = mount(EmailOAuthButtons, {
      props: {
        githubEnabled: true,
        googleEnabled: false,
        shellLabels,
      },
      global: {
        stubs: {
          GitHubMark: true,
          GoogleMark: true,
        },
      },
    })

    expect(wrapper.find('.grid').classes()).not.toContain('sm:grid-cols-2')
    expect(wrapper.get('button').text()).toContain('Configured GitHub sign in')
  })

  it('uses compact labels and two columns when GitHub and Google are both enabled', () => {
    const wrapper = mount(EmailOAuthButtons, {
      props: {
        githubEnabled: true,
        googleEnabled: true,
        shellLabels,
      },
      global: {
        stubs: {
          GitHubMark: true,
          GoogleMark: true,
        },
      },
    })

    expect(wrapper.find('.grid').classes()).toContain('sm:grid-cols-2')
    const buttons = wrapper.findAll('button')
    expect(buttons).toHaveLength(2)
    expect(buttons[0].text()).toContain('GitHub')
    expect(buttons[0].text()).not.toContain('Configured GitHub sign in')
    expect(buttons[1].text()).toContain('Google')
    expect(buttons[1].text()).not.toContain('Configured Google sign in')
  })
})
