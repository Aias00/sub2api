<template>
  <div class="min-h-screen bg-[#101114] text-white">
    <header class="border-b border-white/10 bg-[#15171d] px-6 py-5">
      <nav class="mx-auto flex max-w-7xl items-center justify-between gap-4">
        <RouterLink :to="authRouteDefaults.homePath" class="flex min-w-0 items-center gap-3">
          <div v-if="siteLogo" class="h-9 w-9 shrink-0 overflow-hidden rounded-xl border border-white/10 bg-white/5">
            <img :src="siteLogo" alt="Logo" class="h-full w-full object-contain" />
          </div>
          <span class="truncate text-sm font-semibold text-white">{{ siteName }}</span>
        </RouterLink>

        <div class="flex items-center gap-3">
          <LocaleSwitcher />
          <RouterLink
            :to="isAuthenticated ? dashboardPath : loginPath"
            class="inline-flex items-center rounded-full border border-white/10 bg-white px-4 py-2 text-sm font-semibold text-slate-950 transition hover:bg-white/90"
          >
            {{ copy.accountAction }}
          </RouterLink>
        </div>
      </nav>
    </header>

    <main class="px-6 py-10 sm:py-14">
      <div class="mx-auto max-w-7xl">
        <section class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_320px] lg:items-end">
          <div>
            <p class="text-sm font-semibold uppercase tracking-[0.22em] text-cyan-200/70">
              {{ copy.eyebrow }}
            </p>
            <h1 class="mt-4 max-w-4xl text-4xl font-black leading-tight text-white sm:text-5xl">
              {{ pageTitle }}
            </h1>
            <p class="mt-4 max-w-3xl text-base leading-8 text-white/60">
              {{ pageDescription }}
            </p>
          </div>

          <div class="grid grid-cols-2 gap-3 rounded-2xl border border-white/10 bg-white/[0.035] p-4">
            <div>
              <p class="text-xs text-white/40">{{ copy.total }}</p>
              <p class="mt-1 text-2xl font-black text-white">{{ summary.total }}</p>
            </div>
            <div>
              <p class="text-xs text-white/40">{{ copy.sources }}</p>
              <p class="mt-1 text-2xl font-black text-white">{{ summary.source_count }}</p>
            </div>
            <div>
              <p class="text-xs text-white/40">{{ copy.cases }}</p>
              <p class="mt-1 text-2xl font-black text-white">{{ summary.case_count }}</p>
            </div>
            <div>
              <p class="text-xs text-white/40">{{ copy.templates }}</p>
              <p class="mt-1 text-2xl font-black text-white">{{ summary.template_count }}</p>
            </div>
          </div>
        </section>

        <section class="mt-8 rounded-2xl border border-white/10 bg-[#17181d] p-4 sm:p-5">
          <div class="grid gap-3 lg:grid-cols-[minmax(220px,1fr)_180px_220px_220px_auto] lg:items-center">
            <label class="block">
              <span class="sr-only">{{ copy.search }}</span>
              <input
                v-model.trim="draftSearch"
                type="search"
                class="h-12 w-full rounded-xl border border-white/10 bg-white/[0.045] px-4 text-sm text-white outline-none transition placeholder:text-white/30 focus:border-cyan-300/40 focus:bg-white/[0.065]"
                :placeholder="copy.searchPlaceholder"
                @keyup.enter="applySearch"
              />
            </label>

            <select
              v-model="filters.sourceType"
              class="h-12 rounded-xl border border-white/10 bg-[#111318] px-3 text-sm text-white outline-none focus:border-cyan-300/40"
              @change="reloadFromFirstPage"
            >
              <option value="case">{{ copy.caseOnly }}</option>
              <option value="template">{{ copy.templateOnly }}</option>
              <option value="">{{ copy.allTypes }}</option>
            </select>

            <select
              v-model="filters.sourceProject"
              class="h-12 rounded-xl border border-white/10 bg-[#111318] px-3 text-sm text-white outline-none focus:border-cyan-300/40"
              @change="reloadFromFirstPage"
            >
              <option value="">{{ copy.allSources }}</option>
              <option v-for="source in sourceOptions" :key="source.value" :value="source.value">
                {{ facetLabel(source) }}
              </option>
            </select>

            <select
              v-model="filters.category"
              class="h-12 rounded-xl border border-white/10 bg-[#111318] px-3 text-sm text-white outline-none focus:border-cyan-300/40"
              @change="reloadFromFirstPage"
            >
              <option value="">{{ copy.allCategories }}</option>
              <option v-for="category in categoryOptions" :key="category.value" :value="category.value">
                {{ facetLabel(category) }}
              </option>
            </select>

            <button
              type="button"
              class="inline-flex h-12 items-center justify-center rounded-xl border border-white/10 px-4 text-sm font-semibold transition"
              :class="filters.hasImage ? 'bg-cyan-300 text-slate-950' : 'bg-white/[0.045] text-white/70 hover:bg-white/[0.07]'"
              @click="toggleImageFilter"
            >
              {{ copy.hasImage }}
            </button>
          </div>

          <div class="mt-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <p class="text-sm text-white/45">
              {{ copy.resultPrefix }} {{ total }} · {{ copy.page }} {{ page }} / {{ pages }}
            </p>
            <div class="flex gap-2">
              <button
                type="button"
                class="rounded-xl border border-white/10 px-4 py-2 text-sm font-semibold text-white/70 transition hover:bg-white/[0.06] disabled:cursor-not-allowed disabled:opacity-40"
                :disabled="page <= 1 || loading"
                @click="goToPage(page - 1)"
              >
                {{ copy.previous }}
              </button>
              <button
                type="button"
                class="rounded-xl border border-white/10 px-4 py-2 text-sm font-semibold text-white/70 transition hover:bg-white/[0.06] disabled:cursor-not-allowed disabled:opacity-40"
                :disabled="page >= pages || loading"
                @click="goToPage(page + 1)"
              >
                {{ copy.next }}
              </button>
            </div>
          </div>
        </section>

        <section
          v-if="isAdmin"
          class="mt-5 rounded-2xl border border-cyan-300/20 bg-cyan-300/[0.055] p-4 sm:p-5"
        >
          <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
            <div>
              <p class="text-sm font-bold text-cyan-100">{{ copy.importTitle }}</p>
              <p class="mt-1 text-sm leading-6 text-white/52">{{ copy.importDescription }}</p>
            </div>
            <form class="grid gap-3 lg:min-w-[620px] lg:grid-cols-[160px_minmax(260px,1fr)_auto]" @submit.prevent="importFromSource">
              <select
                v-model="importForm.provider"
                class="h-12 rounded-xl border border-white/10 bg-[#111318] px-3 text-sm text-white outline-none focus:border-cyan-300/40"
                :disabled="importing"
              >
                <option value="x">{{ copy.importProviderX }}</option>
              </select>
              <input
                v-model.trim="importForm.url"
                type="url"
                class="h-12 rounded-xl border border-white/10 bg-white/[0.045] px-4 text-sm text-white outline-none transition placeholder:text-white/30 focus:border-cyan-300/40 focus:bg-white/[0.065]"
                :placeholder="copy.importPlaceholder"
                :disabled="importing"
                required
              />
              <button
                type="submit"
                class="inline-flex h-12 items-center justify-center rounded-xl bg-cyan-300 px-5 text-sm font-black text-slate-950 transition hover:bg-cyan-200 disabled:cursor-not-allowed disabled:opacity-50"
                :disabled="importing || !importForm.url"
              >
                {{ importing ? copy.importing : copy.importAction }}
              </button>
            </form>
          </div>

          <div
            v-if="importMessage"
            class="mt-4 rounded-xl border px-4 py-3 text-sm"
            :class="importError ? 'border-red-400/25 bg-red-500/10 text-red-100' : 'border-cyan-300/20 bg-cyan-300/10 text-cyan-50'"
          >
            {{ importMessage }}
          </div>

          <div
            v-if="importWarnings.length > 0"
            class="mt-3 rounded-xl border border-amber-300/20 bg-amber-300/10 px-4 py-3 text-sm text-amber-50"
          >
            <p class="font-bold">{{ copy.importWarnings }}</p>
            <ul class="mt-2 list-disc space-y-1 pl-5">
              <li v-for="warning in importWarnings" :key="warning">{{ warning }}</li>
            </ul>
          </div>
        </section>

        <div v-if="loading" class="flex justify-center py-20">
          <div class="h-10 w-10 animate-spin rounded-full border-2 border-white/20 border-t-white"></div>
        </div>

        <div
          v-else-if="errorMessage"
          class="mt-8 rounded-2xl border border-red-400/20 bg-red-500/10 px-5 py-4 text-sm text-red-100"
        >
          {{ errorMessage }}
        </div>

        <div
          v-else-if="items.length === 0"
          class="mt-8 rounded-2xl border border-white/10 bg-white/[0.03] px-8 py-16 text-center"
        >
          <h2 class="text-2xl font-bold text-white">{{ copy.emptyTitle }}</h2>
          <p class="mx-auto mt-3 max-w-xl text-sm leading-7 text-white/55">{{ copy.emptyDescription }}</p>
        </div>

        <section v-else class="mt-8 grid gap-5 lg:grid-cols-2 xl:grid-cols-3">
          <article
            v-for="item in items"
            :key="item.id"
            class="flex min-h-[420px] flex-col overflow-hidden rounded-2xl border border-white/10 bg-[#17181d] shadow-[0_20px_60px_rgba(0,0,0,0.24)]"
          >
            <div class="relative aspect-[4/3] bg-[#0f1117]">
              <img
                v-if="item.primary_image_url"
                :src="item.primary_image_url"
                :alt="item.title"
                loading="lazy"
                class="h-full w-full object-contain"
              />
              <div v-else class="flex h-full items-center justify-center px-6 text-center text-sm text-white/35">
                {{ copy.noImage }}
              </div>
            </div>

            <div class="flex flex-1 flex-col p-5">
              <div class="flex flex-wrap gap-2">
                <span
                  v-if="item.source_display_label"
                  class="rounded-full border border-white/10 bg-white/[0.04] px-3 py-1 text-xs font-semibold text-white/60"
                >
                  {{ item.source_display_label }}
                </span>
                <span
                  v-if="item.category"
                  class="rounded-full border border-cyan-300/20 bg-cyan-300/10 px-3 py-1 text-xs font-semibold text-cyan-100"
                >
                  {{ item.category }}
                </span>
              </div>

              <h2 class="mt-4 line-clamp-2 text-xl font-black leading-tight text-white">
                {{ item.title }}
              </h2>
              <p class="mt-3 line-clamp-4 text-sm leading-7 text-white/58">
                {{ item.prompt_preview || item.prompt }}
              </p>

              <div class="mt-4 flex flex-wrap gap-2">
                <span
                  v-for="tag in item.visible_tags"
                  :key="tag"
                  class="rounded-full bg-white/[0.045] px-2.5 py-1 text-xs text-white/50"
                >
                  {{ tag }}
                </span>
              </div>

              <div class="mt-auto flex items-center justify-between gap-3 pt-5">
                <p class="text-xs text-white/35">{{ formattedDate(item.imported_at || item.created_at) }}</p>
                <div class="flex shrink-0 flex-wrap justify-end gap-2">
                  <button
                    type="button"
                    class="rounded-xl border border-white/10 px-3 py-2 text-xs font-semibold text-white/70 transition hover:bg-white/[0.06]"
                    @click="openDetails(item)"
                  >
                    {{ copy.details }}
                  </button>
                  <button
                    type="button"
                    class="rounded-xl border border-violet-300/20 bg-violet-300/10 px-3 py-2 text-xs font-semibold text-violet-100 transition hover:bg-violet-300/20"
                    @click="openGenerator(item)"
                  >
                    {{ copy.generate }}
                  </button>
                  <button
                    type="button"
                    class="rounded-xl border border-cyan-300/20 bg-cyan-300/10 px-3 py-2 text-xs font-semibold text-cyan-100 transition hover:bg-cyan-300/20"
                    @click="copyPrompt(item)"
                  >
                    {{ copy.copyPrompt }}
                  </button>
                </div>
              </div>
            </div>
          </article>
        </section>
      </div>
    </main>

    <BaseDialog
      :show="Boolean(selectedPrompt)"
      :title="selectedPrompt?.title || copy.details"
      width="full"
      close-on-click-outside
      :z-index="80"
      @close="closeDetails"
    >
      <div v-if="selectedPrompt" class="grid max-h-[72vh] gap-6 overflow-y-auto pr-1 text-slate-900 dark:text-white lg:grid-cols-[minmax(0,1fr)_420px]">
        <section class="min-w-0">
          <div class="overflow-hidden rounded-2xl border border-slate-200 bg-slate-950 dark:border-white/10">
            <img
              v-if="selectedPrompt.primary_image_url"
              :src="selectedPrompt.primary_image_url"
              :alt="selectedPrompt.title"
              class="max-h-[62vh] w-full object-contain"
            />
            <div v-else class="flex min-h-[320px] items-center justify-center px-6 text-sm text-white/45">
              {{ copy.noImage }}
            </div>
          </div>

          <div v-if="selectedPrompt.image_urls.length > 1" class="mt-3 flex gap-2 overflow-x-auto pb-1">
            <img
              v-for="image in selectedPrompt.image_urls"
              :key="image"
              :src="image"
              :alt="selectedPrompt.title"
              class="h-20 w-20 shrink-0 rounded-xl border border-slate-200 object-cover dark:border-white/10"
            />
          </div>
        </section>

        <section class="min-w-0 space-y-5">
          <div class="flex flex-wrap gap-2">
            <span
              v-if="selectedPrompt.source_display_label"
              class="rounded-full border border-slate-200 bg-slate-50 px-3 py-1 text-xs font-semibold text-slate-600 dark:border-white/10 dark:bg-white/[0.04] dark:text-white/60"
            >
              {{ selectedPrompt.source_display_label }}
            </span>
            <span
              v-if="selectedPrompt.category"
              class="rounded-full border border-cyan-300/30 bg-cyan-300/10 px-3 py-1 text-xs font-semibold text-cyan-700 dark:text-cyan-100"
            >
              {{ selectedPrompt.category }}
            </span>
          </div>

          <div class="rounded-2xl border border-slate-200 bg-slate-50 dark:border-white/10 dark:bg-white/[0.035]">
            <div class="flex items-center justify-between border-b border-slate-200 px-4 py-3 dark:border-white/10">
              <h3 class="text-sm font-bold text-slate-700 dark:text-white/75">{{ copy.prompt }}</h3>
              <span class="text-xs text-slate-400">{{ selectedPrompt.prompt_char_count }} {{ copy.charUnit }}</span>
            </div>
            <pre class="whitespace-pre-wrap break-words p-4 text-sm leading-7 text-slate-700 dark:text-white/70">{{ selectedPrompt.prompt }}</pre>
          </div>

          <div class="flex flex-wrap gap-2">
            <span
              v-for="tag in selectedPrompt.all_tags"
              :key="tag"
              class="rounded-full bg-slate-100 px-2.5 py-1 text-xs text-slate-500 dark:bg-white/[0.045] dark:text-white/50"
            >
              {{ tag }}
            </span>
          </div>

          <div class="flex flex-wrap gap-3">
            <button
              type="button"
              class="rounded-xl bg-cyan-300 px-4 py-2 text-sm font-bold text-slate-950 transition hover:bg-cyan-200"
              @click="copyPrompt(selectedPrompt)"
            >
              {{ copy.copyPrompt }}
            </button>
            <button
              type="button"
              class="rounded-xl bg-violet-500 px-4 py-2 text-sm font-bold text-white transition hover:bg-violet-400"
              @click="openGenerator(selectedPrompt)"
            >
              {{ copy.generate }}
            </button>
            <a
              v-if="selectedPrompt.source_url"
              :href="selectedPrompt.source_url"
              target="_blank"
              rel="noreferrer"
              class="rounded-xl border border-slate-200 px-4 py-2 text-sm font-bold text-slate-600 transition hover:bg-slate-50 dark:border-white/10 dark:text-white/70 dark:hover:bg-white/[0.06]"
            >
              {{ copy.source }}
            </a>
          </div>
        </section>
      </div>
    </BaseDialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { getLocale } from '@/i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import { useClipboard } from '@/composables/useClipboard'
