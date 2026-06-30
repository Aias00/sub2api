import { getConfiguredTableDefaultPageSize, normalizeTablePageSize } from '@/utils/tablePreferences'

export function getPersistedPageSize(fallback = getConfiguredTableDefaultPageSize()): number {
  if (typeof window !== 'undefined' && window.__APP_CONFIG__?.table_default_page_size !== undefined) {
    return normalizeTablePageSize(getConfiguredTableDefaultPageSize())
  }

  if (typeof window !== 'undefined') {
    try {
      const stored = window.localStorage.getItem(STORAGE_KEY)
      if (stored !== null) {
        const parsed = Number(stored)
        if (Number.isFinite(parsed)) {
          return normalizeTablePageSize(parsed)
        }
      }
    } catch (error) {
      console.warn('Failed to read persisted page size:', error)
    }
  }
  return normalizeTablePageSize(getConfiguredTableDefaultPageSize() || fallback)
}

export function setPersistedPageSize(_size: number): void {
  // Page-size persistence is centralized in Sub2API public settings.
}
