import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

const routerSource = readFileSync(resolve(process.cwd(), 'src/router/index.ts'), 'utf8')

describe('docs route registration', () => {
  it('registers a public docs route backed by DocsView', () => {
    expect(routerSource).toContain("path: '/docs/:pathMatch(.*)*'")
    expect(routerSource).toContain("name: 'Docs'")
    expect(routerSource).toContain("component: () => import('@/views/public/DocsView.vue')")
    expect(routerSource).toContain("titleKey: 'nav.docs'")
  })
})
