<template>
  <div v-if="homeContent" class="min-h-screen">
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <div v-else v-html="homeContent"></div>
  </div>

  <div
    v-else
    class="relative min-h-screen overflow-x-hidden bg-white text-slate-900 dark:bg-dark-950 dark:text-white"
  >
    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <div class="absolute inset-x-0 top-0 h-[36rem] bg-[radial-gradient(circle_at_20%_60%,rgba(59,130,246,0.24),transparent_32%),radial-gradient(circle_at_80%_20%,rgba(96,165,250,0.18),transparent_28%),linear-gradient(180deg,rgba(239,246,255,0.95),rgba(255,255,255,0.96))] dark:bg-[radial-gradient(circle_at_20%_60%,rgba(59,130,246,0.18),transparent_32%),radial-gradient(circle_at_80%_20%,rgba(96,165,250,0.12),transparent_28%),linear-gradient(180deg,rgba(15,23,42,0.92),rgba(2,6,23,1))]"></div>
      <div class="absolute left-[-8rem] top-48 h-64 w-64 rounded-full bg-blue-500/20 blur-3xl"></div>
      <div class="absolute right-[-6rem] top-16 h-72 w-72 rounded-full bg-sky-300/30 blur-3xl"></div>
      <div class="absolute inset-0 bg-[linear-gradient(rgba(148,163,184,0.08)_1px,transparent_1px),linear-gradient(90deg,rgba(148,163,184,0.08)_1px,transparent_1px)] bg-[size:72px_72px] opacity-40 dark:opacity-10"></div>
    </div>

    <header class="relative z-20 px-6 py-5">
      <nav class="mx-auto flex max-w-5xl items-center justify-between">
        <div class="flex items-center gap-3">
          <div v-if="siteLogo" class="h-9 w-9 overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900">
            <img :src="siteLogo" alt="Logo" class="h-full w-full object-contain" />
          </div>
          <div class="flex items-center gap-2">
            <span class="text-sm font-semibold text-slate-900 dark:text-white">{{ siteName }}</span>
          </div>
        </div>

        <div class="hidden items-center gap-8 text-xs font-medium tracking-[0.12em] text-slate-500 lg:flex">
          <template v-for="item in navItems" :key="item.label">
            <DocsLink
              v-if="item.doc"
              :doc-url="docUrl"
              class="transition-colors hover:text-slate-900 dark:hover:text-white"
            >
              {{ item.label }}
            </DocsLink>
            <router-link
              v-else-if="item.to"
              :to="item.to"
              class="transition-colors hover:text-slate-900 dark:hover:text-white"
            >
              {{ item.label }}
            </router-link>
            <a
              v-else
              :href="item.href"
              class="transition-colors hover:text-slate-900 dark:hover:text-white"
            >
              {{ item.label }}
            </a>
          </template>
        </div>

        <div class="flex items-center gap-3">
          <LocaleSwitcher />
          <DocsLink
            :doc-url="docUrl"
            class="hidden rounded-full border border-slate-200 px-4 py-2 text-sm font-medium text-slate-600 transition hover:border-slate-300 hover:text-slate-900 dark:border-dark-700 dark:text-dark-200 dark:hover:border-dark-500 dark:hover:text-white sm:inline-flex"
          >
            {{ copy.viewDocs }}
          </DocsLink>
          <router-link
            :to="isAuthenticated ? dashboardPath : loginPath"
            class="inline-flex items-center rounded-full border border-slate-900 bg-slate-900 px-4 py-2 text-sm font-semibold text-white transition hover:bg-slate-800 dark:border-white dark:bg-white dark:text-slate-900 dark:hover:bg-slate-100"
          >
            {{ isAuthenticated ? copy.dashboard : copy.login }}
          </router-link>
        </div>
      </nav>
    </header>

    <main class="relative z-10">
      <section id="top" class="px-6 pb-28 pt-12 sm:pb-32 sm:pt-16">
        <div class="mx-auto flex max-w-5xl flex-col items-center text-center">
          <div class="mb-6 inline-flex items-center rounded-full border border-slate-300/80 bg-white/70 px-4 py-2 text-sm font-medium text-slate-700 backdrop-blur dark:border-dark-600 dark:bg-dark-900/70 dark:text-dark-100">
            {{ copy.heroBadge }}
          </div>
          <h1 class="max-w-3xl text-balance text-5xl font-black tracking-tight text-slate-950 dark:text-white sm:text-6xl lg:text-7xl">
            {{ copy.heroTitle }}
          </h1>
          <p class="mt-6 max-w-3xl text-balance text-base leading-8 text-slate-600 dark:text-dark-200 sm:text-lg">
            {{ copy.heroDescription }}
          </p>
          <div class="mt-10 flex flex-col items-center gap-3 sm:flex-row">
            <router-link
              :to="isAuthenticated ? dashboardPath : loginPath"
              class="inline-flex items-center rounded-full bg-slate-950 px-7 py-3 text-sm font-semibold text-white shadow-lg shadow-slate-900/10 transition hover:bg-slate-800 dark:bg-white dark:text-slate-950 dark:hover:bg-slate-100"
            >
              {{ isAuthenticated ? copy.dashboard : copy.primaryCta }}
            </router-link>
            <router-link
              :to="homeLinks.modelsPath"
              class="inline-flex items-center rounded-full border border-slate-300 bg-white/70 px-7 py-3 text-sm font-semibold text-slate-700 backdrop-blur transition hover:border-slate-400 hover:text-slate-950 dark:border-dark-600 dark:bg-dark-900/60 dark:text-dark-100 dark:hover:border-dark-400 dark:hover:text-white"
            >
              {{ copy.secondaryCta }}
            </router-link>
          </div>
        </div>
      </section>

      <section id="models" class="px-6 py-12 sm:py-16">
        <div class="mx-auto max-w-5xl">
          <div class="text-center">
            <p class="text-sm font-medium uppercase tracking-[0.2em] text-sky-600 dark:text-sky-400">
              {{ copy.modelMatrixKicker }}
            </p>
            <h2 class="mt-3 text-3xl font-black tracking-tight text-slate-950 dark:text-white sm:text-4xl">
              {{ copy.modelMatrixTitle }}
            </h2>
            <p class="mx-auto mt-4 max-w-xl text-sm leading-7 text-slate-500 dark:text-dark-300 sm:text-base">
              {{ copy.modelMatrixDescription }}
            </p>
          </div>

          <div v-if="catalogLoading" class="mt-10 flex justify-center py-10">
            <div class="h-8 w-8 animate-spin rounded-full border-2 border-sky-500 border-t-transparent"></div>
          </div>

          <div
            v-if="isBusinessHome"
            data-home-model-grid
            class="mt-12 grid gap-5 lg:grid-cols-2"
          >
            <article
              v-for="card in businessCards"
              :key="card.key"
              class="rounded-[28px] border border-slate-200 bg-white p-7 shadow-[0_12px_50px_rgba(15,23,42,0.05)] transition hover:-translate-y-1 hover:shadow-[0_18px_65px_rgba(15,23,42,0.08)] dark:border-dark-700 dark:bg-dark-900 dark:shadow-none"
            >
              <div class="flex items-start justify-between gap-4">
                <div>
                  <p class="text-sm font-medium uppercase tracking-[0.18em] text-sky-600 dark:text-sky-400">
                    {{ card.badge }}
                  </p>
                  <h3 class="mt-3 text-2xl font-bold text-slate-950 dark:text-white">
                    {{ card.title }}
                  </h3>
                </div>
                <div class="flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-br from-sky-500 to-indigo-600 text-sm font-bold text-white">
                  {{ card.title[0] }}
                </div>
              </div>
              <p class="mt-4 max-w-xl text-sm leading-7 text-slate-500 dark:text-dark-300">
                {{ card.description }}
              </p>
              <div class="mt-6 flex flex-wrap gap-2">
                <span
                  v-for="capability in card.capabilityTags"
                  :key="capability"
                  class="rounded-full border border-slate-200 bg-slate-50 px-3 py-1 text-xs font-medium text-slate-600 dark:border-dark-700 dark:bg-dark-800 dark:text-dark-200"
                >
                  {{ capability }}
                </span>
              </div>
              <router-link
                v-if="card.path && card.pathLabel"
                :to="card.path"
                class="mt-6 inline-flex items-center rounded-full border border-slate-300 bg-white/70 px-4 py-2 text-sm font-semibold text-slate-700 backdrop-blur transition hover:border-slate-400 hover:text-slate-950 dark:border-dark-600 dark:bg-dark-900/60 dark:text-dark-100 dark:hover:border-dark-400 dark:hover:text-white"
              >
                {{ card.pathLabel }}
              </router-link>
            </article>
          </div>

          <div
            v-else
            data-home-model-grid
            class="mt-12 grid gap-5"
            :class="modelGridClass"
          >
            <article
              v-for="family in visibleModelFamilies"
              :key="family.key"
              class="rounded-[28px] border border-slate-200 bg-white p-7 shadow-[0_12px_50px_rgba(15,23,42,0.05)] transition hover:-translate-y-1 hover:shadow-[0_18px_65px_rgba(15,23,42,0.08)] dark:border-dark-700 dark:bg-dark-900 dark:shadow-none"
            >
              <div class="flex items-start justify-between gap-4">
                <div>
                  <p class="text-sm font-medium uppercase tracking-[0.18em] text-sky-600 dark:text-sky-400">
                    {{ familyBadge(family.key) }}
                  </p>
                  <h3 class="mt-3 text-2xl font-bold text-slate-950 dark:text-white">
                    {{ family.name }}
                  </h3>
                </div>
                <div
                  class="flex h-12 w-12 items-center justify-center rounded-2xl text-sm font-bold text-white"
                  :class="familyIconClass(family.key)"
                >
                  {{ family.name[0] }}
                </div>
              </div>
              <p class="mt-6 text-sm font-semibold text-slate-800 dark:text-dark-100">
                {{ familyTagline(family.key) }}
              </p>
              <p class="mt-2 max-w-xs text-sm leading-7 text-slate-500 dark:text-dark-300">
                {{ family.models.length > 0 ? familyDescription(family.key) : copy.modelMatrixEmptyCard }}
              </p>
              <div v-if="family.models.length > 0" class="mt-6 flex flex-wrap gap-2">
                <span
                  v-for="capability in familyCapabilities(family.key)"
                  :key="capability"
                  class="rounded-full border border-slate-200 bg-slate-50 px-3 py-1 text-xs font-medium text-slate-600 dark:border-dark-700 dark:bg-dark-800 dark:text-dark-200"
                >
                  {{ capability }}
                </span>
              </div>
              <div v-else class="mt-6">
                <span class="rounded-full border border-sky-200 bg-sky-50 px-3 py-1 text-xs font-medium text-sky-700 dark:border-sky-900/50 dark:bg-sky-950/30 dark:text-sky-300">
                  {{ copy.modelMatrixEmptyPill }}
                </span>
              </div>
            </article>
          </div>
        </div>
      </section>

      <section class="border-y border-sky-100/80 bg-sky-50/70 px-6 py-24 dark:border-dark-800 dark:bg-dark-950/50">
        <div id="experience" class="mx-auto max-w-5xl">
          <div class="text-center">
            <p class="text-sm font-medium uppercase tracking-[0.2em] text-sky-600 dark:text-sky-400">
              {{ copy.experienceKicker }}
            </p>
            <h2 class="mt-3 text-3xl font-black tracking-tight text-slate-950 dark:text-white sm:text-4xl">
              {{ copy.experienceTitle }}
            </h2>
            <p class="mx-auto mt-4 max-w-2xl text-sm leading-7 text-slate-500 dark:text-dark-300 sm:text-base">
              {{ copy.experienceDescription }}
            </p>
          </div>

          <div class="mt-12 grid gap-5 md:grid-cols-2 xl:grid-cols-4">
            <article
              v-for="feature in experienceCards"
              :key="feature.title"
              class="rounded-[24px] border border-white/90 bg-white/85 p-6 shadow-[0_10px_30px_rgba(15,23,42,0.04)] backdrop-blur dark:border-dark-800 dark:bg-dark-900/80 dark:shadow-none"
            >
              <div
                v-if="feature.icon"
                class="flex h-12 w-12 items-center justify-center rounded-2xl text-white"
                :class="feature.iconClass"
              >
                <Icon :name="feature.icon" size="lg" />
              </div>
              <h3 class="mt-5 text-lg font-bold text-slate-950 dark:text-white">{{ feature.title }}</h3>
              <p class="mt-3 text-sm leading-7 text-slate-500 dark:text-dark-300">{{ feature.description }}</p>
            </article>
          </div>

          <div class="mt-24 text-center">
            <p class="text-sm font-medium uppercase tracking-[0.2em] text-sky-600 dark:text-sky-400">
              {{ copy.whyChooseKicker }}
            </p>
            <h2 class="mt-3 text-3xl font-black tracking-tight text-slate-950 dark:text-white sm:text-4xl">
              {{ copy.whyChooseTitle }}
            </h2>
            <p class="mx-auto mt-4 max-w-2xl text-sm leading-7 text-slate-500 dark:text-dark-300 sm:text-base">
              {{ copy.whyChooseDescription }}
            </p>
          </div>

          <div class="mt-10 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
            <article
              v-for="point in whyChooseCards"
              :key="point.title"
              class="rounded-[24px] border border-white/90 bg-white/80 p-5 backdrop-blur dark:border-dark-800 dark:bg-dark-900/80"
            >
              <h3 class="text-lg font-semibold text-slate-950 dark:text-white">{{ point.title }}</h3>
              <p class="mt-3 text-sm leading-7 text-slate-500 dark:text-dark-300">{{ point.description }}</p>
            </article>
          </div>
        </div>
      </section>
    </main>

    <footer class="border-t border-slate-200/80 px-6 py-14 dark:border-dark-800">
      <div class="mx-auto max-w-5xl">
        <div class="grid gap-12 md:grid-cols-[1.5fr_repeat(3,minmax(0,1fr))]">
          <div>
            <div class="flex items-center gap-3">
              <div v-if="siteLogo" class="h-9 w-9 overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900">
                <img :src="siteLogo" alt="Logo" class="h-full w-full object-contain" />
              </div>
              <span class="text-sm font-semibold text-slate-950 dark:text-white">{{ siteName }}</span>
            </div>
            <p class="mt-4 max-w-xs text-sm leading-7 text-slate-500 dark:text-dark-300">
              {{ copy.footerDescription }}
            </p>
          </div>

          <div v-for="section in footerSections" :key="section.title">
            <h3 class="text-sm font-semibold text-slate-950 dark:text-white">{{ section.title }}</h3>
            <ul class="mt-4 space-y-3">
              <li v-for="item in section.items" :key="item.label">
                <a
                  v-if="item.href"
                  :href="item.href"
                  class="text-sm text-slate-500 transition-colors hover:text-slate-900 dark:text-dark-300 dark:hover:text-white"
                >
                  {{ item.label }}
                </a>
                <span v-else class="text-sm text-slate-500 dark:text-dark-300">{{ item.label }}</span>
              </li>
            </ul>
          </div>
        </div>

        <div class="mt-12 flex flex-col gap-4 border-t border-slate-200 pt-6 text-sm text-slate-500 dark:border-dark-800 dark:text-dark-300 sm:flex-row sm:items-center sm:justify-between">
          <p>&copy; {{ currentYear }} {{ siteName }}. {{ copy.allRightsReserved }}</p>
          <div class="flex flex-wrap items-center gap-4">
            <a :href="homeLinks.termsPath" class="hover:text-slate-900 dark:hover:text-white">{{ copy.termsLink }}</a>
            <a :href="homeLinks.privacyPath" class="hover:text-slate-900 dark:hover:text-white">{{ copy.privacyLink }}</a>
          </div>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import DocsLink from '@/components/common/DocsLink.vue'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAuthRouteDefaults } from '@/composables/useAuthRouteDefaults'
