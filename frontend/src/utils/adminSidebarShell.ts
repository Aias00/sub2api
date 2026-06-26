import { resolveRuntimeLocale } from '@/utils/runtimeLocale'

export type AdminSidebarRouteDefaults = {
  adminOpsPath: string
  adminUsersPath: string
  adminGroupsPath: string
  adminChannelsPath: string
  adminChannelPricingPath: string
  adminChannelMonitorPath: string
  adminSubscriptionsPath: string
  adminAccountsPath: string
  adminAnnouncementsPath: string
  adminProxiesPath: string
  adminRiskControlPath: string
  adminRedeemPath: string
  adminPromoCodesPath: string
  adminAffiliatesPath: string
  adminAffiliateOverviewPath: string
  adminAffiliateRulesPath: string
  adminAffiliateCodesPath: string
  adminAffiliateInvitesPath: string
  adminAffiliateRebatesPath: string
  adminAffiliateTransfersPath: string
  adminOrdersRootPath: string
  adminOrdersDashboardPath: string
  adminPaymentPlansPath: string
  adminUsagePath: string
}

const FALLBACK_ADMIN_SIDEBAR_ROUTE_DEFAULTS: AdminSidebarRouteDefaults = {
  adminOpsPath: '/admin/ops',
  adminUsersPath: '/admin/users',
  adminGroupsPath: '/admin/groups',
  adminChannelsPath: '/admin/channels',
  adminChannelPricingPath: '/admin/channels/pricing',
  adminChannelMonitorPath: '/admin/channels/monitor',
  adminSubscriptionsPath: '/admin/subscriptions',
  adminAccountsPath: '/admin/accounts',
  adminAnnouncementsPath: '/admin/announcements',
  adminProxiesPath: '/admin/proxies',
  adminRiskControlPath: '/admin/risk-control',
  adminRedeemPath: '/admin/redeem',
  adminPromoCodesPath: '/admin/promo-codes',
  adminAffiliatesPath: '/admin/affiliates',
  adminAffiliateOverviewPath: '/admin/affiliates/overview',
  adminAffiliateRulesPath: '/admin/affiliates/rules',
  adminAffiliateCodesPath: '/admin/affiliates/codes',
  adminAffiliateInvitesPath: '/admin/affiliates/invites',
  adminAffiliateRebatesPath: '/admin/affiliates/rebates',
  adminAffiliateTransfersPath: '/admin/affiliates/transfers',
  adminOrdersRootPath: '/admin/orders',
  adminOrdersDashboardPath: '/admin/orders/dashboard',
  adminPaymentPlansPath: '/admin/orders/plans',
  adminUsagePath: '/admin/usage',
}

type AdminSidebarRouteKey = keyof AdminSidebarRouteDefaults

const adminSidebarRouteKeys: AdminSidebarRouteKey[] = [
  'adminOpsPath',
  'adminUsersPath',
  'adminGroupsPath',
  'adminChannelsPath',
  'adminChannelPricingPath',
  'adminChannelMonitorPath',
  'adminSubscriptionsPath',
  'adminAccountsPath',
  'adminAnnouncementsPath',
  'adminProxiesPath',
  'adminRiskControlPath',
  'adminRedeemPath',
  'adminPromoCodesPath',
  'adminAffiliatesPath',
  'adminAffiliateOverviewPath',
  'adminAffiliateRulesPath',
  'adminAffiliateCodesPath',
  'adminAffiliateInvitesPath',
  'adminAffiliateRebatesPath',
  'adminAffiliateTransfersPath',
  'adminOrdersRootPath',
  'adminOrdersDashboardPath',
  'adminPaymentPlansPath',
  'adminUsagePath',
]

