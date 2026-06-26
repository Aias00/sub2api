import { readFileSync } from 'node:fs'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ProfileView from '@/views/user/ProfileView.vue'

const profileViewSource = readFileSync('src/views/user/ProfileView.vue', 'utf8')

const {
  fetchPublicSettingsMock,
  refreshUserMock,
  authState,
  appState
} = vi.hoisted(() => ({
  fetchPublicSettingsMock: vi.fn(),
  refreshUserMock: vi.fn(),
  authState: {
    user: null as Record<string, unknown> | null,
    refreshUser: vi.fn()
  },
  appState: {
    cachedPublicSettings: {} as Record<string, unknown>,
    fetchPublicSettings: vi.fn()
  }
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => authState
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => appState
}))

vi.mock('@/utils/format', () => ({
  formatDate: () => 'April 2026'
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
      locale: { value: 'zh-CN' }
    })
  }
})

describe('ProfileView', () => {
  beforeEach(() => {
    refreshUserMock.mockReset()
    fetchPublicSettingsMock.mockReset()
    appState.fetchPublicSettings = fetchPublicSettingsMock
    appState.cachedPublicSettings = {}
    refreshUserMock.mockResolvedValue(undefined)
    authState.refreshUser = refreshUserMock
    authState.user = {
      id: 1,
      username: 'alice',
      email: 'alice@example.com',
      role: 'user',
      balance: 10,
      concurrency: 2,
      status: 'active',
      allowed_groups: null,
      balance_notify_enabled: true,
      balance_notify_threshold: null,
      balance_notify_extra_emails: [],
      created_at: '2026-04-20T00:00:00Z',
      updated_at: '2026-04-20T00:00:00Z'
    }
    fetchPublicSettingsMock.mockImplementation(async () => {
      const settings = {
        contact_info: '',
        balance_low_notify_enabled: false,
        balance_low_notify_threshold: 0,
        password_min_length: 8,
        totp_enabled: false,
        linuxdo_oauth_enabled: true,
        wechat_oauth_enabled: true,
        wechat_oauth_open_enabled: true,
        wechat_oauth_mp_enabled: false,
        oidc_oauth_enabled: true,
        oidc_oauth_provider_name: 'OIDC'
      }
      appState.cachedPublicSettings = {
        ...appState.cachedPublicSettings,
        ...settings,
      }
      return settings
    })
  })

  it('passes configured profile shell labels into the overview and support card', async () => {
    appState.cachedPublicSettings = {
      profile_shell_config: JSON.stringify({
        zh: {
          labels: {
            contactSupport: '配置客服',
            accountBalance: '配置余额',
            providers: {
              wechat: '配置微信'
            },
            sourceAvatar: '配置头像来自 {providerName}',
            changePassword: '配置修改密码',
            balanceNotifyTitle: '配置余额提醒',
            authBindingsTitle: '配置登录绑定'
          }
        }
      })
    }
    authState.user = {
      ...authState.user,
      profile_sources: {
        avatar: 'wechat'
      }
    }
    fetchPublicSettingsMock.mockImplementation(async () => {
      const settings = {
        contact_info: 'support@example.com',
        balance_low_notify_enabled: true,
        balance_low_notify_threshold: 0,
        password_min_length: 8,
        totp_enabled: false,
        linuxdo_oauth_enabled: true,
        wechat_oauth_enabled: true,
        wechat_oauth_open_enabled: true,
        wechat_oauth_mp_enabled: false,
        oidc_oauth_enabled: true,
        oidc_oauth_provider_name: 'OIDC'
      }
      appState.cachedPublicSettings = {
        ...appState.cachedPublicSettings,
        ...settings,
      }
      return settings
    })

    const wrapper = mount(ProfileView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          ProfileInfoCard: {
            props: ['labels'],
            template: '<div data-testid="profile-info-card">{{ labels.accountBalance }} {{ labels.providers.wechat }} {{ labels.sourceAvatar }} {{ labels.authBindingsTitle }}</div>'
          },
          ProfileBalanceNotifyCard: { props: ['labels'], template: '<div data-testid="profile-balance-notify-card">{{ labels.balanceNotifyTitle }}</div>' },
          ProfilePasswordForm: { props: ['emailBound', 'labels'], template: '<div data-testid="profile-password-form">{{ emailBound }} {{ labels.changePassword }}</div>' },
          ProfileTotpCard: { template: '<div data-testid="profile-totp-card" />' },
          Icon: true
        }
      }
    })

    await flushPromises()

    expect(wrapper.text()).toContain('配置客服')
    expect(wrapper.text()).toContain('support@example.com')
    expect(wrapper.get('[data-testid="profile-info-card"]').text()).toContain('配置余额')
    expect(wrapper.get('[data-testid="profile-info-card"]').text()).toContain('配置微信')
    expect(wrapper.get('[data-testid="profile-info-card"]').text()).toContain('配置头像来自 {providerName}')
    expect(wrapper.get('[data-testid="profile-info-card"]').text()).toContain('配置登录绑定')
    expect(wrapper.get('[data-testid="profile-password-form"]').text()).toContain('配置修改密码')
    expect(wrapper.get('[data-testid="profile-balance-notify-card"]').text()).toContain('配置余额提醒')
  })

  it('uses cached profile shell labels while public settings fetch fills runtime flags', async () => {
    appState.cachedPublicSettings = {
      profile_shell_config: JSON.stringify({
        zh: {
          labels: {
            contactSupport: '接口客服',
            accountBalance: '接口余额',
            changePassword: '接口修改密码',
            authBindingsTitle: '接口登录绑定',
            providers: {
              wechat: '接口微信'
            }
          }
        }
      })
    }
    fetchPublicSettingsMock.mockImplementation(async () => {
      const settings = {
        contact_info: 'support@example.com',
        balance_low_notify_enabled: false,
        balance_low_notify_threshold: 0,
        password_min_length: 8,
        totp_enabled: false,
        linuxdo_oauth_enabled: true,
        wechat_oauth_enabled: true,
        wechat_oauth_open_enabled: true,
        wechat_oauth_mp_enabled: false,
        oidc_oauth_enabled: true,
        oidc_oauth_provider_name: 'OIDC'
      }
      appState.cachedPublicSettings = {
        ...appState.cachedPublicSettings,
        ...settings,
      }
      return settings
    })

    const wrapper = mount(ProfileView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          ProfileInfoCard: {
            props: ['labels'],
            template: '<div data-testid="profile-info-card">{{ labels.accountBalance }} {{ labels.providers?.wechat }} {{ labels.authBindingsTitle }}</div>'
          },
          ProfileBalanceNotifyCard: { props: ['labels'], template: '<div data-testid="profile-balance-notify-card" />' },
          ProfilePasswordForm: { props: ['labels'], template: '<div data-testid="profile-password-form">{{ labels.changePassword }}</div>' },
          ProfileTotpCard: { template: '<div data-testid="profile-totp-card" />' },
          Icon: true
        }
      }
    })

    await flushPromises()

    expect(wrapper.text()).toContain('接口客服')
    expect(wrapper.get('[data-testid="profile-info-card"]').text()).toContain('接口余额')
    expect(wrapper.get('[data-testid="profile-info-card"]').text()).toContain('接口微信')
    expect(wrapper.get('[data-testid="profile-info-card"]').text()).toContain('接口登录绑定')
    expect(wrapper.get('[data-testid="profile-password-form"]').text()).toContain('接口修改密码')
  })

  it('renders the simplified single-column profile shell without separate stat cards', async () => {
    const wrapper = mount(ProfileView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          StatCard: { template: '<div class="stat-card" />' },
          ProfileInfoCard: { template: '<div data-testid="profile-info-card" />' },
          ProfileBalanceNotifyCard: { template: '<div data-testid="profile-balance-notify-card" />' },
          ProfilePasswordForm: { props: ['emailBound'], template: '<div data-testid="profile-password-form">{{ emailBound }}</div>' },
          ProfileTotpCard: { template: '<div data-testid="profile-totp-card" />' },
          Icon: true
        }
      }
    })

    await flushPromises()

    expect(wrapper.findAll('.stat-card')).toHaveLength(0)
    expect(wrapper.get('[data-testid="profile-shell"]').exists()).toBe(true)
    expect(wrapper.get('[data-testid="profile-shell"]').html()).toContain('profile-info-card')
    expect(wrapper.get('[data-testid="profile-shell"]').html()).toContain('profile-password-form')
    expect(wrapper.get('[data-testid="profile-shell"]').text()).toContain('false')
    expect(wrapper.get('[data-testid="profile-shell"]').html()).not.toContain('profile-totp-card')
  })

  it('does not keep a frontend-local OIDC provider name default', () => {
    expect(profileViewSource).not.toContain("ref('OIDC')")
    expect(profileViewSource).not.toContain("|| 'OIDC'")
    expect(profileViewSource).toContain("const oidcOAuthProviderName = ref('')")
  })

  it('does not keep a local fetched profile shell config fallback', () => {
    expect(profileViewSource).not.toContain('fetchedProfileShellConfig')
    expect(profileViewSource).toContain('appStore.cachedPublicSettings?.profile_shell_config')
    expect(profileViewSource).not.toContain('profile_shell_config || fetchedProfileShellConfig.value')
  })

  it('renders the TOTP card only when the feature is enabled in public settings', async () => {
    appState.cachedPublicSettings = {
      profile_shell_config: JSON.stringify({
        zh: {
          labels: {
            totpTitle: '配置两步验证'
          }
        }
      })
    }
    fetchPublicSettingsMock.mockResolvedValue({
      contact_info: '',
      balance_low_notify_enabled: false,
      balance_low_notify_threshold: 0,
      password_min_length: 8,
      totp_enabled: true,
      linuxdo_oauth_enabled: true,
      wechat_oauth_enabled: true,
      wechat_oauth_open_enabled: true,
      wechat_oauth_mp_enabled: false,
      oidc_oauth_enabled: true,
      oidc_oauth_provider_name: 'OIDC'
    })

    const wrapper = mount(ProfileView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          StatCard: { template: '<div class="stat-card" />' },
          ProfileInfoCard: { template: '<div data-testid="profile-info-card" />' },
          ProfileBalanceNotifyCard: { template: '<div data-testid="profile-balance-notify-card" />' },
          ProfilePasswordForm: { props: ['emailBound'], template: '<div data-testid="profile-password-form">{{ emailBound }}</div>' },
          ProfileTotpCard: { props: ['labels'], template: '<div data-testid="profile-totp-card">{{ labels.totpTitle }}</div>' },
          Icon: true
        }
      }
    })

    await flushPromises()

    expect(wrapper.get('[data-testid="profile-shell"]').html()).toContain('profile-totp-card')
    expect(wrapper.get('[data-testid="profile-totp-card"]').text()).toContain('配置两步验证')
  })

  it('passes oauth-only users into password setup mode', async () => {
    authState.user = {
      ...authState.user,
      email_bound: false,
    }

    const wrapper = mount(ProfileView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          ProfileInfoCard: { template: '<div data-testid="profile-info-card" />' },
          ProfileBalanceNotifyCard: { template: '<div data-testid="profile-balance-notify-card" />' },
          ProfilePasswordForm: { props: ['emailBound'], template: '<div data-testid="profile-password-form">{{ emailBound }}</div>' },
          ProfileTotpCard: { template: '<div data-testid="profile-totp-card" />' },
          Icon: true
        }
      }
    })

    await flushPromises()
    expect(wrapper.get('[data-testid="profile-password-form"]').text()).toContain('false')
  })

  it('defaults missing email_bound to password setup mode until profile metadata is loaded', async () => {
    authState.user = {
      ...authState.user,
      email_bound: undefined,
    }

    const wrapper = mount(ProfileView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          ProfileInfoCard: { template: '<div data-testid="profile-info-card" />' },
          ProfileBalanceNotifyCard: { template: '<div data-testid="profile-balance-notify-card" />' },
          ProfilePasswordForm: { props: ['emailBound'], template: '<div data-testid="profile-password-form">{{ emailBound }}</div>' },
          ProfileTotpCard: { template: '<div data-testid="profile-totp-card" />' },
          Icon: true
        }
      }
    })

    await flushPromises()
    expect(wrapper.get('[data-testid="profile-password-form"]').text()).toContain('false')
  })

  it('does not render profile shell label keys as fallback copy', () => {
    expect(profileViewSource).not.toContain('const profileLabelKeys')
    expect(profileViewSource).not.toContain('const profileProviderKeys')
    expect(profileViewSource).toContain("profileLabelKeys,\n  profileProviderKeys")
    expect(profileViewSource).not.toContain('profileShellLabels.value[key] || key')
  })
})
