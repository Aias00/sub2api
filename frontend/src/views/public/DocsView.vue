<template>
  <div class="min-h-screen bg-[radial-gradient(circle_at_top,_rgba(59,130,246,0.08),_transparent_30%),linear-gradient(180deg,#f8fafc_0%,#eef2ff_48%,#ffffff_100%)] text-gray-900 dark:bg-[radial-gradient(circle_at_top,_rgba(37,99,235,0.18),_transparent_24%),linear-gradient(180deg,#020817_0%,#081124_48%,#020617_100%)] dark:text-white">
    <header class="sticky top-0 z-40 border-b border-white/70 bg-white/85 backdrop-blur-xl dark:border-white/10 dark:bg-slate-950/75">
      <div class="mx-auto flex max-w-[1600px] items-center justify-between gap-4 px-4 py-3 md:px-6">
        <div class="flex min-w-0 items-center gap-3">
          <RouterLink
            to="/home"
            class="flex h-11 w-11 items-center justify-center overflow-hidden rounded-2xl border border-primary-100 bg-white shadow-sm shadow-primary-500/10 dark:border-white/10 dark:bg-white/5"
          >
            <img
              v-if="siteLogo"
              :src="siteLogo"
              :alt="siteName"
              class="h-full w-full object-cover"
            >
            <span v-else class="text-sm font-semibold tracking-[0.18em] text-primary-600 dark:text-primary-300">
              {{ siteName.slice(0, 2).toUpperCase() }}
            </span>
          </RouterLink>
          <div class="min-w-0">
            <p class="truncate text-sm font-semibold tracking-[0.24em] text-primary-600 dark:text-primary-300">
              {{ siteName }}
            </p>
            <h1 class="truncate text-lg font-semibold text-gray-950 dark:text-white">
              {{ t('nav.docs') }}
            </h1>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <button
            class="inline-flex items-center gap-2 rounded-xl border border-gray-200/80 bg-white px-3 py-2 text-sm font-medium text-gray-700 shadow-sm transition hover:border-gray-300 hover:bg-gray-50 lg:hidden dark:border-white/10 dark:bg-white/5 dark:text-white dark:hover:bg-white/10"
            @click="navOpen = !navOpen"
          >
            <Icon name="menu" size="sm" />
            {{ t('docs.navigation') }}
          </button>
          <router-link
            v-if="isAuthenticated"
            :to="dashboardPath"
            class="inline-flex items-center rounded-xl bg-gray-950 px-4 py-2 text-sm font-medium text-white transition hover:bg-gray-800 dark:bg-primary-500 dark:text-slate-950 dark:hover:bg-primary-400"
          >
            {{ t('home.goToDashboard') }}
          </router-link>
          <router-link
            v-else
            to="/login"
            class="inline-flex items-center rounded-xl bg-gray-950 px-4 py-2 text-sm font-medium text-white transition hover:bg-gray-800 dark:bg-primary-500 dark:text-slate-950 dark:hover:bg-primary-400"
          >
            {{ t('home.login') }}
          </router-link>
        </div>
      </div>
    </header>

    <div class="mx-auto max-w-[1600px] px-4 py-6 md:px-6">
      <div class="grid gap-6 xl:grid-cols-[280px,minmax(0,1fr),220px]">
        <aside
          :class="[
            'rounded-[28px] border border-gray-200/80 bg-white/92 p-4 shadow-[0_24px_60px_-30px_rgba(15,23,42,0.28)] dark:border-white/10 dark:bg-slate-900/70 xl:sticky xl:top-24 xl:h-[calc(100vh-8rem)] xl:overflow-hidden',
            navOpen ? 'block' : 'hidden xl:block',
          ]"
        >
          <div class="flex h-full flex-col">
            <label class="mb-3 block">
              <span class="sr-only">{{ t('docs.searchPlaceholder') }}</span>
              <input
                v-model.trim="searchQuery"
                type="text"
                :placeholder="t('docs.searchPlaceholder')"
                class="w-full rounded-2xl border border-gray-200 bg-gray-50 px-4 py-2.5 text-sm text-gray-700 outline-none transition focus:border-primary-300 focus:bg-white focus:ring-2 focus:ring-primary-200 dark:border-white/10 dark:bg-white/5 dark:text-white dark:placeholder:text-slate-400 dark:focus:border-primary-400 dark:focus:bg-white/10 dark:focus:ring-primary-500/30"
              >
            </label>

            <div class="flex-1 overflow-y-auto pr-1">
              <section
                v-for="section in filteredSections"
                :key="section.id"
                class="mb-5 last:mb-0"
              >
                <h2 class="mb-2 px-2 text-xs font-semibold uppercase tracking-[0.22em] text-gray-500 dark:text-slate-400">
                  {{ section.title }}
                </h2>
                <nav class="space-y-1">
                  <RouterLink
                    v-for="page in section.pages"
                    :key="page.slug"
                    :to="`/docs/${page.slug}`"
                    class="block rounded-2xl px-3 py-2.5 text-sm transition"
                    :class="
                      page.slug === currentPage?.slug
                        ? 'bg-primary-50 text-primary-700 shadow-sm dark:bg-primary-500/15 dark:text-primary-200'
                        : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900 dark:text-slate-300 dark:hover:bg-white/6 dark:hover:text-white'
                    "
                    @click="navOpen = false"
                  >
                    <div class="font-medium">{{ page.title }}</div>
                    <p class="mt-1 line-clamp-2 text-xs text-inherit/75">
                      {{ page.description }}
                    </p>
                  </RouterLink>
                </nav>
              </section>
            </div>
          </div>
        </aside>

        <main class="min-w-0">
          <div class="rounded-[32px] border border-white/75 bg-white/92 px-5 py-6 shadow-[0_30px_80px_-40px_rgba(15,23,42,0.35)] dark:border-white/10 dark:bg-slate-900/70 md:px-8 md:py-8">
            <div v-if="currentPage" class="space-y-6">
              <div class="space-y-3">
                <p class="text-xs font-semibold uppercase tracking-[0.22em] text-primary-600 dark:text-primary-300">
                  {{ currentPage.section }}
                </p>
                <div class="flex flex-wrap items-start justify-between gap-3">
                  <div class="min-w-0">
                    <h2 class="text-3xl font-semibold tracking-tight text-gray-950 dark:text-white md:text-4xl">
                      {{ currentPage.title }}
                    </h2>
                    <p class="mt-3 max-w-3xl text-sm leading-7 text-gray-600 dark:text-slate-300">
                      {{ currentPage.description }}
                    </p>
                  </div>
                </div>
              </div>

              <div ref="articleRef" class="docs-markdown" v-html="renderedHtml"></div>

              <div class="grid gap-3 border-t border-gray-200/80 pt-6 dark:border-white/10 md:grid-cols-2">
                <RouterLink
                  v-if="adjacent.previous"
                  :to="`/docs/${adjacent.previous.slug}`"
                  class="rounded-3xl border border-gray-200/80 bg-gray-50/70 px-5 py-4 transition hover:border-primary-200 hover:bg-primary-50/70 dark:border-white/10 dark:bg-white/5 dark:hover:border-primary-400/40 dark:hover:bg-primary-500/10"
                >
                  <div class="text-xs font-semibold uppercase tracking-[0.2em] text-gray-500 dark:text-slate-400">
                    {{ t('docs.previousPage') }}
                  </div>
                  <div class="mt-2 text-base font-semibold text-gray-900 dark:text-white">
                    {{ adjacent.previous.title }}
                  </div>
                </RouterLink>
                <div v-else class="hidden md:block"></div>

                <RouterLink
                  v-if="adjacent.next"
                  :to="`/docs/${adjacent.next.slug}`"
                  class="rounded-3xl border border-gray-200/80 bg-gray-50/70 px-5 py-4 text-left transition hover:border-primary-200 hover:bg-primary-50/70 dark:border-white/10 dark:bg-white/5 dark:hover:border-primary-400/40 dark:hover:bg-primary-500/10"
                  :class="{ 'md:ml-auto': !adjacent.previous }"
                >
                  <div class="text-xs font-semibold uppercase tracking-[0.2em] text-gray-500 dark:text-slate-400">
                    {{ t('docs.nextPage') }}
                  </div>
                  <div class="mt-2 text-base font-semibold text-gray-900 dark:text-white">
                    {{ adjacent.next.title }}
                  </div>
                </RouterLink>
              </div>
            </div>

            <div v-else class="rounded-3xl border border-dashed border-gray-200 px-6 py-12 text-center dark:border-white/10">
              <p class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t('customPage.notFoundTitle') }}
              </p>
              <p class="mt-2 text-sm text-gray-500 dark:text-slate-400">
                {{ t('customPage.notFoundDesc') }}
              </p>
            </div>
          </div>
        </main>

        <aside class="hidden xl:block">
          <div
            v-if="tocItems.length"
            class="sticky top-24 rounded-[28px] border border-gray-200/80 bg-white/92 p-4 shadow-[0_24px_60px_-30px_rgba(15,23,42,0.28)] dark:border-white/10 dark:bg-slate-900/70"
          >
            <h2 class="mb-3 text-xs font-semibold uppercase tracking-[0.22em] text-gray-500 dark:text-slate-400">
              {{ t('docs.onThisPage') }}
            </h2>
            <nav class="space-y-1.5">
              <a
                v-for="item in tocItems"
                :key="item.id"
                :href="`#${item.id}`"
                class="block rounded-xl px-3 py-2 text-sm transition"
                :class="[
                  item.id === activeHeadingId
                    ? 'bg-primary-50 text-primary-700 dark:bg-primary-500/15 dark:text-primary-200'
                    : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900 dark:text-slate-300 dark:hover:bg-white/6 dark:hover:text-white',
                  item.level >= 3 ? 'ml-3 text-[13px]' : '',
                ]"
                @click.prevent="scrollToHeading(item.id)"
              >
                {{ item.text }}
              </a>
            </nav>
          </div>
        </aside>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'
