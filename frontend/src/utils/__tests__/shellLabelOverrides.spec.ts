import { describe, expect, it } from 'vitest'
import { resolveShellLabelOverrides } from '../shellLabelOverrides'

const allowedKeys = ['title', 'description', 'submit'] as const

describe('resolveShellLabelOverrides', () => {
  it('reads locale labels and filters unknown keys', () => {
    const labels = resolveShellLabelOverrides(
      JSON.stringify({
        zh: {
          labels: {
            title: '标题',
            submit: '提交',
            ignored: 'ignored',
          },
        },
      }),
      'zh-CN',
      allowedKeys,
    )

    expect(labels).toEqual({
      title: '标题',
      submit: '提交',
    })
  })

  it('falls back through en and zh locale branches', () => {
    const labels = resolveShellLabelOverrides(
      JSON.stringify({
        en: {
          labels: {
            description: 'Description',
          },
        },
      }),
      'fr-FR',
      allowedKeys,
    )

    expect(labels).toEqual({
      description: 'Description',
    })
  })

  it('returns empty overrides for invalid JSON', () => {
    expect(resolveShellLabelOverrides('{bad json', 'en', allowedKeys)).toEqual({})
  })
})
