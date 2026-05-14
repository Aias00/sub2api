<template>
  <div class="docs-page min-h-screen bg-slate-50 text-gray-900 dark:bg-slate-950 dark:text-white">
    <header class="sticky top-0 z-40 border-b border-gray-200/80 bg-white/90 backdrop-blur-xl dark:border-white/10 dark:bg-slate-950/80">
      <div class="mx-auto flex max-w-[1600px] items-center justify-between gap-4 px-4 py-3 md:px-6">
        <div class="flex min-w-0 items-center gap-3">
          <RouterLink
            to="/home"
            class="flex h-11 w-11 items-center justify-center overflow-hidden rounded-2xl border border-primary-100 bg-white shadow-sm shadow-primary-500/10 dark:border-white/10 dark:bg-white/5"
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
            <p class="truncate text-sm font-semibold tracking-[0.24em] text-primary-600 dark:text-primary-300">
              {{ siteName }}
            </p>
            <h1 class="truncate text-lg font-semibold text-gray-950 dark:text-white">
              {{ t('nav.docs') }}
            </h1>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <LocaleSwitcher />
          <router-link
            v-if="isAuthenticated"
            :to="dashboardPath"
            class="inline-flex items-center rounded-xl bg-gray-950 px-3 py-2 text-sm font-medium text-white transition hover:bg-gray-800 sm:px-4 dark:bg-primary-500 dark:text-slate-950 dark:hover:bg-primary-400"
          >
            <span class="hidden sm:inline">{{ t('home.goToDashboard') }}</span>
            <span class="sm:hidden">{{ t('nav.dashboard') }}</span>
          </router-link>
          <router-link
            v-else
            to="/login"
            class="inline-flex items-center rounded-xl bg-gray-950 px-3 py-2 text-sm font-medium text-white transition hover:bg-gray-800 sm:px-4 dark:bg-primary-500 dark:text-slate-950 dark:hover:bg-primary-400"
          >
            {{ t('home.login') }}
          </router-link>
        </div>
      </div>
    </header>

    <main class="mx-auto max-w-[1600px] px-4 py-4 md:px-6 md:py-6">
      <div class="rounded-[32px] border border-white/75 bg-white/92 shadow-[0_30px_80px_-40px_rgba(15,23,42,0.35)] dark:border-white/10 dark:bg-slate-900/70">
        <div class="docsify-shell px-2 py-3 md:px-4">
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
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import docsifyScriptUrl from 'docsify/lib/docsify.min.js?url'
import docsifySearchPluginUrl from 'docsify/lib/plugins/search.min.js?url'
import docsifyZoomImagePluginUrl from 'docsify/lib/plugins/zoom-image.min.js?url'
import docsifyThemeUrl from 'docsify/lib/themes/vue.css?url'
import { normalizeDocsHashPath } from '@/utils/docs'

declare global {
  interface Window {
    $docsify?: Record<string, unknown>
  }
}

const route = useRoute()
const router = useRouter()
const { t, locale } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()

const docsifyRoot = ref<HTMLElement | null>(null)
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const isAuthenticated = computed(() => authStore.isAuthenticated)
const dashboardPath = computed(() => (authStore.isAdmin ? '/admin/dashboard' : '/dashboard'))
const docsBasePath = computed(() => (locale.value === 'en' ? '/docs-content/en/' : '/docs-content/'))
const docsContentVersion = encodeURIComponent(import.meta.env.VITE_DOCS_CONTENT_VERSION || '')
const docsVersionQueryKey = '_docs_v'
const appRouteDocsLinks = new Set(['#/home', '#/dashboard', '#/register', '#/purchase'])

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
    nameLink: '/home',
    basePath: docsBasePath.value,
    homepage: 'README.md',
    loadSidebar: '_sidebar.md',
    subMaxLevel: 3,
    auto2top: true,
    relativePath: true,
    requestHeaders: {
      'Cache-Control': 'no-cache',
      Pragma: 'no-cache',
    },
    themeColor: '#2563eb',
    notFoundPage: true,
    plugins: [docsVersionPlugin],
    search: {
      namespace: `cloudbase-docs-${locale.value}-${docsContentVersion}`,
      placeholder: locale.value === 'zh' ? '搜索文档' : 'Search docs',
      noData: locale.value === 'zh' ? '没有找到结果' : 'No results',
      depth: 4,
    },
  }
}

function withDocsContentVersion(hash: string) {
  const normalizedHash = hash.startsWith('#') ? hash : `#${hash}`
  if (!normalizedHash.startsWith('#/')) {
    return normalizedHash
  }

  const [path, query = ''] = normalizedHash.slice(1).split('?')
  const params = new URLSearchParams(query)
  if (params.get(docsVersionQueryKey) !== docsContentVersion) {
    params.set(docsVersionQueryKey, docsContentVersion)
  }
  const queryString = params.toString()

  return queryString ? `#${path}?${queryString}` : `#${path}`
}

function rewriteDocsLinks() {
  const root = document.querySelector('.docsify-shell')
  if (!root) return
  root.querySelectorAll<HTMLAnchorElement>('a[href^="#/"]').forEach((link) => {
    const href = link.getAttribute('href')
    if (!href) return
    if (appRouteDocsLinks.has(href.split('?')[0] ?? href)) {
      link.setAttribute('href', href.slice(1))
      return
    }
    link.setAttribute('href', withDocsContentVersion(href))
  })
}

function docsVersionPlugin(hook: { mounted: (callback: () => void) => void; doneEach: (callback: () => void) => void }) {
  hook.mounted(rewriteDocsLinks)
  hook.doneEach(rewriteDocsLinks)
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
  const targetHash = withDocsContentVersion(docsHash.value)
  if (force || window.location.hash !== targetHash) {
    window.location.hash = targetHash
  }
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

  const initialHash = docsHash.value
  if (route.path !== '/docs') {
    await router.replace('/docs')
  }
  if (typeof window !== 'undefined') {
    window.location.hash = withDocsContentVersion(initialHash)
  }
  await loadDocsify()
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
  border-right: 1px solid rgba(148, 163, 184, 0.18);
  padding: 0.75rem 1rem 2rem !important;
  overflow-y: auto;
  background: transparent;
  z-index: 1;
}

.docsify-shell :deep(.sidebar .app-name) {
  display: none !important;
}

.docsify-shell :deep(.sidebar-nav) {
  padding-bottom: 0 !important;
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
  color: inherit;
}

.docsify-shell :deep(.app-name-link),
.docsify-shell :deep(.sidebar-nav a),
.docsify-shell :deep(.anchor span) {
  color: inherit;
}

.docsify-shell :deep(.sidebar-nav li.active > a) {
  display: block;
  border-radius: 0.85rem;
  background: rgba(37, 99, 235, 0.1);
  box-shadow: inset 3px 0 0 rgba(37, 99, 235, 0.85);
  color: #2563eb !important;
  font-weight: 700;
  padding-left: 0.75rem;
}

.dark .docsify-shell :deep(.sidebar-nav li.active > a) {
  background: rgba(96, 165, 250, 0.16);
  box-shadow: inset 3px 0 0 rgba(96, 165, 250, 0.9);
  color: #bfdbfe !important;
}

.docsify-shell :deep(.sidebar-nav a:hover) {
  color: #2563eb !important;
}

.dark .docsify-shell :deep(.sidebar-nav a:hover) {
  color: #bfdbfe !important;
}

.docsify-shell :deep(.search) {
  margin-bottom: 1rem;
}

.docsify-shell :deep(.search input) {
  border-radius: 1rem;
  border-color: rgba(148, 163, 184, 0.3);
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
