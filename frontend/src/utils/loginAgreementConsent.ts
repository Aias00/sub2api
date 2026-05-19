const LOGIN_AGREEMENT_STORAGE_KEY = 'sub2api_login_agreement_consent'

export interface LoginAgreementAcceptancePayload {
  agreement_accepted: true
  agreement_revision: string
}

interface LoginAgreementConsentRecord {
  revision?: string
  accepted_at?: string
}

export function readStoredLoginAgreementConsent(): LoginAgreementConsentRecord | null {
  try {
    const raw = localStorage.getItem(LOGIN_AGREEMENT_STORAGE_KEY)
    if (!raw) {
      return null
    }
    return JSON.parse(raw) as LoginAgreementConsentRecord
  } catch {
    return null
  }
}

export function hasAcceptedLoginAgreement(revision: string): boolean {
  if (!revision) {
    return false
  }
  return readStoredLoginAgreementConsent()?.revision === revision
}

export function persistLoginAgreementAcceptance(revision: string): void {
  if (!revision) {
    return
  }
  localStorage.setItem(
    LOGIN_AGREEMENT_STORAGE_KEY,
    JSON.stringify({
      revision,
      accepted_at: new Date().toISOString()
    })
  )
}

export function clearLoginAgreementAcceptance(): void {
  localStorage.removeItem(LOGIN_AGREEMENT_STORAGE_KEY)
}

export function buildLoginAgreementAcceptancePayload(): Partial<LoginAgreementAcceptancePayload> {
  const revision = readStoredLoginAgreementConsent()?.revision?.trim()
  if (!revision) {
    return {}
  }
  return {
    agreement_accepted: true,
    agreement_revision: revision
  }
}
