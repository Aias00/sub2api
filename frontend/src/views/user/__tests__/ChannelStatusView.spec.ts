import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { readFileSync } from 'node:fs'
import ChannelStatusView from '../ChannelStatusView.vue'

const channelStatusViewSource = readFileSync('src/views/user/ChannelStatusView.vue', 'utf8')

const listChannelMonitorViews = vi.hoisted(() => vi.fn())
const fetchChannelMonitorDetail = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())
const setAutoRefreshInterval = vi.hoisted(() => vi.fn())
const autoRefreshOptions = vi.hoisted(() => ({
  current: null as null | { defaultInterval?: number },
}))
const appStoreState = vi.hoisted(() => ({
  cachedPublicSettings: null as null | {
    channel_monitor_enabled?: boolean
    channel_status_shell_config?: string
  },
  showError,
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
      locale: { value: 'zh-CN' },
    }),
  }
})

vi.mock('@/api/channelMonitor', () => ({
  list: listChannelMonitorViews,
  status: fetchChannelMonitorDetail,
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => appStoreState,
}))

vi.mock('@/composables/useAutoRefresh', () => ({
  useAutoRefresh: (options: { defaultInterval?: number }) => {
    autoRefreshOptions.current = options
    return {
    enabled: { value: false },
    intervalSeconds: { value: options.defaultInterval || 30 },
    countdown: { value: options.defaultInterval || 30 },
    intervals: [30, 60],
    setEnabled: vi.fn(),
    setInterval: setAutoRefreshInterval,
    start: vi.fn(),
    stop: vi.fn(),
    }
  },
}))

describe('ChannelStatusView', () => {
  beforeEach(() => {
    listChannelMonitorViews.mockReset().mockResolvedValue({
      items: [
        {
          id: 1,
          name: 'Primary Monitor',
          provider: 'openai',
          group_name: 'Default',
          primary_model: 'gpt-5',
          primary_status: 'operational',
          primary_latency_ms: 123,
          primary_ping_latency_ms: 45,
          availability_7d: 99.9,
          extra_models: [{ model: 'gpt-5-mini', status: 'operational', latency_ms: 99 }],
          timeline: [],
        },
      ],
    })
    fetchChannelMonitorDetail.mockReset().mockResolvedValue({
      id: 1,
      name: 'Primary Monitor',
      provider: 'openai',
      group_name: 'Default',
      models: [],
    })
    showError.mockReset()
    setAutoRefreshInterval.mockReset()
    autoRefreshOptions.current = null
    appStoreState.cachedPublicSettings = null
  })

  it('passes Cloudbase channel status shell labels to monitor components', async () => {
    appStoreState.cachedPublicSettings = {
      channel_status_shell_config: JSON.stringify({
        zh: {
          defaults: {
            refreshIntervalSeconds: 120,
          },
          labels: {
            refreshTitle: '配置刷新',
            detailTitle: '配置详情',
            latency: '配置延迟',
            ping: '配置 Ping',
            availabilityPrefix: '配置可用率',
            extraModelsCount: '配置 +{n}',
            emptyTitle: '配置空态标题',
            emptyDescription: '配置空态描述',
            closeDetail: '配置关闭',
            detailLoadError: '配置详情失败',
            windowTab: {
              '7d': '配置 7 天',
              '15d': '配置 15 天',
              '30d': '配置 30 天',
            },
            overall: {
              operational: '配置正常',
              degraded: '配置降级',
            },
            detailColumns: {
              model: '配置模型',
              latestStatus: '配置状态',
            },
          },
        },
      }),
    }

    const wrapper = mount(ChannelStatusView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          MonitorHero: {
            props: ['labels', 'intervalSeconds'],
            template: `
              <section>
                <span>{{ labels.refreshTitle }}</span>
                <span>{{ labels.windowTab['7d'] }}</span>
                <span>{{ labels.overall.operational }}</span>
                <span>interval:{{ intervalSeconds }}</span>
              </section>
            `,
          },
          MonitorCardGrid: {
            props: ['items', 'labels'],
            template: `
              <section>
                <span>{{ labels.emptyTitle }}</span>
                <span>{{ labels.emptyDescription }}</span>
                <span>{{ labels.latency }}</span>
                <span>{{ labels.ping }}</span>
                <span>{{ labels.availabilityPrefix }}</span>
                <span>{{ labels.extraModelsCount }}</span>
                <span v-for="item in items" :key="item.id">{{ item.name }}</span>
              </section>
            `,
          },
          MonitorDetailDialog: {
            props: ['title', 'labels'],
            template: `
              <section>
                <span>{{ title }}</span>
                <span>{{ labels.closeDetail }}</span>
                <span>{{ labels.detailLoadError }}</span>
                <span>{{ labels.detailColumns.model }}</span>
                <span>{{ labels.detailColumns.latestStatus }}</span>
              </section>
            `,
          },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('配置刷新')
    expect(wrapper.text()).toContain('interval:120')
    expect(wrapper.text()).toContain('配置 7 天')
    expect(wrapper.text()).toContain('配置正常')
    expect(wrapper.text()).toContain('配置空态标题')
    expect(wrapper.text()).toContain('配置空态描述')
    expect(wrapper.text()).toContain('配置延迟')
    expect(wrapper.text()).toContain('配置 Ping')
    expect(wrapper.text()).toContain('配置可用率')
    expect(wrapper.text()).toContain('配置 +{n}')
    expect(wrapper.text()).toContain('配置详情')
    expect(wrapper.text()).toContain('配置关闭')
    expect(wrapper.text()).toContain('配置详情失败')
    expect(wrapper.text()).toContain('配置模型')
    expect(wrapper.text()).toContain('配置状态')
    expect(wrapper.text()).toContain('Primary Monitor')
    expect(autoRefreshOptions.current?.defaultInterval).toBe(120)
    expect(setAutoRefreshInterval).not.toHaveBeenCalled()
  })

  it('does not keep the legacy channel status locale section in frontend bundles or route meta', () => {
    const zhLocaleSource = readFileSync('src/i18n/locales/zh.ts', 'utf8')
    const enLocaleSource = readFileSync('src/i18n/locales/en.ts', 'utf8')
    const routerSource = readFileSync('src/router/index.ts', 'utf8')

    for (const source of [zhLocaleSource, enLocaleSource]) {
      expect(source).not.toContain('\n  channelStatus: {')
    }
    expect(routerSource).not.toContain("titleKey: 'nav.channelStatus'")
    expect(channelStatusViewSource).not.toContain('channelStatus.')
    expect(channelStatusViewSource).toContain("from './channelStatusRuntime'")
    expect(channelStatusViewSource).toContain('resolveChannelStatusOverallStatus')
    expect(channelStatusViewSource).toContain('shouldEnsureChannelStatusDetails')
    expect(channelStatusViewSource).not.toContain('DEFAULT_INTERVAL_SECONDS')
    expect(channelStatusViewSource).toContain('refreshIntervalSeconds')
  })
})
