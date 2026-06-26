import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import { readFileSync } from 'node:fs'

import UsageView from '../UsageView.vue'

const usageViewSource = readFileSync('src/views/user/UsageView.vue', 'utf8')

const { query, getStatsByDateRange, list, appStoreState } = vi.hoisted(() => ({
  query: vi.fn(),
  getStatsByDateRange: vi.fn(),
  list: vi.fn(),
  appStoreState: {
    cachedPublicSettings: null as null | {
      usage_shell_config?: string
      pricing_currency_symbol?: string
    },
    showError: vi.fn(),
    showWarning: vi.fn(),
    showSuccess: vi.fn(),
    showInfo: vi.fn(),
  },
}))

const { showError, showWarning, showSuccess, showInfo } = appStoreState

const messages: Record<string, string> = {
  'common.usageMetrics.cacheCreation1hTokens': 'Cache Creation 1h Tokens',
  'common.usageMetrics.cacheCreation5mTokens': 'Cache Creation 5m Tokens',
  'common.usageMetrics.cacheCreationCost': 'Cache Creation Cost',
  'common.usageMetrics.cacheCreationTokens': 'Cache Creation Tokens',
  'common.usageMetrics.cacheReadCost': 'Cache Read Cost',
  'common.usageMetrics.cacheReadTokens': 'Cache Read Tokens',
  'common.usageMetrics.cacheTtlOverridden1h': '1h write reduced to 5m',
  'common.usageMetrics.cacheTtlOverridden5m': '5m write promoted to 1h',
  'common.usageMetrics.cacheTtlOverriddenHint': 'The cache write TTL was adjusted automatically',
  'common.usageMetrics.cacheTtlOverriddenLabel': 'Cache TTL adjusted',
  'common.usageMetrics.costDetails': 'Cost Breakdown',
  'common.usageMetrics.inputCost': 'Input Cost',
  'common.usageMetrics.inputTokenPrice': 'Input price',
  'common.usageMetrics.inputTokens': 'Input Tokens',
  'common.usageMetrics.outputCost': 'Output Cost',
  'common.usageMetrics.outputTokenPrice': 'Output price',
  'common.usageMetrics.outputTokens': 'Output Tokens',
  'common.usageMetrics.perMillionTokens': '/ 1M tokens',
  'common.usageMetrics.tokenDetails': 'Token Details',
  'common.usageMetrics.totalTokens': 'Total Tokens',
  'common.usageMetrics.unitPrice': 'Unit Price',
  'common.serviceTier.label': 'Service tier',
  'common.serviceTier.priority': 'Fast',
  'common.serviceTier.flex': 'Flex',
  'common.serviceTier.standard': 'Standard',
  'common.requestType.ws': 'WS',
  'common.requestType.stream': 'Stream',
  'common.requestType.sync': 'Sync',
  'common.requestType.unknown': 'Unknown',
  'usage.rate': 'Rate',
  'usage.original': 'Original',
  'usage.billed': 'Billed',
  'usage.allApiKeys': 'All API Keys',
  'usage.apiKeyFilter': 'API Key',
  'usage.model': 'Model',
  'usage.reasoningEffort': 'Reasoning Effort',
  'usage.type': 'Type',
  'usage.tokens': 'Tokens',
  'usage.cost': 'Cost',
  'usage.firstToken': 'First Token',
  'usage.duration': 'Duration',
  'usage.time': 'Time',
  'usage.userAgent': 'User Agent',
  'common.imageUsage.unit': ' images',
  'common.imageUsage.count': 'Image count',
  'common.imageUsage.billingSize': 'Billing size',
  'common.imageUsage.inputSize': 'Input size',
  'common.imageUsage.outputSize': 'Output size',
  'common.imageUsage.sizeSource': 'Size source',
  'common.imageUsage.sizeBreakdown': 'Size breakdown',
  'common.imageUsage.sizeSourceOutput': 'Upstream output',
  'common.imageUsage.sizeSourceInput': 'Request input',
  'common.imageUsage.sizeSourceDefault': 'Default billing tier',
  'common.imageUsage.sizeSourceLegacy': 'Legacy record',
  'common.imageUsage.sizeSourceMissing': 'Not recorded',
  'common.imageUsage.sizeNotRecorded': 'not recorded',
  'common.imageUsage.sizeLegacyUnstandardized': 'legacy unstandardized',
  'common.imageUsage.sizeUnknown': 'unknown',
  'common.imageUsage.unitPrice': 'Per-image price',
  'common.imageUsage.totalPrice': 'Image total price',
  'admin.usage.billingModeToken': 'Token',
  'admin.usage.billingModePerRequest': 'Per request',
  'admin.usage.billingModeImage': 'Image',
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
    getStatsByDateRange,
  },
  keysAPI: {
    list,
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

const AppLayoutStub = { template: '<div><slot /></div>' }
const TablePageLayoutStub = {
  template: '<div><slot name="actions" /><slot name="filters" /><slot name="table" /><slot /></div>',
}
const DataTableStub = {
  props: ['data', 'columns'],
  template: `
    <div>
      <div>
        <span v-for="column in columns" :key="column.key">{{ column.label }}</span>
      </div>
      <slot v-if="data.length === 0" name="empty" />
      <div v-for="row in data" :key="row.request_id">
        <slot name="cell-billing_mode" :row="row" />
        <slot name="cell-tokens" :row="row" />
        <slot name="cell-cost" :row="row" />
      </div>
    </div>
  `,
}

describe('user UsageView tooltip', () => {
  beforeEach(() => {
    query.mockReset()
    getStatsByDateRange.mockReset()
    list.mockReset()
    showError.mockReset()
    showWarning.mockReset()
    showSuccess.mockReset()
    showInfo.mockReset()
    appStoreState.cachedPublicSettings = null

    vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockReturnValue({
      x: 0,
      y: 0,
      top: 20,
      left: 20,
      right: 120,
      bottom: 40,
      width: 100,
      height: 20,
      toJSON: () => ({}),
    } as DOMRect)

    ;(globalThis as any).ResizeObserver = class {
      observe() {}
      disconnect() {}
    }
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('renders stable usage shell labels from public settings', async () => {
    appStoreState.cachedPublicSettings = {
      pricing_currency_symbol: '€',
      usage_shell_config: buildUsageShellConfig({
        totalRequests: '配置总请求',
        inSelectedRange: '配置范围',
        totalTokens: '配置 Tokens',
        in: '配置输入',
        out: '配置输出',
        totalCost: '配置总费用',
        actualCost: '配置实际',
        standardCost: '配置标准',
        avgDuration: '配置平均耗时',
        perRequest: '配置单次',
        apiKeyFilter: '配置 API Key',
        allApiKeys: '配置全部密钥',
        timeRange: '配置时间范围',
        refresh: '配置刷新',
        reset: '配置重置',
        exportCsv: '配置导出',
        noRecords: '配置空记录',
        tokenDetails: '配置 Token 明细',
        costDetails: '配置费用明细',
        failedToLoad: '配置加载失败',
        noDataToExport: '配置无导出数据',
        preparingExport: '配置准备导出',
        exportSuccess: '配置导出成功',
        exportFailed: '配置导出失败',
        model: '配置模型',
        billingMode: '配置计费模式',
      }),
    }
    query.mockResolvedValue({
      items: [],
      total: 0,
      pages: 0,
    })
    getStatsByDateRange.mockResolvedValue({
      total_requests: 3,
      total_tokens: 100,
      total_input_tokens: 40,
      total_output_tokens: 60,
      total_actual_cost: 0.12,
      total_cost: 0.2,
      average_duration_ms: 123,
    })
    list.mockResolvedValue({ items: [] })

    const wrapper = mount(UsageView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
          TablePageLayout: TablePageLayoutStub,
          Pagination: true,
          EmptyState: { props: ['message'], template: '<div>{{ message }}</div>' },
          Select: true,
          DateRangePicker: true,
          DataTable: DataTableStub,
          Icon: true,
          Teleport: true,
        },
      },
    })

    await flushPromises()
    await nextTick()

    const text = wrapper.text()
    expect(text).toContain('配置总请求')
    expect(text).toContain('配置范围')
    expect(text).toContain('配置 Tokens')
    expect(text).toContain('配置输入')
    expect(text).toContain('配置输出')
    expect(text).toContain('配置总费用')
    expect(text).toContain('配置实际')
    expect(text).toContain('配置标准')
    expect(text).toContain('配置平均耗时')
    expect(text).toContain('配置单次')
    expect(text).toContain('配置 API Key')
    expect(text).toContain('配置时间范围')
    expect(text).toContain('配置刷新')
    expect(text).toContain('配置重置')
    expect(text).toContain('配置导出')
    expect(text).toContain('配置空记录')
    expect(text).toContain('配置模型')
    expect(text).toContain('配置计费模式')
  })

  it('uses usage shell defaults for the initial date range and API key fetch size', async () => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date(2026, 5, 20, 12, 0, 0))
    appStoreState.cachedPublicSettings = {
      usage_shell_config: buildUsageShellConfig({}, {
        dateRangeDays: 14,
        apiKeyPageSize: 23,
      }),
    }
    query.mockResolvedValue({
      items: [],
      total: 0,
      pages: 0,
    })
    getStatsByDateRange.mockResolvedValue({
      total_requests: 0,
      total_tokens: 0,
      total_cost: 0,
      avg_duration_ms: 0,
    })
    list.mockResolvedValue({ items: [] })

    mount(UsageView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
          TablePageLayout: TablePageLayoutStub,
          Pagination: true,
          EmptyState: true,
          Select: true,
          DateRangePicker: true,
          DataTable: DataTableStub,
          Icon: true,
          Teleport: true,
        },
      },
    })

    await flushPromises()

    expect(list).toHaveBeenCalledWith(1, 23)
    expect(query).toHaveBeenCalledWith(
      expect.objectContaining({
        start_date: '2026-06-07',
        end_date: '2026-06-20',
      }),
      expect.any(Object),
    )
    expect(getStatsByDateRange).toHaveBeenCalledWith('2026-06-07', '2026-06-20', undefined)
  })

  it('shows fast service tier and unit prices in user tooltip', async () => {
    appStoreState.cachedPublicSettings = {
      pricing_currency_symbol: '€',
      usage_shell_config: buildUsageShellConfig({
        cacheCreationCost: '配置缓存创建费用',
        cacheCreationTokens: '配置缓存创建 Tokens',
        cacheReadCost: '配置缓存读取费用',
        cacheReadTokens: '配置缓存读取 Tokens',
        costDetails: '配置费用明细',
        inputCost: '配置输入费用',
        inputTokens: '配置输入 Tokens',
        outputCost: '配置输出费用',
        outputTokens: '配置输出 Tokens',
        tokenDetails: '配置 Token 明细',
      }),
    }
    query.mockResolvedValue({
      items: [
        {
          request_id: 'req-user-1',
          actual_cost: 0.092883,
          total_cost: 0.092883,
          rate_multiplier: 1,
          service_tier: 'priority',
          input_cost: 0.020285,
          output_cost: 0.00303,
          cache_creation_cost: 0,
          cache_read_cost: 0.069568,
          input_tokens: 4057,
          output_tokens: 101,
          cache_creation_tokens: 0,
          cache_read_tokens: 278272,
          cache_creation_5m_tokens: 0,
          cache_creation_1h_tokens: 0,
          image_count: 0,
          image_size: null,
          first_token_ms: null,
          duration_ms: 1,
          created_at: '2026-03-08T00:00:00Z',
        },
      ],
      total: 1,
      pages: 1,
    })
    getStatsByDateRange.mockResolvedValue({
      total_requests: 1,
      total_tokens: 100,
      total_cost: 0.1,
      avg_duration_ms: 1,
    })
    list.mockResolvedValue({ items: [] })

    const wrapper = mount(UsageView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
          TablePageLayout: TablePageLayoutStub,
          Pagination: true,
          EmptyState: true,
          Select: true,
          DateRangePicker: true,
          DataTable: DataTableStub,
          Icon: true,
          Teleport: true,
        },
      },
    })

    await flushPromises()
    await nextTick()

    const setupState = (wrapper.vm as any).$?.setupState
    setupState.tokenTooltipData = {
      input_tokens: 4057,
      output_tokens: 101,
      cache_creation_tokens: 4,
      cache_creation_5m_tokens: 0,
      cache_creation_1h_tokens: 0,
      cache_read_tokens: 278272,
    }
    setupState.tokenTooltipVisible = true
    setupState.tooltipData = {
      request_id: 'req-user-1',
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
    }
    setupState.tooltipVisible = true
    await nextTick()

    const text = wrapper.text()
    expect(text).toContain('Token Details')
    expect(text).toContain('Input Tokens')
    expect(text).toContain('Output Tokens')
    expect(text).toContain('Cache Creation Tokens')
    expect(text).toContain('Cache Read Tokens')
    expect(text).toContain('Cost Breakdown')
    expect(text).toContain('Input Cost')
    expect(text).toContain('Output Cost')
    expect(text).toContain('Cache Creation Cost')
    expect(text).toContain('Cache Read Cost')
    expect(text).toContain('Service tier')
    expect(text).toContain('Fast')
    expect(text).toContain('Rate')
    expect(text).toContain('1.00x')
    expect(text).toContain('Billed')
    expect(text).toContain('€0.092883')
    expect(text).toContain('€5.0000 / 1M tokens')
    expect(text).toContain('€30.0000 / 1M tokens')
    expect(text).not.toContain('$0.092883')
  })

  it('exports csv with input and output unit price columns', async () => {
    appStoreState.cachedPublicSettings = {
      usage_shell_config: buildUsageShellConfig({
        preparingExport: '配置准备导出',
        exportSuccess: '配置导出成功',
      }, {
        exportPageSize: 37,
      }),
    }
    const exportedLogs = [
      {
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
        api_key: { name: 'demo-key' },
      },
    ]

    query.mockResolvedValue({
      items: exportedLogs,
      total: 1,
      pages: 1,
    })
    getStatsByDateRange.mockResolvedValue({
      total_requests: 1,
      total_tokens: 100,
      total_cost: 0.1,
      avg_duration_ms: 1,
    })
    list.mockResolvedValue({ items: [] })

    let exportedBlob: Blob | null = null
    const originalCreateObjectURL = window.URL.createObjectURL
    const originalRevokeObjectURL = window.URL.revokeObjectURL
    window.URL.createObjectURL = vi.fn((blob: Blob | MediaSource) => {
      exportedBlob = blob as Blob
      return 'blob:usage-export'
    }) as typeof window.URL.createObjectURL
    window.URL.revokeObjectURL = vi.fn(() => {}) as typeof window.URL.revokeObjectURL
    const clickSpy = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {})

    const wrapper = mount(UsageView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
          TablePageLayout: TablePageLayoutStub,
          Pagination: true,
          EmptyState: true,
          Select: true,
          DateRangePicker: true,
          DataTable: DataTableStub,
          Icon: true,
          Teleport: true,
        },
      },
    })

    await flushPromises()

    const setupState = (wrapper.vm as any).$?.setupState
    await setupState.exportToCSV()

    expect(exportedBlob).not.toBeNull()
    const hasSortedExportQuery = query.mock.calls.some((call) => {
      const params = call[0] as Record<string, unknown> | undefined
      const config = call[1]
      return (
        params?.page_size === 37 &&
        params?.sort_by === 'created_at' &&
        params?.sort_order === 'desc' &&
        config === undefined
      )
    })
    expect(hasSortedExportQuery).toBe(true)
    expect(clickSpy).toHaveBeenCalled()
    expect(showInfo).toHaveBeenCalledWith('配置准备导出')
    expect(showSuccess).toHaveBeenCalledWith('配置导出成功')

    window.URL.createObjectURL = originalCreateObjectURL
    window.URL.revokeObjectURL = originalRevokeObjectURL
    clickSpy.mockRestore()
  })

  it('exports historical image rows with image billing mode derived from image_count', async () => {
    const exportedLogs = [
      {
        request_id: 'req-user-export-legacy-image',
        actual_cost: 0.2,
        total_cost: 0.2,
        rate_multiplier: 1,
        service_tier: null,
        input_cost: 0,
        output_cost: 0,
        cache_creation_cost: 0,
        cache_read_cost: 0,
        input_tokens: 0,
        output_tokens: 0,
        cache_creation_tokens: 0,
        cache_read_tokens: 0,
        cache_creation_5m_tokens: 0,
        cache_creation_1h_tokens: 0,
        image_count: 1,
        image_size: null,
        billing_mode: null,
        first_token_ms: null,
        duration_ms: 345,
        created_at: '2026-03-08T00:00:00Z',
        model: 'gpt-image-2',
        reasoning_effort: null,
        api_key: { name: 'demo-key' },
      },
    ]

    query.mockResolvedValue({
      items: exportedLogs,
      total: 1,
      pages: 1,
    })
    getStatsByDateRange.mockResolvedValue({
      total_requests: 1,
      total_tokens: 0,
      total_cost: 0.2,
      avg_duration_ms: 1,
    })
    list.mockResolvedValue({ items: [] })

    let exportedBlob: Blob | null = null
    const originalCreateObjectURL = window.URL.createObjectURL
    const originalRevokeObjectURL = window.URL.revokeObjectURL
    window.URL.createObjectURL = vi.fn((blob: Blob | MediaSource) => {
      exportedBlob = blob as Blob
      return 'blob:usage-export'
    }) as typeof window.URL.createObjectURL
    window.URL.revokeObjectURL = vi.fn(() => {}) as typeof window.URL.revokeObjectURL
    const clickSpy = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {})

    const wrapper = mount(UsageView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
          TablePageLayout: TablePageLayoutStub,
          Pagination: true,
          EmptyState: true,
          Select: true,
          DateRangePicker: true,
          DataTable: DataTableStub,
          Icon: true,
          Teleport: true,
        },
      },
    })

    await flushPromises()

    const setupState = (wrapper.vm as any).$?.setupState
    await setupState.exportToCSV()

    expect(exportedBlob).not.toBeNull()
    const csv = await new Promise<string>((resolve, reject) => {
      const reader = new FileReader()
      reader.onload = () => resolve(String(reader.result))
      reader.onerror = () => reject(reader.error)
      reader.readAsText(exportedBlob as Blob)
    })
    expect(csv).toContain('Billing Mode')
    expect(csv).toContain('Image')
    expect(csv).not.toContain(',Token,0,0,0,0,')

    window.URL.createObjectURL = originalCreateObjectURL
    window.URL.revokeObjectURL = originalRevokeObjectURL
    clickSpy.mockRestore()
  })

  it('does not display a 2K fallback for historical image rows with missing size', async () => {
    query.mockResolvedValue({
      items: [
        {
          request_id: 'req-user-legacy-missing-image',
          actual_cost: 0.2,
          total_cost: 0.2,
          rate_multiplier: 1,
          service_tier: null,
          input_cost: 0,
          output_cost: 0,
          cache_creation_cost: 0,
          cache_read_cost: 0,
          input_tokens: 0,
          output_tokens: 0,
          cache_creation_tokens: 0,
          cache_read_tokens: 0,
          cache_creation_5m_tokens: 0,
          cache_creation_1h_tokens: 0,
          image_count: 1,
          image_size: null,
          image_input_size: null,
          image_output_size: null,
          image_size_source: null,
          image_size_breakdown: null,
          billing_mode: null,
          first_token_ms: null,
          duration_ms: 1,
          created_at: '2026-03-08T00:00:00Z',
          model: 'gpt-image-2',
        },
      ],
      total: 1,
      pages: 1,
    })
    getStatsByDateRange.mockResolvedValue({
      total_requests: 1,
      total_tokens: 0,
      total_cost: 0.2,
      avg_duration_ms: 1,
    })
    list.mockResolvedValue({ items: [] })

    const wrapper = mount(UsageView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
          TablePageLayout: TablePageLayoutStub,
          Pagination: true,
          EmptyState: true,
          Select: true,
          DateRangePicker: true,
          DataTable: DataTableStub,
          Icon: true,
          Teleport: true,
        },
      },
    })

    await flushPromises()
    await nextTick()

    const text = wrapper.text()
    expect(text).toContain('Image')
    expect(text).toContain('not recorded')
    expect(text).not.toContain('(2K)')
  })

  it('shows image billing metadata in the user cost tooltip', async () => {
    appStoreState.cachedPublicSettings = {
      usage_shell_config: buildUsageShellConfig(),
    }
    query.mockResolvedValue({
      items: [],
      total: 0,
      pages: 0,
    })
    getStatsByDateRange.mockResolvedValue({
      total_requests: 0,
      total_tokens: 0,
      total_cost: 0,
      avg_duration_ms: 0,
    })
    list.mockResolvedValue({ items: [] })

    const wrapper = mount(UsageView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
          TablePageLayout: TablePageLayoutStub,
          Pagination: true,
          EmptyState: true,
          Select: true,
          DateRangePicker: true,
          DataTable: DataTableStub,
          Icon: true,
          Teleport: true,
        },
      },
    })

    await flushPromises()

    const setupState = (wrapper.vm as any).$?.setupState
    setupState.tooltipData = {
      request_id: 'req-user-output-image',
      actual_cost: 0.8,
      total_cost: 0.8,
      rate_multiplier: 1,
      service_tier: null,
      input_cost: 0,
      output_cost: 0,
      cache_creation_cost: 0,
      cache_read_cost: 0,
      input_tokens: 0,
      output_tokens: 0,
      cache_creation_tokens: 0,
      cache_read_tokens: 0,
      billing_mode: null,
      image_count: 2,
      image_size: '4K',
      image_input_size: '1024x1024',
      image_output_size: '3840x2160',
      image_size_source: 'output',
      image_size_breakdown: { '4K': 2 },
    }
    setupState.tooltipVisible = true
    await nextTick()

    const text = wrapper.text()
    expect(text).toContain('Image count')
    expect(text).toContain('Billing size')
    expect(text).toContain('4K')
    expect(text).toContain('Size source')
    expect(text).toContain('Upstream output')
    expect(text).toContain('Input size')
    expect(text).toContain('1024x1024')
    expect(text).toContain('Output size')
    expect(text).toContain('3840x2160')
    expect(text).toContain('4K x 2')
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
