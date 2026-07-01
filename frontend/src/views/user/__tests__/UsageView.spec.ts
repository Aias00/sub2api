import { readFileSync } from 'node:fs'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import UsageView from '../UsageView.vue'

const usageViewSource = readFileSync('src/views/user/UsageView.vue', 'utf8')

const {
  query,
  getStats,
  getStatsByDateRange,
  getDashboardModels,
  getDashboardSnapshotV2,
  list,
  getAvailable,
  showError,
  showWarning,
  showSuccess,
  showInfo,
} = vi.hoisted(() => ({
  query: vi.fn(),
  getStats: vi.fn(),
  getStatsByDateRange: vi.fn(),
  getDashboardModels: vi.fn(),
  getDashboardSnapshotV2: vi.fn(),
  list: vi.fn(),
  getAvailable: vi.fn(),
  showError: vi.fn(),
  showWarning: vi.fn(),
  showSuccess: vi.fn(),
  showInfo: vi.fn(),
}))

const appStoreState = {
  showError,
  showWarning,
  showSuccess,
  showInfo,
  cachedPublicSettings: null as null | { usage_shell_config?: string; pricing_currency_symbol?: string },
}

const messages: Record<string, string> = {
  'admin.dashboard.timeRange': 'Time range',
  'admin.dashboard.granularity': 'Granularity',
  'admin.dashboard.day': 'Day',
  'admin.dashboard.hour': 'Hour',
  'admin.users.columnSettings': 'Columns',
  'admin.usage.group': 'Group',
  'admin.usage.billingType': 'Billing type',
  'admin.usage.billingMode': 'Billing mode',
  'admin.usage.allTypes': 'All types',
  'admin.usage.allBillingTypes': 'All billing types',
  'admin.usage.billingTypeBalance': 'Balance',
  'admin.usage.billingTypeSubscription': 'Subscription',
  'admin.usage.allBillingModes': 'All billing modes',
  'admin.usage.billingModeToken': 'Token',
  'admin.usage.billingModePerRequest': 'Per request',
  'admin.usage.billingModeImage': 'Image',
  'admin.usage.allGroups': 'All groups',
  'admin.usage.allModels': 'All models',
  'usage.allApiKeys': 'All API Keys',
  'usage.apiKeyFilter': 'API Key',
  'usage.model': 'Model',
  'usage.type': 'Type',
  'usage.ws': 'WS',
  'usage.stream': 'Stream',
  'usage.sync': 'Sync',
  'usage.exporting': 'Exporting',
  'usage.exportCsv': 'Export CSV',
  'usage.failedToLoad': 'Failed to load',
  'usage.noDataToExport': 'No data',
  'usage.preparingExport': 'Preparing export',
  'usage.exportSuccess': 'Export success',
  'usage.exportFailed': 'Export failed',
  'common.refresh': 'Refresh',
  'common.reset': 'Reset',
}

function buildUsageShellConfig(
  overrides: Record<string, string> = {},
  defaults: Record<string, unknown> = {},
): string {
  return JSON.stringify({
    zh: {
      labels: {
        totalRequests: 'Total Requests',
        inSelectedRange: 'In selected range',
        totalTokens: 'Total Tokens',
        in: 'In',
        out: 'Out',
        totalCost: 'Total Cost',
        actualCost: 'Actual Cost',
        standardCost: 'Standard Cost',
        avgDuration: 'Avg Duration',
        perRequest: 'Per request',
        apiKeyFilter: 'API Key',
        allApiKeys: 'All API Keys',
        timeRange: 'Time Range',
        refresh: 'Refresh',
        reset: 'Reset',
        exportCsv: 'Export CSV',
        exporting: 'Exporting...',
        model: 'Model',
        reasoningEffort: 'Reasoning Effort',
        endpoint: 'Endpoint',
        type: 'Type',
        billingMode: 'Billing Mode',
        tokens: 'Tokens',
        cost: 'Cost',
        firstToken: 'First Token',
        duration: 'Duration',
        time: 'Time',
        userAgent: 'User Agent',
        noRecords: 'No usage records',
        rate: 'Rate',
        original: 'Original',
        billed: 'Billed',
        failedToLoad: 'Failed to load usage records',
        noDataToExport: 'No data to export',
        preparingExport: 'Preparing export...',
        exportSuccess: 'Export successful',
        exportFailed: 'Export failed',
        ...overrides,
      },
      defaults,
    },
  })
}

vi.mock('@/api', () => ({
  usageAPI: {
    query,
    getStats,
    getStatsByDateRange,
    getDashboardModels,
    getDashboardSnapshotV2,
  },
  keysAPI: {
    list,
  },
  userGroupsAPI: {
    getAvailable,
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => appStoreState,
}))

vi.mock('vue-router', () => ({
  useRoute: () => ({ query: {} }),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => messages[key] ?? key,
      locale: { value: 'zh-CN' },
    }),
  }
})

const simpleStub = { template: '<div><slot /></div>' }
const chartStub = { template: '<div />' }

