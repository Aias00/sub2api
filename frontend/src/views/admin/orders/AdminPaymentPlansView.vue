<template>
  <AppLayout>
    <div class="space-y-5">
      <div class="overflow-hidden rounded-3xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900">
        <div class="grid gap-5 p-5 lg:grid-cols-[1.2fr_0.8fr] lg:items-center">
          <div>
            <p class="text-xs font-semibold uppercase tracking-[0.24em] text-primary-500">{{ t('payment.admin.catalog.badge') }}</p>
            <h1 class="mt-2 text-2xl font-black tracking-tight text-gray-950 dark:text-white">{{ t('payment.admin.catalog.title') }}</h1>
            <p class="mt-2 max-w-2xl text-sm leading-6 text-gray-500 dark:text-gray-400">
              {{ t('payment.admin.catalog.description') }}
            </p>
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div class="rounded-2xl bg-gray-50 p-4 dark:bg-dark-800/70">
              <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.catalog.stats.plans') }}</p>
              <p class="mt-1 text-2xl font-black text-gray-950 dark:text-white">{{ plans.length }}</p>
            </div>
            <div class="rounded-2xl bg-gray-50 p-4 dark:bg-dark-800/70">
              <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.catalog.stats.onSale') }}</p>
              <p class="mt-1 text-2xl font-black text-gray-950 dark:text-white">{{ onSalePlanCount }}</p>
            </div>
          </div>
        </div>

        <div class="border-t border-gray-100 p-2 dark:border-dark-700">
          <div class="grid gap-2 rounded-2xl bg-gray-100 p-1 dark:bg-dark-800 sm:grid-cols-2">
            <button
              v-for="tab in catalogTabs"
              :key="tab.key"
              type="button"
              class="rounded-xl px-4 py-3 text-sm font-semibold transition-all"
              :class="activeTab === tab.key
                ? 'bg-white text-gray-950 shadow-sm dark:bg-dark-700 dark:text-white'
                : 'text-gray-500 hover:text-gray-800 dark:text-gray-400 dark:hover:text-gray-200'"
              @click="activeTab = tab.key"
            >
              <span class="flex items-center justify-center gap-2">
                <Icon :name="tab.icon" size="sm" />
                {{ tab.label }}
              </span>
            </button>
          </div>
        </div>
      </div>

      <RechargeProductsManager v-if="activeTab === 'recharge'" />

      <div v-else-if="activeTab === 'plans'" class="space-y-4">
        <div class="overflow-hidden rounded-3xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900">
          <div class="flex flex-col gap-4 border-b border-gray-100 p-5 dark:border-dark-700 md:flex-row md:items-center md:justify-between">
            <div>
              <p class="text-xs font-semibold uppercase tracking-[0.24em] text-primary-500">{{ t('payment.admin.catalog.salesBadge') }}</p>
              <h2 class="mt-2 text-xl font-bold text-gray-950 dark:text-white">{{ t('payment.admin.plansPageTitle') }}</h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('payment.admin.catalog.salesDescription') }}</p>
            </div>
            <div class="flex items-center gap-2">
              <button @click="loadPlans" :disabled="plansLoading" class="btn btn-secondary" :title="t('common.refresh')">
                <Icon name="refresh" size="md" :class="plansLoading ? 'animate-spin' : ''" />
              </button>
              <button @click="openPlanEdit(null)" class="btn btn-primary">
                <Icon name="plus" size="sm" />
                {{ t('payment.admin.createPlan') }}
              </button>
            </div>
          </div>

          <DataTable :columns="planColumns" :data="plans" :loading="plansLoading">
            <template #cell-name="{ value, row }">
              <span class="text-sm font-medium" :class="getPlanNameClass(row.group_id)">{{ value }}</span>
            </template>
            <template #cell-group_id="{ value }">
              <span v-if="isGroupMissing(value)" class="text-sm">
                <span class="text-gray-400">#{{ value }}</span>
                <span class="ml-1 badge badge-danger">{{ t('payment.admin.groupMissing') }}</span>
              </span>
              <GroupBadge
                v-else-if="getGroup(value)"
                :name="getGroup(value)!.name"
                :platform="getGroup(value)!.platform"
                :rate-multiplier="getGroup(value)!.rate_multiplier"
              />
              <span v-else class="text-sm text-gray-400">-</span>
            </template>
            <template #cell-price="{ value, row }">
              <div class="text-sm">
                <span class="font-medium text-gray-900 dark:text-white">${{ (value ?? 0).toFixed(2) }}</span>
                <span v-if="row.original_price" class="ml-1 text-xs text-gray-400 line-through">${{ row.original_price.toFixed(2) }}</span>
              </div>
            </template>
            <template #cell-validity_days="{ value, row }">
              <span class="text-sm">{{ formatValidity(value, row.validity_unit) }}</span>
            </template>
            <template #cell-for_sale="{ value, row }">
              <button
                type="button"
                :class="[
                  'relative inline-flex h-5 w-9 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2',
                  value ? 'bg-primary-500' : 'bg-gray-300 dark:bg-dark-600'
                ]"
                @click="toggleForSale(row)"
              >
                <span :class="[
                  'pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
                  value ? 'translate-x-4' : 'translate-x-0'
                ]" />
              </button>
            </template>
            <template #cell-actions="{ row }">
              <div class="flex items-center gap-2">
                <button @click="openPlanEdit(row)" class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-blue-50 hover:text-blue-600 dark:hover:bg-blue-900/20 dark:hover:text-blue-400">
                  <Icon name="edit" size="sm" />
                  <span class="text-xs">{{ t('common.edit') }}</span>
                </button>
                <button @click="confirmDeletePlan(row)" class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400">
                  <Icon name="trash" size="sm" />
                  <span class="text-xs">{{ t('common.delete') }}</span>
                </button>
              </div>
            </template>
          </DataTable>
        </div>
      </div>
    </div>

    <!-- Plan Edit Dialog -->
    <PlanEditDialog :show="showPlanDialog" :plan="editingPlan" :groups="groups" :payment-config="paymentConfig" @close="showPlanDialog = false" @saved="loadPlans" />

    <ConfirmDialog :show="showDeletePlanDialog" :title="t('payment.admin.deletePlan')" :message="t('payment.admin.deletePlanConfirm')" :confirm-text="t('common.delete')" danger @confirm="handleDeletePlan" @cancel="showDeletePlanDialog = false" />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminPaymentAPI } from '@/api/admin/payment'
