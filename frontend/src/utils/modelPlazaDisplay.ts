import { resolveLocalizedShellLabels } from './localizedShell'

export const MODEL_PLAZA_ALL_GROUP_KEY = 'all'
export const MODEL_PLAZA_OTHER_GROUP_KEY = 'other'
export const MODEL_PLAZA_UNKNOWN_PROVIDER_INITIAL = 'M'

export type ModelsPlazaCopy = {
  viewDocs: string
  dashboard: string
  login: string
  badge: string
  title: string
  description: string
  emptyTitle: string
  emptyDescription: string
  quickFind: string
  searchLabel: string
  searchPlaceholder: string
  groupsTitle: string
  currentSearch: string
  browseHint: string
  results: string
  emptyFilteredTitle: string
  emptyFilteredDescription: string
  copyModelIds: string
  modelIdsCopied: string
  inputPrice: string
  outputPrice: string
  cacheReadPrice: string
  cacheWritePrice: string
  modelIdsConfigured: string
  groupAll: string
  groupOther: string
}

const modelPlazaCopyKeys = [
  'viewDocs',
  'dashboard',
  'login',
  'badge',
  'title',
  'description',
  'emptyTitle',
  'emptyDescription',
  'quickFind',
  'searchLabel',
  'searchPlaceholder',
  'groupsTitle',
  'currentSearch',
  'browseHint',
  'results',
  'emptyFilteredTitle',
  'emptyFilteredDescription',
  'copyModelIds',
  'modelIdsCopied',
  'inputPrice',
  'outputPrice',
  'cacheReadPrice',
  'cacheWritePrice',
  'modelIdsConfigured',
  'groupAll',
  'groupOther',
] as const satisfies readonly (keyof ModelsPlazaCopy)[]

export function resolveModelsPlazaCopy(raw: string | undefined, runtimeLocale: string): ModelsPlazaCopy {
  return resolveLocalizedShellLabels(raw, runtimeLocale, modelPlazaCopyKeys)
}

export function resolveModelPlazaProviderGroupKey(provider: string): string {
  const normalized = provider.trim().toLowerCase()
  if (normalized.includes('anthropic') || normalized.includes('claude')) return 'claude'
  if (normalized.includes('openai') || normalized.includes('gpt')) return 'gpt'
  if (normalized.includes('gemini') || normalized.includes('google')) return 'gemini'
  return normalized || MODEL_PLAZA_OTHER_GROUP_KEY
}

export function resolveModelPlazaProviderGroupRank(groupKey: string): number {
  if (groupKey === 'claude') return 0
  if (groupKey === 'gpt') return 1
  if (groupKey === 'gemini') return 2
  if (groupKey === MODEL_PLAZA_OTHER_GROUP_KEY) return 99
  return 50
}

export function resolveModelPlazaProviderGroupLabel(groupKey: string, copy: Pick<ModelsPlazaCopy, 'groupOther'>): string {
  if (groupKey === 'claude') return 'Claude'
  if (groupKey === 'gpt') return 'GPT'
  if (groupKey === 'gemini') return 'Gemini'
  if (groupKey === MODEL_PLAZA_OTHER_GROUP_KEY) return copy.groupOther
  return groupKey.toUpperCase()
}

export function resolveModelPlazaProviderInitial(provider: string): string {
  if (!provider.trim()) return MODEL_PLAZA_UNKNOWN_PROVIDER_INITIAL

  const groupKey = resolveModelPlazaProviderGroupKey(provider)
  if (groupKey === 'claude') return 'C'
  if (groupKey === 'gpt') return 'G'
  if (groupKey === 'gemini') return 'G'
  return groupKey.slice(0, 1).toUpperCase() || MODEL_PLAZA_UNKNOWN_PROVIDER_INITIAL
}

export function resolveModelPlazaProviderIconClass(provider: string): string {
  const groupKey = resolveModelPlazaProviderGroupKey(provider)
  if (groupKey === 'claude') {
    return 'bg-[linear-gradient(135deg,#ef8e67,#d2745c)]'
  }
  if (groupKey === 'gpt') {
    return 'bg-[linear-gradient(135deg,#48b774,#2f9360)]'
  }
  if (groupKey === 'gemini') {
    return 'bg-[linear-gradient(135deg,#5b7cff,#4a5ce4)]'
  }
  return 'bg-[linear-gradient(135deg,#64748b,#475569)]'
}

export function formatModelsPlazaTemplate(template: string, values: Record<string, string>): string {
  return template.replace(/\{(\w+)\}/g, (_match, key: string) => values[key] ?? '')
}
