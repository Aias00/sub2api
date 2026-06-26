import { readFileSync } from 'node:fs'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import HomeView from '../HomeView.vue'

const homeViewSource = readFileSync('src/views/HomeView.vue', 'utf8')
const paymentCatalog = vi.hoisted(() => vi.fn())
const currentRoute = vi.hoisted(() => ({
  path: '/sub',
  name: 'SubHome',
}))

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
      locale: { value: 'en' },
      te: () => false,
    }),
  }
})

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => currentRoute,
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
    home_shell_config: JSON.stringify({
      en: {
        labels: {
          navHome: 'Configured home',
          viewDocs: 'Configured docs',
          dashboard: 'Configured dashboard',
          login: 'Configured login',
          heroBadge: 'Configured badge',
          heroTitle: 'Configured hero',
          heroDescription: 'Configured hero description',
          secondaryCta: 'Configured models',
          modelMatrixTitle: 'Configured matrix',
          familyClaudeReasoning: 'Configured reasoning',
          familyGptAgents: 'Configured agents',
          experienceTitle: 'Configured experience',
          whyChooseTitle: 'Configured why',
          footerDescription: 'Configured footer',
        },
        experienceCards: [
          {
            key: 'unified',
            title: 'Configured unified card',
            description: 'Configured unified description',
          },
        ],
        whyChooseCards: [
          {
            key: 'lowFriction',
            title: 'Configured low friction',
            description: 'Configured low friction description',
          },
        ],
      },
    }),
    home_business_shell_config: JSON.stringify({
      en: {
        labels: {
          heroBadge: 'Configured business badge',
          heroTitle: 'Configured business hero',
          heroDescription: 'Configured business description',
          primaryCta: 'Configured business primary',
          secondaryCta: 'Configured business secondary',
        },
        businessCards: [
          {
            key: 'wechat-export',
            title: 'Configured WeChat Export',
            description: 'Configured business card description',
            capabilityTags: ['Configured capability'],
            path: '/prompts',
            pathLabel: 'Configured card CTA',
          },
        ],
      },
    }),
    auth_shell_config: JSON.stringify({
      en: {
        defaults: {
          loginPath: '/configured-login',
          defaultRedirectPath: '/configured-dashboard',
          adminRedirectPath: '/configured-admin',
        },
      },
    }),
  },
  fetchPublicSettings: vi.fn(),
}))

vi.mock('@/stores', () => ({
  useAuthStore: () => authStoreState,
  useAppStore: () => appStoreState,
}))