import Icon from '@/components/icons/Icon.vue'
import { defaultDocsSlug, docsSections, findDocsPage, getAdjacentDocsPages } from '@/utils/docs'

interface TocItem {
  id: string
  text: string
  level: number
}

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()

const navOpen = ref(false)
const searchQuery = ref('')
const renderedHtml = ref('')
const tocItems = ref<TocItem[]>([])
const articleRef = ref<HTMLElement | null>(null)
const activeHeadingId = ref('')
let scrollRafId = 0

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const isAuthenticated = computed(() => authStore.isAuthenticated)
const dashboardPath = computed(() => (authStore.isAdmin ? '/admin/dashboard' : '/dashboard'))

const currentSlug = computed(() => {
  const raw = route.params.pathMatch
  if (Array.isArray(raw)) {
    return raw.join('/')
  }
  return typeof raw === 'string' && raw ? raw : defaultDocsSlug
})

const currentPage = computed(() => findDocsPage(currentSlug.value) ?? findDocsPage(defaultDocsSlug))

const filteredSections = computed(() => {
  const keyword = searchQuery.value.trim().toLowerCase()
  if (!keyword) {
    return docsSections
  }

  return docsSections
    .map((section) => ({
      ...section,
      pages: section.pages.filter((page) => {
        const haystack = `${page.title} ${page.description} ${page.section}`.toLowerCase()
        return haystack.includes(keyword)
      }),
    }))
    .filter((section) => section.pages.length > 0)
})

