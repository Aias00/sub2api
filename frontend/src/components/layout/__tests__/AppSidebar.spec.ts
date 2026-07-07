import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../AppSidebar.vue')
const componentSource = readFileSync(componentPath, 'utf8')
const userGuideStepsPath = resolve(dirname(fileURLToPath(import.meta.url)), '../../../components/Guide/steps.ts')
const userGuideStepsSource = readFileSync(userGuideStepsPath, 'utf8')

describe('AppSidebar custom SVG styles', () => {
  it('does not override uploaded SVG fill or stroke colors', () => {
    expect(componentSource).toContain('.sidebar-svg-icon {')
    expect(componentSource).toContain('color: currentColor;')
    expect(componentSource).toContain('display: block;')
    expect(componentSource).not.toContain('stroke: currentColor;')
    expect(componentSource).not.toContain('fill: none;')
  })
})

describe('AppSidebar header styles', () => {
  it('does not mount the version badge or trigger upstream update checks', () => {
    expect(componentSource).not.toContain('VersionBadge')
    expect(componentSource).not.toContain('siteVersion')
  })

  it('links the sidebar brand back to the configured home path', () => {
    expect(componentSource).toContain(':to="authRouteDefaults.homePath"')
    expect(componentSource).toContain('class="sidebar-home-link"')
    expect(componentSource).toContain('@click="closeMobile"')
  })
})

describe('AppSidebar regular user navigation', () => {
  it('keeps profile and api key entries in the regular user sidebar', () => {
    expect(componentSource).not.toContain(
      ".filter((item) => item.path !== '/keys' && item.path !== '/profile')"
    )
  })

  it('reads regular user navigation paths from auth route defaults', () => {
    expect(componentSource).toContain('useAuthRouteDefaults')
    expect(componentSource).toContain('const navPaths = authRouteDefaults.value')
    expect(componentSource).toContain('path: navPaths.userRedirectPath')
    expect(componentSource).toContain('path: navPaths.apiKeysPath')
    expect(componentSource).toContain('path: navPaths.usagePath')
    expect(componentSource).toContain('path: navPaths.purchasePath')
    expect(componentSource).toContain('path: navPaths.ordersPath')
    expect(componentSource).toContain('path: navPaths.profilePath')
    expect(componentSource).toContain(':data-tour="item.path === authRouteDefaults.apiKeysPath ?')
    expect(componentSource).not.toMatch(/path: '\/(?:dashboard|keys|usage|available-channels|available-groups|subscriptions|purchase|orders|redeem|affiliate|profile)'/)
    expect(componentSource).not.toContain("item.path === '/keys'")
  })

  it('reads primary admin entry paths from auth route defaults', () => {
    expect(componentSource).toContain('path: navPaths.adminRedirectPath')
    expect(componentSource).toContain('path: `${navPaths.adminSettingsPath}?tab=runtime`')
    expect(componentSource).toContain('path: navPaths.adminSettingsPath')
    expect(componentSource).toContain('function splitNavPath')
    expect(componentSource).toContain('Object.entries(target.query).every')
    expect(componentSource).not.toContain("path: '/admin/dashboard'")
    expect(componentSource).not.toContain("path: '/admin/runtime-settings'")
    expect(componentSource).not.toContain("path: '/admin/settings'")
  })

  it('delegates sidebar section assembly to shared sidebar runtime helpers', () => {
    expect(componentSource).toContain("from './sidebarRuntime'")
    expect(componentSource).toContain('buildSidebarVisibleItemMap')
    expect(componentSource).toContain('buildSidebarSections')
    expect(componentSource).not.toContain('function applyFeatureFlags(')
  })

  it('exposes public creation and content tools in the dashboard sidebar', () => {
    expect(componentSource).not.toContain('const businessNavItems = computed')
    expect(componentSource).not.toContain("t('nav.businessCapabilities')")
    expect(componentSource).toContain("tasks: { path: '/app/tasks', label: t('nav.myTasks')")
    expect(componentSource).toContain("promptCatalog: { path: '/app/prompts', label: t('nav.promptCatalog')")
    expect(componentSource).toContain("imageGenerator: { path: '/app/image-generator', label: t('nav.imageGenerator')")
    expect(componentSource).toContain("wechatExport: { path: '/app/wechat', label: t('nav.wechatExport')")
    expect(componentSource).toContain("hotTopics: { path: '/app/hot', label: t('nav.hotTopics')")
    expect(componentSource).not.toContain("promptCatalog: { path: '/prompts', label: t('nav.promptCatalog')")
  })

  it('keeps async tasks and tool entries in both self sidebar defaults', () => {
    expect(componentSource).toContain("tasks: { path: '/app/tasks', label: t('nav.myTasks')")
    expect(componentSource).toContain("'tasks',")
    expect(componentSource).toContain("'dashboard',\n        'tasks',\n        'promptCatalog',\n        'imageGenerator',\n        'wechatExport',\n        'hotTopics',")
    expect(componentSource).toContain("'admin-personal',\n      items: [\n        'tasks',\n        'promptCatalog',\n        'imageGenerator',\n        'wechatExport',\n        'hotTopics',")
  })
})

describe('Regular user onboarding key entry point', () => {
  it('keeps regular users onboarding anchored to the sidebar key entry', () => {
    const userStepsSection = userGuideStepsSource.split('export const getUserSteps')[1] ?? ''
    expect(userStepsSection).toContain('element: \'[data-tour="sidebar-my-keys"]\'')
    expect(userStepsSection).not.toContain('element: \'[data-tour="dashboard-create-key-shortcut"]\'')
  })
})
