<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">{{ t('admin.userInsights.title') }}</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('admin.userInsights.description') }}</p>
        </div>
        <button
          type="button"
          class="inline-flex items-center gap-2 rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm font-medium text-gray-700 shadow-sm transition-colors hover:bg-gray-50 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-200 dark:hover:bg-dark-700"
          :disabled="loading"
          @click="loadInsights"
        >
          <Icon name="refresh" size="sm" :class="{ 'animate-spin': loading }" />
          {{ t('common.refresh') }}
        </button>
      </div>

      <div v-if="error" class="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-700 dark:border-red-900/40 dark:bg-red-900/20 dark:text-red-300">
        {{ error }}
      </div>

      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        <section class="rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-600 dark:bg-dark-800">
          <h2 class="mb-3 text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.userInsights.classification') }}</h2>
          <div class="space-y-2">
            <InsightBar v-for="item in insights?.classification || []" :key="item.key" :label="item.label" :count="item.count" :total="totalUsers" />
          </div>
        </section>

        <section class="rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-600 dark:bg-dark-800">
          <h2 class="mb-3 text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.userInsights.signupSources') }}</h2>
          <div class="space-y-2">
            <InsightBar v-for="item in insights?.signup_sources || []" :key="item.key" :label="item.label" :count="item.count" :total="totalUsers" />
          </div>
        </section>

        <section class="rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-600 dark:bg-dark-800">
          <h2 class="mb-3 text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.userInsights.funnel') }}</h2>
          <div class="space-y-2">
            <div v-for="step in insights?.funnel || []" :key="step.key" class="flex items-center justify-between gap-3">
              <div class="min-w-0">
                <div class="truncate text-sm font-medium text-gray-800 dark:text-gray-100">{{ step.label }}</div>
                <div class="text-xs text-gray-500 dark:text-dark-400">{{ formatPercent(step.conversion) }}</div>
              </div>
              <div class="text-sm font-semibold text-gray-900 dark:text-white">{{ formatCount(step.count) }}</div>
            </div>
          </div>
        </section>
      </div>

      <div class="grid gap-4 lg:grid-cols-3">
        <DimensionCard :title="t('admin.userInsights.topIPs')" :items="insights?.registration_ips || []" />
        <DimensionCard :title="t('admin.userInsights.topUserAgents')" :items="insights?.user_agents || []" />
        <DimensionCard :title="t('admin.userInsights.topLanguages')" :items="insights?.languages || []" />
      </div>

      <section class="rounded-lg border border-gray-200 bg-white dark:border-dark-600 dark:bg-dark-800">
        <div class="flex items-center justify-between border-b border-gray-100 px-4 py-3 dark:border-dark-700">
          <h2 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.userInsights.riskSamples') }}</h2>
          <span class="text-xs text-gray-500 dark:text-dark-400">
            {{ insights?.generated_at ? formatDateTime(insights.generated_at) : '-' }}
          </span>
        </div>
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-100 dark:divide-dark-700">
            <thead class="bg-gray-50 dark:bg-dark-700/50">
              <tr>
                <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('admin.userInsights.user') }}</th>
                <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('admin.userInsights.label') }}</th>
                <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('admin.userInsights.registrationIp') }}</th>
                <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('admin.userInsights.createdAt') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-for="sample in insights?.risk_samples || []" :key="sample.user_id" class="hover:bg-gray-50 dark:hover:bg-dark-700/40">
                <td class="px-4 py-3">
                  <RouterLink :to="`/admin/users?search=${encodeURIComponent(sample.email)}`" class="text-sm font-medium text-primary-600 hover:text-primary-700 dark:text-primary-400">
                    {{ sample.email }}
                  </RouterLink>
                  <div class="text-xs text-gray-500 dark:text-dark-400">ID {{ sample.user_id }} · {{ sample.username || '-' }}</div>
                </td>
                <td class="px-4 py-3">
                  <span class="inline-flex rounded-full px-2 py-0.5 text-xs font-medium" :class="sample.severity === 'warning' ? 'bg-amber-50 text-amber-700 dark:bg-amber-900/20 dark:text-amber-300' : 'bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-dark-300'">
                    {{ sample.label }}
                  </span>
                  <div class="mt-1 max-w-md text-xs text-gray-500 dark:text-dark-400">{{ sample.reason }}</div>
                </td>
                <td class="px-4 py-3 text-sm text-gray-700 dark:text-gray-200">{{ sample.registration_ip || '-' }}</td>
                <td class="px-4 py-3 text-sm text-gray-700 dark:text-gray-200">{{ formatDateTime(sample.created_at) }}</td>
              </tr>
              <tr v-if="!loading && !(insights?.risk_samples || []).length">
                <td colspan="4" class="px-4 py-8 text-center text-sm text-gray-500 dark:text-dark-400">{{ t('admin.userInsights.noRiskSamples') }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, ref, type PropType } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { adminAPI } from '@/api/admin'
