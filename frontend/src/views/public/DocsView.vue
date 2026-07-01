<template>
  <div class="docs-page home-business-page public-template-page relative min-h-screen overflow-x-hidden">
    <PublicDarkHeader
      :account-label="isAuthenticated ? copy.dashboard : copy.login"
      container-class="max-w-6xl"
    />

    <main class="public-template-main">
      <div class="public-template-container">
        <section class="mb-8">
          <p class="text-sm font-semibold uppercase tracking-[0.22em] text-[var(--public-muted)]">
            {{ docsHeroEyebrow }}
          </p>
          <h1 class="mt-4 max-w-4xl text-4xl font-black leading-tight text-[var(--public-ink)] sm:text-5xl">
            {{ copy.title }}
          </h1>
          <p class="mt-4 max-w-3xl text-base leading-8 text-[var(--public-body)]">
            {{ docsHeroDescription }}
          </p>
        </section>

        <div class="docs-template-shell overflow-hidden rounded-2xl border border-[var(--public-border)] bg-[var(--public-panel)] shadow-[0_18px_48px_rgba(0,0,0,0.08)]">
          <div class="docsify-shell">
            <div id="docsify-app" ref="docsifyRoot" class="min-h-[70vh]"></div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'
import { useAuthRouteDefaults } from '@/composables/useAuthRouteDefaults'
import PublicDarkHeader from '@/components/layout/PublicDarkHeader.vue'
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
const { authRouteDefaults } = useAuthRouteDefaults()

const docsifyRoot = ref<HTMLElement | null>(null)
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName)
const isAuthenticated = computed(() => authStore.isAuthenticated)
const docsLocale = computed<'zh' | 'en'>(() => resolveRuntimeLanguage(locale))
const docsBasePath = computed(() =>
  resolveDocsContentBasePath(appStore.cachedPublicSettings?.docs_content_base_path, docsLocale.value),
)
const docsShellConfig = computed(() =>
  resolveDocsShellConfig(appStore.cachedPublicSettings?.docs_shell_config, docsLocale.value),
)
const copy = computed(() => docsShellConfig.value.labels)
const docsHeroEyebrow = computed(() => (docsLocale.value === 'zh' ? 'Documentation' : 'Documentation'))
const docsHeroDescription = computed(() =>
  docsLocale.value === 'zh'
    ? '查看部署、控制台、业务能力和模型接入指南，快速定位当前需要的配置与流程。'
    : 'Browse deployment, console, business capability, and model integration guides in one place.',
)
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
    themeColor: '#0284c7',
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

.docs-template-shell {
  background: var(--public-panel);
  color: var(--public-ink);
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
  top: 1rem !important;
  left: auto !important;
  bottom: auto !important;
  flex: 0 0 17rem;
  width: 17rem !important;
  height: calc(100vh - 9rem) !important;
  margin: 1rem 0 1rem 1rem;
  border: 1px solid var(--public-border);
  border-radius: 1rem;
  padding: 0.85rem 0.9rem 1rem !important;
  overflow-y: auto;
  background: var(--public-panel-soft);
  z-index: 1;
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
  background: transparent;
}

.docsify-shell :deep(.markdown-section) {
  max-width: min(900px, 100%);
  min-height: 70vh;
  margin: 0;
  padding: 2rem 2.5rem 3rem;
  color: var(--public-body);
}

.docsify-shell :deep(.markdown-section h1),
.docsify-shell :deep(.markdown-section h2),
.docsify-shell :deep(.markdown-section h3),
.docsify-shell :deep(.markdown-section h4) {
  color: var(--public-ink);
  font-weight: 800;
  letter-spacing: 0;
}

.docsify-shell :deep(.markdown-section h1) {
  margin-top: 0;
  font-size: clamp(2rem, 3.2vw, 2.85rem);
  line-height: 1.12;
}

.docsify-shell :deep(.markdown-section h2) {
  margin-top: 2.25rem;
  border-top: 1px solid var(--public-border);
  padding-top: 1.5rem;
  font-size: 1.5rem;
}

.docsify-shell :deep(.markdown-section h3) {
  font-size: 1.125rem;
}

.docsify-shell :deep(.markdown-section p),
.docsify-shell :deep(.markdown-section li) {
  color: var(--public-body);
  line-height: 1.8;
}

