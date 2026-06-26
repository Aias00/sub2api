import { readFileSync } from 'node:fs'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ProfileTotpCard from '@/components/user/profile/ProfileTotpCard.vue'

const { getStatusMock } = vi.hoisted(() => ({
  getStatusMock: vi.fn(),
}))

vi.mock('@/api', () => ({
  totpAPI: {
    getStatus: getStatusMock,
  },
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

const profileTotpCardSource = readFileSync('src/components/user/profile/ProfileTotpCard.vue', 'utf8')

describe('ProfileTotpCard', () => {
  beforeEach(() => {
    getStatusMock.mockReset()
  })

  it('renders configured labels for the not-enabled state', async () => {
    getStatusMock.mockResolvedValue({
      feature_enabled: true,
      enabled: false,
    })

    const wrapper = mount(ProfileTotpCard, {
      props: {
        labels: {
          totpTitle: 'Configured TOTP title',
          totpDescription: 'Configured TOTP description',
          totpNotEnabled: 'Configured not enabled',
          totpNotEnabledHint: 'Configured not enabled hint',
          totpEnable: 'Configured enable action',
        },
      },
      global: {
        stubs: {
          TotpSetupModal: true,
          TotpDisableDialog: true,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('Configured TOTP title')
    expect(wrapper.text()).toContain('Configured TOTP description')
    expect(wrapper.text()).toContain('Configured not enabled')
    expect(wrapper.text()).toContain('Configured not enabled hint')
    expect(wrapper.text()).toContain('Configured enable action')
  })

  it('renders configured labels for disabled and enabled states', async () => {
    getStatusMock.mockResolvedValueOnce({
      feature_enabled: false,
      enabled: false,
    })

    const disabledWrapper = mount(ProfileTotpCard, {
      props: {
        labels: {
          totpFeatureDisabled: 'Configured feature disabled',
          totpFeatureDisabledHint: 'Configured feature disabled hint',
        },
      },
    })
    await flushPromises()

    expect(disabledWrapper.text()).toContain('Configured feature disabled')
    expect(disabledWrapper.text()).toContain('Configured feature disabled hint')

    getStatusMock.mockResolvedValueOnce({
      feature_enabled: true,
      enabled: true,
      enabled_at: 1766102400,
    })

    const enabledWrapper = mount(ProfileTotpCard, {
      props: {
        labels: {
          totpEnabled: 'Configured enabled',
          totpEnabledAt: 'Configured enabled at',
          totpDisable: 'Configured disable action',
        },
      },
    })
    await flushPromises()

    expect(enabledWrapper.text()).toContain('Configured enabled')
    expect(enabledWrapper.text()).toContain('Configured enabled at')
    expect(enabledWrapper.text()).toContain('Configured disable action')
  })

  it('passes configured labels into setup and disable dialogs', async () => {
    getStatusMock.mockResolvedValueOnce({
      feature_enabled: true,
      enabled: false,
    })

    const setupWrapper = mount(ProfileTotpCard, {
      props: {
        labels: {
          totpEnable: 'Configured enable action',
          totpSetupTitle: 'Configured setup title',
        },
      },
      global: {
        stubs: {
          TotpSetupModal: {
            props: ['labels'],
            template: '<div data-testid="totp-setup-modal">{{ labels.totpSetupTitle }}</div>',
          },
          TotpDisableDialog: true,
        },
      },
    })
    await flushPromises()

    await setupWrapper.findAll('button').find((button) => button.text().includes('Configured enable action'))!.trigger('click')

    expect(setupWrapper.get('[data-testid="totp-setup-modal"]').text()).toBe('Configured setup title')

    getStatusMock.mockResolvedValueOnce({
      feature_enabled: true,
      enabled: true,
    })

    const disableWrapper = mount(ProfileTotpCard, {
      props: {
        labels: {
          totpDisable: 'Configured disable action',
          totpDisableTitle: 'Configured disable title',
        },
      },
      global: {
        stubs: {
          TotpSetupModal: true,
          TotpDisableDialog: {
            props: ['labels'],
            template: '<div data-testid="totp-disable-dialog">{{ labels.totpDisableTitle }}</div>',
          },
        },
      },
    })
    await flushPromises()

    await disableWrapper.findAll('button').find((button) => button.text().includes('Configured disable action'))!.trigger('click')

    expect(disableWrapper.get('[data-testid="totp-disable-dialog"]').text()).toBe('Configured disable title')
  })

  it('does not render local label keys as TOTP card fallback copy', () => {
    expect(profileTotpCardSource).not.toContain('return props.labels?.[key] || key')
  })
})
