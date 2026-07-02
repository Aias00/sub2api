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

describe('legacy Touch Vue route aliases', () => {
  it('keeps generic frontend routes as the only Vue Router entries', async () => {
    const { default: router } = await import('@/router')
    const paths = router.getRoutes().map((route) => route.path)

    expect(paths).toContain('/image-generator')
    expect(paths).toContain('/prompts')
    expect(paths).toContain('/wechat')
    expect(paths).toContain('/wechat-export')
    expect(paths).toContain('/pricing')
    expect(paths).toContain('/settings/credits')
    expect(paths).toContain('/admin/runtime-settings')

    expect(paths).not.toContain('/touch/prompts')
    expect(paths).not.toContain('/touch/generator')
    expect(paths).not.toContain('/touch/pricing')
    expect(paths).not.toContain('/touch/credits')
    expect(paths).not.toContain('/admin/touch/settings')
  })
})
