<template>
  <div class="docs-page relative min-h-screen overflow-x-hidden bg-white text-slate-900 dark:bg-dark-950 dark:text-white">
    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <div class="absolute inset-x-0 top-0 h-[28rem] bg-[radial-gradient(circle_at_18%_16%,rgba(125,211,252,0.22),transparent_24%),radial-gradient(circle_at_86%_0%,rgba(191,219,254,0.2),transparent_26%),linear-gradient(180deg,rgba(248,250,252,0.95),rgba(255,255,255,0))] dark:bg-[radial-gradient(circle_at_18%_16%,rgba(56,189,248,0.12),transparent_24%),radial-gradient(circle_at_86%_0%,rgba(59,130,246,0.1),transparent_26%),linear-gradient(180deg,rgba(15,23,42,0.9),rgba(2,6,23,0))]"></div>
      <div class="absolute inset-0 bg-[linear-gradient(rgba(148,163,184,0.08)_1px,transparent_1px),linear-gradient(90deg,rgba(148,163,184,0.08)_1px,transparent_1px)] bg-[size:72px_72px] opacity-30 dark:opacity-10"></div>
    </div>

    <header class="sticky top-0 z-40 border-b border-slate-200/80 bg-white/92 backdrop-blur-xl dark:border-dark-700/80 dark:bg-dark-950/88">
      <div class="mx-auto flex max-w-[1600px] items-center justify-between gap-4 px-4 py-3 md:px-6">
        <div class="flex min-w-0 items-center gap-3">
          <RouterLink
            :to="authRouteDefaults.homePath"
            class="flex h-11 w-11 items-center justify-center overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm shadow-slate-900/5 dark:border-dark-700 dark:bg-white/5"
          >
            <img
              v-if="siteLogo"
              :src="siteLogo"
              :alt="siteName"
              class="h-full w-full object-cover"
            >
            <span
              v-else
              class="text-sm font-semibold tracking-[0.18em] text-primary-600 dark:text-primary-300"
            >
              {{ siteName.slice(0, 2).toUpperCase() }}
            </span>
          </RouterLink>
          <div class="min-w-0">
            <p class="truncate text-xs font-semibold uppercase tracking-[0.24em] text-sky-600 dark:text-sky-300">
              {{ siteName }}
            </p>
            <h1 class="truncate text-lg font-semibold text-slate-950 dark:text-white">
              {{ copy.title }}
            </h1>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <LocaleSwitcher />
          <router-link
            v-if="isAuthenticated"
            :to="dashboardPath"
            class="inline-flex items-center rounded-2xl bg-slate-950 px-3 py-2 text-sm font-semibold text-white transition hover:bg-slate-800 sm:px-4 dark:bg-white dark:text-slate-950 dark:hover:bg-slate-100"
          >
            <span class="hidden sm:inline">{{ copy.dashboard }}</span>
            <span class="sm:hidden">{{ copy.dashboard }}</span>
          </router-link>
          <router-link
            v-else
            :to="loginPath"
            class="inline-flex items-center rounded-2xl bg-slate-950 px-3 py-2 text-sm font-semibold text-white transition hover:bg-slate-800 sm:px-4 dark:bg-white dark:text-slate-950 dark:hover:bg-slate-100"
          >
            {{ copy.login }}
          </router-link>
        </div>
      </div>
    </header>

    <main class="relative z-10 mx-auto max-w-[1600px] px-4 py-4 md:px-6 md:py-6">
      <div class="rounded-[36px] border border-slate-200/80 bg-white shadow-[0_28px_90px_-54px_rgba(15,23,42,0.22)] dark:border-dark-700 dark:bg-dark-950/72">
        <div class="docsify-shell px-2 py-3 md:px-4 md:py-4">
          <div id="docsify-app" ref="docsifyRoot" class="min-h-[70vh]"></div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'
