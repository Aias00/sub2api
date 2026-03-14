import type { ApiKey } from '@/types'

export type GatewayVariantId =
  | 'anthropicMessages'
  | 'openaiChat'
  | 'openaiResponses'
  | 'geminiNative'
  | 'antigravityMessages'
  | 'antigravityGemini'

type GatewayVariantProtocol = 'anthropic' | 'openai' | 'google'
type GatewayBodyKind = 'anthropic' | 'openaiChat' | 'openaiResponses' | 'geminiGenerate'

export interface GatewayVariant {
  id: GatewayVariantId
  translationKey: string
  protocol: GatewayVariantProtocol
  bodyKind: GatewayBodyKind
  headerMode: 'bearer' | 'x-goog-api-key'
  supportsStream: boolean
  defaultModel: string
  modelPlaceholder: string
}

export interface GatewayModelOption {
  id: string
  label: string
  description?: string
}

const gatewayVariants: Record<GatewayVariantId, GatewayVariant> = {
  anthropicMessages: {
    id: 'anthropicMessages',
    translationKey: 'gateway.variants.anthropicMessages',
    protocol: 'anthropic',
    bodyKind: 'anthropic',
    headerMode: 'bearer',
    supportsStream: true,
    defaultModel: 'claude-sonnet-4-5',
    modelPlaceholder: 'claude-sonnet-4-5'
  },
  openaiChat: {
    id: 'openaiChat',
    translationKey: 'gateway.variants.openaiChat',
    protocol: 'openai',
    bodyKind: 'openaiChat',
    headerMode: 'bearer',
    supportsStream: true,
    defaultModel: 'gpt-4.1',
    modelPlaceholder: 'gpt-4.1'
  },
  openaiResponses: {
    id: 'openaiResponses',
    translationKey: 'gateway.variants.openaiResponses',
    protocol: 'openai',
    bodyKind: 'openaiResponses',
    headerMode: 'bearer',
    supportsStream: true,
    defaultModel: 'gpt-4.1',
    modelPlaceholder: 'gpt-4.1'
  },
  geminiNative: {
    id: 'geminiNative',
    translationKey: 'gateway.variants.geminiNative',
    protocol: 'google',
    bodyKind: 'geminiGenerate',
    headerMode: 'x-goog-api-key',
    supportsStream: false,
    defaultModel: 'gemini-2.5-flash',
    modelPlaceholder: 'gemini-2.5-flash'
  },
  antigravityMessages: {
    id: 'antigravityMessages',
    translationKey: 'gateway.variants.antigravityMessages',
    protocol: 'anthropic',
    bodyKind: 'anthropic',
    headerMode: 'bearer',
    supportsStream: true,
    defaultModel: 'claude-sonnet-4-5',
    modelPlaceholder: 'claude-sonnet-4-5'
  },
  antigravityGemini: {
    id: 'antigravityGemini',
    translationKey: 'gateway.variants.antigravityGemini',
    protocol: 'google',
    bodyKind: 'geminiGenerate',
    headerMode: 'x-goog-api-key',
    supportsStream: false,
    defaultModel: 'gemini-2.5-flash',
    modelPlaceholder: 'gemini-2.5-flash'
  }
}

const fallbackGatewayModels: Record<GatewayVariantId, string[]> = {
  anthropicMessages: [
    'claude-sonnet-4-6',
    'claude-sonnet-4-5',
    'claude-sonnet-4-5-thinking',
    'claude-opus-4-6',
    'claude-opus-4-6-thinking',
    'claude-opus-4-5-thinking',
    'claude-haiku-4-5',
    'claude-haiku-4-5-20251001'
  ],
  openaiChat: [
    'gpt-5.4',
    'gpt-5.1',
    'gpt-5',
    'gpt-4.1',
    'gpt-4.1-mini',
    'gpt-4o',
    'gpt-4o-mini',
    'o3',
    'o3-mini'
  ],
  openaiResponses: [
    'gpt-5.4',
    'gpt-5.1',
    'gpt-5',
    'gpt-4.1',
    'gpt-4.1-mini',
    'gpt-4o',
    'gpt-4o-mini',
    'o3',
    'o3-mini'
  ],
  geminiNative: [
    'gemini-2.5-flash',
    'gemini-2.5-pro',
    'gemini-2.5-flash-image',
    'gemini-3-flash-preview',
    'gemini-3-pro-preview'
  ],
  antigravityMessages: [
    'claude-sonnet-4-6',
    'claude-sonnet-4-5',
    'claude-sonnet-4-5-thinking',
    'claude-opus-4-6',
    'claude-opus-4-6-thinking',
    'claude-opus-4-5-thinking'
  ],
  antigravityGemini: [
    'gemini-2.5-flash',
    'gemini-2.5-flash-image',
    'gemini-2.5-flash-lite',
    'gemini-2.5-flash-thinking',
    'gemini-3-flash',
    'gemini-3-pro-high',
    'gemini-3-pro-low',
    'gemini-3.1-pro-high',
    'gemini-3.1-pro-low',
    'gemini-3.1-flash-image',
    'gemini-3-pro-preview',
    'gemini-3-pro-image'
  ]
}

