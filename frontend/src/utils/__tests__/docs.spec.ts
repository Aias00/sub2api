import { describe, expect, it } from 'vitest'

import {
  defaultDocsSlug,
  docsPages,
  docsSections,
  findDocsPage,
  resolveDocsLink,
} from '../docs'

describe('docs registry', () => {
  it('exposes a non-empty default page and grouped sections', () => {
    expect(defaultDocsSlug).toBeTruthy()
    expect(docsPages.length).toBeGreaterThanOrEqual(5)
    expect(docsSections.length).toBeGreaterThanOrEqual(2)
    expect(findDocsPage(defaultDocsSlug)?.content.length).toBeGreaterThan(20)
  })
})

describe('resolveDocsLink', () => {
  it('routes same-origin landing URLs to the internal docs page', () => {
    expect(resolveDocsLink('https://cloudbase.eu.org/', 'https://cloudbase.eu.org')).toEqual({
      internal: true,
      to: '/docs',
      href: '/docs',
    })
    expect(resolveDocsLink('https://cloudbase.eu.org/home', 'https://cloudbase.eu.org')).toEqual({
      internal: true,
      to: '/docs',
      href: '/docs',
    })
  })

  it('keeps external documentation links external', () => {
    expect(resolveDocsLink('https://docs.example.com', 'https://cloudbase.eu.org')).toEqual({
      internal: false,
      to: 'https://docs.example.com/',
      href: 'https://docs.example.com/',
    })
  })
})
