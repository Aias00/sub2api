<template>
  <div class="home-business-page public-dark-page min-h-screen bg-[#101114] text-white">
    <PublicDarkHeader :account-label="copy.accountAction" />

    <main class="px-6 py-10 sm:py-14">
      <div class="mx-auto max-w-7xl">
        <section>
          <div class="grid gap-5 lg:grid-cols-[minmax(0,1fr)_minmax(420px,620px)] lg:items-end">
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

            <div
              v-if="isAdmin"
              class="rounded-2xl border border-cyan-300/20 bg-cyan-300/[0.055] p-3 sm:p-4"
            >
              <div class="mb-3 flex items-center justify-between gap-3">
                <p class="shrink-0 text-sm font-bold text-cyan-100">{{ copy.importTitle }}</p>
                <p class="hidden truncate text-xs text-white/45 xl:block">{{ copy.importDescription }}</p>
              </div>
              <form class="grid gap-2 sm:grid-cols-[140px_minmax(220px,1fr)_auto]" @submit.prevent="importFromSource">
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
          </div>

          <div
            v-if="importMessage"
            class="mt-4 rounded-2xl border px-4 py-3 text-sm"
            :class="importError ? 'border-red-300/20 bg-red-300/10 text-red-100' : 'border-cyan-300/20 bg-cyan-300/10 text-cyan-50'"
          >
            {{ importMessage }}
          </div>

          <div
            v-if="importWarnings.length > 0"
            class="mt-3 rounded-2xl border border-amber-300/20 bg-amber-300/10 px-4 py-3 text-sm text-amber-50"
          >
            <p class="font-bold">{{ copy.importWarnings }}</p>
            <ul class="mt-2 list-disc space-y-1 pl-5">
              <li v-for="warning in importWarnings" :key="warning">{{ warning }}</li>
            </ul>
          </div>
        </section>

        <div class="mt-8 grid gap-6 lg:grid-cols-[300px_minmax(0,1fr)] lg:items-start">
          <aside class="rounded-2xl border border-white/10 bg-[#17181d] p-4 sm:p-5">
            <div class="space-y-4">
              <label class="block">
                <span class="mb-2 block text-xs font-semibold uppercase tracking-[0.16em] text-white/38">{{ copy.search }}</span>
                <input
                  v-model.trim="draftSearch"
                  type="search"
                  class="h-12 w-full rounded-xl border border-white/10 bg-white/[0.045] px-4 text-sm text-white outline-none transition placeholder:text-white/30 focus:border-cyan-300/40 focus:bg-white/[0.065]"
                  :placeholder="copy.searchPlaceholder"
                  @keyup.enter="applySearch"
                />
              </label>

              <div>
                <span class="mb-2 block text-xs font-semibold uppercase tracking-[0.16em] text-white/38">{{ copy.allCategories }}</span>
                <div class="flex flex-wrap items-start gap-2">
                  <button
                    type="button"
                    class="inline-flex min-h-10 max-w-full items-center justify-between gap-3 rounded-2xl border px-3 py-2 text-left text-sm font-semibold transition"
                    :class="!filters.category ? 'border-cyan-300/35 bg-cyan-300/10 text-cyan-50' : 'border-white/10 bg-white/[0.035] text-white/68 hover:bg-white/[0.06]'"
                    @click="setCategoryFilter('')"
                  >
                    <span>{{ copy.allCategories }}</span>
                    <span class="text-xs text-white/35">{{ summary.total }}</span>
                  </button>
                  <button
                    v-for="(category, index) in categoryOptions"
                    :key="category.value"
                    type="button"
                    class="inline-flex min-h-10 max-w-full items-center justify-between gap-3 rounded-2xl border px-3 py-2 text-left text-sm font-semibold transition"
                    :class="[categoryChipClass(index), filters.category === category.value ? 'border-cyan-300/35 bg-cyan-300/10 text-cyan-50' : 'border-white/10 bg-white/[0.035] text-white/68 hover:bg-white/[0.06]']"
                    @click="setCategoryFilter(category.value)"
                  >
                    <span class="min-w-0 truncate">{{ facetLabel(category) }}</span>
                    <span class="shrink-0 text-xs text-white/35">{{ category.count }}</span>
                  </button>
                </div>
              </div>

              <button
                type="button"
                class="inline-flex h-12 w-full items-center justify-center rounded-xl bg-cyan-300 px-4 text-sm font-black text-slate-950 transition hover:bg-cyan-200"
                @click="applySearch"
              >
                {{ copy.search }}
              </button>
            </div>
          </aside>

          <div class="min-w-0">
            <div v-if="loading && items.length === 0" class="flex justify-center py-20">
              <div class="h-10 w-10 animate-spin rounded-full border-2 border-white/20 border-t-white"></div>
            </div>

            <div
              v-else-if="errorMessage"
              class="rounded-2xl border border-red-300/20 bg-red-300/10 px-5 py-4 text-sm text-red-100"
            >
              {{ errorMessage }}
            </div>

            <div
              v-else-if="items.length === 0"
              class="rounded-2xl border border-white/10 bg-white/[0.03] px-8 py-16 text-center"
            >
              <h2 class="text-2xl font-bold text-white">{{ copy.emptyTitle }}</h2>
              <p class="mx-auto mt-3 max-w-xl text-sm leading-7 text-white/55">{{ copy.emptyDescription }}</p>
            </div>

            <section
              v-else
              class="rounded-2xl border border-white/10"
            >
              <div class="grid gap-5 p-1 xl:grid-cols-2 2xl:grid-cols-3">
                <article
                  v-for="item in items"
                  :key="item.id"
                  class="flex min-h-[420px] flex-col overflow-hidden rounded-2xl border border-white/10 bg-[#17181d] shadow-[0_20px_60px_rgba(0,0,0,0.24)]"
                >
                  <button
                    type="button"
                    class="group/image relative aspect-[4/3] w-full overflow-hidden bg-[#0f1117] text-left transition focus:outline-none focus-visible:ring-2 focus-visible:ring-cyan-300/50"
                    @click="openDetails(item)"
                  >
                    <img
                      v-if="item.primary_image_url"
                      :src="item.primary_image_url"
                      :alt="item.title"
                      loading="lazy"
                      class="h-full w-full object-contain transition duration-300 group-hover/image:scale-[1.02]"
                    />
                    <div v-else class="flex h-full items-center justify-center px-6 text-center text-sm text-white/35">
                      {{ copy.noImage }}
                    </div>
                  </button>

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
                        {{ categoryLabel(item.category) }}
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
              </div>

              <div v-if="loadingMore" class="flex justify-center py-6">
                <div class="h-8 w-8 animate-spin rounded-full border-2 border-white/20 border-t-white"></div>
              </div>

              <div
                v-if="!loadingMore && page >= pages && items.length > 0"
                class="py-4 text-center text-xs text-white/30"
              >
                {{ copy.noMoreResults || '· · ·' }}
              </div>
            </section>
          </div>
        </div>
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
              {{ categoryLabel(selectedPrompt.category) }}
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
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { getLocale } from '@/i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import PublicDarkHeader from '@/components/layout/PublicDarkHeader.vue'
import { useClipboard } from '@/composables/useClipboard'
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
  resolvePromptCatalogValueLabel,
  resolvePromptCatalogGeneratorDraftSource,
  resolvePromptCatalogGeneratorPath,
  resolvePromptCatalogImportXAuto,
  resolvePromptCatalogPageDescription,
  resolvePromptCatalogPageTitle,
} from './promptCatalogRuntime'

