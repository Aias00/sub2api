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
      <div class="mx-auto max-w-6xl">
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

        <div v-else class="grid gap-6 xl:grid-cols-2">
          <article
            v-for="item in items"
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

const { locale, t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const loading = ref(false)

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

function localText(zh: string, en: string): string {
  return locale.value.startsWith('zh') ? zh : en
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
