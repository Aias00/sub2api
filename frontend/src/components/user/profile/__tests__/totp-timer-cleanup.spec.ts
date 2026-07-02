import { readFileSync } from 'node:fs'
import { beforeEach, afterEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import TotpSetupModal from '@/components/user/profile/TotpSetupModal.vue'
import TotpDisableDialog from '@/components/user/profile/TotpDisableDialog.vue'

const mocks = vi.hoisted(() => ({
  showSuccess: vi.fn(),
  showError: vi.fn(),
  writeText: vi.fn(),
  getVerificationMethod: vi.fn(),
  sendVerifyCode: vi.fn(),
  initiateSetup: vi.fn(),
  enable: vi.fn(),
  disable: vi.fn()
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key
  })
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showSuccess: mocks.showSuccess,
    showError: mocks.showError
  })
}))

vi.mock('@/api', () => ({
  totpAPI: {
    getVerificationMethod: mocks.getVerificationMethod,
    sendVerifyCode: mocks.sendVerifyCode,
    initiateSetup: mocks.initiateSetup,
    enable: mocks.enable,
    disable: mocks.disable
  }
}))

const flushPromises = async () => {
  await Promise.resolve()
  await Promise.resolve()
}

const totpSetupModalSource = readFileSync('src/components/user/profile/TotpSetupModal.vue', 'utf8')
const totpDisableDialogSource = readFileSync('src/components/user/profile/TotpDisableDialog.vue', 'utf8')

