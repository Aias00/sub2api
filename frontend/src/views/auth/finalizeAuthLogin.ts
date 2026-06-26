import type { OAuthTokenResponse } from '@/api/auth'
import { persistOAuthTokenContext } from '@/api/auth'
import { clearAllAffiliateReferralCodes } from '@/utils/oauthAffiliate'

type AuthStoreLike = {
  setToken(token: string): Promise<unknown>
}

type AppStoreLike = {
  showSuccess(message: string): void
}

type RouterLike = {
  replace(target: string): Promise<unknown> | unknown
}

export async function finalizeAuthLoginSuccess(options: {
  tokenResponse: OAuthTokenResponse
  redirect: string
  authStore: AuthStoreLike
  appStore: AppStoreLike
  router: RouterLike
  successMessage: string
  beforeRedirect?: () => void
}) {
  persistOAuthTokenContext(options.tokenResponse)
  await options.authStore.setToken(options.tokenResponse.access_token)
  options.beforeRedirect?.()
  clearAllAffiliateReferralCodes()
  options.appStore.showSuccess(options.successMessage)
  await options.router.replace(options.redirect)
}
