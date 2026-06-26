export type HomeShellCopy = {
  viewDocs: string
  dashboard: string
  login: string
  primaryCta: string
  secondaryCta: string
  heroBadge: string
  heroTitle: string
  heroDescription: string
  modelMatrixKicker: string
  modelMatrixTitle: string
  modelMatrixDescription: string
  modelMatrixEmptyCard: string
  modelMatrixEmptyPill: string
  experienceKicker: string
  experienceTitle: string
  experienceDescription: string
  whyChooseKicker: string
  whyChooseTitle: string
  whyChooseDescription: string
  footerDescription: string
  allRightsReserved: string
  termsLink: string
  privacyLink: string
  navHome: string
  navDocs: string
  navModels: string
  navExperience: string
  footerProduct: string
  footerCatalog: string
  footerSupport: string
  familyClaudeBadge: string
  familyGptBadge: string
  familyClaudeTagline: string
  familyGptTagline: string
  familyClaudeDescription: string
  familyGptDescription: string
  familyClaudeReasoning: string
  familyClaudeArchitecture: string
  familyClaudeReview: string
  familyGptCoding: string
  familyGptIteration: string
  familyGptAgents: string
}

export type HomeShellConfig = {
  labels: HomeShellCopy
  defaults: HomeShellDefaults
  experienceCards: HomeExperienceCard[]
  whyChooseCards: HomeWhyChooseCard[]
  businessCards: HomeBusinessCard[]
}

export type HomeShellDefaults = {
  links: {
    homeAnchor: string
    modelsPath: string
    experienceAnchor: string
    docsPath: string
    termsPath: string
    privacyPath: string
  }
}

export type HomeExperienceCard = {
  key: string
  icon?: 'server' | 'key' | 'sparkles' | 'chart'
  iconClass: string
  title: string
  description: string
}

export type HomeExperienceCardOverride = Partial<HomeExperienceCard> & {
  key?: string
}

export type HomeWhyChooseCard = {
  key: string
  title: string
  description: string
}

export type HomeWhyChooseCardOverride = Partial<HomeWhyChooseCard> & {
  key?: string
}

export type HomeBusinessCard = {
  key: string
  badge: string
  title: string
  description: string
  capabilityTags: string[]
  path?: string
  pathLabel?: string
}

const homeLabelKeys: Array<keyof HomeShellCopy> = [
  'viewDocs',
  'dashboard',
  'login',
  'primaryCta',
  'secondaryCta',
  'heroBadge',
  'heroTitle',
  'heroDescription',
  'modelMatrixKicker',
  'modelMatrixTitle',
  'modelMatrixDescription',
  'modelMatrixEmptyCard',
  'modelMatrixEmptyPill',
  'experienceKicker',
  'experienceTitle',
  'experienceDescription',
  'whyChooseKicker',
  'whyChooseTitle',
  'whyChooseDescription',
  'footerDescription',
  'allRightsReserved',
  'termsLink',
  'privacyLink',
  'navHome',
  'navDocs',
  'navModels',
  'navExperience',
  'footerProduct',
  'footerCatalog',
  'footerSupport',
  'familyClaudeBadge',
  'familyGptBadge',
  'familyClaudeTagline',
  'familyGptTagline',
  'familyClaudeDescription',
  'familyGptDescription',
  'familyClaudeReasoning',
  'familyClaudeArchitecture',
  'familyClaudeReview',
  'familyGptCoding',
  'familyGptIteration',
  'familyGptAgents',
]

export function resolveHomeShellConfig(raw: string | undefined, selectedLocale: 'zh' | 'en'): HomeShellConfig {
  const emptyConfig = createEmptyHomeShellConfig()
  if (!raw?.trim()) {
    return emptyConfig
  }
  try {
    const parsed = JSON.parse(raw) as unknown
    if (!isRecord(parsed)) {
      return emptyConfig
    }
    const localized = parsed[selectedLocale] ?? parsed.en ?? parsed.zh ?? parsed
    if (!isRecord(localized)) {
      return emptyConfig
    }
    return {
      labels: isRecord(localized.labels) ? readHomeLabels(localized.labels) : emptyConfig.labels,
      defaults: readHomeDefaults(localized.defaults, emptyConfig.defaults),
      businessCards: readHomeBusinessCards(localized.businessCards),
      experienceCards: readHomeExperienceCards(localized.experienceCards),
      whyChooseCards: readHomeWhyChooseCards(localized.whyChooseCards),
    }
  } catch {
    return emptyConfig
  }
}

export function resolveBusinessHomeShellConfig(raw: string | undefined, selectedLocale: 'zh' | 'en'): HomeShellConfig {
  if (raw?.trim()) {
    return resolveHomeShellConfig(raw, selectedLocale)
  }
  return selectedLocale === 'zh' ? businessHomeShellConfigZh : businessHomeShellConfigEn
}

