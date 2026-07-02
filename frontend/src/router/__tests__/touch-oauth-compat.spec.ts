import { describe, expect, it, vi } from 'vitest'

const authStore = vi.hoisted(() => ({
  checkAuth: vi.fn(),
  isAuthenticated: false,
  isAdmin: false,
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

describe('legacy OAuth compatibility routes', () => {
  it('registers legacy callback aliases on the Cloudbase OAuth callback route', async () => {
    const { default: router } = await import('@/router')

    for (const path of ['/auth-callback', '/en/auth-callback', '/zh/auth-callback']) {
      const route = router.getRoutes().find((record) => record.path === path)
      expect(route?.meta.requiresAuth).toBe(false)
      expect(route?.meta.title).toBe('OAuth Callback')
    }
  })

  it('registers legacy popup aliases as public routes', async () => {
    const { default: router } = await import('@/router')

    for (const path of ['/auth-popup', '/en/auth-popup', '/zh/auth-popup']) {
      const route = router.getRoutes().find((record) => record.path === path)
      expect(route?.meta.requiresAuth).toBe(false)
      expect(route?.meta.title).toBe('Login')
      expect(route?.name ?? route?.aliasOf?.name).toBe('AuthPopup')
    }
  })
})
