import { readFileSync } from 'node:fs'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import ImageGeneratorView from '../ImageGeneratorView.vue'
import { createImageWorkspaceTask } from '@/api/image-workspace'

const imageGeneratorViewSource = readFileSync('src/views/public/ImageGeneratorView.vue', 'utf8')
const zhLocaleSource = readFileSync('src/i18n/locales/zh.ts', 'utf8')
const enLocaleSource = readFileSync('src/i18n/locales/en.ts', 'utf8')
const fetchPublicSettings = vi.hoisted(() => vi.fn())
const currentLocale = vi.hoisted(() => ({ value: 'en' }))
const configuredWorkspaceShellConfig = vi.hoisted(() => JSON.stringify({
  en: {
    defaults: {
      catalogPath: '/configured-prompts',
      maxPromptLength: 12,
    },
    catalogLabel: 'Configured catalog',
    eyebrow: 'Configured workspace',
    title: 'Configured Image Workbench',
    heroDescription: 'Configured hero copy from public settings.',
    promptLabel: 'Configured prompt',
    promptPlaceholder: 'Configured placeholder',
    clearLabel: 'Configured clear',
    copyPromptLabel: 'Configured copy',
    workspaceTitle: 'Configured status',
    workspaceDescription: 'Configured workspace description.',
    workspaceStatus: 'Configured workspace status.',
    backToCatalogLabel: 'Configured back',
  },
}))
const imageWorkspaceTasks = vi.hoisted(() => ({
  items: [] as Array<{
    id: number
    status: string
    prompt: string
    negative_prompt: string
    model: string
    provider: string
    size: string
    quality: string
    style: string
    seed?: number
    batch_size: number
    template_id?: number
    worker_lease_until?: string
    cost_estimate: number
    balance_snapshot: number
    error_message: string
    result_json: string
    artifacts?: Array<{
      id: number
      task_id: number
      image_url: string
      storage_provider: string
      storage_key: string
      prompt: string
      mime_type: string
      width: number
      height: number
      file_size: number
      created_at: string
    }>
    created_at: string
    updated_at: string
  }>,
  total: 0,
  pages: 0,
}))

const appStoreState = vi.hoisted(() => ({
  publicSettingsLoaded: true,
  siteName: 'Cloudbase',
  siteLogo: '',
  cachedPublicSettings: {
    site_name: 'Cloudbase',
    site_logo: '',
    workspace_shell_config: configuredWorkspaceShellConfig,
  },
  fetchPublicSettings,
  showError: vi.fn(),
}))
const authStoreState = vi.hoisted(() => ({
  isAuthenticated: true,
  user: {
    balance: 2,
  },
  refreshUser: vi.fn(async () => ({
    balance: 2,
  })),
}))

vi.mock('@/stores', () => ({
  useAppStore: () => appStoreState,
  useAuthStore: () => authStoreState,
}))

vi.mock('@/api/image-workspace', () => ({
  createImageWorkspaceTask: vi.fn(),
  cancelImageWorkspaceTask: vi.fn(),
  downloadImageWorkspaceArtifact: vi.fn(),
  getImageWorkspaceTask: vi.fn(async (id: number) =>
    imageWorkspaceTasks.items.find((task) => task.id === id) ?? { id, status: 'succeeded', artifacts: [] },
  ),
  listImageWorkspaceModels: vi.fn(async () => [
    {
      id: 'configured-image-model',
      label: 'Configured Image Model',
      provider: 'openai',
      default_size: '768x768',
      default_quality: 'draft',
      sizes: ['768x768'],
      qualities: ['draft'],
      cost_per_image: 0.5,
      cost_hint: '1 credit / image',
      enabled: true,
    },
  ]),
  listImageWorkspaceTasks: vi.fn(async () => ({
    items: imageWorkspaceTasks.items,
    total: imageWorkspaceTasks.total,
    page: 1,
    page_size: 20,
    pages: imageWorkspaceTasks.pages,
  })),
}))

vi.mock('@/i18n', () => ({
  getLocale: () => currentLocale.value,
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({
      push: vi.fn(),
    }),
  }
})

