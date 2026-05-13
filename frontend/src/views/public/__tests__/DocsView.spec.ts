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
    expect(docsViewSource).toContain('import LocaleSwitcher')
    expect(docsViewSource).toContain('<LocaleSwitcher />')
    expect(docsViewSource).toContain("locale.value === 'en' ? '/docs-content/en/' : '/docs-content/'")
    expect(docsViewSource).toContain("loadSidebar: '_sidebar.md'")
    expect(docsViewSource).not.toContain('const docsSidebarPath = computed')
    expect(docsViewSource).not.toContain("'/.*/_sidebar.md'")
    expect(docsViewSource).toContain('window.location.reload()')
  })

  it('adds a build-specific cache buster to Docsify hashes and search index', () => {
    expect(docsViewSource).toContain('const docsContentVersion = encodeURIComponent(__DOCS_CONTENT_VERSION__)')
    expect(docsViewSource).toContain("const docsVersionQueryKey = '_docs_v'")
    expect(docsViewSource).toContain('function withDocsContentVersion')
    expect(docsViewSource).toContain('window.location.hash = withDocsContentVersion(initialHash)')
    expect(docsViewSource).toContain('namespace: `cloudbase-docs-${locale.value}-${docsContentVersion}`')
    expect(docsViewSource).toContain('link.setAttribute')
  })

  it('keeps application route links from being treated as Docsify document routes', () => {
    expect(docsViewSource).toContain("const appRouteDocsLinks = new Set(['#/home', '#/dashboard', '#/register', '#/purchase'])")
    expect(docsViewSource).toContain("link.setAttribute('href', href.slice(1))")
  })

  it('overrides Docsify default fixed layout so the product header stays pinned', () => {
    expect(docsViewSource).toContain(".docsify-shell :deep(main)")
    expect(docsViewSource).toContain('display: flex')
    expect(docsViewSource).toContain(".docsify-shell :deep(.sidebar)")
    expect(docsViewSource).toContain('position: sticky !important')
    expect(docsViewSource).toContain(".docsify-shell :deep(.content)")
    expect(docsViewSource).toContain('margin-left: 17rem !important')
    expect(docsViewSource).toContain(".docsify-shell :deep(.sidebar .app-name)")
    expect(docsViewSource).toContain('display: none !important;')
    expect(docsViewSource).toContain("padding: 0.5rem 0.75rem 1rem !important;")
    expect(docsViewSource).toContain('max-height: 16rem !important;')
    expect(docsViewSource).toContain('padding-top: 0 !important;')
    expect(docsViewSource).toContain('py-4 md:px-6 md:py-6')
    expect(docsViewSource).toContain('body.docs-page-body #app')
    expect(docsViewSource).toContain('margin: 0 !important')
  })

  it('does not render the extra docs intro banner above the Docsify shell', () => {
    expect(docsViewSource).not.toContain("{{ t('docs.frameworkHint') }}")
    expect(docsViewSource).not.toContain("{{ t('nav.docs') }}\n          </h2>")
  })

  it('cleans up Docsify global resources when leaving the docs route', () => {
    expect(docsViewSource).toContain("const docsifyResourceIds = ['docsify-theme', 'docsify-runtime', 'docsify-search-plugin', 'docsify-zoom-image-plugin']")
    expect(docsViewSource).toContain("document.body.classList.remove('ready', 'sticky', 'close')")
    expect(docsViewSource).toContain("document.querySelectorAll('body > .progress').forEach((node) => node.remove())")
  })
})
