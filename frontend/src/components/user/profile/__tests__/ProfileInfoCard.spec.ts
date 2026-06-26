import { readFileSync } from 'node:fs'
import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import ProfileInfoCard from '@/components/user/profile/ProfileInfoCard.vue'
import type { User } from '@/types'

vi.mock('vue-router', () => ({
  useRoute: () => ({
    fullPath: '/profile'
  })
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    user: null
  })
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
    showSuccess: vi.fn()
  })
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, string>) => {
        if (key === 'profile.accountBalance') return 'Account Balance'
        if (key === 'profile.concurrencyLimit') return 'Concurrency Limit'
        if (key === 'profile.memberSince') return 'Member Since'
        if (key === 'profile.administrator') return 'Administrator'
        if (key === 'profile.user') return 'User'
        if (key === 'profile.authBindings.providers.email') return 'Email'
        if (key === 'profile.authBindings.providers.linuxdo') return 'LinuxDo'
        if (key === 'profile.authBindings.providers.wechat') return 'WeChat'
        if (key === 'profile.authBindings.providers.oidc') return params?.providerName || 'OIDC'
        if (key === 'profile.authBindings.source.avatar') {
          return `Avatar synced from ${params?.providerName || 'provider'}`
        }
        if (key === 'profile.authBindings.source.username') {
          return `Username synced from ${params?.providerName || 'provider'}`
        }
        return key
      }
    })
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
    ...overrides
  }
}

const profileInfoCardSource = readFileSync('src/components/user/profile/ProfileInfoCard.vue', 'utf8')

const overviewLabels = {
  user: 'Configured user',
  accountBalance: 'Configured balance',
  concurrencyLimit: 'Configured concurrency',
  memberSince: 'Configured member since',
}

