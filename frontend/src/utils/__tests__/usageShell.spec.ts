import { describe, expect, it } from 'vitest'
import {
  DEFAULT_USAGE_API_KEY_PAGE_SIZE,
  DEFAULT_USAGE_DATE_RANGE_DAYS,
  DEFAULT_USAGE_EXPORT_PAGE_SIZE,
  renderUsageShellText,
  resolveUsageShellConfig,
  resolveUsageShellLabels,
  usageShellLabelKeys,
} from '../usageShell'

describe('usage shell helpers', () => {
  it('resolves usage labels from localized shell config', () => {
    const labels = resolveUsageShellLabels(
      JSON.stringify({
        zh: {
          labels: {
            totalRequests: '总请求',
            exportSuccess: '导出成功',
            ignored: 'ignored',
          },
        },
      }),
      'zh-CN',
    )

    expect(usageShellLabelKeys).toContain('exportFailed')
    expect(labels.totalRequests).toBe('总请求')
    expect(labels.exportSuccess).toBe('导出成功')
    expect(labels.failedToLoad).toBeUndefined()
  })

  it('renders empty text for missing usage labels', () => {
    expect(renderUsageShellText({ totalRequests: 'Total' }, 'totalRequests')).toBe('Total')
    expect(renderUsageShellText({}, 'failedToLoad')).toBe('')
  })

  it('resolves behavior defaults from localized shell config', () => {
    const config = resolveUsageShellConfig(
      JSON.stringify({
        zh: {
          defaults: {
            dateRangeDays: 14,
            apiKeyPageSize: 37,
            exportPageSize: 37,
          },
        },
      }),
      'zh-CN',
    )

    expect(config.defaults.dateRangeDays).toBe(14)
    expect(config.defaults.apiKeyPageSize).toBe(37)
    expect(config.defaults.exportPageSize).toBe(37)
  })

  it('falls back to built-in behavior defaults for invalid defaults', () => {
    const invalidJSONDefaults = resolveUsageShellConfig('{bad json', 'zh-CN').defaults
    expect(invalidJSONDefaults.dateRangeDays).toBe(DEFAULT_USAGE_DATE_RANGE_DAYS)
    expect(invalidJSONDefaults.apiKeyPageSize).toBe(DEFAULT_USAGE_API_KEY_PAGE_SIZE)
    expect(invalidJSONDefaults.exportPageSize).toBe(DEFAULT_USAGE_EXPORT_PAGE_SIZE)

    const invalidValueDefaults = resolveUsageShellConfig(JSON.stringify({
      zh: {
        defaults: {
          dateRangeDays: 0,
          apiKeyPageSize: 1001,
          exportPageSize: 0,
        },
      },
    }), 'zh-CN').defaults

    expect(invalidValueDefaults.dateRangeDays).toBe(DEFAULT_USAGE_DATE_RANGE_DAYS)
    expect(invalidValueDefaults.apiKeyPageSize).toBe(DEFAULT_USAGE_API_KEY_PAGE_SIZE)
    expect(invalidValueDefaults.exportPageSize).toBe(DEFAULT_USAGE_EXPORT_PAGE_SIZE)
  })
})
