import { describe, expect, it } from 'vitest'

import {
  buildCreditsPurchaseRoute,
  formatCreditsRatio,
  parseCreditsPerBalance,
  renderCreditsBalanceLabel,
  renderCreditsPerBalance,
  resolveCreditsActionsDescription,
  resolveCreditsActionsTitle,
  resolveCreditsOrdersLabel,
  resolveCreditsOrdersPath,
  resolveCreditsPurchasePath,
  resolveCreditsRechargeLabel,
} from '../creditsRuntime'
import type { CreditsCopy, CreditsShellConfig } from '@/utils/creditsShell'

const copy: CreditsCopy = {
  eyebrow: '',
  title: '',
  description: '',
  purchase: '',
  orders: '',
  credits: '',
  cloudbaseBalance: '',
  conversion: 'Default conversion {creditsPerBalance}',
  balanceLabel: 'Balance {balance}',
  actionsTitle: 'Default actions title',
  actionsDescription: 'Default actions description',
  recharge: 'Default recharge',
  viewOrders: 'Default orders',
}

describe('creditsRuntime', () => {
  it('normalizes the legacy credits-per-balance setting to the unified 1:1 unit', () => {
    expect(parseCreditsPerBalance('12')).toBe(1)
    expect(parseCreditsPerBalance(' 12.5 ')).toBe(1)
    expect(parseCreditsPerBalance('0')).toBe(1)
    expect(parseCreditsPerBalance('bad')).toBe(1)
  })

  it('formats ratio and template placeholders', () => {
    expect(formatCreditsRatio(12)).toBe('12')
    expect(formatCreditsRatio(12.5)).toBe('12.5')
    expect(renderCreditsPerBalance('1 = {creditsPerBalance}', 1)).toBe('1 = 1')
    expect(renderCreditsBalanceLabel('Balance: {balance}', '2.50')).toBe('Balance: 2.50')
  })

  it('resolves shell overrides with fallback copy', () => {
    const shellConfig: CreditsShellConfig = {
      labels: copy,
      defaults: {
        purchasePath: '/purchase',
        ordersPath: '/orders',
      },
      actions: {
        title: 'Configured actions title',
      },
      buttons: {
        recharge: 'Configured recharge',
      },
    }

    expect(resolveCreditsActionsTitle(shellConfig, copy)).toBe('Configured actions title')
    expect(resolveCreditsActionsDescription(shellConfig, copy)).toBe('Default actions description')
    expect(resolveCreditsRechargeLabel(shellConfig, copy)).toBe('Configured recharge')
    expect(resolveCreditsOrdersLabel(shellConfig, copy)).toBe('Default orders')
    expect(resolveCreditsPurchasePath(shellConfig)).toBe('/purchase')
    expect(resolveCreditsOrdersPath(shellConfig)).toBe('/orders')
  })

  it('builds purchase routes without embedding defaults in the view', () => {
    expect(buildCreditsPurchaseRoute('')).toBeNull()
    expect(buildCreditsPurchaseRoute('/purchase')).toBe('/purchase')
    expect(buildCreditsPurchaseRoute('/purchase', 'recharge')).toEqual({
      path: '/purchase',
      query: { tab: 'recharge' },
    })
  })
})
