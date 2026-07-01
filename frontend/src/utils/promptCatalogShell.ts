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

const defaultPromptCatalogCopyByLocale: Record<'zh' | 'en', PromptCatalogCopy> = {
  zh: {
    accountAction: '',
    accountActionAuthenticated: '仪表盘',
    accountActionAnonymous: '登录',
    eyebrow: '提示词案例',
    title: '图片提示词案例库',
    description: '浏览可复用的图片提示词案例，复制提示词或直接带入生图工作台。',
    caseTitle: '图片提示词案例库',
    caseDescription: '从真实案例中查找图片提示词、参考图和标签。',
    templateTitle: '提示词模板',
    templateDescription: '沉淀可复用的提示词模板。',
    total: '总数',
    sources: '来源',
    cases: '案例',
    templates: '模板',
    search: '搜索',
    searchPlaceholder: '搜索标题、提示词或标签',
    caseOnly: '仅案例',
    templateOnly: '仅模板',
    allTypes: '全部类型',
    allSources: '全部来源',
    allCategories: '全部分类',
    hasImage: '只看有图',
    resultPrefix: '结果',
    page: '页',
    previous: '上一页',
    next: '下一页',
    emptyTitle: '暂无提示词案例',
    emptyDescription: '当前筛选条件下没有可展示的提示词案例。',
    noImage: '暂无图片',
    source: '来源',
    details: '详情',
    prompt: '提示词',
    charUnit: '字符',
    copyPrompt: '复制提示词',
    promptCopied: '提示词已复制',
    generate: '去生图',
    importTitle: '导入提示词',
    importDescription: '从 X/Twitter 链接导入图片提示词案例。',
    importProviderX: 'Twitter 导入',
    importPlaceholder: '粘贴 X/Twitter 帖子链接',
    importAction: '导入',
    importing: '导入中',
    importSuccess: '导入成功',
    importWarnings: '导入提示',
    loadError: '加载提示词案例失败',
    noMoreResults: '没有更多结果',
  },
  en: {
    accountAction: '',
    accountActionAuthenticated: 'Dashboard',
    accountActionAnonymous: 'Log in',
    eyebrow: 'Prompt Cases',
    title: 'Image Prompt Catalog',
    description: 'Browse reusable image prompt cases, copy prompts, or send them into the image workspace.',
    caseTitle: 'Image Prompt Catalog',
    caseDescription: 'Find image prompts, references, and tags from real cases.',
    templateTitle: 'Prompt Templates',
    templateDescription: 'Reusable prompt templates for image workflows.',
    total: 'Total',
    sources: 'Sources',
    cases: 'Cases',
    templates: 'Templates',
    search: 'Search',
    searchPlaceholder: 'Search titles, prompts, or tags',
    caseOnly: 'Cases only',
    templateOnly: 'Templates only',
    allTypes: 'All types',
    allSources: 'All sources',
    allCategories: 'All categories',
    hasImage: 'With image only',
    resultPrefix: 'Results',
    page: 'Page',
    previous: 'Previous',
    next: 'Next',
    emptyTitle: 'No prompt cases',
    emptyDescription: 'No prompt cases match the current filters.',
    noImage: 'No image',
    source: 'Source',
    details: 'Details',
    prompt: 'Prompt',
    charUnit: 'chars',
    copyPrompt: 'Copy prompt',
    promptCopied: 'Prompt copied',
    generate: 'Generate',
    importTitle: 'Import prompt',
    importDescription: 'Import an image prompt case from an X/Twitter link.',
    importProviderX: 'Twitter import',
    importPlaceholder: 'Paste an X/Twitter post URL',
    importAction: 'Import',
    importing: 'Importing',
    importSuccess: 'Imported',
    importWarnings: 'Import warnings',
    loadError: 'Failed to load prompt cases',
    noMoreResults: 'No more results',
  },
}

function defaultPromptCatalogCopy(selectedLocale: 'zh' | 'en'): PromptCatalogCopy {
  return { ...defaultPromptCatalogCopyByLocale[selectedLocale] }
}

export function resolvePromptCatalogShellConfig(
  raw: string | undefined,
  selectedLocale: 'zh' | 'en',
): PromptCatalogShellConfig {
  const defaultLabels = defaultPromptCatalogCopy(selectedLocale)
  if (!raw?.trim()) return { labels: defaultLabels }
  try {
    const parsed = JSON.parse(raw) as Record<string, unknown>
    const scoped = selectedLocale === 'zh' ? parsed.zh : parsed.en
    const value = isRecord(scoped) ? scoped : parsed
    const labels = isRecord(value.labels) ? readPromptCatalogLabels(value.labels) : undefined
    return {
      labels: { ...defaultLabels, ...(labels || {}) },
      defaults: isRecord(value.defaults) ? readPromptCatalogDefaults(value.defaults) : undefined,
    }
  } catch {
    return { labels: defaultLabels }
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
