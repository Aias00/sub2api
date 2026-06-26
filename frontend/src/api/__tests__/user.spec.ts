import { beforeEach, describe, expect, it, vi } from 'vitest'

describe('user api oauth binding urls', () => {
  beforeEach(() => {
    vi.resetModules()
  })

  it('builds third-party bind urls against the bind start endpoint', async () => {
    const { buildOAuthBindingStartURL } = await import('@/api/user')

    expect(buildOAuthBindingStartURL('linuxdo', { redirectTo: '/settings/profile' })).toBe(
      '/api/v1/auth/oauth/linuxdo/bind/start?redirect=%2Fsettings%2Fprofile&intent=bind_current_user'
    )
    expect(
      buildOAuthBindingStartURL('linuxdo', {
        redirectTo: '/settings/profile',
        apiBaseSettings: { api_base_url: 'https://runtime.example.com/api/v1/' },
      })
    ).toBe(
      'https://runtime.example.com/api/v1/auth/oauth/linuxdo/bind/start?redirect=%2Fsettings%2Fprofile&intent=bind_current_user'
    )
    expect(
      buildOAuthBindingStartURL('wechat', {
        redirectTo: '/settings/profile',
        apiBaseSettings: { api_base_url: 'https://runtime.example.com/api/v1/' },
        wechatOAuthSettings: {
          wechat_oauth_open_enabled: true,
          wechat_oauth_mp_enabled: false,
          wechat_oauth_mobile_enabled: false
        }
      })
    ).toBe(
      'https://runtime.example.com/api/v1/auth/oauth/wechat/bind/start?redirect=%2Fsettings%2Fprofile&intent=bind_current_user&mode=open'
    )
  })

  it('uses the centralized safe profile redirect for bind urls', async () => {
    const { buildOAuthBindingStartURL } = await import('@/api/user')

    expect(buildOAuthBindingStartURL('linuxdo')).toBe(
      '/api/v1/auth/oauth/linuxdo/bind/start?redirect=%2Fprofile&intent=bind_current_user'
    )
    expect(buildOAuthBindingStartURL('linuxdo', { redirectTo: 'https://evil.example/profile' })).toBe(
      '/api/v1/auth/oauth/linuxdo/bind/start?redirect=%2Fprofile&intent=bind_current_user'
    )
  })
})
