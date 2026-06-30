import type { CreditsCopy, CreditsShellConfig } from '@/utils/creditsShell'

export function parseCreditsPerBalance(_value: unknown): number {
  // Credits are a display name for the backend balance ledger. Keep the legacy
  // setting accepted for compatibility, but normalize it to the single 1:1 unit.
  return 1
}

export function formatCreditsRatio(value: number): string {
  return Number.isInteger(value) ? String(value) : value.toFixed(2).replace(/0+$/, '').replace(/\.$/, '')
}

export function renderCreditsPerBalance(template: string, creditsPerBalance: number): string {
  return template.includes('{creditsPerBalance}')
    ? template.split('{creditsPerBalance}').join(formatCreditsRatio(creditsPerBalance))
    : template
}

export function renderCreditsBalanceLabel(template: string, formattedBalance: string): string {
  return template.includes('{balance}')
    ? template.split('{balance}').join(formattedBalance)
    : `${template} ${formattedBalance}`.trim()
}

export function resolveCreditsActionsTitle(shellConfig: CreditsShellConfig, copy: CreditsCopy): string {
  return shellConfig.actions?.title?.trim() || copy.actionsTitle
}

export function resolveCreditsActionsDescription(shellConfig: CreditsShellConfig, copy: CreditsCopy): string {
  return shellConfig.actions?.description?.trim() || copy.actionsDescription
}

export function resolveCreditsRechargeLabel(shellConfig: CreditsShellConfig, copy: CreditsCopy): string {
  return shellConfig.buttons?.recharge?.trim() || copy.recharge
}

export function resolveCreditsOrdersLabel(shellConfig: CreditsShellConfig, copy: CreditsCopy): string {
  return shellConfig.buttons?.orders?.trim() || copy.viewOrders
}

export function resolveCreditsPurchasePath(shellConfig: CreditsShellConfig): string {
  return shellConfig.defaults?.purchasePath?.trim() || ''
}

export function resolveCreditsOrdersPath(shellConfig: CreditsShellConfig): string {
  return shellConfig.defaults?.ordersPath?.trim() || ''
}

export function buildCreditsPurchaseRoute(path: string, tab?: string) {
  if (!path) return null
  return tab ? { path, query: { tab } } : path
}
