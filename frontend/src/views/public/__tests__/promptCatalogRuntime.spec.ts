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
  resolvePromptCatalogValueLabel,
  type PromptCatalogFilters,
} from '../promptCatalogRuntime'

describe('promptCatalogRuntime', () => {
  it('resolves page title and description', () => {
    const copy = {
      title: 'All',
      description: 'All desc',
      caseTitle: 'Cases',
      caseDescription: 'Cases desc',
      templateTitle: 'Templates',
      templateDescription: 'Templates desc',
    } as any

    expect(resolvePromptCatalogPageTitle(copy, '')).toBe('Cases')
    expect(resolvePromptCatalogPageTitle(copy, 'template')).toBe('Cases')
    expect(resolvePromptCatalogPageDescription(copy, '')).toBe('Cases desc')
    expect(resolvePromptCatalogPageDescription(copy, 'template')).toBe('Cases desc')
  })

  it('applies configured defaults and builds list params', () => {
    const filters: PromptCatalogFilters = {
      search: 'abc',
      category: 'portrait',
      hasImage: false,
    }

    applyPromptCatalogDefaults(filters, {
      hasImage: true,
      pageSize: 12,
      sortBy: 'title',
      sortOrder: 'asc',
    })

    expect(filters.hasImage).toBe(true)

    expect(buildPromptCatalogListParams(filters, {
      pageSize: 12,
      sortBy: 'title',
      sortOrder: 'asc',
    }, 2)).toEqual({
      page: 2,
      page_size: 12,
      category: 'portrait',
      search: 'abc',
      has_image: true,
      sort_by: 'title',
      sort_order: 'asc',
    })
  })

  it('resolves generator/import defaults and formatting helpers', () => {
    expect(resolvePromptCatalogGeneratorPath({ generatorPath: '/image-generator' })).toBe('/image-generator')
    expect(resolvePromptCatalogGeneratorPath(undefined)).toBe('/image-generator')
    expect(resolvePromptCatalogGeneratorDraftSource({ generatorDraftSource: 'catalog' })).toBe('catalog')
    expect(resolvePromptCatalogImportXAuto(undefined)).toBe(true)
    expect(resolvePromptCatalogImportXAuto({ importXAuto: false })).toBe(false)
    expect(resolvePromptCatalogFacetLabel({ display_label: 'API Source' } as any)).toBe('API Source')
    expect(resolvePromptCatalogFacetLabel({ value: 'portrait', display_label: 'portrait' } as any, 'zh')).toBe('肖像')
    expect(resolvePromptCatalogFacetLabel({ value: 'awesome_nano_banana_pro_prompts', display_label: 'awesome_nano_banana_pro_prompts' } as any, 'zh')).toBe('Nano Banana Pro 提示库')
    expect(resolvePromptCatalogValueLabel('Photography & Realism', 'zh')).toBe('摄影与写实')
    expect(buildPromptCatalogImportSuccessMessage('Imported', 'Case A')).toBe('Imported: Case A')
    expect(formatPromptCatalogDate('2026-06-18T00:00:00Z', 'en')).toContain('2026')
    expect(formatPromptCatalogDate('', 'en')).toBe('')
  })

  it('does not return Chinese labels for English locale', () => {
    // resolvePromptCatalogValueLabel should not consult the zh dictionary when locale is 'en'
    expect(resolvePromptCatalogValueLabel('portrait', 'en')).toBe('portrait')
    expect(resolvePromptCatalogValueLabel('Photography & Realism', 'en')).toBe('Photography & Realism')

    // resolvePromptCatalogFacetLabel should fall through to display_label for English
    expect(resolvePromptCatalogFacetLabel({ value: 'portrait', display_label: 'Portrait Cases' } as any, 'en')).toBe('Portrait Cases')
    // When display_label matches value, return the raw value
    expect(resolvePromptCatalogFacetLabel({ value: 'portrait', display_label: 'portrait' } as any, 'en')).toBe('portrait')
    expect(resolvePromptCatalogFacetLabel({ value: 'Infographic / Edu Visual', display_label: '信息图与教育视觉' } as any, 'en')).toBe('Infographic / Edu Visual')
  })
})
