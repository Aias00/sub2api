export function normalizeDocsNamespacePart(value: string) {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
}

export function buildDocsSearchNamespace(parts: string[]) {
  return parts.map(normalizeDocsNamespacePart).filter(Boolean).join('-')
}

export function withDocsContentVersion(hash: string, version: string, queryKey = '_docs_v') {
  const normalizedHash = hash.startsWith('#') ? hash : `#${hash}`
  if (!normalizedHash.startsWith('#/')) {
    return normalizedHash
  }

  const [path, query = ''] = normalizedHash.slice(1).split('?')
  const params = new URLSearchParams(query)
  if (params.get(queryKey) !== version) {
    params.set(queryKey, version)
  }
  const queryString = params.toString()

  return queryString ? `#${path}?${queryString}` : `#${path}`
}

export function getDocsHashPath(hash: string) {
  const normalizedHash = hash.startsWith('#') ? hash : `#${hash}`
  const path = normalizedHash.slice(1).split('?')[0] || '/'

  return path === '/' ? '/' : path.replace(/\/+$/, '')
}

export function resolveInitialDocsHash(routePath: string, currentHash: string, fallbackHash: string) {
  if (routePath === '/docs' && currentHash.startsWith('#/')) {
    return currentHash
  }

  return fallbackHash
}
