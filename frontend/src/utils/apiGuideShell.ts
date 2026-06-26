import { resolveShellLabelOverrides } from './shellLabelOverrides'

export const apiGuideLabelKeys = [
  'badge',
  'title',
  'description',
  'openTester',
  'manageKeys',
  'baseUrl',
  'currentKey',
  'noSelection',
  'selectKeyHint',
  'supportedEndpoints',
  'noGroupAssigned',
  'noKeysTitle',
  'noKeysDescription',
  'keySelector',
  'keySelectorHint',
  'unassignedTitle',
  'unassignedDescription',
  'keySummary',
  'groupName',
  'platform',
  'status',
  'authHeaderTitle',
  'authHeaderDescription',
  'noEndpointVariants',
  'stream',
  'testThisVariant',
  'endpoint',
  'protocol',
  'defaultModel',
  'headerMode',
  'curlExample',
  'copyCurl',
  'copyCurlSuccess',
  'defaultPrompt',
  'loadKeysFailed',
] as const

export type APIGuideLabelKey = typeof apiGuideLabelKeys[number]
export type APIGuideShellLabels = Partial<Record<APIGuideLabelKey, string>>

export type APIGuideShellDefaults = {
  testPath: string
  defaultPrompt: string
  maxTokens: number
  apiKeyPageSize: number
}

export const DEFAULT_API_GUIDE_TEST_PATH = '/gateway-test'
export const DEFAULT_API_GUIDE_PROMPT = '请简短介绍一下你当前命中的模型和主要能力。'
export const DEFAULT_API_GUIDE_MAX_TOKENS = 256
export const DEFAULT_API_GUIDE_API_KEY_PAGE_SIZE = 100

export function resolveAPIGuideShellLabels(raw: string | undefined, runtimeLocale: string): APIGuideShellLabels {
  return resolveShellLabelOverrides(raw, runtimeLocale, apiGuideLabelKeys)
}

export function resolveAPIGuideShellDefaults(raw: string | undefined, runtimeLocale: string): APIGuideShellDefaults {
  const config = pickLocalizedShellConfig(raw, runtimeLocale)
  const defaults = isRecord(config?.defaults) ? config.defaults : null
  const testPath = defaults ? readInternalPath(defaults.testPath) : ''
  const defaultPrompt = defaults ? readNonEmptyString(defaults.defaultPrompt) : ''
  const maxTokens = defaults ? readPositiveInteger(defaults.maxTokens) : 0
  const apiKeyPageSize = defaults ? readPositiveInteger(defaults.apiKeyPageSize, 1000) : 0
  return {
    testPath: testPath || DEFAULT_API_GUIDE_TEST_PATH,
    defaultPrompt: defaultPrompt || DEFAULT_API_GUIDE_PROMPT,
    maxTokens: maxTokens || DEFAULT_API_GUIDE_MAX_TOKENS,
    apiKeyPageSize: apiKeyPageSize || DEFAULT_API_GUIDE_API_KEY_PAGE_SIZE,
  }
}

export function renderAPIGuideShellText(labels: APIGuideShellLabels, key: APIGuideLabelKey): string {
  return labels[key] || ''
}

function pickLocalizedShellConfig(raw: string | undefined, runtimeLocale: string): Record<string, unknown> | null {
  if (!raw?.trim()) return null
  try {
    const parsed = JSON.parse(raw) as unknown
    if (!isRecord(parsed)) return null
    const normalizedLocale = runtimeLocale.toLowerCase()
    const baseLocale = normalizedLocale.split('-')[0]
    const localized = parsed[normalizedLocale] ?? parsed[baseLocale] ?? parsed.en ?? parsed.zh ?? parsed
    return isRecord(localized) ? localized : null
  } catch {
    return null
  }
}

function readInternalPath(value: unknown): string {
  if (typeof value !== 'string') return ''
  const path = value.trim()
  if (!path || !path.startsWith('/') || path.startsWith('//')) return ''
  if (path.includes('://') || path.includes('\n') || path.includes('\r')) return ''
  return path
}

function readNonEmptyString(value: unknown): string {
  if (typeof value !== 'string') return ''
  return value.trim()
}

function readPositiveInteger(value: unknown, max?: number): number {
  const normalized = Number(value)
  if (!Number.isInteger(normalized) || normalized <= 0) return 0
  if (max && normalized > max) return 0
  return normalized
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}
