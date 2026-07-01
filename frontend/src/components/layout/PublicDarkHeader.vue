<template>
  <header class="public-dark-header relative z-20 px-6 py-5">
    <nav class="mx-auto flex items-center justify-between" :class="containerClass">
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
          class="public-dark-header__theme-toggle inline-flex h-10 w-10 items-center justify-center rounded-full border text-sm font-semibold leading-none transition"
          :aria-label="themeToggleLabel"
          @click="togglePublicTheme"
        >
          <Icon :name="isDarkTheme ? 'sun' : 'moon'" size="sm" :stroke-width="2" />
        </button>
        <DocsLink
          :doc-url="docUrl"
          class="public-dark-header__docs hidden h-10 items-center rounded-full border px-4 text-sm font-semibold leading-none transition sm:inline-flex"
        >
          {{ t('nav.docs') }}
        </DocsLink>
        <slot name="actions" />
        <RouterLink
          v-if="!isAuthenticated"
          :to="loginPath"
          class="public-dark-header__account inline-flex h-10 items-center rounded-full border border-slate-900 bg-slate-900 px-5 text-sm font-semibold leading-none text-white transition hover:bg-slate-800"
        >
          {{ resolvedAccountLabel }}
        </RouterLink>
        <div v-if="isAuthenticated" ref="avatarMenuRef" class="public-dark-header__avatar-menu relative">
          <button
            type="button"
            class="public-dark-header__avatar flex h-10 w-10 shrink-0 items-center justify-center overflow-hidden rounded-full border border-slate-200 bg-white text-sm font-black text-slate-900 transition hover:border-slate-300"
            :aria-label="displayName"
            :aria-expanded="avatarMenuOpen"
            aria-haspopup="menu"
            @click="toggleAvatarMenu"
          >
            <img
              v-if="avatarUrl"
              :src="avatarUrl"
              :alt="displayName"
              class="h-full w-full object-cover"
            />
            <span v-else>{{ userInitial }}</span>
          </button>

          <transition name="public-dark-header-menu">
            <div
              v-if="avatarMenuOpen"
              class="public-dark-header__dropdown absolute right-0 mt-3 w-52 overflow-hidden rounded-2xl border p-1.5 shadow-xl"
              role="menu"
            >
              <RouterLink :to="dashboardPath" class="public-dark-header__dropdown-item" role="menuitem" @click="closeAvatarMenu">
                <Icon name="grid" size="sm" />
                <span>{{ t('nav.dashboard') }}</span>
              </RouterLink>
              <RouterLink to="/app/tasks" class="public-dark-header__dropdown-item" role="menuitem" @click="closeAvatarMenu">
                <Icon name="clipboard" size="sm" />
                <span>{{ t('nav.myTasks') }}</span>
              </RouterLink>
              <button
                type="button"
                class="public-dark-header__dropdown-item public-dark-header__dropdown-item-danger w-full"
                role="menuitem"
                @click="handleLogout"
              >
                <Icon name="arrowLeft" size="sm" />
                <span>{{ t('nav.logout') }}</span>
              </button>
            </div>
          </transition>
        </div>
      </div>
    </nav>
  </header>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import DocsLink from '@/components/common/DocsLink.vue'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import { Icon } from '@/components/icons'
import { useAuthRouteDefaults } from '@/composables/useAuthRouteDefaults'
import { usePublicTheme } from '@/composables/usePublicTheme'
import { useAppStore, useAuthStore } from '@/stores'

const props = defineProps<{
  accountLabel?: string
  containerClass?: string
}>()

const { t } = useI18n()
const router = useRouter()
const appStore = useAppStore()
const authStore = useAuthStore()
const { authRouteDefaults, resolveHomePath } = useAuthRouteDefaults()
const { isDarkTheme, togglePublicTheme } = usePublicTheme()
const avatarMenuOpen = ref(false)
const avatarMenuRef = ref<HTMLElement | null>(null)

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'cloudbase')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const isAuthenticated = computed(() => authStore.isAuthenticated)
const dashboardPath = computed(() => resolveHomePath(authStore.isAdmin))
const loginPath = computed(() => authRouteDefaults.value.loginPath)
const avatarUrl = computed(() => authStore.user?.avatar_url?.trim() || '')
const containerClass = computed(() => props.containerClass || 'max-w-7xl')
const displayName = computed(() => {
  const user = authStore.user
  return user?.username?.trim() || user?.email?.trim() || siteName.value || 'User'
})
const userInitial = computed(() => displayName.value.charAt(0).toUpperCase() || 'U')
const resolvedAccountLabel = computed(() => props.accountLabel || t('common.login'))
const themeToggleText = computed(() => (isDarkTheme.value ? t('nav.lightMode') : t('nav.darkMode')))
const themeToggleLabel = computed(() => themeToggleText.value)

function toggleAvatarMenu() {
  avatarMenuOpen.value = !avatarMenuOpen.value
}

function closeAvatarMenu() {
  avatarMenuOpen.value = false
}

async function handleLogout() {
  closeAvatarMenu()
  try {
    await authStore.logout()
  } catch (error) {
    console.error('Logout error:', error)
  }
  await router.push(authRouteDefaults.value.loginPath)
}

function handleDocumentClick(event: MouseEvent) {
  if (!avatarMenuRef.value?.contains(event.target as Node)) {
    closeAvatarMenu()
  }
}

function handleDocumentKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    closeAvatarMenu()
  }
}

onMounted(() => {
  document.addEventListener('click', handleDocumentClick)
  document.addEventListener('keydown', handleDocumentKeydown)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleDocumentClick)
  document.removeEventListener('keydown', handleDocumentKeydown)
})
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

.public-dark-header__dropdown {
  border-color: rgba(229, 231, 235, 0.96);
  background: rgba(255, 255, 255, 0.96);
  color: var(--vercel-ink, #171717);
  backdrop-filter: blur(18px);
}

.public-dark-header__dropdown-item {
  display: flex;
  min-height: 2.5rem;
  align-items: center;
  gap: 0.625rem;
  border-radius: 0.875rem;
  padding: 0.625rem 0.75rem;
  color: inherit;
  font-size: 0.875rem;
  font-weight: 650;
  line-height: 1;
  text-align: left;
  transition:
    background-color 160ms ease,
    color 160ms ease;
}

.public-dark-header__dropdown-item:hover {
  background: rgba(17, 24, 39, 0.06);
}

.public-dark-header__dropdown-item-danger {
  color: #dc2626;
}

.public-dark-header-menu-enter-active,
.public-dark-header-menu-leave-active {
  transition:
    opacity 160ms ease,
    transform 160ms ease;
}

.public-dark-header-menu-enter-from,
.public-dark-header-menu-leave-to {
  opacity: 0;
  transform: translateY(-0.25rem) scale(0.98);
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

.dark .public-dark-header__dropdown {
  border-color: rgba(255, 255, 255, 0.12);
  background: rgba(24, 24, 27, 0.96);
  color: #fff;
}

.dark .public-dark-header__dropdown-item:hover {
  background: rgba(255, 255, 255, 0.08);
}

.dark .public-dark-header__dropdown-item-danger {
  color: #f87171;
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
