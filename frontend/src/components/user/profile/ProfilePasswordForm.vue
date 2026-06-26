<template>
  <div :class="props.embedded ? 'space-y-4' : 'card'">
    <div
      v-if="!props.embedded"
      class="border-b border-gray-100 px-6 py-4 dark:border-dark-700"
    >
      <h2 class="text-lg font-medium text-gray-900 dark:text-white">
        {{ profilePasswordText('changePassword') }}
      </h2>
    </div>
    <div :class="props.embedded ? '' : 'px-6 py-6'">
      <form @submit.prevent="handleChangePassword" class="space-y-4">
        <div v-if="props.embedded">
          <p class="text-sm font-semibold text-gray-900 dark:text-white">
            {{ profilePasswordText('changePassword') }}
          </p>
        </div>
        <div v-if="requiresCurrentPassword">
          <label for="old_password" class="input-label">
            {{ profilePasswordText('currentPassword') }}
          </label>
          <input
            id="old_password"
            v-model="form.old_password"
            type="password"
            required
            autocomplete="current-password"
            class="input"
          />
        </div>
        <div>
          <label for="new_password" class="input-label">
            {{ profilePasswordText('newPassword') }}
          </label>
          <input
            id="new_password"
            v-model="form.new_password"
            type="password"
            required
            autocomplete="new-password"
            class="input"
          />
          <p class="input-hint">
            {{ profilePasswordText('passwordHint', { count: passwordMinLength }) }}
          </p>
        </div>

        <div>
          <label for="confirm_password" class="input-label">
            {{ profilePasswordText('confirmNewPassword') }}
          </label>
          <input
            id="confirm_password"
            v-model="form.confirm_password"
            type="password"
            required
            autocomplete="new-password"
            class="input"
          />
        </div>

        <div class="flex justify-end pt-4">
          <button type="submit" :disabled="loading" class="btn btn-primary">
            {{ loading ? profilePasswordText('changingPassword') : profilePasswordText('changePasswordButton') }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ProfileLabelKey, ProfileLabels } from '@/utils/profileShell'
import { computed, ref } from 'vue'
import { useAppStore } from '@/stores/app'
import { userAPI } from '@/api'
import { resolvePasswordMinLength } from '@/utils/passwordPolicy'

const appStore = useAppStore()
const props = withDefaults(defineProps<{
  embedded?: boolean
  emailBound?: boolean
  labels?: ProfileLabels
}>(), {
  embedded: false,
  emailBound: true,
  labels: () => ({}),
})

const requiresCurrentPassword = computed(() => props.emailBound)
const passwordMinLength = computed(() =>
  resolvePasswordMinLength(appStore.cachedPublicSettings)
)

const loading = ref(false)
const form = ref({
  old_password: '',
  new_password: '',
  confirm_password: ''
})


function interpolateLabel(template: string, params?: Record<string, string | number>): string {
  if (!params) return template
  return template.replace(/\{(\w+)\}/g, (match, key) => {
    const value = params[key]
    return value === undefined ? match : String(value)
  })
}

function profilePasswordText(key: ProfileLabelKey, params?: Record<string, string | number>): string {
  const configured = props.labels?.[key]
  if (configured) {
    return interpolateLabel(configured, params)
  }
  return interpolateLabel(key, params)
}

const handleChangePassword = async () => {
  if (form.value.new_password !== form.value.confirm_password) {
    appStore.showError(profilePasswordText('passwordsNotMatch'))
    return
  }

  if (form.value.new_password.length < passwordMinLength.value) {
    appStore.showError(profilePasswordText('passwordTooShort', { count: passwordMinLength.value }))
    return
  }

  loading.value = true
  try {
    await userAPI.changePassword(form.value.old_password, form.value.new_password)
    form.value = { old_password: '', new_password: '', confirm_password: '' }
    appStore.showSuccess(profilePasswordText('passwordChangeSuccess'))
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || profilePasswordText('passwordChangeFailed'))
  } finally {
    loading.value = false
  }
}
</script>