import { paymentAPI } from '@/api/payment'
import type { HomeCatalogResponse } from '@/types/payment'
import {
  resolveBusinessHomeShellConfig,
  resolveHomeShellConfig,
  type HomeBusinessCard,
  type HomeShellCopy,
} from '@/utils/homeShell'
import { resolveRuntimeLanguage } from '@/utils/runtimeLocale'
import { buildHomeModelFamilies } from '@/views/home/homeCatalog'

const { locale } = useI18n()
const route = useRoute()
const authStore = useAuthStore()
const appStore = useAppStore()
const { authRouteDefaults, resolveHomePath } = useAuthRouteDefaults()

const emptyCatalog: HomeCatalogResponse = {
  recharge_products: [],
  plans: [],
}

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName)
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const isSubHome = computed(() => route.path === '/sub' || route.name === 'SubHome')
const isBusinessHome = computed(() => !isSubHome.value)
const homeContent = computed(() => (isSubHome.value ? appStore.cachedPublicSettings?.home_content || '' : ''))
const homeLocale = computed<'zh' | 'en'>(() => resolveRuntimeLanguage(locale))

const shellConfig = computed(() =>
  isBusinessHome.value
    ? resolveBusinessHomeShellConfig(appStore.cachedPublicSettings?.home_business_shell_config, homeLocale.value)
    : resolveHomeShellConfig(appStore.cachedPublicSettings?.home_shell_config, homeLocale.value),
)
const copy = computed<HomeShellCopy>(() => shellConfig.value.labels)
const homeLinks = computed(() => shellConfig.value.defaults.links)
const businessCards = computed<HomeBusinessCard[]>(() => shellConfig.value.businessCards)

