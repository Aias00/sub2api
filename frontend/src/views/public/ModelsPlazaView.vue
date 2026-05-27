<template>
  <div class="min-h-screen bg-[#101114] text-white">
    <div class="relative overflow-hidden border-b border-white/10">
      <div class="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_top_left,rgba(59,130,246,0.18),transparent_30%),radial-gradient(circle_at_top_right,rgba(16,185,129,0.16),transparent_26%),linear-gradient(180deg,rgba(255,255,255,0.02),transparent_45%)]"></div>

      <header class="relative z-10 px-6 py-5">
        <nav class="mx-auto flex max-w-6xl items-center justify-between">
          <RouterLink to="/home" class="flex items-center gap-3">
            <div class="h-9 w-9 overflow-hidden rounded-xl border border-white/10 bg-white/5">
              <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
            </div>
            <span class="text-sm font-semibold text-white">{{ siteName }}</span>
          </RouterLink>

          <div class="flex items-center gap-3">
            <LocaleSwitcher />
            <DocsLink
              v-if="docUrl"
              :doc-url="docUrl"
              class="hidden rounded-full border border-white/10 px-4 py-2 text-sm font-medium text-white/70 transition hover:border-white/20 hover:text-white sm:inline-flex"
            >
              {{ t('home.viewDocs') }}
            </DocsLink>
            <RouterLink
              :to="isAuthenticated ? dashboardPath : '/login'"
              class="inline-flex items-center rounded-full border border-white/10 bg-white px-4 py-2 text-sm font-semibold text-slate-950 transition hover:bg-white/90"
            >
              {{ isAuthenticated ? t('home.goToDashboard') : t('home.login') }}
            </RouterLink>
          </div>
        </nav>
      </header>

      <section class="relative z-10 px-6 pb-16 pt-10 sm:pb-20 sm:pt-14">
        <div class="mx-auto max-w-5xl text-center">
          <div class="inline-flex items-center rounded-full border border-white/10 bg-white/5 px-4 py-2 text-sm font-medium text-white/70 backdrop-blur">
            {{ localText('模型广场', 'Model Plaza') }}
          </div>
          <h1 class="mt-6 text-balance text-5xl font-black tracking-tight text-white sm:text-6xl">
            {{ localText('公开模型目录', 'Public Model Catalog') }}
          </h1>
          <p class="mx-auto mt-5 max-w-3xl text-balance text-base leading-8 text-white/60 sm:text-lg">
            {{
              localText(
                '从后台直接配置并公开展示可售模型卡片。适合做模型能力说明、价格展示和统一入口。',
                'Configure and publish model cards directly from the admin backend for capability overviews, pricing communication, and a unified entry point.',
              )
            }}
          </p>
        </div>
      </section>
    </div>

    <main class="px-6 py-12 sm:py-16">
      <div class="mx-auto max-w-7xl">
        <div v-if="loading" class="flex justify-center py-20">
          <div class="h-10 w-10 animate-spin rounded-full border-2 border-white/20 border-t-white"></div>
        </div>

        <div
          v-else-if="items.length === 0"
          class="rounded-[32px] border border-white/10 bg-white/[0.03] px-8 py-16 text-center"
        >
          <h2 class="text-2xl font-bold text-white">
            {{ localText('模型广场暂未配置', 'Model plaza is not configured yet') }}
          </h2>
          <p class="mx-auto mt-4 max-w-2xl text-sm leading-7 text-white/60 sm:text-base">
            {{
              localText(
                '管理员完成模型广场配置后，这里会展示公开模型卡片。',
                'Once the admin configures model plaza items, public model cards will appear here.',
              )
            }}
          </p>
        </div>

        <div v-else class="grid gap-8 lg:grid-cols-[260px_minmax(0,1fr)]">
          <aside class="lg:sticky lg:top-6 lg:self-start">
            <div class="rounded-[28px] border border-white/10 bg-[#17181d] p-5 shadow-[0_20px_50px_rgba(0,0,0,0.22)]">
              <div>
                <p class="text-xs font-semibold uppercase tracking-[0.24em] text-white/35">
                  {{ localText('快速查找', 'Quick find') }}
                </p>
                <label class="mt-3 block">
                  <span class="sr-only">{{ localText('搜索模型广场', 'Search model plaza') }}</span>
                  <input
                    v-model.trim="searchQuery"
                    type="search"
                    class="w-full rounded-2xl border border-white/10 bg-white/[0.04] px-4 py-3 text-sm text-white outline-none transition placeholder:text-white/30 focus:border-white/20 focus:bg-white/[0.06]"
                    :placeholder="localText('搜索模型、能力或标签', 'Search models, capabilities, or tags')"
                  />
                </label>
              </div>

              <div class="mt-6">
                <p class="text-xs font-semibold uppercase tracking-[0.24em] text-white/35">
                  {{ localText('平台分组', 'Groups') }}
                </p>
                <div class="mt-3 space-y-2">
                  <button
                    v-for="group in groupOptions"
                    :key="group.key"
                    type="button"
                    class="flex w-full items-center justify-between rounded-2xl px-4 py-3 text-left text-sm transition"
                    :class="
                      activeGroup === group.key
                        ? 'border border-white/15 bg-white text-slate-950 shadow-[0_12px_30px_rgba(255,255,255,0.08)]'
                        : 'border border-white/8 bg-white/[0.03] text-white/70 hover:border-white/15 hover:bg-white/[0.05] hover:text-white'
                    "
                    @click="activeGroup = group.key"
                  >
                    <span class="font-medium">{{ group.label }}</span>
                    <span
                      class="rounded-full px-2 py-1 text-xs"
                      :class="activeGroup === group.key ? 'bg-slate-950/10 text-slate-700' : 'bg-white/5 text-white/40'"
                    >
                      {{ group.count }}
                    </span>
                  </button>
                </div>
              </div>
            </div>
          </aside>

          <section class="min-w-0">
            <div class="mb-6 flex flex-col gap-3 rounded-[28px] border border-white/10 bg-white/[0.03] px-5 py-4 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <p class="text-sm font-semibold text-white/80">
                  {{ activeGroupLabel }}
                </p>
                <p class="mt-1 text-sm text-white/45">
                  {{
                    searchQuery
                      ? localText(`当前搜索：${searchQuery}`, `Current search: ${searchQuery}`)
                      : localText('按平台分组浏览公开模型卡片。', 'Browse public model cards by provider group.')
                  }}
                </p>
              </div>
              <div class="inline-flex items-center rounded-full border border-white/10 bg-white/[0.04] px-4 py-2 text-sm text-white/70">
                {{ localText('结果', 'Results') }} · {{ filteredItems.length }}
              </div>
            </div>

            <div
              v-if="filteredItems.length === 0"
              class="rounded-[32px] border border-white/10 bg-white/[0.03] px-8 py-16 text-center"
            >
              <h2 class="text-2xl font-bold text-white">
                {{ localText('没有匹配的模型卡片', 'No matching model cards') }}
              </h2>
              <p class="mx-auto mt-4 max-w-2xl text-sm leading-7 text-white/60 sm:text-base">
                {{
                  localText(
                    '试试切换分组，或者换一个更宽松的关键词搜索。',
                    'Try another group or broaden the search terms.',
                  )
                }}
              </p>
            </div>

            <div v-else class="grid gap-6 xl:grid-cols-2">
              <article
                v-for="item in filteredItems"
                :key="item.id"
                class="rounded-[32px] border border-white/10 bg-[#17181d] p-7 shadow-[0_20px_60px_rgba(0,0,0,0.28)]"
              >
                <div class="flex items-start justify-between gap-4">
                  <div class="flex items-start gap-4">
                    <div
                      class="flex h-14 w-14 items-center justify-center rounded-2xl text-xl font-black text-white"
                      :class="providerIconClass(item.provider)"
                    >
                      {{ providerInitial(item.provider) }}
                    </div>
                    <div>
                      <div class="flex flex-wrap items-center gap-2">
                        <h2 class="text-2xl font-black text-white sm:text-[2rem]">
                          {{ item.title }}
                        </h2>
                        <span
                          v-if="item.badge"
                          class="rounded-full border border-white/10 bg-white/5 px-3 py-1 text-xs font-semibold uppercase tracking-[0.18em] text-white/60"
                        >
                          {{ item.badge }}
                        </span>
                      </div>
                      <p
                        v-if="item.description"
                        class="mt-3 max-w-2xl text-sm leading-7 text-white/65 sm:text-base"
                      >
                        {{ item.description }}
                      </p>
                    </div>
                  </div>

                  <button
                    v-if="item.model_ids.length > 0"
                    type="button"
                    class="inline-flex h-10 w-10 items-center justify-center rounded-xl border border-white/10 text-white/60 transition hover:border-white/20 hover:text-white"
                    :title="localText('复制模型 ID', 'Copy model IDs')"
                    @click="copyToClipboard(item.model_ids.join('\n'), localText('模型 ID 已复制', 'Model IDs copied'))"
                  >
                    <Icon name="copy" size="sm" />
                  </button>
                </div>

                <div class="mt-8 space-y-2 text-sm text-white/78 sm:text-base">
                  <p v-if="item.input_price">{{ localText('输入价格', 'Input price') }} {{ item.input_price }}</p>
                  <p v-if="item.output_price">{{ localText('输出价格', 'Output price') }} {{ item.output_price }}</p>
                  <p v-if="item.cache_read_price">{{ localText('缓存读取价格', 'Cache read price') }} {{ item.cache_read_price }}</p>
                  <p v-if="item.cache_write_price">{{ localText('缓存创建价格', 'Cache write price') }} {{ item.cache_write_price }}</p>
                </div>

                <div class="mt-8 flex flex-wrap gap-2">
                  <span
                    v-for="tag in item.capability_tags"
                    :key="tag"
                    class="rounded-full border border-white/10 bg-white/5 px-3 py-1 text-xs font-medium text-white/72"
                  >
                    {{ tag }}
                  </span>
                </div>

                <div class="mt-8 flex items-center justify-between gap-3">
                  <span
                    v-if="item.billing_badge"
                    class="rounded-full bg-violet-500/12 px-4 py-2 text-sm font-semibold text-violet-200"
                  >
                    {{ item.billing_badge }}
                  </span>
                  <div v-else></div>
                  <p v-if="item.model_ids.length > 0" class="text-xs text-white/35">
                    {{ localText('已配置模型 ID', 'Model IDs configured') }} · {{ item.model_ids.length }}
                  </p>
                </div>
              </article>
            </div>
          </section>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import DocsLink from '@/components/common/DocsLink.vue'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAuthStore, useAppStore } from '@/stores'
