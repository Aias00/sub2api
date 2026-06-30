export type PromptCatalogCopy = {
  accountAction: string
  accountActionAuthenticated: string
  accountActionAnonymous: string
  eyebrow: string
  title: string
  description: string
  caseTitle: string
  caseDescription: string
  templateTitle: string
  templateDescription: string
  total: string
  sources: string
  cases: string
  templates: string
  search: string
  searchPlaceholder: string
  caseOnly: string
  templateOnly: string
  allTypes: string
  allSources: string
  allCategories: string
  hasImage: string
  resultPrefix: string
  page: string
  previous: string
  next: string
  emptyTitle: string
  emptyDescription: string
  noImage: string
  source: string
  details: string
  prompt: string
  charUnit: string
  copyPrompt: string
  promptCopied: string
  generate: string
  importTitle: string
  importDescription: string
  importProviderX: string
  importPlaceholder: string
  importAction: string
  importing: string
  importSuccess: string
  importWarnings: string
  loadError: string
  noMoreResults: string
}

export type PromptCatalogShellConfig = {
  labels?: Partial<Record<keyof PromptCatalogCopy, string>>
  defaults?: PromptCatalogDefaults
}

export type PromptCatalogDefaults = {
  sourceType?: '' | 'case' | 'template'
  hasImage?: boolean
  pageSize?: number
  sortBy?: 'imported_at' | 'created_at' | 'updated_at' | 'title'
  sortOrder?: 'asc' | 'desc'
  generatorPath?: string
  generatorDraftSource?: string
  importXAuto?: boolean
}

export const promptCatalogCopyKeys: Array<keyof PromptCatalogCopy> = [
  'accountAction',
  'accountActionAuthenticated',
  'accountActionAnonymous',
  'eyebrow',
  'title',
  'description',
  'caseTitle',
  'caseDescription',
  'templateTitle',
  'templateDescription',
  'total',
  'sources',
  'cases',
  'templates',
  'search',
  'searchPlaceholder',
  'caseOnly',
  'templateOnly',
  'allTypes',
  'allSources',
  'allCategories',
  'hasImage',
  'resultPrefix',
  'page',
  'previous',
  'next',
  'emptyTitle',
  'emptyDescription',
  'noImage',
  'source',
  'details',
  'prompt',
  'charUnit',
  'copyPrompt',
  'promptCopied',
  'generate',
  'importTitle',
  'importDescription',
  'importProviderX',
  'importPlaceholder',
  'importAction',
  'importing',
  'importSuccess',
  'importWarnings',
  'loadError',
  'noMoreResults',
]

export function resolvePromptCatalogShellConfig(
  raw: string | undefined,
  selectedLocale: 'zh' | 'en',
): PromptCatalogShellConfig {
  if (!raw?.trim()) return {}
  try {
    const parsed = JSON.parse(raw) as Record<string, unknown>
    const scoped = selectedLocale === 'zh' ? parsed.zh : parsed.en
    const value = isRecord(scoped) ? scoped : parsed
    return {
      labels: isRecord(value.labels) ? readPromptCatalogLabels(value.labels) : undefined,
      defaults: isRecord(value.defaults) ? readPromptCatalogDefaults(value.defaults) : undefined,
    }
  } catch {
    return {}
  }
}

function readPromptCatalogLabels(value: Record<string, unknown>): PromptCatalogShellConfig['labels'] {
  const labels: PromptCatalogShellConfig['labels'] = {}
  for (const key of promptCatalogCopyKeys) {
    const label = readString(value[key])
    if (label) {
      labels[key] = label
    }
  }
  return labels
}

function readPromptCatalogDefaults(value: Record<string, unknown>): PromptCatalogDefaults | undefined {
  const defaults: PromptCatalogDefaults = {}
  const sourceType = readPromptCatalogSourceType(value.sourceType)
  const hasImage = readBoolean(value.hasImage)
  const pageSize = readPositiveInteger(value.pageSize, 100)
  const sortBy = readPromptCatalogSortBy(value.sortBy)
  const sortOrder = readPromptCatalogSortOrder(value.sortOrder)
  const generatorPath = readInternalPath(value.generatorPath)
  const generatorDraftSource = readString(value.generatorDraftSource)
  const importXAuto = readBoolean(value.importXAuto)

  if (sourceType !== undefined) defaults.sourceType = sourceType
  if (hasImage !== undefined) defaults.hasImage = hasImage
  if (pageSize !== undefined) defaults.pageSize = pageSize
  if (sortBy !== undefined) defaults.sortBy = sortBy
  if (sortOrder !== undefined) defaults.sortOrder = sortOrder
  if (generatorPath !== undefined) defaults.generatorPath = generatorPath
  if (generatorDraftSource !== undefined) defaults.generatorDraftSource = generatorDraftSource
  if (importXAuto !== undefined) defaults.importXAuto = importXAuto

  return Object.keys(defaults).length > 0 ? defaults : undefined
}

function readPromptCatalogSourceType(value: unknown): PromptCatalogDefaults['sourceType'] | undefined {
  if (value === 'case' || value === 'template') return value
  if (value === '' || value === 'all') return ''
  return undefined
}

function readPromptCatalogSortBy(value: unknown): PromptCatalogDefaults['sortBy'] | undefined {
  if (value === 'title' || value === 'created_at' || value === 'updated_at' || value === 'imported_at') {
    return value
  }
  return undefined
}

function readPromptCatalogSortOrder(value: unknown): PromptCatalogDefaults['sortOrder'] | undefined {
  if (value === 'asc' || value === 'desc') return value
  return undefined
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

function readString(value: unknown) {
  if (typeof value !== 'string') return undefined
  const trimmed = value.trim()
  return trimmed || undefined
}

function readBoolean(value: unknown): boolean | undefined {
  return typeof value === 'boolean' ? value : undefined
}

function readPositiveInteger(value: unknown, max?: number): number | undefined {
  if (typeof value !== 'number' || !Number.isInteger(value) || value < 1) {
    return undefined
  }
  if (max !== undefined && value > max) {
    return undefined
  }
  return value
}

function readInternalPath(value: unknown): string | undefined {
  const path = readString(value)
  if (!path || !path.startsWith('/') || path.startsWith('//') || path.includes('://')) {
    return undefined
  }
  if (path.includes('\n') || path.includes('\r')) {
    return undefined
  }
  return path
}