const usageLog = {
  id: 1,
  request_id: 'req-user-export',
  actual_cost: 0.092883,
  total_cost: 0.092883,
  rate_multiplier: 1,
  service_tier: 'priority',
  input_cost: 0.020285,
  output_cost: 0.00303,
  cache_creation_cost: 0.000001,
  cache_read_cost: 0.069568,
  input_tokens: 4057,
  output_tokens: 101,
  cache_creation_tokens: 4,
  cache_read_tokens: 278272,
  cache_creation_5m_tokens: 0,
  cache_creation_1h_tokens: 0,
  image_count: 0,
  image_size: null,
  first_token_ms: 12,
  duration_ms: 345,
  created_at: '2026-03-08T00:00:00Z',
  model: 'gpt-5.4',
  reasoning_effort: null,
  ip_address: '203.0.113.10',
  api_key: { name: 'demo-key' },
  billing_mode: 'token',
  request_type: 'sync',
  stream: false,
}

function mountUsageView() {
  return mount(UsageView, {
    global: {
      stubs: {
        AppLayout: simpleStub,
        Pagination: true,
        Select: true,
        DateRangePicker: true,
        Icon: true,
        UsageStatsCards: chartStub,
        UsageTable: chartStub,
        ModelDistributionChart: chartStub,
        GroupDistributionChart: chartStub,
        EndpointDistributionChart: chartStub,
        TokenUsageTrend: chartStub,
      },
    },
  })
}

