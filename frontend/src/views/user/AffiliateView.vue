<template>
  <AppLayout>
    <div class="space-y-6">
      <div v-if="loading" class="flex justify-center py-12">
        <div
          class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"
        ></div>
      </div>

      <template v-else-if="detail">
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <div class="card p-5">
            <p class="flex items-center gap-1.5 text-sm text-gray-500 dark:text-dark-400">
              <Icon name="dollar" size="sm" class="text-primary-500" />
              {{ affiliateText('rebateRate') }}
            </p>
            <p class="mt-2 text-2xl font-semibold text-primary-600 dark:text-primary-400">
              {{ formattedRebateRate }}<span class="ml-0.5 text-base font-medium">%</span>
            </p>
            <p class="mt-1 text-xs text-gray-400 dark:text-dark-500">
              {{ affiliateText('rebateRateHint') }}
            </p>
          </div>
          <div class="card p-5">
            <p class="text-sm text-gray-500 dark:text-dark-400">{{ affiliateText('invitedUsers') }}</p>
            <p class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">
              {{ formatCount(detail.aff_count) }}
            </p>
          </div>
          <div class="card p-5">
            <p class="text-sm text-gray-500 dark:text-dark-400">{{ affiliateText('availableQuota') }}</p>
            <p class="mt-2 text-2xl font-semibold text-emerald-600 dark:text-emerald-400">
              {{ formatCurrency(detail.aff_quota) }}
            </p>
          </div>
          <div class="card p-5">
            <p class="text-sm text-gray-500 dark:text-dark-400">{{ affiliateText('totalQuota') }}</p>
            <p class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">
              {{ formatCurrency(detail.aff_history_quota) }}
            </p>
            <p v-if="detail.aff_frozen_quota > 0" class="mt-1 text-xs text-amber-600 dark:text-amber-400">
              {{ affiliateText('frozenQuota') }}: {{ formatCurrency(detail.aff_frozen_quota) }}
            </p>
          </div>
        </div>

        <div class="card p-6">
          <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ affiliateText('title') }}</h3>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ affiliateText('description') }}</p>

          <div class="mt-5 grid gap-4 md:grid-cols-2">
            <div class="space-y-2">
              <p class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ affiliateText('yourCode') }}</p>
              <div class="flex items-center gap-2 rounded-xl border border-gray-200 bg-gray-50 px-3 py-2 dark:border-dark-700 dark:bg-dark-900">
                <code class="flex-1 truncate text-sm font-semibold text-gray-900 dark:text-white">{{ detail.aff_code }}</code>
                <button class="btn btn-secondary btn-sm" @click="copyCode">
                  <Icon name="copy" size="sm" />
                  <span>{{ affiliateText('copyCode') }}</span>
                </button>
              </div>
            </div>

            <div class="space-y-2">
              <p class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ affiliateText('inviteLink') }}</p>
              <div class="flex items-center gap-2 rounded-xl border border-gray-200 bg-gray-50 px-3 py-2 dark:border-dark-700 dark:bg-dark-900">
                <code class="flex-1 truncate text-sm text-gray-700 dark:text-gray-300">{{ inviteLink }}</code>
                <button class="btn btn-secondary btn-sm" @click="copyInviteLink">
                  <Icon name="copy" size="sm" />
                  <span>{{ affiliateText('copyLink') }}</span>
                </button>
              </div>
            </div>
          </div>

          <div class="mt-5 rounded-xl border border-primary-200 bg-primary-50 p-4 dark:border-primary-900/40 dark:bg-primary-900/20">
            <p class="text-sm font-medium text-primary-800 dark:text-primary-200">{{ affiliateText('tipsTitle') }}</p>
            <ul class="mt-2 space-y-1 text-sm text-primary-700 dark:text-primary-300">
              <li>1. {{ affiliateText('tipShare') }}</li>
              <li>2. {{ affiliateText('tipRebate', { rate: `${formattedRebateRate}%` }) }}</li>
              <li>3. {{ affiliateText('tipTransfer') }}</li>
              <li v-if="detail.aff_frozen_quota > 0">4. {{ affiliateText('tipFreeze') }}</li>
            </ul>
          </div>
        </div>

        <div class="card p-6">
          <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ affiliateText('transferTitle') }}</h3>
              <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ affiliateText('transferDescription') }}</p>
            </div>
            <button
              class="btn btn-primary"
              :disabled="transferring || detail.aff_quota <= 0"
              @click="transferQuota"
            >
              <Icon v-if="transferring" name="refresh" size="sm" class="animate-spin" />
              <Icon v-else name="dollar" size="sm" />
              <span>{{ transferring ? affiliateText('transferring') : affiliateText('transferButton') }}</span>
            </button>
          </div>
          <p v-if="detail.aff_quota <= 0" class="mt-3 text-sm text-amber-600 dark:text-amber-400">
            {{ affiliateText('transferEmpty') }}
          </p>
        </div>

        <div class="card p-6">
          <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ affiliateText('inviteesTitle') }}</h3>
          <div v-if="detail.invitees.length === 0" class="mt-4 rounded-xl border border-dashed border-gray-300 p-6 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-dark-400">
            {{ affiliateText('inviteesEmpty') }}
          </div>
          <div v-else class="mt-4 overflow-x-auto">
            <table class="w-full min-w-[560px] text-left text-sm">
              <thead>
                <tr class="border-b border-gray-200 text-gray-500 dark:border-dark-700 dark:text-dark-400">
                  <th class="px-3 py-2 font-medium">{{ affiliateText('emailColumn') }}</th>
                  <th class="px-3 py-2 font-medium">{{ affiliateText('usernameColumn') }}</th>
                  <th class="px-3 py-2 font-medium text-right">{{ affiliateText('rebateColumn') }}</th>
                  <th class="px-3 py-2 font-medium">{{ affiliateText('joinedAtColumn') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="item in detail.invitees"
                  :key="item.user_id"
                  class="border-b border-gray-100 last:border-b-0 dark:border-dark-800"
                >
                  <td class="px-3 py-3 text-gray-900 dark:text-white">{{ item.email || '-' }}</td>
                  <td class="px-3 py-3 text-gray-700 dark:text-gray-300">{{ item.username || '-' }}</td>
                  <td class="px-3 py-3 text-right font-medium text-emerald-600 dark:text-emerald-400">{{ formatCurrency(item.total_rebate) }}</td>
                  <td class="px-3 py-3 text-gray-700 dark:text-gray-300">{{ formatDateTime(item.created_at) || '-' }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <div class="card p-6">
          <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ affiliateText('rebatesTitle') }}</h3>
          <div v-if="rebateLoading" class="flex justify-center py-8">
            <div class="h-6 w-6 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
          </div>
          <div v-else-if="rebateRecords.length === 0" class="mt-4 rounded-xl border border-dashed border-gray-300 p-6 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-dark-400">
            {{ affiliateText('rebatesEmpty') }}
          </div>
          <div v-else class="mt-4 overflow-x-auto">
            <table class="w-full min-w-[820px] text-left text-sm">
              <thead>
                <tr class="border-b border-gray-200 text-gray-500 dark:border-dark-700 dark:text-dark-400">
                  <th class="px-3 py-2 font-medium">{{ affiliateText('inviteeColumn') }}</th>
                  <th class="px-3 py-2 font-medium text-right">{{ affiliateText('orderAmountColumn') }}</th>
                  <th class="px-3 py-2 font-medium text-right">{{ affiliateText('payAmountColumn') }}</th>
                  <th class="px-3 py-2 font-medium text-right">{{ affiliateText('rebateAmountColumn') }}</th>
                  <th class="px-3 py-2 font-medium">{{ affiliateText('paymentTypeColumn') }}</th>
                  <th class="px-3 py-2 font-medium">{{ affiliateText('orderStatusColumn') }}</th>
                  <th class="px-3 py-2 font-medium">{{ affiliateText('createdAtColumn') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="item in rebateRecords"
                  :key="`${item.order_id}-${item.created_at}`"
                  class="border-b border-gray-100 last:border-b-0 dark:border-dark-800"
                >
                  <td class="px-3 py-3">
                    <div class="text-gray-900 dark:text-white">{{ item.invitee_email || '-' }}</div>
                    <div class="text-xs text-gray-500 dark:text-dark-400">{{ item.invitee_username || '-' }}</div>
                  </td>
                  <td class="px-3 py-3 text-right text-gray-700 dark:text-gray-300">{{ formatCurrency(item.order_amount) }}</td>
                  <td class="px-3 py-3 text-right text-gray-700 dark:text-gray-300">{{ formatCurrency(item.pay_amount) }}</td>
                  <td class="px-3 py-3 text-right font-medium text-emerald-600 dark:text-emerald-400">{{ formatCurrency(item.rebate_amount) }}</td>
                  <td class="px-3 py-3 text-gray-700 dark:text-gray-300">{{ formatPaymentType(item.payment_type) }}</td>
                  <td class="px-3 py-3"><OrderStatusBadge :status="asOrderStatus(item.order_status)" /></td>
                  <td class="px-3 py-3 text-gray-700 dark:text-gray-300">{{ formatDateTime(item.created_at) || '-' }}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <Pagination
            v-if="rebateTotal > rebatePageSize"
            class="mt-4"
            :page="rebatePage"
            :total="rebateTotal"
            :page-size="rebatePageSize"
            @update:page="changeRebatePage"
            @update:pageSize="changeRebatePageSize"
          />
        </div>

        <div class="card p-6">
          <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ affiliateText('transfersTitle') }}</h3>
          <div v-if="transferRecordLoading" class="flex justify-center py-8">
            <div class="h-6 w-6 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
          </div>
          <div v-else-if="transferRecords.length === 0" class="mt-4 rounded-xl border border-dashed border-gray-300 p-6 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-dark-400">
            {{ affiliateText('transfersEmpty') }}
          </div>
          <div v-else class="mt-4 overflow-x-auto">
            <table class="w-full min-w-[820px] text-left text-sm">
              <thead>
                <tr class="border-b border-gray-200 text-gray-500 dark:border-dark-700 dark:text-dark-400">
                  <th class="px-3 py-2 font-medium text-right">{{ affiliateText('amountColumn') }}</th>
                  <th class="px-3 py-2 font-medium text-right">{{ affiliateText('balanceAfterColumn') }}</th>
                  <th class="px-3 py-2 font-medium text-right">{{ affiliateText('availableQuotaAfterColumn') }}</th>
                  <th class="px-3 py-2 font-medium text-right">{{ affiliateText('frozenQuotaAfterColumn') }}</th>
                  <th class="px-3 py-2 font-medium text-right">{{ affiliateText('historyQuotaAfterColumn') }}</th>
                  <th class="px-3 py-2 font-medium">{{ affiliateText('transferredAtColumn') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="item in transferRecords"
                  :key="`${item.ledger_id}-${item.created_at}`"
                  class="border-b border-gray-100 last:border-b-0 dark:border-dark-800"
                >
                  <td class="px-3 py-3 text-right font-medium text-emerald-600 dark:text-emerald-400">{{ formatCurrency(item.amount) }}</td>
                  <td class="px-3 py-3 text-right text-gray-700 dark:text-gray-300">{{ formatNullableCurrency(item.balance_after) }}</td>
                  <td class="px-3 py-3 text-right text-gray-700 dark:text-gray-300">{{ formatNullableCurrency(item.available_quota_after) }}</td>
                  <td class="px-3 py-3 text-right text-gray-700 dark:text-gray-300">{{ formatNullableCurrency(item.frozen_quota_after) }}</td>
                  <td class="px-3 py-3 text-right text-gray-700 dark:text-gray-300">{{ formatNullableCurrency(item.history_quota_after) }}</td>
                  <td class="px-3 py-3 text-gray-700 dark:text-gray-300">{{ formatDateTime(item.created_at) || '-' }}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <Pagination
            v-if="transferTotal > transferPageSize"
            class="mt-4"
            :page="transferPage"
            :total="transferTotal"
            :page-size="transferPageSize"
            @update:page="changeTransferPage"
            @update:pageSize="changeTransferPageSize"
          />
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { resolveRuntimeLocale } from '@/utils/runtimeLocale'
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import Pagination from '@/components/common/Pagination.vue'
import OrderStatusBadge from '@/components/payment/OrderStatusBadge.vue'
import userAPI from '@/api/user'
import type { UserAffiliateDetail, AffiliateRebateRecord, AffiliateTransferRecord } from '@/types'
import type { OrderStatus } from '@/types/payment'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { useClipboard } from '@/composables/useClipboard'
import { formatCurrency, formatDateTime } from '@/utils/format'
import { extractApiErrorMessage } from '@/utils/apiError'
import {
  renderAffiliateShellText,
  resolveAffiliateShellLabels,
  type AffiliateLabelKey,
} from '@/utils/affiliateShell'
import {
  asAffiliateOrderStatus,
  changeAffiliatePage,
  changeAffiliatePageSize,
  formatAffiliateCount,
  formatAffiliateNullableCurrency,
  formatAffiliatePaymentType,
} from './affiliateRuntime'

const { t, te, locale } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const { copyToClipboard } = useClipboard()

const loading = ref(true)
const rebateLoading = ref(false)
const transferRecordLoading = ref(false)
const transferring = ref(false)
const detail = ref<UserAffiliateDetail | null>(null)
const rebateRecords = ref<AffiliateRebateRecord[]>([])
const rebatePage = ref(1)
const rebatePageSize = ref(10)
const rebateTotal = ref(0)
const transferRecords = ref<AffiliateTransferRecord[]>([])
const transferPage = ref(1)
const transferPageSize = ref(10)
const transferTotal = ref(0)


const affiliateLabels = computed(() =>
  resolveAffiliateShellLabels(
    appStore.cachedPublicSettings?.affiliate_shell_config,
    resolveRuntimeLocale(locale),
  ),
)

function affiliateText(key: AffiliateLabelKey, values?: Record<string, string | number>): string {
  return renderAffiliateShellText(affiliateLabels.value, key, values)
}

const inviteLink = computed(() => {
  if (!detail.value) return ''
  if (typeof window === 'undefined') return `/register?aff=${encodeURIComponent(detail.value.aff_code)}`
  return `${window.location.origin}/register?aff=${encodeURIComponent(detail.value.aff_code)}`
})

const formattedRebateRate = computed(() => {
  const v = detail.value?.effective_rebate_rate_percent ?? 0
  const rounded = Math.round(v * 100) / 100
  return Number.isInteger(rounded) ? String(rounded) : rounded.toString()
})

function formatCount(value: number): string {
  return formatAffiliateCount(value)
}

function formatNullableCurrency(value?: number | null): string {
  return formatAffiliateNullableCurrency(value, formatCurrency)
}

function formatPaymentType(paymentType: string): string {
  return formatAffiliatePaymentType(paymentType, te, t)
}

function asOrderStatus(status: string): OrderStatus {
  return asAffiliateOrderStatus(status)
}

async function loadAffiliateDetail(silent = false): Promise<void> {
  if (!silent) {
    loading.value = true
  }
  try {
    detail.value = await userAPI.getAffiliateDetail()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, affiliateText('loadFailed')))
  } finally {
    if (!silent) {
      loading.value = false
    }
  }
}

async function loadRebateRecords(): Promise<void> {
  rebateLoading.value = true
  try {
    const res = await userAPI.getAffiliateRebates({ page: rebatePage.value, page_size: rebatePageSize.value })
    rebateRecords.value = res.items ?? []
    rebateTotal.value = res.total ?? 0
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, affiliateText('loadFailed')))
  } finally {
    rebateLoading.value = false
  }
}

