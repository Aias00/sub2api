import { describe, expect, it } from 'vitest'
import { resolveCreditsShellConfig } from '../creditsShell'

describe('resolveCreditsShellConfig', () => {
  it('resolves labels, actions, buttons, and conversion', () => {
    const shell = resolveCreditsShellConfig(
      JSON.stringify({
        en: {
          defaults: {
            purchasePath: '/configured-purchase',
            ordersPath: '/configured-orders',
          },
          labels: {
            title: 'Wallet',
            purchase: 'Buy balance',
            ignored: 'ignored',
          },
          actions: {
            title: 'Configured actions',
            description: 'Configured action description',
          },
          buttons: {
            recharge: 'Configured recharge',
            orders: 'Configured orders',
          },
          conversion: 'Configured conversion',
        },
      }),
      'en',
    )

    expect(shell.labels.title).toBe('Wallet')
    expect(shell.labels.purchase).toBe('Buy balance')
    expect(shell.labels.credits).toBe('')
    expect(shell.defaults).toEqual({
      purchasePath: '/configured-purchase',
      ordersPath: '/configured-orders',
    })
    expect(shell.actions?.title).toBe('Configured actions')
    expect(shell.buttons?.orders).toBe('Configured orders')
    expect(shell.conversion).toBe('Configured conversion')
  })

  it('filters unsafe route defaults instead of owning local credit paths', () => {
    const shell = resolveCreditsShellConfig(
      JSON.stringify({
        en: {
          defaults: {
            purchasePath: 'https://evil.example/purchase',
            ordersPath: '//evil.example/orders',
          },
        },
      }),
      'en',
    )

    expect(shell.defaults).toEqual({
      purchasePath: undefined,
      ordersPath: undefined,
    })
  })

  it('returns empty labels for invalid JSON', () => {
    const shell = resolveCreditsShellConfig('{bad json', 'en')

    expect(shell.labels.title).toBe('')
    expect(shell.labels.purchase).toBe('')
    expect(shell.labels.viewOrders).toBe('')
  })
})