function createEmptyHomeShellConfig(): HomeShellConfig {
  return {
    labels: homeLabelKeys.reduce((copy, key) => {
      copy[key] = ''
      return copy
    }, {} as HomeShellCopy),
    defaults: {
      links: {
        homeAnchor: '#top',
        modelsPath: '/models',
        experienceAnchor: '#experience',
        docsPath: '/docs',
        termsPath: '/legal/terms',
        privacyPath: '/legal/privacy-policy',
      },
    },
    experienceCards: [],
    whyChooseCards: [],
    businessCards: [],
  }
}

function readHomeDefaults(value: unknown, fallback: HomeShellDefaults): HomeShellDefaults {
  const defaults = isRecord(value) ? value : {}
  const links = isRecord(defaults.links) ? defaults.links : {}

  return {
    links: {
      homeAnchor: readInternalLink(links.homeAnchor, fallback.links.homeAnchor),
      modelsPath: readInternalLink(links.modelsPath, fallback.links.modelsPath),
      experienceAnchor: readInternalLink(links.experienceAnchor, fallback.links.experienceAnchor),
      docsPath: readInternalLink(links.docsPath, fallback.links.docsPath),
      termsPath: readInternalLink(links.termsPath, fallback.links.termsPath),
      privacyPath: readInternalLink(links.privacyPath, fallback.links.privacyPath),
    },
  }
}

function readInternalLink(value: unknown, fallback: string): string {
  if (typeof value !== 'string') return fallback
  const trimmed = value.trim()
  if (!trimmed) return fallback
  if (trimmed.startsWith('#')) return trimmed
  if (!trimmed.startsWith('/') || trimmed.startsWith('//') || trimmed.includes('\\')) return fallback
  return trimmed
}

function readHomeLabels(labels: Record<string, unknown>): HomeShellCopy {
  const copy = createEmptyHomeShellConfig().labels
  for (const key of homeLabelKeys) {
    const label = readString(labels[key])
    if (label) copy[key] = label
  }
  return copy
}

function readHomeExperienceCards(value: unknown): HomeExperienceCard[] {
  if (!Array.isArray(value)) {
    return []
  }
  return value.filter(isRecord).map((card, index) => ({
    key: readString(card.key) || `experience-${index + 1}`,
    icon: readExperienceIcon(card.icon),
    iconClass: readString(card.iconClass) || '',
    title: readString(card.title) || '',
    description: readString(card.description) || '',
  }))
}

function readHomeBusinessCards(value: unknown): HomeBusinessCard[] {
  if (!Array.isArray(value)) {
    return []
  }
  return value.filter(isRecord).map((card, index) => ({
    key: readString(card.key) || `business-${index + 1}`,
    badge: readString(card.badge) || '',
    title: readString(card.title) || '',
    description: readString(card.description) || '',
    capabilityTags: Array.isArray(card.capabilityTags)
      ? card.capabilityTags.filter((item): item is string => typeof item === 'string' && item.trim().length > 0)
      : [],
    path: readInternalLink(card.path, '') || undefined,
    pathLabel: readString(card.pathLabel) || '',
  }))
}

function readHomeWhyChooseCards(value: unknown): HomeWhyChooseCard[] {
  if (!Array.isArray(value)) {
    return []
  }
  return value.filter(isRecord).map((card, index) => ({
    key: readString(card.key) || `why-${index + 1}`,
    title: readString(card.title) || '',
    description: readString(card.description) || '',
  }))
}

