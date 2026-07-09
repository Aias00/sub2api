<template>
  <aside
    class="sidebar"
    :class="[
      sidebarCollapsed ? 'w-[72px]' : 'w-64',
      { '-translate-x-full lg:translate-x-0': !mobileOpen }
    ]"
  >
    <!-- Logo/Brand -->
    <div class="sidebar-header" :class="{ 'sidebar-header-collapsed': sidebarCollapsed }">
      <router-link
        :to="authRouteDefaults.homePath"
        class="sidebar-home-link"
        :class="{ 'sidebar-home-link-collapsed': sidebarCollapsed }"
        :aria-label="siteName"
        @click="closeMobile"
      >
        <!-- Custom Logo or Default Logo -->
        <div v-if="settingsLoaded && siteLogo" class="sidebar-logo flex h-9 w-9 items-center justify-center overflow-hidden rounded-xl shadow-glow">
          <img :src="siteLogo" alt="Logo" class="h-full w-full object-contain" />
        </div>
        <div class="sidebar-brand" :class="{ 'sidebar-brand-collapsed': sidebarCollapsed }" :aria-hidden="sidebarCollapsed ? 'true' : 'false'">
          <span class="sidebar-brand-title text-lg font-bold text-gray-900 dark:text-white">
            {{ siteName }}
          </span>
        </div>
      </router-link>
    </div>

    <!-- Navigation -->
    <nav class="sidebar-nav scrollbar-hide">
      <!-- Admin View: Admin menu first, then personal menu -->
      <template v-if="isAdmin">
        <!-- Admin Section -->
        <div
          v-for="section in adminNavSections"
          :key="section.id"
          class="sidebar-section"
        >
          <div
            v-if="section.showTitle"
            class="sidebar-section-title"
            :class="{ 'sidebar-section-title-collapsed': sidebarCollapsed }"
            :aria-hidden="sidebarCollapsed ? 'true' : 'false'"
          >
            <span class="sidebar-section-title-text" :class="{ 'sidebar-section-title-text-collapsed': sidebarCollapsed }">
              {{ section.title }}
            </span>
          </div>
          <template v-for="item in section.items" :key="item.path">
            <!-- Collapsible group (has children) -->
            <template v-if="item.children?.length">
              <button
                type="button"
                class="sidebar-link mb-1 w-full"
                :class="{
                  'sidebar-link-active': isGroupActive(item) && !isGroupExpanded(item),
                  'sidebar-link-collapsed': sidebarCollapsed
                }"
                :title="sidebarCollapsed ? item.label : undefined"
                @click="handleGroupClick(item)"
              >
                <component :is="item.icon" class="h-5 w-5 flex-shrink-0" />
                <span
                  class="sidebar-label sidebar-label-flex"
                  :class="{ 'sidebar-label-collapsed': sidebarCollapsed }"
                  :aria-hidden="sidebarCollapsed ? 'true' : 'false'"
                >
                  <span class="min-w-0 truncate">{{ item.label }}</span>
                  <ChevronDownIcon
                    class="h-4 w-4 flex-shrink-0 transition-transform duration-200"
                    :class="isGroupExpanded(item) ? 'rotate-180' : ''"
                  />
                </span>
              </button>
              <!-- Children -->
              <div v-if="!sidebarCollapsed && isGroupExpanded(item)" class="mb-1 ml-4 border-l border-gray-200 pl-2 dark:border-dark-600">
                <router-link
                  v-for="child in item.children"
                  :key="child.path"
                  :to="child.path"
                  class="sidebar-link mb-0.5 py-1.5 text-sm"
                  :class="{ 'sidebar-link-active': route.path === child.path }"
                  @click="handleMenuItemClick(child.path)"
                >
                  <component :is="child.icon" class="h-4 w-4 flex-shrink-0" />
                  <span>{{ child.label }}</span>
                </router-link>
              </div>
            </template>
            <!-- Normal item (no children) -->
            <router-link
              v-else
              :to="item.path"
              class="sidebar-link mb-1"
              :class="{ 'sidebar-link-active': isActive(item.path), 'sidebar-link-collapsed': sidebarCollapsed }"
              :title="sidebarCollapsed ? item.label : undefined"
              :id="
                item.path === adminSidebarPaths.adminAccountsPath
                  ? 'sidebar-channel-manage'
                  : item.path === adminSidebarPaths.adminGroupsPath
                    ? 'sidebar-group-manage'
                    : item.path === adminSidebarPaths.adminRedeemPath
                      ? 'sidebar-wallet'
                      : undefined
              "
              @click="handleMenuItemClick(item.path)"
            >
              <span v-if="item.iconSvg" class="h-5 w-5 flex-shrink-0 sidebar-svg-icon" v-html="sanitizeSvg(item.iconSvg)"></span>
              <component v-else :is="item.icon" class="h-5 w-5 flex-shrink-0" />
              <span class="sidebar-label" :class="{ 'sidebar-label-collapsed': sidebarCollapsed }" :aria-hidden="sidebarCollapsed ? 'true' : 'false'">{{ item.label }}</span>
            </router-link>
          </template>
        </div>

        <!-- Personal Section for Admin (hidden in simple mode) -->
        <template v-if="!authStore.isSimpleMode">
          <div
            v-for="section in personalNavSections"
            :key="section.id"
            class="sidebar-section"
          >
            <div
              v-if="section.showTitle"
              class="sidebar-section-title"
              :class="{ 'sidebar-section-title-collapsed': sidebarCollapsed }"
              :aria-hidden="sidebarCollapsed ? 'true' : 'false'"
            >
              <span class="sidebar-section-title-text" :class="{ 'sidebar-section-title-text-collapsed': sidebarCollapsed }">
                {{ section.title }}
              </span>
            </div>

            <router-link
              v-for="item in section.items"
              :key="item.path"
              :to="item.path"
              class="sidebar-link mb-1"
              :class="{ 'sidebar-link-active': isActive(item.path), 'sidebar-link-collapsed': sidebarCollapsed }"
              :title="sidebarCollapsed ? item.label : undefined"
              :data-tour="item.path === authRouteDefaults.apiKeysPath ? 'sidebar-my-keys' : undefined"
              @click="handleMenuItemClick(item.path)"
            >
              <span v-if="item.iconSvg" class="h-5 w-5 flex-shrink-0 sidebar-svg-icon" v-html="sanitizeSvg(item.iconSvg)"></span>
              <component v-else :is="item.icon" class="h-5 w-5 flex-shrink-0" />
              <span class="sidebar-label" :class="{ 'sidebar-label-collapsed': sidebarCollapsed }" :aria-hidden="sidebarCollapsed ? 'true' : 'false'">{{ item.label }}</span>
            </router-link>
          </div>
        </template>
      </template>

      <!-- Regular User View -->
      <template v-else-if="!appStore.backendModeEnabled">
        <div
          v-for="section in userNavSections"
          :key="section.id"
          class="sidebar-section"
        >
          <router-link
            v-for="item in section.items"
            :key="item.path"
            :to="item.path"
            class="sidebar-link mb-1"
            :class="{ 'sidebar-link-active': isActive(item.path), 'sidebar-link-collapsed': sidebarCollapsed }"
            :title="sidebarCollapsed ? item.label : undefined"
            :data-tour="item.path === authRouteDefaults.apiKeysPath ? 'sidebar-my-keys' : undefined"
            @click="handleMenuItemClick(item.path)"
          >
            <span v-if="item.iconSvg" class="h-5 w-5 flex-shrink-0 sidebar-svg-icon" v-html="sanitizeSvg(item.iconSvg)"></span>
            <component v-else :is="item.icon" class="h-5 w-5 flex-shrink-0" />
            <span class="sidebar-label" :class="{ 'sidebar-label-collapsed': sidebarCollapsed }" :aria-hidden="sidebarCollapsed ? 'true' : 'false'">{{ item.label }}</span>
          </router-link>
        </div>
      </template>
    </nav>

    <!-- Bottom Section -->
    <div class="mt-auto border-t border-gray-100 p-3 dark:border-dark-800">
      <!-- Collapse Button -->
      <button
        @click="toggleSidebar"
        class="sidebar-link w-full"
        :class="{ 'sidebar-link-collapsed': sidebarCollapsed }"
        :title="sidebarCollapsed ? t('nav.expand') : t('nav.collapse')"
      >
        <ChevronDoubleLeftIcon v-if="!sidebarCollapsed" class="h-5 w-5 flex-shrink-0" />
        <ChevronDoubleRightIcon v-else class="h-5 w-5 flex-shrink-0" />
        <span class="sidebar-label" :class="{ 'sidebar-label-collapsed': sidebarCollapsed }" :aria-hidden="sidebarCollapsed ? 'true' : 'false'">{{ t('nav.collapse') }}</span>
      </button>
    </div>
  </aside>

  <!-- Mobile Overlay -->
  <transition name="fade">
    <div
      v-if="mobileOpen"
      class="fixed inset-0 z-30 bg-black/50 lg:hidden"
      @click="closeMobile"
    ></div>
  </transition>