const adjacent = computed(() =>
  currentPage.value ? getAdjacentDocsPages(currentPage.value.slug) : { previous: null, next: null }
)

function generateHeadingId(text: string, index: number): string {
  const base = text
    .toLowerCase()
    .replace(/[^\w一-龥]+/g, '-')
    .replace(/^-+|-+$/g, '')
  return base ? `${base}-${index}` : `heading-${index}`
}

function renderPageContent() {
  const page = currentPage.value
  if (!page) {
    renderedHtml.value = ''
    tocItems.value = []
    return
  }

  const html = marked.parse(page.content) as string
  const sanitized = DOMPurify.sanitize(html)
  const toc: TocItem[] = []
  let headingIndex = 0
  renderedHtml.value = sanitized.replace(
    /<(h[2-4])[^>]*>(.*?)<\/h[2-4]>/gi,
    (_, tag: string, content: string) => {
      const level = Number(tag[1])
      const text = content.replace(/<[^>]+>/g, '').trim()
      const id = generateHeadingId(text, headingIndex++)
      toc.push({ id, text, level })
      return `<${tag} id="${id}">${content}</${tag}>`
    }
  )
  tocItems.value = toc
  activeHeadingId.value = toc[0]?.id || ''
}

function scrollToHeading(id: string) {
  const container = articleRef.value
  if (!container) return
  const target = container.querySelector(`#${CSS.escape(id)}`) as HTMLElement | null
  if (!target) return
  target.scrollIntoView({ behavior: 'smooth', block: 'start' })
  activeHeadingId.value = id
}

