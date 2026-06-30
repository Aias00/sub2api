import { resolveRuntimeLocale } from '@/utils/runtimeLocale'

export const adminSidebarItemKeys = [
  'dashboard',
  'ops',
  'users',
  'groups',
  'channels',
  'subscriptions',
  'accounts',
  'announcements',
  'proxies',
  'riskControl',
  'redeem',
  'promoCodes',
  'affiliates',
  'orders',
  'usage',
  'apiKeys',
  'runtimeSettings',
  'settings',
] as const

export type AdminSidebarItemKey = typeof adminSidebarItemKeys[number]

export type AdminSidebarSection = {
  id: string
  items: AdminSidebarItemKey[]
}

export const selfSidebarItemKeys = [
  'dashboard',
  'tasks',
  'promptCatalog',
  'imageGenerator',
  'wechatExport',
  'hotTopics',
  'apiKeys',
  'usage',
  'availableChannels',
  'availableGroups',
  'subscriptions',
  'purchase',
  'orders',
  'redeem',
  'affiliate',
  'profile',
] as const

export type SelfSidebarItemKey = typeof selfSidebarItemKeys[number]

export type SelfSidebarSection = {
  id: string
  items: SelfSidebarItemKey[]
}

const allowedItemKeys = new Set<AdminSidebarItemKey>(adminSidebarItemKeys)
const allowedSelfItemKeys = new Set<SelfSidebarItemKey>(selfSidebarItemKeys)

export function resolveAdminSidebarSections(
  rawAuthShellConfig?: string,
  runtimeLocale?: unknown,
): AdminSidebarSection[] {
  const sections = readConfiguredSections(rawAuthShellConfig, runtimeLocale)
  return sections.length > 0 ? sections : []
}

export function resolveSelfSidebarSections(
  rawAuthShellConfig: string | undefined,
  runtimeLocale: unknown,
  key: 'userSidebarSections' | 'adminPersonalSidebarSections',
): SelfSidebarSection[] {
  const sections = readConfiguredSelfSections(rawAuthShellConfig, runtimeLocale, key)
  return sections.length > 0 ? sections : []
}

function readConfiguredSections(
  rawAuthShellConfig: string | undefined,
  runtimeLocale: unknown,
): AdminSidebarSection[] {
  if (!rawAuthShellConfig?.trim()) {
    return []
  }

  try {
    const parsed = JSON.parse(rawAuthShellConfig) as unknown
    if (!isRecord(parsed)) {
      return []
    }
    const localized = pickLocalizedConfig(parsed, resolveRuntimeLocale(runtimeLocale))
    if (!localized || !isRecord(localized.defaults)) {
      return []
    }
    const rawSections = localized.defaults.adminSidebarSections
    if (!Array.isArray(rawSections)) {
      return []
    }

    const seen = new Set<AdminSidebarItemKey>()
    const sections: AdminSidebarSection[] = []
    for (let index = 0; index < rawSections.length; index += 1) {
      const section = rawSections[index]
      if (!isRecord(section) || !Array.isArray(section.items)) {
        continue
      }
      const id = readSectionID(section.id, index)
      const items = section.items
        .map((value) => (typeof value === 'string' ? value.trim() : ''))
        .filter((value): value is AdminSidebarItemKey => allowedItemKeys.has(value as AdminSidebarItemKey))
        .filter((value) => {
          if (seen.has(value)) return false
          seen.add(value)
          return true
        })

      if (items.length > 0) {
        sections.push({ id, items })
      }
    }
    return sections
  } catch {
    return []
  }
}

function readConfiguredSelfSections(
  rawAuthShellConfig: string | undefined,
  runtimeLocale: unknown,
  key: 'userSidebarSections' | 'adminPersonalSidebarSections',
): SelfSidebarSection[] {
  if (!rawAuthShellConfig?.trim()) {
    return []
  }

  try {
    const parsed = JSON.parse(rawAuthShellConfig) as unknown
    if (!isRecord(parsed)) {
      return []
    }
    const localized = pickLocalizedConfig(parsed, resolveRuntimeLocale(runtimeLocale))
    if (!localized || !isRecord(localized.defaults)) {
      return []
    }
    const rawSections = localized.defaults[key]
    if (!Array.isArray(rawSections)) {
      return []
    }

    const seen = new Set<SelfSidebarItemKey>()
    const sections: SelfSidebarSection[] = []
    for (let index = 0; index < rawSections.length; index += 1) {
      const section = rawSections[index]
      if (!isRecord(section) || !Array.isArray(section.items)) {
        continue
      }
      const id = readSectionID(section.id, index)
      const items = section.items
        .map((value) => (typeof value === 'string' ? value.trim() : ''))
        .filter((value): value is SelfSidebarItemKey => allowedSelfItemKeys.has(value as SelfSidebarItemKey))
        .filter((value) => {
          if (seen.has(value)) return false
          seen.add(value)
          return true
        })

      if (items.length > 0) {
        sections.push({ id, items })
      }
    }
    return sections
  } catch {
    return []
  }
}

function readSectionID(value: unknown, index: number): string {
  if (typeof value !== 'string') {
    return `section-${index + 1}`
  }
  const trimmed = value.trim()
  return trimmed || `section-${index + 1}`
}

function pickLocalizedConfig(
  parsed: Record<string, unknown>,
  runtimeLocale: string,
): Record<string, unknown> | null {
  const normalizedLocale = runtimeLocale.toLowerCase()
  const localeKeys = [normalizedLocale, normalizedLocale.split('-')[0], 'en', 'zh']
  for (const key of localeKeys) {
    const localized = parsed[key]
    if (isRecord(localized)) {
      return localized
    }
  }
  return parsed
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}
