import { describe, expect, it } from 'vitest'
import { formatLegalDocumentTemplate, resolveLegalDocumentCopy } from '../legalDocumentShell'

describe('legal document shell helpers', () => {
  it('resolves legal document copy from localized public settings', () => {
    const copy = resolveLegalDocumentCopy(
      JSON.stringify({
        zh: {
          labels: {
            login: '登录',
            agreementLabel: '登录条款',
            updatedAt: '更新于 {date}',
            ignored: 'ignored',
          },
        },
      }),
      'zh-CN',
    )

    expect(copy.login).toBe('登录')
    expect(copy.agreementLabel).toBe('登录条款')
    expect(copy.updatedAt).toBe('更新于 {date}')
    expect(copy.missingTitle).toBe('')
  })

  it('formats legal document templates outside the view', () => {
    expect(formatLegalDocumentTemplate('Updated {date}', { date: '2026-06-19' })).toBe('Updated 2026-06-19')
    expect(formatLegalDocumentTemplate('Updated {missing}', {})).toBe('Updated ')
  })
})
