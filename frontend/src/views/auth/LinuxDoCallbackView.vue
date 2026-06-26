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
            needsAdoptionConfirmation ||
            needsChooser ||
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
              {{ isSubmitting ? authText('providerCompletingRegistration') : authText('providerCompleteRegistration') }}
            </button>
          </template>

          <template v-else-if="needsAdoptionConfirmation">
            <p class="text-sm text-gray-700 dark:text-gray-300">
              {{ authText('oauthFlowReviewProfileBeforeContinue', { providerName }) }}
            </p>
            <button class="btn btn-primary w-full" :disabled="isSubmitting" @click="handleContinueLogin">
              {{ isSubmitting ? authText('processing') : authText('continue') }}
            </button>
          </template>

          <template v-else-if="needsChooser">
            <div class="rounded-xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-600 dark:bg-dark-800/60">
              <div class="space-y-4">
                <div class="space-y-1">
                  <p class="text-sm font-medium text-gray-900 dark:text-white">
                    {{ authText('oauthFlowChooseHowToContinue') }}
                  </p>
                  <p class="text-xs text-gray-500 dark:text-dark-400">
                    {{
                      pendingAccountEmail
                        ? authText('oauthFlowSuggestedEmail', { email: pendingAccountEmail })
                        : authText('oauthFlowChooseAccountActionHint')
                    }}
                  </p>
                </div>

                <div class="grid gap-3 sm:grid-cols-2">
                  <button
                    class="btn btn-secondary w-full"
                    :disabled="isSubmitting"
                    @click="switchToBindLoginMode()"
                  >
                    {{ authText('oauthFlowBindExistingAccount') }}
                  </button>
                  <button
                    class="btn btn-primary w-full"
                    :disabled="isSubmitting"
                    @click="switchToCreateAccountMode"
                  >
                    {{ authText('oauthFlowCreateNewAccount') }}
                  </button>
                </div>
              </div>
            </div>
          </template>

          <template v-else-if="needsCreateAccount">
            <p class="text-sm text-gray-700 dark:text-gray-300">
              {{ authText('oauthFlowCreateAccountHint') }}
            </p>
            <PendingOAuthCreateAccountForm
              test-id-prefix="linuxdo"
              :initial-email="pendingAccountEmail"
              :is-submitting="isSubmitting"
              :error-message="accountActionError"
              @submit="handleCreateAccount"
              @switch-to-bind="switchToBindLoginMode"
            />
          </template>

          <template v-else-if="needsBindLogin">
            <p class="text-sm text-gray-700 dark:text-gray-300">
              {{ authText('oauthFlowBindLoginHint', { providerName }) }}
            </p>
            <div class="space-y-3">
              <input
                v-model="bindLoginEmail"
                data-testid="linuxdo-bind-login-email"
                type="email"
                class="input w-full"
                :placeholder="authText('emailPlaceholder')"
                :disabled="isSubmitting"
                @keyup.enter="handleBindLogin"
              />
              <input
                v-model="bindLoginPassword"
                data-testid="linuxdo-bind-login-password"
                type="password"
                class="input w-full"
                :placeholder="authText('passwordPlaceholder')"
                :disabled="isSubmitting"
                @keyup.enter="handleBindLogin"
              />
              <button
                data-testid="linuxdo-bind-login-submit"
                class="btn btn-primary w-full"
                :disabled="isSubmitting || !bindLoginEmail.trim() || !bindLoginPassword"
                @click="handleBindLogin"
              >
                {{ isSubmitting ? authText('processing') : authText('oauthFlowLogInAndBind') }}
              </button>
              <button
                v-if="canReturnToCreateAccount"
                class="btn btn-secondary w-full"
                :disabled="isSubmitting"
                @click="switchToCreateAccountMode"
              >
                {{ authText('oauthFlowUseDifferentEmail') }}
              </button>
            </div>
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
                data-testid="linuxdo-bind-login-totp"
                type="text"
                inputmode="numeric"
                maxlength="6"
                class="input w-full"
                placeholder="123456"
                :disabled="isSubmitting"
                @keyup.enter="handleSubmitTotpChallenge"
              />
              <button
                data-testid="linuxdo-bind-login-totp-submit"
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
import { apiClient } from '@/api/client'
import { useAuthStore, useAppStore } from '@/stores'
import { getRequestErrorMessage } from './requestError'
import {
  buildStandardTotpChallengeState,
  buildStandardPendingAccountState,
  buildBindLoginSwitchState,
  buildCreateAccountSwitchState,
  extractStandardPendingAccountEmail,
  isCreateAccountRecoveryError,
  resolveStandardPendingAccountAction,
} from './pendingAccountFlow'
import {
  completeLinuxDoOAuthRegistration,
  exchangePendingOAuthCompletion,
  getOAuthCompletionKind,
  isOAuthLoginCompletion,
  login2FA,
  type OAuthAdoptionDecision,
  type PendingOAuthExchangeResponse
} from '@/api/auth'
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