</template>

<script setup lang="ts">
import { computed, h, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthRouteDefaults } from '@/composables/useAuthRouteDefaults'
import { useAdminSettingsStore, useAppStore, useAuthStore, useOnboardingStore } from '@/stores'
import { sanitizeSvg } from '@/utils/sanitize'
import { resolveAdminSidebarRouteDefaults } from '@/utils/adminSidebarShell'
import {
  resolveAdminSidebarSections,
  type AdminSidebarItemKey,
  type AdminSidebarSection,
  resolveSelfSidebarSections,
  type SelfSidebarItemKey,
  type SelfSidebarSection,
} from '@/utils/adminSidebarSchema'
import { FeatureFlags, makeSidebarFlag } from '@/utils/featureFlags'
import {
  buildSidebarSections,
  buildSidebarVisibleItemMap,
  type SidebarNavItem,
  type SidebarNavSection,
} from './sidebarRuntime'

interface NavItem extends SidebarNavItem {
  path: string
  label: string
  icon: unknown
  iconSvg?: string
  hideInSimpleMode?: boolean
  children?: NavItem[]
  /**
   * When true, the parent item only toggles the expand/collapse state and
   * does NOT navigate to its `path`. The `path` is purely a stable key.
   */
  expandOnly?: boolean
  /**
   * 可选的功能开关 getter。返回 false 时菜单项被隐藏；返回 undefined/true 时显示。
   * 宽容策略（undefined → 显示）避免 public settings 未加载完成时菜单闪烁消失。
   * Getter 里访问的 reactive 来源（store / composable）会被 computed 自动追踪，
   * 开关切换时菜单自动更新。
   */
  featureFlag?: () => boolean | undefined
}

type NavSection = SidebarNavSection & {
  id: string
  items: NavItem[]
}

type TitledNavSection = NavSection & {
  title?: string
  showTitle?: boolean
}

const i18n = useI18n()
const { t, locale } = i18n

const route = useRoute()
const router = useRouter()
const appStore = useAppStore()
const authStore = useAuthStore()
const onboardingStore = useOnboardingStore()
const adminSettingsStore = useAdminSettingsStore()
const { authRouteDefaults } = useAuthRouteDefaults()

const sidebarCollapsed = computed(() => appStore.sidebarCollapsed)
const mobileOpen = computed(() => appStore.mobileOpen)
const isAdmin = computed(() => authStore.isAdmin)

