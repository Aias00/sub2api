import { describe, expect, it } from 'vitest'
import { resolveBusinessHomeShellConfig, resolveHomeShellConfig, type HomeShellCopy } from '../homeShell'

const emptyLabels: HomeShellCopy = {
  viewDocs: '',
  dashboard: '',
  login: '',
  primaryCta: '',
  secondaryCta: '',
  heroBadge: '',
  heroTitle: '',
  heroDescription: '',
  modelMatrixKicker: '',
  modelMatrixTitle: '',
  modelMatrixDescription: '',
  modelMatrixEmptyCard: '',
  modelMatrixEmptyPill: '',
  experienceKicker: '',
  experienceTitle: '',
  experienceDescription: '',
  whyChooseKicker: '',
  whyChooseTitle: '',
  whyChooseDescription: '',
  footerDescription: '',
  allRightsReserved: '',
  termsLink: '',
  privacyLink: '',
  navHome: '',
  navDocs: '',
  navModels: '',
  navExperience: '',
  footerProduct: '',
  footerCatalog: '',
  footerSupport: '',
  familyClaudeBadge: '',
  familyGptBadge: '',
  familyClaudeTagline: '',
  familyGptTagline: '',
  familyClaudeDescription: '',
  familyGptDescription: '',
  familyClaudeReasoning: '',
  familyClaudeArchitecture: '',
  familyClaudeReview: '',
  familyGptCoding: '',
  familyGptIteration: '',
  familyGptAgents: '',
}

const emptyDefaults = {
  links: {
    homeAnchor: '#top',
    modelsPath: '/models',
    experienceAnchor: '#experience',
    docsPath: '/docs',
    termsPath: '/legal/terms',
    privacyPath: '/legal/privacy-policy',
  },
}

describe('resolveHomeShellConfig', () => {
  it('resolves labels and card overrides from localized config', () => {
    const config = resolveHomeShellConfig(
      JSON.stringify({
        en: {
          labels: {
            heroTitle: 'Configured hero',
            navHome: 'Configured home',
            ignored: 'ignored',
          },
          experienceCards: [
            {
              key: 'unified',
              icon: 'sparkles',
              title: 'Configured unified',
            },
          ],
          whyChooseCards: [
            {
              key: 'lowFriction',
              description: 'Configured why',
            },
          ],
          defaults: {
            links: {
              homeAnchor: '#configured-top',
              modelsPath: '/configured-models',
              experienceAnchor: '#configured-experience',
              docsPath: '/configured-docs',
              termsPath: '/configured-terms',
              privacyPath: '/configured-privacy',
            },
          },
        },
      }),
      'en',
    )

    expect(config.labels.heroTitle).toBe('Configured hero')
    expect(config.labels.navHome).toBe('Configured home')
    expect(config.labels.viewDocs).toBe('')
    expect(config.experienceCards?.[0]).toMatchObject({
      key: 'unified',
      icon: 'sparkles',
      title: 'Configured unified',
      iconClass: '',
      description: '',
    })
    expect(config.whyChooseCards?.[0]).toMatchObject({
      key: 'lowFriction',
      description: 'Configured why',
      title: '',
    })
    expect(config.defaults.links).toEqual({
      homeAnchor: '#configured-top',
      modelsPath: '/configured-models',
      experienceAnchor: '#configured-experience',
      docsPath: '/configured-docs',
      termsPath: '/configured-terms',
      privacyPath: '/configured-privacy',
    })
  })

  it('keeps unsafe configured links on local fallbacks', () => {
    const config = resolveHomeShellConfig(
      JSON.stringify({
        en: {
          defaults: {
            links: {
              homeAnchor: 'https://example.com',
              modelsPath: '//example.com/models',
              experienceAnchor: '/bad\\path',
            },
          },
        },
      }),
      'en',
    )

    expect(config.defaults).toEqual(emptyDefaults)
  })

  it('does not assign frontend default icons to experience cards', () => {
    const config = resolveHomeShellConfig(
      JSON.stringify({
        en: {
          experienceCards: [
            { key: 'missing-icon', title: 'Missing icon' },
            { key: 'bad-icon', icon: 'rocket', title: 'Bad icon' },
          ],
        },
      }),
      'en',
    )

    expect(config.experienceCards).toEqual([
      {
        key: 'missing-icon',
        icon: undefined,
        iconClass: '',
        title: 'Missing icon',
        description: '',
      },
      {
        key: 'bad-icon',
        icon: undefined,
        iconClass: '',
        title: 'Bad icon',
        description: '',
      },
    ])
  })

  it('falls back to an empty config for invalid JSON', () => {
    expect(resolveHomeShellConfig('{bad json', 'en')).toEqual({
      labels: emptyLabels,
      defaults: emptyDefaults,
      experienceCards: [],
      whyChooseCards: [],
      businessCards: [],
    })
  })
})

describe('resolveBusinessHomeShellConfig', () => {
  it('returns a dedicated business-home shell for /home', () => {
    const config = resolveBusinessHomeShellConfig(undefined, 'zh')

    expect(config.labels.heroBadge).toBe('业务能力首页')
    expect(config.labels.heroTitle).toBe('面向业务场景的 AI 能力工作台')
    expect(config.labels.navModels).toBe('提示词')
    expect(config.defaults.links.modelsPath).toBe('/prompts')
    expect(config.businessCards.map((card) => card.title)).toEqual([
      '微信导出',
      '热点追踪',
      '图片提示词',
      '生图工作台',
    ])
    expect(config.experienceCards.length).toBeGreaterThan(0)
    expect(config.whyChooseCards.length).toBeGreaterThan(0)
  })

  it('prefers runtime-configured business home copy when provided', () => {
    const config = resolveBusinessHomeShellConfig(
      JSON.stringify({
        en: {
          labels: {
            heroTitle: 'Configured business home',
          },
          businessCards: [
            {
              key: 'prompt-catalog',
              title: 'Configured prompt catalog',
            },
          ],
        },
      }),
      'en',
    )

    expect(config.labels.heroTitle).toBe('Configured business home')
    expect(config.businessCards[0]).toMatchObject({
      key: 'prompt-catalog',
      title: 'Configured prompt catalog',
    })
  })
})