export const DEFAULT_GATEWAY_TEST_PROMPT = '请简短介绍一下你当前命中的模型和主要能力。'

export function getGatewayVariantById(id: GatewayVariantId): GatewayVariant {
  return gatewayVariants[id]
}

export function getGatewayVariantsForApiKey(apiKey: ApiKey | null | undefined): GatewayVariant[] {
  const platform = apiKey?.group?.platform
  if (!platform) return []

  switch (platform) {
    case 'anthropic':
      return [gatewayVariants.anthropicMessages]
    case 'openai':
      return apiKey.group?.allow_messages_dispatch
        ? [
            gatewayVariants.openaiChat,
            gatewayVariants.openaiResponses,
            gatewayVariants.anthropicMessages
          ]
        : [gatewayVariants.openaiChat, gatewayVariants.openaiResponses]
    case 'gemini':
      return [gatewayVariants.geminiNative]
    case 'antigravity':
      return [gatewayVariants.antigravityMessages, gatewayVariants.antigravityGemini]
    default:
      return []
  }
}

export function getGatewayBaseUrl(apiBaseUrl?: string | null): string {
  const trimmed = typeof apiBaseUrl === 'string' ? apiBaseUrl.trim() : ''
  if (trimmed) {
    return trimmed.replace(/\/+$/, '')
  }
  if (typeof window !== 'undefined') {
    return window.location.origin
  }
  return ''
}

export function buildGatewayRelativePath(variantId: GatewayVariantId, model: string): string {
  const normalizedModel = encodeURIComponent(model.trim() || getGatewayVariantById(variantId).defaultModel)

  switch (variantId) {
    case 'anthropicMessages':
      return '/v1/messages'
    case 'openaiChat':
      return '/v1/chat/completions'
    case 'openaiResponses':
      return '/v1/responses'
    case 'geminiNative':
      return `/v1beta/models/${normalizedModel}:generateContent`
    case 'antigravityMessages':
      return '/antigravity/v1/messages'
    case 'antigravityGemini':
      return `/antigravity/v1beta/models/${normalizedModel}:generateContent`
  }
}

export function buildGatewayModelsRelativePath(variantId: GatewayVariantId): string {
  switch (variantId) {
    case 'anthropicMessages':
    case 'openaiChat':
    case 'openaiResponses':
      return '/v1/models'
    case 'geminiNative':
      return '/v1beta/models'
    case 'antigravityMessages':
      return '/antigravity/v1/models'
    case 'antigravityGemini':
      return '/antigravity/v1beta/models'
  }
}

export function buildGatewayAbsoluteUrl(baseUrl: string, variantId: GatewayVariantId, model: string): string {
  return `${getGatewayBaseUrl(baseUrl)}${buildGatewayRelativePath(variantId, model)}`
}

function dedupeGatewayModelOptions(options: GatewayModelOption[], defaultModel: string): GatewayModelOption[] {
  const byID = new Map<string, GatewayModelOption>()

  for (const option of options) {
    const normalizedID = option.id.trim()
    if (!normalizedID) continue
    if (!byID.has(normalizedID)) {
      byID.set(normalizedID, {
        id: normalizedID,
        label: option.label?.trim() || normalizedID,
        description: option.description?.trim() || undefined
      })
    }
  }

  if (defaultModel && !byID.has(defaultModel)) {
    byID.set(defaultModel, { id: defaultModel, label: defaultModel })
  }

  return Array.from(byID.values())
}