function readExperienceIcon(value: unknown): HomeExperienceCard['icon'] | undefined {
  if (value === 'server' || value === 'key' || value === 'sparkles' || value === 'chart') {
    return value
  }
  return undefined
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

function readString(value: unknown): string | undefined {
  return typeof value === 'string' ? value : undefined
}

const businessHomeShellConfigZh: HomeShellConfig = {
  labels: {
    viewDocs: '文档',
    dashboard: '控制台',
    login: '登录',
    primaryCta: '进入能力中台',
    secondaryCta: '查看图片提示词',
    heroBadge: '业务能力首页',
    heroTitle: '面向业务场景的 AI 能力工作台',
    heroDescription:
      'Sub2API 以后沉淀用户、订单、套餐、支付等中台能力；首页重点展示微信导出、热点、图片提示词和生图工作台等可直接理解的业务能力。',
    modelMatrixKicker: '业务能力',
    modelMatrixTitle: '把高频业务能力摆到首页',
    modelMatrixDescription:
      '围绕内容采集、提示词沉淀与图像生产流程，先把可落地的能力入口讲清楚，再由中台承接账户、订单和套餐等底层能力。',
    modelMatrixEmptyCard: '业务能力即将上线',
    modelMatrixEmptyPill: '建设中',
    experienceKicker: '中台定位',
    experienceTitle: '业务能力前台，平台能力落到中台',
    experienceDescription:
      '用户、订单、套餐、支付与账户治理逐步统一收口到 Sub2API 中台，让前台页面更多表达业务价值而不是底层接线细节。',
    whyChooseKicker: '能力组织方式',
    whyChooseTitle: '先讲用户能完成什么，再讲平台怎么支撑',
    whyChooseDescription:
      '首页围绕业务工作流编排；底层代理、模型路由和结算能力继续由中台承接。',
    footerDescription: '聚焦业务能力表达，由中台统一承接用户、订单、套餐和支付能力。',
    allRightsReserved: '保留所有权利。',
    termsLink: '服务条款',
    privacyLink: '隐私政策',
    navHome: '首页',
    navDocs: '文档',
    navModels: '提示词',
    navExperience: '能力',
    footerProduct: '首页入口',
    footerCatalog: '业务能力',
    footerSupport: '支持',
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
  },
  defaults: {
    links: {
      homeAnchor: '#top',
      modelsPath: '/prompts',
      experienceAnchor: '#experience',
      docsPath: '/docs',
      termsPath: '/legal/terms',
      privacyPath: '/legal/privacy-policy',
    },
  },
  businessCards: [
    {
      key: 'wechat-export',
      badge: 'Workflow',
      title: '微信导出',
      description: '沉淀公众号内容导出与整理能力，适合把文章资产回收到统一工作流里。',
      capabilityTags: ['内容导出', '素材整理', '资产回收'],
      path: '/wechat-export',
      pathLabel: '进入微信导出',
    },
    {
      key: 'hot-topics',
      badge: 'Signal',
      title: '热点追踪',
      description: '围绕热点发现、筛选和后续处理，把高频内容观察任务做成稳定入口。',
      capabilityTags: ['热点收集', '线索筛选', '内容观察'],
    },
    {
      key: 'prompt-catalog',
      badge: 'Library',
      title: '图片提示词',
      description: '把沉淀下来的图片提示词案例放到统一目录里，便于检索、复用和二次加工。',
      capabilityTags: ['案例目录', '检索复用', '图像提示词'],
      path: '/prompts',
      pathLabel: '进入提示词库',
    },
    {
      key: 'image-workspace',
      badge: 'Workspace',
      title: '生图工作台',
      description: '以提示词工作流为中心组织图片生成前的整理、复制和后续生产衔接。',
      capabilityTags: ['Prompt 工作流', '生图准备', '工作台'],
      path: '/image-generator',
      pathLabel: '进入工作台',
    },
  ],
  experienceCards: [
    {
      key: 'platform',
      icon: 'server',
      iconClass: 'bg-gradient-to-br from-sky-500 to-blue-600',
      title: '中台统一承接用户与订单',
      description: '前台页面聚焦业务表达，用户、订单、支付和套餐配置逐步收口到统一能力中台。',
    },
    {
      key: 'catalog',
      icon: 'key',
      iconClass: 'bg-gradient-to-br from-indigo-500 to-violet-600',
      title: '内容能力先产品化',
      description: '优先把微信导出、热点、提示词和生图工作流做成稳定能力，再让底层平台持续支撑它们。',
    },
    {
      key: 'ops',
      icon: 'sparkles',
      iconClass: 'bg-gradient-to-br from-emerald-500 to-teal-600',
      title: '前后台职责更清晰',
      description: '首页讲业务价值，后台负责配置、数据和运行时控制，减少首页同时承担两种叙事。',
    },
  ],
  whyChooseCards: [
    {
      key: 'business-first',
      title: '先围绕业务入口组织',
      description: '把用户真正会点开的业务能力放在首页，而不是先暴露中台实现细节。',
    },
    {
      key: 'platform-backbone',
      title: '中台继续做能力骨架',
      description: '账户、订单、套餐与支付能力继续沉到 Sub2API 中台，不需要在首页重复解释。',
    },
    {
      key: 'reuse',
      title: '提示词与内容资产可复用',
      description: '把图片提示词、导出内容和热点线索组织成可持续复用的业务资产。',
    },
    {
      key: 'workflow',
      title: '形成工作流闭环',
      description: '从内容导出、热点发现到提示词沉淀、生图准备，首页直接表达完整业务链路。',
    },
  ],
}

const businessHomeShellConfigEn: HomeShellConfig = {
  labels: {
    viewDocs: 'Docs',
    dashboard: 'Dashboard',
    login: 'Log in',
    primaryCta: 'Open the platform',
    secondaryCta: 'Browse prompt cases',
    heroBadge: 'Business capability home',
    heroTitle: 'An AI workspace organized around business capabilities',
    heroDescription:
      'Sub2API will keep consolidating users, orders, plans, and payment into the capability platform while the homepage highlights concrete workflows such as WeChat export, hot-topic tracking, prompt cases, and the image workspace.',
    modelMatrixKicker: 'Capabilities',
    modelMatrixTitle: 'Put business workflows on the homepage',
    modelMatrixDescription:
      'Lead with content export, discovery, prompt reuse, and image production workflows while the platform layer continues to own accounts, plans, and billing.',
    modelMatrixEmptyCard: 'Business capability coming soon',
    modelMatrixEmptyPill: 'In progress',
    experienceKicker: 'Platform direction',
    experienceTitle: 'Business-facing home, platform-backed operations',
    experienceDescription:
      'Users, orders, plans, payments, and account management continue moving into the Sub2API platform so public pages can focus on user-facing workflows.',
    whyChooseKicker: 'Information architecture',
    whyChooseTitle: 'Explain what users can do before how the platform works',
    whyChooseDescription:
      'The homepage should foreground business workflows while the platform continues to power routing, account management, and settlement behind the scenes.',
    footerDescription: 'Homepage messaging focused on business capabilities, backed by a unified platform for users, plans, orders, and payments.',
    allRightsReserved: 'All rights reserved.',
    termsLink: 'Terms',
    privacyLink: 'Privacy',
    navHome: 'Home',
    navDocs: 'Docs',
    navModels: 'Prompts',
    navExperience: 'Capabilities',
    footerProduct: 'Entry points',
    footerCatalog: 'Workflows',
    footerSupport: 'Support',
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
  },
  defaults: {
    links: {
      homeAnchor: '#top',
      modelsPath: '/prompts',
      experienceAnchor: '#experience',
      docsPath: '/docs',
      termsPath: '/legal/terms',
      privacyPath: '/legal/privacy-policy',
    },
  },
  businessCards: [
    {
      key: 'wechat-export',
      badge: 'Workflow',
      title: 'WeChat Export',
      description: 'Turn WeChat export and article recovery into a stable workflow entry instead of an ad hoc operation.',
      capabilityTags: ['Content export', 'Asset recovery', 'Workflow'],
      path: '/wechat-export',
      pathLabel: 'Open WeChat export',
    },
    {
      key: 'hot-topics',
      badge: 'Signal',
      title: 'Hot Topic Tracking',
      description: 'Package hot-topic discovery and follow-up processing into a clearer product surface.',
      capabilityTags: ['Signal collection', 'Trend tracking', 'Content ops'],
    },
    {
      key: 'prompt-catalog',
      badge: 'Library',
      title: 'Image Prompt Cases',
      description: 'Keep image prompt cases in a searchable catalog so teams can reuse and refine proven material.',
      capabilityTags: ['Prompt library', 'Search', 'Reuse'],
      path: '/prompts',
      pathLabel: 'Open prompt catalog',
    },
    {
      key: 'image-workspace',
      badge: 'Workspace',
      title: 'Image Workspace',
      description: 'Center the image workflow around prompt preparation and handoff instead of exposing only the platform plumbing.',
      capabilityTags: ['Prompt workflow', 'Image prep', 'Workspace'],
      path: '/image-generator',
      pathLabel: 'Open workspace',
    },
  ],
  experienceCards: [
    {
      key: 'platform',
      icon: 'server',
      iconClass: 'bg-gradient-to-br from-sky-500 to-blue-600',
      title: 'Platform-owned users and orders',
      description: 'The public home can focus on business workflows while user, order, payment, and plan capabilities consolidate behind the platform.',
    },
    {
      key: 'catalog',
      icon: 'key',
      iconClass: 'bg-gradient-to-br from-indigo-500 to-violet-600',
      title: 'Productize content workflows first',
      description: 'Lead with WeChat export, hot topics, prompt cases, and image preparation instead of putting infrastructure copy first.',
    },
    {
      key: 'ops',
      icon: 'sparkles',
      iconClass: 'bg-gradient-to-br from-emerald-500 to-teal-600',
      title: 'Clearer split between home and platform',
      description: 'The homepage explains user-facing workflows; the platform continues to own runtime controls, routing, and settlement.',
    },
  ],
  whyChooseCards: [
    {
      key: 'business-first',
      title: 'Organize around workflows users recognize',
      description: 'Put the workflows people actually want to enter from the homepage ahead of the supporting platform internals.',
    },
    {
      key: 'platform-backbone',
      title: 'Keep the platform as the backbone',
      description: 'Users, orders, plans, and payment keep consolidating into Sub2API without forcing every homepage section to explain the machinery.',
    },
    {
      key: 'reuse',
      title: 'Make prompts and content reusable assets',
      description: 'Turn prompt cases, exported content, and hot-topic findings into assets that can be searched, refined, and reused.',
    },
    {
      key: 'workflow',
      title: 'Show a complete workflow story',
      description: 'Move from export and topic discovery into prompt curation and image preparation with a clearer end-to-end capability narrative.',
    },
  ],
}
