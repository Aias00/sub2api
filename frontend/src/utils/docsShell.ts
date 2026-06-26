import { resolveLocalizedShellLabels } from './localizedShell'

export type DocsShellCopy = {
  title: string
  dashboard: string
  login: string
  searchPlaceholder: string
  noData: string
}

export type DocsShellConfig = {
  labels: DocsShellCopy
  defaults: {
    appRouteLinks: string[]
  }
}

const docsShellCopyKeys = [
  'title',
  'dashboard',
  'login',
  'searchPlaceholder',
  'noData',
] as const satisfies readonly (keyof DocsShellCopy)[]

export function resolveDocsShellCopy(raw: string | undefined, runtimeLocale: string): DocsShellCopy {
  return resolveDocsShellConfig(raw, runtimeLocale).labels
}

export function resolveDocsShellConfig(raw: string | undefined, runtimeLocale: string): DocsShellConfig {
  return {
    labels: resolveLocalizedShellLabels(raw, runtimeLocale, docsShellCopyKeys),
    defaults: {
      appRouteLinks: readAppRouteLinks(raw, runtimeLocale),
    },
  }
}

function readAppRouteLinks(raw: string | undefined, runtimeLocale: string): string[] {
  const config = pickDocsConfig(raw, runtimeLocale)
  if (!config || !isRecord(config.defaults) || !Array.isArray(config.defaults.appRouteLinks)) {
    return []
  }

  return [...new Set(config.defaults.appRouteLinks.map(normalizeAppRouteLink).filter(Boolean))]
}

function normalizeAppRouteLink(value: unknown): string {
  if (typeof value !== 'string') return ''
  const trimmed = value.trim()
  if (!trimmed || trimmed.includes('\n') || trimmed.includes('\r')) return ''
  if (trimmed.startsWith('http://') || trimmed.startsWith('https://') || trimmed.startsWith('//')) return ''
  const path = trimmed.startsWith('#/') ? trimmed.slice(1) : trimmed
  if (!path.startsWith('/') || path.startsWith('//')) return ''
  const normalized = path.split('?')[0]?.replace(/\/+$/, '') || '/'
  return normalized === '/' ? '#/' : `#${normalized}`
}

function pickDocsConfig(raw: string | undefined, runtimeLocale: string): Record<string, unknown> | null {
  if (!raw?.trim()) return null
  try {
    const parsed = JSON.parse(raw) as unknown
    if (!isRecord(parsed)) return null

    const normalizedLocale = runtimeLocale.toLowerCase()
    const baseLocale = normalizedLocale.split('-')[0]
    const localeKeys = [normalizedLocale, baseLocale, 'en', 'zh']

    for (const key of localeKeys) {
      const localized = parsed[key]
      if (isRecord(localized)) return localized
    }

    return parsed
  } catch {
    return null
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}
