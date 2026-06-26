import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

const callbackViewFiles = [
  'OAuthCallbackView.vue',
  'OidcCallbackView.vue',
  'LinuxDoCallbackView.vue',
  'DingTalkCallbackView.vue',
  'WechatCallbackView.vue',
]

const sharedOAuthFlowViewFiles = [
  'OidcCallbackView.vue',
  'LinuxDoCallbackView.vue',
  'DingTalkCallbackView.vue',
  'WechatCallbackView.vue',
]

describe('callback auth shell labels', () => {
  it('uses auth shell settings for shared callback processing labels', () => {
    for (const file of callbackViewFiles) {
      const source = readFileSync(resolve(process.cwd(), `src/views/auth/${file}`), 'utf8')

      expect(source, file).toContain('useAuthShellText')
      expect(source, file).toContain("authText('processing')")
      expect(source, file).not.toContain("t('common.processing')")
      expect(source, file).not.toContain('common.processing')
    }
  })

  it('uses auth shell settings for pending OAuth create-account processing labels', () => {
    const source = readFileSync(
      resolve(process.cwd(), 'src/components/auth/PendingOAuthCreateAccountForm.vue'),
      'utf8',
    )

    expect(source).toContain('useAuthShellText')
    expect(source).toContain("authText('processing')")
    expect(source).not.toContain("t('common.processing')")
    expect(source).not.toContain('common.processing')
  })

  it('uses auth shell settings for DingTalk email-completion shell labels', () => {
    const source = readFileSync(
      resolve(process.cwd(), 'src/views/auth/DingTalkEmailCompletionView.vue'),
      'utf8',
    )

    expect(source).toContain('useAuthShellText')
    expect(source).toContain("authText('oauthFlowCreateAccountTitle'")
    expect(source).toContain("authText('oauthFlowCreateAccountHint')")
    expect(source).toContain("authText('dingtalkProviderName')")
    expect(source).not.toContain("t('auth.dingtalk.createAccountTitle')")
    expect(source).not.toContain("t('auth.oauthFlow.createAccountHint')")
  })

  it('uses auth shell defaults for DingTalk callback/email-completion routes', () => {
    const callbackSource = readFileSync(
      resolve(process.cwd(), 'src/views/auth/DingTalkCallbackView.vue'),
      'utf8',
    )
    const emailCompletionSource = readFileSync(
      resolve(process.cwd(), 'src/views/auth/DingTalkEmailCompletionView.vue'),
      'utf8',
    )

    expect(callbackSource).toContain('authRouteDefaults.value.dingtalkEmailCompletionPath')
    expect(callbackSource).not.toContain("'/auth/dingtalk/email-completion?")
    expect(callbackSource).not.toContain('"/auth/dingtalk/email-completion?')

    expect(emailCompletionSource).toContain('authRouteDefaults.value.dingtalkCallbackPath')
    expect(emailCompletionSource).not.toContain("path: '/auth/dingtalk/callback'")
    expect(emailCompletionSource).not.toContain('path: "/auth/dingtalk/callback"')
  })

  it('uses auth shell settings for shared OAuth flow shell labels', () => {
    const migratedKeys = [
      'profileDetailsTitle',
      'profileDetailsDescription',
      'useDisplayName',
      'avatarAlt',
      'useAvatar',
      'bindCurrentAccount',
      'bindCurrentAccountDescription',
      'bindCurrentAccountTitle',
      'bindSignInToExistingAccount',
      'reviewProfileBeforeContinue',
      'chooseHowToContinue',
      'suggestedEmail',
      'signInThenBindDescription',
      'chooseAccountActionHint',
      'bindExistingAccount',
      'createNewAccount',
      'createAccountHint',
      'bindLoginHint',
      'logInAndBind',
      'useDifferentEmail',
      'totpHint',
      'yourAccount',
      'verifyAndContinue',
    ]

    for (const file of sharedOAuthFlowViewFiles) {
      const source = readFileSync(resolve(process.cwd(), `src/views/auth/${file}`), 'utf8')

      expect(source, file).toContain("authText('oauthFlowProfileDetailsTitle'")
      expect(source, file).toContain("authText('oauthFlowBindExistingAccount')")
      expect(source, file).toContain("authText('oauthFlowTotpHint'")
      for (const key of migratedKeys) {
        expect(source, file).not.toContain(`auth.oauthFlow.${key}`)
      }
    }
  })

  it('uses auth shell settings for provider callback shell labels', () => {
    const migratedKeys = [
      'callbackTitle',
      'callbackProcessing',
      'callbackHint',
      'invitationRequired',
      'completing',
      'completeRegistration',
    ]

    for (const file of sharedOAuthFlowViewFiles) {
      const source = readFileSync(resolve(process.cwd(), `src/views/auth/${file}`), 'utf8')

      expect(source, file).toContain("authText('providerCallbackTitle'")
      expect(source, file).toContain("authText('providerCallbackProcessing'")
      expect(source, file).toContain("authText('providerCallbackHint')")
      expect(source, file).toContain("authText('providerInvitationRequired'")
      expect(source, file).toContain("authText('providerCompletingRegistration')")
      expect(source, file).toContain("authText('providerCompleteRegistration')")
      expect(source, file).toContain("authText('invitationCodePlaceholder')")
      expect(source, file).toContain("authText('emailPlaceholder')")
      expect(source, file).toContain("authText('passwordPlaceholder')")
      expect(source, file).toContain("authText('continue')")
      for (const key of migratedKeys) {
        expect(source, file).not.toContain(`t('auth.oidc.${key}'`)
        expect(source, file).not.toContain(`t('auth.linuxdo.${key}'`)
        expect(source, file).not.toContain(`t('auth.dingtalk.${key}'`)
      }
      expect(source, file).not.toContain('auth.invitationCodePlaceholder')
      expect(source, file).not.toContain("t('auth.emailPlaceholder')")
      expect(source, file).not.toContain("t('auth.passwordPlaceholder')")
      expect(source, file).not.toContain("t('auth.continue')")
    }
  })

  it('uses auth shell settings for WeChat callback availability labels', () => {
    const source = readFileSync(resolve(process.cwd(), 'src/views/auth/WechatCallbackView.vue'), 'utf8')

    expect(source).toContain("authText('wechatProviderName')")
    expect(source).toContain("authText('wechatAvailabilityUnknown')")
    expect(source).toContain("authText('wechatSystemBrowserOnly')")
    expect(source).toContain("authText('wechatBrowserOnly')")
    expect(source).toContain("authText('wechatNativeAppOnly')")
    expect(source).toContain("authText('wechatNotConfigured')")
    expect(source).not.toContain("t('auth.wechatProviderName')")
    expect(source).not.toContain('auth.oauthFlow.wechatAvailabilityUnknown')
    expect(source).not.toContain('auth.oauthFlow.wechatSystemBrowserOnly')
    expect(source).not.toContain('auth.oauthFlow.wechatBrowserOnly')
    expect(source).not.toContain('auth.oauthFlow.wechatNativeAppOnly')
    expect(source).not.toContain('auth.oauthFlow.wechatNotConfigured')
  })
})
