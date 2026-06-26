const docsContentLocales = ['zh', 'en'] as const

export type DocsContentLocale = typeof docsContentLocales[number]

export function resolveDocsContentBasePath(raw: string | undefined, locale: DocsContentLocale): string {
  const value = raw?.trim()
  if (!value) return ''

  try {
    const parsed = JSON.parse(value) as unknown
    if (typeof parsed === 'string') {
      return normalizeDocsBasePath(parsed)
    }
    if (isRecord(parsed)) {
      const scoped = parsed[locale] ?? parsed.default
      return normalizeDocsBasePath(typeof scoped === 'string' ? scoped : '')
    }
  } catch {
    return normalizeDocsBasePath(value)
  }

  return ''
}

function normalizeDocsBasePath(raw: string): string {
  const value = raw.trim()
  if (!value) return ''
  if (!isAllowedDocsBasePath(value)) return ''
  const [withoutHash] = value.split('#')
  const [path, query] = withoutHash.split('?')
  const normalizedPath = path.endsWith('/') ? path : `${path}/`
  return query ? `${normalizedPath}?${query}` : normalizedPath
}

function isAllowedDocsBasePath(value: string): boolean {
  return value.startsWith('/') || value.startsWith('https://') || value.startsWith('http://')
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}
