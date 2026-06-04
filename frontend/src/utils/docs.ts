export interface DocsLinkTarget {
  internal: boolean
  to: string
  href: string
}

const INTERNAL_DOCS_PATH = '/docs'
const DOCS_SECTION_INDEX_PATHS = new Set([
  'quickstart',
])

export function resolveDocsLink(docUrl: string, currentOrigin: string): DocsLinkTarget {
  const trimmed = docUrl.trim()
  if (!trimmed) {
    return {
      internal: true,
      to: INTERNAL_DOCS_PATH,
      href: INTERNAL_DOCS_PATH,
    }
  }

  try {
    const baseOrigin = currentOrigin || 'https://local.invalid'
    const parsed = new URL(trimmed, baseOrigin)
    const normalizedPath = parsed.pathname.replace(/\/+$/g, '') || '/'
    const normalizedOrigin = new URL(baseOrigin).origin
    const sameOrigin = parsed.origin === normalizedOrigin
    const shouldUseInternalDocs =
      sameOrigin &&
      (normalizedPath === '/' ||
        normalizedPath === '/home' ||
        normalizedPath === '/docs' ||
        normalizedPath.startsWith('/docs/'))

    if (shouldUseInternalDocs) {
      return {
        internal: true,
        to: INTERNAL_DOCS_PATH,
        href: INTERNAL_DOCS_PATH,
      }
    }

    return {
      internal: false,
      to: parsed.toString(),
      href: parsed.toString(),
    }
  } catch {
    return {
      internal: true,
      to: INTERNAL_DOCS_PATH,
      href: INTERNAL_DOCS_PATH,
    }
  }
}

export function normalizeDocsHashPath(pathMatch: string | string[] | undefined): string {
  const normalized = Array.isArray(pathMatch)
    ? pathMatch.join('/')
    : typeof pathMatch === 'string'
      ? pathMatch
      : ''

  const trimmed = normalized.trim().replace(/^\/+|\/+$/g, '')
  if (!trimmed) return '#/'

  const normalizedIndexPath = DOCS_SECTION_INDEX_PATHS.has(trimmed)
    ? `${trimmed}/README`
    : trimmed

  return `#/${normalizedIndexPath}`
}
