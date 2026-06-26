export type AuthShellLocale = 'zh' | 'en'

export const authShellLabelKeys = [
  'affiliateInvitationDetected',
  'agreementAcceptAndContinue',
  'agreementAcceptedDescription',
  'agreementAcceptedTitle',
  'agreementCheckboxPrefix',
  'agreementRecent',
  'agreementReject',
  'agreementRelevantDocuments',
  'agreementReviewDescription',
  'agreementReviewTitle',
  'agreementTermsUpdateTitle',
  'agreementUpdatedAt',
  'agreementViewAndAccept',
  'agreementViewTerms',
  'allRightsReserved',
  'alreadyHaveAccount',
  'backToLogin',
  'continue',
  'createAccount',
  'createPasswordPlaceholder',
  'codeSentSuccess',
  'confirmPassword',
  'confirmPasswordPlaceholder',
  'dingtalkProviderName',
  'dontHaveAccount',
  'emailLabel',
  'emailPlaceholder',
  'emailVerifyBackToRegistration',
  'emailVerifyClickToResend',
  'emailVerifyCodeHint',
  'emailVerifyCodeLabel',
  'emailVerifyCodeSentSuccess',
  'emailVerifyDescriptionPrefix',
  'emailVerifyResendCode',
  'emailVerifyResendCountdown',
  'emailVerifySessionExpiredDescription',
  'emailVerifySessionExpiredTitle',
  'emailVerifySubmit',
  'emailVerifyTitle',
  'emailVerifyVerifying',
  'forgotPasswordHint',
  'forgotPasswordTitle',
  'forgotPassword',
  'invitationCodeLabel',
  'invitationCodePlaceholder',
  'invitationCodeValid',
  'invalidResetLink',
  'invalidResetLinkHint',
  'newPassword',
  'newPasswordPlaceholder',
  'oauthAlternativeMethods',
  'oauthCallbackCode',
  'oauthCallbackHint',
  'oauthCallbackFullUrl',
  'oauthCallbackInvalidHint',
  'oauthCallbackInvalidTitle',
  'oauthCallbackPasswordOptionalHint',
  'oauthCallbackRegistrationHint',
  'oauthCallbackRegistrationInvitationRequired',
  'oauthCallbackState',
  'oauthCallbackSubmitRegistration',
  'oauthCallbackTitle',
  'oauthFlowAvatarAlt',
  'oauthFlowBindCurrentAccount',
  'oauthFlowBindCurrentAccountDescription',
  'oauthFlowBindCurrentAccountTitle',
  'oauthFlowBindExistingAccount',
  'oauthFlowBindLoginHint',
  'oauthFlowBindSignInToExistingAccount',
  'oauthFlowChooseAccountActionHint',
  'oauthFlowChooseHowToContinue',
  'oauthFlowCreateAccountHint',
  'oauthFlowCreateAccountTitle',
  'oauthFlowCreateNewAccount',
  'oauthFlowLogInAndBind',
  'oauthFlowProfileDetailsDescription',
  'oauthFlowProfileDetailsTitle',
  'oauthFlowReviewProfileBeforeContinue',
  'oauthFlowSuggestedEmail',
  'oauthFlowSignInThenBindDescription',
  'oauthFlowTotpHint',
  'oauthFlowUseAvatar',
  'oauthFlowUseDifferentEmail',
  'oauthFlowUseDisplayName',
  'oauthFlowVerifyAndContinue',
  'oauthFlowYourAccount',
  'optional',
  'passwordHint',
  'passwordLabel',
  'passwordPlaceholder',
  'passwordResetSuccess',
  'passwordResetSuccessHint',
  'processing',
  'providerCallbackHint',
  'providerCallbackProcessing',
  'providerCallbackTitle',
  'providerCompleteRegistration',
  'providerCompletingRegistration',
  'providerInvitationRequired',
  'promoCodeLabel',
  'promoCodePlaceholder',
  'promoCodeValid',
  'registrationDisabled',
  'rememberedPassword',
  'requestNewResetLink',
  'resetEmailSent',
  'resetEmailSentHint',
  'resetPassword',
  'resetPasswordHint',
  'resetPasswordTitle',
  'resettingPassword',
  'resendCountdown',
  'sendResetLink',
  'sendCode',
  'sendingCode',
  'sendingResetLink',
  'signIn',
  'signInToAccount',
  'signingIn',
  'signInWithProvider',
  'signUp',
  'signUpToStart',
  'totpLoginHint',
  'totpLoginTitle',
  'totpCancel',
  'totpVerifying',
  'verificationCodeHint',
  'wechatAvailabilityUnknown',
  'wechatBrowserOnly',
  'wechatNativeAppOnly',
  'wechatNotConfigured',
  'wechatProviderName',
  'wechatSystemBrowserOnly',
  'welcomeBack',
] as const

export type AuthShellLabelKey = typeof authShellLabelKeys[number]
export type AuthShellLabels = Partial<Record<AuthShellLabelKey, string>>

