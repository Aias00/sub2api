<template>
  <div class="vercel-auth-shell flex min-h-screen items-center justify-center bg-slate-50 p-4 dark:bg-slate-950">
    <div
      class="w-full max-w-md space-y-4 rounded-2xl border border-slate-200 bg-white p-6 shadow-lg dark:border-slate-700 dark:bg-slate-900"
    >
      <!-- Amount + Order ID -->
      <div v-if="amount" class="text-center">
        <p class="text-3xl font-bold" :style="{ color: methodColor }">{{ displayAmount }}</p>
        <p v-if="orderId" class="mt-1 text-sm text-gray-500 dark:text-slate-400">
          {{ paymentText('orderId') }}: {{ orderId }}
        </p>
      </div>

      <!-- Error -->
      <div v-if="error" class="space-y-3">
        <div
          class="rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-600 dark:border-red-700 dark:bg-red-900/30 dark:text-red-400"
        >
          {{ error }}
        </div>
        <button
          class="w-full text-sm underline dark:text-blue-400 dark:hover:text-blue-300"
          :style="{ color: methodColor }"
          @click="closeWindow"
        >
          {{ paymentText('close') }}
        </button>
      </div>

      <!-- Success -->
      <div v-else-if="success" class="space-y-3 py-4 text-center">
        <div class="text-5xl text-green-600 dark:text-green-400">✓</div>
        <p class="text-sm text-gray-500 dark:text-slate-400">{{ paymentText('success') }}</p>
        <button
          class="text-sm underline dark:text-blue-400 dark:hover:text-blue-300"
          :style="{ color: methodColor }"
          @click="closeWindow"
        >
          {{ paymentText('close') }}
        </button>
      </div>

      <!-- Loading / Redirecting -->
      <div v-else class="flex items-center justify-center py-8">
        <div
          class="h-8 w-8 animate-spin rounded-full border-2 border-t-transparent"
          :style="{ borderColor: methodColor, borderTopColor: 'transparent' }"
        />
        <span class="ml-3 text-sm text-gray-500 dark:text-slate-400">{{ hint }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { resolveRuntimeLocale } from '@/utils/runtimeLocale'
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { useAppStore } from '@/stores'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { getStripePaymentMethodColor } from '@/components/payment/paymentMethod'
import { isMobileDevice } from '@/utils/device'
import { buildApiUrl } from '@/api/client'

interface StripeWithWechatPay {
  confirmWechatPayPayment(clientSecret: string, options: Record<string, unknown>): Promise<{ error?: { message?: string }; paymentIntent?: { status: string } }>
}

const { t, locale } = useI18n()
const route = useRoute()
const appStore = useAppStore()
const { authRouteDefaults } = useAuthRouteDefaults()

const routeState = resolveStripePopupRouteState(route.query as Record<string, unknown>)
const orderId = routeState.orderId
const method = routeState.method
const amount = routeState.amount
const currency = routeState.currency

const methodColor = computed(() => getStripePaymentMethodColor(method))
const displayAmount = computed(() => formatStripePopupDisplayAmount(amount, currency))

const error = ref('')
const success = ref(false)
const hint = ref('')

let pollTimer: ReturnType<typeof setInterval> | null = null


const stripePopupLabels = computed(() =>
  resolveStripePopupLabels(
    appStore.cachedPublicSettings?.payment_shell_config,
    resolveRuntimeLocale(locale),
  ),
)
const stripeRuntimeDefaults = computed(() =>
  resolveStripePaymentRuntimeDefaults(
    appStore.cachedPublicSettings?.payment_shell_config,
    resolveRuntimeLocale(locale),
  ),
)

function paymentText(key: StripePopupLabelKey): string {
  return renderStripePopupText(stripePopupLabels.value, key)
}

hint.value = paymentText('stripePopupRedirecting')

function closeWindow() { window.close() }

onMounted(() => {
  const handler = (event: MessageEvent) => {
    if (event.origin !== window.location.origin) return
    if (event.data?.type !== 'STRIPE_POPUP_INIT') return
    window.removeEventListener('message', handler)
    initStripe(event.data.clientSecret, event.data.publishableKey)
  }
  window.addEventListener('message', handler)

  if (window.opener) {
    window.opener.postMessage({ type: 'STRIPE_POPUP_READY' }, window.location.origin)
  }

  setTimeout(() => {
    if (!error.value && !success.value) {
      error.value = paymentText('stripePopupTimeout')
    }
  }, stripeRuntimeDefaults.value.popupInitTimeoutMs)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})

async function initStripe(clientSecret: string, publishableKey: string) {
  if (!clientSecret || !publishableKey || !method) {
    error.value = paymentText('stripeMissingParams')
    return
  }
  try {
    const { loadStripe } = await import('@stripe/stripe-js')
    const stripe = await loadStripe(publishableKey)
    if (!stripe) { error.value = paymentText('stripeLoadFailed'); return }

    const returnUrl = paymentResultReturnURL()

    if (method === 'alipay') {
      // Alipay: redirect this popup to Alipay payment page
      const { error: err } = await stripe.confirmAlipayPayment(clientSecret, { return_url: returnUrl })
      if (err) error.value = err.message || paymentText('failed')
    } else if (method === 'wechat_pay') {
      // WeChat: Stripe shows its built-in QR dialog, user scans, promise resolves
      hint.value = paymentText('stripePopupLoadingQr')
      const result = await (stripe as unknown as StripeWithWechatPay).confirmWechatPayPayment(clientSecret, {
        payment_method_options: { wechat_pay: { client: isMobileDevice() ? 'mobile_web' : 'web' } },
      })
      if (result.error) {
        error.value = result.error.message || paymentText('failed')
      } else if (result.paymentIntent?.status === 'succeeded') {
        success.value = true
        setTimeout(closeWindow, stripeRuntimeDefaults.value.closeDelayMs)
      } else {
        // Payment not completed (user closed QR dialog)
        startPolling()
      }
    }
  } catch (err: unknown) {
    error.value = extractI18nErrorMessage(err, t, 'payment.errors', paymentText('stripeLoadFailed'))
  }
}

function startPolling() {
  pollTimer = setInterval(async () => {
    try {
      const token = document.cookie.split('; ').find(c => c.startsWith('token='))?.split('=')[1]
        || localStorage.getItem('token') || ''
      const res = await fetch(buildApiUrl(`/payment/orders/${orderId}`), {
        headers: token ? { Authorization: 'Bearer ' + token } : {},
        credentials: 'include',
      })
      if (!res.ok) return
      const data = await res.json()
      const status = data?.data?.status
      if (status === 'COMPLETED' || status === 'PAID') {
        if (pollTimer) { clearInterval(pollTimer); pollTimer = null }
        success.value = true
        setTimeout(closeWindow, stripeRuntimeDefaults.value.closeDelayMs)
      }
    } catch { /* ignore */ }
  }, stripeRuntimeDefaults.value.pollIntervalMs)
}

function paymentResultReturnURL(): string {
  return buildStripePopupPaymentResultReturnUrl(
    authRouteDefaults.value.paymentResultPath,
    orderId,
    window.location.origin,
  )
}
</script>
