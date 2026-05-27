<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="card p-6">
        <h2 class="text-xl font-semibold text-gray-900 dark:text-white">
          {{ t('admin.affiliates.rules.title') }}
        </h2>
        <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
          {{ t('admin.affiliates.rules.description') }}
        </p>
      </div>

      <div v-if="loading" class="flex justify-center py-12">
        <div class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
      </div>

      <form v-else class="card space-y-6 p-6" @submit.prevent="saveRules">
        <div class="flex items-center justify-between">
          <div>
            <label class="font-medium text-gray-900 dark:text-white">
              {{ t('admin.settings.features.affiliate.enabled') }}
            </label>
            <p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">
              {{ t('admin.settings.features.affiliate.enabledHint') }}
            </p>
          </div>
          <Toggle v-model="form.affiliate_enabled" />
        </div>

        <div class="flex items-center justify-between border-t border-gray-100 pt-4 dark:border-dark-700">
          <div>
            <label class="font-medium text-gray-900 dark:text-white">
              {{ t('admin.settings.registration.invitationCode') }}
            </label>
            <p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">
              {{ t('admin.settings.registration.invitationCodeHint') }}
            </p>
          </div>
          <Toggle v-model="form.invitation_code_enabled" />
        </div>

        <div>
          <label class="input-label">{{ t('admin.settings.features.affiliate.rebateRate') }}</label>
          <div class="relative">
            <input v-model.number="form.affiliate_rebate_rate" type="number" min="0" max="100" step="0.01" class="input pr-8" />
            <span class="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-gray-400">%</span>
          </div>
          <p class="mt-1 text-xs text-gray-400">{{ t('admin.settings.features.affiliate.rebateRateHint') }}</p>
        </div>

        <div>
          <label class="input-label">{{ t('admin.settings.features.affiliate.freezeHours') }}</label>
          <input v-model.number="form.affiliate_rebate_freeze_hours" type="number" min="0" max="720" step="1" class="input" />
          <p class="mt-1 text-xs text-gray-400">{{ t('admin.settings.features.affiliate.freezeHoursDesc') }}</p>
        </div>

        <div>
          <label class="input-label">{{ t('admin.settings.features.affiliate.durationDays') }}</label>
          <input v-model.number="form.affiliate_rebate_duration_days" type="number" min="0" max="3650" step="1" class="input" />
          <p class="mt-1 text-xs text-gray-400">{{ t('admin.settings.features.affiliate.durationDaysDesc') }}</p>
        </div>

        <div>
          <label class="input-label">{{ t('admin.settings.features.affiliate.perInviteeCap') }}</label>
          <input v-model.number="form.affiliate_rebate_per_invitee_cap" type="number" min="0" step="0.01" class="input" />
          <p class="mt-1 text-xs text-gray-400">{{ t('admin.settings.features.affiliate.perInviteeCapDesc') }}</p>
        </div>

        <div class="flex justify-end">
          <button type="submit" class="btn btn-primary" :disabled="saving">
            {{ saving ? t('common.saving') : t('admin.affiliates.rules.save') }}
          </button>
        </div>
      </form>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Toggle from '@/components/common/Toggle.vue'
import { affiliatesAPI } from '@/api/admin/affiliates'
import type { AffiliateRulesSettings } from '@/types'
import { useAppStore } from '@/stores/app'
import { extractI18nErrorMessage } from '@/utils/apiError'

const { t } = useI18n()
const appStore = useAppStore()
const loading = ref(false)
const saving = ref(false)

const form = reactive<AffiliateRulesSettings>({
  affiliate_enabled: false,
  invitation_code_enabled: false,
  affiliate_rebate_rate: 20,
  affiliate_rebate_freeze_hours: 0,
  affiliate_rebate_duration_days: 0,
  affiliate_rebate_per_invitee_cap: 0,
})

async function loadRules() {
  loading.value = true
  try {
    Object.assign(form, await affiliatesAPI.getRules())
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'admin.affiliates.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

async function saveRules() {
  saving.value = true
  try {
    Object.assign(form, await affiliatesAPI.updateRules({
      affiliate_enabled: form.affiliate_enabled,
      invitation_code_enabled: form.invitation_code_enabled,
      affiliate_rebate_rate: Math.min(100, Math.max(0, Number(form.affiliate_rebate_rate) || 0)),
      affiliate_rebate_freeze_hours: Math.max(0, Math.min(720, Math.floor(Number(form.affiliate_rebate_freeze_hours) || 0))),
      affiliate_rebate_duration_days: Math.max(0, Math.min(3650, Math.floor(Number(form.affiliate_rebate_duration_days) || 0))),
      affiliate_rebate_per_invitee_cap: Math.max(0, Number(form.affiliate_rebate_per_invitee_cap) || 0),
    }))
    appStore.showSuccess(t('admin.affiliates.rules.saved'))
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'admin.affiliates.errors', t('common.error')))
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  void loadRules()
})
</script>
