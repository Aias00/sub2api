<template>
  <BaseDialog
    :show="show"
    :title="t('admin.users.profileSummary.title')"
    width="wide"
    :close-on-click-outside="true"
    @close="$emit('close')"
  >
    <div v-if="user" class="space-y-4">
      <div class="rounded-lg bg-gray-50 p-4 dark:bg-dark-700/70">
        <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div class="min-w-0">
            <div class="flex items-center gap-2">
              <div class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-full bg-primary-100 text-sm font-semibold text-primary-700 dark:bg-primary-900/30 dark:text-primary-300">
                {{ user.email.charAt(0).toUpperCase() }}
              </div>
              <div class="min-w-0">
                <div class="truncate text-sm font-semibold text-gray-900 dark:text-white">{{ user.email }}</div>
                <div class="mt-0.5 text-xs text-gray-500 dark:text-dark-400">
                  ID {{ user.public_id || user.id }} · {{ user.username || t('admin.users.profileSummary.noUsername') }}
                </div>
              </div>
            </div>
          </div>
          <div v-if="summary" class="flex flex-wrap items-center gap-2">
            <span class="rounded-full bg-primary-50 px-2.5 py-1 text-xs font-medium text-primary-700 dark:bg-primary-900/30 dark:text-primary-300">
              {{ summary.classification.label }}
            </span>
            <span class="rounded-full bg-gray-100 px-2.5 py-1 text-xs text-gray-600 dark:bg-dark-600 dark:text-dark-300">
              {{ t('admin.users.profileSummary.confidence') }}: {{ summary.classification.confidence }}
            </span>
          </div>
        </div>
      </div>

      <div v-if="loading" class="flex justify-center py-10">
        <svg class="h-7 w-7 animate-spin text-primary-500" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
        </svg>
      </div>

      <div v-else-if="error" class="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-700 dark:border-red-900/40 dark:bg-red-900/20 dark:text-red-300">
        {{ error }}
      </div>

      <template v-else-if="summary">
        <div class="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
          <SummaryMetric :label="t('admin.users.profileSummary.currentBalance')" :value="formatMoney(summary.user.balance)" />
          <SummaryMetric :label="t('admin.users.profileSummary.apiUsage')" :value="formatCount(summary.activity.api_usage_count)" :hint="formatMoney(summary.activity.api_actual_cost)" />
          <SummaryMetric :label="t('admin.users.profileSummary.paidOrders')" :value="formatCount(summary.payments.paid_order_count)" :hint="formatMoney(summary.payments.paid_amount)" />
          <SummaryMetric :label="t('admin.users.profileSummary.businessTasks')" :value="formatCount(summary.business.image_task_count + summary.business.wechat_task_count)" :hint="formatMoney(summary.business.image_actual_cost + summary.business.wechat_actual_cost)" />
        </div>

        <section class="rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-600 dark:bg-dark-800">
          <div class="mb-3 flex items-center justify-between">
            <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.users.profileSummary.riskTags') }}</h3>
            <span class="text-xs text-gray-500 dark:text-dark-400">{{ summary.classification.category }}</span>
          </div>
          <div v-if="summary.risk_tags.length" class="flex flex-wrap gap-2">
            <span
              v-for="tag in summary.risk_tags"
              :key="tag.key"
              class="inline-flex max-w-full items-center gap-1 rounded-full px-2.5 py-1 text-xs"
              :class="tagClass(tag.severity)"
              :title="tag.detail"
            >
              {{ tag.label }}
              <span v-if="tag.detail" class="truncate opacity-75">· {{ tag.detail }}</span>
            </span>
          </div>
          <p v-else class="text-sm text-gray-500 dark:text-dark-400">{{ t('admin.users.profileSummary.noRiskTags') }}</p>
          <ul v-if="summary.classification.reasons.length" class="mt-3 space-y-1 text-xs text-gray-500 dark:text-dark-400">
            <li v-for="reason in summary.classification.reasons" :key="reason">· {{ reason }}</li>
          </ul>
        </section>

        <div class="grid gap-4 lg:grid-cols-2">
          <SummarySection :title="t('admin.users.profileSummary.registration')">
            <SummaryRow :label="t('admin.users.profileSummary.signupSource')" :value="summary.registration.registered_via" />
            <SummaryRow :label="t('admin.users.profileSummary.emailDomain')" :value="summary.registration.email_domain || '-'" />
            <SummaryRow :label="t('admin.users.profileSummary.registrationIp')" :value="summary.registration.registration_ip || '-'" />
            <SummaryRow :label="t('admin.users.profileSummary.userAgent')" :value="summary.registration.user_agent || '-'" />
            <SummaryRow :label="t('admin.users.profileSummary.acceptLanguage')" :value="summary.registration.accept_language || '-'" />
            <SummaryRow :label="t('admin.users.profileSummary.deviceFingerprint')" :value="summary.registration.device_fingerprint || '-'" />
            <SummaryRow :label="t('admin.users.profileSummary.nearbyEvent')" :value="formatNearbyEvent" />
            <SummaryRow :label="t('admin.users.profileSummary.sameIp24h')" :value="formatCount(summary.registration.same_ip_signup_count_24h)" />
            <SummaryRow :label="t('admin.users.profileSummary.sameDomain')" :value="formatCount(summary.registration.same_domain_signup_count)" />
            <div v-if="registrationHeaders.length" class="mt-3 border-t border-gray-100 pt-3 dark:border-dark-700">
              <div class="mb-2 text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('admin.users.profileSummary.headerSnapshot') }}</div>
              <div class="space-y-1">
                <div
                  v-for="[key, value] in registrationHeaders"
                  :key="key"
                  class="grid grid-cols-[8rem_1fr] gap-2 text-xs"
                >
                  <span class="truncate text-gray-500 dark:text-dark-400" :title="key">{{ key }}</span>
                  <span class="break-words text-gray-800 dark:text-gray-100" :title="value">{{ value }}</span>
                </div>
              </div>
            </div>
          </SummarySection>

          <SummarySection :title="t('admin.users.profileSummary.identity')">
            <SummaryRow :label="t('admin.users.profileSummary.createdAt')" :value="formatOptionalDate(summary.user.created_at)" />
            <SummaryRow :label="t('admin.users.profileSummary.lastLogin')" :value="formatOptionalDate(summary.user.last_login_at)" />
            <SummaryRow :label="t('admin.users.profileSummary.lastActive')" :value="formatOptionalDate(summary.user.last_active_at)" />
            <SummaryRow :label="t('admin.users.profileSummary.lastUsed')" :value="formatOptionalDate(summary.user.last_used_at)" />
            <div class="mt-3 border-t border-gray-100 pt-3 dark:border-dark-700">
              <div v-if="summary.auth_identities.length" class="space-y-2">
                <div v-for="identity in summary.auth_identities" :key="`${identity.provider_type}:${identity.provider_key}:${identity.provider_subject}`" class="rounded-md bg-gray-50 p-2 text-xs dark:bg-dark-700">
                  <div class="font-medium text-gray-800 dark:text-gray-100">{{ identity.provider_type }} / {{ identity.provider_key }}</div>
                  <div class="mt-0.5 truncate text-gray-500 dark:text-dark-400" :title="identity.provider_subject">{{ identity.provider_subject }}</div>
                </div>
              </div>
              <p v-else class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.users.profileSummary.noAuthIdentities') }}</p>
            </div>
          </SummarySection>

          <SummarySection :title="t('admin.users.profileSummary.apiAndBilling')">
            <SummaryRow :label="t('admin.users.profileSummary.activeApiKeys')" :value="`${summary.api_keys.active_count} / ${summary.api_keys.total_count}`" />
            <SummaryRow :label="t('admin.users.profileSummary.firstApiKey')" :value="formatOptionalDate(summary.api_keys.first_created_at)" />
            <SummaryRow :label="t('admin.users.profileSummary.lastApiKey')" :value="formatOptionalDate(summary.api_keys.last_created_at)" />
            <SummaryRow :label="t('admin.users.profileSummary.usageWindow')" :value="formatUsageWindow" />
            <SummaryRow :label="t('admin.users.profileSummary.ledger')" :value="`${formatCount(summary.balance.ledger_count)} · ${formatMoney(summary.balance.net_ledger_amount)}`" />
            <SummaryRow :label="t('admin.users.profileSummary.redeem')" :value="`${formatCount(summary.balance.redeem_count)} · ${formatMoney(summary.balance.redeem_balance_amount)}`" />
          </SummarySection>

          <SummarySection :title="t('admin.users.profileSummary.business')">
            <SummaryRow :label="t('admin.users.profileSummary.imageTasks')" :value="`${summary.business.image_success_count} / ${summary.business.image_task_count}`" />
            <SummaryRow :label="t('admin.users.profileSummary.imageCost')" :value="formatMoney(summary.business.image_actual_cost)" />
            <SummaryRow :label="t('admin.users.profileSummary.imageWindow')" :value="formatWindow(summary.business.first_image_task_at, summary.business.last_image_task_at)" />
            <SummaryRow :label="t('admin.users.profileSummary.wechatTasks')" :value="formatCount(summary.business.wechat_task_count)" />
            <SummaryRow :label="t('admin.users.profileSummary.wechatCost')" :value="formatMoney(summary.business.wechat_actual_cost)" />
            <SummaryRow :label="t('admin.users.profileSummary.wechatWindow')" :value="formatWindow(summary.business.first_wechat_task_at, summary.business.last_wechat_task_at)" />
          </SummarySection>
        </div>

        <section class="rounded-lg border border-gray-200 bg-white dark:border-dark-600 dark:bg-dark-800">
          <div class="flex items-center justify-between border-b border-gray-100 px-4 py-3 dark:border-dark-700">
            <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.users.profileSummary.timeline') }}</h3>
            <span class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.users.profileSummary.timelineCount', { count: summary.timeline?.length || 0 }) }}</span>
          </div>
          <div v-if="summary.timeline?.length" class="divide-y divide-gray-100 dark:divide-dark-700">
            <div
              v-for="event in summary.timeline"
              :key="`${event.source}:${event.action}:${event.record_id || event.occurred_at}`"
              class="grid gap-3 px-4 py-3 sm:grid-cols-[9.5rem_1fr]"
            >
              <div class="text-xs text-gray-500 dark:text-dark-400">{{ formatDateTime(event.occurred_at) }}</div>
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-2">
                  <span class="rounded bg-gray-100 px-1.5 py-0.5 text-[11px] font-medium text-gray-600 dark:bg-dark-700 dark:text-dark-300">
                    {{ timelineSourceLabel(event.source) }}
                  </span>
                  <span class="text-sm font-medium text-gray-900 dark:text-white">{{ event.title || event.action }}</span>
                  <span v-if="event.status" class="rounded-full bg-gray-50 px-2 py-0.5 text-[11px] text-gray-500 ring-1 ring-gray-200 dark:bg-dark-700 dark:text-dark-300 dark:ring-dark-600">{{ event.status }}</span>
                  <span v-if="typeof event.amount === 'number'" class="text-xs font-medium" :class="event.amount < 0 ? 'text-red-600 dark:text-red-300' : 'text-emerald-600 dark:text-emerald-300'">
                    {{ formatSignedMoney(event.amount) }}
                  </span>
                </div>
                <div v-if="event.detail" class="mt-1 break-words text-xs text-gray-500 dark:text-dark-400">{{ event.detail }}</div>
                <div v-if="event.ip_address || event.user_agent" class="mt-1 flex flex-wrap gap-x-3 gap-y-1 text-[11px] text-gray-400 dark:text-dark-500">
                  <span v-if="event.ip_address">IP {{ event.ip_address }}</span>
                  <span v-if="event.user_agent" class="max-w-full truncate" :title="event.user_agent">UA {{ event.user_agent }}</span>
                </div>
              </div>
            </div>
          </div>
          <div v-else class="px-4 py-8 text-center text-sm text-gray-500 dark:text-dark-400">
            {{ t('admin.users.profileSummary.noTimeline') }}
          </div>
        </section>
      </template>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type { UserProfileSummary } from '@/api/admin/users'
