import type { UserAuthProvider } from '@/types'

export type ProfileShellLabels<K extends string, P extends string> = Partial<Record<K, string>> & {
  providers?: Partial<Record<P, string>>
}

export const profileLabelKeys = [
  'contactSupport',
  'user',
  'administrator',
  'accountBalance',
  'concurrencyLimit',
  'memberSince',
  'basicsTitle',
  'basicsDescription',
  'linkedProfileSources',
  'linkedProfileSourcesDescription',
  'sourceAvatar',
  'sourceUsername',
  'profileStatusActive',
  'profileStatusDisabled',
  'profileEditTitle',
  'profileUsername',
  'profileUsernamePlaceholder',
  'profileUpdating',
  'profileUpdateAction',
  'profileUsernameRequired',
  'profileUpdateSuccess',
  'profileUpdateFailed',
  'changePassword',
  'currentPassword',
  'newPassword',
  'confirmNewPassword',
  'passwordHint',
  'changingPassword',
  'changePasswordButton',
  'passwordsNotMatch',
  'passwordTooShort',
  'passwordChangeSuccess',
  'passwordChangeFailed',
  'balanceNotifyTitle',
  'balanceNotifyDescription',
  'balanceNotifyEnabled',
  'balanceNotifyThreshold',
  'balanceNotifyThresholdHint',
  'balanceNotifySystemDefault',
  'balanceNotifyThresholdPlaceholder',
  'balanceNotifyExtraEmails',
  'balanceNotifyExtraEmailsHint',
  'balanceNotifyCodePlaceholder',
  'balanceNotifyVerify',
  'balanceNotifyResend',
  'balanceNotifyUnverified',
  'balanceNotifyVerified',
  'balanceNotifyRemoveEmail',
  'balanceNotifySendCode',
  'balanceNotifyEmailPlaceholder',
  'balanceNotifyMaxEmailsReached',
  'balanceNotifyEmailDuplicate',
  'balanceNotifyCodeSent',
  'balanceNotifyVerifySuccess',
  'balanceNotifyRemoveSuccess',
  'balanceNotifySaving',
  'balanceNotifySave',
  'balanceNotifyCancel',
  'balanceNotifyAdd',
  'balanceNotifySaved',
  'balanceNotifyError',
  'avatarTitle',
  'avatarDescription',
  'avatarUploadHint',
  'avatarUploadAction',
  'avatarUploadRequired',
  'avatarReadFailed',
  'avatarCompressFailed',
  'avatarCompressTooLarge',
  'avatarInvalidType',
  'avatarGifTooLarge',
  'avatarSave',
  'avatarDelete',
  'avatarSaveSuccess',
  'avatarEmptyDeleteHint',
  'avatarDeleteSuccess',
  'avatarError',
  'totpTitle',
  'totpDescription',
  'totpFeatureDisabled',
  'totpFeatureDisabledHint',
  'totpEnabled',
  'totpEnabledAt',
  'totpDisable',
  'totpNotEnabled',
  'totpNotEnabledHint',
  'totpEnable',
  'totpSetupTitle',
  'totpVerifyEmailFirst',
  'totpVerifyPasswordFirst',
  'totpSetupStep1',
  'totpSetupStep2',
  'totpEmailCode',
  'totpEnterEmailCode',
  'totpSendCode',
  'totpSending',
  'totpEnterPassword',
  'totpManualEntry',
  'totpEnterCode',
  'totpVerify',
  'totpCancel',
  'totpNext',
  'totpBack',
  'totpLoading',
  'totpVerifying',
  'totpCopied',
  'totpCopyFailed',
  'totpCodeSent',
  'totpSendCodeFailed',
  'totpSetupFailed',
  'totpEnableSuccess',
  'totpVerifyFailed',
  'totpDisableTitle',
  'totpDisableWarning',
  'totpConfirmDisable',
  'totpProcessing',
  'totpDisableSuccess',
  'totpDisableFailed',
  'totpError',
  'authBindingsTitle',
  'authBindingsDescription',
  'authBindingsStatusBound',
  'authBindingsStatusNotBound',
  'authBindingsStatusPasswordNotSet',
  'authBindingsBindAction',
  'authBindingsEmailPlaceholder',
  'authBindingsCodePlaceholder',
  'authBindingsPasswordPlaceholder',
  'authBindingsReplaceEmailPasswordPlaceholder',
  'authBindingsSendCodeAction',
  'authBindingsUnbindAction',
  'authBindingsManageEmailAction',
  'authBindingsHideEmailFormAction',
  'authBindingsConfirmEmailBindAction',
  'authBindingsConfirmEmailReplaceAction',
  'authBindingsBoundCount',
  'authBindingsUnbindSuccess',
  'authBindingsCodeSentTo',
  'authBindingsBindSuccess',
  'authBindingsReplaceSuccess',
  'authBindingsLoading',
  'authBindingsTryAgain',
  'authBindingsEmailRequired',
  'authBindingsInvalidEmail',
  'authBindingsCodeRequired',
  'authBindingsPasswordRequired',
  'authBindingsPasswordMinLength',
  'authBindingsSendCodeFailed',
  'authBindingsNoteEmailManagedFromProfile',
  'authBindingsNoteCanUnbind',
  'authBindingsNoteBindAnotherBeforeUnbind',
] as const

