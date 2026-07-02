import type { AuthShellConfig } from '@/utils/authShell'
import { resolveAuthShellConfig } from '@/utils/authShell'

export type AuthRouteDefaults = {
  homePath: string
  loginPath: string
  registerPath: string
  forgotPasswordPath: string
  emailVerifyPath: string
  apiKeysPath: string
  usagePath: string
  availableChannelsPath: string
  availableGroupsPath: string
  subscriptionsPath: string
  purchasePath: string
  paymentResultPath: string
  ordersPath: string
  redeemPath: string
  affiliatePath: string
  profilePath: string
  userRedirectPath: string
  adminRedirectPath: string
  adminRuntimeSettingsPath: string
  adminSettingsPath: string
}

export const FALLBACK_AUTH_ROUTE_DEFAULTS: AuthRouteDefaults = {
  homePath: '/home',
  loginPath: '/login',
  registerPath: '/register',
  forgotPasswordPath: '/forgot-password',
  emailVerifyPath: '/email-verify',
  apiKeysPath: '/keys',
  usagePath: '/usage',
  availableChannelsPath: '/available-channels',
  availableGroupsPath: '/available-groups',
  subscriptionsPath: '/subscriptions',
  purchasePath: '/purchase',
  paymentResultPath: '/payment/result',
  ordersPath: '/orders',
  redeemPath: '/redeem',
  affiliatePath: '/affiliate',
  profilePath: '/profile',
  userRedirectPath: '/dashboard',
  adminRedirectPath: '/admin/dashboard',
  adminRuntimeSettingsPath: '/admin/runtime-settings',
  adminSettingsPath: '/admin/settings',
}

export function resolveAuthRouteDefaults(
  rawAuthShellConfig?: string,
  runtimeLocale = typeof navigator === 'undefined' ? '' : navigator.language,
): AuthRouteDefaults {
  const defaults = resolveAuthShellConfig(rawAuthShellConfig, runtimeLocale).defaults
  return resolveAuthRouteDefaultsFromShellDefaults(defaults)
}

export function resolveAuthRouteDefaultsFromShellDefaults(
  defaults: AuthShellConfig['defaults'] = {},
): AuthRouteDefaults {
  return {
    homePath: defaults.homePath || FALLBACK_AUTH_ROUTE_DEFAULTS.homePath,
    loginPath: defaults.loginPath || FALLBACK_AUTH_ROUTE_DEFAULTS.loginPath,
    registerPath: defaults.registerPath || FALLBACK_AUTH_ROUTE_DEFAULTS.registerPath,
    forgotPasswordPath: defaults.forgotPasswordPath || FALLBACK_AUTH_ROUTE_DEFAULTS.forgotPasswordPath,
    emailVerifyPath: defaults.emailVerifyPath || FALLBACK_AUTH_ROUTE_DEFAULTS.emailVerifyPath,
    apiKeysPath: defaults.apiKeysPath || FALLBACK_AUTH_ROUTE_DEFAULTS.apiKeysPath,
    usagePath: defaults.usagePath || FALLBACK_AUTH_ROUTE_DEFAULTS.usagePath,
    availableChannelsPath: defaults.availableChannelsPath || FALLBACK_AUTH_ROUTE_DEFAULTS.availableChannelsPath,
    availableGroupsPath: defaults.availableGroupsPath || FALLBACK_AUTH_ROUTE_DEFAULTS.availableGroupsPath,
    subscriptionsPath: defaults.subscriptionsPath || FALLBACK_AUTH_ROUTE_DEFAULTS.subscriptionsPath,
    purchasePath: defaults.purchasePath || FALLBACK_AUTH_ROUTE_DEFAULTS.purchasePath,
    paymentResultPath: defaults.paymentResultPath || FALLBACK_AUTH_ROUTE_DEFAULTS.paymentResultPath,
    ordersPath: defaults.ordersPath || FALLBACK_AUTH_ROUTE_DEFAULTS.ordersPath,
    redeemPath: defaults.redeemPath || FALLBACK_AUTH_ROUTE_DEFAULTS.redeemPath,
    affiliatePath: defaults.affiliatePath || FALLBACK_AUTH_ROUTE_DEFAULTS.affiliatePath,
    profilePath: defaults.profilePath || FALLBACK_AUTH_ROUTE_DEFAULTS.profilePath,
    userRedirectPath: defaults.defaultRedirectPath || FALLBACK_AUTH_ROUTE_DEFAULTS.userRedirectPath,
    adminRedirectPath: defaults.adminRedirectPath || FALLBACK_AUTH_ROUTE_DEFAULTS.adminRedirectPath,
    adminRuntimeSettingsPath: defaults.adminRuntimeSettingsPath || FALLBACK_AUTH_ROUTE_DEFAULTS.adminRuntimeSettingsPath,
    adminSettingsPath: defaults.adminSettingsPath || FALLBACK_AUTH_ROUTE_DEFAULTS.adminSettingsPath,
  }
}

export function resolveRoleHomeRedirect(isAdmin: boolean, defaults = FALLBACK_AUTH_ROUTE_DEFAULTS): string {
  return isAdmin ? defaults.adminRedirectPath : defaults.userRedirectPath
}

export function resolveCompletedSetupRedirectPath(
  isAuthenticated: boolean,
  isAdmin: boolean,
  defaults = FALLBACK_AUTH_ROUTE_DEFAULTS,
): string {
  if (!isAuthenticated) {
    return defaults.loginPath
  }

  return resolveRoleHomeRedirect(isAdmin, defaults)
}