describe('user UsageView', () => {
  beforeEach(() => {
    query.mockReset()
    getStats.mockReset()
    getStatsByDateRange.mockReset()
    getDashboardModels.mockReset()
    getDashboardSnapshotV2.mockReset()
    list.mockReset()
    getAvailable.mockReset()
    showError.mockReset()
    showWarning.mockReset()
    showSuccess.mockReset()
    showInfo.mockReset()
    appStoreState.cachedPublicSettings = null
    window.matchMedia = vi.fn().mockImplementation((query: string) => ({
      matches: true,
      media: query,
      onchange: null,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      addListener: vi.fn(),
      removeListener: vi.fn(),
      dispatchEvent: vi.fn(),
    }))

    query.mockResolvedValue({ items: [usageLog], total: 1, pages: 1 })
    getStatsByDateRange.mockResolvedValue({
      total_requests: 1,
      total_tokens: 30,
      total_input_tokens: 10,
      total_output_tokens: 20,
      total_cost: 0.01,
      total_actual_cost: 0.008,
      average_duration_ms: 250,
    })
    getStats.mockResolvedValue({
      total_requests: 1,
      total_input_tokens: 10,
      total_output_tokens: 20,
      total_cache_tokens: 0,
      total_tokens: 30,
      total_cost: 0.1,
      total_actual_cost: 0.08,
      average_duration_ms: 12,
      endpoints: [],
      upstream_endpoints: [],
      endpoint_paths: [],
    })
    getDashboardModels.mockResolvedValue({
      models: [{ model: 'gpt-5.4', requests: 1, input_tokens: 10, output_tokens: 20, cache_creation_tokens: 0, cache_read_tokens: 0, total_tokens: 30, cost: 0.1, actual_cost: 0.08 }],
      start_date: '2026-03-08',
      end_date: '2026-03-08',
    })
    getDashboardSnapshotV2.mockResolvedValue({
      generated_at: '2026-03-08T00:00:00Z',
      start_date: '2026-03-08',
      end_date: '2026-03-08',
      granularity: 'hour',
      trend: [],
      groups: [],
    })
    list.mockResolvedValue({ items: [{ id: 1, name: 'demo-key' }] })
    getAvailable.mockResolvedValue([{ id: 1, name: 'default' }])
  })

  it('loads logs, date-range stats, and API key options on first render', async () => {
    mountUsageView()
    await flushPromises()

    expect(query).toHaveBeenCalled()
    expect(getStatsByDateRange).toHaveBeenCalled()
    expect(list).toHaveBeenCalledWith(1, 100)
  })

  it('exports csv with current filters and without admin-only fields', async () => {
    const wrapper = mountUsageView()
    await flushPromises()

    let exportedBlob: Blob | null = null
    let csvContent = ''
    const OriginalBlob = globalThis.Blob
    vi.stubGlobal('Blob', vi.fn((parts: BlobPart[], options?: BlobPropertyBag) => {
      csvContent = parts.map((part) => String(part)).join('')
      return new OriginalBlob(parts, options)
    }))
    const originalCreateObjectURL = window.URL.createObjectURL
    const originalRevokeObjectURL = window.URL.revokeObjectURL
    window.URL.createObjectURL = vi.fn((blob: Blob | MediaSource) => {
      exportedBlob = blob as Blob
      return 'blob:usage-export'
    }) as typeof window.URL.createObjectURL
    window.URL.revokeObjectURL = vi.fn(() => {}) as typeof window.URL.revokeObjectURL
    const clickSpy = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {})

    await (wrapper.vm as any).exportToCSV()

    expect(exportedBlob).not.toBeNull()
    expect(query).toHaveBeenCalledWith(expect.objectContaining({
      page_size: 100,
      sort_by: 'created_at',
      sort_order: 'desc',
    }))
    expect(clickSpy).toHaveBeenCalled()
    expect(showSuccess).toHaveBeenCalled()
    expect(csvContent).not.toContain('IP Address')
    expect(csvContent).not.toContain('203.0.113.10')
    expect(csvContent).toContain('Billed Cost')
    expect(csvContent).toContain('Original Cost')
    expect(csvContent).not.toContain('Upstream Endpoint')
    expect(csvContent).not.toContain('account_cost')
    expect(csvContent).not.toContain('account_rate_multiplier')

    window.URL.createObjectURL = originalCreateObjectURL
    window.URL.revokeObjectURL = originalRevokeObjectURL
    vi.unstubAllGlobals()
    clickSpy.mockRestore()
  })

  it('exports historical image rows with image billing mode derived from image_count', async () => {
    query.mockResolvedValue({
      items: [
        {
          ...usageLog,
          request_id: 'req-user-export-legacy-image',
          actual_cost: 0.2,
          total_cost: 0.2,
          input_cost: 0,
          output_cost: 0,
          cache_creation_cost: 0,
          cache_read_cost: 0,
          input_tokens: 0,
          output_tokens: 0,
          cache_creation_tokens: 0,
          cache_read_tokens: 0,
          image_count: 1,
          model: 'gpt-image-2',
          billing_mode: null,
          ip_address: null,
        },
      ],
      total: 1,
      pages: 1,
    })

    const wrapper = mountUsageView()
    await flushPromises()

    let csvContent = ''
    const OriginalBlob = globalThis.Blob
    vi.stubGlobal('Blob', vi.fn((parts: BlobPart[], options?: BlobPropertyBag) => {
      csvContent = parts.map((part) => String(part)).join('')
      return new OriginalBlob(parts, options)
    }))
    const originalCreateObjectURL = window.URL.createObjectURL
    const originalRevokeObjectURL = window.URL.revokeObjectURL
    window.URL.createObjectURL = vi.fn(() => 'blob:usage-export') as typeof window.URL.createObjectURL
    window.URL.revokeObjectURL = vi.fn(() => {}) as typeof window.URL.revokeObjectURL
    const clickSpy = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {})

    await (wrapper.vm as any).exportToCSV()

    expect(csvContent).toContain('Billing Mode')
    expect(csvContent).toContain('Image')
    expect(csvContent).not.toContain(',Token,0,0,0,0,')

    window.URL.createObjectURL = originalCreateObjectURL
    window.URL.revokeObjectURL = originalRevokeObjectURL
    vi.unstubAllGlobals()
    clickSpy.mockRestore()
  })

  it('does not keep usage shell i18n fallback keys in the view bootstrap layer', () => {
    expect(usageViewSource).not.toContain('usageShellFallbackKeys')
    expect(usageViewSource).not.toContain('usage.totalRequests')
    expect(usageViewSource).not.toContain('admin.usage.billingMode')
    expect(usageViewSource).not.toContain('usageShellLabels.value[key] || key')
    expect(usageViewSource).not.toContain('const usageShellLabelKeys')
    expect(usageViewSource).not.toContain('resolveShellLabelOverrides(')
    expect(usageViewSource).not.toContain('const pageSize = 100')
    expect(usageViewSource).toContain('usageShell.value.defaults.exportPageSize')
    expect(usageViewSource).toContain('usageShell.value.defaults.dateRangeDays')
    expect(usageViewSource).toContain('usageShell.value.defaults.apiKeyPageSize')
    expect(usageViewSource).not.toContain('keysAPI.list(1, 100)')
    expect(usageViewSource).not.toContain('weekAgo.setDate(weekAgo.getDate() - 6)')
    expect(usageViewSource).toContain("from './usageRuntime'")
    expect(usageViewSource).toContain('resolveUsageDefaultDateRange')
    expect(usageViewSource).toContain('buildUsageTableQueryParams')
    expect(usageViewSource).not.toContain('usage.image')
    expect(usageViewSource).not.toContain("usageText('serviceTier')")
    expect(usageViewSource).not.toContain("usageText('tokenDetails')")
    expect(usageViewSource).not.toContain("usageText('costDetails')")
    expect(usageViewSource).not.toContain("usageText('ws')")
    expect(usageViewSource).not.toContain("usageText('stream')")
    expect(usageViewSource).not.toContain("usageText('sync')")
    expect(usageViewSource).not.toContain("usageText('unknown')")
    expect(usageViewSource).toContain('resolveUsageShellConfig')
    expect(usageViewSource).toContain('renderUsageShellText')
  })

  it('does not hard-code a dollar prefix for user-visible usage costs', () => {
    expect(usageViewSource).not.toContain('>${{')
    expect(usageViewSource).not.toContain('${{')
    expect(usageViewSource).not.toContain('return `$${formatted}`')
    expect(usageViewSource).toContain('pricing_currency_symbol')
    expect(usageViewSource).toContain('formatUsageCost(')
  })
})
