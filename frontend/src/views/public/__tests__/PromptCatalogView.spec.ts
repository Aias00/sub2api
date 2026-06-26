import { readFileSync } from 'node:fs'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import PromptCatalogView from '../PromptCatalogView.vue'

const promptCatalogViewSource = readFileSync('src/views/public/PromptCatalogView.vue', 'utf8')
const listCases = vi.hoisted(() => vi.fn())
const importTwitter = vi.hoisted(() => vi.fn())
const fetchPublicSettings = vi.hoisted(() => vi.fn())
const copyToClipboard = vi.hoisted(() => vi.fn())
const saveImageGeneratorDraft = vi.hoisted(() => vi.fn())

const authStoreState = vi.hoisted(() => ({
  isAuthenticated: true,
  isAdmin: true,
  checkAuth: vi.fn(),
}))

const appStoreState = vi.hoisted(() => ({
  publicSettingsLoaded: true,
  siteName: 'Sub2API',
  siteLogo: '',
  cachedPublicSettings: {
    site_name: 'Sub2API',
    site_logo: '',
    prompt_catalog_shell_config: JSON.stringify({
      en: {
        defaults: {
          sourceType: 'case',
          hasImage: true,
          pageSize: 12,
          sortBy: 'title',
          sortOrder: 'asc',
          generatorPath: '/configured-generator',
          generatorDraftSource: 'configured-catalog-source',
          importXAuto: false,
        },
        labels: {
          accountActionAuthenticated: 'Configured dashboard',
          eyebrow: 'Configured eyebrow',
          title: 'Configured title',
          description: 'Configured description',
          caseTitle: 'Configured shell cases',
          caseDescription: 'Configured shell case description',
          templateTitle: 'Configured shell templates',
          templateDescription: 'Configured shell template description',
          total: 'Configured total',
          sources: 'Configured sources',
          cases: 'Configured cases',
          templates: 'Configured templates',
          search: 'Configured search',
          searchPlaceholder: 'Configured search placeholder',
          caseOnly: 'Configured cases only',
          templateOnly: 'Configured templates only',
          allTypes: 'Configured all types',
          allSources: 'Configured all sources',
          allCategories: 'Configured all categories',
          hasImage: 'Configured images only',
          resultPrefix: 'Configured results',
          page: 'Configured page',
          previous: 'Configured previous',
          next: 'Configured next',
          emptyTitle: 'Configured empty title',
          emptyDescription: 'Configured empty description',
          noImage: 'Configured no image',
          source: 'Configured source',
          charUnit: 'characters',
          details: 'Open',
          prompt: 'Configured prompt label',
          copyPrompt: 'Configured copy',
          promptCopied: 'Configured copied',
          generate: 'Configured generator',
          importTitle: 'Configured import title',
          importDescription: 'Configured import description',
          importProviderX: 'Configured X source',
          importPlaceholder: 'Configured import placeholder',
          importAction: 'Configured import',
          importing: 'Configured importing',
          importSuccess: 'Configured import success',
          importWarnings: 'Configured import warnings',
          loadError: 'Configured load error',
        },
      },
    }),
  },
  fetchPublicSettings,
}))

vi.mock('@/stores', () => ({
  useAppStore: () => appStoreState,
  useAuthStore: () => authStoreState,
}))

vi.mock('@/i18n', () => ({
  getLocale: () => 'en',
}))

vi.mock('@/api/prompts', async () => {
  const actual = await vi.importActual<typeof import('@/api/prompts')>('@/api/prompts')
  return {
    ...actual,
    promptsAPI: {
      listCases,
      importTwitter,
    },
  }
})

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard,
  }),
}))

vi.mock('@/utils/imageGeneratorDraft', () => ({
  saveImageGeneratorDraft,
}))

