<template>
  <div class="min-h-screen overflow-x-hidden bg-slate-50 dark:bg-[#050b18]">
    <a
      href="#app-main"
      class="sr-only focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-[70] focus:rounded-xl focus:bg-white focus:px-4 focus:py-2 focus:text-sm focus:font-medium focus:text-primary-700 focus:shadow-lg dark:focus:bg-dark-900 dark:focus:text-primary-200"
    >
      Skip to content
    </a>

    <!-- Background Decoration -->
    <div class="pointer-events-none fixed inset-0 overflow-hidden">
      <div class="absolute inset-0 bg-mesh-gradient opacity-95 dark:opacity-80"></div>
      <div class="absolute left-[-10rem] top-[-9rem] h-[28rem] w-[28rem] rounded-full bg-primary-300/28 blur-3xl dark:bg-primary-500/16"></div>
      <div class="absolute right-[-8rem] top-16 h-[24rem] w-[24rem] rounded-full bg-primary-200/18 blur-3xl dark:bg-primary-400/12"></div>
      <div class="absolute bottom-[-12rem] left-[24%] h-[26rem] w-[26rem] rounded-full bg-[#dcb959]/28 blur-3xl dark:bg-[#caac5e]/12"></div>
      <div class="absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-primary-400/40 to-transparent"></div>
    </div>

    <!-- Sidebar -->
    <AppSidebar />

    <!-- Main Content Area -->
    <div
      class="relative min-h-screen transition-[margin] duration-300"
      :class="[sidebarCollapsed ? 'lg:ml-[72px]' : 'lg:ml-64']"
    >
      <!-- Header -->
      <AppHeader />

      <!-- Main Content -->
      <main id="app-main" class="relative p-4 md:p-6 lg:p-8 safe-bottom">
        <div class="mx-auto w-full max-w-[1680px]">
          <slot />
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import '@/styles/onboarding.css'
import { computed, onMounted } from 'vue'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'
import { useOnboardingTour } from '@/composables/useOnboardingTour'
import { useOnboardingStore } from '@/stores/onboarding'
import AppSidebar from './AppSidebar.vue'
import AppHeader from './AppHeader.vue'

const appStore = useAppStore()
const authStore = useAuthStore()
const sidebarCollapsed = computed(() => appStore.sidebarCollapsed)
const isAdmin = computed(() => authStore.user?.role === 'admin')

const { replayTour } = useOnboardingTour({
  storageKey: isAdmin.value ? 'admin_guide' : 'user_guide',
  autoStart: true
})

const onboardingStore = useOnboardingStore()

onMounted(() => {
  onboardingStore.setReplayCallback(replayTour)
})

defineExpose({ replayTour })
</script>
