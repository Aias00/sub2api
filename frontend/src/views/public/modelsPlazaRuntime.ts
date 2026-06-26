import type { ModelPlazaItem } from '@/types'
import type { ModelsPlazaCopy } from '@/utils/modelPlazaDisplay'
import {
  MODEL_PLAZA_ALL_GROUP_KEY,
  resolveModelPlazaProviderGroupKey,
  resolveModelPlazaProviderGroupLabel,
  resolveModelPlazaProviderGroupRank,
} from '@/utils/modelPlazaDisplay'

export type ModelsPlazaGroupOption = {
  key: string
  count: number
  label: string
  rank: number
}

export function resolveVisibleModelsPlazaItems(items: ModelPlazaItem[]) {
  return items
    .filter((item) => item.visible !== false)
    .slice()
    .sort((a, b) => (a.sort_order || 0) - (b.sort_order || 0))
}

export function matchesModelsPlazaSearch(item: ModelPlazaItem, query: string): boolean {
  const normalized = query.trim().toLowerCase()
  if (!normalized) return true

  const haystack = [
    item.title,
    item.provider,
    item.badge,
    item.description,
    item.input_price,
    item.output_price,
    item.cache_read_price,
    item.cache_write_price,
    item.billing_badge,
    ...item.capability_tags,
    ...item.model_ids,
  ]
    .join('\n')
    .toLowerCase()

  return haystack.includes(normalized)
}

export function resolveModelsPlazaGroupOptions(
  items: ModelPlazaItem[],
  copy: Pick<ModelsPlazaCopy, 'groupAll' | 'groupOther'>,
): ModelsPlazaGroupOption[] {
  const counts = new Map<string, number>()

  for (const item of items) {
    const key = resolveModelPlazaProviderGroupKey(item.provider)
    counts.set(key, (counts.get(key) || 0) + 1)
  }

  const groups = Array.from(counts.entries())
    .map(([key, count]) => ({
      key,
      count,
      label: resolveModelPlazaProviderGroupLabel(key, copy),
      rank: resolveModelPlazaProviderGroupRank(key),
    }))
    .sort((a, b) => a.rank - b.rank || a.label.localeCompare(b.label))

  return [
    {
      key: MODEL_PLAZA_ALL_GROUP_KEY,
      label: copy.groupAll,
      count: items.length,
      rank: -1,
    },
    ...groups,
  ]
}

export function resolveModelsPlazaActiveGroupLabel(
  groups: ModelsPlazaGroupOption[],
  activeGroup: string,
  copy: Pick<ModelsPlazaCopy, 'groupAll'>,
) {
  const match = groups.find((group) => group.key === activeGroup)
  return match?.label || copy.groupAll
}

export function filterModelsPlazaItems(
  items: ModelPlazaItem[],
  activeGroup: string,
  query: string,
) {
  return items.filter((item) => {
    if (
      activeGroup !== MODEL_PLAZA_ALL_GROUP_KEY &&
      resolveModelPlazaProviderGroupKey(item.provider) !== activeGroup
    ) {
      return false
    }
    return matchesModelsPlazaSearch(item, query)
  })
}
