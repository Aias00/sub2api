import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const dir = dirname(fileURLToPath(import.meta.url))
const sources = {
  appHeader: readFileSync(resolve(dir, '../AppHeader.vue'), 'utf8'),
  publicDarkHeader: readFileSync(resolve(dir, '../PublicDarkHeader.vue'), 'utf8'),
  homeView: readFileSync(resolve(dir, '../../../views/HomeView.vue'), 'utf8'),
  keyUsageView: readFileSync(resolve(dir, '../../../views/KeyUsageView.vue'), 'utf8'),
}

describe('doc_url sanitization', () => {
  it('sanitizes AppHeader docUrl', () => {
    expect(sources.appHeader).toContain("import { sanitizeUrl } from '@/utils/url'")
    expect(sources.appHeader).toContain('sanitizeUrl(appStore.docUrl)')
  })

  it('sanitizes PublicDarkHeader docUrl', () => {
    expect(sources.publicDarkHeader).toContain("import { sanitizeUrl } from '@/utils/url'")
    expect(sources.publicDarkHeader).toContain('sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl')
  })

  it('sanitizes HomeView docUrl', () => {
    expect(sources.homeView).toContain("import { sanitizeUrl } from '@/utils/url'")
    expect(sources.homeView).toContain('sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl')
  })

  it('sanitizes KeyUsageView docUrl', () => {
    expect(sources.keyUsageView).toContain("import { sanitizeUrl } from '@/utils/url'")
    expect(sources.keyUsageView).toContain('sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl')
  })
})
