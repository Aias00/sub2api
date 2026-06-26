import { flushPromises, mount } from '@vue/test-utils'
import { readFileSync } from 'node:fs'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import WechatPaymentCallbackView from '@/views/auth/WechatPaymentCallbackView.vue'

const { replaceMock, routeState, locationState, showErrorMock, appStoreState } = vi.hoisted(() => ({
  replaceMock: vi.fn(),
  routeState: {
    query: {} as Record<string, unknown>,
  },
  locationState: {
    current: {
      href: 'http://localhost/auth/wechat/payment/callback',
      hash: '',
      search: '',
      pathname: '/auth/wechat/payment/callback',
      origin: 'http://localhost',
    } as Location & { origin: string },
  },
  showErrorMock: vi.fn(),
  appStoreState: {
    cachedPublicSettings: {
      auth_shell_config: JSON.stringify({
        zh: {
          defaults: {
            purchasePath: '/configured-purchase',
          },
        },
      }),
      payment_shell_config: JSON.stringify({
        zh: {
          labels: {
            wechatPaymentCallbackTitle: '配置恢复微信支付',
            wechatPaymentCallbackProcessing: '配置恢复微信支付中...',
            wechatPaymentCallbackBackToPayment: '配置返回支付页',
            wechatPaymentCallbackMissingResumeToken: '配置缺少恢复令牌。',
          },
        },
      }),
    },
  },
}))

vi.mock('vue-router', () => ({
  useRoute: () => routeState,
  useRouter: () => ({
    replace: replaceMock,
  }),
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => {
      if (key === 'auth.wechatPayment.callbackTitle') return '正在恢复微信支付'
      if (key === 'auth.wechatPayment.callbackProcessing') return '正在恢复微信支付...'
      if (key === 'auth.wechatPayment.backToPayment') return '返回支付页'
      if (key === 'auth.wechatPayment.callbackMissingResumeToken') return '微信支付回调缺少恢复令牌。'
      return key
    },
    locale: { value: 'zh-CN' },
  }),
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    ...appStoreState,
    showError: (...args: any[]) => showErrorMock(...args),
  }),
}))

describe('WechatPaymentCallbackView', () => {
  beforeEach(() => {
    replaceMock.mockReset()
    showErrorMock.mockReset()
    appStoreState.cachedPublicSettings = {
      auth_shell_config: JSON.stringify({
        zh: {
          defaults: {
            purchasePath: '/configured-purchase',
          },
        },
      }),
      payment_shell_config: JSON.stringify({
        zh: {
          labels: {
            wechatPaymentCallbackTitle: '配置恢复微信支付',
            wechatPaymentCallbackProcessing: '配置恢复微信支付中...',
            wechatPaymentCallbackBackToPayment: '配置返回支付页',
            wechatPaymentCallbackMissingResumeToken: '配置缺少恢复令牌。',
          },
        },
      }),
    }
    routeState.query = {}
    locationState.current = {
      href: 'http://localhost/auth/wechat/payment/callback',
      hash: '',
      search: '',
      pathname: '/auth/wechat/payment/callback',
      origin: 'http://localhost',
    } as Location & { origin: string }
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: locationState.current,
    })
  })

  it('redirects back to purchase with an opaque resume token from hash fragment', async () => {
    locationState.current.hash = '#wechat_resume_token=resume-token-123&redirect=%2Fpurchase%3Ffrom%3Dwechat'

    mount(WechatPaymentCallbackView)
    await flushPromises()

    expect(replaceMock).toHaveBeenCalledWith({
      path: '/purchase',
      query: {
        from: 'wechat',
        wechat_resume: '1',
        wechat_resume_token: 'resume-token-123',
      },
    })
  })

  it('shows an error when the callback only carries legacy openid payloads without a resume token', async () => {
    locationState.current.hash =
      '#openid=openid-123&state=oauth-state&scope=snsapi_base&payment_type=wxpay_direct&amount=128&order_type=subscription&plan_id=7&redirect=%2Fpayment%3Ffrom%3Dwechat'

    mount(WechatPaymentCallbackView)
    await flushPromises()

    expect(showErrorMock).toHaveBeenCalledWith('配置缺少恢复令牌。')
    expect(replaceMock).not.toHaveBeenCalled()
  })

  it('shows an error when the callback payload is missing the resume token', async () => {
    locationState.current.hash = '#payment_type=wxpay'

    const wrapper = mount(WechatPaymentCallbackView)
    await flushPromises()

    expect(replaceMock).not.toHaveBeenCalled()
    expect(showErrorMock).toHaveBeenCalledWith('配置缺少恢复令牌。')
    expect(wrapper.text()).toContain('配置缺少恢复令牌。')
    expect(wrapper.find('.bg-red-50').exists()).toBe(false)
  })

  it('uses the configured purchase fallback for the error recovery button', async () => {
    locationState.current.hash = '#payment_type=wxpay'

    const wrapper = mount(WechatPaymentCallbackView)
    await flushPromises()

    await wrapper.get('button').trigger('click')

    expect(replaceMock).toHaveBeenCalledWith('/configured-purchase')
  })

  it('uses payment shell settings for visible callback labels', async () => {
    locationState.current.hash = '#payment_type=wxpay'

    const wrapper = mount(WechatPaymentCallbackView)
    await flushPromises()

    expect(wrapper.text()).toContain('配置恢复微信支付')
    expect(wrapper.text()).toContain('配置返回支付页')
  })

  it('does not read visible callback labels from auth i18n', () => {
    const source = readFileSync('src/views/auth/WechatPaymentCallbackView.vue', 'utf8')

    expect(source).toContain('resolveWechatPaymentCallbackLabels')
    expect(source).toContain('useAuthRouteDefaults')
    expect(source).toContain('authRouteDefaults.value.purchasePath')
    expect(source).not.toContain("return '/purchase'")
    expect(source).not.toContain("router.replace('/purchase')")
    expect(source).toContain("paymentText('wechatPaymentCallbackTitle')")
    expect(source).not.toContain("t('auth.wechatPayment.callbackTitle')")
    expect(source).not.toContain("t('auth.wechatPayment.callbackProcessing')")
    expect(source).not.toContain("t('auth.wechatPayment.backToPayment')")
    expect(source).not.toContain("t('auth.wechatPayment.callbackMissingResumeToken')")
  })
})
