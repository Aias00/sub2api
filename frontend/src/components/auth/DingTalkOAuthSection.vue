<template>
  <div class="space-y-4">
    <button type="button" :disabled="disabled" class="btn btn-secondary w-full" @click="startLogin">
      <svg
        class="icon mr-2"
        viewBox="0 0 24 24"
        xmlns="http://www.w3.org/2000/svg"
        width="20"
        height="20"
        aria-hidden="true"
        style="flex-shrink: 0"
      >
        <circle cx="12" cy="12" r="12" fill="#1677FF" />
        <text
          x="12"
          y="17"
          font-family="sans-serif"
          font-size="13"
          font-weight="bold"
          fill="white"
          text-anchor="middle"
        >D</text>
      </svg>
      {{ authText('signInWithProvider', { providerName }) }}
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
import { useRoute } from 'vue-router'
import { buildApiUrl } from '@/api/client'
import { renderAuthShellText, type AuthShellLabelKey, type AuthShellLabels } from '@/utils/authShell'
import { resolveRouteAuthRedirect } from '@/utils/authRedirect'
import { resolveAffiliateReferralCode, storeOAuthAffiliateCode } from '@/utils/oauthAffiliate'

const props = withDefaults(defineProps<{
  disabled?: boolean
  affCode?: string
  showDivider?: boolean
  shellLabels?: AuthShellLabels
}>(), {
  showDivider: true
})

const route = useRoute()
const providerName = 'DingTalk'

function authText(key: AuthShellLabelKey, params: Record<string, string | number> = {}): string {
  return renderAuthShellText(props.shellLabels || {}, key, params)
}

function startLogin(): void {
  const redirectTo = resolveRouteAuthRedirect(route.query.redirect)
  storeOAuthAffiliateCode(resolveAffiliateReferralCode(props.affCode, route.query.aff, route.query.aff_code))
  const startURL = buildApiUrl(`/auth/oauth/dingtalk/start?redirect=${encodeURIComponent(redirectTo)}`)
  window.location.href = startURL
}
</script>