import { useAuthRouteDefaults } from '@/composables/useAuthRouteDefaults'
import { promptsAPI, type PromptCatalogFacet, type PromptCatalogItem, type PromptCatalogSummary } from '@/api/prompts'
import { useAppStore, useAuthStore } from '@/stores'
import { saveImageGeneratorDraft } from '@/utils/imageGeneratorDraft'
import { promptCatalogCopyKeys, resolvePromptCatalogShellConfig, type PromptCatalogCopy } from '@/utils/promptCatalogShell'
import { resolveRuntimeLanguage } from '@/utils/runtimeLocale'
import {
  applyPromptCatalogDefaults,
  buildPromptCatalogImportSuccessMessage,
  buildPromptCatalogListParams,
  formatPromptCatalogDate,
  resolvePromptCatalogFacetLabel,
  resolvePromptCatalogGeneratorDraftSource,
  resolvePromptCatalogGeneratorPath,
  resolvePromptCatalogImportXAuto,
  resolvePromptCatalogPageDescription,
  resolvePromptCatalogPageTitle,
} from './promptCatalogRuntime'

const authStore = useAuthStore()
const appStore = useAppStore()
const { authRouteDefaults, resolveHomePath } = useAuthRouteDefaults()
const { copyToClipboard } = useClipboard()