// Track which parent nav groups are expanded
const expandedGroups = ref<Set<string>>(new Set())

// Site settings from appStore (cached, no flicker)
const siteName = computed(() => appStore.siteName)
const siteLogo = computed(() => appStore.siteLogo)
const settingsLoaded = computed(() => appStore.publicSettingsLoaded)
const adminSidebarPaths = computed(() =>
  resolveAdminSidebarRouteDefaults(appStore.cachedPublicSettings?.auth_shell_config, locale.value),
)
const configuredAdminSidebarSections = computed(() =>
  resolveAdminSidebarSections(appStore.cachedPublicSettings?.auth_shell_config, locale.value),
)
const configuredUserSidebarSections = computed(() =>
  resolveSelfSidebarSections(appStore.cachedPublicSettings?.auth_shell_config, locale.value, 'userSidebarSections'),
)
const configuredAdminPersonalSidebarSections = computed(() =>
  resolveSelfSidebarSections(appStore.cachedPublicSettings?.auth_shell_config, locale.value, 'adminPersonalSidebarSections'),
)

// SVG Icon Components
const DashboardIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M3.75 6A2.25 2.25 0 016 3.75h2.25A2.25 2.25 0 0110.5 6v2.25a2.25 2.25 0 01-2.25 2.25H6a2.25 2.25 0 01-2.25-2.25V6zM3.75 15.75A2.25 2.25 0 016 13.5h2.25a2.25 2.25 0 012.25 2.25V18a2.25 2.25 0 01-2.25 2.25H6A2.25 2.25 0 013.75 18v-2.25zM13.5 6a2.25 2.25 0 012.25-2.25H18A2.25 2.25 0 0120.25 6v2.25A2.25 2.25 0 0118 10.5h-2.25a2.25 2.25 0 01-2.25-2.25V6zM13.5 15.75a2.25 2.25 0 012.25-2.25H18a2.25 2.25 0 012.25 2.25V18A2.25 2.25 0 0118 20.25h-2.25A2.25 2.25 0 0113.5 18v-2.25z'
        })
      ]
    )
}

const KeyIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M15.75 5.25a3 3 0 013 3m3 0a6 6 0 01-7.029 5.912c-.563-.097-1.159.026-1.563.43L10.5 17.25H8.25v2.25H6v2.25H2.25v-2.818c0-.597.237-1.17.659-1.591l6.499-6.499c.404-.404.527-1 .43-1.563A6 6 0 1121.75 8.25z'
        })
      ]
    )
}

const ChartIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z'
        })
      ]
    )
}

const GiftIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M21 11.25v8.25a1.5 1.5 0 01-1.5 1.5H5.25a1.5 1.5 0 01-1.5-1.5v-8.25M12 4.875A2.625 2.625 0 109.375 7.5H12m0-2.625V7.5m0-2.625A2.625 2.625 0 1114.625 7.5H12m0 0V21m-8.625-9.75h18c.621 0 1.125-.504 1.125-1.125v-1.5c0-.621-.504-1.125-1.125-1.125h-18c-.621 0-1.125.504-1.125 1.125v1.5c0 .621.504 1.125 1.125 1.125z'
        })
      ]
    )
}

const UserIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M15.75 6a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0zM4.501 20.118a7.5 7.5 0 0114.998 0A17.933 17.933 0 0112 21.75c-2.676 0-5.216-.584-7.499-1.632z'
        })
      ]
    )
}

const UsersIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zm8.25 2.25a2.625 2.625 0 11-5.25 0 2.625 2.625 0 015.25 0z'
        })
      ]
    )
}

const FolderIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M2.25 12.75V12A2.25 2.25 0 014.5 9.75h15A2.25 2.25 0 0121.75 12v.75m-8.69-6.44l-2.12-2.12a1.5 1.5 0 00-1.061-.44H4.5A2.25 2.25 0 002.25 6v12a2.25 2.25 0 002.25 2.25h15A2.25 2.25 0 0021.75 18V9a2.25 2.25 0 00-2.25-2.25h-5.379a1.5 1.5 0 01-1.06-.44z'
        })
      ]
    )
}

const ChannelIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M6.429 9.75L2.25 12l4.179 2.25m0-4.5l5.571 3 5.571-3m-11.142 0L2.25 7.5 12 2.25l9.75 5.25-4.179 2.25m0 0l4.179 2.25L12 17.25 2.25 12m15.321-2.25l4.179 2.25L12 17.25l-9.75-5.25'
        })
      ]
    )
}

const CreditCardIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M2.25 8.25h19.5M2.25 9h19.5m-16.5 5.25h6m-6 2.25h3m-3.75 3h15a2.25 2.25 0 002.25-2.25V6.75A2.25 2.25 0 0019.5 4.5h-15a2.25 2.25 0 00-2.25 2.25v10.5A2.25 2.25 0 004.5 19.5z'
        })
      ]
    )
}

const RechargeSubscriptionIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'currentColor', viewBox: '0 0 1024 1024' },
      [
        h('path', {
          d: 'M512 992C247.3 992 32 776.7 32 512S247.3 32 512 32s480 215.3 480 480c0 84.4-22.2 167.4-64.2 240-8.9 15.3-28.4 20.6-43.7 11.7-15.3-8.8-20.5-28.4-11.7-43.7 36.4-62.9 55.6-134.8 55.6-208 0-229.4-186.6-416-416-416S96 282.6 96 512s186.6 416 416 416c17.7 0 32 14.3 32 32s-14.3 32-32 32z'
        }),
        h('path', {
          d: 'M640 512H384c-17.7 0-32-14.3-32-32s14.3-32 32-32h256c17.7 0 32 14.3 32 32s-14.3 32-32 32zM640 640H384c-17.7 0-32-14.3-32-32s14.3-32 32-32h256c17.7 0 32 14.3 32 32s-14.3 32-32 32z'
        }),
        h('path', {
          d: 'M512 480c-8.2 0-16.4-3.1-22.6-9.4l-128-128c-12.5-12.5-12.5-32.8 0-45.3s32.8-12.5 45.3 0l128 128c12.5 12.5 12.5 32.8 0 45.3-6.3 6.3-14.5 9.4-22.7 9.4z'
        }),
        h('path', {
          d: 'M512 480c-8.2 0-16.4-3.1-22.6-9.4-12.5-12.5-12.5-32.8 0-45.3l128-128c12.5-12.5 32.8-12.5 45.3 0s12.5 32.8 0 45.3l-128 128c-6.3 6.3-14.5 9.4-22.7 9.4z'
        }),
        h('path', {
          d: 'M512 736c-17.7 0-32-14.3-32-32V448c0-17.7 14.3-32 32-32s32 14.3 32 32v256c0 17.7-14.3 32-32 32zM896 992H512c-17.7 0-32-14.3-32-32s14.3-32 32-32h306.8l-73.4-73.4c-12.5-12.5-12.5-32.8 0-45.3s32.8-12.5 45.3 0l128 128c9.2 9.2 11.9 22.9 6.9 34.9S908.9 992 896 992z'
        })
      ]
    )
}

const GlobeIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M12 21a9.004 9.004 0 008.716-6.747M12 21a9.004 9.004 0 01-8.716-6.747M12 21c2.485 0 4.5-4.03 4.5-9S14.485 3 12 3m0 18c-2.485 0-4.5-4.03-4.5-9S9.515 3 12 3m0 0a8.997 8.997 0 017.843 4.582M12 3a8.997 8.997 0 00-7.843 4.582m15.686 0A11.953 11.953 0 0112 10.5c-2.998 0-5.74-1.1-7.843-2.918m15.686 0A8.959 8.959 0 0121 12c0 .778-.099 1.533-.284 2.253m0 0A17.919 17.919 0 0112 16.5c-3.162 0-6.133-.815-8.716-2.247m0 0A9.015 9.015 0 013 12c0-1.605.42-3.113 1.157-4.418'
        })
      ]
    )
}

const ServerIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M5.25 14.25h13.5m-13.5 0a3 3 0 01-3-3m3 3a3 3 0 100 6h13.5a3 3 0 100-6m-16.5-3a3 3 0 013-3h13.5a3 3 0 013 3m-19.5 0a4.5 4.5 0 01.9-2.7L5.737 5.1a3.375 3.375 0 012.7-1.35h7.126c1.062 0 2.062.5 2.7 1.35l2.587 3.45a4.5 4.5 0 01.9 2.7m0 0a3 3 0 01-3 3m0 3h.008v.008h-.008v-.008zm0-6h.008v.008h-.008v-.008zm-3 6h.008v.008h-.008v-.008zm0-6h.008v.008h-.008v-.008z'
        })
      ]
    )
}

const BellIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M14.857 17.082a23.848 23.848 0 005.454-1.31A8.967 8.967 0 0118 9.75V9a6 6 0 10-12 0v.75a8.967 8.967 0 01-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 01-5.714 0m5.714 0a3 3 0 11-5.714 0'
        })
      ]
    )
}

const TicketIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M16.5 6v.75m0 3v.75m0 3v.75m0 3V18m-9-5.25h5.25M7.5 15h3M3.375 5.25c-.621 0-1.125.504-1.125 1.125v3.026a2.999 2.999 0 010 5.198v3.026c0 .621.504 1.125 1.125 1.125h17.25c.621 0 1.125-.504 1.125-1.125v-3.026a2.999 2.999 0 010-5.198V6.375c0-.621-.504-1.125-1.125-1.125H3.375z'
        })
      ]
    )
}

const CogIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.324.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 011.37.49l1.296 2.247a1.125 1.125 0 01-.26 1.431l-1.003.827c-.293.24-.438.613-.431.992a6.759 6.759 0 010 .255c-.007.378.138.75.43.99l1.005.828c.424.35.534.954.26 1.43l-1.298 2.247a1.125 1.125 0 01-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.57 6.57 0 01-.22.128c-.331.183-.581.495-.644.869l-.213 1.28c-.09.543-.56.941-1.11.941h-2.594c-.55 0-1.02-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 01-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 01-1.369-.49l-1.297-2.247a1.125 1.125 0 01.26-1.431l1.004-.827c.292-.24.437-.613.43-.992a6.932 6.932 0 010-.255c.007-.378-.138-.75-.43-.99l-1.004-.828a1.125 1.125 0 01-.26-1.43l1.297-2.247a1.125 1.125 0 011.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.087.22-.128.332-.183.582-.495.644-.869l.214-1.281z'
        }),
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M15 12a3 3 0 11-6 0 3 3 0 016 0z'
        })
      ]
    )
}

const ChevronDoubleLeftIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'm18.75 4.5-7.5 7.5 7.5 7.5m-6-15L5.25 12l7.5 7.5'
        })
      ]
    )
}

const OrderIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M9 12h3.75M9 15h3.75M9 18h3.75m3 .75H18a2.25 2.25 0 002.25-2.25V6.108c0-1.135-.845-2.098-1.976-2.192a48.424 48.424 0 00-1.123-.08m-5.801 0c-.065.21-.1.433-.1.664 0 .414.336.75.75.75h4.5a.75.75 0 00.75-.75 2.25 2.25 0 00-.1-.664m-5.8 0A2.251 2.251 0 0113.5 2.25H15a2.25 2.25 0 012.15 1.586m-5.8 0c-.376.023-.75.05-1.124.08C9.095 4.01 8.25 4.973 8.25 6.108V8.25m0 0H4.875c-.621 0-1.125.504-1.125 1.125v11.25c0 .621.504 1.125 1.125 1.125h9.75c.621 0 1.125-.504 1.125-1.125V9.375c0-.621-.504-1.125-1.125-1.125H8.25zM6.75 12h.008v.008H6.75V12zm0 3h.008v.008H6.75V15zm0 3h.008v.008H6.75V18z'
        })
      ]
    )
}

const OrderListIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z'
        })
      ]
    )
}

const ChevronDoubleRightIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'm5.25 4.5 7.5 7.5-7.5 7.5m6-15 7.5 7.5-7.5 7.5'
        })
      ]
    )
}

const SignalIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M9.348 14.651a3.75 3.75 0 010-5.303m5.304 0a3.75 3.75 0 010 5.303m-7.425 2.122a6.75 6.75 0 010-9.546m9.546 0a6.75 6.75 0 010 9.546M5.106 18.894c-3.808-3.807-3.808-9.98 0-13.788m13.788 0c3.808 3.807 3.808 9.98 0 13.788M12 12h.008v.008H12V12zm.375 0a.375.375 0 11-.75 0 .375.375 0 01.75 0z'
        })
      ]
    )
}

const ShieldIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z'
        })
      ]
    )
}

const PriceTagIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M9.568 3H5.25A2.25 2.25 0 003 5.25v4.318c0 .597.237 1.17.659 1.591l9.581 9.581c.699.699 1.78.872 2.607.33a18.095 18.095 0 005.223-5.223c.542-.827.369-1.908-.33-2.607L11.16 3.66A2.25 2.25 0 009.568 3z'
        }),
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M6 6h.008v.008H6V6z'
        })
      ]
    )
}

const PromptIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M8.25 6.75h7.5M8.25 10.5h7.5M8.25 14.25h4.5M5.25 3.75h13.5A2.25 2.25 0 0121 6v10.5a2.25 2.25 0 01-2.25 2.25H9l-4.125 2.475A.75.75 0 013.75 20.58V6A2.25 2.25 0 016 3.75z'
        })
      ]
    )
}

const ImageIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M3.75 6A2.25 2.25 0 016 3.75h12A2.25 2.25 0 0120.25 6v12A2.25 2.25 0 0118 20.25H6A2.25 2.25 0 013.75 18V6z'
        }),
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M3.75 16.5l4.72-4.72a1.5 1.5 0 012.12 0l2.66 2.66 1.22-1.22a1.5 1.5 0 012.12 0l3.66 3.66M14.25 8.25h.008v.008h-.008V8.25z'
        })
      ]
    )
}

const FlameIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M15.36 5.64c.58 1.96.26 3.64-.97 5.04-.6.69-1.05 1.36-1.34 2.02-.36-1.85-1.27-3.29-2.72-4.34-2.56 2.23-3.84 4.66-3.84 7.28 0 3.18 2.38 5.61 5.51 5.61s5.51-2.43 5.51-5.61c0-3.04-1.5-6.37-2.15-10z'
        })
      ]
    )
}

const ChevronDownIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'm19.5 8.25-7.5 7.5-7.5-7.5'
        })
      ]
    )
}

// Public-settings flags go through the registry in utils/featureFlags.ts,
// which handles the opt-in vs opt-out fallback when settings haven't loaded
// yet. Admin-only flags (not in public settings) stay inline below.
const flagChannelMonitor = makeSidebarFlag(FeatureFlags.channelMonitor)
const flagPayment = makeSidebarFlag(FeatureFlags.payment)
const flagAvailableChannels = makeSidebarFlag(FeatureFlags.availableChannels)
const flagAffiliate = makeSidebarFlag(FeatureFlags.affiliate)
const flagRiskControl = makeSidebarFlag(FeatureFlags.riskControl)
const flagOpsMonitoring = () => adminSettingsStore.opsMonitoringEnabled
const flagAdminPayment = () => adminSettingsStore.paymentEnabled

// Custom menu items filtered by visibility
const customMenuItemsForUser = computed(() => {
  const items = appStore.cachedPublicSettings?.custom_menu_items ?? []
  return items
    .filter((item) => item.visibility === 'user')
    .sort((a, b) => a.sort_order - b.sort_order)
})

const customMenuItemsForAdmin = computed(() => {
  return adminSettingsStore.customMenuItems
    .filter((item) => item.visibility === 'admin')
    .sort((a, b) => a.sort_order - b.sort_order)
})

