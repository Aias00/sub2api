<template>
  <AppLayout>
    <div class="space-y-8">
      <div v-if="loading" class="flex items-center justify-center py-12"><LoadingSpinner /></div>
      <template v-else-if="stats">
        <section class="page-hero">
          <div class="page-hero-grid items-start">
            <div class="space-y-5">
              <span class="page-kicker">{{ appStore.siteName }}</span>
              <div class="space-y-3">
                <p class="text-xs font-semibold uppercase tracking-[0.22em] text-primary-700 dark:text-primary-300">
                  {{ t('dashboard.title') }}
                </p>
                <h2 class="max-w-3xl text-3xl font-semibold tracking-tight text-gray-950 dark:text-white md:text-4xl">
                  <span class="text-gradient">{{ t('dashboard.welcomeMessage') }}</span>
                </h2>
                <p class="max-w-2xl text-sm leading-7 text-gray-600 dark:text-gray-300">
                  {{ user?.email || appStore.siteName }}
                </p>
              </div>
            </div>

            <div class="grid gap-3 sm:grid-cols-3 xl:grid-cols-1">
              <div class="metric-panel">
                <div class="flex items-center gap-3">
                  <div class="metric-icon text-[#caac5e] dark:text-[#dcb959]">
                    <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M2.25 18.75a60.07 60.07 0 0115.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 013 6h-.75m0 0v-.375c0-.621.504-1.125 1.125-1.125H20.25M2.25 6v9m18-10.5v.75c0 .414.336.75.75.75h.75m-1.5-1.5h.375c.621 0 1.125.504 1.125 1.125v9.75c0 .621-.504 1.125-1.125 1.125h-.375m1.5-1.5H21a.75.75 0 00-.75.75v.75m0 0H3.75m0 0h-.375a1.125 1.125 0 01-1.125-1.125V15m1.5 1.5v-.75A.75.75 0 003 15h-.75M15 10.5a3 3 0 11-6 0 3 3 0 016 0zm3 0h.008v.008H18V10.5zm-12 0h.008v.008H6V10.5z" />
                    </svg>
                  </div>
                  <div>
                    <p class="text-xs font-medium uppercase tracking-[0.18em] text-gray-500 dark:text-gray-400">{{ t('dashboard.balance') }}</p>
                    <p class="text-2xl font-semibold text-gray-950 dark:text-white">${{ formatMoney(user?.balance || 0) }}</p>
                  </div>
                </div>
              </div>

              <div class="metric-panel">
                <div class="flex items-center gap-3">
                  <div class="metric-icon text-primary-600 dark:text-primary-300">
                    <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z" />
                    </svg>
                  </div>
                  <div>
                    <p class="text-xs font-medium uppercase tracking-[0.18em] text-gray-500 dark:text-gray-400">{{ t('dashboard.todayRequests') }}</p>
                    <p class="text-2xl font-semibold text-gray-950 dark:text-white">{{ stats.today_requests }}</p>
                  </div>
                </div>
              </div>

              <div class="metric-panel">
                <div class="flex items-center gap-3">
                  <div class="metric-icon text-primary-600 dark:text-[#dcb959]">
                    <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M7.5 7.5h9m-9 4.5h9m-9 4.5H12M5.25 3.75h13.5A1.5 1.5 0 0120.25 5.25v13.5a1.5 1.5 0 01-1.5 1.5H5.25a1.5 1.5 0 01-1.5-1.5V5.25a1.5 1.5 0 011.5-1.5z" />
                    </svg>
                  </div>
                  <div>
                    <p class="text-xs font-medium uppercase tracking-[0.18em] text-gray-500 dark:text-gray-400">{{ t('dashboard.todayTokens') }}</p>
                    <p class="text-2xl font-semibold text-gray-950 dark:text-white">{{ formatCompact(stats.today_tokens) }}</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <UserDashboardStats :stats="stats" :balance="user?.balance || 0" :is-simple="authStore.isSimpleMode" />
        <UserDashboardCharts v-model:startDate="startDate" v-model:endDate="endDate" v-model:granularity="granularity" :loading="loadingCharts" :trend="trendData" :models="modelStats" @dateRangeChange="loadCharts" @granularityChange="loadCharts" />
        <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
          <div class="lg:col-span-2"><UserDashboardRecentUsage :data="recentUsage" :loading="loadingUsage" /></div>
          <div class="lg:col-span-1"><UserDashboardQuickActions /></div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { ref, computed, onMounted } from 'vue'; import { useAuthStore } from '@/stores/auth'; import { usageAPI, type UserDashboardStats as UserStatsType } from '@/api/usage'
import { useAppStore } from '@/stores'
import AppLayout from '@/components/layout/AppLayout.vue'; import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import UserDashboardStats from '@/components/user/dashboard/UserDashboardStats.vue'; import UserDashboardCharts from '@/components/user/dashboard/UserDashboardCharts.vue'
import UserDashboardRecentUsage from '@/components/user/dashboard/UserDashboardRecentUsage.vue'; import UserDashboardQuickActions from '@/components/user/dashboard/UserDashboardQuickActions.vue'
import type { UsageLog, TrendDataPoint, ModelStat } from '@/types'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore(); const user = computed(() => authStore.user)
const stats = ref<UserStatsType | null>(null); const loading = ref(false); const loadingUsage = ref(false); const loadingCharts = ref(false)
const trendData = ref<TrendDataPoint[]>([]); const modelStats = ref<ModelStat[]>([]); const recentUsage = ref<UsageLog[]>([])

const formatLD = (d: Date) => d.toISOString().split('T')[0]
const formatCompact = (n: number) => {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`
  if (n >= 1000) return `${(n / 1000).toFixed(1)}K`
  return n.toString()
}
const formatMoney = (n: number) => n.toFixed(2)
const startDate = ref(formatLD(new Date(Date.now() - 6 * 86400000))); const endDate = ref(formatLD(new Date())); const granularity = ref('day')

const loadStats = async () => { loading.value = true; try { await authStore.refreshUser(); stats.value = await usageAPI.getDashboardStats() } catch (error) { console.error('Failed to load dashboard stats:', error) } finally { loading.value = false } }
const loadCharts = async () => { loadingCharts.value = true; try { const res = await Promise.all([usageAPI.getDashboardTrend({ start_date: startDate.value, end_date: endDate.value, granularity: granularity.value as any }), usageAPI.getDashboardModels({ start_date: startDate.value, end_date: endDate.value })]); trendData.value = res[0].trend || []; modelStats.value = res[1].models || [] } catch (error) { console.error('Failed to load charts:', error) } finally { loadingCharts.value = false } }
const loadRecent = async () => { loadingUsage.value = true; try { const res = await usageAPI.getByDateRange(startDate.value, endDate.value); recentUsage.value = res.items.slice(0, 5) } catch (error) { console.error('Failed to load recent usage:', error) } finally { loadingUsage.value = false } }

onMounted(() => { loadStats(); loadCharts(); loadRecent() })
</script>
