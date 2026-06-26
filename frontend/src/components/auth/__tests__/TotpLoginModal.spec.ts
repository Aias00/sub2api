import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import TotpLoginModal from '@/components/auth/TotpLoginModal.vue'
import type { AuthShellLabels } from '@/utils/authShell'

const { showErrorMock } = vi.hoisted(() => ({
  showErrorMock: vi.fn(),
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key,
  }),
}))

const shellLabels: AuthShellLabels = {
  totpCancel: 'Configured cancel',
  totpLoginHint: 'Configured login hint',
  totpLoginTitle: 'Configured login title',
  totpVerifying: 'Configured verifying',
}

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    showError: (...args: any[]) => showErrorMock(...args),
  }),
}))

describe('TotpLoginModal', () => {
  beforeEach(() => {
    showErrorMock.mockReset()
  })

  it('sends verification errors to toast and does not render inline red text', async () => {
    const wrapper = mount(TotpLoginModal, {
      props: {
        tempToken: 'temp-token',
        userEmailMasked: 'u***@example.com',
        shellLabels,
      },
    })

    ;(wrapper.vm as unknown as { setError: (message: string) => void }).setError('Invalid code')
    await wrapper.vm.$nextTick()

    expect(showErrorMock).toHaveBeenCalledWith('Invalid code')
    expect(wrapper.text()).not.toContain('Invalid code')
    expect(wrapper.find('.bg-red-50').exists()).toBe(false)
  })

  it('renders login shell copy from auth shell settings', async () => {
    const wrapper = mount(TotpLoginModal, {
      props: {
        tempToken: 'temp-token',
        shellLabels,
      },
    })

    expect(wrapper.text()).toContain('Configured login title')
    expect(wrapper.text()).toContain('Configured login hint')
    expect(wrapper.text()).toContain('Configured cancel')

    ;(wrapper.vm as unknown as { setVerifying: (value: boolean) => void }).setVerifying(true)
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('Configured verifying')
  })
})
