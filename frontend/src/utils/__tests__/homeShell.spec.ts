import { describe, expect, it } from 'vitest'
import {
  applyBusinessHomeCardRuntimeStatuses,
  resolveBusinessHomeCardsForRoutes,
  resolveBusinessHomeShellConfig,
  resolveHomeShellConfig,
  type HomeShellCopy,
} from '../homeShell'

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
    promptsPath: '/prompts',
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
              promptsPath: '/configured-prompts',
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
      promptsPath: '/configured-prompts',
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
              promptsPath: 'https://example.com/prompts',
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
    expect(config.labels.navExperience).toBe('能力入口')
    expect(config.defaults.links.modelsPath).toBe('/models')
    expect(config.defaults.links.promptsPath).toBe('/prompts')
    expect(config.defaults.links.experienceAnchor).toBe('#capabilities')
    expect(config.businessCards.map((card) => card.title)).toEqual([
      '提示词画廊',
      '微信导出',
      '热点追踪',
      '生图工作台',
    ])
    expect(config.businessCards[0]).toMatchObject({
      key: 'prompt-catalog',
      path: '/prompts',
      pathLabel: '进入提示词画廊',
    })
    expect(config.businessCards[1]).toMatchObject({
      key: 'wechat-export',
      path: '/wechat',
      pathLabel: '进入微信导出',
    })
    expect(config.businessCards[2]).toMatchObject({
      key: 'hot-topics',
      path: '/hot',
      pathLabel: '进入热点追踪',
      status: 'available',
      statusLabel: '可用',
      disabled: false,
      visible: true,
    })
    expect(config.businessCards[2].capabilityTags).toContain('内容采集')
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
              status: 'hidden',
              visible: false,
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
      status: 'hidden',
      visible: false,
    })
  })

  it('normalizes backend-style business card fields and status labels', () => {
    const config = resolveBusinessHomeShellConfig(
      JSON.stringify({
        zh: {
          businessCards: [
            {
              key: 'hot-topics',
              title: '热点',
              capability_tags: [' 内容采集 ', ''],
              path: '/hot',
              path_label: '进入热点',
              status: 'in_progress',
            },
            {
              key: 'bad-link',
              title: '外链',
              capability_tags: ['隐藏'],
              path: 'https://example.com',
              status: 'disabled',
              status_label: '管理员关闭',
            },
          ],
        },
      }),
      'zh',
    )

    expect(config.businessCards[0]).toMatchObject({
      key: 'hot-topics',
      capabilityTags: ['内容采集'],
      path: '/hot',
      pathLabel: '进入热点',
      status: 'in_progress',
      statusLabel: '建设中',
      disabled: false,
      visible: true,
    })
    expect(config.businessCards[1]).toMatchObject({
      key: 'bad-link',
      path: undefined,
      status: 'disabled',
      statusLabel: '管理员关闭',
    })
  })
})

describe('resolveBusinessHomeCardsForRoutes', () => {
  it('keeps available cards clickable only when the configured path exists', () => {
    const config = resolveBusinessHomeShellConfig(
      JSON.stringify({
        en: {
          businessCards: [
            {
              key: 'prompt-catalog',
              title: 'Prompt catalog',
              path: '/prompts',
              status: 'available',
            },
            {
              key: 'future-surface',
              title: 'Future surface',
              path: '/future',
              status: 'available',
            },
            {
              key: 'hidden-surface',
              title: 'Hidden surface',
              path: '/hidden',
              status: 'hidden',
              visible: false,
            },
          ],
        },
      }),
      'en',
    )

    const cards = resolveBusinessHomeCardsForRoutes(config.businessCards, ['/prompts'], 'en')

    expect(cards[0]).toMatchObject({
      key: 'prompt-catalog',
      status: 'available',
      disabled: false,
      statusLabel: 'Available',
    })
    expect(cards[1]).toMatchObject({
      key: 'future-surface',
      status: 'in_progress',
      disabled: true,
      statusLabel: 'In progress',
    })
    expect(cards[2]).toMatchObject({
      key: 'hidden-surface',
      status: 'hidden',
      visible: false,
    })
  })

  it('downgrades available cards without a path to an in-progress state', () => {
    const config = resolveBusinessHomeShellConfig(
      JSON.stringify({
        zh: {
          businessCards: [
            {
              key: 'no-route-yet',
              title: '建设中的能力',
              status: 'available',
            },
          ],
        },
      }),
      'zh',
    )

    const cards = resolveBusinessHomeCardsForRoutes(config.businessCards, ['/wechat'], 'zh')

    expect(cards[0]).toMatchObject({
      key: 'no-route-yet',
      status: 'in_progress',
      disabled: true,
      statusLabel: '建设中',
    })
  })
})

describe('applyBusinessHomeCardRuntimeStatuses', () => {
  it('downgrades only cards that are configured as available', () => {
    const config = resolveBusinessHomeShellConfig(
      JSON.stringify({
        en: {
          businessCards: [
            { key: 'prompt-catalog', title: 'Prompts', path: '/prompts', status: 'available' },
            { key: 'manual-progress', title: 'Manual progress', path: '/manual', status: 'in_progress' },
            { key: 'manual-disabled', title: 'Manual disabled', path: '/disabled', status: 'disabled' },
            { key: 'manual-hidden', title: 'Manual hidden', path: '/hidden', status: 'hidden', visible: false },
          ],
        },
      }),
      'en',
    )

    const cards = applyBusinessHomeCardRuntimeStatuses(
      config.businessCards,
      {
        'prompt-catalog': { status: 'in_progress', message: 'Prompt images are not reachable.', count: 12 },
        'manual-progress': { status: 'available', message: 'Configured count is live.', count: 3, statusLabel: 'Live' },
        'manual-disabled': { status: 'available' },
        'manual-hidden': { status: 'available' },
      },
      'en',
    )

    expect(cards[0]).toMatchObject({
      key: 'prompt-catalog',
      status: 'in_progress',
      disabled: true,
      statusLabel: 'In progress',
      statusMessage: 'Prompt images are not reachable.',
      statusCount: 12,
    })
    expect(cards[1]).toMatchObject({
      key: 'manual-progress',
      status: 'in_progress',
    })
    expect(cards[2]).toMatchObject({
      key: 'manual-disabled',
      status: 'disabled',
    })
    expect(cards[3]).toMatchObject({
      key: 'manual-hidden',
      status: 'hidden',
      visible: false,
    })
  })

  it('preserves available cards while attaching runtime status detail', () => {
    const config = resolveBusinessHomeShellConfig(
      JSON.stringify({
        en: {
          businessCards: [
            { key: 'prompt-catalog', title: 'Prompts', path: '/prompts', status: 'available' },
          ],
        },
      }),
      'en',
    )

    const cards = applyBusinessHomeCardRuntimeStatuses(
      config.businessCards,
      {
        'prompt-catalog': {
          status: 'available',
          statusLabel: 'Ready',
          message: '2,299 image prompt cases are indexed.',
          count: 2299,
        },
      },
      'en',
    )

    expect(cards[0]).toMatchObject({
      key: 'prompt-catalog',
      status: 'available',
      disabled: false,
      visible: true,
      statusLabel: 'Ready',
      statusMessage: '2,299 image prompt cases are indexed.',
      statusCount: 2299,
    })
  })
})
