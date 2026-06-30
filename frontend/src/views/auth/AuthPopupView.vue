<template>
  <div class="vercel-auth-shell min-h-screen bg-gray-50 px-4 py-10 dark:bg-dark-900">
    <div class="mx-auto max-w-sm">
      <div class="card p-6 text-center">
        <div class="mx-auto h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
        <p class="mt-4 text-sm text-gray-600 dark:text-gray-400">
          {{ authText('oauthCallbackHint') }}
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthShellText } from '@/composables/useAuthShellText'

const route = useRoute()
const router = useRouter()
const { authText, authRouteDefaults, loadAuthShellConfig } = useAuthShellText()

function safeInternalRedirect(value: unknown): string {
  if (typeof value !== 'string') return ''
  if (!value.startsWith('/') || value.startsWith('//')) return ''
  if (value.includes('://') || value.includes('\n') || value.includes('\r')) return ''
  return value
}

onMounted(async () => {
  await loadAuthShellConfig()
  const redirect = safeInternalRedirect(route.query.callbackUrl)
  router.replace({
    path: authRouteDefaults.value.loginPath,
    query: redirect ? { redirect } : undefined,
  })
})
</script>
