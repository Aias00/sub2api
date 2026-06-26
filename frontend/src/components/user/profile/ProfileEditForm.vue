<template>
  <div :class="props.embedded ? 'space-y-4' : 'card'">
    <div
      v-if="!props.embedded"
      class="border-b border-gray-100 px-6 py-4 dark:border-dark-700"
    >
      <h2 class="text-lg font-medium text-gray-900 dark:text-white">
        {{ profileEditText('profileEditTitle') }}
      </h2>
    </div>
    <div :class="props.embedded ? '' : 'px-6 py-6'">
      <form @submit.prevent="handleUpdateProfile" class="space-y-4">
        <div v-if="props.embedded">
          <p class="text-sm font-semibold text-gray-900 dark:text-white">
            {{ profileEditText('profileEditTitle') }}
          </p>
        </div>
        <div>
          <label for="username" class="input-label">
            {{ profileEditText('profileUsername') }}
          </label>
          <input
            id="username"
            v-model="username"
            type="text"
            class="input"
            :placeholder="profileEditText('profileUsernamePlaceholder')"
          />
        </div>

        <div class="flex justify-end pt-4">
          <button type="submit" :disabled="loading" class="btn btn-primary">
            {{ loading ? profileEditText('profileUpdating') : profileEditText('profileUpdateAction') }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ProfileLabelKey, ProfileLabels } from '@/utils/profileShell'
import { ref, watch } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { userAPI } from '@/api'

const props = withDefaults(defineProps<{
  initialUsername: string
  embedded?: boolean
  labels?: ProfileLabels
}>(), {
  embedded: false,
  labels: () => ({}),
})

const authStore = useAuthStore()
const appStore = useAppStore()

const username = ref(props.initialUsername)
const loading = ref(false)


function profileEditText(key: ProfileLabelKey): string {
  return props.labels?.[key] || ''
}

watch(() => props.initialUsername, (val) => {
  username.value = val
})

const handleUpdateProfile = async () => {
  if (!username.value.trim()) {
    appStore.showError(profileEditText('profileUsernameRequired'))
    return
  }

  loading.value = true
  try {
    const updatedUser = await userAPI.updateProfile({
      username: username.value
    })
    authStore.user = updatedUser
    appStore.showSuccess(profileEditText('profileUpdateSuccess'))
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || profileEditText('profileUpdateFailed'))
  } finally {
    loading.value = false
  }
}
</script>