export function resolveAdminSidebarRouteDefaults(
  rawAuthShellConfig?: string,
  runtimeLocale?: unknown,
): AdminSidebarRouteDefaults {
  const defaults = readAdminSidebarRouteOverrides(rawAuthShellConfig, runtimeLocale)
  return {
    adminOpsPath: defaults.adminOpsPath || FALLBACK_ADMIN_SIDEBAR_ROUTE_DEFAULTS.adminOpsPath,
    adminUsersPath: defaults.adminUsersPath || FALLBACK_ADMIN_SIDEBAR_ROUTE_DEFAULTS.adminUsersPath,
    adminGroupsPath: defaults.adminGroupsPath || FALLBACK_ADMIN_SIDEBAR_ROUTE_DEFAULTS.adminGroupsPath,
    adminChannelsPath: defaults.adminChannelsPath || FALLBACK_ADMIN_SIDEBAR_ROUTE_DEFAULTS.adminChannelsPath,
    adminChannelPricingPath:
      defaults.adminChannelPricingPath || FALLBACK_ADMIN_SIDEBAR_ROUTE_DEFAULTS.adminChannelPricingPath,
    adminChannelMonitorPath:
      defaults.adminChannelMonitorPath || FALLBACK_ADMIN_SIDEBAR_ROUTE_DEFAULTS.adminChannelMonitorPath,
    adminSubscriptionsPath:
      defaults.adminSubscriptionsPath || FALLBACK_ADMIN_SIDEBAR_ROUTE_DEFAULTS.adminSubscriptionsPath,
    adminAccountsPath: defaults.adminAccountsPath || FALLBACK_ADMIN_SIDEBAR_ROUTE_DEFAULTS.adminAccountsPath,
    adminAnnouncementsPath:
      defaults.adminAnnouncementsPath || FALLBACK_ADMIN_SIDEBAR_ROUTE_DEFAULTS.adminAnnouncementsPath,
    adminProxiesPath: defaults.adminProxiesPath || FALLBACK_ADMIN_SIDEBAR_ROUTE_DEFAULTS.adminProxiesPath,
    adminRiskControlPath:
      defaults.adminRiskControlPath || FALLBACK_ADMIN_SIDEBAR_ROUTE_DEFAULTS.adminRiskControlPath,
    adminRedeemPath: defaults.adminRedeemPath || FALLBACK_ADMIN_SIDEBAR_ROUTE_DEFAULTS.adminRedeemPath,
    adminPromoCodesPath:
      defaults.adminPromoCodesPath || FALLBACK_ADMIN_SIDEBAR_ROUTE_DEFAULTS.adminPromoCodesPath,
    adminAffiliatesPath: defaults.adminAffiliatesPath || FALLBACK_ADMIN_SIDEBAR_ROUTE_DEFAULTS.adminAffiliatesPath,
    adminAffiliateOverviewPath:
      defaults.adminAffiliateOverviewPath || FALLBACK_ADMIN_SIDEBAR_ROUTE_DEFAULTS.adminAffiliateOverviewPath,
    adminAffiliateRulesPath:
      defaults.adminAffiliateRulesPath || FALLBACK_ADMIN_SIDEBAR_ROUTE_DEFAULTS.adminAffiliateRulesPath,
    adminAffiliateCodesPath:
      defaults.adminAffiliateCodesPath || FALLBACK_ADMIN_SIDEBAR_ROUTE_DEFAULTS.adminAffiliateCodesPath,
    adminAffiliateInvitesPath:
      defaults.adminAffiliateInvitesPath || FALLBACK_ADMIN_SIDEBAR_ROUTE_DEFAULTS.adminAffiliateInvitesPath,
    adminAffiliateRebatesPath:
      defaults.adminAffiliateRebatesPath || FALLBACK_ADMIN_SIDEBAR_ROUTE_DEFAULTS.adminAffiliateRebatesPath,
    adminAffiliateTransfersPath:
      defaults.adminAffiliateTransfersPath || FALLBACK_ADMIN_SIDEBAR_ROUTE_DEFAULTS.adminAffiliateTransfersPath,
    adminOrdersRootPath:
      defaults.adminOrdersRootPath || FALLBACK_ADMIN_SIDEBAR_ROUTE_DEFAULTS.adminOrdersRootPath,
    adminOrdersDashboardPath:
      defaults.adminOrdersDashboardPath || FALLBACK_ADMIN_SIDEBAR_ROUTE_DEFAULTS.adminOrdersDashboardPath,
    adminPaymentPlansPath:
      defaults.adminPaymentPlansPath || FALLBACK_ADMIN_SIDEBAR_ROUTE_DEFAULTS.adminPaymentPlansPath,
    adminUsagePath: defaults.adminUsagePath || FALLBACK_ADMIN_SIDEBAR_ROUTE_DEFAULTS.adminUsagePath,
  }
}

function readAdminSidebarRouteOverrides(
  rawAuthShellConfig: string | undefined,
  runtimeLocale: unknown,
): Partial<AdminSidebarRouteDefaults> {
  if (!rawAuthShellConfig?.trim()) {
    return {}
  }

  try {
    const parsed = JSON.parse(rawAuthShellConfig) as unknown
    if (!isRecord(parsed)) {
      return {}
    }

    const localized = pickLocalizedAuthShellConfig(parsed, resolveRuntimeLocale(runtimeLocale))
    if (!localized || !isRecord(localized.defaults)) {
      return {}
    }

    const overrides: Partial<AdminSidebarRouteDefaults> = {}
    for (const key of adminSidebarRouteKeys) {
      const path = readInternalPath(localized.defaults[key])
      if (path) {
        overrides[key] = path
      }
    }
    return overrides
  } catch {
    return {}
  }
}

function pickLocalizedAuthShellConfig(
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

function readInternalPath(value: unknown): string | undefined {
  if (typeof value !== 'string') return undefined
  const trimmed = value.trim()
  if (!trimmed || !trimmed.startsWith('/') || trimmed.startsWith('//') || trimmed.includes('://')) {
    return undefined
  }
  if (trimmed.includes('\n') || trimmed.includes('\r')) {
    return undefined
  }
  return trimmed
}
