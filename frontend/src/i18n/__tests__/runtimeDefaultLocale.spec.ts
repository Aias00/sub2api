import { afterEach, describe, expect, it, vi } from 'vitest'

describe('runtime default locale', () => {
  afterEach(() => {
    localStorage.clear()
    delete window.__APP_CONFIG__
    vi.resetModules()
  })

  it('uses public settings default_locale when the user has no saved locale', async () => {
    window.__APP_CONFIG__ = { default_locale: 'zh-CN' } as typeof window.__APP_CONFIG__

    const { getLocale } = await import('../index')

    expect(getLocale()).toBe('zh')
  })

  it('keeps the user saved locale ahead of public settings default_locale', async () => {
    localStorage.setItem('cloudbase_locale', 'en')
    window.__APP_CONFIG__ = { default_locale: 'zh' } as typeof window.__APP_CONFIG__

    const { getLocale } = await import('../index')

    expect(getLocale()).toBe('en')
  })
})
