<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="card p-6">
        <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h2 class="text-xl font-semibold text-gray-900 dark:text-white">
              {{ t('nav.affiliateCodeManagement') }}
            </h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
              {{ t('admin.affiliates.codesDescription') }}
            </p>
          </div>
          <button class="btn btn-primary" @click="openAffiliateModal(null)">
            + {{ t('admin.settings.features.affiliate.customUsers.addButton') }}
          </button>
        </div>
      </div>

      <div class="card p-6">
        <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
          <div class="relative w-full md:max-w-md">
            <Icon name="search" size="md" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
            <input
              v-model="affiliateState.search"
              type="text"
              class="input pl-10"
              :placeholder="t('admin.settings.features.affiliate.customUsers.searchPlaceholder')"
              @input="onAffiliateSearchInput"
            />
          </div>
          <button
            v-if="affiliateState.selected.length > 0"
            class="btn btn-secondary"
            @click="openAffiliateBatchModal"
          >
            {{ t('admin.settings.features.affiliate.customUsers.batchButton', { count: affiliateState.selected.length }) }}
          </button>
        </div>

        <div class="mt-4 overflow-x-auto">
          <table class="w-full min-w-[760px] border-collapse text-sm">
            <thead>
              <tr class="border-b border-gray-200 text-left dark:border-dark-700">
                <th class="px-3 py-2">
                  <input
                    type="checkbox"
                    :checked="affiliateState.entries.length > 0 && affiliateState.selected.length === affiliateState.entries.length"
                    @change="toggleAffiliateSelectAll"
                  />
                </th>
                <th class="px-3 py-2 text-xs font-medium uppercase text-gray-500">{{ t('admin.settings.features.affiliate.customUsers.col.email') }}</th>
                <th class="px-3 py-2 text-xs font-medium uppercase text-gray-500">{{ t('admin.settings.features.affiliate.customUsers.col.username') }}</th>
                <th class="px-3 py-2 text-xs font-medium uppercase text-gray-500">{{ t('admin.settings.features.affiliate.customUsers.col.code') }}</th>
                <th class="px-3 py-2 text-xs font-medium uppercase text-gray-500">{{ t('admin.settings.features.affiliate.customUsers.col.rate') }}</th>
                <th class="px-3 py-2 text-xs font-medium uppercase text-gray-500">{{ t('admin.settings.features.affiliate.customUsers.col.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="affiliateState.loading">
                <td colspan="6" class="px-3 py-8 text-center text-gray-500">{{ t('common.loading') }}</td>
              </tr>
              <tr v-else-if="affiliateState.entries.length === 0">
                <td colspan="6" class="px-3 py-8 text-center text-gray-500">{{ t('admin.settings.features.affiliate.customUsers.empty') }}</td>
              </tr>
              <tr v-for="entry in affiliateState.entries" :key="entry.user_id" class="border-b border-gray-100 dark:border-dark-800">
                <td class="px-3 py-3">
                  <input
                    type="checkbox"
                    :checked="affiliateState.selected.includes(entry.user_id)"
                    @change="toggleAffiliateSelect(entry.user_id)"
                  />
                </td>
                <td class="px-3 py-3 text-gray-900 dark:text-white">{{ entry.email || '-' }}</td>
                <td class="px-3 py-3 text-gray-700 dark:text-gray-300">{{ entry.username || '-' }}</td>
                <td class="px-3 py-3">
                  <div class="flex items-center gap-2">
                    <code class="font-mono text-sm text-gray-900 dark:text-white">{{ entry.aff_code }}</code>
                    <span
                      v-if="entry.aff_code_custom"
                      class="rounded-full bg-primary-50 px-2 py-0.5 text-xs font-medium text-primary-700 dark:bg-primary-900/30 dark:text-primary-300"
                    >
                      {{ t('admin.settings.features.affiliate.customUsers.customBadge') }}
                    </span>
                  </div>
                </td>
                <td class="px-3 py-3 text-gray-700 dark:text-gray-300">
                  <span v-if="entry.aff_rebate_rate_percent != null">{{ entry.aff_rebate_rate_percent }}%</span>
                  <span v-else class="text-gray-400">{{ t('admin.settings.features.affiliate.customUsers.useGlobal') }}</span>
                </td>
                <td class="px-3 py-3">
                  <div class="flex flex-wrap gap-2">
                    <button class="btn btn-secondary btn-sm" @click="openAffiliateModal(entry)">{{ t('common.edit') }}</button>
                    <button class="btn btn-secondary btn-sm" @click="askResetAffiliateUser(entry)">{{ t('common.reset') }}</button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div v-if="affiliateState.total > affiliateState.pageSize" class="mt-3 flex items-center justify-between text-sm">
          <span class="text-gray-500">{{ t('admin.settings.features.affiliate.customUsers.totalLabel', { total: affiliateState.total }) }}</span>
          <div class="flex items-center gap-2">
            <button class="btn btn-secondary btn-sm" :disabled="affiliateState.page <= 1" @click="changeAffiliatePage(affiliateState.page - 1)">‹</button>
            <span class="text-gray-500">{{ affiliateState.page }} / {{ Math.max(1, Math.ceil(affiliateState.total / affiliateState.pageSize)) }}</span>
            <button class="btn btn-secondary btn-sm" :disabled="affiliateState.page >= Math.ceil(affiliateState.total / affiliateState.pageSize)" @click="changeAffiliatePage(affiliateState.page + 1)">›</button>
          </div>
        </div>
      </div>

      <div v-if="affiliateModal.open" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 px-4" @click.self="closeAffiliateModal">
        <div class="w-full max-w-lg rounded-2xl bg-white p-6 shadow-2xl dark:bg-dark-900">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ affiliateModal.mode === 'add' ? t('admin.settings.features.affiliate.modal.addTitle') : t('admin.settings.features.affiliate.modal.editTitle') }}
          </h3>

          <div class="mt-5 space-y-4">
            <div v-if="affiliateModal.mode === 'add'">
              <label class="input-label">{{ t('admin.settings.features.affiliate.modal.userLabel') }}</label>
              <div
                v-if="affiliateModal.selectedUser"
                class="mb-2 flex items-center justify-between rounded-xl border border-gray-200 bg-gray-50 px-3 py-2 dark:border-dark-700 dark:bg-dark-800"
              >
                <div>
                  <span class="font-medium text-gray-900 dark:text-white">{{ affiliateModal.selectedUser.email }}</span>
                  <span class="ml-1 text-xs text-gray-500">({{ affiliateModal.selectedUser.username }})</span>
                </div>
                <button type="button" class="btn btn-secondary btn-sm" :title="t('admin.settings.features.affiliate.modal.changeUser')" @click="clearSelectedAffiliateUser">
                  {{ t('common.change') }}
                </button>
              </div>
              <input
                v-if="!affiliateModal.selectedUser"
                v-model="affiliateModal.userQuery"
                type="text"
                class="input"
                :placeholder="t('admin.settings.features.affiliate.modal.userPlaceholder')"
                @input="onAffiliateUserSearchInput"
              />
              <div v-if="affiliateModal.userResults.length > 0" class="mt-2 max-h-48 overflow-auto rounded-xl border border-gray-200 dark:border-dark-700">
                <button
                  v-for="u in affiliateModal.userResults"
                  :key="u.id"
                  type="button"
                  class="flex w-full items-center justify-between px-3 py-2 text-left hover:bg-gray-50 dark:hover:bg-dark-800"
                  @click="selectAffiliateUser(u)"
                >
                  <span class="text-sm text-gray-900 dark:text-white">{{ u.email }}</span>
                  <span class="text-xs text-gray-500">{{ u.username }}</span>
                </button>
              </div>
            </div>

            <div v-else>
              <label class="input-label">{{ t('admin.settings.features.affiliate.modal.userLabel') }}</label>
              <input class="input" :value="affiliateModal.editingEntry ? affiliateModal.editingEntry.email : ''" disabled />
            </div>

            <div>
              <label class="input-label">{{ t('admin.settings.features.affiliate.modal.codeLabel') }}</label>
              <input
                v-model="affiliateModal.code"
                type="text"
                class="input"
                :placeholder="t('admin.settings.features.affiliate.modal.codePlaceholder')"
              />
              <p class="mt-1 text-xs text-gray-400">{{ t('admin.settings.features.affiliate.modal.codeHint') }}</p>
            </div>

            <div>
              <label class="input-label">{{ t('admin.settings.features.affiliate.modal.rateLabel') }}</label>
              <div class="relative">
                <input
                  v-model="affiliateModal.rate"
                  type="number"
                  min="0"
                  max="100"
                  step="0.01"
                  class="input pr-8"
                  :placeholder="t('admin.settings.features.affiliate.modal.ratePlaceholder')"
                />
                <span class="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-gray-400">%</span>
              </div>
              <p class="mt-1 text-xs text-gray-400">{{ t('admin.settings.features.affiliate.modal.rateHint') }}</p>
            </div>

            <div v-if="!affiliateModalCanSubmit" class="rounded-xl bg-amber-50 px-3 py-2 text-sm text-amber-700 dark:bg-amber-900/20 dark:text-amber-300">
              {{ t('admin.settings.features.affiliate.modal.errorEmpty') }}
            </div>
          </div>

          <div class="mt-6 flex justify-end gap-2">
            <button type="button" class="btn btn-secondary" @click="closeAffiliateModal">{{ t('common.cancel') }}</button>
            <button type="button" class="btn btn-primary" :disabled="affiliateModal.saving || !affiliateModalCanSubmit" @click="submitAffiliateModal">
              {{ affiliateModal.saving ? t('common.saving') : t('common.save') }}
            </button>
          </div>
        </div>
      </div>

      <div v-if="affiliateBatchModal.open" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 px-4" @click.self="affiliateBatchModal.open = false">
        <div class="w-full max-w-md rounded-2xl bg-white p-6 shadow-2xl dark:bg-dark-900">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('admin.settings.features.affiliate.batchModal.title', { count: affiliateState.selected.length }) }}
          </h3>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('admin.settings.features.affiliate.batchModal.hint') }}</p>
          <div class="mt-5">
            <input
              v-model="affiliateBatchModal.rate"
              type="number"
              min="0"
              max="100"
              step="0.01"
              class="input"
              :placeholder="t('admin.settings.features.affiliate.batchModal.placeholder')"
            />
            <p class="mt-1 text-xs text-gray-400">{{ t('admin.settings.features.affiliate.batchModal.clearHint') }}</p>
          </div>
          <div class="mt-6 flex justify-end gap-2">
            <button type="button" class="btn btn-secondary" @click="affiliateBatchModal.open = false">{{ t('common.cancel') }}</button>
            <button type="button" class="btn btn-primary" :disabled="affiliateBatchModal.saving" @click="submitAffiliateBatchModal">
              {{ affiliateBatchModal.saving ? t('common.saving') : t('common.save') }}
            </button>
          </div>
        </div>
      </div>

      <ConfirmDialog
        :show="affiliateConfirmDialog.show"
        :title="affiliateConfirmDialog.title"
        :message="affiliateConfirmDialog.message"
        :confirm-text="affiliateConfirmDialog.confirmText"
        @confirm="handleAffiliateConfirm"
        @cancel="cancelAffiliateConfirm"
      />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { affiliatesAPI, type AffiliateAdminEntry, type SimpleUser as AffiliateSimpleUser } from '@/api/admin/affiliates'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'

