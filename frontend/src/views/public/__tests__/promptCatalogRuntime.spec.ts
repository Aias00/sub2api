import { describe, expect, it } from 'vitest'

import {
  applyPromptCatalogDefaults,
  buildPromptCatalogImportSuccessMessage,
  buildPromptCatalogListParams,
  formatPromptCatalogDate,
  resolvePromptCatalogFacetLabel,
  resolvePromptCatalogGeneratorDraftSource,
  resolvePromptCatalogGeneratorPath,
  resolvePromptCatalogImportXAuto,
  resolvePromptCatalogPageDescription,
  resolvePromptCatalogPageTitle,
  type PromptCatalogFilters,
} from '../promptCatalogRuntime'

describe('promptCatalogRuntime', () => {
  it('resolves page title and description from source type', () => {
    const copy = {
      title: 'All',
      description: 'All desc',
      caseTitle: 'Cases',
      caseDescription: 'Cases desc',
      templateTitle: 'Templates',
      templateDescription: 'Templates desc',
    } as any

    expect(resolvePromptCatalogPageTitle(copy, 'case')).toBe('Cases')
    expect(resolvePromptCatalogPageTitle(copy, 'template')).toBe('Templates')
    expect(resolvePromptCatalogPageDescription(copy, 'case')).toBe('Cases desc')
    expect(resolvePromptCatalogPageDescription(copy, 'template')).toBe('Templates desc')
  })

  it('applies configured defaults and builds list params', () => {
    const filters: PromptCatalogFilters = {
      search: 'abc',
      sourceType: '',
      sourceProject: 'x',
      category: 'portrait',
      hasImage: false,
    }

    applyPromptCatalogDefaults(filters, {
      sourceType: 'template',
      hasImage: true,
      pageSize: 12,
      sortBy: 'title',
      sortOrder: 'asc',
    })

    expect(filters.sourceType).toBe('template')
    expect(filters.hasImage).toBe(true)

    expect(buildPromptCatalogListParams(filters, {
      pageSize: 12,
      sortBy: 'title',
      sortOrder: 'asc',
    }, 2)).toEqual({
      page: 2,
      page_size: 12,
      source_type: 'template',
      source_project: 'x',
      category: 'portrait',
      search: 'abc',
      has_image: true,
      sort_by: 'title',
      sort_order: 'asc',
    })
  })

  it('resolves generator/import defaults and formatting helpers', () => {
    expect(resolvePromptCatalogGeneratorPath({ generatorPath: '/image-generator' })).toBe('/image-generator')
    expect(resolvePromptCatalogGeneratorDraftSource({ generatorDraftSource: 'catalog' })).toBe('catalog')
    expect(resolvePromptCatalogImportXAuto(undefined)).toBe(true)
    expect(resolvePromptCatalogImportXAuto({ importXAuto: false })).toBe(false)
    expect(resolvePromptCatalogFacetLabel({ display_label: 'API Source' } as any)).toBe('API Source')
    expect(buildPromptCatalogImportSuccessMessage('Imported', 'Case A')).toBe('Imported: Case A')
    expect(formatPromptCatalogDate('2026-06-18T00:00:00Z', 'en')).toContain('2026')
    expect(formatPromptCatalogDate('', 'en')).toBe('')
  })
})
