import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

const source = readFileSync(resolve(process.cwd(), 'src/components/auth/OidcOAuthSection.vue'), 'utf8')
const emailOAuthSource = readFileSync(resolve(process.cwd(), 'src/components/auth/EmailOAuthButtons.vue'), 'utf8')
const dingtalkOAuthSource = readFileSync(resolve(process.cwd(), 'src/components/auth/DingTalkOAuthSection.vue'), 'utf8')
const linuxdoOAuthSource = readFileSync(resolve(process.cwd(), 'src/components/auth/LinuxDoOAuthSection.vue'), 'utf8')
const wechatOAuthSource = readFileSync(resolve(process.cwd(), 'src/components/auth/WechatOAuthSection.vue'), 'utf8')

describe('OidcOAuthSection runtime provider name', () => {
  it('does not invent a frontend-local OIDC provider name', () => {
    expect(source).not.toContain("providerName: 'OIDC'")
    expect(source).not.toContain("return name || 'OIDC'")
    expect(source).not.toContain("|| 'O'")
    expect(source).toContain("providerName: ''")
    expect(source).toContain("return name || ''")
  })

  it('keeps OAuth start URL construction out of component-local env fallbacks', () => {
    for (const componentSource of [
      source,
      emailOAuthSource,
      dingtalkOAuthSource,
      linuxdoOAuthSource,
      wechatOAuthSource,
    ]) {
      expect(componentSource).not.toContain('import.meta.env.VITE_API_BASE_URL')
      expect(componentSource).not.toContain("|| '/dashboard'")
      expect(componentSource).toContain('buildApiUrl')
      expect(componentSource).toContain('resolveRouteAuthRedirect')
    }
  })

  it('uses auth shell labels for OAuth entry button and divider copy', () => {
    for (const componentSource of [
      source,
      emailOAuthSource,
      dingtalkOAuthSource,
      linuxdoOAuthSource,
      wechatOAuthSource,
    ]) {
      expect(componentSource).toContain('shellLabels')
      expect(componentSource).toContain('signInWithProvider')
      expect(componentSource).toContain('oauthAlternativeMethods')
      expect(componentSource).not.toContain("t('auth.oauthOrContinue')")
      expect(componentSource).not.toContain("t('auth.emailOAuth.signIn')")
      expect(componentSource).not.toContain("t('auth.oidc.signIn')")
      expect(componentSource).not.toContain("t('auth.linuxdo.signIn')")
      expect(componentSource).not.toContain("t('auth.dingtalk.signIn')")
    }
  })

  it('uses auth shell labels for WeChat OAuth availability hints', () => {
    expect(wechatOAuthSource).toContain("authText('wechatProviderName')")
    expect(wechatOAuthSource).toContain("authText('wechatSystemBrowserOnly')")
    expect(wechatOAuthSource).toContain("authText('wechatBrowserOnly')")
    expect(wechatOAuthSource).toContain("authText('wechatNativeAppOnly')")
    expect(wechatOAuthSource).toContain("authText('wechatNotConfigured')")
    expect(wechatOAuthSource).not.toContain("t('auth.wechatProviderName')")
    expect(wechatOAuthSource).not.toContain('auth.oauthFlow.wechatSystemBrowserOnly')
    expect(wechatOAuthSource).not.toContain('auth.oauthFlow.wechatBrowserOnly')
    expect(wechatOAuthSource).not.toContain('auth.oauthFlow.wechatNativeAppOnly')
    expect(wechatOAuthSource).not.toContain('auth.oauthFlow.wechatNotConfigured')
  })
})
