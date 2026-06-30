import { describe, expect, it, beforeEach } from 'vitest'
import { applyPublicTheme, getStoredPublicTheme, initPublicTheme } from '../usePublicTheme'

describe('usePublicTheme', () => {
  beforeEach(() => {
    localStorage.clear()
    document.documentElement.removeAttribute('data-public-theme')
  })

  it('defaults public tool pages to the dark template', () => {
    expect(getStoredPublicTheme()).toBe('dark')
    initPublicTheme()
    expect(document.documentElement.dataset.publicTheme).toBe('dark')
    expect(localStorage.getItem('public-theme')).toBe('dark')
  })

  it('persists and applies the requested public theme', () => {
    applyPublicTheme('light')
    expect(getStoredPublicTheme()).toBe('light')
    expect(document.documentElement.dataset.publicTheme).toBe('light')
  })
})