export type AuthShellConfig = {
  labels: AuthShellLabels
  defaults: {
    defaultRedirectPath?: string
    bindRedirectPath?: string
    homePath?: string
    loginPath?: string
    registerPath?: string
    forgotPasswordPath?: string
    emailVerifyPath?: string
    apiKeysPath?: string
    usagePath?: string
    availableChannelsPath?: string
    availableGroupsPath?: string
    subscriptionsPath?: string
    purchasePath?: string
    paymentResultPath?: string
    ordersPath?: string
    redeemPath?: string
    affiliatePath?: string
    profilePath?: string
    adminRedirectPath?: string
    adminRuntimeSettingsPath?: string
    adminSettingsPath?: string
    dingtalkCallbackPath?: string
    dingtalkEmailCompletionPath?: string
  }
}

export function resolveAuthShellLabels(
  raw: string | undefined,
  runtimeLocale: string,
): AuthShellLabels {
  return resolveAuthShellConfig(raw, runtimeLocale).labels
}

export function resolveAuthShellConfig(
  raw: string | undefined,
  runtimeLocale: string,
): AuthShellConfig {
  const locale = resolveAuthShellLocale(runtimeLocale)
  const localized = readAuthShellLocalizedConfig(raw, locale)

  return {
    labels: readAuthShellLabels(localized),
    defaults: readAuthShellDefaults(localized),
  }
}

export function renderAuthShellText(
  labels: AuthShellLabels,
  key: AuthShellLabelKey,
  params: Record<string, string | number> = {},
): string {
  let value = labels[key] || ''
  for (const [paramKey, paramValue] of Object.entries(params)) {
    value = value.split(`{${paramKey}}`).join(String(paramValue))
  }
  return value
}

function readAuthShellLocalizedConfig(raw: string | undefined, locale: AuthShellLocale): Record<string, unknown> | null {
  if (!raw?.trim()) {
    return null
  }
  try {
    const parsed = JSON.parse(raw) as unknown
    if (!isRecord(parsed)) {
      return null
    }
    const localized = parsed[locale] ?? parsed.en ?? parsed.zh ?? parsed
    return isRecord(localized) ? localized : null
  } catch {
    return null
  }
}

function readAuthShellLabels(localized: Record<string, unknown> | null): AuthShellLabels {
  if (!localized || !isRecord(localized.labels)) {
    return {}
  }
  try {
    const allowedKeys = new Set<string>(authShellLabelKeys)
    return Object.fromEntries(
      Object.entries(localized.labels).filter((entry): entry is [AuthShellLabelKey, string] =>
        allowedKeys.has(entry[0]) && typeof entry[1] === 'string',
      ),
    )
  } catch {
    return {}
  }
}

function readAuthShellDefaults(localized: Record<string, unknown> | null): AuthShellConfig['defaults'] {
  if (!localized || !isRecord(localized.defaults)) {
    return {}
  }
  return {
    defaultRedirectPath: readInternalPath(localized.defaults.defaultRedirectPath),
    bindRedirectPath: readInternalPath(localized.defaults.bindRedirectPath),
    homePath: readInternalPath(localized.defaults.homePath),
    loginPath: readInternalPath(localized.defaults.loginPath),
    registerPath: readInternalPath(localized.defaults.registerPath),
    forgotPasswordPath: readInternalPath(localized.defaults.forgotPasswordPath),
    emailVerifyPath: readInternalPath(localized.defaults.emailVerifyPath),
    apiKeysPath: readInternalPath(localized.defaults.apiKeysPath),
    usagePath: readInternalPath(localized.defaults.usagePath),
    availableChannelsPath: readInternalPath(localized.defaults.availableChannelsPath),
    availableGroupsPath: readInternalPath(localized.defaults.availableGroupsPath),
    subscriptionsPath: readInternalPath(localized.defaults.subscriptionsPath),
    purchasePath: readInternalPath(localized.defaults.purchasePath),
    paymentResultPath: readInternalPath(localized.defaults.paymentResultPath),
    ordersPath: readInternalPath(localized.defaults.ordersPath),
    redeemPath: readInternalPath(localized.defaults.redeemPath),
    affiliatePath: readInternalPath(localized.defaults.affiliatePath),
    profilePath: readInternalPath(localized.defaults.profilePath),
    adminRedirectPath: readInternalPath(localized.defaults.adminRedirectPath),
    adminRuntimeSettingsPath: readInternalPath(localized.defaults.adminRuntimeSettingsPath),
    adminSettingsPath: readInternalPath(localized.defaults.adminSettingsPath),
    dingtalkCallbackPath: readInternalPath(localized.defaults.dingtalkCallbackPath),
    dingtalkEmailCompletionPath: readInternalPath(localized.defaults.dingtalkEmailCompletionPath),
  }
}

function resolveAuthShellLocale(runtimeLocale: string): AuthShellLocale {
  return runtimeLocale.toLowerCase().startsWith('zh') ? 'zh' : 'en'
}

function readInternalPath(value: unknown): string | undefined {
  if (typeof value !== 'string') return undefined
  const path = value.trim()
  if (!path || !path.startsWith('/') || path.startsWith('//')) return undefined
  if (path.includes('://') || path.includes('\n') || path.includes('\r')) return undefined
  return path
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}
