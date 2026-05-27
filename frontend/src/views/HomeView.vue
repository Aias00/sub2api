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
          <div class="h-9 w-9 overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900">
            <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
          </div>
          <div class="flex items-center gap-2">
            <span class="text-sm font-semibold text-slate-900 dark:text-white">{{ siteName }}</span>
          </div>
        </div>

        <div class="hidden items-center gap-8 text-xs font-medium tracking-[0.12em] text-slate-500 lg:flex">
          <template v-for="item in navItems" :key="item.label">
            <router-link
              v-if="item.to"
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
            v-if="docUrl"
            :doc-url="docUrl"
            class="hidden rounded-full border border-slate-200 px-4 py-2 text-sm font-medium text-slate-600 transition hover:border-slate-300 hover:text-slate-900 dark:border-dark-700 dark:text-dark-200 dark:hover:border-dark-500 dark:hover:text-white sm:inline-flex"
          >
            {{ t('home.viewDocs') }}
          </DocsLink>
          <router-link
            :to="isAuthenticated ? dashboardPath : '/login'"
            class="inline-flex items-center rounded-full border border-slate-900 bg-slate-900 px-4 py-2 text-sm font-semibold text-white transition hover:bg-slate-800 dark:border-white dark:bg-white dark:text-slate-900 dark:hover:bg-slate-100"
          >
            {{ isAuthenticated ? t('home.goToDashboard') : t('home.login') }}
          </router-link>
        </div>
      </nav>
    </header>

    <main class="relative z-10">
      <section id="top" class="px-6 pb-28 pt-12 sm:pb-32 sm:pt-16">
        <div class="mx-auto flex max-w-5xl flex-col items-center text-center">
          <div class="mb-6 inline-flex items-center rounded-full border border-slate-300/80 bg-white/70 px-4 py-2 text-sm font-medium text-slate-700 backdrop-blur dark:border-dark-600 dark:bg-dark-900/70 dark:text-dark-100">
            {{ t('home.heroBadge') }}
          </div>
          <h1 class="max-w-3xl text-balance text-5xl font-black tracking-tight text-slate-950 dark:text-white sm:text-6xl lg:text-7xl">
            {{ t('home.heroTitle') }}
          </h1>
          <p class="mt-6 max-w-3xl text-balance text-base leading-8 text-slate-600 dark:text-dark-200 sm:text-lg">
            {{ t('home.heroDescription') }}
          </p>
          <div class="mt-10 flex flex-col items-center gap-3 sm:flex-row">
            <router-link
              :to="isAuthenticated ? dashboardPath : '/login'"
              class="inline-flex items-center rounded-full bg-slate-950 px-7 py-3 text-sm font-semibold text-white shadow-lg shadow-slate-900/10 transition hover:bg-slate-800 dark:bg-white dark:text-slate-950 dark:hover:bg-slate-100"
            >
              {{ isAuthenticated ? t('home.goToDashboard') : t('home.primaryCta') }}
            </router-link>
            <router-link
              to="/models"
              class="inline-flex items-center rounded-full border border-slate-300 bg-white/70 px-7 py-3 text-sm font-semibold text-slate-700 backdrop-blur transition hover:border-slate-400 hover:text-slate-950 dark:border-dark-600 dark:bg-dark-900/60 dark:text-dark-100 dark:hover:border-dark-400 dark:hover:text-white"
            >
              {{ t('home.secondaryCta') }}
            </router-link>
          </div>
        </div>
      </section>

      <section id="models" class="px-6 py-12 sm:py-16">
        <div class="mx-auto max-w-5xl">
          <div class="text-center">
            <p class="text-sm font-medium uppercase tracking-[0.2em] text-sky-600 dark:text-sky-400">
              {{ t('home.modelMatrixKicker') }}
            </p>
            <h2 class="mt-3 text-3xl font-black tracking-tight text-slate-950 dark:text-white sm:text-4xl">
              {{ t('home.modelMatrixTitle') }}
            </h2>
            <p class="mx-auto mt-4 max-w-xl text-sm leading-7 text-slate-500 dark:text-dark-300 sm:text-base">
              {{ t('home.modelMatrixDescription') }}
            </p>
          </div>

          <div v-if="catalogLoading" class="mt-10 flex justify-center py-10">
            <div class="h-8 w-8 animate-spin rounded-full border-2 border-sky-500 border-t-transparent"></div>
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
                {{ family.models.length > 0 ? familyDescription(family.key) : t('home.modelMatrixEmptyCard') }}
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
                <span class="rounded-full border border-dashed border-slate-300 px-3 py-1 text-xs font-medium text-slate-500 dark:border-dark-600 dark:text-dark-300">
                  {{ t('home.modelMatrixEmptyPill') }}
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
              {{ t('home.experienceKicker') }}
            </p>
            <h2 class="mt-3 text-3xl font-black tracking-tight text-slate-950 dark:text-white sm:text-4xl">
              {{ t('home.experienceTitle') }}
            </h2>
            <p class="mx-auto mt-4 max-w-2xl text-sm leading-7 text-slate-500 dark:text-dark-300 sm:text-base">
              {{ t('home.experienceDescription') }}
            </p>
          </div>

          <div class="mt-12 grid gap-5 md:grid-cols-2 xl:grid-cols-4">
            <article
              v-for="feature in experienceCards"
              :key="feature.title"
              class="rounded-[24px] border border-white/90 bg-white/85 p-6 shadow-[0_10px_30px_rgba(15,23,42,0.04)] backdrop-blur dark:border-dark-800 dark:bg-dark-900/80 dark:shadow-none"
            >
              <div
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
              {{ t('home.whyChooseKicker') }}
            </p>
            <h2 class="mt-3 text-3xl font-black tracking-tight text-slate-950 dark:text-white sm:text-4xl">
              {{ t('home.whyChooseTitle') }}
            </h2>
            <p class="mx-auto mt-4 max-w-2xl text-sm leading-7 text-slate-500 dark:text-dark-300 sm:text-base">
              {{ t('home.whyChooseDescription') }}
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
              <div class="h-9 w-9 overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900">
                <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
              </div>
              <span class="text-sm font-semibold text-slate-950 dark:text-white">{{ siteName }}</span>
            </div>
            <p class="mt-4 max-w-xs text-sm leading-7 text-slate-500 dark:text-dark-300">
              {{ t('home.footerDescription') }}
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
          <p>&copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}</p>
          <div class="flex flex-wrap items-center gap-4">
            <a href="/legal/terms" class="hover:text-slate-900 dark:hover:text-white">{{ t('home.termsLink') }}</a>
            <a href="/legal/privacy-policy" class="hover:text-slate-900 dark:hover:text-white">{{ t('home.privacyLink') }}</a>
          </div>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import DocsLink from '@/components/common/DocsLink.vue'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'
