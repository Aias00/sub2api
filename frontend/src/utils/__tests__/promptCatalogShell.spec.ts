import { describe, expect, it } from 'vitest'
import { resolvePromptCatalogShellConfig } from '../promptCatalogShell'

describe('resolvePromptCatalogShellConfig', () => {
  it('resolves localized prompt catalog labels and filters unknown keys', () => {
    const config = resolvePromptCatalogShellConfig(
      JSON.stringify({
        en: {
          defaults: {
            sourceType: 'template',
            hasImage: true,
            pageSize: 12,
            sortBy: 'title',
            sortOrder: 'asc',
            generatorPath: '/workspace/image',
            generatorDraftSource: 'configured-prompt-catalog',
            importXAuto: false,
          },
          labels: {
            title: 'Prompt Library',
            caseTitle: 'Prompt Cases',
            templateTitle: 'Prompt Templates',
            details: 'Open case',
            importProviderX: 'X source',
            ignored: 'ignored',
          },
        },
      }),
      'en',
    )

    expect(config.labels).toEqual({
      title: 'Prompt Library',
      caseTitle: 'Prompt Cases',
      templateTitle: 'Prompt Templates',
      details: 'Open case',
      importProviderX: 'X source',
    })
    expect(config.defaults).toEqual({
      sourceType: 'template',
      hasImage: true,
      pageSize: 12,
      sortBy: 'title',
      sortOrder: 'asc',
      generatorPath: '/workspace/image',
      generatorDraftSource: 'configured-prompt-catalog',
      importXAuto: false,
    })
  })

  it('drops invalid defaults instead of owning business fallbacks in the frontend', () => {
    const config = resolvePromptCatalogShellConfig(
      JSON.stringify({
        en: {
          defaults: {
            sourceType: 'unknown',
            hasImage: 'yes',
            pageSize: 0,
            sortBy: 'bad',
            sortOrder: 'bad',
            generatorPath: 'https://example.com',
            generatorDraftSource: '',
            importXAuto: 'yes',
          },
        },
      }),
      'en',
    )

    expect(config.defaults).toBeUndefined()
  })

  it('falls back to root labels when the selected locale is missing', () => {
    const config = resolvePromptCatalogShellConfig(
      JSON.stringify({
        labels: {
          searchPlaceholder: 'Search all prompts',
        },
      }),
      'zh',
    )

    expect(config.labels?.searchPlaceholder).toBe('Search all prompts')
  })

  it('returns an empty config for invalid JSON', () => {
    expect(resolvePromptCatalogShellConfig('{bad json', 'en')).toEqual({})
  })

  it('rejects pageSize values exceeding the maximum allowed', () => {
    const config = resolvePromptCatalogShellConfig(
      JSON.stringify({
        en: {
          defaults: {
            pageSize: 1000000,
          },
        },
      }),
      'en',
    )

    expect(config.defaults?.pageSize).toBeUndefined()
  })
})
