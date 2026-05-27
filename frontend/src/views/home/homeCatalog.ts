import type { HomeCatalogResponse } from '@/types/payment'

export interface HomeModelFamily {
  key: 'claude' | 'gpt'
  name: string
  models: string[]
  priceHint?: string
}

const PLATFORM_TO_FAMILY: Record<string, HomeModelFamily['key'] | ''> = {
  anthropic: 'claude',
  openai: 'gpt',
}

const FAMILY_ORDER: HomeModelFamily['key'][] = ['claude', 'gpt']

const FAMILY_DISPLAY_NAME: Record<HomeModelFamily['key'], string> = {
  claude: 'Claude',
  gpt: 'GPT',
}

export function buildHomeModelFamilies(catalog: HomeCatalogResponse): HomeModelFamily[] {
  const buckets = new Map<HomeModelFamily['key'], Set<string>>()

  for (const plan of catalog.plans || []) {
    const family = PLATFORM_TO_FAMILY[(plan.group_platform || '').trim().toLowerCase()]
    if (!family) continue
    const set = buckets.get(family) ?? new Set<string>()
    for (const model of plan.supported_model_scopes || []) {
      const normalized = model.trim()
      if (normalized) set.add(normalized)
    }
    if (set.size === 0 && plan.group_name?.trim()) {
      set.add(plan.group_name.trim())
    }
    buckets.set(family, set)
  }

  return FAMILY_ORDER
    .filter((family) => (buckets.get(family)?.size ?? 0) > 0)
    .map((family) => ({
      key: family,
      name: FAMILY_DISPLAY_NAME[family],
      models: Array.from(buckets.get(family) ?? []).slice(0, 3),
    }))
}
