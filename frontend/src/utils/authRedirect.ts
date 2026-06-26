import { FALLBACK_AUTH_ROUTE_DEFAULTS } from '@/router/setupRedirect'

export const DEFAULT_AUTH_REDIRECT_PATH = FALLBACK_AUTH_ROUTE_DEFAULTS.userRedirectPath
export const DEFAULT_AUTH_BIND_REDIRECT_PATH = FALLBACK_AUTH_ROUTE_DEFAULTS.profilePath

export function sanitizeAuthRedirectPath(
  path?: string | null,
  fallback = DEFAULT_AUTH_REDIRECT_PATH,
): string {
  const trimmed = String(path || '').trim()
  if (!trimmed) return fallback
  if (!trimmed.startsWith('/')) return fallback
  if (trimmed.startsWith('//')) return fallback
  if (trimmed.includes('://')) return fallback
  if (trimmed.includes('\n') || trimmed.includes('\r')) return fallback
  return trimmed
}

export function resolveRouteAuthRedirect(
  redirect: unknown,
  fallback = DEFAULT_AUTH_REDIRECT_PATH,
): string {
  const value = Array.isArray(redirect) ? redirect[0] : redirect
  return sanitizeAuthRedirectPath(typeof value === 'string' ? value : '', fallback)
}

export function resolveAuthBindRedirect(redirect?: string | null): string {
  return sanitizeAuthRedirectPath(redirect, DEFAULT_AUTH_BIND_REDIRECT_PATH)
}
