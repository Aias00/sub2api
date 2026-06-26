import { readFileSync } from 'node:fs'
import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ProfileBalanceNotifyCard from '@/components/user/profile/ProfileBalanceNotifyCard.vue'
import { userAPI } from '@/api'

const { showErrorMock, showSuccessMock } = vi.hoisted(() => ({
  showErrorMock: vi.fn(),
  showSuccessMock: vi.fn(),
}))

vi.mock('@/api', () => ({
  userAPI: {
    updateProfile: vi.fn(),
    toggleNotifyEmail: vi.fn(),
    sendNotifyEmailCode: vi.fn(),
    verifyNotifyEmail: vi.fn(),
    removeNotifyEmail: vi.fn(),
    getProfile: vi.fn(),
  },
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    user: null,
  }),
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showSuccess: showSuccessMock,
    showError: showErrorMock,
  }),
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

const profileBalanceNotifyCardSource = readFileSync('src/components/user/profile/ProfileBalanceNotifyCard.vue', 'utf8')

describe('ProfileBalanceNotifyCard', () => {
  beforeEach(() => {
    showErrorMock.mockReset()
    showSuccessMock.mockReset()
    vi.mocked(userAPI.updateProfile).mockReset()
  })

  it('renders configured balance notification labels', () => {
    const wrapper = mount(ProfileBalanceNotifyCard, {
      props: {
        enabled: true,
        threshold: null,
        extraEmails: [],
        systemDefaultThreshold: 5,
        userEmail: 'alice@example.com',
        labels: {
          balanceNotifyTitle: 'Configured balance title',
          balanceNotifyDescription: 'Configured balance description',
          balanceNotifyEnabled: 'Configured enabled label',
          balanceNotifyThreshold: 'Configured threshold',
          balanceNotifyThresholdHint: 'Configured threshold hint',
          balanceNotifySystemDefault: 'Configured default',
          balanceNotifyExtraEmails: 'Configured emails',
          balanceNotifyExtraEmailsHint: 'Configured email hint',
          balanceNotifyEmailPlaceholder: 'Configured email placeholder',
          balanceNotifySaving: 'Configured saving',
          balanceNotifySave: 'Configured save',
          balanceNotifyAdd: 'Configured add',
        },
      },
    })

    expect(wrapper.text()).toContain('Configured balance title')
    expect(wrapper.text()).toContain('Configured balance description')
    expect(wrapper.text()).toContain('Configured enabled label')
    expect(wrapper.text()).toContain('Configured threshold')
    expect(wrapper.text()).toContain('Configured threshold hint')
    expect(wrapper.text()).toContain('Configured emails')
    expect(wrapper.text()).toContain('Configured email hint')
    expect(wrapper.get('input[type="number"]').attributes('placeholder')).toBe('Configured default $5')
    expect(wrapper.get('input[type="email"]').attributes('placeholder')).toBe('Configured email placeholder')
    expect(wrapper.text()).toContain('Configured save')
    expect(wrapper.text()).toContain('Configured add')
  })

  it('uses configured duplicate-email feedback', async () => {
    const wrapper = mount(ProfileBalanceNotifyCard, {
      props: {
        enabled: true,
        threshold: null,
        extraEmails: [
          { email: 'alice@example.com', verified: true, disabled: false },
        ],
        systemDefaultThreshold: 0,
        userEmail: 'alice@example.com',
        labels: {
          balanceNotifyEmailDuplicate: 'Configured duplicate email',
        },
      },
    })

    await wrapper.get('input[type="email"]').setValue('ALICE@example.com')
    await wrapper.get('input[type="email"]').trigger('keyup.enter')

    expect(showErrorMock).toHaveBeenCalledWith('Configured duplicate email')
  })

  it('uses configured save success and error fallback feedback', async () => {
    vi.mocked(userAPI.updateProfile).mockResolvedValue({
      balance_notify_extra_emails: [],
    } as any)

    const wrapper = mount(ProfileBalanceNotifyCard, {
      props: {
        enabled: true,
        threshold: null,
        extraEmails: [],
        systemDefaultThreshold: 0,
        userEmail: 'alice@example.com',
        labels: {
          balanceNotifySave: 'Configured save',
          balanceNotifySaved: 'Configured saved',
          balanceNotifyError: 'Configured error',
        },
      },
    })

    await wrapper.get('button.btn-primary').trigger('click')
    expect(showSuccessMock).toHaveBeenCalledWith('Configured saved')

    vi.mocked(userAPI.updateProfile).mockRejectedValueOnce({})
    await wrapper.get('button.btn-primary').trigger('click')
    expect(showErrorMock).toHaveBeenCalledWith('Configured error')
  })

  it('does not render local label keys as balance notification fallback copy', () => {
    expect(profileBalanceNotifyCardSource).not.toContain('return props.labels?.[key] || key')
  })
})
