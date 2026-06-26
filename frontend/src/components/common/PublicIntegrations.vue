<script setup lang="ts">
import { computed, onBeforeUnmount, watch } from 'vue'

import { useAppStore } from '@/stores'
import { applyPublicIntegrations, clearPublicIntegrations } from '@/utils/publicIntegrations'

const appStore = useAppStore()

const shouldInject = computed(() => {
  return appStore.cachedPublicSettings?.public_integrations_enabled !== false
})

watch(
  () => appStore.cachedPublicSettings,
  (settings) => {
    applyPublicIntegrations(settings, { enabled: shouldInject.value })
  },
  { immediate: true, deep: true }
)

onBeforeUnmount(() => {
  clearPublicIntegrations()
})
</script>

<template>
  <span v-if="false" aria-hidden="true"></span>
</template>
