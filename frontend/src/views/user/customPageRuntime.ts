type CustomMenuLike = {
  id: string
  url: string
  page_slug?: string
}

export function resolveCustomPageMenuItem(
  id: string,
  publicItems: CustomMenuLike[],
  adminItems: CustomMenuLike[],
  isAdmin: boolean,
) {
  const found = publicItems.find((item) => item.id === id) ?? null
  if (found) return found
  if (isAdmin) {
    return adminItems.find((item) => item.id === id) ?? null
  }
  return null
}

export function resolveCustomPageMarkdownSlug(
  item: CustomMenuLike | null,
): string {
  if (!item) return ''
  if (item.page_slug) return item.page_slug
  if (item.url?.startsWith('md:')) return item.url.slice(3)
  return ''
}

export function isRelativeCustomPageMarkdownAsset(src: string): boolean {
  const trimmed = src.trim()
  if (!trimmed || /^[a-z][a-z0-9+.-]*:/i.test(trimmed) || trimmed.startsWith('//') || trimmed.startsWith('/')) {
    return false
  }
  const [pathPart] = trimmed.split(/([?#].*)/, 2)
  return pathPart
    .split('/')
    .filter((part) => part && part !== '.')
    .every((part) => part !== '..' && !part.includes('\\'))
}

export function buildCustomPageImageUrl(slug: string, src: string): string {
  const trimmed = src.trim()
  const [pathPart, suffix = ''] = trimmed.split(/([?#].*)/, 2)
  const encodedPath = pathPart
    .split('/')
    .filter((part) => part && part !== '.')
    .map((part) => encodeURIComponent(part))
    .join('/')
  return `/api/v1/pages/${encodeURIComponent(slug)}/images/${encodedPath}${suffix}`
}
