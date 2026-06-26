<template>
  <AuthLayout>
    <div class="space-y-6">
      <div class="text-center">
        <h2 class="text-2xl font-bold text-gray-900 dark:text-white">
          {{ authText('providerCallbackTitle', { providerName }) }}
        </h2>
        <p class="mt-2 text-sm text-gray-500 dark:text-dark-400">
          {{
            isProcessing
              ? authText('providerCallbackProcessing', { providerName })
              : authText('providerCallbackHint')
          }}
        </p>
      </div>

      <transition name="fade">
        <div
          v-if="
            needsInvitation ||
            needsChooser ||
            needsAdoptionConfirmation ||
            needsCreateAccount ||
            needsBindLogin ||
            needsTotpChallenge
          "
          class="space-y-4"
        >
          <div
            v-if="adoptionRequired && (suggestedDisplayName || suggestedAvatarUrl)"
            class="rounded-xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-600 dark:bg-dark-800/60"
          >
            <div class="space-y-3">
              <div class="space-y-1">
                <p class="text-sm font-medium text-gray-900 dark:text-white">
                  {{ authText('oauthFlowProfileDetailsTitle', { providerName }) }}
                </p>
                <p class="text-xs text-gray-500 dark:text-dark-400">
                  {{ authText('oauthFlowProfileDetailsDescription', { providerName }) }}
                </p>
              </div>

              <label
                v-if="suggestedDisplayName"
                class="flex items-start gap-3 rounded-lg border border-gray-200 bg-white p-3 text-sm dark:border-dark-600 dark:bg-dark-900/50"
              >
                <input v-model="adoptDisplayName" type="checkbox" class="mt-1 h-4 w-4" />
                <span class="space-y-1">
                  <span class="block font-medium text-gray-900 dark:text-white">
                    {{ authText('oauthFlowUseDisplayName') }}
                  </span>
                  <span class="block text-gray-500 dark:text-dark-400">
                    {{ suggestedDisplayName }}
                  </span>
                </span>
              </label>

              <label
                v-if="suggestedAvatarUrl"
                class="flex items-start gap-3 rounded-lg border border-gray-200 bg-white p-3 text-sm dark:border-dark-600 dark:bg-dark-900/50"
              >
                <input v-model="adoptAvatar" type="checkbox" class="mt-1 h-4 w-4" />
                <img
                  :src="suggestedAvatarUrl"
                  :alt="authText('oauthFlowAvatarAlt', { providerName })"
                  class="h-10 w-10 rounded-full border border-gray-200 object-cover dark:border-dark-600"
                />
                <span class="space-y-1">
                  <span class="block font-medium text-gray-900 dark:text-white">
                    {{ authText('oauthFlowUseAvatar') }}
                  </span>
                  <span class="block break-all text-gray-500 dark:text-dark-400">
                    {{ suggestedAvatarUrl }}
                  </span>
                </span>
              </label>
            </div>
          </div>

          <template v-if="needsInvitation">
            <p class="text-sm text-gray-700 dark:text-gray-300">
              {{ authText('providerInvitationRequired', { providerName }) }}
            </p>
            <div>
              <input
                v-model="invitationCode"
                type="text"
                class="input w-full"
                :placeholder="authText('invitationCodePlaceholder')"
                :disabled="isSubmitting"
                @keyup.enter="handleSubmitInvitation"
              />
            </div>
            <button
              class="btn btn-primary w-full"
              :disabled="isSubmitting || !invitationCode.trim()"
              @click="handleSubmitInvitation"
            >
              {{
                isSubmitting
                  ? authText('providerCompletingRegistration')
                : authText('providerCompleteRegistration')
              }}
            </button>

            <div
              class="rounded-xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-600 dark:bg-dark-800/60"
            >
              <div class="space-y-3">
                <div class="space-y-1">
                  <p class="text-sm font-medium text-gray-900 dark:text-white">
                    {{ authText('alreadyHaveAccount') }}
                  </p>
                  <p class="text-xs text-gray-500 dark:text-dark-400">
                    {{
                      hasCurrentAuthToken
                        ? authText('oauthFlowBindCurrentAccountDescription', { providerName })
                        : authText('oauthFlowSignInThenBindDescription', { providerName })
                    }}
                  </p>
                </div>

                <input
                  v-if="!hasCurrentAuthToken"
                  v-model="existingAccountEmail"
                  data-testid="existing-account-email"
                  type="email"
                  class="input w-full"
                  :placeholder="authText('emailPlaceholder')"
                  :disabled="isSubmitting"
                />

                <button
                  data-testid="existing-account-submit"
                  type="button"
                  class="btn btn-secondary w-full"
                  :disabled="isSubmitting"
                  @click="handleExistingAccountBinding"
                >
                  {{ hasCurrentAuthToken ? authText('oauthFlowBindCurrentAccount') : authText('signIn') }}
                </button>
              </div>
            </div>
          </template>

          <template v-else-if="needsChooser">
            <div
              class="rounded-xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-600 dark:bg-dark-800/60"
            >
              <div class="space-y-4">
                <div class="space-y-1">
                  <p class="text-sm font-medium text-gray-900 dark:text-white">
                    {{ authText('oauthFlowChooseHowToContinue') }}
                  </p>
                  <p class="text-xs text-gray-500 dark:text-dark-400">
                    {{ authText('oauthFlowChooseAccountActionHint') }}
                  </p>
                </div>

                <button
                  data-testid="wechat-choice-bind-existing"
                  type="button"
                  class="btn btn-primary w-full"
                  :disabled="isSubmitting"
                  @click="switchToBindLoginMode()"
                >
                  {{ authText('oauthFlowBindExistingAccount') }}
                </button>

                <button
                  data-testid="wechat-choice-create-account"
                  type="button"
                  class="btn btn-secondary w-full"
                  :disabled="isSubmitting"
                  @click="switchToCreateAccountMode()"
                >
                  {{ authText('oauthFlowCreateNewAccount') }}
                </button>
              </div>
            </div>
          </template>

          <template v-else-if="needsAdoptionConfirmation">
            <p class="text-sm text-gray-700 dark:text-gray-300">
              {{ authText('oauthFlowReviewProfileBeforeContinue', { providerName }) }}
            </p>
            <button class="btn btn-primary w-full" :disabled="isSubmitting" @click="handleContinueLogin">
              {{ isSubmitting ? authText('processing') : authText('continue') }}
            </button>
          </template>

          <template v-else-if="needsCreateAccount">
            <p class="text-sm text-gray-700 dark:text-gray-300">
              {{ authText('oauthFlowCreateAccountHint') }}
            </p>
            <PendingOAuthCreateAccountForm
              test-id-prefix="wechat"
              :initial-email="pendingAccountEmail"
              :is-submitting="isSubmitting"
              :error-message="accountActionError"
              @submit="handleCreateAccount"
              @switch-to-bind="switchToBindLoginMode"
            />
            <button
              v-if="showBackToChooser"
              class="btn btn-secondary w-full"
              :disabled="isSubmitting"
              @click="switchToCreateAccountMode()"
            >
              {{ authText('oauthFlowCreateNewAccount') }}
            </button>
          </template>

          <template v-else-if="needsBindLogin">
            <p class="text-sm text-gray-700 dark:text-gray-300">
              {{ authText('oauthFlowBindSignInToExistingAccount', { providerName }) }}
            </p>
            <div
              v-if="hasCurrentAuthToken"
              class="rounded-xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-600 dark:bg-dark-800/60"
            >
              <div class="space-y-3">
                <div class="space-y-1">
                  <p class="text-sm font-medium text-gray-900 dark:text-white">
                    {{ authText('oauthFlowBindCurrentAccountTitle') }}
                  </p>
                  <p class="text-xs text-gray-500 dark:text-dark-400">
                    {{ authText('oauthFlowBindCurrentAccountDescription', { providerName }) }}
                  </p>
                </div>

                <button
                  data-testid="existing-account-submit"
                  type="button"
                  class="btn btn-primary w-full"
                  :disabled="isSubmitting"
                  @click="handleBindCurrentAccount"
                >
                  {{ isSubmitting ? authText('processing') : authText('oauthFlowBindCurrentAccount') }}
                </button>
              </div>
            </div>
            <div v-else class="space-y-3">
              <input
                v-model="bindLoginEmail"
                data-testid="wechat-bind-login-email"
                type="email"
                class="input w-full"
                :placeholder="authText('emailPlaceholder')"
                :disabled="isSubmitting"
                @keyup.enter="handleBindLogin"
              />
              <input
                v-model="bindLoginPassword"
                data-testid="wechat-bind-login-password"
                type="password"
                class="input w-full"
                :placeholder="authText('passwordPlaceholder')"
                :disabled="isSubmitting"
                @keyup.enter="handleBindLogin"
              />
              <button
                data-testid="wechat-bind-login-submit"
                class="btn btn-primary w-full"
                :disabled="isSubmitting || !bindLoginEmail.trim() || !bindLoginPassword"
                @click="handleBindLogin"
              >
                {{ isSubmitting ? authText('processing') : authText('oauthFlowLogInAndBind') }}
              </button>
            </div>
            <button
              v-if="showBackToChooser"
              class="btn btn-secondary w-full"
              :disabled="isSubmitting"
              @click="switchToCreateAccountMode()"
            >
              {{ authText('oauthFlowCreateNewAccount') }}
            </button>
          </template>

          <template v-else-if="needsTotpChallenge">
            <p class="text-sm text-gray-700 dark:text-gray-300">
              {{
                authText('oauthFlowTotpHint', {
                  providerName,
                  account: totpUserEmailMasked || authText('oauthFlowYourAccount')
                })
              }}
            </p>
            <div class="space-y-3">
              <input
                v-model="totpCode"
                data-testid="wechat-bind-login-totp"
                type="text"
                inputmode="numeric"
                maxlength="6"
                class="input w-full"
                placeholder="123456"
                :disabled="isSubmitting"
                @keyup.enter="handleSubmitTotpChallenge"
              />
              <button
                data-testid="wechat-bind-login-totp-submit"
                class="btn btn-primary w-full"
                :disabled="isSubmitting || totpCode.trim().length !== 6"
                @click="handleSubmitTotpChallenge"
              >
                {{ isSubmitting ? authText('processing') : authText('oauthFlowVerifyAndContinue') }}
              </button>
            </div>
          </template>
        </div>
      </transition>
    </div>
  </AuthLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { AuthLayout } from '@/components/layout'