// Invitation code flow state
const needsInvitation = ref(false)
const invitationCode = ref('')
const isSubmitting = ref(false)
const invitationError = ref('')
const redirectTo = ref(DEFAULT_AUTH_REDIRECT_PATH)
const adoptionRequired = ref(false)
const suggestedDisplayName = ref('')
const suggestedAvatarUrl = ref('')
const adoptDisplayName = ref(true)
const adoptAvatar = ref(true)
const needsAdoptionConfirmation = ref(false)
const pendingAccountAction = ref<'none' | 'choose_account_action' | 'create_account' | 'bind_login'>('none')
const pendingAccountEmail = ref('')
const bindLoginEmail = ref('')
const bindLoginPassword = ref('')
const accountActionError = ref('')
const canReturnToCreateAccount = ref(false)
const bindSuccessMessage = t('profile.authBindings.bindSuccess')
const needsTotpChallenge = ref(false)
const totpTempToken = ref('')
const totpCode = ref('')
const totpError = ref('')
const totpUserEmailMasked = ref('')
const providerName = 'LinuxDo'

const needsCreateAccount = computed(() => pendingAccountAction.value === 'create_account')
const needsChooser = computed(() => pendingAccountAction.value === 'choose_account_action')
const needsBindLogin = computed(() => pendingAccountAction.value === 'bind_login')

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

type LinuxDoPendingActionResponse = PendingOAuthExchangeResponse & {
  step?: string
  intent?: string
  email?: string
  resolved_email?: string
  pending_email?: string
  existing_account_email?: string
}

