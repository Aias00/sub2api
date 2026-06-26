import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import {
  DEFAULT_AUTH_BIND_REDIRECT_PATH,
  DEFAULT_AUTH_REDIRECT_PATH,
  resolveAuthBindRedirect,
  resolveRouteAuthRedirect,
  sanitizeAuthRedirectPath,
} from '../authRedirect'

const authRedirectSource = readFileSync('src/utils/authRedirect.ts', 'utf8')

describe('auth redirect helpers', () => {
  it('centralizes the default auth redirect', () => {
    expect(resolveRouteAuthRedirect('/billing?plan=pro')).toBe('/billing?plan=pro')
    expect(resolveRouteAuthRedirect(['//evil.example'])).toBe(DEFAULT_AUTH_REDIRECT_PATH)
    expect(resolveRouteAuthRedirect('')).toBe(DEFAULT_AUTH_REDIRECT_PATH)
  })

  it('rejects external or malformed redirect paths', () => {
    expect(sanitizeAuthRedirectPath('https://evil.example/path')).toBe(DEFAULT_AUTH_REDIRECT_PATH)
    expect(sanitizeAuthRedirectPath('//evil.example/path')).toBe(DEFAULT_AUTH_REDIRECT_PATH)
    expect(sanitizeAuthRedirectPath('dashboard')).toBe(DEFAULT_AUTH_REDIRECT_PATH)
    expect(sanitizeAuthRedirectPath('/safe\nLocation')).toBe(DEFAULT_AUTH_REDIRECT_PATH)
  })

  it('centralizes the default auth bind redirect', () => {
    expect(resolveAuthBindRedirect('/profile/security')).toBe('/profile/security')
    expect(resolveAuthBindRedirect('')).toBe(DEFAULT_AUTH_BIND_REDIRECT_PATH)
    expect(resolveAuthBindRedirect('https://evil.example/profile')).toBe(DEFAULT_AUTH_BIND_REDIRECT_PATH)
  })

  it('reuses shared auth route defaults instead of local path literals', () => {
    expect(authRedirectSource).toContain('FALLBACK_AUTH_ROUTE_DEFAULTS.userRedirectPath')
    expect(authRedirectSource).toContain('FALLBACK_AUTH_ROUTE_DEFAULTS.profilePath')
    expect(authRedirectSource).not.toContain("DEFAULT_AUTH_REDIRECT_PATH = '/dashboard'")
    expect(authRedirectSource).not.toContain("DEFAULT_AUTH_BIND_REDIRECT_PATH = '/profile'")
  })
})
