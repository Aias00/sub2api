import { ref } from 'vue'
import accountsAPI, { type UpstreamPreset } from '@/api/admin/accounts'

/**
 * Resolved form values derived from a vendor preset, ready to be applied to the
 * create-account form. Kept as a pure function so it is trivially unit-testable
 * independent of the (very large) CreateAccountModal component.
 */
export interface UpstreamPresetApply {
  platform: string
  accountType: string
  baseUrl: string
  models: string[]
  apiStyle: string
}

/**
 * Pure transform: map a vendor preset to the fields the create-account form
 * needs. Returns a fresh models array so callers never mutate the shared preset.
 */
export function resolveUpstreamPresetApply(preset: UpstreamPreset): UpstreamPresetApply {
  return {
    platform: preset.platform,
    accountType: preset.account_type,
    baseUrl: preset.base_url,
    models: Array.isArray(preset.default_models) ? [...preset.default_models] : [],
    apiStyle: preset.api_style
  }
}

// Module-level cache so the catalog is fetched once across modal opens.
const presets = ref<UpstreamPreset[]>([])
const loading = ref(false)
const loaded = ref(false)
let pendingLoad: Promise<UpstreamPreset[]> | null = null

/**
 * useUpstreamPresets loads and exposes the curated upstream vendor catalog.
 */
export function useUpstreamPresets() {
  async function load(force = false): Promise<UpstreamPreset[]> {
    if (loaded.value && !force) {
      return presets.value
    }
    if (pendingLoad && !force) {
      return pendingLoad
    }
    loading.value = true
    pendingLoad = accountsAPI
      .listUpstreamPresets()
      .then((items) => {
        presets.value = items
        loaded.value = true
        return items
      })
      .catch(() => {
        presets.value = []
        loaded.value = true
        return []
      })
      .finally(() => {
        loading.value = false
        pendingLoad = null
      })
    return pendingLoad
  }

  function findById(id: string): UpstreamPreset | undefined {
    return presets.value.find((p) => p.id === id)
  }

  return {
    presets,
    loading,
    loaded,
    load,
    findById,
    resolveApply: resolveUpstreamPresetApply
  }
}

// Test-only helper to reset module cache between unit tests.
export function __resetUpstreamPresetsCacheForTest() {
  presets.value = []
  loading.value = false
  loaded.value = false
  pendingLoad = null
}
