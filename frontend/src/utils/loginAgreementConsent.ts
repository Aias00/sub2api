const LOGIN_AGREEMENT_STORAGE_KEY = 'sub2api_login_agreement_consent'

export interface LoginAgreementAcceptancePayload {
  agreement_accepted: true
  agreement_revision: string
}

interface LoginAgreementConsentEntry {
  revision?: string
  accepted_at?: string
}

interface LegacyLoginAgreementConsentRecord extends LoginAgreementConsentEntry {
  subject?: string
}

interface LoginAgreementConsentState {
  anonymous?: LoginAgreementConsentEntry
  subjects?: Record<string, LoginAgreementConsentEntry>
}

function normalizeSubject(subject?: string | null): string {
  return String(subject || '').trim().toLowerCase()
}

function cleanupState(state: LoginAgreementConsentState): LoginAgreementConsentState | null {
  const subjects: Record<string, LoginAgreementConsentEntry> = {}
  for (const [subject, entry] of Object.entries(state.subjects || {})) {
    const normalizedSubject = normalizeSubject(subject)
    const revision = String(entry?.revision || '').trim()
    if (!normalizedSubject || !revision) {
      continue
    }
    subjects[normalizedSubject] = {
      revision,
      accepted_at: entry.accepted_at
    }
  }

  const anonymousRevision = String(state.anonymous?.revision || '').trim()
  const cleaned: LoginAgreementConsentState = {}
  if (anonymousRevision) {
    cleaned.anonymous = {
      revision: anonymousRevision,
      accepted_at: state.anonymous?.accepted_at
    }
  }
  if (Object.keys(subjects).length > 0) {
    cleaned.subjects = subjects
  }

  if (!cleaned.anonymous && !cleaned.subjects) {
    return null
  }
  return cleaned
}

function parseStoredState(raw: string): LoginAgreementConsentState | null {
  const parsed = JSON.parse(raw) as LoginAgreementConsentState | LegacyLoginAgreementConsentRecord | null
  if (!parsed || typeof parsed !== 'object') {
    return null
  }

  // Backward compatibility with the old single-record shape.
  if ('revision' in parsed || 'subject' in parsed) {
    const legacy = parsed as LegacyLoginAgreementConsentRecord
    const revision = String(legacy.revision || '').trim()
    if (!revision) {
      return null
    }
    const subject = normalizeSubject(legacy.subject)
    if (subject) {
      return cleanupState({
        subjects: {
          [subject]: {
            revision,
            accepted_at: legacy.accepted_at
          }
        }
      })
    }
    return cleanupState({
      anonymous: {
        revision,
        accepted_at: legacy.accepted_at
      }
    })
  }

  return cleanupState(parsed as LoginAgreementConsentState)
}

function readStoredState(): LoginAgreementConsentState | null {
  try {
    const raw = localStorage.getItem(LOGIN_AGREEMENT_STORAGE_KEY)
    if (!raw) {
      return null
    }
    return parseStoredState(raw)
  } catch {
    return null
  }
}

function persistState(state: LoginAgreementConsentState | null): void {
  if (!state) {
    localStorage.removeItem(LOGIN_AGREEMENT_STORAGE_KEY)
    return
  }
  localStorage.setItem(LOGIN_AGREEMENT_STORAGE_KEY, JSON.stringify(state))
}

function getConsentEntry(subject?: string | null): LoginAgreementConsentEntry | null {
  const state = readStoredState()
  if (!state) {
    return null
  }
  const normalizedSubject = normalizeSubject(subject)
  if (normalizedSubject) {
    return state.subjects?.[normalizedSubject] || null
  }
  return state.anonymous || null
}

export function hasAcceptedLoginAgreement(revision: string, subject?: string | null): boolean {
  const normalizedRevision = String(revision || '').trim()
  if (!normalizedRevision) {
    return false
  }
  return getConsentEntry(subject)?.revision === normalizedRevision
}

export function persistLoginAgreementAcceptance(revision: string, subject?: string | null): void {
  const normalizedRevision = String(revision || '').trim()
  if (!normalizedRevision) {
    return
  }

  const state = readStoredState() || {}
  const entry: LoginAgreementConsentEntry = {
    revision: normalizedRevision,
    accepted_at: new Date().toISOString()
  }
  const normalizedSubject = normalizeSubject(subject)
  if (normalizedSubject) {
    state.subjects = {
      ...(state.subjects || {}),
      [normalizedSubject]: entry
    }
  } else {
    state.anonymous = entry
  }
  persistState(cleanupState(state))
}

export function bindAnonymousLoginAgreementAcceptanceToSubject(subject?: string | null): void {
  const normalizedSubject = normalizeSubject(subject)
  const state = readStoredState()
  if (!normalizedSubject || !state?.anonymous?.revision) {
    return
  }

  state.subjects = {
    ...(state.subjects || {}),
    [normalizedSubject]: {
      revision: state.anonymous.revision,
      accepted_at: state.anonymous.accepted_at || new Date().toISOString()
    }
  }
  delete state.anonymous
  persistState(cleanupState(state))
}

export function clearLoginAgreementAcceptance(subject?: string | null): void {
  const state = readStoredState()
  if (!state) {
    return
  }
  const normalizedSubject = normalizeSubject(subject)
  if (normalizedSubject) {
    if (state.subjects) {
      delete state.subjects[normalizedSubject]
    }
  } else {
    delete state.anonymous
  }
  persistState(cleanupState(state))
}

export function clearAllLoginAgreementAcceptance(): void {
  localStorage.removeItem(LOGIN_AGREEMENT_STORAGE_KEY)
}

export function buildLoginAgreementAcceptancePayload(subject?: string | null): Partial<LoginAgreementAcceptancePayload> {
  const revision = String(getConsentEntry(subject)?.revision || '').trim()
  if (!revision) {
    return {}
  }
  return {
    agreement_accepted: true,
    agreement_revision: revision
  }
}
