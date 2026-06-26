import { resolveShellLabelOverrides } from './shellLabelOverrides'
import { DEFAULT_GATEWAY_TEST_MAX_TOKENS, DEFAULT_GATEWAY_TEST_PROMPT } from './gatewayDocs'

export const apiTestLabelKeys = [
  'badge',
  'title',
  'description',
  'openGuide',
  'send',
  'sending',
  'keySelector',
  'noSelection',
  'noGroupAssigned',
  'protocol',
  'model',
  'loading',
  'noOptionsFound',
  'stream',
  'requestMeta',
  'noKeysTitle',
  'noKeysDescription',
  'manageKeys',
  'modelPlaceholder',
  'modelSearchPlaceholder',
  'modelHint',
  'customModel',
  'customModelHint',
  'customModelOption',
  'customModelOptionHint',
  'prompt',
  'promptHint',
  'promptPlaceholder',
  'streamHint',
  'unassignedTitle',
  'unassignedDescription',
  'liveBillingTitle',
  'liveBillingDescription',
  'copyCurl',
  'platform',
  'headerMode',
  'requestPreview',
  'copyRequest',
  'responsePreview',
  'statusCode',
  'duration',
  'copyResponse',
  'responseSummary',
  'usageRecordTitle',
  'openUsage',
  'rawResponse',
  'responsePending',
  'notReady',
  'copyCurlSuccess',
  'copyRequestSuccess',
  'copyResponseSuccess',
  'usageRecordSyncing',
  'usageRecordFound',
  'usageRecordPending',
  'usageRecordIdle',
  'loadKeysFailed',
  'unknownError',
] as const

export type APITestLabelKey = typeof apiTestLabelKeys[number]
export type APITestShellLabels = Partial<Record<APITestLabelKey, string>>

export type APITestShellDefaults = {
  guidePath: string
  defaultPrompt: string
  maxTokens: number
  apiKeyPageSize: number
  usageSyncPageSize: number
}

export const DEFAULT_API_TEST_GUIDE_PATH = '/gateway-guide'
export const DEFAULT_API_TEST_MAX_TOKENS = DEFAULT_GATEWAY_TEST_MAX_TOKENS
export const DEFAULT_API_TEST_API_KEY_PAGE_SIZE = 100
export const DEFAULT_API_TEST_USAGE_SYNC_PAGE_SIZE = 10

export function resolveAPITestShellLabels(raw: string | undefined, runtimeLocale: string): APITestShellLabels {
  return resolveShellLabelOverrides(raw, runtimeLocale, apiTestLabelKeys)
}

export function resolveAPITestShellDefaults(raw: string | undefined, runtimeLocale: string): APITestShellDefaults {
  const config = pickLocalizedShellConfig(raw, runtimeLocale)
  const defaults = isRecord(config?.defaults) ? config.defaults : null
  const guidePath = defaults ? readInternalPath(defaults.guidePath) : ''
  const defaultPrompt = defaults ? readNonEmptyString(defaults.defaultPrompt) : ''
  const maxTokens = defaults ? readPositiveInteger(defaults.maxTokens) : undefined
  const apiKeyPageSize = defaults ? readPositiveInteger(defaults.apiKeyPageSize, 1000) : undefined
  const usageSyncPageSize = defaults ? readPositiveInteger(defaults.usageSyncPageSize, 1000) : undefined
  return {
    guidePath: guidePath || DEFAULT_API_TEST_GUIDE_PATH,
    defaultPrompt: defaultPrompt || DEFAULT_GATEWAY_TEST_PROMPT,
    maxTokens: maxTokens || DEFAULT_API_TEST_MAX_TOKENS,
    apiKeyPageSize: apiKeyPageSize || DEFAULT_API_TEST_API_KEY_PAGE_SIZE,
    usageSyncPageSize: usageSyncPageSize || DEFAULT_API_TEST_USAGE_SYNC_PAGE_SIZE,
  }
}

export function renderAPITestShellText(
  labels: APITestShellLabels,
  key: APITestLabelKey,
  values?: Record<string, string | number>,
): string {
  let text = labels[key] || ''
  if (!values) return text
  for (const [name, value] of Object.entries(values)) {
    text = text.split(`{${name}}`).join(String(value))
  }
  return text
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

function readPositiveInteger(value: unknown, max?: number): number | undefined {
  if (typeof value !== 'number' || !Number.isInteger(value) || value < 1) {
    return undefined
  }
  if (max && value > max) return undefined
  return value
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}
