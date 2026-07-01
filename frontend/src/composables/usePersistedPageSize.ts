import { getConfiguredTableDefaultPageSize, normalizeTablePageSize } from '@/utils/tablePreferences'

export function getPersistedPageSize(fallback = getConfiguredTableDefaultPageSize()): number {
  return normalizeTablePageSize(getConfiguredTableDefaultPageSize() || fallback)
}

export function setPersistedPageSize(_size: number): void {
  // Page-size persistence is centralized in Sub2API public settings.
}