const { t } = useI18n()
const appStore = useAppStore()

interface AffiliateState {
  loading: boolean
  entries: AffiliateAdminEntry[]
  page: number
  pageSize: number
  total: number
  search: string
  selected: number[]
  searchTimer: number | null
}

interface AffiliateModalState {
  open: boolean
  mode: 'add' | 'edit'
  userQuery: string
  userResults: AffiliateSimpleUser[]
  selectedUser: AffiliateSimpleUser | null
  editingEntry: AffiliateAdminEntry | null
  code: string
  rate: string
  saving: boolean
  searchTimer: number | null
}

const affiliateState = reactive<AffiliateState>({
  loading: false,
  entries: [],
  page: 1,
  pageSize: 20,
  total: 0,
  search: '',
  selected: [],
  searchTimer: null,
})

const affiliateModal = reactive<AffiliateModalState>({
  open: false,
  mode: 'add',
  userQuery: '',
  userResults: [],
  selectedUser: null,
  editingEntry: null,
  code: '',
  rate: '',
  saving: false,
  searchTimer: null,
})

const affiliateBatchModal = reactive({
  open: false,
  rate: '',
  saving: false,
})

const affiliateConfirmDialog = reactive({
  show: false,
  title: '',
  message: '',
  confirmText: '',
  pending: null as null | (() => Promise<void>),
})

