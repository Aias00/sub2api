import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

const source = readFileSync(resolve(process.cwd(), 'src/components/common/LocaleSwitcher.vue'), 'utf8')

describe('LocaleSwitcher visual treatment', () => {
  it('uses a compact docs-style trigger with globe icon and uppercase locale code', () => {
    expect(source).toContain("class=\"flex h-9 items-center gap-2 rounded-xl border border-gray-200/80 bg-white/85 px-3")
    expect(source).toContain('<Icon name="globe" size="sm"')
    expect(source).toContain('{{ currentLocale?.code.toUpperCase() }}')
    expect(source).not.toContain("{{ currentLocale?.flag }}</span>")
  })

  it('renders richer locale rows inside the dropdown', () => {
    expect(source).toContain("class=\"inline-flex h-7 w-7 items-center justify-center rounded-full bg-gray-100")
    expect(source).toContain('{{ locale.name }}')
    expect(source).toContain('{{ locale.code }}')
  })
})
