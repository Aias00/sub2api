import { describe, it, expect, vi, beforeEach } from 'vitest'
import type { UpstreamPreset } from '@/api/admin/accounts'
import {
  resolveUpstreamPresetApply,
  useUpstreamPresets,
  __resetUpstreamPresetsCacheForTest
} from '../useUpstreamPresets'

const samplePreset: UpstreamPreset = {
  id: 'deepseek',
  display_name: 'DeepSeek',
  platform: 'openai',
  account_type: 'apikey',
  api_style: 'openai',
  base_url: 'https://api.deepseek.com/v1',
  default_models: ['deepseek-chat', 'deepseek-reasoner'],
  docs_url: 'https://api-docs.deepseek.com'
}

const listUpstreamPresets = vi.fn()

vi.mock('@/api/admin/accounts', () => ({
  default: {
    listUpstreamPresets: (...args: unknown[]) => listUpstreamPresets(...args)
  }
}))

describe('resolveUpstreamPresetApply', () => {
  it('maps preset fields to form apply values', () => {
    const result = resolveUpstreamPresetApply(samplePreset)
    expect(result).toEqual({
      platform: 'openai',
      accountType: 'apikey',
      baseUrl: 'https://api.deepseek.com/v1',
      models: ['deepseek-chat', 'deepseek-reasoner'],
      apiStyle: 'openai'
    })
  })

  it('returns a fresh models array (no shared mutation)', () => {
    const result = resolveUpstreamPresetApply(samplePreset)
    result.models.push('mutated')
    expect(samplePreset.default_models).toEqual(['deepseek-chat', 'deepseek-reasoner'])
  })

  it('tolerates missing default_models', () => {
    const result = resolveUpstreamPresetApply({ ...samplePreset, default_models: undefined as unknown as string[] })
    expect(result.models).toEqual([])
  })
})

describe('useUpstreamPresets', () => {
  beforeEach(() => {
    __resetUpstreamPresetsCacheForTest()
    listUpstreamPresets.mockReset()
  })

  it('loads presets and finds by id', async () => {
    listUpstreamPresets.mockResolvedValue([samplePreset])
    const { load, findById, presets, loaded } = useUpstreamPresets()

    await load()

    expect(presets.value).toHaveLength(1)
    expect(loaded.value).toBe(true)
    expect(findById('deepseek')?.display_name).toBe('DeepSeek')
    expect(findById('missing')).toBeUndefined()
  })

  it('caches after first load (only one API call)', async () => {
    listUpstreamPresets.mockResolvedValue([samplePreset])
    const { load } = useUpstreamPresets()

    await load()
    await load()

    expect(listUpstreamPresets).toHaveBeenCalledTimes(1)
  })

  it('degrades gracefully on API error', async () => {
    listUpstreamPresets.mockRejectedValue(new Error('boom'))
    const { load, presets, loaded } = useUpstreamPresets()

    const result = await load()

    expect(result).toEqual([])
    expect(presets.value).toEqual([])
    expect(loaded.value).toBe(true)
  })
})
