import type { PromptCatalogFacet } from '@/api/prompts'
import type { PromptCatalogCopy, PromptCatalogDefaults } from '@/utils/promptCatalogShell'

export type PromptCatalogFilters = {
  search: string
  sourceType: string
  sourceProject: string
  category: string
  hasImage: boolean
}

export function applyPromptCatalogDefaults(
  filters: PromptCatalogFilters,
  defaults: PromptCatalogDefaults | undefined,
) {
  filters.sourceType = defaults?.sourceType || ''
  filters.hasImage = Boolean(defaults?.hasImage)
}

export function resolvePromptCatalogPageTitle(copy: PromptCatalogCopy, sourceType: string): string {
  if (sourceType === 'template') {
    return copy.templateTitle || copy.title
  }
  return copy.caseTitle || copy.title
}

export function resolvePromptCatalogPageDescription(copy: PromptCatalogCopy, sourceType: string): string {
  if (sourceType === 'template') {
    return copy.templateDescription || copy.description
  }
  return copy.caseDescription || copy.description
}

export function resolvePromptCatalogPageSize(defaults: PromptCatalogDefaults | undefined): number | undefined {
  return defaults?.pageSize
}

export function resolvePromptCatalogSortBy(defaults: PromptCatalogDefaults | undefined): string | undefined {
  return defaults?.sortBy
}

export function resolvePromptCatalogSortOrder(defaults: PromptCatalogDefaults | undefined): string | undefined {
  return defaults?.sortOrder
}

export function resolvePromptCatalogGeneratorPath(defaults: PromptCatalogDefaults | undefined): string | undefined {
  return defaults?.generatorPath
}

export function resolvePromptCatalogGeneratorDraftSource(
  defaults: PromptCatalogDefaults | undefined,
): string | undefined {
  return defaults?.generatorDraftSource
}

export function resolvePromptCatalogImportXAuto(defaults: PromptCatalogDefaults | undefined): boolean {
  return defaults?.importXAuto ?? true
}

export function buildPromptCatalogListParams(filters: PromptCatalogFilters, defaults: PromptCatalogDefaults | undefined, page: number) {
  return {
    page,
    page_size: resolvePromptCatalogPageSize(defaults),
    source_type: filters.sourceType || undefined,
    source_project: filters.sourceProject || undefined,
    category: filters.category || undefined,
    search: filters.search || undefined,
    has_image: filters.hasImage ? true : undefined,
    sort_by: resolvePromptCatalogSortBy(defaults),
    sort_order: resolvePromptCatalogSortOrder(defaults),
  }
}

export function formatPromptCatalogDate(value: string | null | undefined, locale: string): string {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  return new Intl.DateTimeFormat(locale, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  }).format(date)
}

export function resolvePromptCatalogFacetLabel(facet: PromptCatalogFacet): string {
  return facet.display_label
}

export function buildPromptCatalogImportSuccessMessage(label: string, title: string): string {
  return `${label}: ${title}`
}
