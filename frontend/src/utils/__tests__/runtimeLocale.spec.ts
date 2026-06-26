import { describe, expect, it } from 'vitest'

import { resolveRuntimeLanguage, resolveRuntimeLocale } from '../runtimeLocale'

describe('runtimeLocale', () => {
  it('resolves raw locale strings', () => {
    expect(resolveRuntimeLocale('zh-CN')).toBe('zh-CN')
    expect(resolveRuntimeLanguage('zh-CN')).toBe('zh')
  })

  it('resolves ref-like locale objects', () => {
    expect(resolveRuntimeLocale({ value: 'en-US' })).toBe('en-US')
    expect(resolveRuntimeLanguage({ value: 'en-US' })).toBe('en')
  })

  it('falls back to an empty locale for missing values', () => {
    expect(resolveRuntimeLocale(null)).toBe('')
    expect(resolveRuntimeLocale({ value: undefined })).toBe('')
    expect(resolveRuntimeLanguage(null)).toBe('en')
  })
})
