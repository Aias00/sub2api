import { describe, expect, it, vi } from 'vitest'

const authStore = vi.hoisted(() => ({
  checkAuth: vi.fn(),
  isAuthenticated: true,
  isAdmin: true,
  isSimpleMode: false,
  hasPendingAuthSession: false,
}))

const appStore = vi.hoisted(() => ({
  siteName: 'Cloudbase',
  backendModeEnabled: false,
  cachedPublicSettings: null as null | Record<string, unknown>,
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => authStore,
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => appStore,
}))

vi.mock('@/stores/adminSettings', () => ({
  useAdminSettingsStore: () => ({
    customMenuItems: [],
  }),
}))

vi.mock('@/composables/useNavigationLoading', () => ({
  useNavigationLoadingState: () => ({
    startNavigation: vi.fn(),
    endNavigation: vi.fn(),
    isLoading: { value: false },
  }),
}))

vi.mock('@/composables/useRoutePrefetch', () => ({
  useRoutePrefetch: () => ({
    triggerPrefetch: vi.fn(),
    cancelPendingPrefetch: vi.fn(),
    resetPrefetchState: vi.fn(),
  }),
}))

describe('runtime settings route', () => {
  it('uses only the generic primary admin path in the Vue router', async () => {
    const { default: router } = await import('@/router')

    const primary = router.getRoutes().find((record) => record.path === '/admin/runtime-settings')
    expect(primary?.name).toBe('AdminRuntimeSettings')
    expect(primary?.meta.title).toBe('Runtime Settings')
    expect(primary?.aliasOf).toBeUndefined()

    const legacyAlias = router.getRoutes().find((record) => record.path === '/admin/touch/settings')
    expect(legacyAlias).toBeUndefined()
  })
})
