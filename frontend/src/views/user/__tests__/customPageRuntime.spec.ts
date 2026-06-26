import { describe, expect, it } from 'vitest'

import {
  buildCustomPageImageUrl,
  isRelativeCustomPageMarkdownAsset,
  resolveCustomPageMarkdownSlug,
  resolveCustomPageMenuItem,
} from '../customPageRuntime'

describe('customPageRuntime', () => {
  it('resolves menu items and markdown slug', () => {
    const publicItems = [{ id: 'docs', url: 'md:guide' }]
    const adminItems = [{ id: 'admin', url: 'https://example.com' }]

    expect(resolveCustomPageMenuItem('docs', publicItems as any, adminItems as any, false)?.id).toBe('docs')
    expect(resolveCustomPageMenuItem('admin', publicItems as any, adminItems as any, true)?.id).toBe('admin')
    expect(resolveCustomPageMarkdownSlug({ id: 'docs', url: 'md:guide' } as any)).toBe('guide')
  })

  it('handles relative markdown assets and builds image urls', () => {
    expect(isRelativeCustomPageMarkdownAsset('images/a.png')).toBe(true)
    expect(isRelativeCustomPageMarkdownAsset('https://example.com/a.png')).toBe(false)
    expect(buildCustomPageImageUrl('guide', 'images/a.png')).toBe('/api/v1/pages/guide/images/images/a.png')
  })
})
