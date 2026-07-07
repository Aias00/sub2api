<template>
  <AppLayout>
    <div class="mx-auto flex max-w-md flex-col items-center space-y-6 py-8">
      <h2 class="text-xl font-semibold text-gray-900 dark:text-white">
        {{ qrUrl ? scanTitle : paymentText('payInNewWindow') }}
      </h2>
      <div v-if="qrUrl" class="rounded-2xl bg-white p-6 shadow-lg dark:bg-dark-800">
        <canvas ref="qrCanvas" class="mx-auto"></canvas>
      </div>
      <!-- Scan prompt for QR code -->
      <p v-if="qrUrl && !expired && scanHint" class="text-center text-sm text-gray-500 dark:text-gray-400">
        {{ scanHint }}
      </p>
      <div v-if="expired" class="text-center">
        <p class="text-lg font-medium text-red-500">{{ paymentText('expired') }}</p>
        <button class="btn btn-primary mt-4" @click="router.push(authRouteDefaults.purchasePath)">{{ paymentText('backToRecharge') }}</button>
      </div>
      <div v-else class="text-center">
        <p class="text-sm text-gray-500 dark:text-gray-400">{{ qrUrl ? paymentText('expiresIn') : paymentText('payInNewWindowHint') }}</p>
        <p class="mt-1 text-2xl font-bold tabular-nums text-gray-900 dark:text-white">{{ countdownDisplay }}</p>
        <p class="mt-2 text-sm text-gray-400 dark:text-gray-500">{{ paymentText('waitingPayment') }}</p>
      </div>
      <a v-if="payUrl && !qrUrl && !expired" :href="payUrl" target="_blank" rel="noopener noreferrer"
        class="btn btn-primary w-full py-3">
        {{ paymentText('openPayWindow') }}
      </a>
      <!-- Cancel button -->
      <button v-if="!expired && orderId" class="btn btn-secondary w-full" :disabled="cancelling" @click="handleCancel">
        {{ cancelling ? paymentText('processing') : paymentText('cancelOrder') }}
      </button>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { resolveRuntimeLocale } from '@/utils/runtimeLocale'
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import { usePaymentStore } from '@/stores/payment'
import { paymentAPI } from '@/api/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import {
  renderPaymentQRText,
  resolvePaymentQRLabels,
  resolvePaymentStatusPollingDefaults,
  type PaymentQRLabelKey,
} from '@/utils/paymentShell'
import { useAuthRouteDefaults } from '@/composables/useAuthRouteDefaults'
import { useAppStore } from '@/stores'
import { isBuiltInAlipayMethod, isBuiltInWxpayMethod } from '@/components/payment/providerConfig'
import QRCode from 'qrcode'
import alipayIcon from '@/assets/icons/alipay.svg'
import wxpayIcon from '@/assets/icons/wxpay.svg'
import {
  formatPaymentQrCountdown,
  isPaymentQrCompleted,
  isPaymentQrTerminal,
  resolvePaymentQrRouteState,
  resolvePaymentQrSecondsUntil,
} from './paymentQrRuntime'

const { t, locale } = useI18n()
const route = useRoute()
const router = useRouter()
const paymentStore = usePaymentStore()
const appStore = useAppStore()
const { authRouteDefaults } = useAuthRouteDefaults()

const qrCanvas = ref<HTMLCanvasElement | null>(null)
const qrUrl = ref('')
const payUrl = ref('')
const orderId = ref(0)
const remainingSeconds = ref(0)
const expired = ref(false)
const cancelling = ref(false)
const paymentType = ref('')

let pollTimer: ReturnType<typeof setInterval> | null = null
let countdownTimer: ReturnType<typeof setInterval> | null = null

const countdownDisplay = computed(() => formatPaymentQrCountdown(remainingSeconds.value))

const isAlipay = computed(() => isBuiltInAlipayMethod(paymentType.value))
const isWxpay = computed(() => isBuiltInWxpayMethod(paymentType.value))


const paymentQRLabels = computed(() =>
  resolvePaymentQRLabels(
    appStore.cachedPublicSettings?.payment_shell_config,
    resolveRuntimeLocale(locale),
  ),
)
const paymentStatusPollingDefaults = computed(() =>
  resolvePaymentStatusPollingDefaults(
    appStore.cachedPublicSettings?.payment_shell_config,
    resolveRuntimeLocale(locale),
  ),
)