const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => resolveHomePath(isAdmin.value))
const loginPath = computed(() => authRouteDefaults.value.loginPath)
const currentYear = computed(() => new Date().getFullYear())

const navItems = computed(() => [
  { href: homeLinks.value.homeAnchor, label: copy.value.navHome },
  { doc: true, label: copy.value.navDocs },
  { to: homeLinks.value.modelsPath, label: copy.value.navModels },
  { href: homeLinks.value.experienceAnchor, label: copy.value.navExperience },
])

const publicCatalog = ref<HomeCatalogResponse>(emptyCatalog)
const catalogLoading = ref(false)

const modelFamilies = computed(() => buildHomeModelFamilies(publicCatalog.value))
const visibleModelFamilies = computed(() =>
  modelFamilies.value.length > 0
    ? modelFamilies.value
    : [
        { key: 'claude' as const, name: 'Claude', models: [] },
        { key: 'gpt' as const, name: 'GPT', models: [] },
      ],
)
const modelGridClass = computed(() => {
  if (visibleModelFamilies.value.length <= 1) {
    return 'mx-auto max-w-xl grid-cols-1'
  }
  if (visibleModelFamilies.value.length === 2) {
    return 'mx-auto max-w-4xl md:grid-cols-2'
  }
  return 'lg:grid-cols-3'
})