import PendingOAuthCreateAccountForm, {
  type PendingOAuthCreateAccountPayload
} from '@/components/auth/PendingOAuthCreateAccountForm.vue'
import { useAuthShellText } from '@/composables/useAuthShellText'
import { apiClient, buildApiUrl } from '@/api/client'
import { useAuthStore, useAppStore } from '@/stores'
import { getRequestErrorMessage } from './requestError'
import {
  completeWeChatOAuthRegistration,
  exchangePendingOAuthCompletion,
  getAuthToken,
  hasExplicitWeChatOAuthCapabilities,
  getOAuthCompletionKind,
  isOAuthLoginCompletion,
  login2FA,
  prepareOAuthBindAccessTokenCookie,
  resolveWeChatOAuthStartStrict,
  type OAuthAdoptionDecision,
  type PendingOAuthExchangeResponse
} from '@/api/auth'
import {
  buildStandardPendingAccountStateForChoiceMode,
  extractPendingAccountEmail as extractSharedPendingAccountEmail,
  resolvePendingAccountAction as resolveSharedPendingAccountAction,
} from './pendingAccountFlow'
import {
  clearAllAffiliateReferralCodes,
  loadOAuthAffiliateCode,
  oauthAffiliatePayload
} from '@/utils/oauthAffiliate'
import { finalizeAuthLoginSuccess } from './finalizeAuthLogin'
import { buildLoginAgreementAcceptancePayload } from '@/utils/loginAgreementConsent'
import { DEFAULT_AUTH_REDIRECT_PATH, sanitizeAuthRedirectPath } from '@/utils/authRedirect'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const { authText, defaultBindRedirectPath, defaultRedirectPath, loadAuthShellConfig } = useAuthShellText()

