import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

const source = readFileSync(resolve(process.cwd(), 'src/components/common/LocaleSwitcher.vue'), 'utf8')

describe('LocaleSwitcher visual treatment', () => {
  it('uses a text trigger with globe icon and current language name', () => {
    expect(source).toContain('class="group inline-flex h-10 items-center gap-2 rounded-full')
    expect(source).toContain('max-w-28 truncate text-left text-sm font-semibold leading-none')
    expect(source).toContain('<Icon name="globe" size="md"')
    expect(source).toContain('{{ currentLocale?.name }}')
    expect(source).not.toContain('{{ currentLocale?.code.toUpperCase() }}')
    expect(source).not.toContain("{{ currentLocale?.flag }}</span>")
  })

  it('renders a homepage-aligned rounded language menu with text-only rows', () => {
    expect(source).toContain('w-40 overflow-hidden rounded-2xl border border-slate-200/80 bg-white/95 p-2')
    expect(source).toContain('dark:border-white/10 dark:bg-slate-900/95')
    expect(source).toContain('rounded-xl px-4 py-2.5 text-left text-sm')
    expect(source).toContain("'bg-slate-100 text-slate-950 dark:bg-white/10 dark:text-white'")
    expect(source).toContain('{{ locale.name }}')
    expect(source).not.toContain('{{ locale.code }}')
    expect(source).not.toContain('bg-[#343434]')
  })
})