function openAffiliateConfirm(title: string, message: string, confirmText: string, fn: () => Promise<void>) {
  affiliateConfirmDialog.show = true
  affiliateConfirmDialog.title = title
  affiliateConfirmDialog.message = message
  affiliateConfirmDialog.confirmText = confirmText
  affiliateConfirmDialog.pending = fn
}

async function handleAffiliateConfirm() {
  const fn = affiliateConfirmDialog.pending
  affiliateConfirmDialog.show = false
  affiliateConfirmDialog.pending = null
  if (!fn) return
  try {
    await fn()
    appStore.showSuccess(t('common.saved'))
    await loadAffiliateUsers()
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  }
}

function cancelAffiliateConfirm() {
  affiliateConfirmDialog.show = false
  affiliateConfirmDialog.pending = null
}

function debounceTimer(slot: { searchTimer: number | null }, delayMs: number, run: () => void) {
  if (slot.searchTimer != null) window.clearTimeout(slot.searchTimer)
  slot.searchTimer = window.setTimeout(run, delayMs)
}

function parseRebateRate(raw: unknown): number | null | undefined {
  const s = String(raw ?? '').trim()
  if (s === '') return null
  const parsed = Number(s)
  if (Number.isNaN(parsed) || parsed < 0 || parsed > 100) {
    appStore.showError(t('admin.settings.features.affiliate.modal.errorBadRate'))
    return undefined
  }
  return parsed
}