import { useAuthRouteDefaults } from '@/composables/useAuthRouteDefaults'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import docsifyScriptUrl from 'docsify/lib/docsify.min.js?url'
import docsifySearchPluginUrl from 'docsify/lib/plugins/search.min.js?url'
import docsifyZoomImagePluginUrl from 'docsify/lib/plugins/zoom-image.min.js?url'
import docsifyThemeUrl from 'docsify/lib/themes/vue.css?url'
import { normalizeDocsHashPath } from '@/utils/docs'
import { resolveDocsContentBasePath } from '@/utils/docsContentBasePath'
import { resolveDocsShellConfig } from '@/utils/docsShell'
import { resolveRuntimeLanguage } from '@/utils/runtimeLocale'
import {
  buildDocsSearchNamespace,
  getDocsHashPath,
  resolveInitialDocsHash,
  withDocsContentVersion,
} from './docsRuntime'

declare global {
  interface Window {
    $docsify?: Record<string, unknown>
  }
}

const route = useRoute()
const router = useRouter()
const { locale } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const { authRouteDefaults, resolveHomePath } = useAuthRouteDefaults()

const docsifyRoot = ref<HTMLElement | null>(null)
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName)
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const isAuthenticated = computed(() => authStore.isAuthenticated)
const dashboardPath = computed(() => resolveHomePath(authStore.isAdmin))
const loginPath = computed(() => authRouteDefaults.value.loginPath)
const docsLocale = computed<'zh' | 'en'>(() => resolveRuntimeLanguage(locale))
const docsBasePath = computed(() =>
  resolveDocsContentBasePath(appStore.cachedPublicSettings?.docs_content_base_path, docsLocale.value),
)
const docsShellConfig = computed(() =>
  resolveDocsShellConfig(appStore.cachedPublicSettings?.docs_shell_config, docsLocale.value),
)
const copy = computed(() => docsShellConfig.value.labels)
const docsContentVersion = computed(() =>
  encodeURIComponent(appStore.cachedPublicSettings?.version || ''),
)
const docsSearchNamespace = computed(() =>
  buildDocsSearchNamespace([siteName.value, locale.value, docsContentVersion.value]),
)
const docsVersionQueryKey = '_docs_v'
const appRouteDocsLinks = computed(() => new Set(docsShellConfig.value.defaults.appRouteLinks))

const docsHash = computed(() => normalizeDocsHashPath(route.params.pathMatch as string | string[] | undefined))

let docsifyLoaded = false
const docsifyResourceIds = ['docsify-theme', 'docsify-runtime', 'docsify-search-plugin', 'docsify-zoom-image-plugin']

function ensureStylesheet(id: string, href: string) {
  if (document.getElementById(id)) return
  const link = document.createElement('link')
  link.id = id
  link.rel = 'stylesheet'
  link.href = href
  document.head.appendChild(link)
}

async function ensureScript(id: string, src: string) {
  const existing = document.getElementById(id) as HTMLScriptElement | null
  if (existing) {
    if ((existing as HTMLScriptElement).dataset.loaded === 'true') return
    await new Promise<void>((resolve, reject) => {
      existing.addEventListener('load', () => resolve(), { once: true })
      existing.addEventListener('error', () => reject(new Error(`Failed to load ${src}`)), { once: true })
    })
    return
  }

  await new Promise<void>((resolve, reject) => {
    const script = document.createElement('script')
    script.id = id
    script.src = src
    script.async = false
    script.dataset.loaded = 'false'
    script.addEventListener(
      'load',
      () => {
        script.dataset.loaded = 'true'
        resolve()
      },
      { once: true }
    )
    script.addEventListener('error', () => reject(new Error(`Failed to load ${src}`)), {
      once: true,
    })
    document.body.appendChild(script)
  })
}

function configureDocsify() {
  window.$docsify = {
    el: '#docsify-app',
    name: siteName.value,
    nameLink: authRouteDefaults.value.homePath,
    basePath: docsBasePath.value,
    homepage: 'README.md',
    loadSidebar: '_sidebar.md',
    alias: {
      '/.*/_sidebar.md': '/_sidebar.md',
    },
    subMaxLevel: 0,
    auto2top: true,
    relativePath: false,
    requestHeaders: {
      'Cache-Control': 'no-cache',
      Pragma: 'no-cache',
    },
    themeColor: '#4c409c',
    notFoundPage: true,
    plugins: [docsVersionPlugin],
    search: {
      namespace: docsSearchNamespace.value,
      placeholder: copy.value.searchPlaceholder,
      noData: copy.value.noData,
      depth: 4,
    },
  }
}