const loading = ref(false)
const importing = ref(false)
const errorMessage = ref('')
const importMessage = ref('')
const importError = ref(false)
const importWarnings = ref<string[]>([])
const items = ref<PromptCatalogItem[]>([])
const emptySummary = (): PromptCatalogSummary => ({
  total: 0,
  case_count: 0,
  template_count: 0,
  source_count: 0,
  category_count: 0,
  sources: [],
  categories: [],
  template_groups: [],
})
const summary = ref<PromptCatalogSummary>(emptySummary())
const total = ref(0)
const page = ref(1)
const pages = ref(1)
const draftSearch = ref('')
const selectedPrompt = ref<PromptCatalogItem | null>(null)

const filters = reactive({
  search: '',
  sourceType: '',
  sourceProject: '',
  category: '',
  hasImage: false,
})

const importForm = reactive({
  provider: 'x',
  url: '',
})

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName)
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => resolveHomePath(isAdmin.value))
const loginPath = computed(() => authRouteDefaults.value.loginPath)
const locale = computed(() => getLocale())
const promptCatalogLocale = computed<'zh' | 'en'>(() => resolveRuntimeLanguage(locale))

const shellConfig = computed(() =>
  resolvePromptCatalogShellConfig(appStore.cachedPublicSettings?.prompt_catalog_shell_config, promptCatalogLocale.value),
)
const copy = computed<PromptCatalogCopy>(() => {
  const labels = shellConfig.value.labels || {}
  const merged = Object.fromEntries(
    promptCatalogCopyKeys.map((key) => [key, labels[key] || '']),
  ) as PromptCatalogCopy
  return {
    ...merged,
    accountAction: isAuthenticated.value
      ? merged.accountActionAuthenticated
      : merged.accountActionAnonymous,
  }
})