const authStore = useAuthStore()
const appStore = useAppStore()

const isProcessing = ref(true)
const errorMessage = ref('')
const needsInvitation = ref(false)
const needsChooser = ref(false)
const invitationCode = ref('')
const isSubmitting = ref(false)
const invitationError = ref('')
const redirectTo = ref(DEFAULT_AUTH_REDIRECT_PATH)
const adoptionRequired = ref(false)
const suggestedDisplayName = ref('')
const suggestedAvatarUrl = ref('')
const existingAccountEmail = ref('')
const adoptDisplayName = ref(true)
const adoptAvatar = ref(true)
const needsAdoptionConfirmation = ref(false)
const pendingAccountAction = ref<'none' | 'choice' | 'create_account' | 'bind_login'>('none')
const pendingAccountEmail = ref('')
const bindLoginEmail = ref('')
const bindLoginPassword = ref('')
const accountActionError = ref('')
const needsTotpChallenge = ref(false)
const totpTempToken = ref('')
const totpCode = ref('')
const totpError = ref('')
const totpUserEmailMasked = ref('')
const bindSuccessMessage = t('profile.authBindings.bindSuccess')

const providerName = computed(() => authText('wechatProviderName'))
const showBackToChooser = computed(
  () => pendingAccountAction.value === 'create_account' || pendingAccountAction.value === 'bind_login'
)
const needsCreateAccount = computed(() => pendingAccountAction.value === 'create_account')
const needsBindLogin = computed(() => pendingAccountAction.value === 'bind_login')
const hasCurrentAuthToken = computed(() => Boolean(getAuthToken()))

