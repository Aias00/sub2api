<template>
  <div class="vercel-auth-shell min-h-screen bg-gray-50 px-4 py-10 dark:bg-dark-900">
    <div class="mx-auto max-w-2xl">
      <div v-if="isProcessing" class="card p-6 text-center">
        <div class="mx-auto h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
        <h1 class="mt-4 text-lg font-semibold text-gray-900 dark:text-white">
          {{ authText('oauthCallbackTitle') }}
        </h1>
        <p class="mt-2 text-sm text-gray-600 dark:text-gray-400">
          {{ authText('oauthCallbackHint') }}
        </p>
      </div>

      <div v-else-if="needsRegistrationCompletion" class="card p-6">
        <h1 class="text-lg font-semibold text-gray-900 dark:text-white">
          {{ authText('oauthCallbackTitle') }}
        </h1>
        <p class="mt-2 text-sm text-gray-600 dark:text-gray-400">
          {{ registrationHint }}
        </p>

        <div class="mt-6 space-y-4">
          <div>
            <label class="input-label">{{ authText('emailLabel') }}</label>
            <input
              class="input w-full"
              type="email"
              :value="registrationEmail"
              readonly
              disabled
            />
          </div>
          <div>
            <label class="input-label">
              {{ authText('passwordLabel') }}
              <span
                v-if="passwordOptional"
                class="ml-1 text-xs font-normal text-gray-400 dark:text-gray-500"
              >
                ({{ authText('optional') }})
              </span>
            </label>
            <input
              v-model="password"
              type="password"
              class="input w-full"
              :placeholder="authText('createPasswordPlaceholder')"
              :disabled="isSubmitting"
              autocomplete="new-password"
              @keyup.enter="handleSubmitRegistration"
            />
            <p v-if="passwordOptional" class="mt-2 text-xs text-gray-500 dark:text-gray-400">
              {{ authText('oauthCallbackPasswordOptionalHint', { providerName }) }}
            </p>
          </div>
          <div>
            <label class="input-label">
              {{ authText('confirmPassword') }}
              <span
                v-if="passwordOptional"
                class="ml-1 text-xs font-normal text-gray-400 dark:text-gray-500"
              >
                ({{ authText('optional') }})
              </span>
            </label>
            <input
              v-model="confirmPassword"
              type="password"
              class="input w-full"
              :placeholder="authText('confirmPasswordPlaceholder')"
              :disabled="isSubmitting"
              autocomplete="new-password"
              @keyup.enter="handleSubmitRegistration"
            />
          </div>
          <div v-if="invitationRequired">
            <label class="input-label">{{ authText('invitationCodeLabel') }}</label>
            <input
              v-model="invitationCode"
              type="text"
              class="input w-full"
              :placeholder="authText('invitationCodePlaceholder')"
              :disabled="isSubmitting"
              @keyup.enter="handleSubmitRegistration"
            />
            <p v-if="invitationGateSatisfiedByAffiliate" class="mt-2 text-xs text-primary-700 dark:text-primary-300">
              {{ authText('affiliateInvitationDetected') }}
            </p>
          </div>
          <p v-if="registrationError" class="text-sm text-red-600 dark:text-red-400">
            {{ registrationError }}
          </p>
          <button
            class="btn btn-primary w-full"
            type="button"
            :disabled="isSubmitting || !canSubmitRegistration"
            @click="handleSubmitRegistration"
          >
            {{ isSubmitting ? authText('processing') : authText('oauthCallbackSubmitRegistration') }}
          </button>
        </div>
      </div>

      <div v-else-if="invalidCallback" class="card p-6 text-center">
        <h1 class="text-lg font-semibold text-gray-900 dark:text-white">
          {{ authText('oauthCallbackInvalidTitle') }}
        </h1>
        <p class="mt-2 text-sm text-gray-600 dark:text-gray-400">
          {{ authText('oauthCallbackInvalidHint') }}
        </p>
        <button class="btn btn-primary mt-6" type="button" @click="router.replace(authRouteDefaults.loginPath)">
          {{ authText('backToLogin') }}
        </button>
      </div>

      <div v-else class="card p-6">
        <h1 class="text-lg font-semibold text-gray-900 dark:text-white">
          {{ authText('oauthCallbackTitle') }}
        </h1>
        <p class="mt-2 text-sm text-gray-600 dark:text-gray-400">
          {{ authText('oauthCallbackHint') }}
        </p>

        <div class="mt-6 space-y-4">
          <div>
            <label class="input-label">{{ authText('oauthCallbackCode') }}</label>
            <div class="flex gap-2">
              <input class="input flex-1 font-mono text-sm" :value="code" readonly />
              <button class="btn btn-secondary" type="button" :disabled="!code" @click="copy(code)">
                {{ t('common.copy') }}
              </button>
            </div>
          </div>

          <div>
            <label class="input-label">{{ authText('oauthCallbackState') }}</label>
            <div class="flex gap-2">
              <input class="input flex-1 font-mono text-sm" :value="state" readonly />
              <button
                class="btn btn-secondary"
                type="button"
                :disabled="!state"
                @click="copy(state)"
              >
                {{ t('common.copy') }}
              </button>
            </div>
          </div>

          <div>
            <label class="input-label">{{ authText('oauthCallbackFullUrl') }}</label>
            <div class="flex gap-2">
              <input class="input flex-1 font-mono text-xs" :value="fullUrl" readonly />
              <button
                class="btn btn-secondary"
                type="button"
                :disabled="!fullUrl"
                @click="copy(fullUrl)"
              >
                {{ t('common.copy') }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { useClipboard } from '@/composables/useClipboard'
import { useAppStore, useAuthStore } from '@/stores'
import { apiClient, buildApiUrl } from '@/api/client'
import {
  exchangePendingOAuthCompletion,
  type OAuthTokenResponse
} from '@/api/auth'
import { useAuthShellText } from '@/composables/useAuthShellText'
import {
  loadOAuthAffiliateCode,
  oauthAffiliatePayload
} from '@/utils/oauthAffiliate'
import {
  deriveEmailOAuthRegistrationState,
  isEmailOAuthTokenResponse,
  resolveEmailOAuthProvider,
  type EmailOAuthPendingCompletion,
  type EmailOAuthProvider,
} from './emailOAuthFlow'
import { finalizeAuthLoginSuccess } from './finalizeAuthLogin'
import { buildLoginAgreementAcceptancePayload } from '@/utils/loginAgreementConsent'
import { DEFAULT_AUTH_REDIRECT_PATH, sanitizeAuthRedirectPath } from '@/utils/authRedirect'
import { resolvePasswordMinLength } from '@/utils/passwordPolicy'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const { authText, authRouteDefaults, defaultRedirectPath, loadAuthShellConfig } = useAuthShellText()
const { copyToClipboard } = useClipboard()
const appStore = useAppStore()
const authStore = useAuthStore()
const isProcessing = ref(false)
const isSubmitting = ref(false)
const needsRegistrationCompletion = ref(false)
const invitationRequired = ref(false)
const registrationEmail = ref('')
const password = ref('')
const confirmPassword = ref('')
const invitationCode = ref('')
const registrationError = ref('')
const pendingProvider = ref<EmailOAuthProvider>('github')
const redirectTo = ref(DEFAULT_AUTH_REDIRECT_PATH)
const invalidCallback = ref(false)
const EMAIL_OAUTH_PENDING_PROVIDER_KEY = 'email_oauth_pending_provider'

const code = computed(() => (route.query.code as string) || '')
const state = computed(() => (route.query.state as string) || '')
const error = computed(
  () => (route.query.error as string) || (route.query.error_description as string) || ''
)

const fullUrl = computed(() => {
  if (typeof window === 'undefined') return ''
  return window.location.href
})
const providerName = computed(() =>
  pendingProvider.value === 'google' ? 'Google' : 'GitHub'
)
const registrationHint = computed(() =>
  invitationRequired.value
    ? authText('oauthCallbackRegistrationInvitationRequired', { providerName: providerName.value })
    : authText('oauthCallbackRegistrationHint')
)
const passwordMinLength = computed(() =>
  resolvePasswordMinLength(appStore.cachedPublicSettings)
)
const passwordOptional = computed(() => pendingProvider.value === 'google')
const oauthAffiliateCodeForGate = computed(() => loadOAuthAffiliateCode())
const invitationGateSatisfiedByAffiliate = computed(
  () => invitationRequired.value && !invitationCode.value.trim() && !!oauthAffiliateCodeForGate.value
)
const canSubmitRegistration = computed(() => {
  if (!registrationEmail.value.trim()) return false
  if (invitationRequired.value && !invitationCode.value.trim() && !oauthAffiliateCodeForGate.value) return false
  if (!passwordOptional.value) {
    if (password.value.length < passwordMinLength.value) return false
    if (password.value !== confirmPassword.value) return false
    return true
  }
  if (!password.value && !confirmPassword.value) return true
  if (password.value.length < passwordMinLength.value) return false
  if (password.value !== confirmPassword.value) return false
  return true
})

function parseFragmentParams(): URLSearchParams {
  const raw = typeof window !== 'undefined' ? window.location.hash : ''
  const hash = raw.startsWith('#') ? raw.slice(1) : raw
  return new URLSearchParams(hash)
}

function readTokenResponse(params: URLSearchParams): OAuthTokenResponse | null {
  const accessToken = params.get('access_token')?.trim() || ''
  if (!accessToken) return null

  const response: OAuthTokenResponse = { access_token: accessToken }
  const refreshToken = params.get('refresh_token')?.trim() || ''
  if (refreshToken) response.refresh_token = refreshToken
  const expiresIn = Number.parseInt(params.get('expires_in')?.trim() || '', 10)
  if (Number.isFinite(expiresIn) && expiresIn > 0) response.expires_in = expiresIn
  const tokenType = params.get('token_type')?.trim() || ''
  if (tokenType) response.token_type = tokenType
  return response
}

function readPendingEmailOAuthProvider(): EmailOAuthProvider | null {
  if (typeof window === 'undefined') return null
  return resolveEmailOAuthProvider(window.sessionStorage.getItem(EMAIL_OAUTH_PENDING_PROVIDER_KEY))
}

function redirectProviderCallbackToBackend(provider: EmailOAuthProvider): void {
  if (typeof window === 'undefined') return
  const params = new URLSearchParams()
  for (const [key, value] of Object.entries(route.query)) {
    if (Array.isArray(value)) {
      value.forEach((item) => {
        if (item != null) params.append(key, String(item))
      })
    } else if (value != null) {
      params.set(key, String(value))
    }
  }
  const suffix = params.toString() ? `?${params.toString()}` : ''
  window.location.href = buildApiUrl(`/auth/oauth/${provider}/callback${suffix}`)
}

async function finalizeTokenResponse(tokenResponse: OAuthTokenResponse, redirect: string) {
  await finalizeAuthLoginSuccess({
    tokenResponse,
    redirect: sanitizeAuthRedirectPath(redirect, defaultRedirectPath.value),
    authStore,
    appStore,
    router,
    successMessage: t('auth.loginSuccess'),
    beforeRedirect: () => {
      if (typeof window !== 'undefined') {
        window.sessionStorage.removeItem(EMAIL_OAUTH_PENDING_PROVIDER_KEY)
      }
    },
  })
}

async function resumePendingEmailOAuth() {
  isProcessing.value = true
  try {
    const completion = await exchangePendingOAuthCompletion() as EmailOAuthPendingCompletion
    if (isEmailOAuthTokenResponse(completion)) {
      await finalizeTokenResponse(completion as OAuthTokenResponse, completion.redirect || defaultRedirectPath.value)
      return
    }

    const nextState = deriveEmailOAuthRegistrationState(completion, defaultRedirectPath.value)
    if (nextState.provider) {
      pendingProvider.value = nextState.provider
    }
    redirectTo.value = sanitizeAuthRedirectPath(nextState.redirect, defaultRedirectPath.value)

    if (nextState.requiresRegistrationCompletion) {
      invitationRequired.value = nextState.invitationRequired
      registrationEmail.value = nextState.registrationEmail
      needsRegistrationCompletion.value = true
      isProcessing.value = false
      return
    }

    appStore.showError(completion.error || t('auth.loginFailed'))
  } catch (e: unknown) {
    const err = e as { message?: string; response?: { data?: { message?: string } } }
    const message = err.response?.data?.message || err.message || t('auth.loginFailed')
    appStore.showError(message)
    invalidCallback.value = true
  } finally {
    if (!needsRegistrationCompletion.value) {
      isProcessing.value = false
    }
  }
}

async function handleSubmitRegistration() {
  registrationError.value = ''
  if (!registrationEmail.value.trim()) {
    registrationError.value = t('auth.emailRequired')
    return
  }
  const needsPasswordValidation = !passwordOptional.value || !!password.value || !!confirmPassword.value
  if (needsPasswordValidation) {
    if (password.value.length < passwordMinLength.value) {
      registrationError.value = t('auth.passwordMinLength', { count: passwordMinLength.value })
      return
    }
    if (password.value !== confirmPassword.value) {
      registrationError.value = t('auth.passwordsDoNotMatch')
      return
    }
  }
  const code = invitationCode.value.trim()
  const affCode = loadOAuthAffiliateCode()
  if (invitationRequired.value && !code && !affCode) return

  isSubmitting.value = true
  try {
    const payload: { password?: string; invitation_code?: string; aff_code?: string } = {
      ...oauthAffiliatePayload(affCode)
    }
    if (password.value.trim()) {
      payload.password = password.value
    }
    if (invitationRequired.value) {
      payload.invitation_code = code
    }
    const { data } = await apiClient.post<OAuthTokenResponse>(
      `/auth/oauth/${pendingProvider.value}/complete-registration`,
      {
        ...payload,
        ...buildLoginAgreementAcceptancePayload()
      }
    )
    await finalizeTokenResponse(data, redirectTo.value)
  } catch (e: unknown) {
    const err = e as { message?: string; response?: { data?: { message?: string } } }
    registrationError.value =
      err.response?.data?.message || err.message || t('auth.oidc.completeRegistrationFailed')
  } finally {
    isSubmitting.value = false
  }
}

onMounted(async () => {
  await loadAuthShellConfig()
  redirectTo.value = defaultRedirectPath.value

  const params = parseFragmentParams()
  const tokenResponse = readTokenResponse(params)
  const fragmentError = params.get('error') || ''
  const fragmentErrorDescription =
    params.get('error_description') || params.get('error_message') || ''

  if (fragmentError) {
    appStore.showError(fragmentErrorDescription || fragmentError)
    return
  }
  if (!tokenResponse) {
    if (route.path === '/auth/oauth/callback') {
      const pendingEmailOAuthProvider = readPendingEmailOAuthProvider()
      if (pendingEmailOAuthProvider && code.value && state.value) {
        redirectProviderCallbackToBackend(pendingEmailOAuthProvider)
        return
      }
      await resumePendingEmailOAuth()
    }
    return
  }

  isProcessing.value = true
  try {
    await finalizeTokenResponse(tokenResponse, params.get('redirect') || defaultRedirectPath.value)
  } catch (error: unknown) {
    const message = (error as { message?: string })?.message || t('auth.loginFailed')
    appStore.showError(message)
    isProcessing.value = false
  }
})

watch(
  error,
  (message) => {
    if (message) {
      appStore.showError(message)
    }
  },
  { immediate: true }
)

const copy = (value: string) => {
  if (!value) return
  copyToClipboard(value)
}
</script>