const pageTitle = computed(() => {
  return resolvePromptCatalogPageTitle(copy.value, filters.sourceType)
})

const pageDescription = computed(() => {
  return resolvePromptCatalogPageDescription(copy.value, filters.sourceType)
})

const sourceOptions = computed(() => summary.value.sources)
const categoryOptions = computed(() => summary.value.categories)
const catalogDefaults = computed(() => shellConfig.value.defaults)

function applyConfiguredDefaults() {
  applyPromptCatalogDefaults(filters, catalogDefaults.value)
}

async function fetchCatalog(nextPage = 1) {
  loading.value = true
  errorMessage.value = ''

  try {
    const { data } = await promptsAPI.listCases(buildPromptCatalogListParams(filters, catalogDefaults.value, nextPage))

    items.value = data.items
    summary.value = data.summary
    total.value = data.total
    page.value = data.page
    pages.value = Math.max(data.pages, 1)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : copy.value.loadError
  } finally {
    loading.value = false
  }
}

function applySearch() {
  filters.search = draftSearch.value
  reloadFromFirstPage()
}

function reloadFromFirstPage() {
  void fetchCatalog(1)
}

function goToPage(nextPage: number) {
  void fetchCatalog(Math.min(Math.max(nextPage, 1), pages.value))
}

function toggleImageFilter() {
  filters.hasImage = !filters.hasImage
  reloadFromFirstPage()
}