export type ProfileLabelKey = typeof profileLabelKeys[number]
export type ProfileLabels = Partial<Record<ProfileLabelKey, string>>

export const profileProviderKeys = ['email', 'linuxdo', 'dingtalk', 'oidc', 'wechat', 'github', 'google'] as const

export type ProfileProviderKey = typeof profileProviderKeys[number]
export type ProfileViewShellLabels = ProfileShellLabels<ProfileLabelKey, ProfileProviderKey>

export const authBindingLabelKeys = [
  'authBindingsTitle',
  'authBindingsDescription',
  'authBindingsStatusBound',
  'authBindingsStatusNotBound',
  'authBindingsStatusPasswordNotSet',
  'authBindingsBindAction',
  'authBindingsEmailPlaceholder',
  'authBindingsCodePlaceholder',
  'authBindingsPasswordPlaceholder',
  'authBindingsReplaceEmailPasswordPlaceholder',
  'authBindingsSendCodeAction',
  'authBindingsUnbindAction',
  'authBindingsManageEmailAction',
  'authBindingsHideEmailFormAction',
  'authBindingsConfirmEmailBindAction',
  'authBindingsConfirmEmailReplaceAction',
  'authBindingsBoundCount',
  'authBindingsUnbindSuccess',
  'authBindingsCodeSentTo',
  'authBindingsBindSuccess',
  'authBindingsReplaceSuccess',
  'authBindingsLoading',
  'authBindingsTryAgain',
  'authBindingsEmailRequired',
  'authBindingsInvalidEmail',
  'authBindingsCodeRequired',
  'authBindingsPasswordRequired',
  'authBindingsPasswordMinLength',
  'authBindingsSendCodeFailed',
  'authBindingsNoteEmailManagedFromProfile',
  'authBindingsNoteCanUnbind',
  'authBindingsNoteBindAnotherBeforeUnbind',
] as const

export type AuthBindingLabelKey = typeof authBindingLabelKeys[number]

export type AuthBindingLabels = Partial<Record<AuthBindingLabelKey, string>> & {
  providers?: Partial<Record<UserAuthProvider, string>>
}

export const authBindingLabelKeySet = new Set<string>(authBindingLabelKeys)

export const legacyAuthBindingNoteKeys: Record<string, AuthBindingLabelKey> = {
  'Primary account email is managed from the profile form.':
    'authBindingsNoteEmailManagedFromProfile',
  'You can unbind this sign-in method.': 'authBindingsNoteCanUnbind',
  'Bind another sign-in method before unbinding.':
    'authBindingsNoteBindAnotherBeforeUnbind',
}

export const authBindingNoteKeyMap: Record<string, AuthBindingLabelKey> = {
  'profile.authBindings.notes.emailManagedFromProfile': 'authBindingsNoteEmailManagedFromProfile',
  'profile.authBindings.notes.canUnbind': 'authBindingsNoteCanUnbind',
  'profile.authBindings.notes.bindAnotherBeforeUnbind': 'authBindingsNoteBindAnotherBeforeUnbind',
}

export function interpolateProfileShellLabel(
  template: string,
  params?: Record<string, string | number>,
): string {
  if (!params) return template
  return template.replace(/\{(\w+)\}/g, (match, key) => {
    const value = params[key]
    return value === undefined ? match : String(value)
  })
}

export function resolveAuthBindingText(
  labels: AuthBindingLabels | undefined,
  key: AuthBindingLabelKey,
  params?: Record<string, string | number>,
): string {
  const configured = labels?.[key]
  return interpolateProfileShellLabel(configured || '', params)
}

export function resolveAuthBindingProviderLabel(
  labels: AuthBindingLabels | undefined,
  provider: UserAuthProvider,
  oidcProviderName: string,
): string {
  const configured = labels?.providers?.[provider]
  if (provider === 'oidc') {
    return configured?.replace(/\{providerName\}/g, oidcProviderName) || ''
  }
  return configured || ''
}

export function resolveProfileShellLabels<K extends string, P extends string>(
  raw: string | undefined,
  runtimeLocale: string,
  allowedKeys: readonly K[],
  providerKeys: readonly P[],
): ProfileShellLabels<K, P> {
  if (!raw) return {}
  try {
    const parsed = JSON.parse(raw) as Record<string, { labels?: Record<string, unknown> } | undefined>
    const normalizedLocale = runtimeLocale.toLowerCase()
    const localeKeys = [normalizedLocale, normalizedLocale.split('-')[0], 'en', 'zh']
    for (const key of localeKeys) {
      const labels = parsed[key]?.labels
      if (!labels) continue
      const result: ProfileShellLabels<K, P> = {}
      const baseLabels: Partial<Record<K, string>> = {}
      for (const labelKey of allowedKeys) {
        const value = labels[labelKey]
        if (typeof value === 'string') {
          baseLabels[labelKey] = value
        }
      }
      Object.assign(result, baseLabels)
      if (labels.providers && typeof labels.providers === 'object') {
        const providers = labels.providers as Record<string, unknown>
        result.providers = {}
        for (const providerKey of providerKeys) {
          if (typeof providers[providerKey] === 'string') {
            result.providers[providerKey] = providers[providerKey]
          }
        }
      }
      return result
    }
  } catch {
    return {}
  }
  return {}
}