async function loadAffiliateUsers() {
  affiliateState.loading = true
  try {
    const res = await affiliatesAPI.listUsers({
      page: affiliateState.page,
      page_size: affiliateState.pageSize,
      search: affiliateState.search,
    })
    affiliateState.entries = res.items ?? []
    affiliateState.total = res.total ?? 0
    const visibleIds = new Set(affiliateState.entries.map((e) => e.user_id))
    affiliateState.selected = affiliateState.selected.filter((id) => visibleIds.has(id))
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    affiliateState.loading = false
  }
}

function onAffiliateSearchInput() {
  debounceTimer(affiliateState, 300, () => {
    affiliateState.page = 1
    void loadAffiliateUsers()
  })
}

function changeAffiliatePage(page: number) {
  if (page < 1) return
  affiliateState.page = page
  void loadAffiliateUsers()
}

function toggleAffiliateSelectAll(e: Event) {
  const checked = (e.target as HTMLInputElement).checked
  affiliateState.selected = checked ? affiliateState.entries.map((entry) => entry.user_id) : []
}

function toggleAffiliateSelect(userId: number) {
  const idx = affiliateState.selected.indexOf(userId)
  if (idx >= 0) affiliateState.selected.splice(idx, 1)
  else affiliateState.selected.push(userId)
}

function openAffiliateModal(entry: AffiliateAdminEntry | null) {
  affiliateModal.open = true
  affiliateModal.mode = entry ? 'edit' : 'add'
  affiliateModal.userQuery = ''
  affiliateModal.userResults = []
  affiliateModal.selectedUser = null
  affiliateModal.editingEntry = entry
  affiliateModal.code = entry?.aff_code_custom ? entry.aff_code : ''
  affiliateModal.rate = entry?.aff_rebate_rate_percent != null ? String(entry.aff_rebate_rate_percent) : ''
}

