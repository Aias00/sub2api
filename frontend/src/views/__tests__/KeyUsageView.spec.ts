import { describe, expect, it, beforeEach, afterEach, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import { readFileSync } from 'node:fs'

import KeyUsageView from '../KeyUsageView.vue'

const keyUsageViewSource = readFileSync('src/views/KeyUsageView.vue', 'utf8')

const { showInfo, showSuccess, showError, fetchPublicSettings } = vi.hoisted(() => ({
  showInfo: vi.fn(),
  showSuccess: vi.fn(),
  showError: vi.fn(),
  fetchPublicSettings: vi.fn(),
}))
const publicSettings = vi.hoisted(() => ({
  value: null as null | Record<string, unknown>,
}))

const messages: Record<string, string> = {
  'keyUsage.title': 'API Key Usage',
  'keyUsage.subtitle': 'Usage status',
  'keyUsage.placeholder': 'sk-test',
  'keyUsage.query': 'Query',
  'keyUsage.querying': 'Querying...',
  'keyUsage.privacyNote': 'Privacy note',
  'keyUsage.dateRange': 'Date Range:',
  'keyUsage.dateRangeToday': 'Today',
  'keyUsage.dateRange7d': '7 Days',
  'keyUsage.dateRange30d': '30 Days',
  'keyUsage.dateRange90d': '90 Days',
  'keyUsage.dateRangeCustom': 'Custom',
  'keyUsage.apply': 'Apply',
  'keyUsage.used': 'Used',
  'keyUsage.detailInfo': 'Detail Information',
  'keyUsage.tokenStats': 'Token Statistics',
  'keyUsage.dailyDetail': 'Daily Detail',
  'keyUsage.date': 'Date',
  'keyUsage.requests': 'Requests',
  'keyUsage.inputTokens': 'Input Tokens',
  'keyUsage.outputTokens': 'Output Tokens',
  'keyUsage.cacheReadTokens': 'Cache Read',
  'keyUsage.cacheWriteTokens': 'Cache Write',
  'keyUsage.cost': 'Cost',
  'keyUsage.quotaMode': 'Key Quota Mode',
  'keyUsage.walletBalance': 'Wallet Balance',
  'keyUsage.totalQuota': 'Total Quota',
  'keyUsage.limit5h': '5-Hour Limit',
  'keyUsage.limitDaily': 'Daily Limit',
  'keyUsage.limit7d': '7-Day Limit',
  'keyUsage.limitWeekly': 'Weekly Limit',
  'keyUsage.limitMonthly': 'Monthly Limit',
  'keyUsage.remainingQuota': 'Remaining Quota',
  'keyUsage.usedQuota': 'Used Quota',
  'keyUsage.subscriptionType': 'Subscription Type',
  'keyUsage.todayRequests': 'Today Requests',
  'keyUsage.todayInputTokens': 'Today Input',
  'keyUsage.todayOutputTokens': 'Today Output',
  'keyUsage.todayTokens': 'Today Tokens',
  'keyUsage.todayCacheCreation': 'Today Cache Creation',
  'keyUsage.todayCacheRead': 'Today Cache Read',
  'keyUsage.todayCost': 'Today Cost',
  'keyUsage.rpmTpm': 'RPM / TPM',
  'keyUsage.totalRequests': 'Total Requests',
  'keyUsage.totalInputTokens': 'Total Input',
  'keyUsage.totalOutputTokens': 'Total Output',
  'keyUsage.totalTokensLabel': 'Total Tokens',
  'keyUsage.totalCacheCreation': 'Total Cache Creation',
  'keyUsage.totalCacheRead': 'Total Cache Read',
  'keyUsage.totalCost': 'Total Cost',
  'keyUsage.avgDuration': 'Avg Duration',
  'keyUsage.querySuccess': 'Query successful',
  'keyUsage.queryFailed': 'Query failed',
  'keyUsage.queryFailedRetry': 'Query failed, please try again later',
}

function buildKeyUsageShellConfig(
  overrides: Record<string, string> = {},
  defaults: Record<string, unknown> = {},
): string {
  const labels = Object.fromEntries(
    Object.entries(messages)
      .filter(([key]) => key.startsWith('keyUsage.'))
      .map(([key, value]) => [key.replace('keyUsage.', ''), value])
  )

  return JSON.stringify({
    en: {
      labels: {
        ...labels,
        allRightsReserved: 'All rights reserved.',
        docs: 'Docs',
        ...overrides,
      },
      defaults,
    },
  })
}

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => messages[key] ?? key,
      locale: { value: 'en' },
    }),
  }
})

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    cachedPublicSettings: publicSettings.value,
    siteName: 'Cloudbase',
    siteLogo: '',
    docUrl: '',
    publicSettingsLoaded: true,
    fetchPublicSettings,
    showInfo,
    showSuccess,
    showError,
  }),
}))

