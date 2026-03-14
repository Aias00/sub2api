import { describe, expect, it } from 'vitest'
import type { ApiKey, GroupPlatform } from '@/types'
import {
  buildGatewayCurlExample,
  buildGatewayModelsRelativePath,
  buildGatewayRelativePath,
  buildGatewayRequestBody,
  extractGatewayModelOptions,
  extractGatewayResponsePreview,
  getGatewayFallbackModelOptions,
  getGatewayVariantsForApiKey
} from '@/utils/gatewayDocs'

function createApiKey(platform: GroupPlatform, allowMessagesDispatch?: boolean): ApiKey {
  return {
    id: 1,
    user_id: 1,
    key: 'sk-test',
    name: 'Test Key',
    group_id: 1,
    status: 'active',
    ip_whitelist: [],
    ip_blacklist: [],
    last_used_at: null,
    quota: 0,
    quota_used: 0,
    expires_at: null,
    created_at: '2026-03-13T00:00:00Z',
    updated_at: '2026-03-13T00:00:00Z',
    rate_limit_5h: 0,
    rate_limit_1d: 0,
    rate_limit_7d: 0,
    usage_5h: 0,
    usage_1d: 0,
    usage_7d: 0,
    window_5h_start: null,
    window_1d_start: null,
    window_7d_start: null,
    reset_5h_at: null,
    reset_1d_at: null,
    reset_7d_at: null,
    group: {
      id: 1,
      name: 'Test Group',
      description: null,
      platform,
      rate_multiplier: 1,
      is_exclusive: false,
      status: 'active',
      subscription_type: 'standard',
      daily_limit_usd: null,
      weekly_limit_usd: null,
      monthly_limit_usd: null,
      image_price_1k: null,
      image_price_2k: null,
      image_price_4k: null,
      sora_image_price_360: null,
      sora_image_price_540: null,
      sora_video_price_per_request: null,
      sora_video_price_per_request_hd: null,
      sora_storage_quota_bytes: 0,
      claude_code_only: false,
      fallback_group_id: null,
      fallback_group_id_on_invalid_request: null,
      allow_messages_dispatch: allowMessagesDispatch,
      created_at: '2026-03-13T00:00:00Z',
      updated_at: '2026-03-13T00:00:00Z'
    }
  }
}

describe('gatewayDocs', () => {
  it('returns OpenAI variants plus Anthropic compatibility when messages dispatch is enabled', () => {
    const variants = getGatewayVariantsForApiKey(createApiKey('openai', true))
    expect(variants.map(variant => variant.id)).toEqual([
      'openaiChat',
      'openaiResponses',
      'anthropicMessages'
    ])
  })

  it('returns Antigravity dedicated routes', () => {
    const variants = getGatewayVariantsForApiKey(createApiKey('antigravity'))
    expect(variants.map(variant => variant.id)).toEqual([
      'antigravityMessages',
      'antigravityGemini'
    ])
  })

  it('builds Gemini paths with encoded model names', () => {
    expect(buildGatewayRelativePath('geminiNative', 'gemini-2.5-flash')).toBe(
      '/v1beta/models/gemini-2.5-flash:generateContent'
    )
  })

  it('builds model list paths for each protocol family', () => {
    expect(buildGatewayModelsRelativePath('anthropicMessages')).toBe('/v1/models')
    expect(buildGatewayModelsRelativePath('geminiNative')).toBe('/v1beta/models')
    expect(buildGatewayModelsRelativePath('antigravityGemini')).toBe('/antigravity/v1beta/models')
  })

  it('builds Gemini request bodies in native format', () => {
    expect(buildGatewayRequestBody('antigravityGemini', 'gemini-2.5-flash', 'hello', false)).toEqual({
      contents: [{ parts: [{ text: 'hello' }] }]
    })
  })

  it('uses x-goog-api-key in Gemini curl examples', () => {
    const curl = buildGatewayCurlExample(
      'http://localhost:8080',
      'sk-test',
      'geminiNative',
      'gemini-2.5-flash',
      'hello',
      false
    )
    expect(curl).toContain("x-goog-api-key: sk-test")
    expect(curl).toContain('/v1beta/models/gemini-2.5-flash:generateContent')
  })

  it('extracts response previews from Anthropic-compatible responses', () => {
    const preview = extractGatewayResponsePreview('anthropicMessages', JSON.stringify({
      content: [{ type: 'text', text: 'Hello from Claude' }]
    }))
    expect(preview).toBe('Hello from Claude')
  })

  it('extracts response previews from Anthropic SSE streams', () => {
    const preview = extractGatewayResponsePreview(
      'anthropicMessages',
      [
        'event: content_block_start',
        'data: {"content_block":{"text":"","type":"text"},"index":0,"type":"content_block_start"}',
        '',
        'event: content_block_delta',
        'data: {"delta":{"text":"Hello"},"index":0,"type":"content_block_delta"}',
        '',
        'event: content_block_delta',
        'data: {"delta":{"text":" world"},"index":0,"type":"content_block_delta"}'
      ].join('\n')
    )

    expect(preview).toBe('Hello world')
  })

  it('extracts response previews from OpenAI chat SSE streams', () => {
    const preview = extractGatewayResponsePreview(
      'openaiChat',
      [
        'data: {"choices":[{"delta":{"content":"Hello"}}]}',
        '',
        'data: {"choices":[{"delta":{"content":" world"}}]}',
        '',
        'data: [DONE]'
      ].join('\n')
    )

    expect(preview).toBe('Hello world')
  })

  it('extracts Claude model options from Anthropic-style list responses', () => {
    const options = extractGatewayModelOptions('antigravityMessages', JSON.stringify({
      data: [
        { id: 'claude-sonnet-4-5', display_name: 'Claude Sonnet 4.5' },
        { id: 'gemini-2.5-flash', display_name: 'Gemini 2.5 Flash' }
      ]
    }))

    expect(options.map(option => option.id)).toEqual(['claude-sonnet-4-5'])
  })

  it('extracts Gemini model options from Google-style list responses', () => {
    const options = extractGatewayModelOptions('antigravityGemini', JSON.stringify({
      models: [
        { name: 'models/gemini-2.5-flash', displayName: 'Gemini 2.5 Flash' },
        { name: 'models/gemini-3.1-pro-high', displayName: 'Gemini 3.1 Pro High' }
      ]
    }))

    expect(options.map(option => option.id)).toEqual([
      'gemini-2.5-flash',
      'gemini-3.1-pro-high'
    ])
    expect(options[0]?.label).toBe('Gemini 2.5 Flash')
  })

  it('falls back to curated model options when parsing fails', () => {
    const options = extractGatewayModelOptions('geminiNative', 'not-json')
    expect(options).toEqual(getGatewayFallbackModelOptions('geminiNative'))
  })
})
