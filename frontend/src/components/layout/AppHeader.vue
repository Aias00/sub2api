<template>
  <header class="glass sticky top-0 z-30 border-b border-gray-200/50 px-4 md:px-6 lg:px-8 dark:border-dark-700/50">
    <div class="mx-auto flex h-16 w-full items-center justify-between" :class="containerClass">
      <!-- Left: Mobile Menu Toggle + Page Title -->
      <div class="flex items-center gap-4">
        <button
          @click="toggleMobileSidebar"
          class="btn-ghost btn-icon lg:hidden"
          aria-label="Toggle Menu"
        >
          <Icon name="menu" size="md" />
        </button>

        <div class="hidden lg:block">
          <h1 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ pageTitle }}
          </h1>
        </div>
      </div>

      <!-- Right: Announcements + Docs + Language + Subscriptions + Balance + User Dropdown -->
      <div class="flex items-center gap-3">
        <!-- Announcement Bell -->
        <AnnouncementBell v-if="user" />

        <!-- Docs Link -->
        <DocsLink
          :doc-url="docUrl"
          class="group relative inline-flex h-9 w-9 items-center justify-center rounded-xl text-gray-600 transition-colors hover:bg-gray-100 hover:text-gray-900 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white"
          :aria-label="t('nav.docs')"
          :title="t('nav.docs')"
        >
          <Icon name="book" size="sm" />
          <span
            class="pointer-events-none absolute left-1/2 top-full z-20 mt-2 -translate-x-1/2 whitespace-nowrap rounded-lg bg-gray-950 px-2.5 py-1 text-xs font-medium text-white opacity-0 shadow-lg transition duration-150 group-hover:opacity-100 dark:bg-white dark:text-gray-900"
          >
            {{ t('nav.docs') }}
          </span>
        </DocsLink>

        <!-- Language Switcher -->
        <LocaleSwitcher />

        <!-- Subscription Progress (for users with active subscriptions) -->
        <SubscriptionProgressMini v-if="user" />

        <!-- User Dropdown -->
        <div v-if="user" class="relative" ref="dropdownRef">
          <button
            @click="toggleDropdown"
            class="flex h-10 w-10 shrink-0 items-center justify-center overflow-hidden rounded-full border border-gray-200 bg-white text-sm font-black text-gray-900 shadow-sm transition hover:border-gray-300 hover:bg-gray-50 dark:border-dark-700 dark:bg-dark-800 dark:text-white dark:hover:border-dark-600 dark:hover:bg-dark-700"
            :aria-label="displayName"
            :aria-expanded="dropdownOpen"
            aria-haspopup="menu"
          >
            <img
              v-if="avatarUrl"
              :src="avatarUrl"
              :alt="displayName"
              class="h-full w-full object-cover"
            >
            <span v-else>{{ userInitials }}</span>
          </button>

          <!-- Dropdown Menu -->
          <transition name="dropdown">
            <div v-if="dropdownOpen" class="app-header-user-dropdown" role="menu">
              <div v-if="showDropdownPrimaryActions" class="app-header-user-dropdown-group">
                <router-link :to="authRouteDefaults.userRedirectPath" @click="closeDropdown" class="app-header-user-dropdown-item" role="menuitem">
                  <Icon name="grid" size="sm" />
                  {{ t('nav.dashboard') }}
                </router-link>

                <router-link to="/app/tasks" @click="closeDropdown" class="app-header-user-dropdown-item" role="menuitem">
                  <Icon name="clipboard" size="sm" />
                  {{ t('nav.myTasks') }}
                </router-link>

                <template v-if="showDropdownAccountLinks">
                  <router-link :to="authRouteDefaults.profilePath" @click="closeDropdown" class="app-header-user-dropdown-item" role="menuitem">
                    <Icon name="user" size="sm" />
                    {{ t('nav.profile') }}
                  </router-link>

                  <router-link :to="authRouteDefaults.apiKeysPath" @click="closeDropdown" class="app-header-user-dropdown-item" role="menuitem">
                    <Icon name="key" size="sm" />
                    {{ t('nav.apiKeys') }}
                  </router-link>
                </template>

                <a
                  v-if="showGithubLink"
                  href="https://github.com/Wei-Shaw/cloudbase"
                  target="_blank"
                  rel="noopener noreferrer"
                  @click="closeDropdown"
                  class="app-header-user-dropdown-item"
                  role="menuitem"
                >
                  <svg class="h-4 w-4" fill="currentColor" viewBox="0 0 24 24">
                    <path
                      fill-rule="evenodd"
                      clip-rule="evenodd"
                      d="M12 2C6.477 2 2 6.477 2 12c0 4.42 2.865 8.17 6.839 9.49.5.092.682-.217.682-.482 0-.237-.008-.866-.013-1.7-2.782.604-3.369-1.34-3.369-1.34-.454-1.156-1.11-1.464-1.11-1.464-.908-.62.069-.608.069-.608 1.003.07 1.531 1.03 1.531 1.03.892 1.529 2.341 1.087 2.91.831.092-.646.35-1.086.636-1.336-2.22-.253-4.555-1.11-4.555-4.943 0-1.091.39-1.984 1.029-2.683-.103-.253-.446-1.27.098-2.647 0 0 .84-.269 2.75 1.025A9.578 9.578 0 0112 6.836c.85.004 1.705.114 2.504.336 1.909-1.294 2.747-1.025 2.747-1.025.546 1.377.203 2.394.1 2.647.64.699 1.028 1.592 1.028 2.683 0 3.842-2.339 4.687-4.566 4.935.359.309.678.919.678 1.852 0 1.336-.012 2.415-.012 2.743 0 .267.18.578.688.48C19.138 20.167 22 16.418 22 12c0-5.523-4.477-10-10-10z"
                    />
                  </svg>
                  {{ t('nav.github') }}
                </a>

              </div>

              <div class="app-header-user-dropdown-group app-header-user-dropdown-group-danger">
                <button
                  @click="handleLogout"
                  class="app-header-user-dropdown-item app-header-user-dropdown-item-danger w-full"
                  role="menuitem"
                >
                  <svg
                    class="h-4 w-4"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    stroke-width="1.5"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      d="M15.75 9V5.25A2.25 2.25 0 0013.5 3h-6a2.25 2.25 0 00-2.25 2.25v13.5A2.25 2.25 0 007.5 21h6a2.25 2.25 0 002.25-2.25V15M12 9l-3 3m0 0l3 3m-3-3h12.75"
                    />
                  </svg>
                  {{ t('nav.logout') }}
                </button>
              </div>
            </div>
          </transition>
        </div>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAppStore, useAuthStore } from '@/stores'
