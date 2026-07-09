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
  it('redirects the legacy runtime settings path into the unified settings center', async () => {
    const { default: router } = await import('@/router')

    const primary = router.getRoutes().find((record) => record.path === '/admin/runtime-settings')
    expect(typeof primary?.redirect).toBe('function')
    expect(typeof primary?.redirect === 'function' ? primary.redirect({} as never) : primary?.redirect).toEqual({
      path: '/admin/settings',
      query: { tab: 'runtime' },
    })
    expect(primary?.meta.title).toBe('Runtime Settings')
    expect(primary?.aliasOf).toBeUndefined()

    const workers = router.getRoutes().find((record) => record.path === '/admin/workers')
    expect(workers?.redirect).toBeUndefined()
    expect(workers?.name).toBe('AdminWorkers')
    expect(workers?.meta.title).toBe('Worker Management')
    expect(workers?.meta.titleKey).toBe('nav.workers')
    expect(workers?.aliasOf).toBeUndefined()

    const legacyAlias = router.getRoutes().find((record) => record.path === '/admin/touch/settings')
    expect(legacyAlias).toBeUndefined()
  })

  it('uses configured admin settings path for legacy runtime and worker redirects', async () => {
    appStore.cachedPublicSettings = {
      auth_shell_config: JSON.stringify({
        en: {
          defaults: {
            adminSettingsPath: '/configured-admin/settings',
          },
        },
      }),
    }
    const { default: router } = await import('@/router')

    const primary = router.getRoutes().find((record) => record.path === '/admin/runtime-settings')
    expect(typeof primary?.redirect === 'function' ? primary.redirect({} as never) : primary?.redirect).toEqual({
      path: '/configured-admin/settings',
      query: { tab: 'runtime' },
    })

    const workers = router.getRoutes().find((record) => record.path === '/admin/workers')
    expect(workers?.redirect).toBeUndefined()
    expect(workers?.name).toBe('AdminWorkers')
  })
})
