<template>
  <div class="flex min-h-screen items-center justify-center bg-gray-50 px-4 dark:bg-dark-900">
    <div class="w-full max-w-md space-y-6">
      <!-- Loading -->
      <div v-if="loading" class="flex items-center justify-center py-20">
        <div class="h-8 w-8 animate-spin rounded-full border-4 border-primary-500 border-t-transparent"></div>
      </div>
      <template v-else>
        <!-- Status Icon -->
        <div class="text-center">
          <div v-if="isSuccess"
            class="mx-auto flex h-20 w-20 items-center justify-center rounded-full bg-green-100 dark:bg-green-900/30">
            <svg class="h-10 w-10 text-green-500" fill="none" viewBox="0 0 24 24" stroke="currentColor"
              stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <div v-else-if="isPending"
            class="mx-auto flex h-20 w-20 items-center justify-center rounded-full bg-yellow-100 dark:bg-yellow-900/30">
            <div class="h-10 w-10 animate-spin rounded-full border-4 border-yellow-500 border-t-transparent"></div>
          </div>
          <div v-else
            class="mx-auto flex h-20 w-20 items-center justify-center rounded-full bg-red-100 dark:bg-red-900/30">
            <svg class="h-10 w-10 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </div>
          <h2 class="mt-4 text-2xl font-bold text-gray-900 dark:text-white">
            {{ statusTitle }}
          </h2>
          <p v-if="isPending" class="mt-2 text-sm text-gray-500 dark:text-gray-400">
            {{ resultText('processingHint') }}
          </p>
        </div>
        <!-- Order Info -->
        <div v-if="order" class="rounded-xl bg-white p-5 shadow-sm dark:bg-dark-800">
          <div class="space-y-3 text-sm">
            <div class="flex justify-between">
              <span class="text-gray-500 dark:text-gray-400">{{ resultText('orderId') }}</span>
              <span class="font-medium text-gray-900 dark:text-white">#{{ order.id }}</span>
            </div>
            <div v-if="order.out_trade_no" class="flex justify-between">
              <span class="text-gray-500 dark:text-gray-400">{{ resultText('orderNo') }}</span>
              <span class="font-medium text-gray-900 dark:text-white">{{ order.out_trade_no }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-500 dark:text-gray-400">{{ resultText('baseAmount') }}</span>
              <span class="font-medium text-gray-900 dark:text-white">{{ formatGatewayAmount(baseAmount) }}</span>
            </div>
            <div v-if="order.fee_rate > 0" class="flex justify-between">
              <span class="text-gray-500 dark:text-gray-400">{{ resultText('fee') }} ({{ order.fee_rate }}%)</span>
              <span class="font-medium text-gray-900 dark:text-white">{{ formatGatewayAmount(feeAmount) }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-500 dark:text-gray-400">{{ resultText('payAmount') }}</span>
              <span class="font-bold text-primary-600 dark:text-primary-400">{{ formatGatewayAmount(order.pay_amount) }}</span>
            </div>
            <div v-if="order.amount !== order.pay_amount" class="flex justify-between">
              <span class="text-gray-500 dark:text-gray-400">{{ resultText('creditedAmount') }}</span>
              <span class="font-medium text-gray-900 dark:text-white">{{ formatOrderCreditedAmount(order) }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-500 dark:text-gray-400">{{ resultText('paymentMethod') }}</span>
              <span class="font-medium text-gray-900 dark:text-white">{{ paymentMethodLabel(order.payment_type) }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-500 dark:text-gray-400">{{ resultText('status') }}</span>
              <OrderStatusBadge :status="order.status" :labels="statusBadgeLabels" />
            </div>
          </div>
        </div>
        <!-- EasyPay return info (when no order loaded) -->
        <div v-else-if="returnInfo" class="rounded-xl bg-white p-5 shadow-sm dark:bg-dark-800">
          <div class="space-y-3 text-sm">
            <div v-if="returnInfo.outTradeNo" class="flex justify-between">
              <span class="text-gray-500 dark:text-gray-400">{{ resultText('orderId') }}</span>
              <span class="font-medium text-gray-900 dark:text-white">{{ returnInfo.outTradeNo }}</span>
            </div>
            <div v-if="returnInfo.money" class="flex justify-between">
              <span class="text-gray-500 dark:text-gray-400">{{ resultText('payAmount') }}</span>
              <span class="font-medium text-gray-900 dark:text-white">{{ formatGatewayAmount(Number(returnInfo.money) || 0) }}</span>
            </div>
            <div v-if="returnInfo.type" class="flex justify-between">
              <span class="text-gray-500 dark:text-gray-400">{{ resultText('paymentMethod') }}</span>
              <span class="font-medium text-gray-900 dark:text-white">{{ paymentMethodLabel(returnInfo.type) }}</span>
            </div>
          </div>
        </div>
        <!-- Actions -->
        <div class="flex gap-3">
          <button class="btn btn-secondary flex-1" @click="router.push(authRouteDefaults.purchasePath)">{{ resultText('backToRecharge') }}</button>
          <button class="btn btn-primary flex-1" @click="router.push(authRouteDefaults.ordersPath)">{{ resultText('viewOrders') }}</button>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { resolveRuntimeLocale } from '@/utils/runtimeLocale'
import { ref, computed, onBeforeUnmount, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import OrderStatusBadge from '@/components/payment/OrderStatusBadge.vue'
import {
  PAYMENT_RECOVERY_STORAGE_KEY,
  clearPaymentRecoverySnapshot,
} from '@/components/payment/paymentFlow'
import { usePaymentStore } from '@/stores/payment'
import { useAppStore } from '@/stores'
import { paymentAPI } from '@/api/payment'
import type { PaymentOrder } from '@/types/payment'
import { formatPaymentAmount, normalizePaymentCurrency } from '@/components/payment/currency'
import { normalizePaymentMethodForDisplay } from './paymentUx'
import { useAuthRouteDefaults } from '@/composables/useAuthRouteDefaults'
import {
  renderPaymentResultText,
  resolvePaymentResultDefaults,
  resolvePaymentResultLabels,
  type PaymentResultLabelKey,
} from '@/utils/paymentShell'
import { formatOrderCreditedAmount } from '@/utils/paymentCurrency'
import {
  applyResolvedPaymentOrder,
  calculatePaymentBaseAmount,
  calculatePaymentFeeAmount,
  clearPaymentRecoverySnapshotForTerminalStatus,
  isPaymentResultPending,
  isPaymentResultSuccess,
  readPaymentResultQueryString,
  restorePaymentRecoverySnapshot,
} from './paymentResultRuntime'

const i18n = useI18n()
const route = useRoute()
const router = useRouter()
const paymentStore = usePaymentStore()
const appStore = useAppStore()
const { authRouteDefaults } = useAuthRouteDefaults()

const order = ref<PaymentOrder | null>(null)
const loading = ref(true)
const currency = ref('')

interface ReturnInfo {
  outTradeNo: string
  money: string
  type: string
  tradeStatus: string
}
const returnInfo = ref<ReturnInfo | null>(null)

let statusRefreshTimer: ReturnType<typeof setTimeout> | null = null
const refreshAttempts = ref(0)

/** 充值金额 = pay_amount / (1 + fee_rate/100)，fee_rate=0 时等于 pay_amount */
const baseAmount = computed(() => calculatePaymentBaseAmount(order.value))

/** 手续费 = pay_amount - baseAmount */
const feeAmount = computed(() => calculatePaymentFeeAmount(order.value))

const localeCode = computed(() => {
  const raw = i18n.locale as unknown
  if (typeof raw === 'string') return raw
  if (raw && typeof raw === 'object' && 'value' in raw) {
    return String((raw as { value?: string }).value || '')
  }
  return undefined
})

const isSuccess = computed(() => {
  return isPaymentResultSuccess(order.value?.status)
})

const isPending = computed(() => {
  return isPaymentResultPending(order.value?.status)
})

const statusTitle = computed(() => {
  if (isSuccess.value) {
    return resultText('success')
  }
  if (isPending.value) {
    return resultText('processing')
  }
  return resultText('failed')
})


const paymentResultLabels = computed(() =>
  resolvePaymentResultLabels(
    appStore.cachedPublicSettings?.payment_shell_config,
    resolveRuntimeLocale(i18n.locale),
  ),
)
const paymentResultDefaults = computed(() =>
  resolvePaymentResultDefaults(
    appStore.cachedPublicSettings?.payment_shell_config,
    resolveRuntimeLocale(i18n.locale),
  ),
)

function resultText(key: PaymentResultLabelKey): string {
  return renderPaymentResultText(paymentResultLabels.value, key)
}

const statusBadgeLabels = computed(() => ({
  PENDING: resultText('statusPending'),
  PAID: resultText('statusPaid'),
  RECHARGING: resultText('statusRecharging'),
  COMPLETED: resultText('statusCompleted'),
  EXPIRED: resultText('statusExpired'),
  CANCELLED: resultText('statusCancelled'),
  FAILED: resultText('statusFailed'),
  REFUND_REQUESTED: resultText('statusRefundRequested'),
  REFUNDING: resultText('statusRefunding'),
  REFUNDED: resultText('statusRefunded'),
  PARTIALLY_REFUNDED: resultText('statusPartiallyRefunded'),
  REFUND_FAILED: resultText('statusRefundFailed'),
}))

function normalizedOrderPaymentType(paymentType: string): string {
  return normalizePaymentMethodForDisplay(paymentType) || paymentType
}

function paymentMethodLabel(paymentType: string): string {
  const normalized = normalizedOrderPaymentType(paymentType)
  if (normalized === 'alipay') return resultText('methodAlipay')
  if (normalized === 'wxpay') return resultText('methodWxpay')
  if (normalized === 'stripe') return resultText('methodStripe')
  if (normalized === 'airwallex') return resultText('methodAirwallex')
  return normalized
}

function formatGatewayAmount(value: number): string {
  return formatPaymentAmount(value, currency.value, localeCode.value)
}

function setResolvedOrder(nextOrder: PaymentOrder | null): void {
  const resolved = applyResolvedPaymentOrder(nextOrder, currency.value)
  order.value = resolved.order
  currency.value = resolved.currency
}

function readRouteQueryString(key: string): string {
  return readPaymentResultQueryString(route.query, key)
}

function restoreRecoverySnapshot(context: {
  resumeToken: string
  routeOrderId: number
  routeOutTradeNo: string
}) {
  return restorePaymentRecoverySnapshot(
    typeof window === 'undefined' ? null : window.localStorage,
    context,
  )
}

async function resolveOrderFromResumeToken(resumeToken: string): Promise<PaymentOrder | null> {
  try {
    const result = await paymentAPI.resolveOrderPublicByResumeToken(resumeToken)
    return result.data
  } catch (_err: unknown) {
    return null
  }
}

async function resolveOrderFromOutTradeNo(outTradeNo: string): Promise<PaymentOrder | null> {
  try {
    const result = await paymentAPI.verifyOrder(outTradeNo)
    return result.data
  } catch (_err: unknown) {
    try {
      const result = await paymentAPI.verifyOrderPublic(outTradeNo)
      return result.data
    } catch (_innerErr: unknown) {
      return null
    }
  }
}

function clearStatusRefreshTimer(): void {
  if (statusRefreshTimer !== null) {
    clearTimeout(statusRefreshTimer)
    statusRefreshTimer = null
  }
}

function clearRecoverySnapshot(): void {
  if (typeof window === 'undefined') return
  clearPaymentRecoverySnapshot(window.localStorage, PAYMENT_RECOVERY_STORAGE_KEY)
}

function clearRecoverySnapshotForTerminalStatus(status: string | null | undefined): void {
  clearPaymentRecoverySnapshotForTerminalStatus(
    typeof window === 'undefined' ? null : window.localStorage,
    status,
  )
}

function scheduleStatusRefresh(refreshOrder: (() => Promise<PaymentOrder | null>) | null): void {
  clearStatusRefreshTimer()
  if (!refreshOrder || !isPending.value || refreshAttempts.value >= paymentResultDefaults.value.maxRefreshAttempts) {
    return
  }

  statusRefreshTimer = setTimeout(async () => {
    refreshAttempts.value += 1
    const refreshedOrder = await refreshOrder()
    if (refreshedOrder) {
      setResolvedOrder(refreshedOrder)
      clearRecoverySnapshotForTerminalStatus(refreshedOrder.status)
    }

    if (isPaymentResultPending(order.value?.status)) {
      scheduleStatusRefresh(refreshOrder)
    }
  }, paymentResultDefaults.value.refreshIntervalMs)
}

onMounted(async () => {
  const resumeToken = readRouteQueryString('resume_token')
  const routeOrderId = Number(readRouteQueryString('order_id')) || 0
  let outTradeNo = readRouteQueryString('out_trade_no')
  let orderId = 0
  let resumeTokenLookupFailed = false

  const restored = restoreRecoverySnapshot({
    resumeToken,
    routeOrderId,
    routeOutTradeNo: outTradeNo,
  })
  if (restored?.orderId) {
    orderId = restored.orderId
  }
  if (restored?.currency) {
    currency.value = normalizePaymentCurrency(restored.currency)
  }
  if (!outTradeNo && restored?.outTradeNo) {
    outTradeNo = restored.outTradeNo
  }

  if (resumeToken) {
    const resolvedOrder = await resolveOrderFromResumeToken(resumeToken)
    if (resolvedOrder) {
      setResolvedOrder(resolvedOrder)
      if (!orderId) {
        orderId = resolvedOrder.id
      }
    } else if (routeOrderId > 0) {
      resumeTokenLookupFailed = true
      orderId = routeOrderId
    } else {
      resumeTokenLookupFailed = true
    }
  } else if (routeOrderId > 0) {
    orderId = routeOrderId
  }

  const hasLegacyFallbackContext = readRouteQueryString('trade_status').trim() !== ''
  const shouldUsePublicOutTradeNo = outTradeNo !== '' && (hasLegacyFallbackContext || routeOrderId > 0 || orderId > 0)

  if (!order.value && orderId && (!resumeToken || routeOrderId > 0)) {
    try {
      setResolvedOrder(await paymentStore.pollOrderStatus(orderId))
    } catch (_err: unknown) {
      // Order lookup failed, will try legacy fallback below when possible.
    }
  }

  if (!order.value && shouldUsePublicOutTradeNo && (!resumeToken || resumeTokenLookupFailed)) {
    const legacyOrder = await resolveOrderFromOutTradeNo(outTradeNo)
    if (legacyOrder) {
      setResolvedOrder(legacyOrder)
      if (!orderId) {
        orderId = legacyOrder.id
      }
    }
  }

  if (!order.value && !orderId && outTradeNo && hasLegacyFallbackContext) {
    returnInfo.value = {
      outTradeNo,
      money: String(route.query.money || ''),
      type: String(route.query.type || ''),
      tradeStatus: String(route.query.trade_status || ''),
    }
  }

  const refreshOrder = async (): Promise<PaymentOrder | null> => {
    if (resumeToken) {
      const resolvedOrder = await resolveOrderFromResumeToken(resumeToken)
      if (resolvedOrder) {
        return resolvedOrder
      }
    }

    if (orderId) {
      try {
        return await paymentStore.pollOrderStatus(orderId)
      } catch (_err: unknown) {
        // Fall through to legacy public verification when order polling is unavailable.
      }
    }

    if (shouldUsePublicOutTradeNo) {
      return await resolveOrderFromOutTradeNo(outTradeNo)
    }

    return null
  }

  if (isPaymentResultPending(order.value?.status)) {
    scheduleStatusRefresh(refreshOrder)
  } else if (order.value) {
    clearRecoverySnapshotForTerminalStatus(order.value.status)
  } else if (returnInfo.value) {
    clearRecoverySnapshot()
  }
  loading.value = false
})

onBeforeUnmount(() => {
  clearStatusRefreshTimer()
})
</script>
