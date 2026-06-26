import { describe, expect, it } from 'vitest'
import { resolvePricingShellConfig } from '../pricingShell'

describe('resolvePricingShellConfig', () => {
  it('resolves localized labels, button, and groups', () => {
    const shell = resolvePricingShellConfig(
      JSON.stringify({
        en: {
          button: { title: 'Configured buy' },
          defaults: {
            promptsPath: '/configured-prompts',
            purchasePath: '/configured-purchase',
          },
          groups: [
            { name: 'one-time', title: 'One-time' },
            { name: 'subscription', title: 'Subscription' },
            'ignored',
          ],
          labels: {
            eyebrow: 'Plans',
            rechargeCta: 'Top up',
            ignored: 'ignored',
          },
        },
      }),
      'en',
    )

    expect(shell.button?.title).toBe('Configured buy')
    expect(shell.defaults).toEqual({
      promptsPath: '/configured-prompts',
      purchasePath: '/configured-purchase',
    })
    expect(shell.groups).toEqual([
      { name: 'one-time', title: 'One-time' },
      { name: 'subscription', title: 'Subscription' },
    ])
    expect(shell.labels.eyebrow).toBe('Plans')
    expect(shell.labels.rechargeCta).toBe('Top up')
    expect(shell.labels.title).toBe('')
  })

  it('filters unsafe route defaults instead of owning local pricing paths', () => {
    const shell = resolvePricingShellConfig(
      JSON.stringify({
        en: {
          defaults: {
            promptsPath: 'https://evil.example/prompts',
            purchasePath: '//evil.example/purchase',
          },
        },
      }),
      'en',
    )

    expect(shell.defaults).toEqual({
      promptsPath: undefined,
      purchasePath: undefined,
    })
  })

  it('falls back to root config when locale branch is missing', () => {
    const shell = resolvePricingShellConfig(
      JSON.stringify({
        labels: {
          title: 'Root pricing',
        },
      }),
      'zh',
    )

    expect(shell.labels.title).toBe('Root pricing')
  })

  it('returns empty labels for invalid JSON', () => {
    const shell = resolvePricingShellConfig('{bad json', 'en')

    expect(shell.labels.title).toBe('')
    expect(shell.labels.rechargeCta).toBe('')
    expect(shell.labels.month).toBe('')
  })
})
