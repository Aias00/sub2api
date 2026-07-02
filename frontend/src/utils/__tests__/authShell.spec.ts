import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

import { authShellLabelKeys, renderAuthShellText, resolveAuthShellConfig, resolveAuthShellLabels } from '../authShell'

const authShellSource = readFileSync('src/utils/authShell.ts', 'utf8')

describe('authShell', () => {
  it('resolves configured locale labels and renders template parameters', () => {
    const labels = resolveAuthShellLabels(
      JSON.stringify({
        zh: {
          labels: {
            welcomeBack: '配置欢迎',
            allRightsReserved: '配置版权。',
            signUpToStart: '注册以开始使用 {siteName}',
            signInWithProvider: '用 {providerName} 继续',
            oauthFlowCreateAccountTitle: '配置完成 {providerName} 账号',
            oauthCallbackHint: '配置回调提示',
            oauthCallbackTitle: '配置 OAuth 回调',
            oauthCallbackPasswordOptionalHint: '配置 {providerName} 可选密码',
            oauthFlowProfileDetailsTitle: '配置使用 {providerName} 资料',
            oauthFlowBindCurrentAccount: '配置绑定当前账户',
            oauthFlowTotpHint: '配置 {account} 绑定 {providerName}',
            forgotPasswordTitle: '配置找回标题',
            emailVerifyTitle: '配置邮箱验证标题',
            emailVerifyResendCountdown: '配置 {countdown} 秒后重发',
            resetPasswordTitle: '配置重置标题',
            passwordResetSuccess: '配置重置成功',
            providerCallbackTitle: '配置 {providerName} 登录标题',
            providerCallbackProcessing: '配置 {providerName} 登录处理中',
            providerInvitationRequired: '配置 {providerName} 邀请码提示',
            resetEmailSent: '配置已发送',
            totpLoginTitle: '配置两步验证',
            agreementUpdatedAt: '配置条款已于 {date} 更新',
            resendCountdown: '{countdown} 秒后重发',
            optional: '可配置可选',
            ignored: '不应出现',
          },
        },
      }),
      'zh-CN',
    )

    expect(labels.welcomeBack).toBe('配置欢迎')
    expect(labels.allRightsReserved).toBe('配置版权。')
    expect(labels.optional).toBe('可配置可选')
    expect(labels.totpLoginTitle).toBe('配置两步验证')
    expect(labels.oauthCallbackHint).toBe('配置回调提示')
    expect(labels.oauthCallbackTitle).toBe('配置 OAuth 回调')
    expect(labels.oauthFlowProfileDetailsTitle).toBe('配置使用 {providerName} 资料')
    expect(labels.oauthFlowBindCurrentAccount).toBe('配置绑定当前账户')
    expect(labels.forgotPasswordTitle).toBe('配置找回标题')
    expect(labels.emailVerifyTitle).toBe('配置邮箱验证标题')
    expect(labels.resetPasswordTitle).toBe('配置重置标题')
    expect(labels.passwordResetSuccess).toBe('配置重置成功')
    expect(labels.providerCallbackTitle).toBe('配置 {providerName} 登录标题')
    expect(labels.providerCallbackProcessing).toBe('配置 {providerName} 登录处理中')
    expect(labels.providerInvitationRequired).toBe('配置 {providerName} 邀请码提示')
    expect(labels.resetEmailSent).toBe('配置已发送')
    expect(labels).not.toHaveProperty('ignored')
    expect(renderAuthShellText(labels, 'signUpToStart', { siteName: 'Cloudbase' })).toBe('注册以开始使用 Cloudbase')
    expect(renderAuthShellText(labels, 'signInWithProvider', { providerName: 'GitHub' })).toBe('用 GitHub 继续')
    expect(renderAuthShellText(labels, 'oauthFlowCreateAccountTitle', { providerName: 'DingTalk' })).toBe('配置完成 DingTalk 账号')
    expect(renderAuthShellText(labels, 'oauthCallbackPasswordOptionalHint', { providerName: 'Google' })).toBe('配置 Google 可选密码')
    expect(renderAuthShellText(labels, 'providerCallbackTitle', { providerName: 'OIDC' })).toBe('配置 OIDC 登录标题')
    expect(renderAuthShellText(labels, 'providerInvitationRequired', { providerName: 'LinuxDo' })).toBe('配置 LinuxDo 邀请码提示')
    expect(renderAuthShellText(labels, 'oauthFlowTotpHint', { account: 'a***@example.com', providerName: 'OIDC' })).toBe('配置 a***@example.com 绑定 OIDC')
    expect(renderAuthShellText(labels, 'agreementUpdatedAt', { date: '2026-05-18' })).toBe('配置条款已于 2026-05-18 更新')
    expect(renderAuthShellText(labels, 'emailVerifyResendCountdown', { countdown: 60 })).toBe('配置 60 秒后重发')
    expect(renderAuthShellText(labels, 'resendCountdown', { countdown: 30 })).toBe('30 秒后重发')
  })

  it('resolves labels from public auth shell config without adding local defaults', () => {
    const labels = resolveAuthShellLabels(
      JSON.stringify({
        en: {
          labels: {
            welcomeBack: 'Configured welcome',
            signIn: 'Configured sign in',
            allRightsReserved: 'Configured rights',
          },
        },
      }),
      'en',
    )

    expect(labels.welcomeBack).toBe('Configured welcome')
    expect(labels.signIn).toBe('Configured sign in')
    expect(labels.allRightsReserved).toBe('Configured rights')
    expect(labels.emailLabel).toBeUndefined()
  })

  it('resolves auth redirect defaults from public auth shell config', () => {
    const config = resolveAuthShellConfig(
      JSON.stringify({
        en: {
          defaults: {
            defaultRedirectPath: '/configured-dashboard',
            bindRedirectPath: '/configured-profile',
            homePath: '/configured-home',
            loginPath: '/configured-login',
            registerPath: '/configured-register',
            forgotPasswordPath: '/configured-forgot-password',
            emailVerifyPath: '/configured-email-verify',
            apiKeysPath: '/configured-keys',
            usagePath: '/configured-usage',
            availableChannelsPath: '/configured-available-channels',
            availableGroupsPath: '/configured-available-groups',
            subscriptionsPath: '/configured-subscriptions',
            purchasePath: '/configured-purchase',
            paymentResultPath: '/configured-payment-result',
            ordersPath: '/configured-orders',
            redeemPath: '/configured-redeem',
            affiliatePath: '/configured-affiliate',
            profilePath: '/configured-profile-page',
            adminRedirectPath: '/configured-admin',
            adminRuntimeSettingsPath: '/configured-admin-runtime-settings',
            adminSettingsPath: '/configured-admin-settings',
          },
          labels: {
            signIn: 'Configured sign in',
          },
        },
      }),
      'en',
    )

    expect(config.labels.signIn).toBe('Configured sign in')
    expect(config.defaults).toEqual({
      defaultRedirectPath: '/configured-dashboard',
      bindRedirectPath: '/configured-profile',
      homePath: '/configured-home',
      loginPath: '/configured-login',
      registerPath: '/configured-register',
      forgotPasswordPath: '/configured-forgot-password',
      emailVerifyPath: '/configured-email-verify',
      apiKeysPath: '/configured-keys',
      usagePath: '/configured-usage',
      availableChannelsPath: '/configured-available-channels',
      availableGroupsPath: '/configured-available-groups',
      subscriptionsPath: '/configured-subscriptions',
      purchasePath: '/configured-purchase',
      paymentResultPath: '/configured-payment-result',
      ordersPath: '/configured-orders',
      redeemPath: '/configured-redeem',
      affiliatePath: '/configured-affiliate',
      profilePath: '/configured-profile-page',
      adminRedirectPath: '/configured-admin',
      adminRuntimeSettingsPath: '/configured-admin-runtime-settings',
      adminSettingsPath: '/configured-admin-settings',
    })
  })

  it('filters unsafe auth redirect defaults', () => {
    const config = resolveAuthShellConfig(
      JSON.stringify({
        en: {
          defaults: {
            defaultRedirectPath: 'https://evil.example/dashboard',
            bindRedirectPath: '//evil.example/profile',
            homePath: 'https://evil.example/home',
            loginPath: 'login',
            registerPath: 'https://evil.example/register',
            forgotPasswordPath: '//evil.example/forgot-password',
            emailVerifyPath: '/email-verify\nbad',
            apiKeysPath: 'https://evil.example/keys',
            usagePath: '//evil.example/usage',
            availableChannelsPath: '/available\nchannels',
            availableGroupsPath: 'available-groups',
            subscriptionsPath: 'https://evil.example/subscriptions',
            purchasePath: '//evil.example/purchase',
            paymentResultPath: 'https://evil.example/payment-result',
            ordersPath: 'orders',
            redeemPath: '/redeem\rbad',
            affiliatePath: 'https://evil.example/affiliate',
            profilePath: '//evil.example/profile',
            adminRedirectPath: '/admin\nLocation',
            adminRuntimeSettingsPath: 'https://evil.example/admin/runtime-settings',
            adminSettingsPath: 'https://evil.example/admin/settings',
          },
        },
      }),
      'en',
    )

    expect(config.defaults).toEqual({
      defaultRedirectPath: undefined,
      bindRedirectPath: undefined,
      homePath: undefined,
      loginPath: undefined,
      registerPath: undefined,
      forgotPasswordPath: undefined,
      emailVerifyPath: undefined,
      apiKeysPath: undefined,
      usagePath: undefined,
      availableChannelsPath: undefined,
      availableGroupsPath: undefined,
      subscriptionsPath: undefined,
      purchasePath: undefined,
      paymentResultPath: undefined,
      ordersPath: undefined,
      redeemPath: undefined,
      affiliatePath: undefined,
      profilePath: undefined,
      adminRedirectPath: undefined,
      adminRuntimeSettingsPath: undefined,
      adminSettingsPath: undefined,
    })
  })

  it('returns no labels when auth shell config is invalid', () => {
    const labels = resolveAuthShellLabels('{bad json', 'zh')

    expect(labels).toEqual({})
    expect(renderAuthShellText(labels, 'createAccount')).toBe('')
    expect(renderAuthShellText(labels, 'passwordHint', { count: 8 })).toBe('')
  })

  it('does not embed default auth shell labels in the frontend parser', () => {
    expect(authShellLabelKeys).toContain('welcomeBack')
    expect(authShellLabelKeys).toContain('allRightsReserved')
    expect(authShellLabelKeys).toContain('sendCode')
    expect(authShellLabelKeys).toContain('signInWithProvider')
    expect(authShellLabelKeys).toContain('oauthCallbackHint')
    expect(authShellLabelKeys).toContain('oauthCallbackTitle')
    expect(authShellLabelKeys).toContain('oauthCallbackInvalidTitle')
    expect(authShellLabelKeys).toContain('oauthCallbackSubmitRegistration')
    expect(authShellLabelKeys).toContain('oauthFlowProfileDetailsTitle')
    expect(authShellLabelKeys).toContain('oauthFlowBindExistingAccount')
    expect(authShellLabelKeys).toContain('oauthFlowBindCurrentAccount')
    expect(authShellLabelKeys).toContain('oauthFlowBindCurrentAccountDescription')
    expect(authShellLabelKeys).toContain('oauthFlowCreateAccountTitle')
    expect(authShellLabelKeys).toContain('oauthFlowTotpHint')
    expect(authShellLabelKeys).toContain('forgotPasswordTitle')
    expect(authShellLabelKeys).toContain('emailVerifyTitle')
    expect(authShellLabelKeys).toContain('emailVerifySubmit')
    expect(authShellLabelKeys).toContain('emailVerifyResendCountdown')
    expect(authShellLabelKeys).toContain('resetPasswordTitle')
    expect(authShellLabelKeys).toContain('resetPassword')
    expect(authShellLabelKeys).toContain('passwordResetSuccess')
    expect(authShellLabelKeys).toContain('providerCallbackTitle')
    expect(authShellLabelKeys).toContain('providerCallbackProcessing')
    expect(authShellLabelKeys).toContain('providerCallbackHint')
    expect(authShellLabelKeys).toContain('providerInvitationRequired')
    expect(authShellLabelKeys).toContain('providerCompletingRegistration')
    expect(authShellLabelKeys).toContain('providerCompleteRegistration')
    expect(authShellLabelKeys).toContain('newPasswordPlaceholder')
    expect(authShellLabelKeys).toContain('sendResetLink')
    expect(authShellLabelKeys).toContain('resetEmailSent')
    expect(authShellLabelKeys).toContain('totpLoginTitle')
    expect(authShellLabelKeys).toContain('totpCancel')
    expect(authShellLabelKeys).toContain('agreementAcceptAndContinue')
    expect(authShellLabelKeys).toContain('agreementUpdatedAt')
    expect(authShellLabelKeys).toContain('verificationCodeHint')
    expect(authShellSource).not.toContain('DEFAULT_AUTH_SHELL_LABELS')
    expect(authShellSource).not.toContain('AuthShellLabels = Record<string, string>')
    expect(authShellSource).not.toContain("welcomeBack: '欢迎回来'")
    expect(authShellSource).not.toContain("welcomeBack: 'Welcome Back'")
    expect(authShellSource).not.toContain("allRightsReserved: '保留所有权利。'")
    expect(authShellSource).not.toContain("allRightsReserved: 'All rights reserved.'")
    expect(authShellSource).not.toContain('labels[key] || key')
  })
})
