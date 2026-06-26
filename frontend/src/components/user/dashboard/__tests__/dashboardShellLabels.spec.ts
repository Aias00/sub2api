import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'
import { resolveDashboardShellConfig, resolveDashboardShellLabels } from '../dashboardShellLabels'

const dashboardStatsSource = readFileSync('src/components/user/dashboard/UserDashboardStats.vue', 'utf8')
const dashboardChartsSource = readFileSync('src/components/user/dashboard/UserDashboardCharts.vue', 'utf8')
const dashboardQuickActionsSource = readFileSync('src/components/user/dashboard/UserDashboardQuickActions.vue', 'utf8')
const dashboardRecentUsageSource = readFileSync('src/components/user/dashboard/UserDashboardRecentUsage.vue', 'utf8')
const dashboardShellLabelsSource = readFileSync('src/components/user/dashboard/dashboardShellLabels.ts', 'utf8')
const dashboardViewSource = readFileSync('src/views/user/DashboardView.vue', 'utf8')
const zhLocaleSource = readFileSync('src/i18n/locales/zh.ts', 'utf8')
const enLocaleSource = readFileSync('src/i18n/locales/en.ts', 'utf8')

describe('resolveDashboardShellLabels', () => {
  it('resolves configured labels without frontend fallback copy', () => {
    const labels = resolveDashboardShellLabels(
      JSON.stringify({
        en: {
          labels: {
            balance: 'Configured balance',
            platformCount: '{count} configured platforms',
            ignored: 'ignored',
          },
        },
      }),
      'en-US',
    )

    expect(labels.balance).toBe('Configured balance')
    expect(labels.platformCount).toBe('{count} configured platforms')
    expect(labels.apiKeys).toBeUndefined()
  })

  it('returns empty labels when JSON is invalid', () => {
    const labels = resolveDashboardShellLabels('{bad json', 'zh-CN')

    expect(labels).toEqual({})
  })

  it('resolves configured quick action paths from dashboard shell defaults', () => {
    const config = resolveDashboardShellConfig(
      JSON.stringify({
        zh: {
          defaults: {
            dateRangeDays: 14,
            defaultGranularity: 'hour',
            recentUsageLimit: 8,
            quickActions: {
              createApiKeyPath: '/custom/keys',
              usagePath: '/custom/usage',
              redeemPath: '/custom/redeem',
            },
          },
        },
      }),
      'zh-CN',
      {
        createApiKeyPath: '/keys',
        usagePath: '/usage',
        redeemPath: '/redeem',
      },
    )

    expect(config.defaults.quickActions).toEqual({
      createApiKeyPath: '/custom/keys',
      usagePath: '/custom/usage',
      redeemPath: '/custom/redeem',
    })
    expect(config.defaults.dateRangeDays).toBe(14)
    expect(config.defaults.defaultGranularity).toBe('hour')
    expect(config.defaults.recentUsageLimit).toBe(8)
  })

  it('falls back to built-in quick action paths when configured paths are unsafe', () => {
    const config = resolveDashboardShellConfig(
      JSON.stringify({
        en: {
          defaults: {
            dateRangeDays: 0,
            defaultGranularity: 'minute',
            recentUsageLimit: -1,
            quickActions: {
              createApiKeyPath: 'https://example.com/keys',
              usagePath: '//example.com/usage',
              redeemPath: '/redeem\\bad',
            },
          },
        },
      }),
      'en-US',
      {
        createApiKeyPath: '/keys',
        usagePath: '/usage',
        redeemPath: '/redeem',
      },
    )

    expect(config.defaults.quickActions).toEqual({
      createApiKeyPath: '/keys',
      usagePath: '/usage',
      redeemPath: '/redeem',
    })
    expect(config.defaults.dateRangeDays).toBe(7)
    expect(config.defaults.defaultGranularity).toBe('day')
    expect(config.defaults.recentUsageLimit).toBe(5)
  })

  it('uses supplied auth route defaults as quick action fallbacks', () => {
    const config = resolveDashboardShellConfig(
      JSON.stringify({
        en: {
          defaults: {
            quickActions: {
              createApiKeyPath: 'https://example.com/keys',
              usagePath: '',
              redeemPath: '//example.com/redeem',
            },
          },
        },
      }),
      'en-US',
      {
        createApiKeyPath: '/configured-keys',
        usagePath: '/configured-usage',
        redeemPath: '/configured-redeem',
      },
    )

    expect(config.defaults.quickActions).toEqual({
      createApiKeyPath: '/configured-keys',
      usagePath: '/configured-usage',
      redeemPath: '/configured-redeem',
    })
  })

  it('does not let dashboard child components render label keys as fallback copy', () => {
    expect(dashboardStatsSource).not.toContain('props.labels?.[key] || key')
    expect(dashboardChartsSource).not.toContain('props.labels?.[key] || key')
    expect(dashboardQuickActionsSource).not.toContain('props.labels?.[key] || key')
    expect(dashboardRecentUsageSource).not.toContain('props.labels?.[key] || key')
  })

  it('does not keep the legacy dashboard locale section in frontend bundles', () => {
    expect(zhLocaleSource).not.toContain('\n  dashboard: {')
    expect(enLocaleSource).not.toContain('\n  dashboard: {')
  })

  it('keeps quick action destinations supplied by dashboard shell config', () => {
    expect(dashboardShellLabelsSource).not.toContain('FALLBACK_AUTH_ROUTE_DEFAULTS')
    expect(dashboardShellLabelsSource).not.toContain("createApiKeyPath: '/keys'")
    expect(dashboardShellLabelsSource).not.toContain("usagePath: '/usage'")
    expect(dashboardShellLabelsSource).not.toContain("redeemPath: '/redeem'")
    expect(dashboardQuickActionsSource).toContain('actionDefaults.createApiKeyPath')
    expect(dashboardQuickActionsSource).toContain('actionDefaults.usagePath')
    expect(dashboardQuickActionsSource).toContain('actionDefaults.redeemPath')
    expect(dashboardViewSource).toContain('useAuthRouteDefaults')
    expect(dashboardViewSource).toContain('createApiKeyPath: authRouteDefaults.value.apiKeysPath')
    expect(dashboardViewSource).toContain('usagePath: authRouteDefaults.value.usagePath')
    expect(dashboardViewSource).toContain('redeemPath: authRouteDefaults.value.redeemPath')
    expect(dashboardViewSource).toContain('dashboardShell.value.defaults.dateRangeDays')
    expect(dashboardViewSource).toContain('dashboardShell.value.defaults.defaultGranularity')
    expect(dashboardViewSource).toContain('dashboardShell.value.defaults.recentUsageLimit')
    expect(dashboardViewSource).not.toContain('6 * 86400000')
    expect(dashboardViewSource).not.toContain("granularity = ref('day')")
    expect(dashboardViewSource).not.toContain('res.items.slice(0, 5)')
    expect(dashboardQuickActionsSource).not.toContain("router.push('/keys')")
    expect(dashboardQuickActionsSource).not.toContain("router.push('/usage')")
    expect(dashboardQuickActionsSource).not.toContain("router.push('/redeem')")
  })

  it('keeps the recent usage detail link supplied by dashboard shell config', () => {
    expect(dashboardRecentUsageSource).toContain(':to="usagePath"')
    expect(dashboardRecentUsageSource).toContain('usagePath: string')
    expect(dashboardRecentUsageSource).not.toContain('to="/usage"')
  })

  it('does not hard-code dollar-prefixed dashboard cost output', () => {
    expect(dashboardStatsSource).not.toContain('${{')
    expect(dashboardChartsSource).not.toContain('${{')
    expect(dashboardRecentUsageSource).not.toContain('${{')
    expect(dashboardStatsSource).toContain('formatPublicMoneyAmount')
    expect(dashboardChartsSource).toContain('formatPublicMoneyAmount')
    expect(dashboardRecentUsageSource).toContain('formatPublicMoneyAmount')
  })
})
