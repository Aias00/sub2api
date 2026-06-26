import { describe, expect, it } from 'vitest'
import { redeemLabelKeys, renderRedeemShellText, resolveRedeemShellLabels } from '../redeemShell'

describe('redeem shell helpers', () => {
  it('resolves redeem labels from localized shell config', () => {
    const labels = resolveRedeemShellLabels(
      JSON.stringify({
        en: {
          labels: {
            currentBalance: 'Balance',
            subscriptionDays: '{days} days',
            ignored: 'ignored',
          },
        },
      }),
      'en-US',
    )

    expect(redeemLabelKeys).toContain('currentBalance')
    expect(labels.currentBalance).toBe('Balance')
    expect(labels.subscriptionDays).toBe('{days} days')
    expect(labels.unknown).toBeUndefined()
  })

  it('renders configured redeem labels with placeholders', () => {
    const labels = {
      subscriptionDays: '{days} configured days',
      currentBalance: 'Balance',
    }

    expect(renderRedeemShellText(labels, 'subscriptionDays', { days: 30 })).toBe('30 configured days')
    expect(renderRedeemShellText(labels, 'currentBalance')).toBe('Balance')
    expect(renderRedeemShellText(labels, 'failedToRedeem')).toBe('')
  })
})
