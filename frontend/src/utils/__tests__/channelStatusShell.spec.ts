import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

import {
  DEFAULT_CHANNEL_STATUS_REFRESH_INTERVAL_SECONDS,
  resolveChannelStatusShellConfig,
  resolveChannelStatusShellLabels,
} from '../channelStatusShell'

const channelStatusShellSource = readFileSync('src/utils/channelStatusShell.ts', 'utf8')

describe('resolveChannelStatusShellLabels', () => {
  it('resolves configured locale labels without adding frontend defaults', () => {
    const labels = resolveChannelStatusShellLabels(
      JSON.stringify({
        zh: {
          labels: {
            refreshTitle: '配置刷新',
            detailTitle: '配置详情',
            windowTab: {
              '7d': '配置 7 天',
            },
            overall: {
              operational: '配置正常',
            },
            detailColumns: {
              model: '配置模型',
            },
          },
        },
      }),
      'zh-CN',
    )

    expect(labels.refreshTitle).toBe('配置刷新')
    expect(labels.detailTitle).toBe('配置详情')
    expect(labels.windowTab['7d']).toBe('配置 7 天')
    expect(labels.windowTab['15d']).toBe('')
    expect(labels.overall.operational).toBe('配置正常')
    expect(labels.overall.degraded).toBe('')
    expect(labels.detailColumns.model).toBe('配置模型')
    expect(labels.detailColumns.latestStatus).toBe('')
  })

  it('returns empty structural labels for invalid config', () => {
    const labels = resolveChannelStatusShellLabels('{bad json', 'en')

    expect(labels.refreshTitle).toBe('')
    expect(labels.detailTitle).toBe('')
    expect(labels.windowTab['7d']).toBe('')
    expect(labels.overall.operational).toBe('')
    expect(labels.detailColumns.model).toBe('')
  })

  it('resolves refresh interval defaults from public settings', () => {
    const config = resolveChannelStatusShellConfig(
      JSON.stringify({
        zh: {
          defaults: {
            refreshIntervalSeconds: 120,
          },
        },
      }),
      'zh-CN',
    )

    expect(config.defaults.refreshIntervalSeconds).toBe(120)
  })

  it('falls back to the built-in refresh interval for invalid defaults', () => {
    expect(resolveChannelStatusShellConfig('{bad json', 'en').defaults.refreshIntervalSeconds)
      .toBe(DEFAULT_CHANNEL_STATUS_REFRESH_INTERVAL_SECONDS)
    expect(resolveChannelStatusShellConfig(JSON.stringify({
      en: {
        defaults: {
          refreshIntervalSeconds: 5,
        },
      },
    }), 'en').defaults.refreshIntervalSeconds).toBe(DEFAULT_CHANNEL_STATUS_REFRESH_INTERVAL_SECONDS)
    expect(resolveChannelStatusShellConfig(JSON.stringify({
      en: {
        defaults: {
          refreshIntervalSeconds: 'never',
        },
      },
    }), 'en').defaults.refreshIntervalSeconds).toBe(DEFAULT_CHANNEL_STATUS_REFRESH_INTERVAL_SECONDS)
  })

  it('does not embed default channel status copy in the frontend parser', () => {
    expect(channelStatusShellSource).not.toContain('EMPTY_LABELS')
    expect(channelStatusShellSource).not.toContain('DEFAULT_LABELS')
    expect(channelStatusShellSource).not.toContain("refreshTitle: '刷新'")
    expect(channelStatusShellSource).not.toContain("detailTitle: '渠道详情'")
    expect(channelStatusShellSource).not.toContain("refreshTitle: 'Refresh'")
    expect(channelStatusShellSource).not.toContain("detailTitle: 'Channel Detail'")
  })
})
