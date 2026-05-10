import { describe, expect, it } from 'vitest'

import { resolveDocsLink } from '../docs'

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
