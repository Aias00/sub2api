import { describe, expect, it } from 'vitest'
import { affiliateLabelKeys, renderAffiliateShellText, resolveAffiliateShellLabels } from '../affiliateShell'

describe('affiliate shell helpers', () => {
  it('resolves affiliate labels from localized shell config', () => {
    const labels = resolveAffiliateShellLabels(
      JSON.stringify({
        en: {
          labels: {
            title: 'Affiliate',
            tipRebate: 'Rebate {rate}',
            ignored: 'ignored',
          },
        },
      }),
      'en-US',
    )

    expect(affiliateLabelKeys).toContain('transferSuccess')
    expect(labels.title).toBe('Affiliate')
    expect(labels.tipRebate).toBe('Rebate {rate}')
    expect(labels.loadFailed).toBeUndefined()
  })

  it('renders affiliate labels with placeholders', () => {
    const labels = {
      tipRebate: 'Rebate {rate}',
      transferSuccess: 'Transferred {amount}',
    }

    expect(renderAffiliateShellText(labels, 'tipRebate', { rate: '12.5%' })).toBe('Rebate 12.5%')
    expect(renderAffiliateShellText(labels, 'transferSuccess', { amount: '$10.00' })).toBe('Transferred $10.00')
    expect(renderAffiliateShellText(labels, 'loadFailed')).toBe('')
  })
})