import type { UserInsightDimension, UserProfileInsights } from '@/api/admin/users'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()
const insights = ref<UserProfileInsights | null>(null)
const loading = ref(false)
const error = ref('')

const totalUsers = computed(() => {
  return (insights.value?.classification || []).reduce((sum, item) => sum + item.count, 0)
})

const InsightBar = defineComponent({
  props: {
    label: { type: String, required: true },
    count: { type: Number, required: true },
    total: { type: Number, required: true },
  },
  setup(props) {
    return () => {
      const percent = props.total > 0 ? Math.round((props.count / props.total) * 100) : 0
      return h('div', [
        h('div', { class: 'mb-1 flex items-center justify-between gap-3 text-sm' }, [
          h('span', { class: 'truncate text-gray-700 dark:text-gray-200' }, props.label),
          h('span', { class: 'font-semibold text-gray-900 dark:text-white' }, props.count.toLocaleString()),
        ]),
        h('div', { class: 'h-2 rounded-full bg-gray-100 dark:bg-dark-700' }, [
          h('div', {
            class: 'h-2 rounded-full bg-primary-500',
            style: { width: `${Math.min(percent, 100)}%` },
          }),
        ]),
      ])
    }
  },
})

const DimensionCard = defineComponent({
  props: {
    title: { type: String, required: true },
    items: { type: Array as PropType<UserInsightDimension[]>, required: true },
  },
  setup(props) {
    return () => h('section', { class: 'rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-600 dark:bg-dark-800' }, [
      h('h2', { class: 'mb-3 text-sm font-semibold text-gray-900 dark:text-white' }, props.title),
      props.items.length
        ? h('div', { class: 'space-y-3' }, props.items.map((item) =>
            h('div', { class: 'min-w-0' }, [
              h('div', { class: 'flex items-start justify-between gap-3' }, [
                h('div', { class: 'min-w-0 break-words text-sm text-gray-800 dark:text-gray-100', title: item.value }, item.value),
                h('div', { class: 'flex-shrink-0 text-sm font-semibold text-gray-900 dark:text-white' }, item.count.toLocaleString()),
              ]),
              item.last_seen ? h('div', { class: 'mt-0.5 text-xs text-gray-500 dark:text-dark-400' }, formatDateTime(item.last_seen)) : null,
            ])
          ))
        : h('div', { class: 'py-6 text-center text-sm text-gray-500 dark:text-dark-400' }, '-'),
    ])
  },
})

onMounted(() => {
  loadInsights()
})

async function loadInsights() {
  loading.value = true
  error.value = ''
  try {
    insights.value = await adminAPI.users.getUserProfileInsights(12)
  } catch (err: any) {
    error.value = err?.response?.data?.message || err?.message || t('admin.userInsights.loadFailed')
  } finally {
    loading.value = false
  }
}

function formatCount(value: number): string {
  return Number(value || 0).toLocaleString()
}

function formatPercent(value: number): string {
  return `${Math.round(Number(value || 0) * 100)}%`
}
</script>