function rewriteDocsLinks() {
  const root = document.querySelector('.docsify-shell')
  if (!root) return
  root.querySelectorAll<HTMLAnchorElement>('a[href^="#/"]').forEach((link) => {
    const href = link.getAttribute('href')
    if (!href) return
    if (appRouteDocsLinks.value.has(href.split('?')[0] ?? href)) {
      link.setAttribute('href', href.slice(1))
      return
    }
    link.setAttribute('href', withDocsContentVersion(href, docsContentVersion.value, docsVersionQueryKey))
  })
}

function syncSidebarActiveLink() {
  const root = document.querySelector('.docsify-shell')
  if (!root) return

  const currentPath = getDocsHashPath(window.location.hash)
  const links = Array.from(root.querySelectorAll<HTMLAnchorElement>('.sidebar-nav a[href^="#/"]'))
  root.querySelectorAll('.sidebar-nav li.active').forEach((item) => item.classList.remove('active'))
  root.querySelectorAll('.sidebar-nav a.active').forEach((link) => link.classList.remove('active'))

  const activeLink = links.find((link) => getDocsHashPath(link.getAttribute('href') || '') === currentPath)
  if (!activeLink) return

  activeLink.classList.add('active')
  activeLink.closest('li')?.classList.add('active')

  const sidebar = activeLink.closest('.sidebar') as HTMLElement | null
  const sidebarNav = activeLink.closest('.sidebar-nav') as HTMLElement | null
  if (!sidebar) return
  const scrollContainer =
    sidebarNav && sidebarNav.scrollHeight > sidebarNav.clientHeight + 1 ? sidebarNav : sidebar
  const linkRect = activeLink.getBoundingClientRect()
  const sidebarRect = scrollContainer.getBoundingClientRect()
  const safePadding = 24
  if (linkRect.top < sidebarRect.top + safePadding) {
    scrollContainer.scrollTop -= sidebarRect.top + safePadding - linkRect.top
  } else if (linkRect.bottom > sidebarRect.bottom - safePadding) {
    scrollContainer.scrollTop += linkRect.bottom - sidebarRect.bottom + safePadding
  }
}

function docsVersionPlugin(hook: { mounted: (callback: () => void) => void; doneEach: (callback: () => void) => void }) {
  const syncDocsLinks = () => {
    rewriteDocsLinks()
    syncSidebarActiveLink()
    window.requestAnimationFrame(() => {
      rewriteDocsLinks()
      syncSidebarActiveLink()
    })
    window.setTimeout(() => {
      rewriteDocsLinks()
      syncSidebarActiveLink()
    }, 120)
    window.setTimeout(() => {
      rewriteDocsLinks()
      syncSidebarActiveLink()
    }, 500)
  }
  hook.mounted(syncDocsLinks)
  hook.doneEach(syncDocsLinks)
}

async function loadDocsify() {
  ensureStylesheet('docsify-theme', docsifyThemeUrl)
  configureDocsify()
  await ensureScript('docsify-runtime', docsifyScriptUrl)
  await ensureScript('docsify-search-plugin', docsifySearchPluginUrl)
  await ensureScript('docsify-zoom-image-plugin', docsifyZoomImagePluginUrl)
  docsifyLoaded = true
}

function syncHash(force = false) {
  if (typeof window === 'undefined') return
  const targetHash = withDocsContentVersion(docsHash.value, docsContentVersion.value, docsVersionQueryKey)
  if (force || window.location.hash !== targetHash) {
    window.location.hash = targetHash
  }
}

function getInitialDocsHash() {
  if (typeof window === 'undefined') return docsHash.value
  return resolveInitialDocsHash(route.path, window.location.hash, docsHash.value)
}

watch(
  docsHash,
  () => {
    if (docsifyLoaded) {
      syncHash()
    }
  },
  { immediate: false }
)

watch(locale, () => {
  if (!docsifyLoaded) return
  window.location.reload()
})

