<template>
  <header class="public-dark-header relative z-20 px-6 py-5">
    <nav class="mx-auto flex max-w-7xl items-center justify-between">
      <RouterLink
        :to="authRouteDefaults.homePath"
        class="public-dark-header__brand flex min-w-0 items-center gap-3 rounded-full transition focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-sky-400"
        :aria-label="siteName"
      >
        <div v-if="siteLogo" class="public-dark-header__logo h-9 w-9 shrink-0 overflow-hidden rounded-xl border border-white/10 bg-white/5">
          <img :src="siteLogo" alt="Logo" class="h-full w-full object-contain" />
        </div>
        <span class="public-dark-header__site truncate text-sm font-semibold leading-none text-white">{{ siteName }}</span>
      </RouterLink>

      <div class="public-dark-header__actions flex min-w-0 items-center gap-3">
        <LocaleSwitcher />
        <button
          type="button"
          class="public-dark-header__theme-toggle inline-flex h-10 items-center gap-2 rounded-full border px-3 text-sm font-semibold leading-none transition"
          :aria-label="themeToggleLabel"
          @click="togglePublicTheme"
        >
          <Icon :name="isDarkTheme ? 'sun' : 'moon'" size="sm" :stroke-width="2" />
          <span class="hidden sm:inline">{{ themeToggleText }}</span>
        </button>
        <slot name="actions" />
        <RouterLink
          :to="isAuthenticated ? dashboardPath : loginPath"
          class="public-dark-header__account inline-flex h-10 items-center rounded-full border border-slate-900 bg-slate-900 px-5 text-sm font-semibold leading-none text-white transition hover:bg-slate-800"
        >
          {{ resolvedAccountLabel }}
        </RouterLink>
        <RouterLink
          v-if="isAuthenticated"
          :to="isAuthenticated ? dashboardPath : loginPath"
          class="public-dark-header__avatar flex h-10 w-10 shrink-0 items-center justify-center overflow-hidden rounded-full border border-slate-200 bg-white text-sm font-black text-slate-900 transition hover:border-slate-300"
          :aria-label="displayName"
        >
          <img
            v-if="avatarUrl"
            :src="avatarUrl"
            :alt="displayName"
            class="h-full w-full object-cover"
          />
          <span v-else>{{ userInitial }}</span>
        </RouterLink>
      </div>
    </nav>
  </header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import { Icon } from '@/components/icons'
import { useAuthRouteDefaults } from '@/composables/useAuthRouteDefaults'
import { usePublicTheme } from '@/composables/usePublicTheme'
import { useAppStore, useAuthStore } from '@/stores'

const props = defineProps<{
  accountLabel?: string
}>()

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const { authRouteDefaults, resolveHomePath } = useAuthRouteDefaults()
const { isDarkTheme, togglePublicTheme } = usePublicTheme()

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const isAuthenticated = computed(() => authStore.isAuthenticated)
const dashboardPath = computed(() => resolveHomePath(authStore.isAdmin))
const loginPath = computed(() => authRouteDefaults.value.loginPath)
const avatarUrl = computed(() => authStore.user?.avatar_url?.trim() || '')
const displayName = computed(() => {
  const user = authStore.user
  return user?.username?.trim() || user?.email?.trim() || siteName.value || 'User'
})
const userInitial = computed(() => displayName.value.charAt(0).toUpperCase() || 'U')
const resolvedAccountLabel = computed(() => props.accountLabel || (isAuthenticated.value ? t('nav.dashboard') : t('common.login')))
const themeToggleText = computed(() => (isDarkTheme.value ? t('nav.lightMode') : t('nav.darkMode')))
const themeToggleLabel = computed(() => themeToggleText.value)
</script>

<style scoped>
.public-dark-header {
  background: rgba(255, 255, 255, 0.9);
  border-bottom: 1px solid var(--vercel-hairline, #ebebeb);
  box-shadow: inset 0 -1px 0 rgba(0, 0, 0, 0.04);
  backdrop-filter: blur(18px);
}

.public-dark-header__brand {
  color: var(--vercel-ink, #171717);
}

.public-dark-header__logo {
  background: #050505 !important;
  border-color: rgba(0, 0, 0, 0.08) !important;
  box-shadow: 0 8px 18px -14px rgba(0, 0, 0, 0.48);
}

.public-dark-header__site {
  color: var(--vercel-ink, #171717) !important;
}

.public-dark-header__account,
.public-dark-header__theme-toggle,
.public-dark-header__actions :deep(a) {
  min-height: 40px;
  border-color: #e5e7eb !important;
  background: rgba(255, 255, 255, 0.78) !important;
  color: #4b5563 !important;
  box-shadow: 0 1px 1px rgba(0, 0, 0, 0.02);
}

.public-dark-header__actions .public-dark-header__account {
  border-color: #111827 !important;
  background: #111827 !important;
  color: #fff !important;
}

.public-dark-header__account:hover,
.public-dark-header__theme-toggle:hover,
.public-dark-header__actions :deep(a:hover) {
  border-color: #d4d4d4 !important;
  background: #fff !important;
}

.public-dark-header__actions .public-dark-header__account:hover {
  border-color: #1f2937 !important;
  background: #1f2937 !important;
  color: #fff !important;
}

.public-dark-header__avatar {
  background: #fff !important;
  border-color: var(--vercel-hairline, #ebebeb) !important;
  color: var(--vercel-ink, #171717) !important;
  box-shadow: 0 1px 1px rgba(0, 0, 0, 0.02);
}

.dark .public-dark-header {
  background: rgba(5, 5, 5, 0.9);
  border-bottom-color: rgba(255, 255, 255, 0.12);
}

.dark .public-dark-header__site {
  color: #fff !important;
}

.dark .public-dark-header__actions :deep(a),
.dark .public-dark-header__theme-toggle,
.dark .public-dark-header__avatar {
  border-color: rgba(255, 255, 255, 0.12) !important;
  background: rgba(255, 255, 255, 0.06) !important;
  color: #fff !important;
}

.dark .public-dark-header__actions :deep(a:hover),
.dark .public-dark-header__theme-toggle:hover,
.dark .public-dark-header__avatar:hover {
  background: rgba(255, 255, 255, 0.1) !important;
}

.dark .public-dark-header__actions .public-dark-header__account {
  border-color: #fff !important;
  background: #fff !important;
  color: #0f172a !important;
}

.dark .public-dark-header__actions .public-dark-header__account:hover {
  background: #f1f5f9 !important;
}

@media (max-width: 640px) {
  .public-dark-header {
    padding-inline: 1rem;
  }

  .public-dark-header__actions {
    gap: 0.5rem;
  }

  .public-dark-header__account,
  .public-dark-header__theme-toggle,
  .public-dark-header__actions :deep(a) {
    padding-inline: 0.875rem;
  }
}
</style>
