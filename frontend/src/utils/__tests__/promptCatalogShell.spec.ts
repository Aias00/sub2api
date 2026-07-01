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

    expect(config.labels).toMatchObject({
      title: 'Prompt Library',
      caseTitle: 'Prompt Cases',
      templateTitle: 'Prompt Templates',
      details: 'Open case',
      importProviderX: 'X source',
      importPlaceholder: 'Paste an X/Twitter post URL',
      importAction: 'Import',
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
    expect(config.labels?.importProviderX).toBe('Twitter import')
    expect(config.labels?.importAction).toBe('Import')
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
    expect(config.labels?.importProviderX).toBe('Twitter 导入')
  })

  it('returns an empty config for invalid JSON', () => {
    expect(resolvePromptCatalogShellConfig('{bad json', 'en').labels?.importAction).toBe('Import')
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

  it('fills missing localized labels from defaults so admin import controls are never blank', () => {
    const config = resolvePromptCatalogShellConfig(
      JSON.stringify({
        zh: {
          labels: {
            title: '配置标题',
            copyPrompt: '配置复制',
          },
        },
      }),
      'zh',
    )

    expect(config.labels?.title).toBe('配置标题')
    expect(config.labels?.copyPrompt).toBe('配置复制')
    expect(config.labels?.importProviderX).toBe('Twitter 导入')
    expect(config.labels?.importPlaceholder).toBe('粘贴 X/Twitter 帖子链接')
    expect(config.labels?.importAction).toBe('导入')
    expect(config.labels?.search).toBe('搜索')
  })
})
