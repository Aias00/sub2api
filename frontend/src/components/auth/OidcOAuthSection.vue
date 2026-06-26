<template>
  <div class="space-y-4">
    <button type="button" :disabled="disabled" class="btn btn-secondary w-full" @click="startLogin">
      <span
        class="mr-2 inline-flex h-5 w-5 items-center justify-center rounded-full bg-primary-100 text-xs font-semibold text-primary-700 dark:bg-primary-900/30 dark:text-primary-300"
      >
        {{ providerInitial }}
      </span>
      {{ authText('signInWithProvider', { providerName: normalizedProviderName }) }}
    </button>

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
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { buildApiUrl } from '@/api/client'
import { renderAuthShellText, type AuthShellLabelKey, type AuthShellLabels } from '@/utils/authShell'
import { resolveRouteAuthRedirect } from '@/utils/authRedirect'
import { resolveAffiliateReferralCode, storeOAuthAffiliateCode } from '@/utils/oauthAffiliate'

const props = withDefaults(defineProps<{
  disabled?: boolean
  affCode?: string
  providerName?: string
  agreementRevision?: string
  turnstileToken?: string
  showDivider?: boolean
  shellLabels?: AuthShellLabels
}>(), {
  providerName: '',
  showDivider: true
})

const route = useRoute()

const normalizedProviderName = computed(() => {
  const name = props.providerName?.trim()
  return name || ''
})

const providerInitial = computed(() => normalizedProviderName.value.charAt(0).toUpperCase())

function authText(key: AuthShellLabelKey, params: Record<string, string | number> = {}): string {
  return renderAuthShellText(props.shellLabels || {}, key, params)
}

function startLogin(): void {
  const redirectTo = resolveRouteAuthRedirect(route.query.redirect)
  storeOAuthAffiliateCode(resolveAffiliateReferralCode(props.affCode, route.query.aff, route.query.aff_code))
  const params = new URLSearchParams({ redirect: redirectTo })
  const agreementRevision = String(props.agreementRevision || '').trim()
  if (agreementRevision) {
    params.set('agreement_revision', agreementRevision)
  }
  const turnstileToken = String(props.turnstileToken || '').trim()
  if (turnstileToken) {
    params.set('turnstile_token', turnstileToken)
  }
  const startURL = buildApiUrl(`/auth/oauth/oidc/start?${params.toString()}`)
  window.location.href = startURL
}
</script>
