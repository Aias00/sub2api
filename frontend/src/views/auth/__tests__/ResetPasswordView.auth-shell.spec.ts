import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import ResetPasswordView from '../ResetPasswordView.vue'

const resetPasswordViewSource = readFileSync(resolve(process.cwd(), 'src/views/auth/ResetPasswordView.vue'), 'utf8')
const { getPublicSettingsMock, routeState } = vi.hoisted(() => ({
  getPublicSettingsMock: vi.fn(),
  routeState: {
    query: {
      email: 'user@example.com',
      token: 'reset-token',
    } as Record<string, unknown>,
  },
}))

vi.mock('vue-router', () => ({
  useRoute: () => routeState,
}))

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
  resetPassword: vi.fn(),
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    cachedPublicSettings: { password_min_length: 8 },
    showError: vi.fn(),
    showSuccess: vi.fn(),
  }),
}))

describe('ResetPasswordView auth shell', () => {
  beforeEach(() => {
    routeState.query = {
      email: 'user@example.com',
      token: 'reset-token',
    }
    getPublicSettingsMock.mockReset()
    getPublicSettingsMock.mockResolvedValue({
      auth_shell_config: JSON.stringify({
        en: {
          defaults: {
            loginPath: '/configured-login',
            forgotPasswordPath: '/configured-forgot-password',
          },
          labels: {
            resetPasswordTitle: 'Configured reset password title',
            resetPasswordHint: 'Configured reset password hint',
            emailLabel: 'Configured email',
            newPassword: 'Configured new password',
            newPasswordPlaceholder: 'configured-new-password',
            confirmPassword: 'Configured confirm password',
            confirmPasswordPlaceholder: 'configured-confirm-password',
            resetPassword: 'Configured submit reset',
            rememberedPassword: 'Configured remembered password',
            signIn: 'Configured sign in',
          },
        },
      }),
    })
  })

  it('renders reset-password form shell labels from public settings', async () => {
    const wrapper = mount(ResetPasswordView, {
      global: {
        stubs: {
          AuthLayout: { template: '<div><slot /><slot name="footer" /></div>' },
          Icon: true,
          RouterLink: {
            name: 'RouterLink',
            props: ['to'],
            template: '<a :href="to"><slot /></a>',
          },
        },
      },
    })

    await flushPromises()

    const text = wrapper.text()
    expect(text).toContain('Configured reset password title')
    expect(text).toContain('Configured reset password hint')
    expect(text).toContain('Configured email')
    expect(text).toContain('Configured new password')
    expect(text).toContain('Configured confirm password')
    expect(text).toContain('Configured submit reset')
    expect(text).toContain('Configured remembered password')
    expect(text).toContain('Configured sign in')
    expect(wrapper.get('#password').attributes('placeholder')).toBe('configured-new-password')
    expect(wrapper.get('#confirmPassword').attributes('placeholder')).toBe('configured-confirm-password')
    expect(wrapper.find('a[href="/configured-login"]').exists()).toBe(true)
  })

  it('uses auth shell route defaults for invalid reset links', async () => {
    routeState.query = {}

    const wrapper = mount(ResetPasswordView, {
      global: {
        stubs: {
          AuthLayout: { template: '<div><slot /><slot name="footer" /></div>' },
          Icon: true,
          RouterLink: {
            name: 'RouterLink',
            props: ['to'],
            template: '<a :href="to"><slot /></a>',
          },
        },
      },
    })

    await flushPromises()

    expect(wrapper.find('a[href="/configured-forgot-password"]').exists()).toBe(true)
  })

  it('keeps reset-password navigation on auth shell route defaults', () => {
    expect(resetPasswordViewSource).toContain('useAuthShellText')
    expect(resetPasswordViewSource).toContain('authRouteDefaults.loginPath')
    expect(resetPasswordViewSource).toContain('authRouteDefaults.forgotPasswordPath')
    expect(resetPasswordViewSource).not.toContain('to="/login"')
    expect(resetPasswordViewSource).not.toContain('to="/forgot-password"')
  })
})