const experienceCards = computed(() => shellConfig.value.experienceCards)
const whyChooseCards = computed(() => shellConfig.value.whyChooseCards)

const footerSections = computed(() => [
  {
    title: copy.value.footerProduct,
    items: [
      { label: copy.value.navHome, href: homeLinks.value.homeAnchor },
      { label: copy.value.navModels, href: homeLinks.value.modelsPath },
      { label: isAuthenticated.value ? copy.value.dashboard : copy.value.login, href: isAuthenticated.value ? dashboardPath.value : loginPath.value },
    ],
  },
  {
    title: copy.value.footerCatalog,
    items: [
      { label: copy.value.navModels, href: homeLinks.value.modelsPath },
      { label: copy.value.navExperience, href: homeLinks.value.experienceAnchor },
    ],
  },
  {
    title: copy.value.footerSupport,
    items: [
      { label: copy.value.viewDocs, href: docUrl.value || homeLinks.value.docsPath },
      { label: copy.value.termsLink, href: homeLinks.value.termsPath },
      { label: copy.value.privacyLink, href: homeLinks.value.privacyPath },
    ],
  },
])

function familyBadge(key: string): string {
  switch (key) {
    case 'claude':
      return copy.value.familyClaudeBadge
    case 'gpt':
      return copy.value.familyGptBadge
    default:
      return siteName.value
  }
}