describe('TOTP 弹窗定时器清理', () => {
  let intervalSeed = 1000
  let setIntervalSpy: ReturnType<typeof vi.spyOn>
  let clearIntervalSpy: ReturnType<typeof vi.spyOn>

  beforeEach(() => {
    intervalSeed = 1000
    mocks.showSuccess.mockReset()
    mocks.showError.mockReset()
    mocks.writeText.mockReset()
    mocks.getVerificationMethod.mockReset()
    mocks.sendVerifyCode.mockReset()
    mocks.initiateSetup.mockReset()
    mocks.enable.mockReset()
    mocks.disable.mockReset()

    mocks.getVerificationMethod.mockResolvedValue({ method: 'email' })
    mocks.sendVerifyCode.mockResolvedValue({ success: true })
    mocks.initiateSetup.mockResolvedValue({
      qr_code_url: 'otpauth://totp/Cloudbase:test?secret=ABC123',
      secret: 'ABC123',
      setup_token: 'setup-token'
    })
    mocks.enable.mockResolvedValue({ success: true })
    mocks.disable.mockResolvedValue({ success: true })
    mocks.writeText.mockResolvedValue(undefined)
    Object.defineProperty(navigator, 'clipboard', {
      value: {
        writeText: mocks.writeText,
      },
      configurable: true,
    })

    setIntervalSpy = vi.spyOn(window, 'setInterval').mockImplementation(((handler: TimerHandler) => {
      void handler
      intervalSeed += 1
      return intervalSeed as unknown as number
    }) as typeof window.setInterval)
    clearIntervalSpy = vi.spyOn(window, 'clearInterval')
  })

  afterEach(() => {
    setIntervalSpy.mockRestore()
    clearIntervalSpy.mockRestore()
  })

  it('TotpSetupModal 卸载时清理倒计时定时器', async () => {
    const wrapper = mount(TotpSetupModal, {
      props: {
        labels: {
          totpSendCode: 'Configured send code',
        },
      },
    })
    await flushPromises()

    const sendButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('Configured send code'))

    expect(sendButton).toBeTruthy()
    await sendButton!.trigger('click')
    await flushPromises()

    expect(setIntervalSpy).toHaveBeenCalledTimes(1)
    const timerId = setIntervalSpy.mock.results[0]?.value

    wrapper.unmount()

    expect(clearIntervalSpy).toHaveBeenCalledWith(timerId)
  })

  it('TotpDisableDialog 卸载时清理倒计时定时器', async () => {
    const wrapper = mount(TotpDisableDialog, {
      props: {
        labels: {
          totpSendCode: 'Configured send code',
        },
      },
    })
    await flushPromises()

    const sendButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('Configured send code'))

    expect(sendButton).toBeTruthy()
    await sendButton!.trigger('click')
    await flushPromises()

    expect(setIntervalSpy).toHaveBeenCalledTimes(1)
    const timerId = setIntervalSpy.mock.results[0]?.value

    wrapper.unmount()

    expect(clearIntervalSpy).toHaveBeenCalledWith(timerId)
  })

  it('TotpSetupModal 失败时改用 toast 并不渲染内联错误', async () => {
    mocks.getVerificationMethod.mockResolvedValue({ method: 'password' })
    mocks.initiateSetup.mockRejectedValue({
      response: { data: { message: 'setup failed' } }
    })

    const wrapper = mount(TotpSetupModal)
    await flushPromises()

    await wrapper.get('input[type="password"]').setValue('correct horse battery staple')
    await wrapper.get('button[type="button"].btn-primary').trigger('click')
    await flushPromises()

    expect(mocks.showError).toHaveBeenCalledWith('setup failed')
    expect(wrapper.text()).not.toContain('setup failed')
    expect(wrapper.find('.bg-red-50').exists()).toBe(false)
  })

  it('TotpSetupModal 使用配置文案并保留验证码发送流程', async () => {
    mocks.getVerificationMethod.mockResolvedValue({ method: 'email' })

    const wrapper = mount(TotpSetupModal, {
      props: {
        labels: {
          currentPassword: 'Configured current password',
          totpSetupTitle: 'Configured setup title',
          totpVerifyEmailFirst: 'Configured verify email first',
          totpEmailCode: 'Configured email code',
          totpEnterEmailCode: 'Configured enter email code',
          totpSendCode: 'Configured send code',
          totpSending: 'Configured sending',
          totpCancel: 'Configured cancel',
          totpNext: 'Configured next',
          totpCodeSent: 'Configured code sent',
        },
      },
    })
    await flushPromises()

    expect(wrapper.text()).toContain('Configured setup title')
    expect(wrapper.text()).toContain('Configured verify email first')
    expect(wrapper.text()).toContain('Configured email code')
    expect((wrapper.get('input[type="text"]').element as HTMLInputElement).placeholder).toBe('Configured enter email code')
    expect(wrapper.findAll('button').some((button) => button.text() === 'Configured cancel')).toBe(true)
    expect(wrapper.findAll('button').some((button) => button.text() === 'Configured next')).toBe(true)

    const sendButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('Configured send code'))
    expect(sendButton).toBeTruthy()

    await sendButton!.trigger('click')
    await flushPromises()

    expect(mocks.sendVerifyCode).toHaveBeenCalled()
    expect(mocks.showSuccess).toHaveBeenCalledWith('Configured code sent')
  })

  it('TotpSetupModal 使用配置文案展示后续步骤和启用成功反馈', async () => {
    mocks.getVerificationMethod.mockResolvedValue({ method: 'password' })

    const wrapper = mount(TotpSetupModal, {
      props: {
        labels: {
          totpSetupStep1: 'Configured scan step',
          totpSetupStep2: 'Configured verify step',
          totpManualEntry: 'Configured manual entry',
          totpEnterCode: 'Configured enter totp',
          totpVerify: 'Configured verify action',
          totpNext: 'Configured next',
          totpBack: 'Configured back',
          totpCopied: 'Configured copied',
          totpEnableSuccess: 'Configured enabled',
        },
      },
    })
    await flushPromises()

    await wrapper.get('input[type="password"]').setValue('correct horse battery staple')
    await wrapper.get('button[type="button"].btn-primary').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Configured scan step')
    expect(wrapper.text()).toContain('Configured manual entry')
    await wrapper.find('button.rounded.p-1\\.5').trigger('click')
    await flushPromises()
    expect(mocks.writeText).toHaveBeenCalledWith('ABC123')
    expect(mocks.showSuccess).toHaveBeenCalledWith('Configured copied')

    await wrapper.findAll('button').find((button) => button.text().includes('Configured next'))!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Configured verify step')
    expect(wrapper.text()).toContain('Configured enter totp')
    expect(wrapper.text()).toContain('Configured verify action')
    expect(wrapper.findAll('button').some((button) => button.text() === 'Configured back')).toBe(true)

    const inputs = wrapper.findAll('input')
    for (const [index, input] of inputs.entries()) {
      await input.setValue(String(index + 1))
    }
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(mocks.enable).toHaveBeenCalledWith({
      totp_code: '123456',
      setup_token: 'setup-token',
    })
    expect(mocks.showSuccess).toHaveBeenCalledWith('Configured enabled')
  })

  it('TotpSetupModal 使用配置的通用错误兜底', async () => {
    mocks.getVerificationMethod.mockRejectedValue(new Error('network failed'))

    mount(TotpSetupModal, {
      props: {
        labels: {
          totpError: 'Configured setup error',
        },
      },
    })
    await flushPromises()

    expect(mocks.showError).toHaveBeenCalledWith('Configured setup error')
  })

  it('TotpDisableDialog 失败时改用 toast 并不渲染内联错误', async () => {
    mocks.getVerificationMethod.mockResolvedValue({ method: 'password' })
    mocks.disable.mockRejectedValue({
      response: { data: { message: 'disable failed' } }
    })

    const wrapper = mount(TotpDisableDialog)
    await flushPromises()

    await wrapper.get('input[type="password"]').setValue('correct horse battery staple')
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(mocks.showError).toHaveBeenCalledWith('disable failed')
    expect(wrapper.text()).not.toContain('disable failed')
    expect(wrapper.find('.bg-red-50').exists()).toBe(false)
  })

  it('TotpDisableDialog 使用配置文案并保留禁用流程', async () => {
    mocks.getVerificationMethod.mockResolvedValue({ method: 'password' })
    let resolveDisable: ((value: { success: boolean }) => void) | undefined
    mocks.disable.mockImplementation(() => new Promise((resolve) => {
      resolveDisable = resolve
    }))

    const wrapper = mount(TotpDisableDialog, {
      props: {
        labels: {
          currentPassword: 'Configured current password',
          totpDisableTitle: 'Configured disable title',
          totpDisableWarning: 'Configured disable warning',
          totpEnterPassword: 'Configured password placeholder',
          totpConfirmDisable: 'Configured confirm disable',
          totpCancel: 'Configured cancel',
          totpProcessing: 'Configured processing',
          totpDisableSuccess: 'Configured disabled',
        },
      },
    })
    await flushPromises()

    expect(wrapper.text()).toContain('Configured disable title')
    expect(wrapper.text()).toContain('Configured disable warning')
    expect(wrapper.text()).toContain('Configured current password')
    expect((wrapper.get('input[type="password"]').element as HTMLInputElement).placeholder).toBe('Configured password placeholder')
    expect(wrapper.get('button[type="submit"]').text()).toBe('Configured confirm disable')
    expect(wrapper.findAll('button').some((button) => button.text() === 'Configured cancel')).toBe(true)

    await wrapper.get('input[type="password"]').setValue('correct horse battery staple')
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(mocks.disable).toHaveBeenCalledWith({ password: 'correct horse battery staple' })
    expect(wrapper.get('button[type="submit"]').text()).toBe('Configured processing')
    resolveDisable?.({ success: true })
    await flushPromises()
    expect(mocks.showSuccess).toHaveBeenCalledWith('Configured disabled')
  })

  it('TotpDisableDialog 使用配置的通用错误兜底', async () => {
    mocks.getVerificationMethod.mockRejectedValue(new Error('network failed'))

    mount(TotpDisableDialog, {
      props: {
        labels: {
          totpError: 'Configured TOTP error',
        },
      },
    })
    await flushPromises()

    expect(mocks.showError).toHaveBeenCalledWith('Configured TOTP error')
  })

  it('TOTP dialogs do not render local label keys as fallback copy', () => {
    expect(totpSetupModalSource).not.toContain('return props.labels?.[key] || key')
    expect(totpDisableDialogSource).not.toContain('return props.labels?.[key] || key')
  })
})
