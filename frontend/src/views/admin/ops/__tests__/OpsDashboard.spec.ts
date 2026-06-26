import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

const opsDashboardSource = readFileSync('src/views/admin/ops/OpsDashboard.vue', 'utf8')

describe('OpsDashboard', () => {
  it('uses shared auth route defaults for the settings redirect', () => {
    expect(opsDashboardSource).toContain('useAuthRouteDefaults')
    expect(opsDashboardSource).toContain('authRouteDefaults.value.adminSettingsPath')
    expect(opsDashboardSource).not.toContain("router.replace('/admin/settings')")
    expect(opsDashboardSource).not.toContain('router.replace("/admin/settings")')
  })
})