function closeAffiliateModal() {
  affiliateModal.open = false
  if (affiliateModal.searchTimer != null) {
    window.clearTimeout(affiliateModal.searchTimer)
    affiliateModal.searchTimer = null
  }
}

function onAffiliateUserSearchInput() {
  const q = affiliateModal.userQuery.trim()
  if (!q) {
    affiliateModal.userResults = []
    return
  }
  debounceTimer(affiliateModal, 300, async () => {
    try {
      affiliateModal.userResults = await affiliatesAPI.lookupUsers(q)
    } catch (err) {
      appStore.showError(extractApiErrorMessage(err, t('common.error')))
    }
  })
}

function selectAffiliateUser(user: AffiliateSimpleUser) {
  affiliateModal.selectedUser = user
  affiliateModal.userQuery = ''
  affiliateModal.userResults = []
}

function clearSelectedAffiliateUser() {
  affiliateModal.selectedUser = null
}

const affiliateModalCanSubmit = computed(() => {
  if (affiliateModal.mode === 'add') {
    if (!affiliateModal.selectedUser) return false
  } else if (!affiliateModal.editingEntry) {
    return false
  }
  const codeFilled = affiliateModal.code.trim() !== ''
  const rateFilled = String(affiliateModal.rate ?? '').trim() !== ''
  if (codeFilled || rateFilled) return true
  return affiliateModal.mode === 'edit' && affiliateModal.editingEntry?.aff_rebate_rate_percent != null
})

async function submitAffiliateModal() {
  if (!affiliateModalCanSubmit.value) {
    appStore.showError(t('admin.settings.features.affiliate.modal.errorEmpty'))
    return
  }

  const userId = affiliateModal.mode === 'add' ? affiliateModal.selectedUser!.id : affiliateModal.editingEntry!.user_id
  const payload: Parameters<typeof affiliatesAPI.updateUserSettings>[1] = {}
  const codeRaw = affiliateModal.code.trim()
  if (codeRaw) payload.aff_code = codeRaw.toUpperCase()

  const rateInput = parseRebateRate(affiliateModal.rate)
  if (rateInput === undefined) return
  if (rateInput === null) {
    if (affiliateModal.mode === 'edit' && affiliateModal.editingEntry?.aff_rebate_rate_percent != null) {
      payload.clear_rebate_rate = true
    }
  } else {
    payload.aff_rebate_rate_percent = rateInput
  }

  affiliateModal.saving = true
  try {
    await affiliatesAPI.updateUserSettings(userId, payload)
    appStore.showSuccess(t('common.saved'))
    closeAffiliateModal()
    affiliateState.page = 1
    await loadAffiliateUsers()
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    affiliateModal.saving = false
  }
}

function askResetAffiliateUser(entry: AffiliateAdminEntry) {
  openAffiliateConfirm(
    t('admin.settings.features.affiliate.customUsers.resetTitle'),
    t('admin.settings.features.affiliate.customUsers.resetMessage', {
      email: entry.email || `#${entry.user_id}`,
    }),
    t('common.delete'),
    async () => {
      await affiliatesAPI.clearUserSettings(entry.user_id)
    },
  )
}

function openAffiliateBatchModal() {
  if (affiliateState.selected.length === 0) return
  affiliateBatchModal.open = true
  affiliateBatchModal.rate = ''
}

async function submitAffiliateBatchModal() {
  const rateInput = parseRebateRate(affiliateBatchModal.rate)
  if (rateInput === undefined) return
  const userIDs = [...affiliateState.selected]
  const payload: Parameters<typeof affiliatesAPI.batchSetRate>[0] =
    rateInput === null ? { user_ids: userIDs, clear: true } : { user_ids: userIDs, aff_rebate_rate_percent: rateInput }

  affiliateBatchModal.saving = true
  try {
    await affiliatesAPI.batchSetRate(payload)
    appStore.showSuccess(t('common.saved'))
    affiliateBatchModal.open = false
    affiliateState.selected = []
    await loadAffiliateUsers()
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    affiliateBatchModal.saving = false
  }
}

onMounted(() => {
  void loadAffiliateUsers()
})
</script>
