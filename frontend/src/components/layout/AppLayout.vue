<template>
  <div class="vercel-app-shell min-h-screen overflow-x-hidden bg-white text-slate-950 dark:bg-dark-950 dark:text-white">
    <div class="pointer-events-none fixed inset-0 overflow-hidden">
      <div class="absolute inset-x-0 top-0 h-[28rem] bg-[radial-gradient(circle_at_20%_55%,rgba(59,130,246,0.16),transparent_32%),radial-gradient(circle_at_82%_12%,rgba(96,165,250,0.14),transparent_28%),linear-gradient(180deg,rgba(239,246,255,0.82),rgba(255,255,255,0.9))] dark:bg-[radial-gradient(circle_at_20%_55%,rgba(59,130,246,0.14),transparent_32%),radial-gradient(circle_at_82%_12%,rgba(96,165,250,0.1),transparent_28%),linear-gradient(180deg,rgba(15,23,42,0.9),rgba(2,6,23,1))]"></div>
      <div class="absolute inset-0 bg-[linear-gradient(rgba(148,163,184,0.07)_1px,transparent_1px),linear-gradient(90deg,rgba(148,163,184,0.07)_1px,transparent_1px)] bg-[size:72px_72px] opacity-35 dark:opacity-10"></div>
    </div>
    <a
      href="#app-main"
      class="sr-only focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-[70] focus:rounded-xl focus:bg-white focus:px-4 focus:py-2 focus:text-sm focus:font-medium focus:text-primary-700 focus:shadow-lg dark:focus:bg-dark-900 dark:focus:text-primary-200"
    >
      Skip to content
    </a>

    <!-- Sidebar -->
    <AppSidebar />

    <!-- Main Content Area -->
    <div
      class="relative min-h-screen transition-[margin] duration-300"
      :class="[sidebarCollapsed ? 'lg:ml-[72px]' : 'lg:ml-64']"
    >
      <!-- Header -->
      <AppHeader :container-class="headerContainerClass" />

      <!-- Main Content -->
      <main id="app-main" class="relative p-4 md:p-6 lg:p-8 safe-bottom">
        <div class="mx-auto w-full" :class="contentContainerClass">
          <slot />
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAppStore } from '@/stores'
import AppSidebar from './AppSidebar.vue'
import AppHeader from './AppHeader.vue'

withDefaults(defineProps<{
  contentContainerClass?: string
  headerContainerClass?: string
}>(), {
  contentContainerClass: 'max-w-[1680px]',
  headerContainerClass: 'max-w-[1680px]',
})

const appStore = useAppStore()
const sidebarCollapsed = computed(() => appStore.sidebarCollapsed)
</script>
