export type EmailOAuthProvider = 'github' | 'google'

export type EmailOAuthPendingCompletion = {
  error?: string
  provider?: string
  redirect?: string
  email?: string
  resolved_email?: string
  invitation_required?: boolean
  access_token?: string
  refresh_token?: string
  expires_in?: number
  token_type?: string
}

export function resolveEmailOAuthProvider(value: unknown): EmailOAuthProvider | null {
  if (typeof value !== 'string') {
    return null
  }
  const normalized = value.trim().toLowerCase()
  return normalized === 'github' || normalized === 'google'
    ? normalized
    : null
}

export function isEmailOAuthTokenResponse(
  value: Partial<EmailOAuthPendingCompletion>,
): value is Required<Pick<EmailOAuthPendingCompletion, 'access_token'>> & Partial<EmailOAuthPendingCompletion> {
  return typeof value.access_token === 'string' && value.access_token.trim() !== ''
}

export function deriveEmailOAuthRegistrationState(
  completion: EmailOAuthPendingCompletion,
  defaultRedirectPath: string,
) {
  const provider = resolveEmailOAuthProvider(completion.provider)
  const redirect = (completion.redirect || defaultRedirectPath).trim() || defaultRedirectPath
  const invitationRequired =
    completion.error === 'invitation_required' || completion.invitation_required === true
  const requiresRegistrationCompletion =
    completion.error === 'invitation_required' ||
    completion.error === 'registration_completion_required'
  const registrationEmail = String(completion.resolved_email || completion.email || '').trim()

  return {
    provider,
    redirect,
    invitationRequired,
    requiresRegistrationCompletion,
    registrationEmail,
  }
}
