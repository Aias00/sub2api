<template>
  <div class="vercel-auth-shell min-h-screen bg-gray-50 px-4 py-10 dark:bg-dark-900">
    <div class="mx-auto max-w-2xl">
      <div class="card p-6">
        <h1 class="text-lg font-semibold text-gray-900 dark:text-white">
          {{ callbackTitleText }}
        </h1>
        <p class="mt-2 text-sm text-gray-600 dark:text-gray-400">
          {{ errorMessage || callbackProcessingText }}
        </p>

        <div
          v-if="!errorMessage"
          class="mt-6 flex items-center justify-center py-10"
        >
          <div
            class="h-8 w-8 animate-spin rounded-full border-4 border-primary-500 border-t-transparent"
          ></div>
        </div>

        <div
          v-else
          class="mt-6 rounded-lg border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-800/80"
        >
          <p class="text-sm text-gray-700 dark:text-gray-300">
            {{ errorMessage }}
          </p>
          <button
            class="btn btn-primary mt-4"
            type="button"
            @click="goBackToPayment"
          >
            {{ backToPaymentText }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { useAppStore } from '@/stores'
import { useAuthRouteDefaults } from '@/composables/useAuthRouteDefaults'
import {
  renderWechatPaymentCallbackText,
  resolveWechatPaymentCallbackLabels,
} from '@/utils/paymentShell'
import { resolveRuntimeLocale } from '@/utils/runtimeLocale'

const { locale } = useI18n()
const route = useRoute()
const router = useRouter()
const appStore = useAppStore()
const { authRouteDefaults } = useAuthRouteDefaults()

const errorMessage = ref('')

watch(errorMessage, (message) => {
  if (message) {
    appStore.showError(message)
  }
})

const paymentShellLabels = computed(() =>
  resolveWechatPaymentCallbackLabels(
    appStore.cachedPublicSettings?.payment_shell_config,
    resolveRuntimeLocale(locale),
  ),
)

function paymentText(key: Parameters<typeof renderWechatPaymentCallbackText>[1]): string {
  return renderWechatPaymentCallbackText(paymentShellLabels.value, key)
}

const callbackProcessingText = computed(() => paymentText('wechatPaymentCallbackProcessing'))
const callbackTitleText = computed(() => paymentText('wechatPaymentCallbackTitle'))
const backToPaymentText = computed(() => paymentText('wechatPaymentCallbackBackToPayment'))

function readQueryString(key: string): string {
  const value = route.query[key]
  if (Array.isArray(value)) {
    return typeof value[0] === 'string' ? value[0] : ''
  }
  return typeof value === 'string' ? value : ''
}

function parseFragmentParams(): URLSearchParams {
  const raw = typeof window !== 'undefined' ? window.location.hash : ''
  const hash = raw.startsWith('#') ? raw.slice(1) : raw
  return new URLSearchParams(hash)
}

function normalizeRedirectPath(path: string | null | undefined): string {
  const fallbackPath = authRouteDefaults.value.purchasePath
  const value = (path || '').trim()
  if (!value) return fallbackPath
  if (!value.startsWith('/')) return fallbackPath
  if (value.startsWith('//') || value.includes('://')) return fallbackPath
  if (value === '/payment') return fallbackPath
  if (value.startsWith('/payment?')) return mergeRedirectQuery(fallbackPath, value.slice('/payment'.length))
  return value
}

function mergeRedirectQuery(basePath: string, querySuffix: string): string {
  const url = new URL(basePath, window.location.origin)
  const query = new URLSearchParams(querySuffix.startsWith('?') ? querySuffix.slice(1) : querySuffix)
  query.forEach((value, key) => {
    url.searchParams.set(key, value)
  })
  return url.pathname + url.search + url.hash
}

function goBackToPayment() {
  void router.replace(authRouteDefaults.value.purchasePath)
}

onMounted(async () => {
  const fragment = parseFragmentParams()
  const readParam = (key: string) => fragment.get(key) || readQueryString(key)

  const error = readParam('error') || readParam('err_msg') || readParam('errmsg')
  const errorDescription = readParam('error_description') || readParam('message')

  if (error) {
    errorMessage.value = errorDescription || error
    return
  }

  const resumeToken = readParam('wechat_resume_token')
  const redirectURL = new URL(
    normalizeRedirectPath(readParam('redirect')),
    window.location.origin,
  )

  if (!resumeToken) {
    errorMessage.value = paymentText('wechatPaymentCallbackMissingResumeToken')
    return
  }

  const query: Record<string, string> = {
    ...Object.fromEntries(redirectURL.searchParams.entries()),
    wechat_resume: '1',
    wechat_resume_token: resumeToken,
  }

  await router.replace({
    path: redirectURL.pathname,
    query,
  })
})
</script>
