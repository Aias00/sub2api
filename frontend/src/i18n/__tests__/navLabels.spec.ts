import { describe, expect, it } from 'vitest'
import zh from '../locales/zh'
import en from '../locales/en'

const requiredSidebarNavKeys = [
  'promptCatalog',
  'imageGenerator',
  'wechatExport',
  'hotTopics',
  'availableGroups',
  'workers',
  'runtimeSettings',
] as const

describe('sidebar nav labels', () => {
  it('defines localized labels for public capability sidebar entries', () => {
    for (const key of requiredSidebarNavKeys) {
      expect(zh.nav[key]).toBeTruthy()
      expect(zh.nav[key]).not.toBe(`nav.${key}`)
      expect(en.nav[key]).toBeTruthy()
      expect(en.nav[key]).not.toBe(`nav.${key}`)
    }
  })
})