function onArticleScroll() {
  if (scrollRafId) return
  scrollRafId = requestAnimationFrame(() => {
    scrollRafId = 0
    const container = articleRef.value
    if (!container || tocItems.value.length === 0) return

    const containerRect = container.getBoundingClientRect()
    let current = activeHeadingId.value

    for (const item of tocItems.value) {
      const heading = container.querySelector(`#${CSS.escape(item.id)}`) as HTMLElement | null
      if (!heading) continue
      const headingRect = heading.getBoundingClientRect()
      if (headingRect.top - containerRect.top <= 120) {
        current = item.id
      }
    }

    activeHeadingId.value = current
  })
}

watch(
  currentSlug,
  async () => {
    if (!findDocsPage(currentSlug.value) && defaultDocsSlug) {
      await router.replace(`/docs/${defaultDocsSlug}`)
      return
    }
    renderPageContent()
    navOpen.value = false
    await nextTick()
    window.scrollTo({ top: 0, behavior: 'auto' })
  },
  { immediate: true }
)

onMounted(async () => {
  if (!appStore.publicSettingsLoaded) {
    await appStore.fetchPublicSettings()
  }
  window.addEventListener('scroll', onArticleScroll, { passive: true })
})

onUnmounted(() => {
  window.removeEventListener('scroll', onArticleScroll)
  if (scrollRafId) {
    cancelAnimationFrame(scrollRafId)
  }
})
</script>

<style scoped>
.docs-markdown {
  line-height: 1.8;
  color: inherit;
}

.docs-markdown :deep(h2) {
  @apply mt-8 border-b border-gray-200 pb-3 text-2xl font-semibold tracking-tight text-gray-950 dark:border-white/10 dark:text-white;
}

.docs-markdown :deep(h3) {
  @apply mt-6 text-xl font-semibold text-gray-900 dark:text-white;
}

.docs-markdown :deep(h4) {
  @apply mt-4 text-lg font-semibold text-gray-900 dark:text-white;
}

.docs-markdown :deep(p) {
  @apply mb-4 text-[15px] leading-8 text-gray-700 dark:text-slate-300;
}

.docs-markdown :deep(ul) {
  @apply mb-4 list-disc space-y-2 pl-6 text-[15px] text-gray-700 dark:text-slate-300;
}

.docs-markdown :deep(ol) {
  @apply mb-4 list-decimal space-y-2 pl-6 text-[15px] text-gray-700 dark:text-slate-300;
}

.docs-markdown :deep(a) {
  @apply text-primary-600 underline decoration-primary-200 underline-offset-4 transition hover:text-primary-700 dark:text-primary-300 dark:decoration-primary-500/40 dark:hover:text-primary-200;
}

.docs-markdown :deep(blockquote) {
  @apply my-5 rounded-r-2xl border-l-4 border-primary-300 bg-primary-50/70 px-5 py-4 italic text-gray-700 dark:border-primary-400/40 dark:bg-primary-500/10 dark:text-slate-200;
}

.docs-markdown :deep(code) {
  @apply rounded-md bg-gray-100 px-1.5 py-0.5 font-mono text-sm text-gray-800 dark:bg-white/10 dark:text-slate-100;
}

.docs-markdown :deep(pre) {
  @apply my-5 overflow-x-auto rounded-3xl bg-slate-950 p-5 text-sm text-slate-100 shadow-inner shadow-black/20;
}

.docs-markdown :deep(pre code) {
  @apply bg-transparent p-0 text-inherit;
}

.docs-markdown :deep(table) {
  @apply my-6 w-full border-collapse overflow-hidden rounded-2xl border border-gray-200 dark:border-white/10;
}

.docs-markdown :deep(th) {
  @apply bg-gray-50 px-4 py-3 text-left text-sm font-semibold text-gray-900 dark:bg-white/5 dark:text-white;
}

.docs-markdown :deep(td) {
  @apply border-t border-gray-200 px-4 py-3 text-sm text-gray-700 dark:border-white/10 dark:text-slate-300;
}

.docs-markdown :deep(hr) {
  @apply my-8 border-gray-200 dark:border-white/10;
}
</style>
