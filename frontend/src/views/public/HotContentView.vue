<template>
  <div class="home-business-page public-template-page" :class="props.appShell ? 'min-h-0' : 'min-h-screen'">
    <PublicDarkHeader v-if="!props.appShell" :account-label="t('hotContent.goConsole')" container-class="max-w-6xl" />

    <main :class="props.appShell ? 'py-0' : 'public-template-main'">
      <section class="public-template-container">
        <div class="max-w-4xl">
          <div>
            <p class="text-sm font-black uppercase tracking-[0.24em] text-[var(--public-muted)]">{{ t('hotContent.signalDesk') }}</p>
            <h1 class="mt-4 max-w-3xl text-5xl font-black leading-tight sm:text-6xl">
              {{ t('hotContent.title') }}
            </h1>
            <p class="mt-5 max-w-2xl text-base leading-8 text-[var(--public-body)]">
              {{ t('hotContent.subtitle') }}
            </p>
          </div>
        </div>

        <!-- Hot stream -->
        <div class="mt-10">
          <div class="min-w-0">
            <div class="flex flex-col gap-3 sm:flex-row">
              <input
                v-model="query"
                class="min-w-0 flex-1 rounded-2xl border public-template-input px-4 py-3 text-sm text-[var(--public-ink)] outline-none placeholder:text-[var(--public-faint)] focus:border-cyan-300/45"
                :placeholder="searchPlaceholder"
                @keyup.enter="searchAndRefresh"
              />
              <button type="button" :disabled="loading" class="public-template-button-primary rounded-xl px-6 py-3 text-sm font-black disabled:opacity-50" @click="searchAndRefresh">
                {{ t('hotContent.search') }}
              </button>
            </div>

            <div v-if="errorMessage" class="mt-6 public-template-error p-4 text-sm">
              {{ errorMessage }}
            </div>

            <div v-if="loading" class="mt-6 flex items-center justify-center gap-3 py-12 text-sm text-[var(--public-muted)]">
              <svg class="h-5 w-5 animate-spin" viewBox="0 0 24 24" fill="none"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z" /></svg>
              {{ t('hotContent.loading') }}
            </div>

            <section v-if="!loading" class="mt-6 grid gap-4">
              <article
                v-for="item in items"
                :key="item.id"
                class="rounded-[1.75rem] public-template-panel p-5 shadow-sm transition hover:-translate-y-0.5 hover:shadow-xl hover:shadow-black/20"
              >
                <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                  <div class="min-w-0">
                    <div class="flex flex-wrap gap-2 text-xs font-bold text-[var(--public-accent-strong)]">
                      <span>{{ hotItemSourceLabel(item) }}</span>
                      <span v-if="item.badge">· {{ item.badge }}</span>
                      <span v-if="item.score">· {{ item.score }}</span>
                      <span v-if="item.published_at">· {{ formatDate(item.published_at) }}</span>
                    </div>
                    <h2 class="mt-2 text-2xl font-black leading-snug">
                      <a v-if="item.canonical_url" :href="item.canonical_url" target="_blank" rel="noreferrer" class="hover:text-[var(--public-accent-strong)]">
                        {{ item.title }}
                      </a>
                      <span v-else>{{ item.title }}</span>
                    </h2>
                    <p class="mt-3 text-sm leading-7 text-[var(--public-body)]">{{ item.summary || item.reason || item.body }}</p>
                  </div>
                  <span class="shrink-0 rounded-full bg-[var(--public-panel-soft)] px-3 py-1 text-xs font-black uppercase text-[var(--public-muted)]">
                    {{ item.content_type }}
                  </span>
                </div>
              </article>
              <EmptyState v-if="items.length === 0" :text="t('hotContent.emptyItems')" />
              <div v-if="itemTotalPages > 1" class="mt-6 flex flex-col items-center gap-3 sm:flex-row sm:justify-between">
                <p class="text-xs text-[var(--public-muted)]">
                  {{ t('hotContent.paginationInfo', { page: itemPage, totalPages: itemTotalPages, total: itemTotal }) }}
                </p>
                <nav class="flex items-center gap-1">
                  <button
                    type="button"
                    class="inline-flex h-8 items-center justify-center rounded-lg public-template-panel-muted px-2 text-sm text-[var(--public-body)] transition hover:bg-[var(--public-panel-soft)] hover:text-[var(--public-ink)] disabled:cursor-not-allowed disabled:opacity-30"
                    :disabled="itemPage <= 1 || loading"
                    @click="goToItemPage(itemPage - 1)"
                  >
                    ‹
                  </button>
                  <template v-for="(pageNum, idx) in getVisiblePages(itemPage, itemTotalPages)" :key="`${pageNum}-${idx}`">
                    <span v-if="typeof pageNum === 'string'" class="inline-flex h-8 w-8 items-center justify-center text-xs text-[var(--public-faint)]">...</span>
                    <button
                      v-else
                      type="button"
                      class="inline-flex h-8 w-8 items-center justify-center rounded-lg text-sm font-medium transition"
                      :class="pageNum === itemPage ? 'border border-cyan-200/40 bg-cyan-200/15 text-[var(--public-accent-strong)]' : 'public-template-panel-muted text-[var(--public-body)] hover:bg-[var(--public-panel-soft)] hover:text-[var(--public-ink)]'"
                      :disabled="loading"
                      @click="goToItemPage(pageNum)"
                    >
                      {{ pageNum }}
                    </button>
                  </template>
                  <button
                    type="button"
                    class="inline-flex h-8 items-center justify-center rounded-lg public-template-panel-muted px-2 text-sm text-[var(--public-body)] transition hover:bg-[var(--public-panel-soft)] hover:text-[var(--public-ink)] disabled:cursor-not-allowed disabled:opacity-30"
                    :disabled="itemPage >= itemTotalPages || loading"
                    @click="goToItemPage(itemPage + 1)"
                  >
                    ›
                  </button>
                </nav>
              </div>
            </section>
          </div>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import PublicDarkHeader from '@/components/layout/PublicDarkHeader.vue'
