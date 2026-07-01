import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

const dashboardSource = readFileSync(resolve(process.cwd(), 'src/views/user/DashboardView.vue'), 'utf8')

describe('User dashboard overview layout', () => {
  it('does not render the redundant overview hero for normal users', () => {
    expect(dashboardSource).not.toContain('class="page-hero"')
    expect(dashboardSource).not.toContain('page-hero-grid')
    expect(dashboardSource).not.toContain("{{ t('dashboard.welcomeMessage') }}")
    expect(dashboardSource).not.toContain('class="metric-panel"')
  })

  it('keeps the main dashboard cards and charts as the first loaded content', () => {
    expect(dashboardSource).toContain('<AppLayout content-container-class="max-w-6xl" header-container-class="max-w-6xl">')
    expect(dashboardSource).toContain('class="space-y-8"')
    expect(dashboardSource).not.toContain('max-w-[1680px] space-y-8')
    expect(dashboardSource).not.toContain('max-w-7xl space-y-8')
    expect(dashboardSource).toContain('<UserDashboardStats :stats="stats"')
    expect(dashboardSource).toContain('<UserDashboardBusinessOverview')
    expect(dashboardSource).toContain('<UserDashboardCharts')
    expect(dashboardSource).toContain('<UserDashboardRecentUsage')
    expect(dashboardSource).toContain('<UserDashboardQuickActions')
    expect(dashboardSource).toContain(':labels="dashboardLabels"')
    expect(dashboardSource).toContain(':usage-path="dashboardShell.defaults.quickActions.usagePath"')
    expect(dashboardSource).toContain(':action-defaults="dashboardShell.defaults.quickActions"')
    expect(dashboardSource).toContain("UserDashboardBusinessOverview from '@/components/user/dashboard/UserDashboardBusinessOverview.vue'")
    expect(dashboardSource).toContain("from './dashboardRuntime'")
    expect(dashboardSource).toContain('resolveDashboardStartDate')
    expect(dashboardSource).toContain('resolveDashboardEndDate')
    expect(dashboardSource).toContain('selectDashboardRecentUsage')
  })
})
