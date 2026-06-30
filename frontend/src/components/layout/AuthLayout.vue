<template>
  <div class="vercel-auth-shell relative flex min-h-screen items-center justify-center overflow-hidden bg-white p-4 text-slate-900 dark:bg-dark-950 dark:text-white">
    <!-- Background -->
    <div
      class="absolute inset-0 bg-[linear-gradient(180deg,rgba(239,246,255,0.9),rgba(255,255,255,0.96)_38%,rgba(255,255,255,1)),radial-gradient(circle_at_82%_10%,rgba(125,211,252,0.22),transparent_28%)] dark:bg-[linear-gradient(180deg,rgba(15,23,42,0.94),rgba(2,6,23,1))]"
    ></div>

    <!-- Decorative Elements -->
    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <!-- Grid Pattern -->
      <div
        class="absolute inset-0 bg-[linear-gradient(rgba(148,163,184,0.08)_1px,transparent_1px),linear-gradient(90deg,rgba(148,163,184,0.08)_1px,transparent_1px)] bg-[size:64px_64px] dark:opacity-20"
      ></div>
    </div>

    <!-- Content Container -->
    <div class="relative z-10 w-full max-w-md">
      <!-- Logo/Brand -->
      <div class="mb-8 text-center">
        <router-link
          v-if="settingsLoaded"
          :to="authRouteDefaults.homePath"
          class="inline-flex max-w-full flex-col items-center rounded-3xl transition focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-sky-400"
          :aria-label="siteName"
        >
          <!-- Custom Logo or Default Logo -->
          <div
            v-if="siteLogo"
            class="mb-4 inline-flex h-16 w-16 items-center justify-center overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-lg shadow-slate-900/10 dark:border-dark-700 dark:bg-dark-900"
          >
            <img :src="siteLogo" alt="Logo" class="h-full w-full object-contain" />
          </div>
          <h1 class="mb-2 max-w-full truncate text-3xl font-black tracking-tight text-slate-950 dark:text-white">
            {{ siteName }}
          </h1>
        </router-link>
      </div>

      <!-- Card Container -->
      <div class="rounded-2xl border border-slate-200/80 bg-white/90 p-8 shadow-[0_24px_70px_rgba(15,23,42,0.10)] backdrop-blur dark:border-dark-700 dark:bg-dark-900/90">
        <slot />
      </div>

      <!-- Footer Links -->
      <div class="mt-6 text-center text-sm">
        <slot name="footer" />
      </div>

      <!-- Copyright -->
      <div class="mt-8 text-center text-xs text-gray-400 dark:text-dark-500">
        &copy; {{ currentYear }} {{ siteName }}. {{ authText('allRightsReserved') }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useAuthShellText } from '@/composables/useAuthShellText'
import { useAuthRouteDefaults } from '@/composables/useAuthRouteDefaults'
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'

const appStore = useAppStore()
const { authText, loadAuthShellConfig } = useAuthShellText()
const { authRouteDefaults } = useAuthRouteDefaults()

const siteName = computed(() => appStore.siteName)
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const settingsLoaded = computed(() => appStore.publicSettingsLoaded)

const currentYear = computed(() => new Date().getFullYear())

onMounted(() => {
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
  void loadAuthShellConfig()
})
</script>
