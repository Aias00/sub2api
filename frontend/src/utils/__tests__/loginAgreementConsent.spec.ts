import { beforeEach, describe, expect, it } from 'vitest'

import {
  buildLoginAgreementAcceptancePayload,
  clearAllLoginAgreementAcceptance,
  hasAcceptedLoginAgreement,
  persistLoginAgreementAcceptance,
} from '@/utils/loginAgreementConsent'

describe('loginAgreementConsent', () => {
  beforeEach(() => {
    sessionStorage.clear()
  })

  it('stores acceptance only for the current auth attempt', () => {
    persistLoginAgreementAcceptance('rev-1')

    expect(hasAcceptedLoginAgreement('rev-1')).toBe(true)
    expect(hasAcceptedLoginAgreement('rev-1', 'user@example.com')).toBe(true)
    expect(buildLoginAgreementAcceptancePayload('user@example.com')).toEqual({
      agreement_accepted: true,
      agreement_revision: 'rev-1',
    })
  })

  it('ignores malformed or incomplete session state', () => {
    sessionStorage.setItem('sub2api_login_agreement_attempt', JSON.stringify({
      revision: 'legacy-rev',
      accepted_at: '2026-05-22T00:00:00.000Z',
    }))

    expect(hasAcceptedLoginAgreement('legacy-rev')).toBe(true)
    sessionStorage.setItem('sub2api_login_agreement_attempt', JSON.stringify({ accepted_at: '2026-05-22T00:00:00.000Z' }))
    expect(hasAcceptedLoginAgreement('legacy-rev')).toBe(false)
  })

  it('clears all consent records when requested', () => {
    persistLoginAgreementAcceptance('rev-1')
    clearAllLoginAgreementAcceptance()

    expect(hasAcceptedLoginAgreement('rev-1')).toBe(false)
    expect(buildLoginAgreementAcceptancePayload()).toEqual({})
  })
})
