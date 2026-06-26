import { describe, expect, it } from 'vitest'
import { resolveDocsContentBasePath } from '../docsContentBasePath'

describe('resolveDocsContentBasePath', () => {
  it('does not invent bundled docs-content paths when config is missing', () => {
    expect(resolveDocsContentBasePath(undefined, 'zh')).toBe('')
    expect(resolveDocsContentBasePath('', 'en')).toBe('')
  })

  it('resolves locale-scoped JSON paths', () => {
    const raw = JSON.stringify({
      zh: '/managed-docs/zh',
      en: 'https://cdn.example.com/docs/en',
    })

    expect(resolveDocsContentBasePath(raw, 'zh')).toBe('/managed-docs/zh/')
    expect(resolveDocsContentBasePath(raw, 'en')).toBe('https://cdn.example.com/docs/en/')
  })

  it('supports direct path values', () => {
    expect(resolveDocsContentBasePath('/managed-docs', 'zh')).toBe('/managed-docs/')
  })

  it('rejects unsafe protocols', () => {
    expect(resolveDocsContentBasePath('javascript:alert(1)', 'zh')).toBe('')
  })
})
