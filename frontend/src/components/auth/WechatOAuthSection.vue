<template>
  <div class="space-y-4">
    <button type="button" :disabled="buttonDisabled" class="btn btn-secondary w-full" @click="startLogin">
      <span
        class="mr-2 inline-flex h-5 w-5 items-center justify-center rounded-full bg-green-100 text-xs font-semibold text-green-700 dark:bg-green-900/30 dark:text-green-300"
      >
        W
      </span>
      {{ authText('signInWithProvider', { providerName }) }}
    </button>

    <p
      v-if="disabledHint"
      data-testid="wechat-oauth-hint"
      class="text-sm text-amber-600 dark:text-amber-400"
    >
      {{ disabledHint }}
    </p>

    <div v-if="showDivider" class="flex items-center gap-3">
      <div class="h-px flex-1 bg-gray-200 dark:bg-dark-700"></div>
      <span class="text-xs text-gray-500 dark:text-dark-400">
        {{ authText('oauthAlternativeMethods') }}
      </span>
      <div class="h-px flex-1 bg-gray-200 dark:bg-dark-700"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { resolveWeChatOAuthStart } from '@/api/auth'
import { buildApiUrl } from '@/api/client'
import { useAppStore } from '@/stores'
import { renderAuthShellText, type AuthShellLabelKey, type AuthShellLabels } from '@/utils/authShell'
import { resolveRouteAuthRedirect } from '@/utils/authRedirect'
import { resolveAffiliateReferralCode, storeOAuthAffiliateCode } from '@/utils/oauthAffiliate'

const props = withDefaults(defineProps<{
  disabled?: boolean
  affCode?: string
  agreementRevision?: string
  turnstileToken?: string
  showDivider?: boolean
  shellLabels?: AuthShellLabels
}>(), {
  showDivider: true,
})

const appStore = useAppStore()
const route = useRoute()

function authText(key: AuthShellLabelKey, params: Record<string, string | number> = {}): string {
  return renderAuthShellText(props.shellLabels || {}, key, params)
}

const providerName = computed(() => authText('wechatProviderName'))
const resolvedStart = computed(() => resolveWeChatOAuthStart(appStore.cachedPublicSettings))
const buttonDisabled = computed(() => props.disabled || resolvedStart.value.mode === null)
const disabledHint = computed(() => {
  if (props.disabled) {
    return ''
  }
  switch (resolvedStart.value.unavailableReason) {
    case 'external_browser_required':
      return authText('wechatSystemBrowserOnly')
    case 'wechat_browser_required':
      return authText('wechatBrowserOnly')
    case 'native_app_required':
      return authText('wechatNativeAppOnly')
    case 'not_configured':
      return authText('wechatNotConfigured')
    default:
      return ''
  }
})

onMounted(() => {
  if (!appStore.cachedPublicSettings && !appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})

function startLogin(): void {
  if (buttonDisabled.value || !resolvedStart.value.mode) {
    return
  }
  const redirectTo = resolveRouteAuthRedirect(route.query.redirect)
  storeOAuthAffiliateCode(resolveAffiliateReferralCode(props.affCode, route.query.aff, route.query.aff_code))
  const mode = resolvedStart.value.mode
  const params = new URLSearchParams({
    mode,
    redirect: redirectTo,
  })
  const agreementRevision = String(props.agreementRevision || '').trim()
  if (agreementRevision) {
    params.set('agreement_revision', agreementRevision)
  }
  const turnstileToken = String(props.turnstileToken || '').trim()
  if (turnstileToken) {
    params.set('turnstile_token', turnstileToken)
  }
  const startURL = buildApiUrl(`/auth/oauth/wechat/start?${params.toString()}`, appStore.cachedPublicSettings)
  window.location.href = startURL
}
</script>
