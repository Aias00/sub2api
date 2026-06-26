import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import {
  authBindingLabelKeySet,
  authBindingNoteKeyMap,
  legacyAuthBindingNoteKeys,
  type ProfileLabels,
  profileLabelKeys,
  profileProviderKeys,
  resolveAuthBindingProviderLabel,
  resolveAuthBindingText,
  resolveProfileShellLabels,
  type AuthBindingLabels,
} from '../profileShell'

const labelKeys = ['accountBalance', 'changePassword'] as const
const providerKeys = ['email', 'wechat', 'github'] as const

describe('resolveProfileShellLabels', () => {
  it('resolves profile labels and provider labels', () => {
    const labels = resolveProfileShellLabels(
      JSON.stringify({
        zh: {
          labels: {
            accountBalance: '账户余额',
            ignored: 'ignored',
            providers: {
              wechat: '微信',
              github: 'GitHub',
              unknown: 'Unknown',
            },
          },
        },
      }),
      'zh-CN',
      labelKeys,
      providerKeys,
    )

    expect(labels).toEqual({
      accountBalance: '账户余额',
      providers: {
        wechat: '微信',
        github: 'GitHub',
      },
    })
  })

  it('returns empty labels for invalid JSON', () => {
    expect(resolveProfileShellLabels('{bad json', 'en', labelKeys, providerKeys)).toEqual({})
  })

  it('centralizes auth binding labels, interpolation, and provider names', () => {
    const labels: AuthBindingLabels = {
      authBindingsBindAction: 'Bind {providerName}',
      authBindingsCodeSentTo: 'Code sent to {email}',
      providers: {
        oidc: 'Configured {providerName}',
        wechat: 'WeChat',
      },
    }

    expect(authBindingLabelKeySet.has('authBindingsBindAction')).toBe(true)
    expect(resolveAuthBindingText(labels, 'authBindingsBindAction', { providerName: 'WeChat' })).toBe('Bind WeChat')
    expect(resolveAuthBindingText(labels, 'authBindingsCodeSentTo', { email: 'a@example.com' })).toBe(
      'Code sent to a@example.com'
    )
    expect(resolveAuthBindingText(labels, 'authBindingsEmailRequired')).toBe('')
    expect(resolveAuthBindingProviderLabel(labels, 'oidc', 'ExampleID')).toBe('Configured ExampleID')
    expect(resolveAuthBindingProviderLabel(labels, 'wechat', 'ExampleID')).toBe('WeChat')
  })

  it('centralizes auth binding note compatibility maps', () => {
    expect(legacyAuthBindingNoteKeys['You can unbind this sign-in method.']).toBe('authBindingsNoteCanUnbind')
    expect(authBindingNoteKeyMap['profile.authBindings.notes.bindAnotherBeforeUnbind']).toBe(
      'authBindingsNoteBindAnotherBeforeUnbind'
    )
  })

  it('centralizes profile view label and provider schemas', () => {
    const labels: ProfileLabels = { accountBalance: 'Configured balance' }

    expect(labels.accountBalance).toBe('Configured balance')
    expect(profileLabelKeys).toContain('accountBalance')
    expect(profileLabelKeys).toContain('authBindingsTitle')
    expect(profileProviderKeys).toContain('wechat')
    expect(profileProviderKeys).toContain('oidc')
  })

  it('keeps profile child components from owning local label-key unions', () => {
    const componentFiles = [
      'src/components/user/profile/ProfileAvatarCard.vue',
      'src/components/user/profile/ProfileBalanceNotifyCard.vue',
      'src/components/user/profile/ProfileEditForm.vue',
      'src/components/user/profile/ProfileInfoCard.vue',
      'src/components/user/profile/ProfilePasswordForm.vue',
      'src/components/user/profile/ProfileTotpCard.vue',
      'src/components/user/profile/TotpDisableDialog.vue',
      'src/components/user/profile/TotpSetupModal.vue',
    ]

    for (const file of componentFiles) {
      const source = readFileSync(file, 'utf8')
      expect(source).not.toMatch(/type\s+\w*LabelKey\s*=/)
      expect(source).not.toMatch(/Partial<Record<ProfileLabelKey,\s*string>>/)
      expect(source).toContain("from '@/utils/profileShell'")
    }
  })
})
