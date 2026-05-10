import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

const dashboardSource = readFileSync(resolve(process.cwd(), 'src/views/user/DashboardView.vue'), 'utf8')

describe('Dashboard hero typography', () => {
  it('uses a smaller welcome heading size than the previous oversized variant', () => {
    expect(dashboardSource).toContain(
      'class="max-w-3xl text-2xl font-semibold tracking-tight text-gray-950 dark:text-white md:text-3xl"'
    )
    expect(dashboardSource).not.toContain(
      'class="max-w-3xl text-3xl font-semibold tracking-tight text-gray-950 dark:text-white md:text-4xl"'
    )
  })

  it('keeps the hero copy focused on the welcome headline', () => {
    expect(dashboardSource).not.toContain('<span class="page-kicker">{{ appStore.siteName }}</span>')
    expect(dashboardSource).not.toContain("{{ t('dashboard.title') }}")
    expect(dashboardSource).not.toContain('{{ user?.email || appStore.siteName }}')
  })
})