function persistPendingAuthSession(redirect?: string) {
  authStore.setPendingAuthSession({
    token: '',
    provider: 'linuxdo',
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

function currentAdoptionDecision(): OAuthAdoptionDecision {
  return {
    adoptDisplayName: adoptDisplayName.value,
    adoptAvatar: adoptAvatar.value
  }
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

function applyAdoptionSuggestionState(completion: {
  adoption_required?: boolean
  suggested_display_name?: string
  suggested_avatar_url?: string
}) {
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

function hasSuggestedProfile(completion: {
  suggested_display_name?: string
  suggested_avatar_url?: string
}): boolean {
  return Boolean(completion.suggested_display_name || completion.suggested_avatar_url)
}

function extractPendingAccountEmail(completion: LinuxDoPendingActionResponse): string {
  return extractStandardPendingAccountEmail(completion)
}

function resolvePendingAccountAction(
  completion: LinuxDoPendingActionResponse
): 'none' | 'choose_account_action' | 'create_account' | 'bind_login' {
  return resolveStandardPendingAccountAction(completion.step || completion.error || completion.intent) as
    | 'none'
    | 'choose_account_action'
    | 'create_account'
    | 'bind_login'
}

function applyPendingAccountAction(completion: LinuxDoPendingActionResponse) {
  const action = resolvePendingAccountAction(completion)
  accountActionError.value = ''
  needsTotpChallenge.value = false
  totpTempToken.value = ''
  totpCode.value = ''
  totpError.value = ''
  totpUserEmailMasked.value = ''

  const email = extractPendingAccountEmail(completion)
  const nextState = buildStandardPendingAccountState(action, email)
  pendingAccountAction.value = nextState.action
  pendingAccountEmail.value = nextState.pendingAccountEmail
  bindLoginEmail.value = nextState.bindLoginEmail
  bindLoginPassword.value = nextState.bindLoginPassword
  canReturnToCreateAccount.value = nextState.canReturnToCreateAccount
}

function applyTotpChallenge(completion: LinuxDoPendingActionResponse): boolean {
  const nextState = buildStandardTotpChallengeState(completion)
  if (!nextState) {
    return false
  }

  pendingAccountAction.value = nextState.action
  needsInvitation.value = nextState.needsInvitation
  needsAdoptionConfirmation.value = nextState.needsAdoptionConfirmation
  needsTotpChallenge.value = nextState.needsTotpChallenge
  totpTempToken.value = nextState.totpTempToken
  totpCode.value = nextState.totpCode
  totpError.value = nextState.totpError
  totpUserEmailMasked.value = nextState.totpUserEmailMasked
  isProcessing.value = nextState.isProcessing
  return true
}

function switchToBindLoginMode(nextEmail?: string) {
  const nextState = buildBindLoginSwitchState(
    bindLoginEmail.value,
    pendingAccountEmail.value,
    nextEmail,
  )
  pendingAccountAction.value = nextState.action
  bindLoginEmail.value = nextState.bindLoginEmail
  bindLoginPassword.value = nextState.bindLoginPassword
  accountActionError.value = nextState.accountActionError
  canReturnToCreateAccount.value = nextState.canReturnToCreateAccount
}

function switchToCreateAccountMode() {
  const nextState = buildCreateAccountSwitchState(
    pendingAccountEmail.value,
    bindLoginEmail.value,
  )
  pendingAccountAction.value = nextState.action
  pendingAccountEmail.value = nextState.pendingAccountEmail
  accountActionError.value = nextState.accountActionError
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
    throw new Error(t('auth.linuxdo.callbackMissingToken'))
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

async function finalizePendingAccountResponse(completion: LinuxDoPendingActionResponse) {
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
    const completion: LinuxDoPendingActionResponse = affCode
      ? await completeLinuxDoOAuthRegistration(invitationCode.value.trim(), decision, affCode)
      : await completeLinuxDoOAuthRegistration(invitationCode.value.trim(), decision)
    await finalizePendingAccountResponse(completion)
  } catch (e: unknown) {
    const err = e as { message?: string; response?: { data?: { message?: string } } }
    invitationError.value =
      err.response?.data?.message || err.message || t('auth.linuxdo.completeRegistrationFailed')
  } finally {
    isSubmitting.value = false
  }
}

async function handleContinueLogin() {
  isSubmitting.value = true
  try {
    const completion = await exchangePendingOAuthCompletion(currentAdoptionDecision()) as LinuxDoPendingActionResponse
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
    const { data } = await apiClient.post<LinuxDoPendingActionResponse>('/auth/oauth/pending/create-account', {
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
    const { data } = await apiClient.post<LinuxDoPendingActionResponse>('/auth/oauth/pending/bind-login', {
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
  const params = parseFragmentParams()
  const error = params.get('error')
  const errorDesc = params.get('error_description') || params.get('error_message') || ''
  try {
    if (error) {
      errorMessage.value = errorDesc || error
      isProcessing.value = false
      return
    }

    const completion = await exchangePendingOAuthCompletion()
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

    if (applyTotpChallenge(completion as LinuxDoPendingActionResponse)) {
      persistPendingAuthSession(completionRedirect)
      return
    }

    applyPendingAccountAction(completion as LinuxDoPendingActionResponse)
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
