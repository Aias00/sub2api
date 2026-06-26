import { describe, expect, it } from 'vitest'
import { resolveDocsShellConfig, resolveDocsShellCopy } from '../docsShell'

describe('docs shell helpers', () => {
  it('resolves docs shell copy from localized public settings', () => {
    const copy = resolveDocsShellCopy(
      JSON.stringify({
        en: {
          labels: {
            title: 'Docs',
            dashboard: 'Dashboard',
            searchPlaceholder: 'Search docs',
            ignored: 'ignored',
          },
        },
      }),
      'en-US',
    )

    expect(copy.title).toBe('Docs')
    expect(copy.dashboard).toBe('Dashboard')
    expect(copy.searchPlaceholder).toBe('Search docs')
    expect(copy.login).toBe('')
  })

  it('resolves app route links from public docs shell defaults', () => {
    const config = resolveDocsShellConfig(
      JSON.stringify({
        en: {
          defaults: {
            appRouteLinks: [
              '/home',
              '#/dashboard?from=docs',
              '/register/',
              '/home',
              'https://evil.example/path',
              '//evil.example/path',
              'relative-path',
              '',
            ],
          },
        },
      }),
      'en-US',
    )

    expect(config.defaults.appRouteLinks).toEqual(['#/home', '#/dashboard', '#/register'])
  })

  it('does not synthesize default app route links when docs shell config is absent', () => {
    expect(resolveDocsShellConfig(undefined, 'en').defaults.appRouteLinks).toEqual([])
    expect(resolveDocsShellConfig('{bad json', 'en').defaults.appRouteLinks).toEqual([])
  })
})
