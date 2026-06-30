import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

const source = readFileSync(resolve(process.cwd(), 'src/components/layout/AuthLayout.vue'), 'utf8')

describe('AuthLayout brand block', () => {
  it('shows the site name without the subtitle line', () => {
    expect(source).toContain('{{ siteName }}')
    expect(source).not.toContain('{{ siteSubtitle }}')
    expect(source).not.toContain('Subscription to API Conversion Platform')
  })

  it('reads footer copy from auth shell public settings', () => {
    expect(source).toContain("authText('allRightsReserved')")
    expect(source).toContain('useAuthShellText')
    expect(source).toContain('loadAuthShellConfig')
    expect(source).not.toContain('function authText(key: string')
    expect(source).not.toContain('All rights reserved.')
  })

  it('links the brand block back to the configured home path', () => {
    expect(source).toContain('useAuthRouteDefaults')
    expect(source).toContain('const { authRouteDefaults } = useAuthRouteDefaults()')
    expect(source).toContain(':to="authRouteDefaults.homePath"')
    expect(source).not.toContain('to="/home"')
  })
})