import { useAdminSettingsStore } from '@/stores/adminSettings'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import SubscriptionProgressMini from '@/components/common/SubscriptionProgressMini.vue'
import AnnouncementBell from '@/components/common/AnnouncementBell.vue'
import DocsLink from '@/components/common/DocsLink.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAuthRouteDefaults } from '@/composables/useAuthRouteDefaults'
import {
  resolveHeaderDisplayName,
  resolveHeaderPageTitle,
  resolveHeaderUserInitials,
} from './appHeaderRuntime'

withDefaults(defineProps<{
  containerClass?: string
}>(), {
  containerClass: '',
})

const router = useRouter()
const route = useRoute()
const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const adminSettingsStore = useAdminSettingsStore()
const { authRouteDefaults } = useAuthRouteDefaults()

const user = computed(() => authStore.user)
const dropdownOpen = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)
const docUrl = computed(() => appStore.docUrl)
const avatarUrl = computed(() => user.value?.avatar_url?.trim() || '')
const showDropdownTaskLink = computed(() => true)
const showDropdownAccountLinks = computed(() => false)
const showGithubLink = computed(() => false)
const showDropdownPrimaryActions = computed(
  () => showDropdownTaskLink.value || showDropdownAccountLinks.value || showGithubLink.value
)

const userInitials = computed(() => resolveHeaderUserInitials(user.value))

