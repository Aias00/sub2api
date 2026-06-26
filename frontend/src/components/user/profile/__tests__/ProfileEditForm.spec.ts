import { readFileSync } from 'node:fs'
import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ProfileEditForm from '@/components/user/profile/ProfileEditForm.vue'
import type { User } from '@/types'

const {
  updateProfileMock,
  showSuccessMock,
  showErrorMock,
  authStoreState,
} = vi.hoisted(() => ({
  updateProfileMock: vi.fn(),
  showSuccessMock: vi.fn(),
  showErrorMock: vi.fn(),
  authStoreState: {
    user: null as User | null,
  },
}))

vi.mock('@/api', () => ({
  userAPI: {
    updateProfile: updateProfileMock,
  },
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => authStoreState,
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

function createUser(overrides: Partial<User> = {}): User {
  return {
    id: 5,
    username: 'alice',
    email: 'alice@example.com',
    avatar_url: null,
    role: 'user',
    balance: 10,
    concurrency: 2,
    status: 'active',
    allowed_groups: null,
    balance_notify_enabled: true,
    balance_notify_threshold: null,
    balance_notify_extra_emails: [],
    created_at: '2026-04-20T00:00:00Z',
    updated_at: '2026-04-20T00:00:00Z',
    ...overrides,
  }
}

const configuredLabels = {
  profileEditTitle: 'Configured edit title',
  profileUsername: 'Configured username',
  profileUsernamePlaceholder: 'Configured username placeholder',
  profileUpdating: 'Configured updating',
  profileUpdateAction: 'Configured update action',
  profileUsernameRequired: 'Configured username required',
  profileUpdateSuccess: 'Configured update success',
  profileUpdateFailed: 'Configured update failed',
}

const profileEditFormSource = readFileSync('src/components/user/profile/ProfileEditForm.vue', 'utf8')

describe('ProfileEditForm', () => {
  beforeEach(() => {
    updateProfileMock.mockReset()
    showSuccessMock.mockReset()
    showErrorMock.mockReset()
    authStoreState.user = null
  })

  it('renders configured profile edit labels', () => {
    const wrapper = mount(ProfileEditForm, {
      props: {
        initialUsername: 'alice',
        labels: configuredLabels,
      },
    })

    expect(wrapper.text()).toContain('Configured edit title')
    expect(wrapper.text()).toContain('Configured username')
    expect((wrapper.get('#username').element as HTMLInputElement).placeholder).toBe('Configured username placeholder')
    expect(wrapper.get('button[type="submit"]').text()).toBe('Configured update action')
  })

  it('uses configured validation and success feedback', async () => {
    const updatedUser = createUser({ username: 'bob' })
    updateProfileMock.mockResolvedValue(updatedUser)

    const wrapper = mount(ProfileEditForm, {
      props: {
        initialUsername: '',
        labels: configuredLabels,
      },
    })

    await wrapper.get('form').trigger('submit.prevent')
    expect(showErrorMock).toHaveBeenCalledWith('Configured username required')

    await wrapper.get('#username').setValue('bob')
    await wrapper.get('form').trigger('submit.prevent')

    expect(updateProfileMock).toHaveBeenCalledWith({ username: 'bob' })
    expect(authStoreState.user?.username).toBe('bob')
    expect(showSuccessMock).toHaveBeenCalledWith('Configured update success')
  })

  it('uses configured failure fallback when update has no server detail', async () => {
    updateProfileMock.mockRejectedValue({})

    const wrapper = mount(ProfileEditForm, {
      props: {
        initialUsername: 'alice',
        labels: configuredLabels,
      },
    })

    await wrapper.get('#username').setValue('bob')
    await wrapper.get('form').trigger('submit.prevent')

    expect(showErrorMock).toHaveBeenCalledWith('Configured update failed')
  })

  it('does not render local label keys as profile edit fallback copy', () => {
    expect(profileEditFormSource).not.toContain('return props.labels?.[key] || key')
  })
})
