import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

const notFoundViewSource = readFileSync('src/views/NotFoundView.vue', 'utf8')

describe('NotFoundView', () => {
  it('uses auth route defaults for the dashboard action instead of a local route fallback', () => {
    expect(notFoundViewSource).toContain('useAuthRouteDefaults')
    expect(notFoundViewSource).toContain(':to="dashboardPath"')
    expect(notFoundViewSource).not.toContain('to="/dashboard"')
    expect(notFoundViewSource).not.toContain("'/admin/dashboard'")
  })
})