import type { AdminUser } from '@/types'
import { formatDateTime } from '@/utils/format'
import BaseDialog from '@/components/common/BaseDialog.vue'

const props = defineProps<{ show: boolean; user: AdminUser | null }>()
defineEmits<{ close: [] }>()

const { t } = useI18n()
const summary = ref<UserProfileSummary | null>(null)
const loading = ref(false)
const error = ref('')

const SummaryMetric = defineComponent({
  props: {
    label: { type: String, required: true },
    value: { type: String, required: true },
    hint: { type: String, default: '' }
  },
  setup(props) {
    return () => h('div', { class: 'rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-600 dark:bg-dark-800' }, [
      h('div', { class: 'text-xs text-gray-500 dark:text-dark-400' }, props.label),
      h('div', { class: 'mt-1 text-lg font-semibold text-gray-900 dark:text-white' }, props.value),
      props.hint ? h('div', { class: 'mt-1 text-xs text-gray-500 dark:text-dark-400' }, props.hint) : null
    ])
  }
})

const SummarySection = defineComponent({
  props: {
    title: { type: String, required: true }
  },
  setup(props, { slots }) {
    return () => h('section', { class: 'rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-600 dark:bg-dark-800' }, [
      h('h3', { class: 'mb-3 text-sm font-semibold text-gray-900 dark:text-white' }, props.title),
      h('div', { class: 'space-y-2' }, slots.default?.())
    ])
  }
})