function paymentText(key: PaymentQRLabelKey): string {
  return renderPaymentQRText(paymentQRLabels.value, key)
}

const scanTitle = computed(() => {
  if (isAlipay.value) return paymentText('scanAlipay')
  if (isWxpay.value) return paymentText('scanWxpay')
  return paymentText('scanToPay')
})

const scanHint = computed(() => {
  if (isAlipay.value) return paymentText('scanAlipayHint')
  if (isWxpay.value) return paymentText('scanWxpayHint')
  return ''
})

function getLogoForType(): string | null {
  if (isAlipay.value) return alipayIcon
  if (isWxpay.value) return wxpayIcon
  return null
}

async function renderQR() {
  await nextTick()
  if (!qrCanvas.value || !qrUrl.value) return

  // Use medium error correction to support logo overlay while keeping QR code scannable
  const logoSrc = getLogoForType()
  await QRCode.toCanvas(qrCanvas.value, qrUrl.value, {
    width: 256,
    margin: 2,
    errorCorrectionLevel: logoSrc ? 'M' : 'L',
  })

  if (!logoSrc) return

  // Draw logo in center of QR code
  const canvas = qrCanvas.value
  const ctx = canvas.getContext('2d')
  if (!ctx) return

  const img = new Image()
  img.src = logoSrc
  img.onload = () => {
    const logoSize = 48
    const x = (canvas.width - logoSize) / 2
    const y = (canvas.height - logoSize) / 2
    // White background with rounded corners
    const pad = 5
    ctx.fillStyle = '#FFFFFF'
    ctx.beginPath()
    const r = 6
    ctx.moveTo(x - pad + r, y - pad)
    ctx.arcTo(x + logoSize + pad, y - pad, x + logoSize + pad, y + logoSize + pad, r)
    ctx.arcTo(x + logoSize + pad, y + logoSize + pad, x - pad, y + logoSize + pad, r)
    ctx.arcTo(x - pad, y + logoSize + pad, x - pad, y - pad, r)
    ctx.arcTo(x - pad, y - pad, x + logoSize + pad, y - pad, r)
    ctx.fill()
    // Draw logo
    ctx.drawImage(img, x, y, logoSize, logoSize)
  }
}

async function pollStatus() {
  if (!orderId.value) return
  const order = await paymentStore.pollOrderStatus(orderId.value)
  if (!order) return
  if (isPaymentQrCompleted(order.status)) {
    cleanup()
    router.push({ path: authRouteDefaults.value.paymentResultPath, query: { order_id: String(orderId.value), status: 'success' } })
  } else if (isPaymentQrTerminal(order.status)) {
    cleanup()
    expired.value = true
  }
}

function startCountdown(seconds: number) {
  remainingSeconds.value = Math.max(0, seconds)
  if (remainingSeconds.value <= 0) {
    expired.value = true
    return
  }
  countdownTimer = setInterval(() => {
    remainingSeconds.value--
    if (remainingSeconds.value <= 0) {
      expired.value = true
      cleanup()
    }
  }, 1000)
}

async function handleCancel() {
  if (!orderId.value || cancelling.value) return
  cancelling.value = true
  try {
    await paymentAPI.cancelOrder(orderId.value)
    cleanup()
    router.push(authRouteDefaults.value.purchasePath)
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', paymentText('errorFallback')))
  } finally {
    cancelling.value = false
  }
}

function cleanup() {
  if (pollTimer) { clearInterval(pollTimer); pollTimer = null }
  if (countdownTimer) { clearInterval(countdownTimer); countdownTimer = null }
}

watch(qrUrl, () => renderQR())

onMounted(() => {
  const routeState = resolvePaymentQrRouteState(route.query)
  orderId.value = routeState.orderId
  qrUrl.value = routeState.qrUrl
  payUrl.value = routeState.payUrl
  paymentType.value = routeState.paymentType

  // Calculate countdown from expiresAt
  const seconds = resolvePaymentQrSecondsUntil(routeState.expiresAt)
  startCountdown(seconds)
  if (!expired.value) {
    pollTimer = setInterval(pollStatus, paymentStatusPollingDefaults.value.pollIntervalMs)
  }
  renderQR()
})

onUnmounted(() => cleanup())
</script>