function buildSelfNavItemMap(withDashboard: boolean): Partial<Record<SelfSidebarItemKey, NavItem>> {
  const navPaths = authRouteDefaults.value
  const itemMap: Partial<Record<SelfSidebarItemKey, NavItem>> = {
    tasks: { path: '/app/tasks', label: t('nav.myTasks'), icon: OrderIcon },
    promptCatalog: { path: '/app/prompts', label: t('nav.promptCatalog'), icon: PromptIcon },
    imageGenerator: { path: '/app/image-generator', label: t('nav.imageGenerator'), icon: ImageIcon },
    wechatExport: { path: '/app/wechat', label: t('nav.wechatExport'), icon: OrderListIcon },
    hotTopics: { path: '/app/hot', label: t('nav.hotTopics'), icon: FlameIcon },
    apiKeys: { path: navPaths.apiKeysPath, label: t('nav.apiKeys'), icon: KeyIcon },
    usage: { path: navPaths.usagePath, label: t('nav.usage'), icon: ChartIcon, hideInSimpleMode: true },
    availableChannels: { path: navPaths.availableChannelsPath, label: t('nav.availableChannels'), icon: ChannelIcon, hideInSimpleMode: true, featureFlag: flagAvailableChannels },
    availableGroups: { path: navPaths.availableGroupsPath, label: t('nav.availableGroups'), icon: FolderIcon },
    subscriptions: { path: navPaths.subscriptionsPath, label: t('nav.mySubscriptions'), icon: CreditCardIcon, hideInSimpleMode: true },
    purchase: { path: navPaths.purchasePath, label: t('nav.buySubscription'), icon: RechargeSubscriptionIcon, hideInSimpleMode: true, featureFlag: flagPayment },
    orders: { path: navPaths.ordersPath, label: t('nav.myOrders'), icon: OrderListIcon, hideInSimpleMode: true, featureFlag: flagPayment },
    redeem: { path: navPaths.redeemPath, label: t('nav.redeem'), icon: GiftIcon, hideInSimpleMode: true },
    affiliate: { path: navPaths.affiliatePath, label: t('nav.affiliate'), icon: UsersIcon, hideInSimpleMode: true, featureFlag: flagAffiliate },
    profile: { path: navPaths.profilePath, label: t('nav.profile'), icon: UserIcon },
  }
  if (withDashboard) {
    itemMap.dashboard = { path: navPaths.userRedirectPath, label: t('nav.dashboard'), icon: DashboardIcon }
  }

  return buildSidebarVisibleItemMap(
    itemMap,
    authStore.isSimpleMode,
  ) as Partial<Record<SelfSidebarItemKey, NavItem>>
}

function buildSelfNavSections(
  configuredSections: SelfSidebarSection[],
  defaultSections: SelfSidebarSection[],
  visibleMap: Partial<Record<SelfSidebarItemKey, NavItem>>,
  customItems: NavItem[],
): NavSection[] {
  return buildSidebarSections(
    configuredSections,
    defaultSections,
    visibleMap,
    customItems,
    'self-more',
    'self-custom',
  ) as NavSection[]
}

const userNavSections = computed((): NavSection[] => {
  const visibleMap = buildSelfNavItemMap(true)
  const defaultSections: SelfSidebarSection[] = [
    {
      id: 'user-main',
      items: [
        'dashboard',
        'tasks',
        'promptCatalog',
        'imageGenerator',
        'wechatExport',
        'hotTopics',
        'apiKeys',
        'usage',
        'availableChannels',
        'availableGroups',
        'subscriptions',
        'purchase',
        'orders',
        'redeem',
        'affiliate',
        'profile',
      ],
    },
  ]
  const customItems = customMenuItemsForUser.value.map((item): NavItem => ({
    path: `/custom/${item.id}`,
    label: item.label,
    icon: null,
    iconSvg: item.icon_svg,
  }))
  return buildSelfNavSections(configuredUserSidebarSections.value, defaultSections, visibleMap, customItems)
})

const personalNavSections = computed((): TitledNavSection[] => {
  const visibleMap = buildSelfNavItemMap(false)
  const defaultSections: SelfSidebarSection[] = [
    {
      id: 'admin-personal',
      items: [
        'tasks',
        'promptCatalog',
        'imageGenerator',
        'wechatExport',
        'hotTopics',
        'apiKeys',
        'usage',
        'availableChannels',
        'availableGroups',
        'subscriptions',
        'purchase',
        'orders',
        'redeem',
        'affiliate',
        'profile',
      ],
    },
  ]
  const sections = buildSelfNavSections(
    configuredAdminPersonalSidebarSections.value,
    defaultSections,
    visibleMap,
    [],
  )
  return sections.map((section, index) => ({
    ...section,
    title: t('nav.myAccount'),
    showTitle: index === 0,
  }))
})