const displayName = computed(() => resolveHeaderDisplayName(user.value))

const pageTitle = computed(() => {
  return resolveHeaderPageTitle({
    routeName: route.name,
    routeCustomId: route.params.id as string | undefined,
    routeMetaTitleKey: route.meta.titleKey as string | undefined,
    routeMetaTitle: route.meta.title as string | undefined,
    publicMenuItems: appStore.cachedPublicSettings?.custom_menu_items ?? [],
    adminMenuItems: adminSettingsStore.customMenuItems,
    isAdmin: authStore.isAdmin,
    translate: t,
  })
})

function toggleMobileSidebar() {
  appStore.toggleMobileSidebar()
}

function toggleDropdown() {
  dropdownOpen.value = !dropdownOpen.value
}

function closeDropdown() {
  dropdownOpen.value = false
}

async function handleLogout() {
  closeDropdown()
  try {
    await authStore.logout()
  } catch (error) {
    // Ignore logout errors - still redirect to login
    console.error('Logout error:', error)
  }
  await router.push(authRouteDefaults.value.loginPath)
}

function handleClickOutside(event: MouseEvent) {
  if (dropdownRef.value && !dropdownRef.value.contains(event.target as Node)) {
    closeDropdown()
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.app-header-user-dropdown {
  position: absolute;
  right: 0;
  z-index: 50;
  margin-top: 0.75rem;
  width: 13rem;
  overflow: hidden;
  border: 1px solid rgba(229, 231, 235, 0.96);
  border-radius: 1.25rem;
  background: rgba(255, 255, 255, 0.96);
  padding: 0.375rem;
  color: #374151;
  box-shadow:
    0 1px 2px rgba(15, 23, 42, 0.06),
    0 24px 60px rgba(15, 23, 42, 0.14);
  backdrop-filter: blur(18px);
  transform-origin: top right;
}

.app-header-user-dropdown-group {
  display: grid;
  gap: 0.125rem;
}

.app-header-user-dropdown-group + .app-header-user-dropdown-group {
  margin-top: 0.375rem;
  border-top: 1px solid rgba(229, 231, 235, 0.72);
  padding-top: 0.375rem;
}

.app-header-user-dropdown-item {
  display: flex;
  min-height: 2.5rem;
  align-items: center;
  gap: 0.625rem;
  border-radius: 0.875rem;
  padding: 0 0.75rem;
  color: inherit;
  font-size: 0.875rem;
  font-weight: 650;
  line-height: 1;
  text-align: left;
  transition:
    background-color 160ms ease,
    color 160ms ease;
}

.app-header-user-dropdown-item:hover {
  background: rgba(17, 24, 39, 0.045);
  color: #111827;
}

.app-header-user-dropdown-item-danger {
  color: #dc2626;
}

.app-header-user-dropdown-item-danger:hover {
  background: rgba(220, 38, 38, 0.07);
  color: #dc2626;
}

.dark .app-header-user-dropdown {
  border-color: rgba(255, 255, 255, 0.12);
  background: rgba(24, 24, 27, 0.96);
  color: #e5e7eb;
  box-shadow:
    0 1px 0 rgba(255, 255, 255, 0.03) inset,
    0 24px 60px rgba(0, 0, 0, 0.32);
}

.dark .app-header-user-dropdown-group + .app-header-user-dropdown-group {
  border-top-color: rgba(255, 255, 255, 0.1);
}

.dark .app-header-user-dropdown-item:hover {
  background: rgba(255, 255, 255, 0.08);
  color: #fff;
}

.dark .app-header-user-dropdown-item-danger {
  color: #f87171;
}

.dark .app-header-user-dropdown-item-danger:hover {
  background: rgba(248, 113, 113, 0.12);
  color: #fca5a5;
}

.dropdown-enter-active,
.dropdown-leave-active {
  transition: all 0.2s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: scale(0.95) translateY(-4px);
}
</style>