describe('PromptCatalogView', () => {
  beforeEach(() => {
    fetchPublicSettings.mockReset()
    copyToClipboard.mockReset()
    saveImageGeneratorDraft.mockReset()
    importTwitter.mockReset()
    listCases.mockReset()
    authStoreState.checkAuth.mockReset()
    listCases.mockResolvedValue({
      data: {
        items: [
          {
            id: 'case-1',
            title: 'Prompt case',
            prompt: 'A configurable prompt',
            prompt_preview: 'A configurable prompt',
            source_type: 'case',
            source_project: 'x',
            source_label: 'X',
            source_display_label: 'API Source',
            source_url: '',
            category: 'portrait',
            tags: [],
            model_tags: [],
            display_tags: [],
            all_tags: ['OpenAI Image', 'editorial'],
            visible_tags: ['OpenAI Image'],
            image_url: '',
            primary_image_url: 'https://static.example.com/primary.jpg',
            image_urls: [],
            prompt_char_count: 42,
            created_at: '2026-06-18T00:00:00Z',
            imported_at: '2026-06-18T00:00:00Z',
          },
        ],
        summary: {
          total: 1,
          case_count: 1,
          template_count: 0,
          source_count: 1,
          sources: [{ value: 'x', label: 'X', count: 1, display_label: 'API Source Filter (1)' }],
          categories: [{ value: 'portrait', count: 1, display_label: 'API Category Filter (1)' }],
        },
        total: 1,
        page: 1,
        pages: 1,
      },
    })
  })

  it('uses prompt catalog shell labels for detail metadata', async () => {
    const wrapper = mount(PromptCatalogView, {
      global: {
        stubs: {
          RouterLink: {
            props: ['to'],
            template: '<a :href="typeof to === \'string\' ? to : to.path"><slot /></a>',
          },
          LocaleSwitcher: { template: '<div>locale</div>' },
          BaseDialog: {
            props: ['show', 'title'],
            template: '<div v-if="show"><h2>{{ title }}</h2><slot /></div>',
          },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('Configured shell cases')
    expect(wrapper.text()).toContain('Configured shell case description')
    expect(wrapper.text()).toContain('Configured dashboard')
    expect(wrapper.text()).toContain('Configured eyebrow')
    expect(wrapper.text()).toContain('Configured X source')
    expect(wrapper.text()).toContain('API Source')
    expect(wrapper.text()).toContain('API Source Filter (1)')
    expect(wrapper.text()).toContain('API Category Filter (1)')
    expect(wrapper.find('img[src="https://static.example.com/primary.jpg"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('OpenAI Image')
    expect(wrapper.text()).not.toContain('editorial')
    const detailsButton = wrapper.findAll('button').find((button) => button.text() === 'Open')
    expect(detailsButton).toBeTruthy()
    await detailsButton!.trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('characters')
    expect(wrapper.text()).toContain('42 characters')
    expect(wrapper.text()).toContain('editorial')
    expect(listCases).toHaveBeenCalledWith(expect.objectContaining({
      source_type: 'case',
      has_image: true,
      page_size: 12,
      sort_by: 'title',
      sort_order: 'asc',
    }))
  })

  it('opens the configured generator route with configured draft source', async () => {
    const originalLocation = window.location
    const assign = vi.fn()
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: { ...originalLocation, assign },
    })

    try {
      const wrapper = mount(PromptCatalogView, {
        global: {
          stubs: {
            RouterLink: {
              props: ['to'],
              template: '<a :href="typeof to === \'string\' ? to : to.path"><slot /></a>',
            },
            LocaleSwitcher: { template: '<div>locale</div>' },
            BaseDialog: {
              props: ['show', 'title'],
              template: '<div v-if="show"><h2>{{ title }}</h2><slot /></div>',
            },
          },
        },
      })

      await flushPromises()

      const generatorButton = wrapper.findAll('button').find((button) => button.text() === 'Configured generator')
      expect(generatorButton).toBeTruthy()
      await generatorButton!.trigger('click')

      expect(saveImageGeneratorDraft).toHaveBeenCalledWith({
        prompt: 'A configurable prompt',
        title: 'Prompt case',
        sourcePromptId: 'case-1',
        source: 'configured-catalog-source',
      })
      expect(assign).toHaveBeenCalledWith('/configured-generator')
    } finally {
      Object.defineProperty(window, 'location', {
        configurable: true,
        value: originalLocation,
      })
    }
  })

  it('uses configured X import automation mode from prompt catalog shell defaults', async () => {
    importTwitter.mockResolvedValue({
      data: {
        item: {
          id: 'case-imported',
          title: 'Imported case',
          prompt: 'Imported prompt',
          prompt_preview: 'Imported prompt',
          source_type: 'case',
          source_project: 'x',
          source_label: 'X',
          source_display_label: 'X',
          source_url: 'https://x.com/example/status/1',
          category: '',
          tags: [],
          model_tags: [],
          display_tags: [],
          all_tags: [],
          visible_tags: [],
          image_url: '',
          primary_image_url: '',
          image_urls: [],
          prompt_char_count: 15,
          created_at: '2026-06-18T00:00:00Z',
          imported_at: '2026-06-18T00:00:00Z',
        },
        warnings: [],
      },
    })

    const wrapper = mount(PromptCatalogView, {
      global: {
        stubs: {
          RouterLink: {
            props: ['to'],
            template: '<a :href="typeof to === \'string\' ? to : to.path"><slot /></a>',
          },
          LocaleSwitcher: { template: '<div>locale</div>' },
          BaseDialog: {
            props: ['show', 'title'],
            template: '<div v-if="show"><h2>{{ title }}</h2><slot /></div>',
          },
        },
      },
    })

    await flushPromises()

    await wrapper.find('input[type="url"]').setValue('https://x.com/example/status/1')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(importTwitter).toHaveBeenCalledWith({
      url: 'https://x.com/example/status/1',
      x_auto: false,
    })
    expect(wrapper.text()).toContain('Configured import success: Imported case')
  })

  it('does not embed default prompt catalog shell copy in the Vue view', () => {
    expect(promptCatalogViewSource).toContain('useAuthRouteDefaults')
    expect(promptCatalogViewSource).toContain(':to="authRouteDefaults.homePath"')
    expect(promptCatalogViewSource).not.toContain('to="/home"')
    expect(promptCatalogViewSource).not.toContain("isAuthenticated ? dashboardPath : '/login'")
    expect(promptCatalogViewSource).not.toContain("authStore.isAdmin ? '/admin/dashboard' : '/dashboard'")
    expect(promptCatalogViewSource).not.toContain('EMPTY_PROMPT_CATALOG_COPY')
    expect(promptCatalogViewSource).not.toContain('FALLBACK_PROMPT_CATALOG_COPY')
    expect(promptCatalogViewSource).not.toContain("title: 'Prompt Catalog'")
    expect(promptCatalogViewSource).not.toContain("description: 'Browse prompt cases from the shared prompt API.'")
    expect(promptCatalogViewSource).not.toContain("searchPlaceholder: 'Search prompts'")
    expect(promptCatalogViewSource).toContain('resolvePromptCatalogPageTitle')
    expect(promptCatalogViewSource).toContain('resolvePromptCatalogPageDescription')
    expect(promptCatalogViewSource).not.toContain('prompt_cases_title')
    expect(promptCatalogViewSource).not.toContain('prompt_cases_description')
    expect(promptCatalogViewSource).not.toContain('prompt_templates_title')
    expect(promptCatalogViewSource).not.toContain('prompt_templates_description')
    expect(promptCatalogViewSource).not.toContain("importTitle: 'Import from link'")
    expect(promptCatalogViewSource).not.toContain("importPlaceholder: 'Paste an X/Twitter post URL'")
    expect(promptCatalogViewSource).not.toContain('<option value="x">X / Twitter</option>')
    expect(promptCatalogViewSource).not.toContain('x_auto: true')
    expect(promptCatalogViewSource).toContain('x_auto: resolvePromptCatalogImportXAuto(catalogDefaults.value)')
    expect(promptCatalogViewSource).not.toContain('const PAGE_SIZE = 24')
    expect(promptCatalogViewSource).not.toContain('|| 24')
    expect(promptCatalogViewSource).not.toContain("sort_by: 'imported_at'")
    expect(promptCatalogViewSource).not.toContain("sort_order: 'desc'")
    expect(promptCatalogViewSource).not.toContain("|| 'imported_at'")
    expect(promptCatalogViewSource).not.toContain("|| 'desc'")
    expect(promptCatalogViewSource).not.toContain("window.location.assign('/image-generator')")
    expect(promptCatalogViewSource).not.toContain("source: 'sub2api-vue-prompt-catalog'")
    expect(promptCatalogViewSource).not.toContain("|| '/image-generator'")
    expect(promptCatalogViewSource).not.toContain("|| 'sub2api-vue-prompt-catalog'")
    expect(promptCatalogViewSource).not.toContain("loadError: 'Failed to load prompt cases'")
    expect(promptCatalogViewSource).not.toContain('function isSupportedTwitterURL')
    expect(promptCatalogViewSource).not.toContain('copy.value.importInvalidUrl')
    expect(promptCatalogViewSource).not.toContain('function cardImageUrl')
    expect(promptCatalogViewSource).not.toContain('function sourceDisplayLabel')
    expect(promptCatalogViewSource).not.toContain('function promptCharCount')
    expect(promptCatalogViewSource).not.toContain('function visibleTags')
    expect(promptCatalogViewSource).not.toContain('function allTags')
    expect(promptCatalogViewSource).not.toContain('data.summary || emptySummary()')
    expect(promptCatalogViewSource).not.toContain('data.pages || 1')
    expect(promptCatalogViewSource).toContain('item.primary_image_url')
    expect(promptCatalogViewSource).toContain('item.source_display_label')
    expect(promptCatalogViewSource).toContain('item.visible_tags')
    expect(promptCatalogViewSource).toContain('selectedPrompt.prompt_char_count')
    expect(promptCatalogViewSource).toContain('selectedPrompt.all_tags')
  })
})
