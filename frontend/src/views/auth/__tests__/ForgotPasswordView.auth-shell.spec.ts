import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import ForgotPasswordView from '../ForgotPasswordView.vue'

const forgotPasswordViewSource = readFileSync(resolve(process.cwd(), 'src/views/auth/ForgotPasswordView.vue'), 'utf8')
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

vi.mock('@/api/auth', () => ({
  getPublicSettings: (...args: unknown[]) => getPublicSettingsMock(...args),
  forgotPassword: vi.fn(),
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
    showSuccess: vi.fn(),
  }),
}))

describe('ForgotPasswordView auth shell', () => {
  it('renders forgot-password shell labels from public settings', async () => {
    getPublicSettingsMock.mockResolvedValue({
      turnstile_enabled: false,
      turnstile_site_key: '',
      auth_shell_config: JSON.stringify({
        en: {
          defaults: {
            loginPath: '/configured-login',
          },
          labels: {
            forgotPasswordTitle: 'Configured reset title',
            forgotPasswordHint: 'Configured reset hint',
            emailLabel: 'Configured email',
            emailPlaceholder: 'configured-email@example.com',
            sendResetLink: 'Configured send reset',
            rememberedPassword: 'Configured remembered password',
            signIn: 'Configured sign in',
          },
        },
      }),
    })

    const wrapper = mount(ForgotPasswordView, {
      global: {
        stubs: {
          AuthLayout: { template: '<div><slot /><slot name="footer" /></div>' },
          Icon: true,
          RouterLink: {
            name: 'RouterLink',
            props: ['to'],
            template: '<a :href="to"><slot /></a>',
          },
          TurnstileWidget: true,
        },
      },
    })

    await flushPromises()

    const text = wrapper.text()
    expect(text).toContain('Configured reset title')
    expect(text).toContain('Configured reset hint')
    expect(text).toContain('Configured email')
    expect(text).toContain('Configured send reset')
    expect(text).toContain('Configured remembered password')
    expect(text).toContain('Configured sign in')
    expect(wrapper.get('#email').attributes('placeholder')).toBe('configured-email@example.com')
    expect(wrapper.find('a[href="/configured-login"]').exists()).toBe(true)
  })

  it('keeps forgot-password navigation on auth shell route defaults', () => {
    expect(forgotPasswordViewSource).toContain('useAuthShellText')
    expect(forgotPasswordViewSource).toContain('applyAuthShellConfig(settings.auth_shell_config)')
    expect(forgotPasswordViewSource).toContain('authRouteDefaults.loginPath')
    expect(forgotPasswordViewSource).not.toContain('to="/login"')
  })
})
