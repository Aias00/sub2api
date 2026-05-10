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
})