watch(invitationError, value => {
  if (value) {
    appStore.showError(value)
  }
})

watch(accountActionError, value => {
  if (value) {
    appStore.showError(value)
  }
})

watch(totpError, value => {
  if (value) {
    appStore.showError(value)
  }
})

watch(errorMessage, value => {
  if (value) {
    appStore.showError(value)
  }
})

type PendingWeChatCompletion = PendingOAuthExchangeResponse & {
  step?: string
  status?: string
  state?: string
  pending_email?: string
  resolved_email?: string
  existing_account_email?: string
  email?: string
  intent?: string
  requires_2fa?: boolean
  temp_token?: string
  user_email_masked?: string
}

function persistPendingAuthSession(redirect?: string) {
  authStore.setPendingAuthSession({
    token: '',
    provider: 'wechat',
    redirect: sanitizeAuthRedirectPath(redirect || redirectTo.value, defaultRedirectPath.value)
  })
}

function clearPendingAuthSession() {
  authStore.clearPendingAuthSession()
}

function parseFragmentParams(): URLSearchParams {
  const raw = typeof window !== 'undefined' ? window.location.hash : ''
  const hash = raw.startsWith('#') ? raw.slice(1) : raw
  return new URLSearchParams(hash)
}

async function ensurePublicSettingsLoaded(): Promise<void> {
  if (hasExplicitWeChatOAuthCapabilities(appStore.cachedPublicSettings)) {
    return
  }

  if (appStore.publicSettingsLoaded) {
    return
  }

  await appStore.fetchPublicSettings()
}

function resolveConfiguredWeChatOAuthMode(): 'open' | 'mp' | null {
  if (!hasExplicitWeChatOAuthCapabilities(appStore.cachedPublicSettings)) {
    return null
  }

  return resolveWeChatOAuthStartStrict(appStore.cachedPublicSettings).mode
}

function resolveWeChatOAuthUnavailableMessage(): string {
  const resolved = resolveWeChatOAuthStartStrict(appStore.cachedPublicSettings)

  switch (resolved.unavailableReason) {
    case 'capability_unknown':
      return authText('wechatAvailabilityUnknown')
    case 'external_browser_required':
      return authText('wechatSystemBrowserOnly')
    case 'wechat_browser_required':
      return authText('wechatBrowserOnly')
    case 'native_app_required':
      return authText('wechatNativeAppOnly')
    case 'not_configured':
      return authText('wechatNotConfigured')
    default:
      return t('auth.loginFailed')
  }
}