describe('HomeView', () => {
  beforeEach(() => {
    currentRoute.path = '/sub'
    currentRoute.name = 'SubHome'
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

    expect(wrapper.text()).toContain('Configured home')
    expect(wrapper.text()).toContain('Configured docs')
    expect(wrapper.text()).toContain('Configured login')
    expect(wrapper.text()).toContain('Configured badge')
    expect(wrapper.text()).toContain('Configured hero')
    expect(wrapper.text()).toContain('Configured hero description')
    expect(wrapper.text()).toContain('Configured models')
    expect(wrapper.text()).toContain('Configured matrix')
    expect(wrapper.text()).toContain('Configured reasoning')
    expect(wrapper.text()).toContain('Configured agents')
    expect(wrapper.text()).toContain('Configured experience')
    expect(wrapper.text()).toContain('Configured why')
    expect(wrapper.text()).toContain('Configured footer')
    expect(wrapper.text()).toContain('Configured unified card')
    expect(wrapper.text()).toContain('Configured unified description')
    expect(wrapper.text()).toContain('Configured low friction')
    expect(wrapper.text()).toContain('Configured low friction description')
    expect(wrapper.text()).not.toContain('Claude Opus 4.6')
    expect(wrapper.text()).not.toContain('GPT-5.4')
    expect(wrapper.text()).not.toContain('Starter Pack')
    expect(wrapper.text()).not.toContain('Claude Pro')
    expect(wrapper.text()).not.toContain('Gemini')
    expect(wrapper.get('[data-home-model-grid]').classes()).toContain('md:grid-cols-2')
    expect(wrapper.get('[data-home-model-grid]').classes()).toContain('max-w-4xl')
    expect(paymentCatalog).toHaveBeenCalledTimes(1)
  })

  it('renders /home with the business capability copy instead of the sub-home runtime copy', async () => {
    currentRoute.path = '/home'
    currentRoute.name = 'Home'

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

    expect(wrapper.text()).toContain('Configured business badge')
    expect(wrapper.text()).toContain('Configured business hero')
    expect(wrapper.text()).toContain('Configured business description')
    expect(wrapper.text()).toContain('Configured business primary')
    expect(wrapper.text()).toContain('Configured business secondary')
    expect(wrapper.text()).toContain('Configured WeChat Export')
    expect(wrapper.text()).toContain('Configured business card description')
    expect(wrapper.text()).toContain('Configured capability')
    expect(wrapper.text()).toContain('Configured card CTA')
    expect(wrapper.text()).not.toContain('Configured hero')
    expect(paymentCatalog).not.toHaveBeenCalled()
  })

  it('does not embed default home shell copy in the Vue view', () => {
    expect(homeViewSource).toContain('resolveBusinessHomeShellConfig')
    expect(homeViewSource).toContain("route.path === '/sub'")
    expect(homeViewSource).toContain('useAuthRouteDefaults')
    expect(homeViewSource).not.toContain("isAuthenticated ? dashboardPath : '/login'")
    expect(homeViewSource).not.toContain("isAdmin.value ? '/admin/dashboard' : '/dashboard'")
    expect(homeViewSource).not.toContain('const EMPTY_HOME_COPY')
    expect(homeViewSource).not.toContain('EMPTY_HOME_COPY')
    expect(homeViewSource).not.toContain('EMPTY_HOME_EXPERIENCE_CARDS')
    expect(homeViewSource).not.toContain('EMPTY_HOME_WHY_CHOOSE_CARDS')
    expect(homeViewSource).not.toContain('FALLBACK_HOME_COPY')
    expect(homeViewSource).not.toContain('FALLBACK_HOME_EXPERIENCE_CARDS')
    expect(homeViewSource).not.toContain('FALLBACK_HOME_WHY_CHOOSE_CARDS')
    expect(homeViewSource).not.toContain('mergeHomeExperienceCards')
    expect(homeViewSource).not.toContain('mergeHomeWhyChooseCards')
    expect(homeViewSource).not.toContain('defaultExperienceIcons')
    expect(homeViewSource).toContain('homeLinks.value.modelsPath')
    expect(homeViewSource).toContain('homeLinks.value.docsPath')
    expect(homeViewSource).toContain('homeLinks.value.termsPath')
    expect(homeViewSource).toContain('homeLinks.value.privacyPath')
    expect(homeViewSource).not.toContain('to="/models"')
    expect(homeViewSource).not.toContain("href: '/models'")
    expect(homeViewSource).not.toContain("href: '/docs'")
    expect(homeViewSource).not.toContain('href="/legal/terms"')
    expect(homeViewSource).not.toContain('href="/legal/privacy-policy"')
    expect(homeViewSource).not.toContain("href: '/legal/terms'")
    expect(homeViewSource).not.toContain("href: '/legal/privacy-policy'")
  })

  it('does not keep legacy home page copy in locale bundles', () => {
    const zhLocaleSource = readFileSync('src/i18n/locales/zh.ts', 'utf8')
    const enLocaleSource = readFileSync('src/i18n/locales/en.ts', 'utf8')
    for (const source of [zhLocaleSource, enLocaleSource]) {
      expect(source).not.toContain('\n  home: {')
      expect(source).not.toContain('AI 编码工作台')
      expect(source).not.toContain('AI Coding Workspace')
      expect(source).not.toContain('模型矩阵')
      expect(source).not.toContain('Model Matrix')
      expect(source).not.toContain('为什么选择我们')
      expect(source).not.toContain('Why Choose Us')
      expect(source).not.toContain('pricingKicker')
      expect(source).not.toContain('pricingTitle')
      expect(source).not.toContain('pricingDescription')
      expect(source).not.toContain('pricingUnavailable')
      expect(source).not.toContain('pricingEmptyPlans')
      expect(source).not.toContain('pricingEmptyRecharge')
      expect(source).not.toContain('rechargeProductsTitle')
      expect(source).not.toContain('subscriptionPlansTitle')
      expect(source).not.toContain('topUpCreditLine')
      expect(source).not.toContain('topUpPriceLabel')
      expect(source).not.toContain('planValidityDays')
      expect(source).not.toContain('planValidityMonths')
      expect(source).not.toContain('planValidityYears')
    }
  })
})
