import { afterEach, describe, expect, it } from 'vitest'

import {
  clearImageGeneratorDraft,
  loadImageGeneratorDraft,
  saveImageGeneratorDraft,
} from '@/utils/imageGeneratorDraft'

const DRAFT_KEY = 'cloudbase:image-generator:draft'

describe('imageGeneratorDraft', () => {
  afterEach(() => {
    window.sessionStorage.clear()
  })

  it('stores drafts under the generic Cloudbase key', () => {
    saveImageGeneratorDraft({
      prompt: 'A cinematic city skyline',
      title: 'City',
      sourcePromptId: 'prompt-1',
      source: 'catalog',
    })

    expect(window.sessionStorage.getItem(DRAFT_KEY)).toContain('A cinematic city skyline')
    expect(loadImageGeneratorDraft()).toEqual({
      prompt: 'A cinematic city skyline',
      title: 'City',
      sourcePromptId: 'prompt-1',
      source: 'catalog',
    })
  })

  it('ignores legacy Touch draft keys', () => {
    window.sessionStorage.setItem('touch:image-generator:draft', JSON.stringify({
      prompt: 'Legacy prompt',
      title: 'Legacy title',
    }))

    expect(loadImageGeneratorDraft()).toBeNull()
  })

  it('clears the generic draft key', () => {
    window.sessionStorage.setItem(DRAFT_KEY, '{"prompt":"new"}')

    clearImageGeneratorDraft()

    expect(window.sessionStorage.getItem(DRAFT_KEY)).toBeNull()
  })

  it('returns null and clears storage when JSON is corrupted', () => {
    window.sessionStorage.setItem(DRAFT_KEY, '{invalid json')

    const result = loadImageGeneratorDraft()

    expect(result).toBeNull()
    expect(window.sessionStorage.getItem(DRAFT_KEY)).toBeNull()
  })

  it('returns null when draft has no prompt', () => {
    window.sessionStorage.setItem(DRAFT_KEY, JSON.stringify({ title: 'No prompt' }))

    expect(loadImageGeneratorDraft()).toBeNull()
  })
})