function resolveRuntimeWeChatOAuthMode(): 'open' | 'mp' {
  if (typeof navigator === 'undefined') {
    return 'open'
  }
  return /MicroMessenger/i.test(navigator.userAgent) ? 'mp' : 'open'
}

function normalizeWeChatOAuthMode(value: unknown): 'open' | 'mp' | null {
  return value === 'open' || value === 'mp' ? value : null
}

function resolveRequestedWeChatOAuthMode(): 'open' | 'mp' | null {
  const configuredMode = resolveConfiguredWeChatOAuthMode()
  if (configuredMode) {
    return configuredMode
  }

  const queryMode = normalizeWeChatOAuthMode(route.query.mode)
  if (queryMode) {
    return queryMode
  }

  return resolveRuntimeWeChatOAuthMode()
}

function resolveRedirectTarget(): string {
  return sanitizeAuthRedirectPath(
    (route.query.redirect as string | undefined) || redirectTo.value || defaultRedirectPath.value,
    defaultRedirectPath.value
  )
}

function resolveWeChatStartURL(intent: 'bind_current_user' | 'adopt_existing_user_by_email'): string | null {
  const mode = resolveRequestedWeChatOAuthMode()
  if (!mode) {
    return null
  }
  const params = new URLSearchParams({
    mode,
    redirect: resolveRedirectTarget(),
    intent,
  })

  return buildApiUrl(`/auth/oauth/wechat/start?${params.toString()}`, appStore.cachedPublicSettings)
}

function buildExistingAccountResumePath(): string | null {
  const mode = resolveRequestedWeChatOAuthMode()
  if (!mode) {
    return null
  }

  const params = new URLSearchParams({
    wechat_bind_existing: '1',
    redirect: resolveRedirectTarget(),
    mode,
  })

  const email = existingAccountEmail.value.trim()
  if (email) {
    params.set('email', email)
  }

  return `/auth/wechat/callback?${params.toString()}`
}

function currentAdoptionDecision(): OAuthAdoptionDecision {
  return {
    adoptDisplayName: adoptDisplayName.value,
    adoptAvatar: adoptAvatar.value
  }
}

function resolveResumeEmail(): string {
  return typeof route.query.email === 'string' ? route.query.email.trim() : ''
}

function serializeAdoptionDecision(decision: OAuthAdoptionDecision): Record<string, boolean> {
  const payload: Record<string, boolean> = {}
  if (typeof decision.adoptDisplayName === 'boolean') {
    payload.adopt_display_name = decision.adoptDisplayName
  }
  if (typeof decision.adoptAvatar === 'boolean') {
    payload.adopt_avatar = decision.adoptAvatar
  }
  return payload
}

async function handleBindCurrentAccount() {
  const unavailableMessage = resolveConfiguredWeChatOAuthMode() === null
    ? resolveWeChatOAuthUnavailableMessage()
    : ''

  const startURL = resolveWeChatStartURL('bind_current_user')
  if (!startURL) {
    errorMessage.value = unavailableMessage || resolveWeChatOAuthUnavailableMessage()
    return
  }

  try {
    await prepareOAuthBindAccessTokenCookie()
    window.location.href = startURL
  } catch (e: unknown) {
    errorMessage.value = getRequestErrorMessage(e, t('auth.loginFailed'))
  }
}

async function handleExistingAccountBinding() {
  if (getAuthToken()) {
    await handleBindCurrentAccount()
    return
  }

  const resumePath = buildExistingAccountResumePath()
  if (!resumePath) {
    errorMessage.value = resolveWeChatOAuthUnavailableMessage()
    return
  }

  const params = new URLSearchParams({
    redirect: resumePath,
  })
  const email = existingAccountEmail.value.trim()
  if (email) {
    params.set('email', email)
  }
  await router.replace(`/login?${params.toString()}`)
}

