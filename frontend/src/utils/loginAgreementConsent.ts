const LOGIN_AGREEMENT_STORAGE_KEY = 'cloudbase_login_agreement_attempt'

export interface LoginAgreementAcceptancePayload {
  agreement_accepted: true
  agreement_revision: string
}

interface LoginAgreementConsentAttempt {
  revision?: string
  accepted_at?: string
}

function getStorage(): Storage | null {
  try {
    return window.sessionStorage
  } catch {
    return null
  }
}

function readAttempt(): LoginAgreementConsentAttempt | null {
  const storage = getStorage()
  if (!storage) return null
  try {
    const raw = storage.getItem(LOGIN_AGREEMENT_STORAGE_KEY)
    if (!raw) return null
    const parsed = JSON.parse(raw) as LoginAgreementConsentAttempt | null
    const revision = String(parsed?.revision || '').trim()
    if (!revision) return null
    return {
      revision,
      accepted_at: parsed?.accepted_at,
    }
  } catch {
    return null
  }
}

function persistAttempt(attempt: LoginAgreementConsentAttempt | null): void {
  const storage = getStorage()
  if (!storage) return
  if (!attempt?.revision) {
    storage.removeItem(LOGIN_AGREEMENT_STORAGE_KEY)
    return
  }
  storage.setItem(LOGIN_AGREEMENT_STORAGE_KEY, JSON.stringify(attempt))
}

export function hasAcceptedLoginAgreement(revision: string, _subject?: string | null): boolean {
  const normalizedRevision = String(revision || '').trim()
  if (!normalizedRevision) {
    return false
  }
  return readAttempt()?.revision === normalizedRevision
}

export function persistLoginAgreementAcceptance(revision: string, _subject?: string | null): void {
  const normalizedRevision = String(revision || '').trim()
  if (!normalizedRevision) {
    return
  }
  persistAttempt({
    revision: normalizedRevision,
    accepted_at: new Date().toISOString(),
  })
}

export function bindAnonymousLoginAgreementAcceptanceToSubject(_subject?: string | null): void {
  // 条款确认只在当前登录尝试内生效，不再绑定到账号或跨会话复用。
}

export function clearLoginAgreementAcceptance(_subject?: string | null): void {
  persistAttempt(null)
}

export function clearAllLoginAgreementAcceptance(): void {
  persistAttempt(null)
}

export function buildLoginAgreementAcceptancePayload(_subject?: string | null): Partial<LoginAgreementAcceptancePayload> {
  const revision = String(readAttempt()?.revision || '').trim()
  if (!revision) {
    return {}
  }
  return {
    agreement_accepted: true,
    agreement_revision: revision,
  }
}