const SummaryRow = defineComponent({
  props: {
    label: { type: String, required: true },
    value: { type: String, required: true }
  },
  setup(props) {
    return () => h('div', { class: 'flex items-start justify-between gap-3 text-sm' }, [
      h('span', { class: 'text-gray-500 dark:text-dark-400' }, props.label),
      h('span', { class: 'max-w-[65%] break-words text-right font-medium text-gray-900 dark:text-white' }, props.value || '-')
    ])
  }
})

watch(
  () => [props.show, props.user?.public_id || props.user?.id] as const,
  ([show, userID]) => {
    if (show && userID) {
      loadSummary(userID)
    } else if (!show) {
      summary.value = null
      error.value = ''
    }
  },
  { immediate: true }
)

async function loadSummary(userID: string | number) {
  loading.value = true
  error.value = ''
  try {
    summary.value = await adminAPI.users.getUserProfileSummary(userID)
  } catch (err: any) {
    error.value = err?.response?.data?.message || err?.message || t('admin.users.profileSummary.loadFailed')
  } finally {
    loading.value = false
  }
}

const formatNearbyEvent = computed(() => {
  if (!summary.value) return '-'
  const event = summary.value.registration.nearby_auth_event || '-'
  const status = summary.value.registration.nearby_auth_status || '-'
  const at = formatOptionalDate(summary.value.registration.nearby_auth_at)
  return `${event} · ${status} · ${at}`
})

