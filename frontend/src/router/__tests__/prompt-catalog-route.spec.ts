import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it, vi } from 'vitest'

const routerSource = readFileSync(resolve(process.cwd(), 'src/router/index.ts'), 'utf8')

const authStore = vi.hoisted(() => ({
  checkAuth: vi.fn(),
  isAuthenticated: true,
  isAdmin: true,
  isSimpleMode: false,
  hasPendingAuthSession: false,
}))

const appStore = vi.hoisted(() => ({
  siteName: 'Sub2API',
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

describe('prompt catalog route registration', () => {
  it('registers a public /prompts route backed by PromptCatalogView', () => {
    expect(routerSource).toContain("path: '/prompts'")
    expect(routerSource).toContain("name: 'PromptCatalog'")
    expect(routerSource).toContain("component: () => import('@/views/public/PromptCatalogView.vue')")
    expect(routerSource).toContain("titleKey: 'nav.promptCatalog'")
  })

  it('marks /prompts as public (requiresAuth: false)', () => {
    // Static analysis: the route block for /prompts must contain requiresAuth: false
    const promptsBlockStart = routerSource.indexOf("path: '/prompts'")
    expect(promptsBlockStart).toBeGreaterThan(-1)

    const nextRouteStart = routerSource.indexOf("path: '/", promptsBlockStart + 1)
    const block = routerSource.slice(promptsBlockStart, nextRouteStart)
    expect(block).toContain('requiresAuth: false')
  })

  it('includes /prompts in the backend-mode public route allowlist', () => {
    expect(routerSource).toContain("'/prompts'")
  })

  it('is registered at runtime with correct meta', async () => {
    const { default: router } = await import('@/router')
    const route = router.getRoutes().find((r) => r.path === '/prompts')

    expect(route).toBeDefined()
    expect(route?.meta.requiresAuth).toBe(false)
    expect(route?.name).toBe('PromptCatalog')
  })
})
