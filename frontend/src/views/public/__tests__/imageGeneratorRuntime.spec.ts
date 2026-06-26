import { describe, expect, it } from 'vitest'

import { applyImageGeneratorDraft, resolveImageGeneratorCatalogPath } from '../imageGeneratorRuntime'

describe('imageGeneratorRuntime', () => {
  it('applies draft prompt/title safely and trims prompt length', () => {
    expect(
      applyImageGeneratorDraft(
        {
          prompt: '  hello world  ',
          title: '  draft title  ',
        },
        5,
      ),
    ).toEqual({
      prompt: 'hello',
      title: 'draft title',
    })
  })

  it('resolves catalog path safely', () => {
    expect(resolveImageGeneratorCatalogPath('/prompts')).toBe('/prompts')
    expect(resolveImageGeneratorCatalogPath('')).toBe('')
    expect(resolveImageGeneratorCatalogPath(undefined)).toBe('')
  })
})