function applyAdoptionSuggestionState(completion: PendingOAuthExchangeResponse) {
  adoptionRequired.value = completion.adoption_required === true
  suggestedDisplayName.value = completion.suggested_display_name || ''
  suggestedAvatarUrl.value = completion.suggested_avatar_url || ''

  if (!suggestedDisplayName.value) {
    adoptDisplayName.value = false
  }
  if (!suggestedAvatarUrl.value) {
    adoptAvatar.value = false
  }
}

function hasSuggestedProfile(completion: PendingOAuthExchangeResponse): boolean {
  return Boolean(completion.suggested_display_name || completion.suggested_avatar_url)
}

function extractPendingAccountEmail(completion: PendingWeChatCompletion): string {
  return extractSharedPendingAccountEmail(
    completion,
    [
    resolveResumeEmail() ||
      '',
    ],
  )
}

function resolvePendingAccountAction(
  completion: PendingWeChatCompletion
): 'none' | 'choice' | 'create_account' | 'bind_login' {
  return resolveSharedPendingAccountAction(
    completion.step || completion.status || completion.state || completion.error || completion.intent,
    { chooseResult: 'choice' },
  ) as 'none' | 'choice' | 'create_account' | 'bind_login'
}

function applyPendingAccountAction(completion: PendingWeChatCompletion) {
  const action = resolvePendingAccountAction(completion)
  accountActionError.value = ''
  needsChooser.value = false
  needsTotpChallenge.value = false
  totpTempToken.value = ''
  totpCode.value = ''
  totpError.value = ''
  totpUserEmailMasked.value = ''

  const email = extractPendingAccountEmail(completion)
  const nextState = buildStandardPendingAccountStateForChoiceMode(action, email)
  pendingAccountAction.value = nextState.action
  pendingAccountEmail.value = nextState.pendingAccountEmail
  bindLoginEmail.value = nextState.bindLoginEmail
  bindLoginPassword.value = nextState.bindLoginPassword

  if (nextState.action === 'choice') {
    needsChooser.value = true
    return
  }
}

function applyTotpChallenge(completion: PendingWeChatCompletion): boolean {
  if (completion.requires_2fa !== true || !completion.temp_token) {
    return false
  }

  pendingAccountAction.value = 'none'
  needsChooser.value = false
  needsInvitation.value = false
  needsAdoptionConfirmation.value = false
  needsTotpChallenge.value = true
  totpTempToken.value = completion.temp_token
  totpCode.value = ''
  totpError.value = ''
  totpUserEmailMasked.value = completion.user_email_masked || ''
  isProcessing.value = false
  return true
}

function switchToBindLoginMode(nextEmail?: string) {
  pendingAccountAction.value = 'bind_login'
  needsChooser.value = false
  bindLoginEmail.value = bindLoginEmail.value.trim() || nextEmail?.trim() || pendingAccountEmail.value.trim()
  bindLoginPassword.value = ''
  accountActionError.value = ''
}

function switchToCreateAccountMode() {
  pendingAccountAction.value = 'create_account'
  needsChooser.value = false
  pendingAccountEmail.value = pendingAccountEmail.value.trim() || bindLoginEmail.value.trim()
  accountActionError.value = ''
}

function isCreateAccountRecoveryError(error: unknown): boolean {
  const data = (error as {
    response?: {
      data?: {
        reason?: string
        error?: string
        code?: string
        step?: string
        intent?: string
      }
    }
  }).response?.data
  const states = [data?.reason, data?.error, data?.code, data?.step, data?.intent]
    .map(value => value?.trim().toLowerCase())
    .filter((value): value is string => Boolean(value))

  return states.includes('email_exists') ||
    states.includes('bind_login_required') ||
    states.includes('bind_login') ||
    states.includes('adopt_existing_user_by_email') ||
    states.includes('existing_account_required') ||
    states.includes('existing_account_binding_required')
}