function normalizeGoogleModelName(name: unknown): string {
  if (typeof name !== 'string') return ''
  return name.replace(/^models\//, '').trim()
}

function shouldIncludeModelForVariant(variantId: GatewayVariantId, modelID: string): boolean {
  if (!modelID) return false

  switch (variantId) {
    case 'anthropicMessages':
    case 'antigravityMessages':
      return modelID.startsWith('claude-')
    case 'geminiNative':
    case 'antigravityGemini':
      return modelID.startsWith('gemini-')
    default:
      return true
  }
}

export function getGatewayFallbackModelOptions(variantId: GatewayVariantId): GatewayModelOption[] {
  return dedupeGatewayModelOptions(
    fallbackGatewayModels[variantId].map((id) => ({ id, label: id })),
    gatewayVariants[variantId].defaultModel
  )
}

export function extractGatewayModelOptions(
  variantId: GatewayVariantId,
  responseText: string
): GatewayModelOption[] {
  if (!responseText.trim()) {
    return getGatewayFallbackModelOptions(variantId)
  }

  try {
    const parsed = JSON.parse(responseText)
    const variant = getGatewayVariantById(variantId)

    if (variant.bodyKind === 'geminiGenerate') {
      const models = Array.isArray(parsed?.models) ? parsed.models : []
      const options = models
        .map((model: any): GatewayModelOption => {
          const id = normalizeGoogleModelName(model?.name)
          return {
            id,
            label: typeof model?.displayName === 'string' && model.displayName.trim()
              ? model.displayName.trim()
              : id,
            description: id && model?.displayName && model.displayName !== id ? id : undefined
          }
        })
        .filter((option: GatewayModelOption) => shouldIncludeModelForVariant(variantId, option.id))

      return dedupeGatewayModelOptions(options, variant.defaultModel)
    }

    const data = Array.isArray(parsed?.data) ? parsed.data : []
    const options = data
      .map((model: any): GatewayModelOption => {
        const id = typeof model?.id === 'string'
          ? model.id.trim()
          : normalizeGoogleModelName(model?.name)
        return {
          id,
          label: typeof model?.display_name === 'string' && model.display_name.trim()
            ? model.display_name.trim()
            : typeof model?.displayName === 'string' && model.displayName.trim()
              ? model.displayName.trim()
              : id,
          description: id && model?.display_name && model.display_name !== id ? id : undefined
        }
      })
      .filter((option: GatewayModelOption) => shouldIncludeModelForVariant(variantId, option.id))

    return dedupeGatewayModelOptions(options, variant.defaultModel)
  } catch {
    return getGatewayFallbackModelOptions(variantId)
  }
}

export function buildGatewayHeaders(apiKey: string, variantId: GatewayVariantId): Record<string, string> {
  const variant = getGatewayVariantById(variantId)
  const headers: Record<string, string> = {
    'Content-Type': 'application/json'
  }

  if (variant.headerMode === 'x-goog-api-key') {
    headers['x-goog-api-key'] = apiKey
  } else {
    headers.Authorization = `Bearer ${apiKey}`
  }

  return headers
}

export function buildGatewayRequestBody(
  variantId: GatewayVariantId,
  model: string,
  prompt: string,
  stream: boolean,
): Record<string, unknown> {
  const variant = getGatewayVariantById(variantId)
  const selectedModel = model.trim() || variant.defaultModel
  const selectedPrompt = prompt.trim() || DEFAULT_GATEWAY_TEST_PROMPT

  switch (variant.bodyKind) {
    case 'anthropic':
      return {
        model: selectedModel,
        max_tokens: 256,
        stream,
        messages: [{ role: 'user', content: selectedPrompt }]
      }
    case 'openaiChat':
      return {
        model: selectedModel,
        stream,
        messages: [{ role: 'user', content: selectedPrompt }]
      }
    case 'openaiResponses':
      return {
        model: selectedModel,
        stream,
        input: selectedPrompt
      }
    case 'geminiGenerate':
      return {
        contents: [
          {
            parts: [{ text: selectedPrompt }]
          }
        ]
      }
  }
}

function shellQuote(value: string): string {
  return `'${value.replace(/'/g, `'\"'\"'`)}'`
}

export function buildGatewayCurlExample(
  baseUrl: string,
  apiKey: string,
  variantId: GatewayVariantId,
  model: string,
  prompt: string,
  stream: boolean,
): string {
  const body = buildGatewayRequestBody(variantId, model, prompt, stream)
  const headers = buildGatewayHeaders(apiKey, variantId)
  const headerLines = Object.entries(headers).map(([name, value]) => `  -H ${shellQuote(`${name}: ${value}`)}`)
  const payload = shellQuote(JSON.stringify(body, null, 2))

  return [
    `curl ${buildGatewayAbsoluteUrl(baseUrl, variantId, model)} \\`,
    ...headerLines.map((line) => `${line} \\`),
    `  -d ${payload}`
  ].join('\n')
}

function truncateText(value: string, limit: number = 600): string {
  if (value.length <= limit) return value
  return `${value.slice(0, limit)}...`
}

function extractAnthropicText(payload: any): string {
  const parts = Array.isArray(payload?.content) ? payload.content : []
  const texts = parts
    .map((part: any) => (typeof part?.text === 'string' ? part.text : ''))
    .filter(Boolean)
  return texts.join('\n').trim()
}

function extractOpenAIResponsesText(payload: any): string {
  if (typeof payload?.output_text === 'string' && payload.output_text.trim()) {
    return payload.output_text.trim()
  }

  const outputItems = Array.isArray(payload?.output) ? payload.output : []
  const texts: string[] = []
  for (const item of outputItems) {
    const contentParts = Array.isArray(item?.content) ? item.content : []
    for (const part of contentParts) {
      if (typeof part?.text === 'string' && part.text.trim()) {
        texts.push(part.text.trim())
      }
    }
  }
  return texts.join('\n').trim()
}

function extractGeminiText(payload: any): string {
  const candidates = Array.isArray(payload?.candidates) ? payload.candidates : []
  const texts: string[] = []
  for (const candidate of candidates) {
    const parts = Array.isArray(candidate?.content?.parts) ? candidate.content.parts : []
    for (const part of parts) {
      if (typeof part?.text === 'string' && part.text.trim()) {
        texts.push(part.text.trim())
      }
    }
  }
  return texts.join('\n').trim()
}

function extractSSEPayloads(responseText: string): any[] {
  return responseText
    .split(/\r?\n/)
    .filter((line) => line.startsWith('data:'))
    .map((line) => line.slice(5).trim())
    .filter((payload) => payload && payload !== '[DONE]')
    .flatMap((payload) => {
      try {
        return [JSON.parse(payload)]
      } catch {
        return []
      }
    })
}

function extractAnthropicSSEText(responseText: string): string {
  const texts: string[] = []

  for (const payload of extractSSEPayloads(responseText)) {
    const blockText = payload?.content_block?.text
    if (typeof blockText === 'string' && blockText) {
      texts.push(blockText)
    }

    const deltaText = payload?.delta?.text
    if (typeof deltaText === 'string' && deltaText) {
      texts.push(deltaText)
    }
  }

  return texts.join('').trim()
}

function extractOpenAIChatSSEText(responseText: string): string {
  const texts: string[] = []

  for (const payload of extractSSEPayloads(responseText)) {
    const content = payload?.choices?.[0]?.delta?.content
    if (typeof content === 'string' && content) {
      texts.push(content)
      continue
    }

    if (Array.isArray(content)) {
      for (const part of content) {
        if (typeof part?.text === 'string' && part.text) {
          texts.push(part.text)
        }
      }
    }
  }

  return texts.join('').trim()
}

function extractOpenAIResponsesSSEText(responseText: string): string {
  const texts: string[] = []

  for (const payload of extractSSEPayloads(responseText)) {
    if (typeof payload?.delta === 'string' && payload.delta) {
      texts.push(payload.delta)
      continue
    }

    if (typeof payload?.text === 'string' && payload.text) {
      texts.push(payload.text)
      continue
    }

    const nestedText = extractOpenAIResponsesText(payload?.response ?? payload)
    if (nestedText) {
      texts.push(nestedText)
    }
  }

  return texts.join('').trim()
}

function extractGatewaySSEPreview(variantId: GatewayVariantId, responseText: string): string {
  switch (variantId) {
    case 'anthropicMessages':
    case 'antigravityMessages':
      return extractAnthropicSSEText(responseText)
    case 'openaiChat':
      return extractOpenAIChatSSEText(responseText)
    case 'openaiResponses':
      return extractOpenAIResponsesSSEText(responseText)
    case 'geminiNative':
    case 'antigravityGemini':
      return ''
  }
}

export function extractGatewayResponsePreview(variantId: GatewayVariantId, responseText: string): string {
  if (!responseText.trim()) return ''

  const ssePreview = extractGatewaySSEPreview(variantId, responseText)
  if (ssePreview) {
    return truncateText(ssePreview)
  }

  try {
    const parsed = JSON.parse(responseText)
    const errorMessage =
      parsed?.error?.message ??
      parsed?.error?.error?.message ??
      parsed?.message
    if (typeof errorMessage === 'string' && errorMessage.trim()) {
      return truncateText(errorMessage.trim())
    }

    switch (variantId) {
      case 'anthropicMessages':
      case 'antigravityMessages': {
        const text = extractAnthropicText(parsed)
        return text ? truncateText(text) : truncateText(responseText)
      }
      case 'openaiChat': {
        const text = parsed?.choices?.[0]?.message?.content
        return typeof text === 'string' && text.trim() ? truncateText(text.trim()) : truncateText(responseText)
      }
      case 'openaiResponses': {
        const text = extractOpenAIResponsesText(parsed)
        return text ? truncateText(text) : truncateText(responseText)
      }
      case 'geminiNative':
      case 'antigravityGemini': {
        const text = extractGeminiText(parsed)
        return text ? truncateText(text) : truncateText(responseText)
      }
    }
  } catch {
    return truncateText(responseText)
  }
}
