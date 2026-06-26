import { describe, expect, it } from 'vitest'

import {
  platformAccentBarClass,
  platformBadgeClass,
  platformBadgeLightClass,
  platformLabel,
  platformTextClass,
} from '../platformColors'

describe('platformColors', () => {
  it('uses product labels only when a platform is provided', () => {
    expect(platformLabel('anthropic')).toBe('Anthropic')
    expect(platformLabel('openai')).toBe('OpenAI')
    expect(platformLabel('antigravity')).toBe('Antigravity')
    expect(platformLabel('gemini')).toBe('Gemini')
    expect(platformLabel('custom-platform')).toBe('custom-platform')
  })

  it('does not invent a platform label for missing backend data', () => {
    expect(platformLabel('')).toBe('')
    expect(platformLabel(null)).toBe('')
    expect(platformLabel(undefined)).toBe('')
  })

  it('keeps neutral styling defaults for missing or unknown platform data', () => {
    expect(platformBadgeClass(undefined)).toContain('slate')
    expect(platformBadgeLightClass(null)).toContain('slate')
    expect(platformTextClass('')).toContain('primary')
    expect(platformAccentBarClass('custom-platform')).toContain('primary')
  })
})