function familyTagline(key: string): string {
  switch (key) {
    case 'claude':
      return copy.value.familyClaudeTagline
    case 'gpt':
      return copy.value.familyGptTagline
    default:
      return ''
  }
}

function familyDescription(key: string): string {
  switch (key) {
    case 'claude':
      return copy.value.familyClaudeDescription
    case 'gpt':
      return copy.value.familyGptDescription
    default:
      return ''
  }
}

function familyCapabilities(key: string): string[] {
  switch (key) {
    case 'claude':
      return [
        copy.value.familyClaudeReasoning,
        copy.value.familyClaudeArchitecture,
        copy.value.familyClaudeReview,
      ]
    case 'gpt':
      return [
        copy.value.familyGptCoding,
        copy.value.familyGptIteration,
        copy.value.familyGptAgents,
      ]
    default:
      return []
  }
}

function familyIconClass(key: string): string {
  switch (key) {
    case 'claude':
      return 'bg-gradient-to-br from-orange-400 to-orange-500'
    case 'gpt':
      return 'bg-gradient-to-br from-emerald-500 to-green-600'
    default:
      return 'bg-gradient-to-br from-slate-500 to-slate-700'
  }
}

async function loadPublicCatalog() {
  catalogLoading.value = true
  try {
    const response = await paymentAPI.getPublicCatalog()
    publicCatalog.value = response.data || emptyCatalog
  } catch (error) {
    console.error('[home] failed to load public catalog', error)
    publicCatalog.value = emptyCatalog
  } finally {
    catalogLoading.value = false
  }
}

onMounted(() => {
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
  if (!homeContent.value.trim() && !isBusinessHome.value) {
    void loadPublicCatalog()
  }
})
</script>
