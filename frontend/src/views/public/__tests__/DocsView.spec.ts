import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

const docsViewSource = readFileSync(resolve(process.cwd(), 'src/views/public/DocsView.vue'), 'utf8')

describe('DocsView docsify integration', () => {
  it('loads local docsify runtime assets instead of the previous custom markdown registry', () => {
    expect(docsViewSource).toContain('docsify/lib/docsify.min.js?url')
    expect(docsViewSource).toContain("docsify/lib/themes/vue.css?url")
    expect(docsViewSource).toContain('window.$docsify')
    expect(docsViewSource).toContain("basePath: '/docs-content/'")
    expect(docsViewSource).not.toContain('findDocsPage(')
    expect(docsViewSource).not.toContain('marked.parse(')
  })

  it('overrides Docsify default fixed layout so the product header stays pinned', () => {
    expect(docsViewSource).toContain(".docsify-shell :deep(main)")
    expect(docsViewSource).toContain('display: flex')
    expect(docsViewSource).toContain(".docsify-shell :deep(.sidebar)")
    expect(docsViewSource).toContain('position: sticky !important')
    expect(docsViewSource).toContain(".docsify-shell :deep(.content)")
    expect(docsViewSource).toContain('margin-left: 17rem !important')
    expect(docsViewSource).toContain('body.docs-page-body #app')
    expect(docsViewSource).toContain('margin: 0 !important')
  })

  it('does not render the extra docs intro banner above the Docsify shell', () => {
    expect(docsViewSource).not.toContain("{{ t('docs.frameworkHint') }}")
    expect(docsViewSource).not.toContain("{{ t('nav.docs') }}\n          </h2>")
  })
})