vi.mock('vue-i18n', () => {
  const zh: Record<string, string> = {
    'imageWorkspace.negativePrompt': '反向提示词',
    'imageWorkspace.negativePromptPlaceholder': '不希望出现的元素',
    'imageWorkspace.model': '模型',
    'imageWorkspace.size': '尺寸',
    'imageWorkspace.quality': '质量',
    'imageWorkspace.batchSize': '批量',
    'imageWorkspace.style': '风格备注',
    'imageWorkspace.stylePlaceholder': '电影感、产品渲染',
    'imageWorkspace.startGenerating': '开始生图',
    'imageWorkspace.queuing': '排队中...',
    'imageWorkspace.loginRequired': '登录后可以创建生图任务',
    'imageWorkspace.estimatedCost': '预计消耗',
    'imageWorkspace.currentBalance': '当前余额',
    'imageWorkspace.insufficientBalance': '余额低于本次任务预计消耗',
    'imageWorkspace.topUp': '充值',
    'imageWorkspace.noTasks': '还没有生图任务。',
    'imageWorkspace.generationHistory': '生成历史',
    'imageWorkspace.imageTasks': '生图任务',
    'imageWorkspace.refresh': '刷新',
    'imageWorkspace.filterAll': '全部',
    'imageWorkspace.statusQueued': '排队中',
    'imageWorkspace.statusRunning': '生成中',
    'imageWorkspace.statusSucceeded': '已完成',
    'imageWorkspace.statusFailed': '失败',
    'imageWorkspace.statusCancelled': '已取消',
    'imageWorkspace.totalTasks': '共 {count} 个任务',
    'imageWorkspace.pageInfo': '第 {current} / {total} 页',
    'imageWorkspace.paginationAriaLabel': '分页',
    'imageWorkspace.prevPage': '上一页',
    'imageWorkspace.nextPage': '下一页',
    'imageWorkspace.batchLabel': '批量',
    'imageWorkspace.cost': '消耗',
    'imageWorkspace.balanceSnapshot': '余额快照',
    'imageWorkspace.workerGenerating': 'Worker 正在生成图片，页面会自动刷新。',
    'imageWorkspace.workerLeaseUntil': 'Worker 租约到期：',
    'imageWorkspace.imageLoadFailed': '图片加载失败',
    'imageWorkspace.goConsole': '去控制台',
    'imageWorkspace.preview': '预览',
    'imageWorkspace.viewFullSize': '查看大图',
    'imageWorkspace.download': '下载',
    'imageWorkspace.close': '关闭',
    'imageWorkspace.downloadOriginal': '下载原图',
    'imageWorkspace.cancel': '取消',
    'imageWorkspace.cancelling': '取消中...',
    'imageWorkspace.taskQueued': '生图任务 #{id} 已进入队列。',
    'imageWorkspace.errorUpstream404': '生图上游返回 404',
    'imageWorkspace.errorUpstreamApiKeyMissing': '生图上游 API Key 未配置',
    'imageWorkspace.errorUpstreamAuth': '生图上游鉴权失败',
    'imageWorkspace.errorRequestFailed': '请求失败',
    'imageWorkspace.taskId': '任务 #{id}',
    'imageWorkspace.imageId': '图片 #{id}',
  }
  return {
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => {
        let result = zh[key] ?? key
        if (params) {
          result = Object.entries(params).reduce(
            (str, [k, v]) => str.replace(`{${k}}`, String(v)),
            result,
          )
        }
        return result
      },
    }),
  }
})

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: vi.fn(),
  }),
}))