describe('KeyUsageView daily detail', () => {
  beforeEach(() => {
    showInfo.mockReset()
    showSuccess.mockReset()
    showError.mockReset()
    fetchPublicSettings.mockReset()
    publicSettings.value = {
      key_usage_shell_config: buildKeyUsageShellConfig(),
    }
    localStorage.clear()

    Object.defineProperty(window, 'matchMedia', {
      configurable: true,
      value: vi.fn().mockReturnValue({ matches: false }),
    })
    vi.stubGlobal('requestAnimationFrame', (cb: FrameRequestCallback) => window.setTimeout(() => cb(0), 0))
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({
        mode: 'quota_limited',
        isValid: true,
        status: 'active',
        quota: {
          limit: 10,
          used: 1,
          remaining: 9,
          unit: 'USD',
        },
        usage: {
          today: {
            requests: 1,
            input_tokens: 10,
            output_tokens: 20,
            cache_creation_tokens: 0,
            cache_read_tokens: 0,
            total_tokens: 30,
            actual_cost: 0.01,
          },
          total: {
            requests: 12,
            input_tokens: 100,
            output_tokens: 200,
            cache_creation_tokens: 10,
            cache_read_tokens: 30,
            total_tokens: 340,
            actual_cost: 0.12,
          },
          rpm: 0,
          tpm: 0,
        },
        daily_usage: [
          {
            date: '2026-05-19',
            requests: 12,
            input_tokens: 100,
            output_tokens: 200,
            cache_read_tokens: 30,
            cache_write_tokens: 10,
            total_tokens: 340,
            cost: 0.15,
            actual_cost: 0.12,
          },
        ],
      }),
    }))
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    vi.useRealTimers()
  })

  it('renders daily usage detail rows after a successful query', async () => {
    const wrapper = mount(KeyUsageView, {
      global: {
        stubs: {
          RouterLink: { template: '<a><slot /></a>' },
          LocaleSwitcher: true,
          Icon: true,
        },
      },
    })

    await wrapper.find('input').setValue('sk-test-key')
    await wrapper.find('input').trigger('keydown.enter')
    await flushPromises()
    await nextTick()

    const fetchMock = vi.mocked(fetch)
    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining('/v1/usage?'),
      expect.objectContaining({
        headers: { Authorization: 'Bearer sk-test-key' },
      })
    )
    expect(String(fetchMock.mock.calls[0][0])).toContain('days=30')

    const text = wrapper.text()
    expect(text).toContain('Daily Detail')
    expect(text).toContain('Date')
    expect(text).toContain('Cache Read')
    expect(text).toContain('Cache Write')
    expect(text).toContain('2026-05-19')
    expect(text).toContain('12')
    expect(text).toContain('100')
    expect(text).toContain('200')
    expect(text).toContain('30')
    expect(text).toContain('10')
    expect(text).toContain('$0.12')

    wrapper.unmount()
  })

  it('uses key usage shell defaults for initial date range and daily detail days', async () => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date(Date.UTC(2026, 5, 20, 12, 0, 0)))
    publicSettings.value = {
      key_usage_shell_config: buildKeyUsageShellConfig({}, {
        defaultDateRange: '7d',
        dailyUsageDays: 90,
      }),
    }

    const wrapper = mount(KeyUsageView, {
      global: {
        stubs: {
          RouterLink: { template: '<a><slot /></a>' },
          LocaleSwitcher: true,
          Icon: true,
        },
      },
    })

    await wrapper.find('input').setValue('sk-test-key')
    await wrapper.find('input').trigger('keydown.enter')
    await flushPromises()

    const fetchMock = vi.mocked(fetch)
    const requestUrl = String(fetchMock.mock.calls[0][0])
    expect(requestUrl).toContain('start_date=2026-06-14')
    expect(requestUrl).toContain('end_date=2026-06-20')
    expect(requestUrl).toContain('days=90')

    wrapper.unmount()
  })

  it('renders key usage shell labels from public settings', () => {
    publicSettings.value = {
      key_usage_shell_config: JSON.stringify({
        en: {
          labels: {
            title: 'Configured Usage Title',
            subtitle: 'Configured usage subtitle',
            placeholder: 'configured-placeholder',
            query: 'Configured Query',
            privacyNote: 'Configured privacy note',
            docs: 'Configured Docs',
            allRightsReserved: 'Configured rights',
          },
        },
      }),
      docs_shell_config: JSON.stringify({
        en: { labels: { title: 'Wrong Docs Source' } },
      }),
      home_shell_config: JSON.stringify({
        en: { labels: { allRightsReserved: 'Wrong Home Source' } },
      }),
      doc_url: '/docs',
    }

    const wrapper = mount(KeyUsageView, {
      global: {
        stubs: {
          RouterLink: { template: '<a><slot /></a>' },
          LocaleSwitcher: true,
          Icon: true,
        },
      },
    })

    const text = wrapper.text()
    expect(text).toContain('Configured Usage Title')
    expect(text).toContain('Configured usage subtitle')
    expect(text).toContain('Configured Query')
    expect(text).toContain('Configured privacy note')
    expect(text).toContain('Configured Docs')
    expect(text).toContain('Configured rights')
    expect(text).not.toContain('Wrong Docs Source')
    expect(text).not.toContain('Wrong Home Source')
    expect(wrapper.find('input').attributes('placeholder')).toBe('configured-placeholder')

    wrapper.unmount()
  })

  it('keeps shell label parsing in the shared key usage shell helper', () => {
    expect(keyUsageViewSource).toContain("} from '@/utils/keyUsageShell'")
    expect(keyUsageViewSource).toContain('resolveKeyUsageShellConfig(')
    expect(keyUsageViewSource).toContain('renderKeyUsageShellText(')
    expect(keyUsageViewSource).not.toContain('resolveKeyUsageShellLabels(')
    expect(keyUsageViewSource).not.toContain("import { resolveLocalizedShellLabels } from '@/utils/localizedShell'")
    expect(keyUsageViewSource).not.toContain('resolveLocalizedShellLabels(')
    expect(keyUsageViewSource).not.toContain('const keyUsageShellLabelKeys')
    expect(keyUsageViewSource).not.toContain('function readLocalizedShellLabels')
    expect(keyUsageViewSource).not.toContain('function isRecord')
  })

  it('does not keep locale-specific key usage fallback copy in the view bootstrap layer', () => {
    expect(keyUsageViewSource).toContain('useAuthRouteDefaults')
    expect(keyUsageViewSource).toContain(':to="authRouteDefaults.homePath"')
    expect(keyUsageViewSource).not.toContain('to="/home"')
    expect(keyUsageViewSource).toContain("from './keyUsageRuntime'")
    expect(keyUsageViewSource).toContain('buildKeyUsageDateParams')
    expect(keyUsageViewSource).toContain('resolveKeyUsageStatusInfo')
    expect(keyUsageViewSource).not.toContain('FALLBACK_KEY_USAGE_LABELS')
    expect(keyUsageViewSource).not.toContain('EMPTY_KEY_USAGE_LABELS')
    expect(keyUsageViewSource).not.toContain('keyUsageShellLabels.value[key as KeyUsageShellLabelKey] || key')
    expect(keyUsageViewSource).not.toContain('API Key Usage')
    expect(keyUsageViewSource).not.toContain('API Key 用量查询')
    expect(keyUsageViewSource).not.toContain("currentRange = ref<DateRangeKey>('today')")
    expect(keyUsageViewSource).not.toContain('dailyUsageDays = ref<7 | 30 | 90>(30)')
    expect(keyUsageViewSource).not.toContain('7 * 86400000')
    expect(keyUsageViewSource).not.toContain('30 * 86400000')
  })

  it('uses the shared runtime locale helper instead of direct locale value checks', () => {
    expect(keyUsageViewSource).toContain("from '@/utils/runtimeLocale'")
    expect(keyUsageViewSource).toContain('resolveRuntimeLanguage(locale)')
    expect(keyUsageViewSource).toContain('resolveRuntimeLocale(locale)')
    expect(keyUsageViewSource).not.toContain("locale.value === 'zh'")
    expect(keyUsageViewSource).not.toContain('locale.value === "zh"')
  })
})
