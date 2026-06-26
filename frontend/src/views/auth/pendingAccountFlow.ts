export type PendingAccountAction =
  | 'none'
  | 'choose_account_action'
  | 'choice'
  | 'create_account'
  | 'bind_login'

export type StandardPendingAccountState = {
  action: 'none' | 'choose_account_action' | 'create_account' | 'bind_login'
  pendingAccountEmail: string
  bindLoginEmail: string
  bindLoginPassword: string
  canReturnToCreateAccount: boolean
}

export type StandardTotpChallengeState = {
  action: 'none'
  needsInvitation: false
  needsAdoptionConfirmation: false
  needsTotpChallenge: true
  totpTempToken: string
  totpCode: string
  totpError: string
  totpUserEmailMasked: string
  isProcessing: false
}

export type PendingAccountEmailSource = {
  pending_email?: string
  existing_account_email?: string
  compat_email?: string
  resolved_email?: string
  email?: string
}

export function normalizePendingState(value: string | null | undefined): string {
  return value?.trim().toLowerCase() || ''
}

export function extractPendingAccountEmail(
  completion: PendingAccountEmailSource,
  extraFallbacks: Array<string | null | undefined> = [],
): string {
  return (
    completion.pending_email ||
    completion.existing_account_email ||
    completion.compat_email ||
    completion.resolved_email ||
    completion.email ||
    extraFallbacks.find((value) => typeof value === 'string' && value.trim()) ||
    ''
  ).trim()
}

export function extractStandardPendingAccountEmail(
  completion: PendingAccountEmailSource,
): string {
  return extractPendingAccountEmail(completion)
}

export function resolvePendingAccountAction(
  value: string | null | undefined,
  options: {
    chooseAliases?: string[]
    createAliases?: string[]
    bindAliases?: string[]
    chooseResult?: PendingAccountAction
  } = {},
): PendingAccountAction {
  const raw = normalizePendingState(value)
  const chooseAliases = options.chooseAliases || [
    'choice',
    'choose_account_action_required',
    'choose_account_action',
    'choose_account',
    'choose',
  ]
  const createAliases = options.createAliases || [
    'email_required',
    'create_account_required',
    'create_account',
  ]
  const bindAliases = options.bindAliases || [
    'bind_login_required',
    'bind_login',
    'existing_account',
    'existing_account_required',
    'existing_account_binding_required',
    'adopt_existing_user_by_email',
  ]
  const chooseResult = options.chooseResult || 'choose_account_action'

  if (chooseAliases.includes(raw)) {
    return chooseResult
  }
  if (createAliases.includes(raw)) {
    return 'create_account'
  }
  if (bindAliases.includes(raw)) {
    return 'bind_login'
  }
  return 'none'
}

export function resolveStandardPendingAccountAction(
  value: string | null | undefined,
): PendingAccountAction {
  return resolvePendingAccountAction(value)
}

export function buildStandardPendingAccountState(
  action: 'none' | 'choose_account_action' | 'create_account' | 'bind_login',
  email: string,
): StandardPendingAccountState {
  if (action === 'choose_account_action') {
    return {
      action,
      pendingAccountEmail: email,
      bindLoginEmail: email,
      bindLoginPassword: '',
      canReturnToCreateAccount: false,
    }
  }

  if (action === 'create_account') {
    return {
      action,
      pendingAccountEmail: email,
      bindLoginEmail: '',
      bindLoginPassword: '',
      canReturnToCreateAccount: true,
    }
  }

  if (action === 'bind_login') {
    return {
      action,
      pendingAccountEmail: email,
      bindLoginEmail: email,
      bindLoginPassword: '',
      canReturnToCreateAccount: false,
    }
  }

  return {
    action: 'none',
    pendingAccountEmail: email,
    bindLoginEmail: '',
    bindLoginPassword: '',
    canReturnToCreateAccount: false,
  }
}

export function buildStandardPendingAccountStateForChoiceMode(
  action: 'none' | 'choice' | 'create_account' | 'bind_login',
  email: string,
): {
  action: 'none' | 'choice' | 'create_account' | 'bind_login'
  pendingAccountEmail: string
  bindLoginEmail: string
  bindLoginPassword: string
} {
  if (action === 'create_account') {
    return {
      action,
      pendingAccountEmail: email,
      bindLoginEmail: '',
      bindLoginPassword: '',
    }
  }

  if (action === 'bind_login') {
    return {
      action,
      pendingAccountEmail: email,
      bindLoginEmail: email,
      bindLoginPassword: '',
    }
  }

  if (action === 'choice') {
    return {
      action,
      pendingAccountEmail: email,
      bindLoginEmail: '',
      bindLoginPassword: '',
    }
  }

  return {
    action: 'none',
    pendingAccountEmail: email,
    bindLoginEmail: '',
    bindLoginPassword: '',
  }
}

export function buildStandardTotpChallengeState(
  completion: { requires_2fa?: boolean; temp_token?: string; user_email_masked?: string },
): StandardTotpChallengeState | null {
  if (completion.requires_2fa !== true || !completion.temp_token) {
    return null
  }

  return {
    action: 'none',
    needsInvitation: false,
    needsAdoptionConfirmation: false,
    needsTotpChallenge: true,
    totpTempToken: completion.temp_token,
    totpCode: '',
    totpError: '',
    totpUserEmailMasked: completion.user_email_masked || '',
    isProcessing: false,
  }
}

export function buildBindLoginSwitchState(
  currentBindLoginEmail: string,
  currentPendingAccountEmail: string,
  nextEmail?: string,
) {
  return {
    action: 'bind_login' as const,
    bindLoginEmail: currentBindLoginEmail.trim() || nextEmail?.trim() || currentPendingAccountEmail.trim(),
    bindLoginPassword: '',
    accountActionError: '',
    canReturnToCreateAccount: true,
  }
}

export function buildCreateAccountSwitchState(
  currentPendingAccountEmail: string,
  currentBindLoginEmail: string,
) {
  return {
    action: 'create_account' as const,
    pendingAccountEmail: currentPendingAccountEmail.trim() || currentBindLoginEmail.trim(),
    accountActionError: '',
  }
}

export function isCreateAccountRecoveryError(error: unknown): boolean {
  const data = (error as {
    response?: {
      data?: {
        reason?: string
        error?: string
        code?: string
        step?: string
        intent?: string
      }
    }
  }).response?.data

  const states = [data?.reason, data?.error, data?.code, data?.step, data?.intent]
    .map((value) => value?.trim().toLowerCase())
    .filter((value): value is string => Boolean(value))

  return states.includes('email_exists') ||
    states.includes('bind_login_required') ||
    states.includes('bind_login') ||
    states.includes('adopt_existing_user_by_email') ||
    states.includes('existing_account_required') ||
    states.includes('existing_account_binding_required')
}
