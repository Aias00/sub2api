import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ProfilePasswordForm from '@/components/user/profile/ProfilePasswordForm.vue'

const { changePasswordMock, showSuccessMock, showErrorMock } = vi.hoisted(() => ({
  changePasswordMock: vi.fn(),
  showSuccessMock: vi.fn(),
  showErrorMock: vi.fn()
}))

vi.mock('@/api', () => ({
  userAPI: {
    changePassword: changePasswordMock
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    cachedPublicSettings: {
      password_min_length: 8,
    },
    showSuccess: showSuccessMock,
    showError: showErrorMock
  })
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => {
        const translations: Record<string, string> = {
          'profile.changePassword': 'Change Password',
          'profile.currentPassword': 'Current Password',
          'profile.newPassword': 'New Password',
          'profile.confirmNewPassword': 'Confirm New Password',
          'profile.passwordHint': 'Password must be at least 8 characters long',
          'profile.changingPassword': 'Changing...',
          'profile.changePasswordButton': 'Change Password',
          'profile.passwordsNotMatch': 'New passwords do not match',
          'profile.passwordTooShort': 'Password must be at least 8 characters long',
          'profile.passwordChangeSuccess': 'Password changed successfully',
          'profile.passwordChangeFailed': 'Failed to change password'
        }
        return translations[key] ?? key
      }
    })
  }
})

describe('ProfilePasswordForm', () => {
  beforeEach(() => {
    changePasswordMock.mockReset()
    showSuccessMock.mockReset()
    showErrorMock.mockReset()
  })

  it('renders configured shell labels and interpolates validation text', async () => {
    const wrapper = mount(ProfilePasswordForm, {
      props: {
        labels: {
          changePassword: 'Configured Password Title',
          currentPassword: 'Configured Current Password',
          newPassword: 'Configured New Password',
          confirmNewPassword: 'Configured Confirm Password',
          passwordHint: 'Configured minimum {count} chars',
          passwordTooShort: 'Configured too short: {count}',
        },
      },
    })

    expect(wrapper.text()).toContain('Configured Password Title')
    expect(wrapper.text()).toContain('Configured Current Password')
    expect(wrapper.text()).toContain('Configured New Password')
    expect(wrapper.text()).toContain('Configured Confirm Password')
    expect(wrapper.text()).toContain('Configured minimum 8 chars')

    await wrapper.get('#old_password').setValue('old-password')
    await wrapper.get('#new_password').setValue('short')
    await wrapper.get('#confirm_password').setValue('short')
    await wrapper.get('form').trigger('submit.prevent')

    expect(changePasswordMock).not.toHaveBeenCalled()
    expect(showErrorMock).toHaveBeenCalledWith('Configured too short: 8')
  })

  it('shows validation failures as toast messages instead of inline errors', async () => {
    const wrapper = mount(ProfilePasswordForm)

    await wrapper.get('#old_password').setValue('old-password')
    await wrapper.get('#new_password').setValue('new-password')
    await wrapper.get('#confirm_password').setValue('different-password')
    await wrapper.get('form').trigger('submit.prevent')

    expect(changePasswordMock).not.toHaveBeenCalled()
    expect(showErrorMock).toHaveBeenCalledWith('passwordsNotMatch')
    expect(wrapper.find('.input-error-text').exists()).toBe(false)
  })

  it('shows API failures as toast messages', async () => {
    changePasswordMock.mockRejectedValue({
      response: { data: { detail: 'backend failure' } }
    })

    const wrapper = mount(ProfilePasswordForm)

    await wrapper.get('#old_password').setValue('old-password')
    await wrapper.get('#new_password').setValue('new-password')
    await wrapper.get('#confirm_password').setValue('new-password')
    await wrapper.get('form').trigger('submit.prevent')

    expect(changePasswordMock).toHaveBeenCalledWith('old-password', 'new-password')
    expect(showErrorMock).toHaveBeenCalledWith('backend failure')
    expect(wrapper.find('.input-error-text').exists()).toBe(false)
  })

  it('supports password setup mode for oauth-only users without asking for the current password', async () => {
    const wrapper = mount(ProfilePasswordForm, {
      props: {
        emailBound: false,
      },
    })

    expect(wrapper.find('#old_password').exists()).toBe(false)

    await wrapper.get('#new_password').setValue('new-password')
    await wrapper.get('#confirm_password').setValue('new-password')
    await wrapper.get('form').trigger('submit.prevent')

    expect(changePasswordMock).toHaveBeenCalledWith('', 'new-password')
  })
})