// Admin navigation sections
const adminNavSections = computed((): TitledNavSection[] => {
  const navPaths = authRouteDefaults.value
  const adminPaths = adminSidebarPaths.value
  const builtInItemMap: Record<AdminSidebarItemKey, NavItem> = {
    dashboard: { path: navPaths.adminRedirectPath, label: t('nav.dashboard'), icon: DashboardIcon },
    ops: { path: adminPaths.adminOpsPath, label: t('nav.ops'), icon: ChartIcon, featureFlag: flagOpsMonitoring },
    users: { path: adminPaths.adminUsersPath, label: t('nav.users'), icon: UsersIcon, hideInSimpleMode: true },
    userInsights: { path: adminPaths.adminUserInsightsPath, label: t('nav.userInsights'), icon: ChartIcon, hideInSimpleMode: true },
    groups: { path: adminPaths.adminGroupsPath, label: t('nav.groups'), icon: FolderIcon, hideInSimpleMode: true },
    channels: {
      path: adminPaths.adminChannelsPath,
      label: t('nav.channelManagement'),
      icon: ChannelIcon,
      hideInSimpleMode: true,
      expandOnly: true,
      children: [
        { path: adminPaths.adminChannelPricingPath, label: t('nav.channelPricing'), icon: PriceTagIcon },
        { path: adminPaths.adminChannelMonitorPath, label: t('nav.channelMonitor'), icon: SignalIcon, featureFlag: flagChannelMonitor },
      ],
    },
    subscriptions: { path: adminPaths.adminSubscriptionsPath, label: t('nav.subscriptions'), icon: CreditCardIcon, hideInSimpleMode: true },
    accounts: { path: adminPaths.adminAccountsPath, label: t('nav.accounts'), icon: GlobeIcon },
    announcements: { path: adminPaths.adminAnnouncementsPath, label: t('nav.announcements'), icon: BellIcon },
    proxies: { path: adminPaths.adminProxiesPath, label: t('nav.proxies'), icon: ServerIcon },
    riskControl: { path: adminPaths.adminRiskControlPath, label: t('nav.riskControl'), icon: ShieldIcon, hideInSimpleMode: true, featureFlag: flagRiskControl },
    redeem: { path: adminPaths.adminRedeemPath, label: t('nav.redeemCodes'), icon: TicketIcon, hideInSimpleMode: true },
    promoCodes: { path: adminPaths.adminPromoCodesPath, label: t('nav.promoCodes'), icon: GiftIcon, hideInSimpleMode: true },
    affiliates: {
      path: adminPaths.adminAffiliatesPath,
      label: t('nav.affiliateManagement'),
      icon: UsersIcon,
      hideInSimpleMode: true,
      expandOnly: true,
      featureFlag: flagAffiliate,
      children: [
        { path: adminPaths.adminAffiliateOverviewPath, label: t('nav.affiliateOverview'), icon: DashboardIcon },
        { path: adminPaths.adminAffiliateRulesPath, label: t('nav.affiliateRules'), icon: CogIcon },
        { path: adminPaths.adminAffiliateCodesPath, label: t('nav.affiliateCodeManagement'), icon: TicketIcon },
        { path: adminPaths.adminAffiliateInvitesPath, label: t('nav.affiliateInviteRecords'), icon: UsersIcon },
        { path: adminPaths.adminAffiliateRebatesPath, label: t('nav.affiliateRebateRecords'), icon: OrderIcon },
        { path: adminPaths.adminAffiliateTransfersPath, label: t('nav.affiliateTransferRecords'), icon: CreditCardIcon },
      ],
    },
    orders: {
      path: adminPaths.adminOrdersRootPath,
      label: t('nav.orderManagement'),
      icon: OrderIcon,
      hideInSimpleMode: true,
      expandOnly: true,
      featureFlag: flagAdminPayment,
      children: [
        { path: adminPaths.adminOrdersDashboardPath, label: t('nav.paymentDashboard'), icon: ChartIcon },
        { path: adminPaths.adminOrdersRootPath, label: t('nav.orderManagement'), icon: OrderIcon },
        { path: adminPaths.adminPaymentPlansPath, label: t('nav.paymentPlans'), icon: CreditCardIcon },
      ],
    },
    usage: { path: adminPaths.adminUsagePath, label: t('nav.usage'), icon: ChartIcon },
    apiKeys: { path: authRouteDefaults.value.apiKeysPath, label: t('nav.apiKeys'), icon: KeyIcon },
    workers: { path: '/admin/workers', label: t('nav.workers'), icon: ServerIcon },
    runtimeSettings: { path: `${navPaths.adminSettingsPath}?tab=runtime`, label: t('nav.runtimeSettings'), icon: CogIcon },
    settings: { path: navPaths.adminSettingsPath, label: t('nav.settings'), icon: CogIcon },
  }

  const visibleMap = buildSidebarVisibleItemMap(
    builtInItemMap,
    authStore.isSimpleMode,
  ) as Partial<Record<AdminSidebarItemKey, NavItem>>

  const defaultSections: AdminSidebarSection[] = authStore.isSimpleMode
    ? [
        {
          id: 'admin-main',
          items: [
            'dashboard',
            'ops',
            'accounts',
            'announcements',
            'proxies',
            'usage',
            'apiKeys',
            'settings',
          ],
        },
      ]
    : [
        {
          id: 'admin-monitoring',
          items: ['dashboard', 'ops', 'usage'],
        },
        {
          id: 'admin-users',
          items: ['users', 'userInsights', 'groups'],
        },
        {
          id: 'admin-channels',
          items: ['accounts', 'channels'],
        },
        {
          id: 'admin-monetization',
          items: ['subscriptions', 'redeem', 'promoCodes', 'affiliates', 'orders'],
        },
        {
          id: 'admin-comms',
          items: ['announcements', 'riskControl', 'proxies'],
        },
        {
          id: 'admin-system',
          items: ['settings', 'runtimeSettings', 'workers'],
        },
      ]

  const configuredSections =
    configuredAdminSidebarSections.value.length > 0 && !authStore.isSimpleMode
      ? configuredAdminSidebarSections.value
      : defaultSections

  const customItems = customMenuItemsForAdmin.value.map((cm): NavItem => ({
    path: `/custom/${cm.id}`,
    label: cm.label,
    icon: null,
    iconSvg: cm.icon_svg,
  }))
  const sections = buildSidebarSections(
    configuredSections,
    defaultSections,
    visibleMap,
    customItems,
    'admin-more',
    'admin-custom',
  ) as NavSection[]

  // Default sections get visible group titles; configured/custom/fallback
  // sections render without a title (preserving prior behavior for admins who
  // supply their own auth_shell_config sidebar layout).
  const sectionTitles: Record<string, string> = {
    'admin-monitoring': t('nav.groupOpsMonitoring'),
    'admin-users': t('nav.groupUsers'),
    'admin-channels': t('nav.groupChannels'),
    'admin-monetization': t('nav.groupMonetization'),
    'admin-comms': t('nav.groupComms'),
    'admin-system': t('nav.groupSystem'),
  }
  return sections.map((section) => ({
    ...section,
    title: sectionTitles[section.id],
    showTitle: Boolean(sectionTitles[section.id]),
  }))
})

function toggleSidebar() {
  appStore.toggleSidebar()
}

function closeMobile() {
  appStore.setMobileOpen(false)
}

function handleMenuItemClick(itemPath: string) {
  if (mobileOpen.value) {
    setTimeout(() => {
      appStore.setMobileOpen(false)
    }, 150)
  }

  // Map paths to tour selectors
  const pathToSelector: Record<string, string> = {
    [adminSidebarPaths.value.adminGroupsPath]: '#sidebar-group-manage',
    [adminSidebarPaths.value.adminAccountsPath]: '#sidebar-channel-manage',
    [authRouteDefaults.value.apiKeysPath]: '[data-tour="sidebar-my-keys"]'
  }

  const selector = pathToSelector[itemPath]
  if (selector && onboardingStore.isCurrentStep(selector)) {
    onboardingStore.nextStep(500)
  }
}

