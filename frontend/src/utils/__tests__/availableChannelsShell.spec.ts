import { describe, expect, it } from 'vitest'
import {
  availableChannelsLabelKeys,
  resolveAvailableChannelsShellLabels,
  resolveConfiguredAvailableChannelsShellLabels,
} from '../availableChannelsShell'

const emptyLabels = {
  searchPlaceholder: '',
  refreshTitle: '',
  columnName: '',
  columnDescription: '',
  pricingBillingMode: '',
  pricingInputPrice: '',
}
const configuredLabels = {
  searchPlaceholder: 'Search',
  refreshTitle: 'Refresh',
  columnName: 'Channel',
  columnDescription: 'Description',
  pricingBillingMode: 'Billing',
  pricingInputPrice: 'Input',
}
const keys = Object.keys(emptyLabels) as Array<keyof typeof emptyLabels>

describe('resolveAvailableChannelsShellLabels', () => {
  it('resolves labels and nested column labels', () => {
    const labels = resolveAvailableChannelsShellLabels(
      JSON.stringify({
        zh: {
          labels: {
            searchPlaceholder: '配置搜索',
            ignored: 'ignored',
            columns: {
              name: '配置渠道名',
              description: '配置描述',
              ignored: 'ignored',
            },
            pricing: {
              billingMode: '配置计费',
              inputPrice: '配置输入',
              ignored: 'ignored',
            },
          },
        },
      }),
      'zh-CN',
      keys,
    )

    expect(labels).toEqual({
      searchPlaceholder: '配置搜索',
      refreshTitle: '',
      columnName: '配置渠道名',
      columnDescription: '配置描述',
      pricingBillingMode: '配置计费',
      pricingInputPrice: '配置输入',
    })
  })

  it('returns empty labels for missing or invalid public settings config', () => {
    expect(resolveAvailableChannelsShellLabels(undefined, 'zh', keys)).toEqual(emptyLabels)
    expect(resolveAvailableChannelsShellLabels('{bad json', 'zh', keys)).toEqual(emptyLabels)
  })

  it('ignores caller-side default copy', () => {
    const labels = resolveAvailableChannelsShellLabels(
      JSON.stringify({ en: { labels: { searchPlaceholder: configuredLabels.searchPlaceholder } } }),
      'en',
      keys,
    )

    expect(labels.searchPlaceholder).toBe(configuredLabels.searchPlaceholder)
    expect(labels.columnName).toBe('')
    expect(labels.pricingBillingMode).toBe('')
  })

  it('centralizes the full available channels shell label contract', () => {
    const labels = resolveConfiguredAvailableChannelsShellLabels(
      JSON.stringify({
        en: {
          labels: {
            searchPlaceholder: 'Search channels',
            exclusive: 'Exclusive',
            columns: {
              supportedModels: 'Models',
            },
            pricing: {
              billingModeToken: 'Per token',
            },
          },
        },
      }),
      'en-US',
    )

    expect(availableChannelsLabelKeys).toContain('columnSupportedModels')
    expect(availableChannelsLabelKeys).toContain('pricingBillingModeToken')
    expect(labels.searchPlaceholder).toBe('Search channels')
    expect(labels.exclusive).toBe('Exclusive')
    expect(labels.columnSupportedModels).toBe('Models')
    expect(labels.pricingBillingModeToken).toBe('Per token')
    expect(labels.loadError).toBe('')
  })
})
