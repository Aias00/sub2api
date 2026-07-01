import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import WeChatExportView from '../WeChatExportView.vue'
import { bindWeChatAccount, createWeChatExportTask, getWeChatSession, importWeChatArticleLink, searchWeChatAccounts, syncWeChatAccount } from '@/api/wechat-export'

const weChatExportViewSource = readFileSync(resolve(process.cwd(), 'src/views/public/WeChatExportView.vue'), 'utf8')

const authStoreState = vi.hoisted(() => ({
  isAuthenticated: false,
  isAdmin: false,
}))

const appStoreState = vi.hoisted(() => ({
  siteName: 'Sub2API',
  siteLogo: '',
  cachedPublicSettings: {
    site_name: 'cloudbase',
    site_logo: '',
  },
}))

vi.mock('@/stores', () => ({
  useAuthStore: () => authStoreState,
  useAppStore: () => appStoreState,
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
      locale: { value: 'zh' },
      te: () => false,
    }),
  }
})

vi.mock('@/api/wechat-export', () => ({
  getWeChatSession: vi.fn(async () => ({ id: 1, status: 'ready', login_account_name: 'Test WeChat' })),
  createWeChatQRCodeSession: vi.fn(),
  pollWeChatSession: vi.fn(),
  logoutWeChatSession: vi.fn(),
  validateWeChatSession: vi.fn(),
  listWeChatArticles: vi.fn(async () => ({ items: [], total: 0, page: 1, page_size: 50, pages: 1 })),
  listWeChatExportTasks: vi.fn(async () => ({ items: [], total: 0, page: 1, page_size: 20, pages: 1 })),
  getWeChatExportWorkerStatus: vi.fn(async () => null),
  importWeChatArticleLink: vi.fn(),
  createWeChatExportTask: vi.fn(),
  listWeChatExportArtifacts: vi.fn(async () => []),
  listWeChatExportTaskLogs: vi.fn(async () => []),
  cancelWeChatExportTask: vi.fn(),
  retryWeChatExportTask: vi.fn(),
  downloadWeChatExportTaskZip: vi.fn(),
  downloadWeChatExportArtifact: vi.fn(),
  quoteWeChatExportTask: vi.fn(async () => ({ estimated_credits: 0 })),
  searchWeChatAccounts: vi.fn(async () => []),
  bindWeChatAccount: vi.fn(async () => ({
    account: { id: 1, fakeid: 'test-account', nickname: 'Test Account' },
    sync_required: true,
  })),
  syncWeChatAccount: vi.fn(async () => ({
    account: { id: 1, fakeid: 'test-account' },
    status: 'synced',
    result: { synced_count: 10, total_count: 10, has_more: false },
  })),
}))