describe('ProfileInfoCard', () => {
  it('renders basic account information inside the new overview shell', () => {
    const wrapper = mount(ProfileInfoCard, {
      props: {
        user: createUser(),
        labels: overviewLabels,
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    expect(wrapper.text()).toContain('alice@example.com')
    expect(wrapper.text()).toContain('alice')
    expect(wrapper.text()).toContain('Configured user')
    expect(wrapper.get('[data-testid="profile-basics-panel"]').exists()).toBe(true)
    expect(wrapper.get('[data-testid="profile-auth-bindings-panel"]').exists()).toBe(true)
  })

  it('renders third-party source hints from profile sources', () => {
    const wrapper = mount(ProfileInfoCard, {
      props: {
        user: createUser({
          avatar_url: 'https://cdn.example.com/linuxdo.png',
          profile_sources: {
            avatar: { provider: 'linuxdo', source: 'linuxdo' },
            username: { provider: 'linuxdo', source: 'linuxdo' }
          }
        }),
        labels: {
          sourceAvatar: 'Configured avatar from {providerName}',
          sourceUsername: 'Configured username from {providerName}',
          providers: {
            linuxdo: 'Configured LinuxDo',
          },
        },
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    expect(wrapper.text()).toContain('Configured avatar from Configured LinuxDo')
    expect(wrapper.text()).toContain('Configured username from Configured LinuxDo')
  })

  it('uses the configured OIDC provider name in source hints', () => {
    const wrapper = mount(ProfileInfoCard, {
      props: {
        user: createUser({
          profile_sources: {
            username: { provider: 'oidc', source: 'oidc' }
          }
        }),
        oidcProviderName: 'ExampleID',
        labels: {
          sourceUsername: 'Configured username from {providerName}',
          providers: {
            oidc: '{providerName}',
          },
        },
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    expect(wrapper.text()).toContain('Configured username from ExampleID')
  })

  it('does not display synthetic oauth-only emails as a real bound email', () => {
    const wrapper = mount(ProfileInfoCard, {
      props: {
        user: createUser({
          email: 'legacy-user@oidc-connect.invalid',
          email_bound: false,
          auth_bindings: {
            email: { bound: false }
          }
        })
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    expect(wrapper.text()).not.toContain('legacy-user@oidc-connect.invalid')
  })

  it('does not display synthetic oauth-only emails when only legacy identity bindings mark email as unbound', () => {
    const wrapper = mount(ProfileInfoCard, {
      props: {
        user: createUser({
          email: 'legacy-user@wechat-connect.invalid',
          identity_bindings: {
            email: { bound: false }
          }
        })
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    expect(wrapper.text()).not.toContain('legacy-user@wechat-connect.invalid')
  })

  it('renders the approved overview hero and two-column content shell', () => {
    const wrapper = mount(ProfileInfoCard, {
      props: {
        user: createUser(),
        labels: overviewLabels,
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    expect(wrapper.get('[data-testid="profile-overview-hero"]').text()).toContain('alice@example.com')
    expect(wrapper.get('[data-testid="profile-overview-metric-balance"]').text()).toContain('Configured balance')
    expect(wrapper.get('[data-testid="profile-overview-metric-concurrency"]').text()).toContain('Configured concurrency')
    expect(wrapper.get('[data-testid="profile-overview-metric-member-since"]').text()).toContain('Configured member since')
    expect(wrapper.find('[data-testid="profile-info-summary-grid"]').exists()).toBe(false)
    expect(wrapper.get('[data-testid="profile-main-column"]').exists()).toBe(true)
    expect(wrapper.get('[data-testid="profile-side-column"]').exists()).toBe(true)
    expect(wrapper.get('[data-testid="profile-basics-panel"]').exists()).toBe(true)
    expect(wrapper.get('[data-testid="profile-auth-bindings-panel"]').exists()).toBe(true)
  })

  it('renders configured status labels in the overview hero', () => {
    const activeWrapper = mount(ProfileInfoCard, {
      props: {
        user: createUser({ status: 'active' }),
        labels: {
          profileStatusActive: 'Configured active',
          profileStatusDisabled: 'Configured disabled',
        },
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    expect(activeWrapper.get('[data-testid="profile-overview-hero"]').text()).toContain('Configured active')

    const disabledWrapper = mount(ProfileInfoCard, {
      props: {
        user: createUser({ status: 'disabled' }),
        labels: {
          profileStatusActive: 'Configured active',
          profileStatusDisabled: 'Configured disabled',
        },
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    expect(disabledWrapper.get('[data-testid="profile-overview-hero"]').text()).toContain('Configured disabled')
  })

  it('uses configured provider labels in source hints', () => {
    const wrapper = mount(ProfileInfoCard, {
      props: {
        user: createUser({
          profile_sources: {
            username: { provider: 'github', source: 'github' },
            avatar: { provider: 'google', source: 'google' },
          },
        }),
        labels: {
          sourceAvatar: 'Configured avatar from {providerName}',
          sourceUsername: 'Configured username from {providerName}',
          providers: {
            github: 'Configured GitHub',
            google: 'Configured Google',
          },
        },
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    expect(wrapper.text()).toContain('Configured username from Configured GitHub')
    expect(wrapper.text()).toContain('Configured avatar from Configured Google')
  })

  it('passes configured profile shell labels into the avatar card', () => {
    const wrapper = mount(ProfileInfoCard, {
      props: {
        user: createUser(),
        labels: {
          avatarTitle: 'Configured avatar title',
        },
      },
      global: {
        stubs: {
          Icon: true,
          ProfileAvatarCard: {
            props: ['labels'],
            template: '<div data-testid="profile-avatar-card">{{ labels.avatarTitle }}</div>',
          },
        },
      },
    })

    expect(wrapper.get('[data-testid="profile-avatar-card"]').text()).toContain('Configured avatar title')
  })

  it('passes configured profile shell labels into the edit form', () => {
    const wrapper = mount(ProfileInfoCard, {
      props: {
        user: createUser(),
        labels: {
          profileEditTitle: 'Configured edit title',
        },
      },
      global: {
        stubs: {
          Icon: true,
          ProfileEditForm: {
            props: ['labels'],
            template: '<div data-testid="profile-edit-form">{{ labels.profileEditTitle }}</div>',
          },
        },
      },
    })

    expect(wrapper.get('[data-testid="profile-edit-form"]').text()).toContain('Configured edit title')
  })

  it('passes configured profile shell labels into the identity bindings section', () => {
    const wrapper = mount(ProfileInfoCard, {
      props: {
        user: createUser(),
        labels: {
          authBindingsTitle: 'Configured bindings title',
        },
      },
      global: {
        stubs: {
          Icon: true,
          ProfileIdentityBindingsSection: {
            props: ['labels'],
            template: '<div data-testid="profile-identity-bindings">{{ labels.authBindingsTitle }}</div>',
          },
        },
      },
    })

    expect(wrapper.get('[data-testid="profile-identity-bindings"]').text()).toContain('Configured bindings title')
  })

  it('does not render local label keys or provider names as fallback copy', () => {
    expect(profileInfoCardSource).not.toContain('return interpolateLabel(configured || key, params)')
    expect(profileInfoCardSource).not.toContain("props.labels?.providers?.email || 'email'")
    expect(profileInfoCardSource).not.toContain("props.labels?.providers?.github || 'GitHub'")
    expect(profileInfoCardSource).not.toContain("props.labels?.providers?.google || 'Google'")
  })
})
