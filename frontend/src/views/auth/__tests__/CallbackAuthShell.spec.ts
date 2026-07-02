import { existsSync, readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

const removedOAuthViewFiles = [
  'OidcCallbackView.vue',
  'LinuxDoCallbackView.vue',
  'DingTalkCallbackView.vue',
  'WechatCallbackView.vue',
  'DingTalkEmailCompletionView.vue',
]

describe('callback auth shell labels', () => {
  it('keeps the shared GitHub/Google OAuth callback on auth shell labels', () => {
    const source = readFileSync(resolve(process.cwd(), 'src/views/auth/OAuthCallbackView.vue'), 'utf8')

    expect(source).toContain('useAuthShellText')
    expect(source).toContain("authText('oauthCallbackTitle')")
    expect(source).toContain("authText('oauthCallbackHint')")
    expect(source).toContain("authText('oauthCallbackInvalidTitle')")
    expect(source).toContain("authText('oauthCallbackSubmitRegistration')")
    expect(source).not.toContain("t('common.processing')")
    expect(source).not.toContain("t('auth.oidc.")
    expect(source).not.toContain("t('auth.linuxdo.")
    expect(source).not.toContain("t('auth.dingtalk.")
    expect(source).not.toContain("t('auth.wechat.")
  })

  it('keeps pending OAuth create-account labels generic', () => {
    const source = readFileSync(
      resolve(process.cwd(), 'src/components/auth/PendingOAuthCreateAccountForm.vue'),
      'utf8',
    )

    expect(source).toContain('useAuthShellText')
    expect(source).toContain("authText('processing')")
    expect(source).not.toContain("t('common.processing')")
    expect(source).not.toContain('common.processing')
  })

  it('does not keep legacy provider-specific callback views', () => {
    for (const file of removedOAuthViewFiles) {
      expect(existsSync(resolve(process.cwd(), `src/views/auth/${file}`)), file).toBe(false)
    }
  })
})
