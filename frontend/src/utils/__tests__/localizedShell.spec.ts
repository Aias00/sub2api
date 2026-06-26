import { describe, expect, it } from 'vitest'
import { resolveLocalizedShellLabels } from '../localizedShell'

const keys = ['title', 'login', 'noData'] as const

describe('resolveLocalizedShellLabels', () => {
  it('resolves allowed localized labels without fallback copy', () => {
    const labels = resolveLocalizedShellLabels(
      JSON.stringify({
        zh: {
          labels: {
            title: '文档',
            login: '登录',
            ignored: 'ignored',
          },
        },
      }),
      'zh-CN',
      keys,
    )

    expect(labels).toEqual({
      title: '文档',
      login: '登录',
      noData: '',
    })
  })

  it('falls back through en and zh locale branches', () => {
    const labels = resolveLocalizedShellLabels(
      JSON.stringify({
        en: {
          labels: {
            noData: 'Nothing found',
          },
        },
      }),
      'fr-FR',
      keys,
    )

    expect(labels.noData).toBe('Nothing found')
    expect(labels.title).toBe('')
  })

  it('uses root labels when locale branches are absent', () => {
    const labels = resolveLocalizedShellLabels(
      JSON.stringify({
        labels: {
          title: 'Root title',
        },
      }),
      'en',
      keys,
    )

    expect(labels.title).toBe('Root title')
    expect(labels.login).toBe('')
  })

  it('returns empty labels for invalid JSON', () => {
    expect(resolveLocalizedShellLabels('{bad json', 'zh', keys)).toEqual({
      title: '',
      login: '',
      noData: '',
    })
  })
})