import type { AdminPaymentConfig } from '@/api/admin/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import adminAPI from '@/api/admin'
import type { SubscriptionPlan } from '@/types/payment'
import type { AdminGroup } from '@/types'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import PlanEditDialog from './PlanEditDialog.vue'
import RechargeProductsManager from './RechargeProductsManager.vue'
import { platformTextClass } from '@/utils/platformColors'

type CatalogTabKey = 'recharge' | 'plans'
type CatalogTab = {
  key: CatalogTabKey
  label: string
  icon: 'gift' | 'creditCard'
}

const { t } = useI18n()
const appStore = useAppStore()

const activeTab = ref<CatalogTabKey>('plans')
const catalogTabs = computed<CatalogTab[]>(() => [
  { key: 'recharge' as const, label: t('payment.admin.catalog.tabs.recharge'), icon: 'gift' },
  { key: 'plans' as const, label: t('payment.admin.catalog.tabs.plans'), icon: 'creditCard' },
])

// ==================== Groups ====================

const groups = ref<AdminGroup[]>([])
const paymentConfig = ref<AdminPaymentConfig | null>(null)

async function loadGroups() {
  try {
    groups.value = await adminAPI.groups.getAll()
  } catch { /* ignore */ }
}

async function loadPaymentConfig() {
  try {
    const res = await adminPaymentAPI.getConfig()
    paymentConfig.value = res.data
  } catch { /* preview only */ }
}

function getGroup(id: number): AdminGroup | undefined {
  return groups.value.find(g => g.id === id)
}

function isGroupMissing(id: number): boolean {
  return id > 0 && !groups.value.find(g => g.id === id)
}

function getPlanNameClass(groupId: number): string {
  const group = getGroup(groupId)
  return group ? platformTextClass(group.platform) : 'text-gray-900 dark:text-white'
}

function formatValidity(days: number, unit: string): string {
  const normalizedUnit = typeof unit === 'string' ? unit.trim() : ''
  return [days, normalizedUnit ? t('payment.admin.' + normalizedUnit) : ''].filter(Boolean).join(' ')
}

// ==================== Plans ====================

const plansLoading = ref(false)
const plans = ref<SubscriptionPlan[]>([])
const showPlanDialog = ref(false)
const showDeletePlanDialog = ref(false)
const editingPlan = ref<SubscriptionPlan | null>(null)
const deletingPlanId = ref<number | null>(null)
const onSalePlanCount = computed(() => plans.value.filter((plan) => plan.for_sale).length)

const planColumns = computed((): Column[] => [
  { key: 'id', label: 'ID' },
  { key: 'name', label: t('payment.admin.planName') },
  { key: 'group_id', label: t('payment.admin.group') },
  { key: 'price', label: t('payment.admin.price') },
  { key: 'validity_days', label: t('payment.admin.validityDays') },
  { key: 'for_sale', label: t('payment.admin.forSale') },
  { key: 'sort_order', label: t('payment.admin.sortOrder') },
  { key: 'actions', label: t('common.actions') },
])

async function loadPlans() {
  plansLoading.value = true
  try {
    const res = await adminPaymentAPI.getPlans()
    // Backend returns features as newline-separated string; parse to array
    plans.value = (res.data || []).map((p: Omit<SubscriptionPlan, 'features'> & { features: string | string[] }) => ({
      ...p,
      features: typeof p.features === 'string'
        ? p.features.split('\n').map((f: string) => f.trim()).filter(Boolean)
        : (p.features || []),
    }))
  }
  catch (err: unknown) { appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error'))) }
  finally { plansLoading.value = false }
}

function openPlanEdit(plan: SubscriptionPlan | null) {
  editingPlan.value = plan
  showPlanDialog.value = true
}


/** Quick toggle for_sale from the list */
async function toggleForSale(plan: SubscriptionPlan) {
  try {
    await adminPaymentAPI.updatePlan(plan.id, { for_sale: !plan.for_sale })
    plan.for_sale = !plan.for_sale
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  }
}

function confirmDeletePlan(plan: SubscriptionPlan) { deletingPlanId.value = plan.id; showDeletePlanDialog.value = true }
async function handleDeletePlan() {
  if (!deletingPlanId.value) return
  try { await adminPaymentAPI.deletePlan(deletingPlanId.value); appStore.showSuccess(t('common.deleted')); showDeletePlanDialog.value = false; loadPlans() }
  catch (err: unknown) { appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error'))) }
}

// ==================== Lifecycle ====================

onMounted(() => {
  loadGroups()
  loadPaymentConfig()
  loadPlans()
})
</script>