onMounted(async () => {
  document.body.classList.add('docs-page-body')

  if (!appStore.publicSettingsLoaded) {
    await appStore.fetchPublicSettings()
  }

  const initialHash = getInitialDocsHash()
  if (route.path !== '/docs') {
    await router.replace('/docs')
  }
  if (typeof window !== 'undefined') {
    window.location.hash = withDocsContentVersion(initialHash, docsContentVersion.value, docsVersionQueryKey)
  }
  await loadDocsify()
  window.addEventListener('hashchange', syncSidebarActiveLink)
})

onBeforeUnmount(() => {
  docsifyLoaded = false
  docsifyResourceIds.forEach((id) => {
    document.getElementById(id)?.remove()
  })
  document.querySelectorAll('body > .progress').forEach((node) => node.remove())
  document.querySelectorAll('body > iframe').forEach((node) => {
    const iframe = node as HTMLIFrameElement
    if (iframe.src === 'about:blank') {
      iframe.remove()
    }
  })
  document.body.classList.remove('ready', 'sticky', 'close')
  document.body.classList.remove('docs-page-body')
  window.removeEventListener('hashchange', syncSidebarActiveLink)
  delete window.$docsify
})
</script>

<style scoped>
.docs-page {
  min-height: 100vh;
}

.docsify-shell :deep(section.cover) {
  display: none;
}

.docsify-shell :deep(main) {
  position: relative !important;
  display: flex;
  align-items: flex-start;
  width: 100% !important;
  min-height: 70vh;
  height: auto !important;
}

.docsify-shell :deep(.sidebar) {
  position: sticky !important;
  top: 0 !important;
  left: auto !important;
  bottom: auto !important;
  flex: 0 0 15.5rem;
  width: 15.5rem !important;
  height: calc(100vh - 12rem) !important;
  border-right: 1px solid rgba(226, 232, 240, 0.96);
  padding: 0.75rem 1rem 2rem !important;
  overflow-y: auto;
  background: linear-gradient(180deg, rgba(248, 250, 252, 0.96), rgba(255, 255, 255, 0.92));
  z-index: 1;
}

.dark .docsify-shell :deep(.sidebar) {
  border-right-color: rgba(51, 65, 85, 0.82);
  background: linear-gradient(180deg, rgba(15, 23, 42, 0.76), rgba(2, 6, 23, 0.7));
}

.docsify-shell :deep(.sidebar .app-name) {
  display: none !important;
}

.docsify-shell :deep(.sidebar-nav) {
  padding-bottom: 0 !important;
}

.docsify-shell :deep(.sidebar-nav ul) {
  list-style: none;
  margin: 0.15rem 0 0.45rem;
  padding-left: 0.7rem;
}

.docsify-shell :deep(.sidebar-nav > ul) {
  padding-left: 0;
}

.docsify-shell :deep(.sidebar-nav li) {
  list-style: none;
  margin: 0.12rem 0;
}

.docsify-shell :deep(.sidebar-nav li::marker) {
  content: '';
}

.docsify-shell :deep(.content) {
  position: relative !important;
  left: auto !important;
  right: auto !important;
  flex: 1 1 auto;
  min-width: 0;
  margin-left: 0 !important;
  padding-top: 0 !important;
  width: auto !important;
}

.docsify-shell :deep(.markdown-section) {
  max-width: min(980px, 100%);
  min-height: 70vh;
  margin: 0;
  padding: 1.5rem 2.5rem 3rem;
}

.docsify-shell :deep(.markdown-section h1),
.docsify-shell :deep(.markdown-section h2),
.docsify-shell :deep(.markdown-section h3),
.docsify-shell :deep(.markdown-section h4) {
  color: #0f172a;
  font-weight: 800;
}

.dark .docsify-shell :deep(.markdown-section h1),
.dark .docsify-shell :deep(.markdown-section h2),
.dark .docsify-shell :deep(.markdown-section h3),
.dark .docsify-shell :deep(.markdown-section h4) {
  color: #f8fafc;
}

.docsify-shell :deep(.markdown-section a) {
  color: #0369a1;
}