function splitNavPath(path: string): { pathname: string; query: Record<string, string> } {
  const [pathname, search = ''] = path.split('?')
  return {
    pathname,
    query: Object.fromEntries(new URLSearchParams(search)),
  }
}

function isActive(path: string): boolean {
  const target = splitNavPath(path)
  const pathMatches = route.path === target.pathname || route.path.startsWith(target.pathname + '/')
  if (!pathMatches) return false

  const queryMatches = Object.entries(target.query).every(([key, value]) => {
    const current = route.query[key]
    return Array.isArray(current) ? current[0] === value : current === value
  })
  if (!queryMatches) return false

  // The unified settings center is the only place that uses ?tab=... to
  // switch between sibling sidebar items (plain "settings" vs "runtime").
  // For that path only, a query-less target should not highlight when a tab
  // is present, so the two items don't compete. Other query-less admin pages
  // may use ?tab=... for their own UI and must stay active regardless.
  if (
    Object.keys(target.query).length === 0 &&
    target.pathname === authRouteDefaults.value.adminSettingsPath
  ) {
    const currentTab = route.query.tab
    const currentTabValue = Array.isArray(currentTab) ? currentTab[0] : currentTab
    return !currentTabValue
  }

  return true
}

function isGroupActive(item: NavItem): boolean {
  if (!item.children) return false
  return item.children.some(child => route.path === child.path)
}

function isGroupExpanded(item: NavItem): boolean {
  return expandedGroups.value.has(item.path) || isGroupActive(item)
}

function toggleGroup(item: NavItem) {
  if (expandedGroups.value.has(item.path)) {
    expandedGroups.value.delete(item.path)
  } else {
    expandedGroups.value.add(item.path)
  }
}

/**
 * Click handler for collapsible parent items.
 * - When sidebar is collapsed: do nothing (children are not visible).
 * - When `expandOnly` is true: only toggle expand state.
 * - Otherwise (default, e.g. /admin/orders): navigate to the parent path
 *   (router-link semantics) and ensure the group is expanded.
 */
function handleGroupClick(item: NavItem) {
  if (sidebarCollapsed.value) return
  if (item.expandOnly) {
    toggleGroup(item)
    return
  }
  // Push to path and ensure expanded
  if (route.path !== item.path) {
    router.push(item.path)
  }
  if (!expandedGroups.value.has(item.path)) {
    expandedGroups.value.add(item.path)
  }
}

// Fetch admin settings (for feature-gated nav items like Ops).
watch(
  isAdmin,
  (v) => {
    if (v) {
      adminSettingsStore.fetch()
    }
  },
  { immediate: true }
)

onMounted(() => {
  if (isAdmin.value) {
    adminSettingsStore.fetch()
  }
})
</script>

<style scoped>
.sidebar-logo {
  flex: 0 0 2.25rem;
  min-width: 2.25rem;
}

.sidebar-home-link {
  display: flex;
  min-width: 0;
  flex: 1 1 auto;
  align-items: center;
  gap: 0.75rem;
  border-radius: 0.875rem;
  outline: none;
}

.sidebar-home-link:focus-visible {
  outline: 2px solid rgba(14, 165, 233, 0.65);
  outline-offset: 4px;
}

.sidebar-home-link-collapsed {
  justify-content: center;
  gap: 0;
}

.sidebar-header-collapsed {
  gap: 0;
  padding-left: 1.125rem;
  padding-right: 1.125rem;
}

.sidebar-brand {
  min-width: 0;
  flex: 1 1 auto;
  white-space: nowrap;
  transition:
    max-width 0.22s ease,
    opacity 0.14s ease,
    transform 0.14s ease;
  max-width: 12rem;
}

.sidebar-brand-collapsed {
  max-width: 0;
  overflow: hidden;
  opacity: 0;
  transform: translateX(-4px);
  pointer-events: none;
}

.sidebar-brand-title {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sidebar-link-collapsed {
  gap: 0;
  padding-left: 0.875rem;
  padding-right: 0.875rem;
}

.sidebar-section-title {
  position: relative;
  display: flex;
  align-items: center;
  min-height: 1.25rem;
  overflow: hidden;
  white-space: nowrap;
}

.sidebar-section-title-text {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  transition:
    opacity 0.16s ease,
    transform 0.16s ease;
}

.sidebar-section-title::after {
  content: '';
  position: absolute;
  left: 0.75rem;
  right: 0.75rem;
  top: 50%;
  height: 1px;
  background: rgb(229 231 235);
  opacity: 0;
  transform: translateY(-50%);
  transition: opacity 0.18s ease;
}

.dark .sidebar-section-title::after {
  background: rgb(55 65 81);
}

.sidebar-section-title-text-collapsed {
  opacity: 0;
  transform: translateX(-4px);
}

.sidebar-section-title-collapsed::after {
  opacity: 1;
  transition-delay: 0.08s;
}

.sidebar-label {
  display: block;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  transition:
    max-width 0.2s ease,
    opacity 0.12s ease,
    transform 0.12s ease;
  max-width: 12rem;
}

.sidebar-label-flex {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
}

.sidebar-label-collapsed {
  max-width: 0;
  opacity: 0;
  transform: translateX(-4px);
  pointer-events: none;
}

/* Custom SVG icon in sidebar: constrain size without overriding uploaded SVG colors */
.sidebar-svg-icon {
  color: currentColor;
}

.sidebar-svg-icon :deep(svg) {
  display: block;
  width: 1.25rem;
  height: 1.25rem;
}
</style>
