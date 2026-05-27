import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import HomeView from '../HomeView.vue'

const paymentCatalog = vi.hoisted(() => vi.fn())

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, string | number>) => {
        if (params?.count !== undefined) return `${key}:${params.count}`
        if (params?.amount !== undefined) return `${key}:${params.amount}`
        return key
      },
      te: () => false,
    }),
  }
})

vi.mock('@/api/payment', () => ({
  paymentAPI: {
    getPublicCatalog: paymentCatalog,
  },
}))

const authStoreState = vi.hoisted(() => ({
  isAuthenticated: false,
  isAdmin: false,
  user: null as null | { email: string },
  checkAuth: vi.fn(),
}))

const appStoreState = vi.hoisted(() => ({
  publicSettingsLoaded: true,
  siteName: 'Sub2API',
  siteLogo: '',
  docUrl: '',
  cachedPublicSettings: {
    site_name: 'cloudbase',
    site_logo: '',
    site_subtitle: 'AI Coding Workspace',
    doc_url: '/docs',
    home_content: '',
  },
  fetchPublicSettings: vi.fn(),
}))

vi.mock('@/stores', () => ({
  useAuthStore: () => authStoreState,
  useAppStore: () => appStoreState,
}))

describe('HomeView', () => {
  beforeEach(() => {
    authStoreState.isAuthenticated = false
    authStoreState.isAdmin = false
    authStoreState.user = null
    authStoreState.checkAuth.mockReset()
    appStoreState.fetchPublicSettings.mockReset()
    paymentCatalog.mockReset().mockResolvedValue({
      data: {
        recharge_products: [
          {
            id: 'starter',
            name: 'Starter Pack',
            description: 'Quick trial',
            amount: 30,
            credited_amount: 45,
            badge: 'HOT',
            recommended: true,
            features: ['45 credits'],
            sort_order: 1,
          },
        ],
        plans: [
          {
            id: 1,
            group_id: 11,
            group_platform: 'anthropic',
            group_name: 'Claude',
            supported_model_scopes: ['Claude Opus 4.6'],
            name: 'Claude Pro',
            description: 'Reasoning heavy',
            price: 59,
            validity_days: 30,
            validity_unit: 'day',
            features: ['priority'],
            for_sale: true,
            sort_order: 1,
          },
          {
            id: 2,
            group_id: 12,
            group_platform: 'openai',
            group_name: 'GPT',
            supported_model_scopes: ['GPT-5.4'],
            name: 'GPT Builder',
            description: 'Feature delivery',
            price: 49,
            validity_days: 30,
            validity_unit: 'day',
            features: ['coding'],
            for_sale: true,
            sort_order: 2,
          },
        ],
      },
    })
  })

  it('renders the landing page with model matrix capability tags but no homepage pricing content', async () => {
    const wrapper = mount(HomeView, {
      global: {
        stubs: {
          RouterLink: {
            props: ['to'],
            template: '<a :href="typeof to === \'string\' ? to : to.path"><slot /></a>',
          },
          DocsLink: { template: '<a><slot /></a>' },
          LocaleSwitcher: { template: '<div>locale</div>' },
          Icon: { template: '<i />' },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('home.heroTitle')
    expect(wrapper.text()).toContain('home.secondaryCta')
    expect(wrapper.text()).toContain('home.modelMatrixTitle')
    expect(wrapper.text()).toContain('home.familyCapabilities.claude.reasoning')
    expect(wrapper.text()).toContain('home.familyCapabilities.gpt.agents')
    expect(wrapper.text()).toContain('home.experienceTitle')
    expect(wrapper.text()).toContain('home.whyChooseTitle')
    expect(wrapper.text()).not.toContain('Claude Opus 4.6')
    expect(wrapper.text()).not.toContain('GPT-5.4')
    expect(wrapper.text()).not.toContain('Starter Pack')
    expect(wrapper.text()).not.toContain('Claude Pro')
    expect(wrapper.text()).not.toContain('Gemini')
    expect(wrapper.get('[data-home-model-grid]').classes()).toContain('md:grid-cols-2')
    expect(wrapper.get('[data-home-model-grid]').classes()).toContain('max-w-4xl')
    expect(paymentCatalog).toHaveBeenCalledTimes(1)
  })
})
