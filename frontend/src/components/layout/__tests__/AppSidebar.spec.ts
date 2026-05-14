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
})

describe('AppSidebar regular user navigation', () => {
  it('keeps profile and api key entries in the regular user sidebar', () => {
    expect(componentSource).not.toContain(
      ".filter((item) => item.path !== '/keys' && item.path !== '/profile')"
    )
  })
})

describe('Regular user onboarding key entry point', () => {
  it('keeps regular users onboarding anchored to the sidebar key entry', () => {
    const userStepsSection = userGuideStepsSource.split('export const getUserSteps')[1] ?? ''
    expect(userStepsSection).toContain('element: \'[data-tour="sidebar-my-keys"]\'')
    expect(userStepsSection).not.toContain('element: \'[data-tour="dashboard-create-key-shortcut"]\'')
  })
})