const formatUsageWindow = computed(() => {
  if (!summary.value) return '-'
  return formatWindow(summary.value.activity.first_api_usage_at, summary.value.activity.last_api_usage_at)
})

const registrationHeaders = computed<[string, string][]>(() => {
  const headers = summary.value?.registration.header_snapshot
  if (!headers) return []
  return Object.entries(headers).sort(([a], [b]) => a.localeCompare(b))
})

function formatMoney(value: number | undefined | null): string {
  return `$${Number(value || 0).toFixed(4)}`
}

function formatSignedMoney(value: number | undefined | null): string {
  const amount = Number(value || 0)
  const sign = amount < 0 ? '-' : amount > 0 ? '+' : ''
  return `${sign}$${Math.abs(amount).toFixed(4)}`
}

function formatCount(value: number | undefined | null): string {
  return Number(value || 0).toLocaleString()
}

function formatOptionalDate(value?: string | null): string {
  return value ? formatDateTime(value) : '-'
}

function formatWindow(start?: string | null, end?: string | null): string {
  if (!start && !end) return '-'
  if (start === end || !end) return formatOptionalDate(start)
  if (!start) return formatOptionalDate(end)
  return `${formatDateTime(start)} → ${formatDateTime(end)}`
}

function tagClass(severity: string): string {
  switch (severity) {
    case 'danger':
      return 'bg-red-50 text-red-700 ring-1 ring-red-200 dark:bg-red-900/20 dark:text-red-300 dark:ring-red-900/40'
    case 'warning':
      return 'bg-amber-50 text-amber-700 ring-1 ring-amber-200 dark:bg-amber-900/20 dark:text-amber-300 dark:ring-amber-900/40'
    default:
      return 'bg-gray-100 text-gray-700 ring-1 ring-gray-200 dark:bg-dark-700 dark:text-dark-300 dark:ring-dark-600'
  }
}

function timelineSourceLabel(source: string): string {
  return t(`admin.users.profileSummary.timelineSources.${source}`, source)
}
</script>
