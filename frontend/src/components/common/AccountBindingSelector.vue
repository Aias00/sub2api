<template>
  <div class="space-y-3">
    <div class="flex items-center justify-between gap-3 text-xs text-gray-500 dark:text-gray-400">
      <span>{{ t('common.selectedCount', { count: modelValue.length }) }}</span>
      <span v-if="loading">{{ t('common.loading') }}</span>
    </div>

    <input
      v-model.trim="keyword"
      type="text"
      class="input"
      :placeholder="placeholder"
    />

    <div class="max-h-64 space-y-2 overflow-y-auto rounded-xl border border-gray-200 bg-gray-50 p-2 dark:border-dark-600 dark:bg-dark-800/70">
      <label
        v-for="account in filteredAccounts"
        :key="account.id"
        class="flex cursor-pointer items-start gap-3 rounded-lg border border-transparent bg-white px-3 py-2.5 transition-colors hover:border-primary-200 hover:bg-primary-50/50 dark:bg-dark-700/80 dark:hover:border-primary-500/30 dark:hover:bg-dark-700"
      >
        <input
          type="checkbox"
          :value="account.id"
          :checked="modelValue.includes(account.id)"
          class="mt-0.5 h-4 w-4 shrink-0 rounded border-gray-300 text-primary-500 focus:ring-primary-500 dark:border-dark-500"
          @change="toggle(account.id, ($event.target as HTMLInputElement).checked)"
        />

        <div class="min-w-0 flex-1">
          <div class="truncate text-sm font-medium text-gray-900 dark:text-white">
            {{ account.name }}
          </div>
          <div class="mt-1 flex flex-wrap items-center gap-1.5 text-[11px]">
            <span class="inline-flex items-center gap-1 rounded-full bg-gray-100 px-2 py-0.5 text-gray-600 dark:bg-dark-600 dark:text-gray-300">
              <PlatformIcon :platform="account.platform" size="xs" />
              {{ t(`admin.groups.platforms.${account.platform}`) }}
            </span>
            <span
              :class="[
                'inline-flex rounded-full px-2 py-0.5',
                account.status === 'active'
                  ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300'
                  : account.status === 'inactive'
                    ? 'bg-amber-100 text-amber-700 dark:bg-amber-500/15 dark:text-amber-300'
                    : 'bg-red-100 text-red-700 dark:bg-red-500/15 dark:text-red-300'
              ]"
            >
              {{ t(`admin.accounts.status.${account.status}`) }}
            </span>
            <span
              v-if="account.mixedScheduling"
              class="inline-flex rounded-full bg-blue-100 px-2 py-0.5 text-blue-700 dark:bg-blue-500/15 dark:text-blue-300"
            >
              {{ t('admin.groups.directAccounts.mixedScheduling') }}
            </span>
          </div>
        </div>
      </label>

      <div
        v-if="filteredAccounts.length === 0"
        class="rounded-lg border border-dashed border-gray-300 px-4 py-6 text-center text-sm text-gray-500 dark:border-dark-600 dark:text-gray-400"
      >
        {{ emptyText }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import type { AccountPlatform } from '@/types'

interface AccountBindingOption {
  id: number
  name: string
  platform: AccountPlatform
  status: 'active' | 'inactive' | 'error'
  mixedScheduling: boolean
}

interface Props {
  modelValue: number[]
  accounts: AccountBindingOption[]
  loading?: boolean
  placeholder: string
  emptyText: string
}

const props = defineProps<Props>()
const emit = defineEmits<{
  'update:modelValue': [value: number[]]
}>()

const { t } = useI18n()
const keyword = ref('')

const filteredAccounts = computed(() => {
  const search = keyword.value.trim().toLowerCase()
  if (!search) {
    return props.accounts
  }
  return props.accounts.filter(account => {
    return (
      account.name.toLowerCase().includes(search) ||
      String(account.id).includes(search) ||
      account.platform.toLowerCase().includes(search)
    )
  })
})

function toggle(accountID: number, checked: boolean) {
  const nextValue = checked
    ? [...props.modelValue, accountID]
    : props.modelValue.filter(id => id !== accountID)
  emit('update:modelValue', Array.from(new Set(nextValue)))
}
</script>