.docsify-shell :deep(.markdown-section a:hover) {
  color: #0f172a;
}

.dark .docsify-shell :deep(.markdown-section a:hover) {
  color: #e2e8f0;
}

.docsify-shell :deep(.app-name-link),
.docsify-shell :deep(.sidebar-nav a),
.docsify-shell :deep(.anchor span) {
  color: inherit;
}

.docsify-shell :deep(.sidebar-nav a) {
  display: flex;
  width: 100%;
  min-height: 2.15rem;
  align-items: center;
  overflow: hidden;
  border: 1px solid transparent;
  border-radius: 999px;
  padding: 0.42rem 0.75rem !important;
  text-overflow: ellipsis;
  white-space: nowrap;
  text-decoration: none;
  transition:
    background-color 160ms ease,
    border-color 160ms ease,
    color 160ms ease;
}

.docsify-shell :deep(.sidebar-nav > ul > li > a) {
  font-weight: 700;
}

.docsify-shell :deep(.sidebar-nav li.active > a) {
  border: 1px solid rgba(56, 189, 248, 0.35);
  border-radius: 999px;
  background: rgba(224, 242, 254, 0.88);
  box-shadow: 0 8px 20px -18px rgba(14, 165, 233, 0.8);
  color: #0f172a !important;
  font-weight: 800;
  line-height: 1.25;
  text-decoration: none;
}

.dark .docsify-shell :deep(.sidebar-nav li.active > a) {
  border-color: rgba(56, 189, 248, 0.38);
  background: rgba(14, 165, 233, 0.14);
  box-shadow: none;
  color: #e0f2fe !important;
}

.docsify-shell :deep(.sidebar-nav a:hover) {
  border-color: rgba(148, 163, 184, 0.28);
  background: rgba(248, 250, 252, 0.96);
  color: #0f172a !important;
}

.dark .docsify-shell :deep(.sidebar-nav a:hover) {
  background: rgba(15, 23, 42, 0.72);
  color: #e2e8f0 !important;
}

.docsify-shell :deep(.search) {
  margin-bottom: 1rem;
}

.docsify-shell :deep(.search input) {
  border-radius: 999px;
  border-color: rgba(203, 213, 225, 0.92);
  background: rgba(255, 255, 255, 0.98);
  box-shadow: inset 0 1px 2px rgba(15, 23, 42, 0.04);
  color: #0f172a;
}

.dark .docsify-shell :deep(.search input) {
  border-color: rgba(51, 65, 85, 0.9);
  background: rgba(15, 23, 42, 0.78);
  color: #f8fafc;
}

.docsify-shell :deep(.sidebar-toggle) {
  display: none;
}

.docsify-shell :deep(.markdown-section code) {
  color: inherit;
}

@media (max-width: 960px) {
  .docsify-shell :deep(main) {
    display: block;
  }

  .docsify-shell :deep(.sidebar) {
    position: relative !important;
    width: 100% !important;
    max-height: 16rem !important;
    height: auto !important;
    flex-basis: auto;
    border-right: 0;
    border-bottom: 1px solid rgba(148, 163, 184, 0.18);
    padding: 0.5rem 0.75rem 1rem !important;
    overflow-y: auto;
  }

  .docsify-shell :deep(.sidebar-nav) {
    padding-top: 0 !important;
  }

  .docsify-shell :deep(.sidebar-nav > ul),
  .docsify-shell :deep(.sidebar-nav > ul > li:first-child) {
    margin-top: 0 !important;
  }

  .docsify-shell :deep(.content) {
    margin-left: 0 !important;
    width: 100% !important;
  }

  .docsify-shell :deep(.markdown-section) {
    padding: 1rem 1.25rem 2rem;
  }
}
</style>

<style>
body.docs-page-body #app {
  margin: 0 !important;
  text-align: initial !important;
  font-size: inherit !important;
  font-weight: inherit !important;
}

body.docs-page-body {
  position: static !important;
  top: auto !important;
}

body.docs-page-body.sticky .sidebar,
body.docs-page-body.sticky .sidebar-toggle {
  position: sticky !important;
}
</style>