async function loadTransferRecords(): Promise<void> {
  transferRecordLoading.value = true
  try {
    const res = await userAPI.getAffiliateTransfers({ page: transferPage.value, page_size: transferPageSize.value })
    transferRecords.value = res.items ?? []
    transferTotal.value = res.total ?? 0
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, affiliateText('loadFailed')))
  } finally {
    transferRecordLoading.value = false
  }
}

async function copyCode(): Promise<void> {
  if (!detail.value?.aff_code) return
  await copyToClipboard(detail.value.aff_code, affiliateText('codeCopied'))
}

async function copyInviteLink(): Promise<void> {
  if (!inviteLink.value) return
  await copyToClipboard(inviteLink.value, affiliateText('linkCopied'))
}

async function transferQuota(): Promise<void> {
  if (!detail.value || detail.value.aff_quota <= 0 || transferring.value) return
  transferring.value = true
  try {
    const resp = await userAPI.transferAffiliateQuota()
    appStore.showSuccess(affiliateText('transferSuccess', { amount: formatCurrency(resp.transferred_quota) }))
    await Promise.all([
      loadAffiliateDetail(true),
      loadTransferRecords(),
      authStore.refreshUser().catch(() => undefined),
    ])
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, affiliateText('transferFailed')))
  } finally {
    transferring.value = false
  }
}

function changeRebatePage(page: number) {
  changeAffiliatePage(page, (next) => { rebatePage.value = next }, loadRebateRecords)
}

function changeRebatePageSize(pageSize: number) {
  changeAffiliatePageSize(
    pageSize,
    (next) => { rebatePageSize.value = next },
    () => { rebatePage.value = 1 },
    loadRebateRecords,
  )
}

function changeTransferPage(page: number) {
  changeAffiliatePage(page, (next) => { transferPage.value = next }, loadTransferRecords)
}

function changeTransferPageSize(pageSize: number) {
  changeAffiliatePageSize(
    pageSize,
    (next) => { transferPageSize.value = next },
    () => { transferPage.value = 1 },
    loadTransferRecords,
  )
}

onMounted(() => {
  void Promise.all([
    loadAffiliateDetail(),
    loadRebateRecords(),
    loadTransferRecords(),
  ])
})
</script>
