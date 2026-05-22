import { beforeEach, describe, expect, it } from 'vitest'

import {
  bindAnonymousLoginAgreementAcceptanceToSubject,
  buildLoginAgreementAcceptancePayload,
  clearAllLoginAgreementAcceptance,
  hasAcceptedLoginAgreement,
  persistLoginAgreementAcceptance,
} from '@/utils/loginAgreementConsent'

describe('loginAgreementConsent', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('does not reuse anonymous acceptance for a different email subject', () => {
    persistLoginAgreementAcceptance('rev-1')

    expect(hasAcceptedLoginAgreement('rev-1')).toBe(true)
    expect(hasAcceptedLoginAgreement('rev-1', 'user@example.com')).toBe(false)
    expect(buildLoginAgreementAcceptancePayload('user@example.com')).toEqual({})
  })

  it('binds anonymous acceptance to a concrete subject after login', () => {
    persistLoginAgreementAcceptance('rev-1')
    bindAnonymousLoginAgreementAcceptanceToSubject('user@example.com')

    expect(hasAcceptedLoginAgreement('rev-1')).toBe(false)
    expect(hasAcceptedLoginAgreement('rev-1', 'user@example.com')).toBe(true)
    expect(buildLoginAgreementAcceptancePayload('user@example.com')).toEqual({
      agreement_accepted: true,
      agreement_revision: 'rev-1',
    })
  })

  it('keeps different subjects isolated', () => {
    persistLoginAgreementAcceptance('rev-1', 'first@example.com')
    persistLoginAgreementAcceptance('rev-1', 'second@example.com')

    expect(hasAcceptedLoginAgreement('rev-1', 'first@example.com')).toBe(true)
    expect(hasAcceptedLoginAgreement('rev-1', 'second@example.com')).toBe(true)
    expect(hasAcceptedLoginAgreement('rev-1', 'third@example.com')).toBe(false)
  })

  it('reads old single-record storage as anonymous consent for backward compatibility', () => {
    localStorage.setItem('sub2api_login_agreement_consent', JSON.stringify({
      revision: 'legacy-rev',
      accepted_at: '2026-05-22T00:00:00.000Z',
    }))

    expect(hasAcceptedLoginAgreement('legacy-rev')).toBe(true)
    expect(hasAcceptedLoginAgreement('legacy-rev', 'user@example.com')).toBe(false)
  })

  it('clears all consent records when requested', () => {
    persistLoginAgreementAcceptance('rev-1')
    persistLoginAgreementAcceptance('rev-1', 'user@example.com')
    clearAllLoginAgreementAcceptance()

    expect(hasAcceptedLoginAgreement('rev-1')).toBe(false)
    expect(hasAcceptedLoginAgreement('rev-1', 'user@example.com')).toBe(false)
  })
})
