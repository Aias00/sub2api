import { describe, expect, it } from 'vitest'
import {
  MODEL_PLAZA_ALL_GROUP_KEY,
  MODEL_PLAZA_OTHER_GROUP_KEY,
  MODEL_PLAZA_UNKNOWN_PROVIDER_INITIAL,
  formatModelsPlazaTemplate,
  resolveModelPlazaProviderGroupKey,
  resolveModelPlazaProviderGroupLabel,
  resolveModelPlazaProviderGroupRank,
  resolveModelPlazaProviderIconClass,
  resolveModelPlazaProviderInitial,
  resolveModelsPlazaCopy,
} from '../modelPlazaDisplay'

describe('model plaza display helpers', () => {
  it('centralizes provider group keys', () => {
    expect(MODEL_PLAZA_ALL_GROUP_KEY).toBe('all')
    expect(resolveModelPlazaProviderGroupKey('Anthropic Claude')).toBe('claude')
    expect(resolveModelPlazaProviderGroupKey('OpenAI GPT')).toBe('gpt')
    expect(resolveModelPlazaProviderGroupKey('Google Gemini')).toBe('gemini')
    expect(resolveModelPlazaProviderGroupKey('')).toBe(MODEL_PLAZA_OTHER_GROUP_KEY)
  })

  it('centralizes provider group ranking and initials', () => {
    expect(resolveModelPlazaProviderGroupRank('claude')).toBeLessThan(resolveModelPlazaProviderGroupRank('gpt'))
    expect(resolveModelPlazaProviderGroupRank(MODEL_PLAZA_OTHER_GROUP_KEY)).toBe(99)
    expect(resolveModelPlazaProviderGroupLabel('claude', { groupOther: '其他' })).toBe('Claude')
    expect(resolveModelPlazaProviderGroupLabel('gpt', { groupOther: '其他' })).toBe('GPT')
    expect(resolveModelPlazaProviderGroupLabel('gemini', { groupOther: '其他' })).toBe('Gemini')
    expect(resolveModelPlazaProviderGroupLabel(MODEL_PLAZA_OTHER_GROUP_KEY, { groupOther: '其他' })).toBe('其他')
    expect(resolveModelPlazaProviderGroupLabel('custom', { groupOther: '其他' })).toBe('CUSTOM')
    expect(resolveModelPlazaProviderInitial('claude')).toBe('C')
    expect(resolveModelPlazaProviderInitial('openai')).toBe('G')
    expect(resolveModelPlazaProviderInitial('')).toBe(MODEL_PLAZA_UNKNOWN_PROVIDER_INITIAL)
  })

  it('centralizes provider icon styling', () => {
    expect(resolveModelPlazaProviderIconClass('claude')).toContain('#ef8e67')
    expect(resolveModelPlazaProviderIconClass('gpt')).toContain('#48b774')
    expect(resolveModelPlazaProviderIconClass('gemini')).toContain('#5b7cff')
    expect(resolveModelPlazaProviderIconClass('unknown')).toContain('#64748b')
  })

  it('centralizes copy resolution and simple template formatting', () => {
    const copy = resolveModelsPlazaCopy(
      JSON.stringify({
        zh: {
          labels: {
            title: '模型目录',
            currentSearch: '搜索：{query}',
            ignored: 'ignored',
          },
        },
      }),
      'zh-CN',
    )

    expect(copy.title).toBe('模型目录')
    expect(copy.currentSearch).toBe('搜索：{query}')
    expect(copy.login).toBe('')
    expect(formatModelsPlazaTemplate(copy.currentSearch, { query: 'claude' })).toBe('搜索：claude')
  })
})
