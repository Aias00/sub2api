<template>
  <AppLayout>
    <div class="mx-auto max-w-4xl space-y-6">
      <!-- Header -->
      <section
        class="rounded-2xl border border-gray-200 bg-white p-6 shadow-sm dark:border-dark-700 dark:bg-dark-900 sm:p-8"
      >
        <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h1 class="text-3xl font-black text-gray-950 dark:text-white">
              {{ t('credits.ledger.title') }}
            </h1>
            <p class="mt-3 text-sm text-gray-500 dark:text-dark-300">
              {{ t('credits.ledger.description') }}
            </p>
          </div>
          <RouterLink to="/settings/credits" class="btn btn-secondary">
            {{ t('credits.ledger.backToCredits') }}
          </RouterLink>
        </div>
      </section>

      <!-- Filter -->
      <section
        class="rounded-2xl border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-900"
      >
        <div class="flex flex-wrap items-center gap-3">
          <select v-model="selectedEntryType" class="form-select rounded-lg border-gray-300 dark:border-dark-600 dark:bg-dark-800">
            <option value="">{{ t('credits.ledger.filter.allTypes') }}</option>
            <option v-for="type in entryTypes" :key="type" :value="type">
              {{ entryTypeDisplayName[type] }}
            </option>
          </select>

          <button @click="loadLedger" class="btn btn-primary">
            {{ t('common.filter') }}
          </button>

          <button v-if="selectedEntryType" @click="resetFilter" class="btn btn-secondary">
            {{ t('common.reset') }}
          </button>
        </div>
      </section>

      <!-- Loading -->
      <div v-if="loading" class="flex justify-center py-8">
        <div class="animate-spin h-8 w-8 rounded-full border-4 border-primary-600 border-t-transparent"></div>
      </div>

      <!-- Empty State -->
      <section
        v-else-if="entries.length === 0"
        class="rounded-2xl border border-gray-200 bg-white p-8 shadow-sm dark:border-dark-700 dark:bg-dark-900 text-center"
      >
        <p class="text-gray-500 dark:text-dark-300">{{ t('credits.ledger.empty') }}</p>
      </section>

      <!-- Ledger Table -->
      <section
        v-else
        class="rounded-2xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900 overflow-hidden"
      >
        <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-700">
          <thead class="bg-gray-50 dark:bg-dark-800">
            <tr>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-dark-300 uppercase tracking-wider">
                {{ t('credits.ledger.col.time') }}
              </th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-dark-300 uppercase tracking-wider">
                {{ t('credits.ledger.col.type') }}
              </th>
              <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-dark-300 uppercase tracking-wider">
                {{ t('credits.ledger.col.amount') }}
              </th>
              <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-dark-300 uppercase tracking-wider">
                {{ t('credits.ledger.col.before') }}
              </th>
              <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-dark-300 uppercase tracking-wider">
                {{ t('credits.ledger.col.after') }}
              </th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-dark-300 uppercase tracking-wider">
                {{ t('credits.ledger.col.description') }}
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200 dark:divide-dark-700">
            <tr v-for="entry in entries" :key="entry.id" class="hover:bg-gray-50 dark:hover:bg-dark-800">
              <td class="px-4 py-3 text-sm text-gray-500 dark:text-dark-300 whitespace-nowrap">
                {{ formatTime(entry.created_at) }}
              </td>
              <td class="px-4 py-3 text-sm whitespace-nowrap">
                <span :class="getEntryTypeBadgeClass(entry.entry_type)">
                  {{ entryTypeDisplayName[entry.entry_type] }}
                </span>
              </td>
              <td class="px-4 py-3 text-sm text-right whitespace-nowrap font-medium" :class="getAmountClass(entry.amount)">
                {{ formatLedgerAmount(entry.amount) }}
              </td>
              <td class="px-4 py-3 text-sm text-right text-gray-500 dark:text-dark-300 whitespace-nowrap">
                {{ formatBalanceSnapshot(entry.balance_before) }}
              </td>
              <td class="px-4 py-3 text-sm text-right text-gray-500 dark:text-dark-300 whitespace-nowrap">
                {{ formatBalanceSnapshot(entry.balance_after) }}
              </td>
              <td class="px-4 py-3 text-sm text-gray-500 dark:text-dark-300">
                {{ entry.description || '-' }}
              </td>
            </tr>
          </tbody>
        </table>

        <!-- Pagination -->
        <div class="flex items-center justify-between px-4 py-3 border-t border-gray-200 dark:border-dark-700">
          <div class="text-sm text-gray-500 dark:text-dark-300">
            {{ t('pagination.showing') }} {{ startItem }}-{{ endItem }} {{ t('pagination.of') }} {{ total }} {{ t('pagination.results') }}
          </div>
          <div class="flex gap-2">
            <button
              :disabled="page <= 1"
              @click="prevPage"
              class="btn btn-secondary btn-sm"
              :class="{ 'opacity-50 cursor-not-allowed': page <= 1 }"
            >
              {{ t('pagination.prev') }}
            </button>
            <button
              :disabled="page * pageSize >= total"
              @click="nextPage"
              class="btn btn-secondary btn-sm"
              :class="{ 'opacity-50 cursor-not-allowed': page * pageSize >= total }"
            >
              {{ t('pagination.next') }}
            </button>
          </div>
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { RouterLink } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import {
  getUserBalanceLedger,
  type BalanceLedgerEntry,
  type BalanceLedgerEntryType,
  entryTypeDisplayName,
  entryTypeColor,
  formatLedgerAmount,
  formatBalanceSnapshot,
} from '@/api/userBalanceLedger'

const { t } = useI18n()

// State
const entries = ref<BalanceLedgerEntry[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(true)
const selectedEntryType = ref<BalanceLedgerEntryType | ''>('')

// Entry types for filter dropdown
const entryTypes = computed(() => {
  return Object.keys(entryTypeDisplayName) as BalanceLedgerEntryType[]
})

// Pagination computed
const startItem = computed(() => (page.value - 1) * pageSize.value + 1)
const endItem = computed(() => Math.min(page.value * pageSize.value, total.value))

// Load ledger data
async function loadLedger() {
  loading.value = true
  try {
    const filter = {
      page: page.value,
      page_size: pageSize.value,
      entry_types: selectedEntryType.value ? [selectedEntryType.value as BalanceLedgerEntryType] : undefined,
    }
    const response = await getUserBalanceLedger(filter)
    entries.value = response.entries
    total.value = response.total
  } catch (error) {
    console.error('[CreditsLedger] failed to load ledger', error)
  } finally {
    loading.value = false
  }
}

// Reset filter
function resetFilter() {
  selectedEntryType.value = ''
  page.value = 1
  loadLedger()
}

// Pagination
function prevPage() {
  if (page.value > 1) {
    page.value--
    loadLedger()
  }
}

function nextPage() {
  if (page.value * pageSize.value < total.value) {
    page.value++
    loadLedger()
  }
}

// Format time
function formatTime(time: string): string {
  const date = new Date(time)
  return date.toLocaleString()
}

// Get amount class
function getAmountClass(amount: number): string {
  return amount >= 0 ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'
}

// Get entry type badge class
function getEntryTypeBadgeClass(type: BalanceLedgerEntryType): string {
  const color = entryTypeColor[type]
  const classes: Record<string, string> = {
    green: 'inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200',
    red: 'inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200',
    blue: 'inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200',
    yellow: 'inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200',
    purple: 'inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200',
    gray: 'inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200',
    orange: 'inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200',
  }
  return classes[color] || classes.gray
}

// Initialize
onMounted(() => {
  loadLedger()
})
</script>