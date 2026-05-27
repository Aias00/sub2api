<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="card p-6">
        <h2 class="text-xl font-semibold text-gray-900 dark:text-white">
          {{ t('admin.affiliates.moduleOverview.title') }}
        </h2>
        <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
          {{ t('admin.affiliates.moduleOverview.description') }}
        </p>
      </div>

      <div v-if="loading" class="flex justify-center py-12">
        <div class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
      </div>

      <template v-else-if="overview">
        <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <MetricCard :label="t('admin.affiliates.moduleOverview.affiliateEnabled')" :value="overview.affiliate_enabled ? t('admin.affiliates.moduleOverview.enabled') : t('admin.affiliates.moduleOverview.disabled')" />
          <MetricCard :label="t('admin.affiliates.moduleOverview.invitationEnabled')" :value="overview.invitation_code_enabled ? t('admin.affiliates.moduleOverview.enabled') : t('admin.affiliates.moduleOverview.disabled')" />
          <MetricCard :label="t('admin.affiliates.moduleOverview.rebateRate')" :value="formatPercent(overview.affiliate_rebate_rate)" />
          <MetricCard :label="t('admin.affiliates.moduleOverview.freezeHours')" :value="formatHours(overview.affiliate_rebate_freeze_hours)" />
          <MetricCard :label="t('admin.affiliates.moduleOverview.durationDays')" :value="formatDays(overview.affiliate_rebate_duration_days)" />
          <MetricCard :label="t('admin.affiliates.moduleOverview.capPerInvitee')" :value="formatCap(overview.affiliate_rebate_per_invitee_cap)" />
          <MetricCard :label="t('admin.affiliates.moduleOverview.invitedUsers')" :value="formatCount(overview.invited_user_count)" />
          <MetricCard :label="t('admin.affiliates.moduleOverview.rebatedInvitees')" :value="formatCount(overview.rebated_invitee_count)" />
        </div>

        <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <MetricCard :label="t('admin.affiliates.moduleOverview.availableQuota')" :value="formatCurrency(overview.available_quota_total)" />
          <MetricCard :label="t('admin.affiliates.moduleOverview.frozenQuota')" :value="formatCurrency(overview.frozen_quota_total)" />
          <MetricCard :label="t('admin.affiliates.moduleOverview.historyQuota')" :value="formatCurrency(overview.history_quota_total)" />
          <MetricCard :label="t('admin.affiliates.moduleOverview.recentRebates')" :value="formatCount(overview.recent_rebate_record_count)" />
        </div>

        <div class="card p-6">
          <h3 class="text-base font-semibold text-gray-900 dark:text-white">
            {{ t('admin.affiliates.moduleOverview.quickLinks') }}
          </h3>
          <div class="mt-4 grid gap-3 md:grid-cols-2 xl:grid-cols-5">
            <router-link to="/admin/affiliates/rules" class="btn btn-secondary justify-center">{{ t('nav.affiliateRules') }}</router-link>
            <router-link to="/admin/affiliates/codes" class="btn btn-secondary justify-center">{{ t('nav.affiliateCodeManagement') }}</router-link>
            <router-link to="/admin/affiliates/invites" class="btn btn-secondary justify-center">{{ t('nav.affiliateInviteRecords') }}</router-link>
            <router-link to="/admin/affiliates/rebates" class="btn btn-secondary justify-center">{{ t('nav.affiliateRebateRecords') }}</router-link>
            <router-link to="/admin/affiliates/transfers" class="btn btn-secondary justify-center">{{ t('nav.affiliateTransferRecords') }}</router-link>
          </div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { defineComponent, h, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import { affiliatesAPI } from '@/api/admin/affiliates'
import type { AffiliateAdminOverview } from '@/types'
import { useAppStore } from '@/stores/app'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { formatCurrency as formatCurrencyBase } from '@/utils/format'

const { t } = useI18n()
const appStore = useAppStore()
const loading = ref(false)
const overview = ref<AffiliateAdminOverview | null>(null)

const MetricCard = defineComponent({
  name: 'AffiliateOverviewMetricCard',
  props: {
    label: { type: String, required: true },
    value: { type: String, required: true },
  },
  setup(props) {
    return () => h('div', { class: 'card p-5' }, [
      h('p', { class: 'text-sm text-gray-500 dark:text-dark-400' }, props.label),
      h('p', { class: 'mt-2 break-all text-2xl font-semibold text-gray-900 dark:text-white' }, props.value),
    ])
  },
})

function formatCount(value: number): string {
  return value.toLocaleString()
}

function formatPercent(value: number): string {
  return `${(Math.round(value * 100) / 100).toString()}%`
}

function formatCurrency(value: number): string {
  return formatCurrencyBase(value)
}

function formatHours(value: number): string {
  if (value <= 0) return t('admin.affiliates.moduleOverview.disabled')
  return t('admin.affiliates.moduleOverview.hoursUnit', { count: value })
}

function formatDays(value: number): string {
  if (value <= 0) return t('admin.affiliates.moduleOverview.noLimit')
  return t('admin.affiliates.moduleOverview.daysUnit', { count: value })
}

function formatCap(value: number): string {
  if (value <= 0) return t('admin.affiliates.moduleOverview.noLimit')
  return formatCurrency(value)
}

async function loadOverview() {
  loading.value = true
  try {
    overview.value = await affiliatesAPI.getOverview()
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'admin.affiliates.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void loadOverview()
})
</script>
