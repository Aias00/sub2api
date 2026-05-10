export interface DocsPageDefinition {
  slug: string
  title: string
  description: string
  section: string
  order: number
  file: string
}

export interface DocsPage extends DocsPageDefinition {
  content: string
}

export interface DocsSection {
  id: string
  title: string
  pages: DocsPage[]
}

export interface DocsLinkTarget {
  internal: boolean
  to: string
  href: string
}

const INTERNAL_DOCS_PATH = '/docs'

const rawDocs = import.meta.glob('../docs/pages/**/*.md', {
  query: '?raw',
  import: 'default',
  eager: true,
}) as Record<string, string>

const docsPageDefinitions: DocsPageDefinition[] = [
  {
    slug: 'getting-started/overview',
    title: '概览',
    description: '了解 cloudbase 的核心能力、适用场景和整体使用方式。',
    section: '开始使用',
    order: 10,
    file: '../docs/pages/getting-started/overview.md',
  },
  {
    slug: 'getting-started/quick-start',
    title: '快速开始',
    description: '从注册、获取 API Key 到第一次成功调用接口的最短路径。',
    section: '开始使用',
    order: 20,
    file: '../docs/pages/getting-started/quick-start.md',
  },
  {
    slug: 'guides/api-keys',
    title: 'API 密钥',
    description: '创建、绑定、管理和安全使用 API Key 的说明。',
    section: '核心指南',
    order: 30,
    file: '../docs/pages/guides/api-keys.md',
  },
  {
    slug: 'guides/gateway',
    title: '网关调用',
    description: '如何使用统一网关地址、鉴权头和示例代码接入模型。',
    section: '核心指南',
    order: 40,
    file: '../docs/pages/guides/gateway.md',
  },
  {
    slug: 'guides/billing',
    title: '充值与订阅',
    description: '充值商品、订阅套餐、订单查询与支付排查说明。',
    section: '核心指南',
    order: 50,
    file: '../docs/pages/guides/billing.md',
  },
  {
    slug: 'guides/security',
    title: '账号与安全',
    description: '资料维护、密码修改与双因素认证（2FA）相关流程。',
    section: '核心指南',
    order: 60,
    file: '../docs/pages/guides/security.md',
  },
]

export const docsPages: DocsPage[] = docsPageDefinitions
  .map((definition) => ({
    ...definition,
    content: rawDocs[definition.file] || '',
  }))
  .sort((left, right) => left.order - right.order)

export const defaultDocsSlug = docsPages[0]?.slug || ''

export const docsSections: DocsSection[] = Array.from(
  docsPages.reduce((sections, page) => {
    const section = sections.get(page.section) ?? {
      id: page.section.toLowerCase().replace(/[^\w一-龥]+/g, '-'),
      title: page.section,
      pages: [],
    }
    section.pages.push(page)
    sections.set(page.section, section)
    return sections
  }, new Map<string, DocsSection>())
).map(([, section]) => ({
  ...section,
  pages: section.pages.sort((left, right) => left.order - right.order),
}))

export function findDocsPage(slug: string): DocsPage | null {
  return docsPages.find((page) => page.slug === slug) ?? null
}

export function getAdjacentDocsPages(slug: string): {
  previous: DocsPage | null
  next: DocsPage | null
} {
  const currentIndex = docsPages.findIndex((page) => page.slug === slug)
  if (currentIndex === -1) {
    return { previous: null, next: null }
  }
  return {
    previous: currentIndex > 0 ? docsPages[currentIndex - 1] : null,
    next: currentIndex < docsPages.length - 1 ? docsPages[currentIndex + 1] : null,
  }
}

export function resolveDocsLink(docUrl: string, currentOrigin: string): DocsLinkTarget {
  const trimmed = docUrl.trim()
  if (!trimmed) {
    return {
      internal: true,
      to: INTERNAL_DOCS_PATH,
      href: INTERNAL_DOCS_PATH,
    }
  }

  try {
    const baseOrigin = currentOrigin || 'https://local.invalid'
    const parsed = new URL(trimmed, baseOrigin)
    const normalizedPath = parsed.pathname.replace(/\/+$/g, '') || '/'
    const normalizedOrigin = new URL(baseOrigin).origin
    const sameOrigin = parsed.origin === normalizedOrigin
    const shouldUseInternalDocs =
      sameOrigin &&
      (normalizedPath === '/' ||
        normalizedPath === '/home' ||
        normalizedPath === '/docs' ||
        normalizedPath.startsWith('/docs/'))

    if (shouldUseInternalDocs) {
      return {
        internal: true,
        to: INTERNAL_DOCS_PATH,
        href: INTERNAL_DOCS_PATH,
      }
    }

    return {
      internal: false,
      to: parsed.toString(),
      href: parsed.toString(),
    }
  } catch {
    return {
      internal: true,
      to: INTERNAL_DOCS_PATH,
      href: INTERNAL_DOCS_PATH,
    }
  }
}
