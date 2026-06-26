import { describe, expect, it } from 'vitest'

import {
  DEFAULT_KEY_USAGE_DAILY_USAGE_DAYS,
  DEFAULT_KEY_USAGE_DATE_RANGE,
  keyUsageShellLabelKeys,
  renderKeyUsageShellText,
  resolveKeyUsageShellConfig,
  resolveKeyUsageShellLabels,
} from '../keyUsageShell'

describe('keyUsageShell', () => {
  it('resolves key usage shell labels and filters unknown keys', () => {
    const labels = resolveKeyUsageShellLabels(
      JSON.stringify({
        en: {
          labels: {
            title: 'Configured usage',
            ignored: 'should not appear',
          },
        },
      }),
      'en',
    )

    expect(labels.title).toBe('Configured usage')
    expect(labels).not.toHaveProperty('ignored')
  })

  it('renders placeholder values from configured labels', () => {
    expect(renderKeyUsageShellText({ ...emptyLabels(), daysLeft: '{days} days left' }, 'daysLeft', { days: 7 })).toBe(
      '7 days left',
    )
  })

  it('centralizes the key usage shell schema', () => {
    expect(keyUsageShellLabelKeys).toContain('title')
    expect(keyUsageShellLabelKeys).toContain('dailyDetail')
    expect(keyUsageShellLabelKeys).toContain('walletBalance')
  })

  it('resolves key usage defaults from localized shell config', () => {
    expect(resolveKeyUsageShellConfig(JSON.stringify({
      en: {
        defaults: {
          defaultDateRange: '7d',
          dailyUsageDays: 90,
        },
      },
    }), 'en').defaults).toEqual({
      defaultDateRange: '7d',
      dailyUsageDays: 90,
    })

    expect(resolveKeyUsageShellConfig(JSON.stringify({
      en: {
        defaults: {
          defaultDateRange: 'custom',
          dailyUsageDays: 31,
        },
      },
    }), 'en').defaults).toEqual({
      defaultDateRange: DEFAULT_KEY_USAGE_DATE_RANGE,
      dailyUsageDays: DEFAULT_KEY_USAGE_DAILY_USAGE_DAYS,
    })
  })
})

function emptyLabels(): Record<(typeof keyUsageShellLabelKeys)[number], string> {
  return Object.fromEntries(keyUsageShellLabelKeys.map((key) => [key, ''])) as Record<
    (typeof keyUsageShellLabelKeys)[number],
    string
  >
}