async function finalizeCompletion(completion: PendingOAuthExchangeResponse, redirect: string) {
  if (getOAuthCompletionKind(completion) === 'bind') {
    const bindRedirect = sanitizeAuthRedirectPath(
      completion.redirect || defaultBindRedirectPath.value,
      defaultBindRedirectPath.value
    )
    clearPendingAuthSession()
    clearAllAffiliateReferralCodes()
    appStore.showSuccess(bindSuccessMessage)
    await router.replace(bindRedirect)
    return
  }

  if (!isOAuthLoginCompletion(completion)) {
    throw new Error(t('auth.oidc.callbackMissingToken'))
  }

  await finalizeAuthLoginSuccess({
    tokenResponse: completion,
    redirect,
    authStore,
    appStore,
    router,
    successMessage: t('auth.loginSuccess'),
  })
}

async function finalizePendingAccountResponse(completion: PendingWeChatCompletion) {
  applyAdoptionSuggestionState(completion)
  const redirect = sanitizeAuthRedirectPath(
    completion.redirect || redirectTo.value,
    defaultRedirectPath.value
  )

  if (completion.error === 'invitation_required') {
    pendingAccountAction.value = 'none'
    needsInvitation.value = true
    needsAdoptionConfirmation.value = false
    isProcessing.value = false
    persistPendingAuthSession(redirect)
    return
  }

  if (applyTotpChallenge(completion)) {
    persistPendingAuthSession(redirect)
    return
  }

  applyPendingAccountAction(completion)
  if (pendingAccountAction.value !== 'none') {
    needsInvitation.value = false
    needsAdoptionConfirmation.value = false
    isProcessing.value = false
    persistPendingAuthSession(redirect)
    return
  }

  if (completion.auth_result === 'pending_session') {
    needsInvitation.value = false
    needsAdoptionConfirmation.value = false
    isProcessing.value = false
    persistPendingAuthSession(redirect)
    return
  }

  await finalizeCompletion(completion, redirect)
}

async function handleSubmitInvitation() {
  invitationError.value = ''
  if (!invitationCode.value.trim()) return

  isSubmitting.value = true
  try {
    const affCode = loadOAuthAffiliateCode()
    const decision = currentAdoptionDecision()
    const completion: PendingWeChatCompletion = affCode
      ? await completeWeChatOAuthRegistration(invitationCode.value.trim(), decision, affCode)
      : await completeWeChatOAuthRegistration(invitationCode.value.trim(), decision)
    await finalizePendingAccountResponse(completion)
  } catch (e: unknown) {
    const err = e as { message?: string; response?: { data?: { message?: string } } }
    invitationError.value =
      err.response?.data?.message || err.message || t('auth.oidc.completeRegistrationFailed')
  } finally {
    isSubmitting.value = false
  }
}

async function handleContinueLogin() {
  isSubmitting.value = true
  try {
    const completion = await exchangePendingOAuthCompletion(currentAdoptionDecision()) as PendingWeChatCompletion
    await finalizePendingAccountResponse(completion)
  } catch (e: unknown) {
    errorMessage.value = getRequestErrorMessage(e, t('auth.loginFailed'))
    needsAdoptionConfirmation.value = false
  } finally {
    isSubmitting.value = false
  }
}

async function handleCreateAccount(payload: PendingOAuthCreateAccountPayload) {
  accountActionError.value = ''
  if (!payload.email || !payload.password) return

  isSubmitting.value = true
  try {
    const { data } = await apiClient.post<PendingWeChatCompletion>('/auth/oauth/pending/create-account', {
      email: payload.email,
      password: payload.password,
      verify_code: payload.verifyCode || undefined,
      invitation_code: payload.invitationCode || undefined,
      ...oauthAffiliatePayload(loadOAuthAffiliateCode()),
      ...serializeAdoptionDecision(currentAdoptionDecision()),
      ...buildLoginAgreementAcceptancePayload()
    })
    await finalizePendingAccountResponse(data)
  } catch (e: unknown) {
    if (isCreateAccountRecoveryError(e)) {
      switchToBindLoginMode(payload.email.trim())
      return
    }
    accountActionError.value = getRequestErrorMessage(e, t('auth.loginFailed'))
  } finally {
    isSubmitting.value = false
  }
}

