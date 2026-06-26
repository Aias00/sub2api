import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

const setupWizardSource = readFileSync('src/views/setup/SetupWizardView.vue', 'utf8')

describe('SetupWizardView', () => {
  it('uses shared auth route defaults for the post-install login redirect', () => {
    expect(setupWizardSource).toContain('resolveAuthRouteDefaultsFromShellDefaults().loginPath')
    expect(setupWizardSource).toContain('window.location.href = setupLoginPath')
    expect(setupWizardSource).not.toContain("window.location.href = '/login'")
    expect(setupWizardSource).not.toContain('window.location.href = "/login"')
  })
})