import { paymentAPI } from '@/api/payment'
import type { HomeCatalogResponse } from '@/types/payment'
import { buildHomeModelFamilies } from '@/views/home/homeCatalog'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()

const emptyCatalog: HomeCatalogResponse = {
  recharge_products: [],
  plans: [],
}

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')

const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/dashboard'))
const currentYear = computed(() => new Date().getFullYear())

const navItems = computed(() => [
  { href: '#top', label: t('home.nav.home') },
  { to: '/models', label: t('home.nav.models') },
  { href: '#experience', label: t('home.nav.experience') },
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

const experienceCards = computed(() => [
  {
    icon: 'server' as const,
    iconClass: 'bg-gradient-to-br from-sky-500 to-blue-600',
    title: t('home.cards.unified.title'),
    description: t('home.cards.unified.description'),
  },
  {
    icon: 'key' as const,
    iconClass: 'bg-gradient-to-br from-indigo-500 to-violet-600',
    title: t('home.cards.setup.title'),
    description: t('home.cards.setup.description'),
  },
  {
    icon: 'sparkles' as const,
    iconClass: 'bg-gradient-to-br from-emerald-500 to-teal-600',
    title: t('home.cards.stability.title'),
    description: t('home.cards.stability.description'),
  },
  {
    icon: 'chart' as const,
    iconClass: 'bg-gradient-to-br from-fuchsia-500 to-purple-600',
    title: t('home.cards.billing.title'),
    description: t('home.cards.billing.description'),
  },
])

const whyChooseCards = computed(() => [
  {
    title: t('home.whyCards.lowFriction.title'),
    description: t('home.whyCards.lowFriction.description'),
  },
  {
    title: t('home.whyCards.transparent.title'),
    description: t('home.whyCards.transparent.description'),
  },
  {
    title: t('home.whyCards.routing.title'),
    description: t('home.whyCards.routing.description'),
  },
  {
    title: t('home.whyCards.team.title'),
    description: t('home.whyCards.team.description'),
  },
])

const footerSections = computed(() => [
  {
    title: t('home.footerSections.product'),
    items: [
      { label: t('home.nav.home'), href: '#top' },
      { label: t('home.nav.models'), href: '/models' },
      { label: isAuthenticated.value ? t('home.goToDashboard') : t('home.login'), href: isAuthenticated.value ? dashboardPath.value : '/login' },
    ],
  },
  {
    title: t('home.footerSections.catalog'),
    items: [
      { label: t('home.nav.models'), href: '/models' },
      { label: t('home.nav.experience'), href: '#experience' },
    ],
  },
  {
    title: t('home.footerSections.support'),
    items: [
      { label: t('home.viewDocs'), href: docUrl.value || '/docs' },
      { label: t('home.termsLink'), href: '/legal/terms' },
      { label: t('home.privacyLink'), href: '/legal/privacy-policy' },
    ],
  },
])

function familyBadge(key: string): string {
  switch (key) {
    case 'claude':
      return t('home.familyBadges.claude')
    case 'gpt':
      return t('home.familyBadges.gpt')
    default:
      return siteName.value
  }
}

function familyTagline(key: string): string {
  switch (key) {
    case 'claude':
      return t('home.familyContent.claude.tagline')
    case 'gpt':
      return t('home.familyContent.gpt.tagline')
    default:
      return ''
  }
}

function familyDescription(key: string): string {
  switch (key) {
    case 'claude':
      return t('home.familyContent.claude.description')
    case 'gpt':
      return t('home.familyContent.gpt.description')
    default:
      return ''
  }
}

function familyCapabilities(key: string): string[] {
  switch (key) {
    case 'claude':
      return [
        t('home.familyCapabilities.claude.reasoning'),
        t('home.familyCapabilities.claude.architecture'),
        t('home.familyCapabilities.claude.review'),
      ]
    case 'gpt':
      return [
        t('home.familyCapabilities.gpt.coding'),
        t('home.familyCapabilities.gpt.iteration'),
        t('home.familyCapabilities.gpt.agents'),
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
  if (!homeContent.value.trim()) {
    void loadPublicCatalog()
  }
})
</script>