describe('ImageGeneratorView', () => {
  beforeEach(() => {
    currentLocale.value = 'en'
    fetchPublicSettings.mockReset()
    appStoreState.publicSettingsLoaded = true
    appStoreState.cachedPublicSettings = {
      site_name: 'Cloudbase',
      site_logo: '',
      workspace_shell_config: configuredWorkspaceShellConfig,
    }
    appStoreState.showError.mockReset()
    authStoreState.isAuthenticated = true
    authStoreState.user = {
      balance: 2,
    }
    authStoreState.refreshUser.mockClear()
    vi.mocked(createImageWorkspaceTask).mockReset()
    vi.mocked(createImageWorkspaceTask).mockResolvedValue({
      id: 901,
      status: 'queued',
      prompt: 'short prompt',
      negative_prompt: '',
      model: 'configured-image-model',
      provider: 'openai',
      size: '768x768',
      quality: 'auto',
      style: '',
      batch_size: 1,
      cost_estimate: 0.5,
      balance_snapshot: 2,
      error_message: '',
      result_json: '',
      created_at: '2026-07-07T01:00:00Z',
      updated_at: '2026-07-07T01:00:00Z',
    })
    imageWorkspaceTasks.items = []
    imageWorkspaceTasks.total = 0
    imageWorkspaceTasks.pages = 0
  })

  it('renders workspace shell copy from public settings', async () => {
    const wrapper = mount(ImageGeneratorView, {
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

    expect(wrapper.text()).toContain('Configured catalog')
    expect(wrapper.text()).toContain('Configured workspace')
    expect(wrapper.text()).toContain('Configured Image Workbench')
    expect(wrapper.text()).toContain('Configured hero copy from public settings.')
    expect(wrapper.text()).toContain('Configured prompt')
    expect(wrapper.find('textarea').attributes('placeholder')).toBe('Configured placeholder')
    expect(wrapper.text()).toContain('Configured clear')
    expect(wrapper.text()).toContain('Configured copy')
    expect(wrapper.text()).toContain('Configured back')
    expect(wrapper.text()).not.toContain('Configured status')
    expect(wrapper.text()).not.toContain('Configured workspace description.')
    expect(wrapper.text()).not.toContain('Configured workspace status.')
    await flushPromises()
    expect(wrapper.text()).toContain('Configured Image Model')
    expect(wrapper.text()).toContain('1 credit / image')
    expect(wrapper.text()).not.toContain('质量')
    expect(wrapper.text()).not.toContain('余额保护')
    expect(wrapper.text()).not.toContain('模板')
    expect(wrapper.text()).toContain('0 / 12')
    const catalogLinks = wrapper.findAll('a[href="/configured-prompts"]')
    expect(catalogLinks).toHaveLength(2)
  })

  it('does not render removed template and balance concepts from stale workspace shell copy', async () => {
    appStoreState.cachedPublicSettings = {
      site_name: 'Cloudbase',
      site_logo: '',
      workspace_shell_config: JSON.stringify({
        zh: {
          workspaceTitle: '任务与产物状态',
          workspaceDescription: '模型配置、任务历史、参数模板、余额预授权和产物存储已由 Cloudbase 生图工作台统一管理。',
          workspaceStatus: '登录后可创建真实生图任务。',
        },
      }),
    }
    currentLocale.value = 'zh'

    const wrapper = mount(ImageGeneratorView, {
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

    expect(wrapper.text()).not.toContain('任务与产物状态')
    expect(wrapper.text()).not.toContain('模型配置、任务历史和产物存储已由 Cloudbase 生图工作台统一管理。')
    expect(wrapper.text()).not.toContain('登录后可创建真实生图任务。')
    expect(wrapper.text()).not.toContain('参数模板')
    expect(wrapper.text()).not.toContain('余额预授权')
    expect(wrapper.text()).not.toContain('余额保护')
    expect(wrapper.text()).not.toContain('保存模板')
  })

  it('blocks generation when the selected model cost exceeds current balance', async () => {
    authStoreState.user = {
      balance: 0.25,
    }
    const wrapper = mount(ImageGeneratorView, {
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
    await wrapper.find('textarea').setValue('short prompt')
    await flushPromises()

    const generateButton = wrapper.findAll('button').find((button) => button.text() === '开始生图')
    expect(generateButton?.attributes('disabled')).toBeDefined()
    expect(wrapper.text()).toContain('余额低于本次任务预计消耗')
    expect(wrapper.text()).toContain('当前余额 0.25')
    expect(wrapper.text()).toContain('预计消耗 0.50')
    expect(wrapper.text()).toContain('充值')
  })

  it('hides quality controls and always submits auto quality', async () => {
    const wrapper = mount(ImageGeneratorView, {
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
    await wrapper.find('textarea').setValue('short')
    await flushPromises()

    expect(wrapper.text()).not.toContain('质量')
    expect(imageGeneratorViewSource).not.toContain('v-model="quality"')
    expect(imageGeneratorViewSource).not.toContain('selectedQualityOptions')
    expect(imageGeneratorViewSource).not.toContain('task.quality')

    const generateButton = wrapper.findAll('button').find((button) => button.text() === '开始生图')
    await generateButton?.trigger('click')
    await flushPromises()

    expect(createImageWorkspaceTask).toHaveBeenCalledWith(expect.objectContaining({
      prompt: 'short',
      model: 'configured-image-model',
      size: '768x768',
      quality: 'auto',
    }))
  })

  it('renders image artifacts with preserved aspect ratio, download action, and lightbox preview', async () => {
    imageWorkspaceTasks.items = [
      {
        id: 201,
        status: 'succeeded',
        prompt: 'wide cinematic image',
        negative_prompt: '',
        model: 'gpt-image-2',
        provider: 'openai',
        size: '1536x1024',
        quality: 'standard',
        style: '',
        batch_size: 1,
        cost_estimate: 0.5,
        balance_snapshot: 1.5,
        error_message: '',
        result_json: '',
        artifacts: [
          {
            id: 301,
            task_id: 201,
            image_url: 'https://static.example/image-301.png',
            storage_provider: 'r2',
            storage_key: 'image-301.png',
            prompt: 'wide cinematic image',
            mime_type: 'image/png',
            width: 1536,
            height: 1024,
            file_size: 2_097_152,
            created_at: '2026-06-27T01:00:00Z',
          },
        ],
        created_at: '2026-06-27T01:00:00Z',
        updated_at: '2026-06-27T01:00:00Z',
      },
    ]
    imageWorkspaceTasks.total = 1
    imageWorkspaceTasks.pages = 1

    const wrapper = mount(ImageGeneratorView, {
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

    const artifactCard = wrapper.find('[style*="1536 / 1024"]')
    expect(artifactCard.exists()).toBe(true)
    expect(artifactCard.attributes('style')).toContain('width: 100%')
    expect(imageGeneratorViewSource).toContain('grid min-w-0 gap-4 sm:grid-cols-2 xl:grid-cols-3')
    expect(imageGeneratorViewSource).not.toContain('columns-2')
    expect(imageGeneratorViewSource).toContain('const downloadingArtifactIds = ref<Record<number, boolean>>({})')
    expect(imageGeneratorViewSource).toContain(":disabled=\"isArtifactDownloading(artifact.id)\"")
    expect(imageGeneratorViewSource).toContain("isArtifactDownloading(lightboxArtifact.id) ? t('imageWorkspace.downloading') : t('imageWorkspace.downloadOriginal')")
    expect(imageGeneratorViewSource).toContain('void loadLocalArtifactBlobs(taskResult.items)')
    expect(imageGeneratorViewSource).toContain('const cachedBlobUrl = localArtifactBlobUrls.value[artifact.id]')
    expect(imageGeneratorViewSource).toContain('if (!cachedBlobUrl) {')
    expect(wrapper.find('img[src="https://static.example/image-301.png"]').classes()).toContain('object-contain')
    expect(wrapper.text()).toContain('共 1 个任务')

    // Click the image to open lightbox
    const imageButton = artifactCard.find('button')
    await imageButton.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('图片 #301')
    expect(wrapper.text()).toContain('1536×1024')
    expect(wrapper.text()).toContain('2.0 MB')
    expect(wrapper.text()).toContain('下载原图')
  })

  it('shows page-based pagination when there are multiple pages', async () => {
    imageWorkspaceTasks.items = [
      {
        id: 1,
        status: 'succeeded',
        prompt: 'test',
        negative_prompt: '',
        model: 'gpt-image-2',
        provider: 'openai',
        size: '1024x1024',
        quality: 'standard',
        style: '',
        batch_size: 1,
        cost_estimate: 0,
        balance_snapshot: 0,
        error_message: '',
        result_json: '',
        created_at: '2026-06-27T01:00:00Z',
        updated_at: '2026-06-27T01:00:00Z',
      },
    ]
    imageWorkspaceTasks.total = 45
    imageWorkspaceTasks.pages = 3

    const wrapper = mount(ImageGeneratorView, {
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

    expect(wrapper.text()).toContain('共 45 个任务')
    expect(wrapper.text()).toContain('第 1 / 3 页')

    const nav = wrapper.find('nav[aria-label="分页"]')
    expect(nav.exists()).toBe(true)

    const pageButtons = nav.findAll('button')
    expect(pageButtons.length).toBeGreaterThanOrEqual(4)

    const prevButton = pageButtons[0]
    expect(prevButton.attributes('disabled')).toBeDefined()

    const nextButton = pageButtons[pageButtons.length - 1]
    expect(nextButton.attributes('disabled')).toBeUndefined()
  })

  it('hides retry action for upstream policy violation failures', async () => {
    imageWorkspaceTasks.items = [
      {
        id: 35,
        status: 'failed',
        prompt: 'blocked prompt',
        negative_prompt: '',
        model: 'gpt-image-2',
        provider: 'openai',
        size: '1024x1024',
        quality: 'standard',
        style: '',
        batch_size: 1,
        cost_estimate: 0,
        balance_snapshot: 0,
        error_message: 'upstream safety_error: request violates content policy',
        result_json: '{"error":{"type":"safety_error","message":"flagged by content policy"}}',
        created_at: '2026-06-27T01:00:00Z',
        updated_at: '2026-06-27T01:00:00Z',
      },
      {
        id: 36,
        status: 'failed',
        prompt: 'temporary failure',
        negative_prompt: '',
        model: 'gpt-image-2',
        provider: 'openai',
        size: '1024x1024',
        quality: 'standard',
        style: '',
        batch_size: 1,
        cost_estimate: 0,
        balance_snapshot: 0,
        error_message: 'upstream timeout',
        result_json: '{}',
        created_at: '2026-06-27T01:00:00Z',
        updated_at: '2026-06-27T01:00:00Z',
      },
    ]
    imageWorkspaceTasks.total = 2
    imageWorkspaceTasks.pages = 1

    const wrapper = mount(ImageGeneratorView, {
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

    expect(wrapper.text()).toContain('任务 #35')
    expect(wrapper.text()).toContain('任务 #36')
    expect(wrapper.findAll('button').filter((button) => button.text() === 'imageWorkspace.retry')).toHaveLength(1)
    expect(imageGeneratorViewSource).toContain('function canRetryTask(task: ImageWorkspaceTask)')
    expect(imageGeneratorViewSource).toContain("v-if=\"canRetryTask(task)\"")
    expect(imageGeneratorViewSource).toContain("'safety_error'")
    expect(imageGeneratorViewSource).toContain("'content policy'")
    expect(imageGeneratorViewSource).toContain("'违规'")
  })

  it('does not show pagination when there is only one page', async () => {
    imageWorkspaceTasks.items = []
    imageWorkspaceTasks.total = 0
    imageWorkspaceTasks.pages = 0

    const wrapper = mount(ImageGeneratorView, {
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

    expect(wrapper.find('nav[aria-label="分页"]').exists()).toBe(false)
  })

  it('renders status filter buttons in the task section', async () => {
    imageWorkspaceTasks.items = []
    imageWorkspaceTasks.total = 0
    imageWorkspaceTasks.pages = 0

    const wrapper = mount(ImageGeneratorView, {
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

    const section = wrapper.find('[ref="taskListSection"]')
    expect(section.exists() || wrapper.text().includes('生成历史')).toBe(true)

    const filterButtons = wrapper.findAll('button').filter((btn) => {
      const text = btn.text()
      return text === '全部' || text === '排队中' || text === '生成中' || text === '已完成' || text === '失败'
    })
    expect(filterButtons.length).toBe(5)
  })

  it('does not embed default workspace shell copy in the Vue view', () => {
    expect(imageGeneratorViewSource).toContain('PublicDarkHeader')
    expect(imageGeneratorViewSource).toContain('container-class="max-w-6xl"')
    expect(imageGeneratorViewSource).toContain('public-template-container')
    expect(imageGeneratorViewSource).toContain("t('imageWorkspace.goConsole')")
    expect(imageGeneratorViewSource).toContain('useAuthRouteDefaults')
    expect(imageGeneratorViewSource).toContain('authRouteDefaults.value.purchasePath')
    expect(imageGeneratorViewSource).toContain(':to="rechargeRoute"')
    expect(imageGeneratorViewSource).toContain("query: { tab: 'recharge' }")
    expect(imageGeneratorViewSource).not.toContain(':to="authRouteDefaults.homePath"')
    expect(imageGeneratorViewSource).not.toContain('resolveHomePath')
    expect(imageGeneratorViewSource).not.toContain(':to="isAuthenticated ? dashboardPath : loginPath"')
    expect(imageGeneratorViewSource).not.toContain('const avatarUrl = computed(() => authStore.user?.avatar_url?.trim() || \'\')')
    expect(imageGeneratorViewSource).not.toContain(':aria-label="displayName"')
    expect(imageGeneratorViewSource).not.toContain('{{ userInitial }}')
    expect(imageGeneratorViewSource).toContain('resolveWorkspaceShellDefaults')
    expect(imageGeneratorViewSource).toContain(':to="catalogPath"')
    expect(imageGeneratorViewSource).toContain('workspaceShellDefaults.value.maxPromptLength')
    expect(imageGeneratorViewSource).toContain("from './imageGeneratorRuntime'")
    expect(imageGeneratorViewSource).toContain('applyImageGeneratorDraft')
    expect(imageGeneratorViewSource).toContain('resolveImageGeneratorCatalogPath')
    expect(imageGeneratorViewSource).not.toContain('MAX_PROMPT_LENGTH')
    expect(imageGeneratorViewSource).not.toContain('to="/home"')
    expect(imageGeneratorViewSource).not.toContain('to="/prompts"')
    expect(imageGeneratorViewSource).not.toContain("|| '/prompts'")
    expect(imageGeneratorViewSource).not.toContain('const EMPTY_WORKSPACE_SHELL')
    expect(imageGeneratorViewSource).not.toContain('EMPTY_WORKSPACE_SHELL')
    expect(imageGeneratorViewSource).not.toContain('DEFAULT_WORKSPACE_SHELL')
    expect(imageGeneratorViewSource).not.toContain('function formatTemplate')
    expect(imageGeneratorViewSource).not.toContain('formatWorkspaceShellTemplate')
    expect(imageGeneratorViewSource).toContain('defaultWorkspaceShellCopy')
    expect(imageGeneratorViewSource).toContain("clearLabel: t('imageWorkspace.clearLabel')")
    expect(imageGeneratorViewSource).toContain("copyPromptLabel: t('imageWorkspace.copyPromptLabel')")
    expect(imageGeneratorViewSource).toContain('...defaultWorkspaceShellCopy()')
    expect(imageGeneratorViewSource).not.toMatch(/resolveWorkspaceShellConfig\([^\n]*,[^\n]*,[^\n]*\)/)
    expect(imageGeneratorViewSource).not.toContain("catalogLabel: '提示词案例'")
    expect(imageGeneratorViewSource).not.toContain("eyebrow: '提示词工作台'")
    expect(imageGeneratorViewSource).not.toContain("title: 'AI 生图工作区'")
    expect(imageGeneratorViewSource).not.toContain("promptPlaceholder: '输入或从案例库导入提示词'")
    expect(imageGeneratorViewSource).not.toContain("copyPromptLabel: '复制提示词'")
    expect(imageGeneratorViewSource).not.toContain("backToCatalogLabel: '返回案例库'")
    expect(imageGeneratorViewSource).not.toContain("catalogLabel: 'Prompt catalog'")
    expect(imageGeneratorViewSource).not.toContain("eyebrow: 'Prompt Workspace'")
    expect(imageGeneratorViewSource).not.toContain("title: 'AI Image Workspace'")
    expect(imageGeneratorViewSource).not.toContain("promptPlaceholder: 'Enter a prompt or import one from the catalog'")
    expect(imageGeneratorViewSource).not.toContain("copyPromptLabel: 'Copy prompt'")
    expect(imageGeneratorViewSource).not.toContain("backToCatalogLabel: 'Back to catalog'")
  })

  it('keeps admin workspace shell placeholders aligned with real image tasks', () => {
    expect(zhLocaleSource).toContain('AI 生图工作台')
    expect(zhLocaleSource).toContain('直接创建生图任务')
    expect(zhLocaleSource).not.toContain('参数模板、余额预授权')
    expect(zhLocaleSource).not.toContain('任务与产物状态')
    expect(zhLocaleSource).not.toContain('当前版本不会直接发起模型调用')
    expect(zhLocaleSource).not.toContain('后续 Cloudbase 原生生成流程')

    expect(enLocaleSource).toContain('AI Image Workspace')
    expect(enLocaleSource).toContain('create an image task')
    expect(enLocaleSource).not.toContain('Task and artifact status')
    expect(enLocaleSource).not.toContain('Prompt Staging Area')
    expect(enLocaleSource).not.toContain('This version does not call a model directly')
    expect(enLocaleSource).not.toContain('future Cloudbase-native generation flow')
  })
})