.docsify-shell :deep(.markdown-section a) {
  color: var(--public-accent-strong);
  text-decoration: none;
}

.docsify-shell :deep(.markdown-section a:hover) {
  color: var(--public-ink);
}

.docsify-shell :deep(.app-name-link),
.docsify-shell :deep(.sidebar-nav a),
.docsify-shell :deep(.anchor span) {
  color: inherit;
}

.docsify-shell :deep(.sidebar-nav) {
  color: var(--public-body);
}

.docsify-shell :deep(.sidebar-nav > ul > li) {
  margin-bottom: 0.75rem;
}

.docsify-shell :deep(.sidebar-nav > ul > li > p),
.docsify-shell :deep(.sidebar-nav > ul > li > strong) {
  margin: 0.8rem 0 0.35rem;
  color: var(--public-muted);
  font-size: 0.75rem;
  font-weight: 800;
  letter-spacing: 0;
}

.docsify-shell :deep(.sidebar-nav a) {
  display: flex;
  width: 100%;
  min-height: 2rem;
  align-items: center;
  overflow: hidden;
  border: 1px solid transparent;
  border-radius: 0.75rem;
  padding: 0.36rem 0.65rem !important;
  color: var(--public-body) !important;
  font-size: 0.9rem;
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
  border: 1px solid var(--public-border-strong);
  border-radius: 0.75rem;
  background: var(--public-panel);
  color: var(--public-ink) !important;
  font-weight: 800;
  line-height: 1.25;
  text-decoration: none;
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.04);
}

.docsify-shell :deep(.sidebar-nav a:hover) {
  border-color: var(--public-border);
  background: var(--public-panel-soft);
  color: var(--public-ink) !important;
}

.docsify-shell :deep(.search) {
  margin-bottom: 1rem;
}

.docsify-shell :deep(.search input) {
  border-radius: 0.875rem;
  border-color: var(--public-border);
  background: var(--public-panel);
  box-shadow: none;
  color: var(--public-ink);
}

.docsify-shell :deep(.search input::placeholder) {
  color: var(--public-faint);
}

.docsify-shell :deep(.sidebar-toggle) {
  display: none;
}

.docsify-shell :deep(.markdown-section code) {
  border: 1px solid var(--public-border);
  border-radius: 0.375rem;
  background: var(--public-panel-muted);
  color: var(--public-ink);
}

.docsify-shell :deep(.markdown-section pre) {
  border: 1px solid var(--public-border);
  border-radius: 0.875rem;
  background: var(--public-panel-muted);
}

.docsify-shell :deep(.markdown-section blockquote) {
  border-left: 3px solid var(--public-border-strong);
  color: var(--public-body);
}

.docsify-shell :deep(.markdown-section table) {
  overflow: hidden;
  border: 1px solid var(--public-border);
  border-radius: 0.875rem;
}

.docsify-shell :deep(.markdown-section th) {
  background: var(--public-panel-muted);
  color: var(--public-ink);
}

.docsify-shell :deep(.markdown-section td),
.docsify-shell :deep(.markdown-section th) {
  border-color: var(--public-border);
}

@media (max-width: 960px) {
  .docsify-shell :deep(main) {
    display: block;
  }

  .docsify-shell :deep(.sidebar) {
    position: relative !important;
    width: calc(100% - 2rem) !important;
    max-height: 12rem !important;
    height: auto !important;
    margin: 1rem;
    flex-basis: auto;
    border-right: 1px solid var(--public-border);
    border-bottom: 1px solid var(--public-border);
    padding: 0.65rem 0.75rem !important;
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
    padding: 1.25rem 1.25rem 2rem;
  }

  .docsify-shell :deep(.markdown-section h1) {
    font-size: 2rem;
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

body.docs-page-body .public-dark-header nav {
  position: static !important;
  inset: auto !important;
  display: flex !important;
  width: 100% !important;
  max-width: 72rem !important;
  height: auto !important;
  margin: 0 auto !important;
  padding: 0 !important;
}

body.docs-page-body .public-dark-header__brand {
  margin: 0 !important;
  padding: 0 !important;
}

body.docs-page-body.sticky .sidebar,
body.docs-page-body.sticky .sidebar-toggle {
  position: sticky !important;
}
</style>