import { useAppStore } from '@/stores'
import {
  listHotItems,
  listHotSources,
  type HotItem,
  type HotSource,
} from '@/api/hot-content'

const props = withDefaults(defineProps<{
  appShell?: boolean
}>(), {
  appShell: false,
})

const EmptyState = defineComponent({
  props: {
    text: {
      type: String,
      required: true,
    },
    theme: {
      type: String as () => 'dark' | 'light',
      default: 'dark',
    },
  },
  setup(props) {
    return () => h('p', {
      class: props.theme === 'light'
        ? 'rounded-[1.75rem] border border-black/10 bg-black/[0.035] p-8 text-center text-sm text-black/45'
        : 'rounded-[1.75rem] public-template-panel-muted p-8 text-center text-sm text-[var(--public-muted)]',
    }, props.text)
  },
})

const { t } = useI18n()
const appStore = useAppStore()
const sources = ref<HotSource[]>([])
const items = ref<HotItem[]>([])
const itemTotal = ref(0)
const itemPage = ref(1)
const itemPageSize = 30
const query = ref('')
const loading = ref(false)
const errorMessage = ref('')

const searchPlaceholder = computed(() => t('hotContent.searchItems'))
const itemTotalPages = computed(() => Math.ceil(itemTotal.value / itemPageSize))

async function withLoading(task: () => Promise<void>) {
  loading.value = true
  errorMessage.value = ''
  try {
    await task()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : t('hotContent.loadFailed')
  } finally {
    loading.value = false
  }
}

async function loadItems() {
  const result = await listHotItems({
    page: itemPage.value,
    page_size: itemPageSize,
    q: query.value.trim(),
  })
  items.value = result.items
  itemTotal.value = result.total
}

async function refreshActiveTab() {
  await withLoading(loadItems)
}

async function searchAndRefresh() {
  itemPage.value = 1
  await refreshActiveTab()
}

async function loadSources() {
  try {
    sources.value = await listHotSources()
  } catch {
    sources.value = []
  }
}

function formatDate(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, '0')
  const d = String(date.getDate()).padStart(2, '0')
  const h = String(date.getHours()).padStart(2, '0')
  const min = String(date.getMinutes()).padStart(2, '0')
  return `${y}-${m}-${d} ${h}:${min}`
}

function hotItemSourceLabel(item: HotItem): string {
  const label = item.source_name || item.source_id
  return label.toLowerCase().includes('rss') ? t('hotContent.sourceAggregated') : label
}

function getVisiblePages(current: number, total: number): (number | string)[] {
  const pages: (number | string)[] = []
  const maxVisible = 5
  if (total <= maxVisible) {
    for (let i = 1; i <= total; i++) pages.push(i)
  } else {
    pages.push(1)
    const start = Math.max(2, current - 1)
    const end = Math.min(total - 1, current + 1)
    if (start > 2) pages.push('...')
    for (let i = start; i <= end; i++) pages.push(i)
    if (end < total - 1) pages.push('...')
    pages.push(total)
  }
  return pages
}

function goToItemPage(page: number) {
  if (page >= 1 && page <= itemTotalPages.value && page !== itemPage.value) {
    itemPage.value = page
    void withLoading(loadItems)
  }
}

onMounted(async () => {
  if (!appStore.publicSettingsLoaded) {
    await appStore.fetchPublicSettings()
  }
  await Promise.all([loadSources(), refreshActiveTab()])
})
</script>