async function handleBindLogin() {
  accountActionError.value = ''
  const email = bindLoginEmail.value.trim()
  const password = bindLoginPassword.value
  if (!email || !password) return

  isSubmitting.value = true
  try {
    const { data } = await apiClient.post<PendingWeChatCompletion>('/auth/oauth/pending/bind-login', {
      email,
      password,
      ...serializeAdoptionDecision(currentAdoptionDecision()),
      ...buildLoginAgreementAcceptancePayload()
    })
    await finalizePendingAccountResponse(data)
  } catch (e: unknown) {
    accountActionError.value = getRequestErrorMessage(e, t('auth.loginFailed'))
  } finally {
    isSubmitting.value = false
  }
}

async function handleSubmitTotpChallenge() {
  totpError.value = ''
  const code = totpCode.value.trim()
  if (!totpTempToken.value || code.length !== 6) return

  isSubmitting.value = true
  try {
    const completion = await login2FA({
      temp_token: totpTempToken.value,
      totp_code: code
    })
    await finalizeAuthLoginSuccess({
      tokenResponse: completion,
      redirect: redirectTo.value,
      authStore,
      appStore,
      router,
      successMessage: t('auth.loginSuccess'),
    })
  } catch (e: unknown) {
    totpError.value = getRequestErrorMessage(e, t('auth.loginFailed'))
  } finally {
    isSubmitting.value = false
  }
}

onMounted(async () => {
  await loadAuthShellConfig()
  redirectTo.value = defaultRedirectPath.value
  try {
    await ensurePublicSettingsLoaded()
  } catch {
    // Binding recovery requires confirmed capability flags. Use the strict guard below.
  }

  if (typeof route.query.email === 'string') {
    const email = route.query.email.trim()
    existingAccountEmail.value = email
    bindLoginEmail.value = email
    pendingAccountEmail.value = email
  }

  if (route.query.wechat_bind_existing === '1') {
    if (getAuthToken()) {
      await handleBindCurrentAccount()
      return
    }

    const resumePath = buildExistingAccountResumePath()
    if (!resumePath) {
      errorMessage.value = resolveWeChatOAuthUnavailableMessage()
      isProcessing.value = false
      return
    }

    const params = new URLSearchParams({
      redirect: resumePath,
    })
    const email = existingAccountEmail.value.trim()
    if (email) {
      params.set('email', email)
    }
    await router.replace(`/login?${params.toString()}`)
    return
  }

  const params = parseFragmentParams()
  const error = params.get('error')
  const errorDesc = params.get('error_description') || params.get('error_message') || ''
  try {
    if (error) {
      errorMessage.value = errorDesc || error
      isProcessing.value = false
      return
    }

    const completion = await exchangePendingOAuthCompletion() as PendingWeChatCompletion
    const completionRedirect = sanitizeAuthRedirectPath(
      completion.redirect || (route.query.redirect as string | undefined) || defaultRedirectPath.value,
      defaultRedirectPath.value
    )
    applyAdoptionSuggestionState(completion)
    redirectTo.value = completionRedirect

    if (completion.error === 'invitation_required') {
      needsInvitation.value = true
      isProcessing.value = false
      persistPendingAuthSession(completionRedirect)
      return
    }

    if (applyTotpChallenge(completion)) {
      persistPendingAuthSession(completionRedirect)
      return
    }

    applyPendingAccountAction(completion)
    if (pendingAccountAction.value !== 'none') {
      isProcessing.value = false
      persistPendingAuthSession(completionRedirect)
      return
    }

    if (adoptionRequired.value && hasSuggestedProfile(completion)) {
      needsAdoptionConfirmation.value = true
      isProcessing.value = false
      persistPendingAuthSession(completionRedirect)
      return
    }

    await finalizeCompletion(completion, completionRedirect)
  } catch (e: unknown) {
    clearPendingAuthSession()
    errorMessage.value = getRequestErrorMessage(e, t('auth.loginFailed'))
    isProcessing.value = false
  }
})
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: all 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