function facetLabel(facet: PromptCatalogFacet): string {
  return resolvePromptCatalogFacetLabel(facet)
}

function formattedDate(value?: string | null): string {
  return formatPromptCatalogDate(value, locale.value)
}

function openDetails(item: PromptCatalogItem) {
  selectedPrompt.value = item
}

function closeDetails() {
  selectedPrompt.value = null
}

function copyPrompt(item: PromptCatalogItem) {
  void copyToClipboard(item.prompt, copy.value.promptCopied)
}

function openGenerator(item: PromptCatalogItem) {
  const path = resolvePromptCatalogGeneratorPath(catalogDefaults.value)
  if (!path) return

  try {
    saveImageGeneratorDraft({
      prompt: item.prompt,
      title: item.title,
      sourcePromptId: item.id,
      source: resolvePromptCatalogGeneratorDraftSource(catalogDefaults.value),
    })
  } catch {
    // Navigation still works if browser storage is unavailable.
  }

  window.location.assign(path)
}

async function importFromSource() {
  const url = importForm.url.trim()
  importMessage.value = ''
  importError.value = false
  importWarnings.value = []

  importing.value = true
  try {
    const { data } = await promptsAPI.importTwitter({
      url,
      x_auto: resolvePromptCatalogImportXAuto(catalogDefaults.value),
    })
    importForm.url = ''
    importWarnings.value = data.warnings || []
    importMessage.value = buildPromptCatalogImportSuccessMessage(copy.value.importSuccess, data.item.title)
    selectedPrompt.value = data.item
    await fetchCatalog(1)
  } catch (error) {
    importError.value = true
    importMessage.value = error instanceof Error ? error.message : copy.value.loadError
  } finally {
    importing.value = false
  }
}

onMounted(async () => {
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) {
    await appStore.fetchPublicSettings()
  }
  applyConfiguredDefaults()
  await fetchCatalog(1)
})
</script>