import { useClipboard } from '@/composables/useClipboard'
import type { ModelPlazaItem } from '@/types'

const { locale, t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const loading = ref(false)
const searchQuery = ref('')
const activeGroup = ref('all')

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const isAuthenticated = computed(() => authStore.isAuthenticated)
const dashboardPath = computed(() => (authStore.isAdmin ? '/admin/dashboard' : '/dashboard'))

const items = computed(() =>
  (appStore.cachedPublicSettings?.model_plaza_items || [])
    .filter((item) => item.visible !== false)
    .slice()
    .sort((a, b) => (a.sort_order || 0) - (b.sort_order || 0)),
)

const groupOptions = computed(() => {
  const counts = new Map<string, number>()

  for (const item of items.value) {
    const key = providerGroupKey(item.provider)
    counts.set(key, (counts.get(key) || 0) + 1)
  }

  const groups = Array.from(counts.entries())
    .map(([key, count]) => ({
      key,
      count,
      label: providerGroupLabel(key),
      rank: providerGroupRank(key),
    }))
    .sort((a, b) => a.rank - b.rank || a.label.localeCompare(b.label))

  return [
    {
      key: 'all',
      label: localText('全部模型', 'All models'),
      count: items.value.length,
      rank: -1,
    },
    ...groups,
  ]
})

const activeGroupLabel = computed(() => {
  const match = groupOptions.value.find((group) => group.key === activeGroup.value)
  return match?.label || localText('全部模型', 'All models')
})

const filteredItems = computed(() =>
  items.value.filter((item) => {
    if (activeGroup.value !== 'all' && providerGroupKey(item.provider) !== activeGroup.value) {
      return false
    }
    return matchesSearch(item, searchQuery.value)
  }),
)

function localText(zh: string, en: string): string {
  return locale.value.startsWith('zh') ? zh : en
}

function providerGroupKey(provider: string): string {
  const normalized = provider.trim().toLowerCase()
  if (normalized.includes('anthropic') || normalized.includes('claude')) return 'claude'
  if (normalized.includes('openai') || normalized.includes('gpt')) return 'gpt'
  if (normalized.includes('gemini') || normalized.includes('google')) return 'gemini'
  return normalized || 'other'
}

function providerGroupLabel(groupKey: string): string {
  if (groupKey === 'claude') return 'Claude'
  if (groupKey === 'gpt') return 'GPT'
  if (groupKey === 'gemini') return 'Gemini'
  if (groupKey === 'other') return localText('其他', 'Other')
  return groupKey.toUpperCase()
}

function providerGroupRank(groupKey: string): number {
  if (groupKey === 'claude') return 0
  if (groupKey === 'gpt') return 1
  if (groupKey === 'gemini') return 2
  if (groupKey === 'other') return 99
  return 50
}

function matchesSearch(item: ModelPlazaItem, query: string): boolean {
  const normalized = query.trim().toLowerCase()
  if (!normalized) return true

  const haystack = [
    item.title,
    item.provider,
    item.badge,
    item.description,
    item.input_price,
    item.output_price,
    item.cache_read_price,
    item.cache_write_price,
    item.billing_badge,
    ...item.capability_tags,
    ...item.model_ids,
  ]
    .join('\n')
    .toLowerCase()

  return haystack.includes(normalized)
}

function providerInitial(provider: string): string {
  const normalized = provider.trim().toLowerCase()
  if (normalized.includes('anthropic') || normalized.includes('claude')) return 'C'
  if (normalized.includes('openai') || normalized.includes('gpt')) return 'G'
  if (normalized.includes('gemini') || normalized.includes('google')) return 'G'
  return normalized.slice(0, 1).toUpperCase() || 'M'
}

function providerIconClass(provider: string): string {
  const normalized = provider.trim().toLowerCase()
  if (normalized.includes('anthropic') || normalized.includes('claude')) {
    return 'bg-[linear-gradient(135deg,#ef8e67,#d2745c)]'
  }
  if (normalized.includes('openai') || normalized.includes('gpt')) {
    return 'bg-[linear-gradient(135deg,#48b774,#2f9360)]'
  }
  if (normalized.includes('gemini') || normalized.includes('google')) {
    return 'bg-[linear-gradient(135deg,#5b7cff,#4a5ce4)]'
  }
  return 'bg-[linear-gradient(135deg,#64748b,#475569)]'
}

onMounted(async () => {
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) {
    loading.value = true
    try {
      await appStore.fetchPublicSettings()
    } finally {
      loading.value = false
    }
  }
})
</script>
