import { describe, expect, it } from 'vitest'

import en from '../locales/en'
import zh from '../locales/zh'

describe('locale coverage for recent public/auth and admin additions', () => {
  it('covers legal document copy in both locales and keeps model plaza copy in shell config', () => {
    expect(zh.legalDocument.loadFailedTitle).toBe('文档加载失败')
    expect(en.legalDocument.loadFailedTitle).toBe('Failed to load document')

    expect('modelsPlaza' in zh).toBe(false)
    expect('modelsPlaza' in en).toBe(false)
  })

  it('covers login agreement warnings and DingTalk auth copy in both locales', () => {
    expect(zh.auth.loginAgreementMustAcceptLogin).toContain('同意最新条款')
    expect(en.auth.loginAgreementMustAcceptLogin).toContain('accept the latest agreement')

    expect(zh.auth.dingtalk.signIn).toBe('使用钉钉登录')
    expect(en.auth.dingtalk.signIn).toBe('Continue with DingTalk')
    expect(zh.profile.authBindings.providers.dingtalk).toBe('钉钉')
    expect(en.profile.authBindings.providers.dingtalk).toBe('DingTalk')
  })

  it('covers targeted admin account, settings, and user locale additions', () => {
    expect(zh.admin.accounts.openai.responsesMode).toBe('文本接口模式')
    expect(en.admin.accounts.openai.responsesMode).toBe('Text endpoint mode')
    expect(zh.admin.accounts.syncUpstreamModels).toBe('同步上游模型')
    expect(en.admin.accounts.syncUpstreamModels).toBe('Sync upstream models')

    expect(zh.admin.settings.dingtalk.title).toBe('钉钉登录')
    expect(en.admin.settings.dingtalk.title).toBe('DingTalk Sign-In')
    expect(zh.admin.settings.emailTemplates.title).toBe('邮件模板')
    expect(en.admin.settings.subscriptionExpiryNotify.title).toBe('Subscription expiry reminders')

    expect(zh.admin.users.platformQuota.title).toBe('用户平台限额')
    expect(en.admin.users.platformQuota.title).toBe('User platform quotas')
    expect(zh.admin.users.columns.balancePlatformQuota).toBe('余额 / 平台配额')
    expect(en.admin.users.columns.balancePlatformQuota).toBe('Balance / Platform Quota')
  })
})
