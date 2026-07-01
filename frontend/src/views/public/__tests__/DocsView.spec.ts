import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

const docsViewSource = readFileSync(resolve(process.cwd(), 'src/views/public/DocsView.vue'), 'utf8')

describe('DocsView docsify integration', () => {
  it('loads local docsify runtime assets instead of the previous custom markdown registry', () => {
    expect(docsViewSource).toContain('docsify/lib/docsify.min.js?url')
    expect(docsViewSource).toContain("docsify/lib/themes/vue.css?url")
    expect(docsViewSource).toContain('window.$docsify')
    expect(docsViewSource).toContain('basePath: docsBasePath.value')
    expect(docsViewSource).toContain('requestHeaders:')
    expect(docsViewSource).toContain("'Cache-Control': 'no-cache'")
    expect(docsViewSource).toContain('plugins: [docsVersionPlugin]')
    expect(docsViewSource).not.toContain('findDocsPage(')
    expect(docsViewSource).not.toContain('marked.parse(')
  })

  it('exposes the shared locale switcher and routes Docsify content by locale', () => {
    expect(docsViewSource).toContain('import PublicDarkHeader')
    expect(docsViewSource).toContain('<PublicDarkHeader')
    expect(docsViewSource).toContain('resolveDocsContentBasePath(')
    expect(docsViewSource).toContain('appStore.cachedPublicSettings?.docs_content_base_path')
    expect(docsViewSource).toContain("loadSidebar: '_sidebar.md'")
    expect(docsViewSource).not.toContain('const docsSidebarPath = computed')
    expect(docsViewSource).toContain("'/.*/_sidebar.md': '/_sidebar.md'")
    expect(docsViewSource).toContain('window.location.reload()')
  })

  it('delegates docs content path defaults to public settings', () => {
    expect(docsViewSource).toContain('appStore.cachedPublicSettings?.docs_content_base_path')
    expect(docsViewSource).not.toContain('/docs-content/')
    expect(docsViewSource).not.toContain('/docs-content/en/')
  })

  it('adds a build-specific cache buster to Docsify hashes and search index', () => {
    expect(docsViewSource).toContain('appStore.cachedPublicSettings?.version')
    expect(docsViewSource).not.toContain('VITE_DOCS_CONTENT_VERSION')
    expect(docsViewSource).toContain("const docsVersionQueryKey = '_docs_v'")
    expect(docsViewSource).toContain("from './docsRuntime'")
    expect(docsViewSource).toContain('withDocsContentVersion')
    expect(docsViewSource).toContain('window.location.hash = withDocsContentVersion(initialHash, docsContentVersion.value, docsVersionQueryKey)')
    expect(docsViewSource).toContain('namespace: docsSearchNamespace.value')
    expect(docsViewSource).toContain('link.setAttribute')
  })

  it('derives Docsify search namespace from runtime site settings instead of local branding', () => {
    expect(docsViewSource).toContain('const docsSearchNamespace = computed')
    expect(docsViewSource).toContain('buildDocsSearchNamespace')
    expect(docsViewSource).not.toContain('cloudbase-docs')
  })

  it('preserves direct Docsify hash deep links on initial load', () => {
    expect(docsViewSource).toContain('function getInitialDocsHash')
    expect(docsViewSource).toContain('resolveInitialDocsHash')
    expect(docsViewSource).toContain('const initialHash = getInitialDocsHash()')
  })

  it('keeps application route links from being treated as Docsify document routes', () => {
    expect(docsViewSource).toContain('resolveDocsShellConfig(')
    expect(docsViewSource).toContain('docsShellConfig.value.defaults.appRouteLinks')
    expect(docsViewSource).toContain('appRouteDocsLinks.value.has')
    expect(docsViewSource).not.toContain("const appRouteDocsLinks = new Set(['#/home', '#/dashboard', '#/register', '#/purchase'])")
    expect(docsViewSource).toContain("document.querySelector('.docsify-shell')")
    expect(docsViewSource).toContain("link.setAttribute('href', href.slice(1))")
  })

  it('overrides Docsify default fixed layout so the product header stays pinned', () => {
    expect(docsViewSource).toContain(".docsify-shell :deep(main)")
    expect(docsViewSource).toContain('display: flex')
    expect(docsViewSource).toContain(".docsify-shell :deep(.sidebar)")
    expect(docsViewSource).toContain('position: sticky !important')
    expect(docsViewSource).toContain(".docsify-shell :deep(.content)")
    expect(docsViewSource).toContain('margin-left: 0 !important')
    expect(docsViewSource).toContain(".docsify-shell :deep(.sidebar .app-name)")
    expect(docsViewSource).toContain('display: none !important;')
    expect(docsViewSource).toContain("padding: 0.65rem 0.75rem !important;")
    expect(docsViewSource).toContain('max-height: 12rem !important;')
    expect(docsViewSource).toContain('padding-top: 0 !important;')
    expect(docsViewSource).toContain('public-template-main')
    expect(docsViewSource).toContain('body.docs-page-body #app')
    expect(docsViewSource).toContain('margin: 0 !important')
  })

  it('uses the public Vercel template shell for docs pages', () => {
    expect(docsViewSource).toContain('home-business-page public-template-page')
    expect(docsViewSource).toContain('public-template-container')
    expect(docsViewSource).toContain('docsHeroDescription')
    expect(docsViewSource).toContain('docs-template-shell overflow-hidden rounded-2xl border')
    expect(docsViewSource).toContain('background: var(--public-panel)')
    expect(docsViewSource).toContain('color: var(--public-body)')
    expect(docsViewSource).toContain('background: var(--public-panel-soft)')
    expect(docsViewSource).toContain('border-radius: 0.75rem')
    expect(docsViewSource).not.toContain('rounded-[36px]')
    expect(docsViewSource).not.toContain('bg-white/92')
  })

  it('does not render the extra docs intro banner above the Docsify shell', () => {
    expect(docsViewSource).not.toContain("{{ t('docs.frameworkHint') }}")
    expect(docsViewSource).not.toContain("{{ t('nav.docs') }}\n          </h2>")
  })

  it('reads docs shell labels from public settings before configuring Docsify', () => {
    expect(docsViewSource).toContain('resolveDocsShellConfig(')
    expect(docsViewSource).toContain('appStore.cachedPublicSettings?.docs_shell_config')
    expect(docsViewSource).toContain('useAuthRouteDefaults')
    expect(docsViewSource).toContain('<PublicDarkHeader')
    expect(docsViewSource).toContain('nameLink: authRouteDefaults.value.homePath')
    expect(docsViewSource).not.toContain('to="/home"')
    expect(docsViewSource).not.toContain("nameLink: '/home'")
    expect(docsViewSource).not.toContain('to="/login"')
    expect(docsViewSource).not.toContain("authStore.isAdmin ? '/admin/dashboard' : '/dashboard'")
    expect(docsViewSource).toContain('isAuthenticated ? copy.dashboard : copy.login')
    expect(docsViewSource).toContain('placeholder: copy.value.searchPlaceholder')
    expect(docsViewSource).toContain('noData: copy.value.noData')
  })

  it('keeps the header brand aligned with other public pages', () => {
    expect(docsViewSource).toContain('PublicDarkHeader')
    expect(docsViewSource).toContain('max-w-6xl')
    expect(docsViewSource).toContain('body.docs-page-body .public-dark-header nav')
    expect(docsViewSource).toContain('position: static !important')
    expect(docsViewSource).toContain('max-width: 72rem !important')
    expect(docsViewSource).toContain('body.docs-page-body .public-dark-header__brand')
    expect(docsViewSource).not.toContain('max-w-[1600px]')
    expect(docsViewSource).not.toContain('<header class="sticky')
    expect(docsViewSource).toContain('{{ copy.title }}')
    expect(docsViewSource).not.toContain('uppercase tracking-[0.24em] text-sky-600')
  })

  it('does not keep locale-specific docs fallback copy in the view bootstrap layer', () => {
    expect(docsViewSource).not.toContain('EMPTY_DOCS_COPY')
    expect(docsViewSource).not.toContain('DEFAULT_DOCS_COPY')
    expect(docsViewSource).not.toContain("title: '文档'")
    expect(docsViewSource).not.toContain("dashboard: '控制台'")
    expect(docsViewSource).not.toContain("searchPlaceholder: '搜索文档'")
    expect(docsViewSource).not.toContain("title: 'Docs'")
    expect(docsViewSource).not.toContain("searchPlaceholder: 'Search docs'")
    expect(docsViewSource).not.toContain('type DocsShellCopy =')
    expect(docsViewSource).not.toContain('const docsShellCopyKeys')
    expect(docsViewSource).not.toContain('resolveLocalizedShellLabels(')
  })

  it('does not keep the legacy docs locale section in frontend bundles', () => {
    const zhLocaleSource = readFileSync(resolve(process.cwd(), 'src/i18n/locales/zh.ts'), 'utf8')
    const enLocaleSource = readFileSync(resolve(process.cwd(), 'src/i18n/locales/en.ts'), 'utf8')
    for (const source of [zhLocaleSource, enLocaleSource]) {
      expect(source).not.toContain('\n  docs: {')
      expect(source).not.toContain('frameworkHint')
      expect(source).not.toContain('On this page')
      expect(source).not.toContain('本页目录')
    }
  })

  it('cleans up Docsify global resources when leaving the docs route', () => {
    expect(docsViewSource).toContain("const docsifyResourceIds = ['docsify-theme', 'docsify-runtime', 'docsify-search-plugin', 'docsify-zoom-image-plugin']")
    expect(docsViewSource).toContain("document.body.classList.remove('ready', 'sticky', 'close')")
    expect(docsViewSource).toContain("document.querySelectorAll('body > .progress').forEach((node) => node.remove())")
  })
})