describe('WeChatExportView', () => {
  beforeEach(() => {
    authStoreState.isAuthenticated = true
    authStoreState.isAdmin = false
    vi.mocked(getWeChatSession).mockResolvedValue({ id: 1, status: 'ready', login_account_name: 'Test WeChat' })
    vi.mocked(importWeChatArticleLink).mockReset()
    vi.mocked(createWeChatExportTask).mockReset()
  })

  it('renders the WeChat export workspace controls', () => {
    const wrapper = mount(WeChatExportView, {
      global: {
        stubs: {
          RouterLink: {
            props: ['to'],
            template: '<a :href="typeof to === \'string\' ? to : to.path"><slot /></a>',
          },
          LocaleSwitcher: { template: '<div>locale</div>' },
        },
      },
    })

    // The mock t() returns the key itself, so we check for i18n keys
    expect(wrapper.text()).toContain('wechatExport.pageTitle')
    expect(wrapper.text()).toContain('wechatExport.pageHint')
    // UI has Chinese text: 文章列表, 导出操作, 任务监控
    expect(wrapper.text()).toContain('文章列表')
    expect(wrapper.text()).toContain('导出操作')
    expect(wrapper.text()).not.toContain('创建导出任务')
    expect(wrapper.text()).toContain('任务监控')
    expect(wrapper.find('input[type="url"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('html')
    expect(wrapper.text()).toContain('markdown')
    expect(wrapper.text()).toContain('互动数据')
    expect(wrapper.findAll('input[type="checkbox"]').length).toBeGreaterThanOrEqual(3)
  })

  it('aligns the shared header with the main content grid', () => {
    expect(weChatExportViewSource).toContain('container-class="max-w-6xl"')
    expect(weChatExportViewSource).toContain('public-template-container')
  })

  it('prompts anonymous visitors to log in', () => {
    authStoreState.isAuthenticated = false
    const wrapper = mount(WeChatExportView, {
      global: {
        stubs: {
          RouterLink: {
            props: ['to'],
            template: '<a :href="typeof to === \'string\' ? to : to.path"><slot /></a>',
          },
          LocaleSwitcher: { template: '<div>locale</div>' },
        },
      },
    })

    // The mock t() returns the key itself, so we check for the i18n key
    expect(wrapper.text()).toContain('wechatExport.loginRequired')
  })

  it('keeps WeChat warning copy readable on the light business shell', () => {
    expect(weChatExportViewSource).toContain('wechat-export-warning')
    expect(weChatExportViewSource).toContain('color: rgb(120, 53, 15) !important;')
    expect(weChatExportViewSource).not.toContain('text-amber-100/85')
    expect(weChatExportViewSource).not.toContain('text-amber-50')
  })

  it('keeps task batch actions readable when disabled', () => {
    expect(weChatExportViewSource).toContain('wechat-task-action-danger')
    expect(weChatExportViewSource).toContain('.wechat-task-action:disabled')
    expect(weChatExportViewSource).toContain('color: rgb(100, 116, 139) !important;')
    expect(weChatExportViewSource).not.toContain('border border-amber-200/20 px-3 py-1.5 text-sm font-semibold text-amber-100')
  })

  it('links insufficient export balance warnings to the configured recharge page', () => {
    expect(weChatExportViewSource).toContain('useAuthRouteDefaults')
    expect(weChatExportViewSource).toContain('authRouteDefaults.value.purchasePath')
    expect(weChatExportViewSource).toContain(':to="rechargeRoute"')
    expect(weChatExportViewSource).toContain("query: { tab: 'recharge' }")
    expect(weChatExportViewSource).toContain('余额不足，请充值后再创建任务')
    expect(weChatExportViewSource).toContain('去充值')
  })

  it('keeps worker status messages readable on the light business shell', () => {
    expect(weChatExportViewSource).toContain('wechat-worker-status-message')
    expect(weChatExportViewSource).toContain('color: rgb(154, 52, 18) !important;')
    expect(weChatExportViewSource).toContain("return 'text-amber-700'")
    expect(weChatExportViewSource).not.toContain('border border-amber-200/20 bg-amber-300/10 px-3 py-2 text-xs text-amber-100')
  })

  it('reports partial failures during batch ZIP downloads', () => {
    expect(weChatExportViewSource).toContain('let successCount = 0')
    expect(weChatExportViewSource).toContain('const failures: string[] = []')
    expect(weChatExportViewSource).toContain('已下载 ${successCount}/${downloadableTaskIds.length} 个任务 ZIP。')
    expect(weChatExportViewSource).toContain('部分任务 ZIP 下载失败')
    expect(weChatExportViewSource).not.toContain('已下载 ${downloadableTaskIds.length} 个任务 ZIP。')
  })

  it('loads the first WeChat article page and keeps remote pagination reachable', () => {
    expect(weChatExportViewSource).toContain('listWeChatArticles({ page: 1, page_size: articleRemotePageSize })')
    expect(weChatExportViewSource).toContain('const articleRemotePage = ref(1)')
    expect(weChatExportViewSource).toContain('const articleRemotePages = ref(0)')
    expect(weChatExportViewSource).toContain('const hasMoreRemoteArticles = computed')
    expect(weChatExportViewSource).toContain('async function loadMoreWeChatArticles()')
    expect(weChatExportViewSource).toContain('listWeChatArticles({ page: nextPage, page_size: articleRemotePageSize })')
    expect(weChatExportViewSource).toContain('加载更多文章')
  })

  describe('Search result bind auto-sync contract', () => {
    beforeEach(() => {
      vi.mocked(searchWeChatAccounts).mockReset()
      vi.mocked(searchWeChatAccounts).mockResolvedValue([])
      vi.mocked(bindWeChatAccount).mockReset()
      vi.mocked(bindWeChatAccount).mockResolvedValue({
        account: { id: 1, fakeid: 'test-account', nickname: 'Test Account' },
        sync_required: true,
      })
      vi.mocked(syncWeChatAccount).mockReset()
      vi.mocked(syncWeChatAccount).mockResolvedValue({
        account: { id: 1, fakeid: 'test-account' },
        status: 'synced',
        result: { synced_count: 10, total_count: 10, has_more: false },
      })
    })

    it('does not render manual fakeid binding controls', () => {
      const wrapper = mount(WeChatExportView, {
        global: {
          stubs: {
            RouterLink: {
              props: ['to'],
              template: '<a :href="typeof to === \'string\' ? to : to.path"><slot /></a>',
            },
            LocaleSwitcher: { template: '<div>locale</div>' },
          },
        },
      })

      expect(wrapper.find('input[placeholder="fakeid"]').exists()).toBe(false)
      expect(wrapper.find('input[placeholder="名称"]').exists()).toBe(false)
      expect(wrapper.text()).toContain('公众号管理')
    })

    it('calls syncWeChatAccount when binding a search result returns sync_required=true', async () => {
      vi.mocked(searchWeChatAccounts).mockImplementation(async (_query?: string, remote?: boolean) => (
        remote ? [{ id: 1, fakeid: 'test-fakeid', nickname: 'Test Account' }] : []
      ))
      vi.mocked(bindWeChatAccount).mockResolvedValueOnce({
        account: { id: 1, fakeid: 'test-fakeid', nickname: 'Test Account' },
        sync_required: true,
      })

      const wrapper = mount(WeChatExportView, {
        global: {
          stubs: {
            RouterLink: {
              props: ['to'],
              template: '<a :href="typeof to === \'string\' ? to : to.path"><slot /></a>',
            },
            LocaleSwitcher: { template: '<div>locale</div>' },
          },
        },
      })

      await flushPromises()
      vi.mocked(searchWeChatAccounts).mockClear()

      await wrapper.find('input[placeholder="搜索公众号"]').setValue('test')
      const searchButton = wrapper.findAll('button').find(btn => btn.text().includes('查找'))
      await searchButton?.trigger('click')
      await flushPromises()

      expect(wrapper.text()).toContain('Test Account')
      const bindButton = wrapper.findAll('button').find(btn => btn.text() === '绑定')
      expect(bindButton).toBeTruthy()
      await bindButton?.trigger('click')
      await flushPromises()

      expect(searchWeChatAccounts).toHaveBeenCalledWith('test', true)
      expect(bindWeChatAccount).toHaveBeenCalledWith({
        fakeid: 'test-fakeid',
        nickname: 'Test Account',
        alias: undefined,
        avatar: undefined,
        description: undefined,
      })
      expect(syncWeChatAccount).toHaveBeenCalledWith('test-fakeid', 0)
    })

    it('does NOT call syncWeChatAccount when binding a search result returns sync_required=false', async () => {
      vi.mocked(searchWeChatAccounts).mockImplementation(async (_query?: string, remote?: boolean) => (
        remote ? [{ id: 2, fakeid: 'no-sync-fakeid', nickname: 'No Sync' }] : []
      ))
      vi.mocked(bindWeChatAccount).mockResolvedValueOnce({
        account: { id: 2, fakeid: 'no-sync-fakeid', nickname: 'No Sync' },
        sync_required: false,
      })

      const wrapper = mount(WeChatExportView, {
        global: {
          stubs: {
            RouterLink: {
              props: ['to'],
              template: '<a :href="typeof to === \'string\' ? to : to.path"><slot /></a>',
            },
            LocaleSwitcher: { template: '<div>locale</div>' },
          },
        },
      })

      await flushPromises()

      await wrapper.find('input[placeholder="搜索公众号"]').setValue('no-sync')
      const searchButton = wrapper.findAll('button').find(btn => btn.text().includes('查找'))
      await searchButton?.trigger('click')
      await flushPromises()

      expect(wrapper.text()).toContain('No Sync')
      const bindButton = wrapper.findAll('button').find(btn => btn.text() === '绑定')
      expect(bindButton).toBeTruthy()
      await bindButton?.trigger('click')
      await flushPromises()

      expect(bindWeChatAccount).toHaveBeenCalled()
      expect(syncWeChatAccount).not.toHaveBeenCalled()
    })

    it('does not show 0/0 while sync total is still unknown', async () => {
      vi.mocked(searchWeChatAccounts).mockImplementation(async (_query?: string, remote?: boolean) => (
        remote ? [{ id: 3, fakeid: 'sync-fakeid', nickname: 'Sync Account' }] : []
      ))
      vi.mocked(bindWeChatAccount).mockResolvedValueOnce({
        account: { id: 3, fakeid: 'sync-fakeid', nickname: 'Sync Account' },
        sync_required: true,
      })
      vi.mocked(syncWeChatAccount).mockImplementationOnce(() => new Promise(() => {}))

      const wrapper = mount(WeChatExportView, {
        global: {
          stubs: {
            RouterLink: {
              props: ['to'],
              template: '<a :href="typeof to === \'string\' ? to : to.path"><slot /></a>',
            },
            LocaleSwitcher: { template: '<div>locale</div>' },
          },
        },
      })

      await flushPromises()
      vi.mocked(searchWeChatAccounts).mockClear()

      await wrapper.find('input[placeholder="搜索公众号"]').setValue('sync')
      const searchButton = wrapper.findAll('button').find(btn => btn.text().includes('查找'))
      await searchButton?.trigger('click')
      await flushPromises()

      const bindButton = wrapper.findAll('button').find(btn => btn.text() === '绑定')
      expect(bindButton).toBeTruthy()
      await bindButton?.trigger('click')
      await flushPromises()

      expect(wrapper.text()).toContain('正在同步文章...: 已完成 0 篇')
      expect(wrapper.text()).not.toContain('0/0')
    })

    it('preserves the known total when continuing a multi-batch sync', () => {
      expect(weChatExportViewSource).toContain("const previousProgress = syncingProgress.value?.fakeid === fakeid ? syncingProgress.value : null")
      expect(weChatExportViewSource).toContain('const knownTotal = previousProgress && previousProgress.total > 0 ? previousProgress.total : 0')
      expect(weChatExportViewSource).toContain('total: knownTotal')
    })

    it('polls to confirm sync results after a transient request timeout', () => {
      expect(weChatExportViewSource).toContain('function isTransientSyncRequestError(error: unknown)')
      expect(weChatExportViewSource).toContain('async function pollSyncConfirmation')
      expect(weChatExportViewSource).toContain('同步请求已提交，前端连接超时，正在轮询确认结果...')
      expect(weChatExportViewSource).toContain('const confirmed = await pollSyncConfirmation(fakeid, startFrom, baselineArticleTotal, baselineLastSyncedAt)')
      expect(weChatExportViewSource).not.toContain("errorMessage.value = error instanceof Error ? error.message : '请求失败'")
    })

    it('requires a ready WeChat session before account and export actions', async () => {
      vi.mocked(getWeChatSession).mockResolvedValueOnce({ id: 9, status: 'expired' })
      vi.mocked(searchWeChatAccounts).mockImplementation(async (_query?: string, remote?: boolean) => (
        remote ? [{ id: 4, fakeid: 'expired-fakeid', nickname: 'Expired Session Account' }] : [
          { id: 4, fakeid: 'expired-fakeid', nickname: 'Expired Session Account', alias: '', avatar: '', description: '', is_active: true },
        ]
      ))

      const wrapper = mount(WeChatExportView, {
        global: {
          stubs: {
            RouterLink: {
              props: ['to'],
              template: '<a :href="typeof to === \'string\' ? to : to.path"><slot /></a>',
            },
            LocaleSwitcher: { template: '<div>locale</div>' },
          },
        },
      })

      await flushPromises()

      expect(wrapper.text()).toContain('微信会话未登录或已失效')
      const searchInput = wrapper.find('input[placeholder="搜索公众号"]')
      expect(searchInput.attributes('disabled')).toBeDefined()
      const importInput = wrapper.find('input[placeholder="粘贴文章链接 mp.weixin.qq.com/s/..."]')
      expect(importInput.attributes('disabled')).toBeDefined()

      await searchInput.setValue('test')
      const searchButton = wrapper.findAll('button').find(btn => btn.text().includes('查找'))
      await searchButton?.trigger('click')
      expect(searchWeChatAccounts).not.toHaveBeenCalledWith('test', true)

      const exportButton = wrapper.findAll('button').find(btn => btn.text().includes('导出 '))
      expect(exportButton?.attributes('disabled')).toBeDefined()
      expect(importWeChatArticleLink).not.toHaveBeenCalled()
      expect(createWeChatExportTask).not.toHaveBeenCalled()
    })
  })
})