const authStore = useAuthStore()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const loading = ref(false)
const loadingMore = ref(false)
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
const allCategories = ref<PromptCatalogFacet[]>([])
const total = ref(0)
const page = ref(1)
const pages = ref(1)
const draftSearch = ref('')
const selectedPrompt = ref<PromptCatalogItem | null>(null)

const filters = reactive({
  search: '',
  category: '',
  hasImage: false,
})

const importForm = reactive({
  provider: 'x',
  url: '',
})

const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
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
  return resolvePromptCatalogPageTitle(copy.value, '')
})

const pageDescription = computed(() => {
  return resolvePromptCatalogPageDescription(copy.value, '')
})

const categoryOptions = computed(() => allCategories.value.length > 0 ? allCategories.value : summary.value.categories)
const catalogDefaults = computed(() => shellConfig.value.defaults)

function applyConfiguredDefaults() {
  applyPromptCatalogDefaults(filters, catalogDefaults.value)
}

async function fetchCatalog(nextPage = 1) {
  const isLoadMore = nextPage > 1
  if (isLoadMore) {
    loadingMore.value = true
  } else {
    loading.value = true
  }
  errorMessage.value = ''

  try {
    const { data } = await promptsAPI.listCases(buildPromptCatalogListParams(filters, catalogDefaults.value, nextPage))

    if (isLoadMore) {
      items.value = [...items.value, ...data.items]
    } else {
      items.value = data.items
    }
    summary.value = data.summary
    if (!filters.category) {
      allCategories.value = data.summary.categories
    }
    total.value = data.total
    page.value = data.page
    pages.value = Math.max(data.pages, 1)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : copy.value.loadError
  } finally {
    loading.value = false
    loadingMore.value = false
  }
}

function applySearch() {
  filters.search = draftSearch.value
  reloadFromFirstPage()
}

function reloadFromFirstPage() {
  // Sync any pending search input before reloading so the API call uses the
  // text the user currently sees in the search box, even if they haven't
  // pressed Enter yet.
  if (filters.search !== draftSearch.value) {
    filters.search = draftSearch.value
  }
  void fetchCatalog(1)
}

function setCategoryFilter(category: string) {
  filters.category = category
  reloadFromFirstPage()
}

function categoryChipClass(index: number): string {
  const variants = [
    'basis-[58%] -translate-y-0.5',
    'basis-[38%] translate-y-1',
    'basis-[48%]',
    'basis-[44%] -translate-y-1',
    'basis-[64%] translate-y-0.5',
    'basis-[34%]',
  ]
  return variants[index % variants.length]
}

function handlePageScroll() {
  if (loading.value || loadingMore.value || page.value >= pages.value) return
  const threshold = 200
  const scrollTop = window.scrollY || document.documentElement.scrollTop
  const viewportHeight = window.innerHeight || document.documentElement.clientHeight
  const scrollHeight = document.documentElement.scrollHeight
  if (scrollHeight - scrollTop - viewportHeight < threshold) {
    void fetchCatalog(page.value + 1)
  }
}

function facetLabel(facet: PromptCatalogFacet): string {
  return resolvePromptCatalogFacetLabel(facet, promptCatalogLocale.value)
}

function categoryLabel(value: string): string {
  return resolvePromptCatalogValueLabel(value, promptCatalogLocale.value)
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
  window.addEventListener('scroll', handlePageScroll, { passive: true })
})

onUnmounted(() => {
  window.removeEventListener('scroll', handlePageScroll)
})
</script>
