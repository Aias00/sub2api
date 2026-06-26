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
  sub2apiBalance: '',
  conversion: 'Default conversion {creditsPerBalance}',
  balanceLabel: 'Balance {balance}',
  actionsTitle: 'Default actions title',
  actionsDescription: 'Default actions description',
  recharge: 'Default recharge',
  viewOrders: 'Default orders',
}

describe('creditsRuntime', () => {
  it('parses credits-per-balance safely', () => {
    expect(parseCreditsPerBalance('12')).toBe(12)
    expect(parseCreditsPerBalance(' 12.5 ')).toBe(12.5)
    expect(parseCreditsPerBalance('0')).toBe(0)
    expect(parseCreditsPerBalance('bad')).toBe(0)
  })

  it('formats ratio and template placeholders', () => {
    expect(formatCreditsRatio(12)).toBe('12')
    expect(formatCreditsRatio(12.5)).toBe('12.5')
    expect(renderCreditsPerBalance('1 = {creditsPerBalance}', 8)).toBe('1 = 8')
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
