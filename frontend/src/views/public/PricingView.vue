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
            v-if="promptsPath"
            :to="promptsPath"
            class="rounded-full border border-white/10 px-4 py-2 text-sm font-semibold text-white/70 transition hover:bg-white/[0.06] hover:text-white"
          >
            {{ copy.prompts }}
          </RouterLink>
        </div>
      </nav>
    </header>

    <main class="px-6 py-10 sm:py-14">
      <div class="mx-auto max-w-7xl">
        <section class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_320px] lg:items-end">
          <div>
            <p class="text-sm font-semibold uppercase tracking-[0.22em] text-emerald-200/70">
              {{ eyebrowLabel }}
            </p>
            <h1 class="mt-4 max-w-4xl text-4xl font-black leading-tight text-white sm:text-5xl">
              {{ pageTitle }}
            </h1>
            <p class="mt-4 max-w-3xl text-base leading-8 text-white/60">
              {{ pageDescription }}
            </p>
          </div>

          <div class="rounded-2xl border border-white/10 bg-white/[0.035] p-5">
            <p class="text-xs font-semibold uppercase tracking-[0.2em] text-white/35">
              {{ catalogStatusLabel }}
            </p>
            <div class="mt-4 grid grid-cols-2 gap-4">
              <div>
                <p class="text-xs text-white/40">{{ rechargeProductsLabel }}</p>
                <p class="mt-1 text-2xl font-black text-white">{{ rechargeProducts.length }}</p>
              </div>
              <div>
                <p class="text-xs text-white/40">{{ subscriptionPlansLabel }}</p>
                <p class="mt-1 text-2xl font-black text-white">{{ subscriptionPlans.length }}</p>
              </div>
            </div>
          </div>
        </section>

        <section class="mt-8 rounded-2xl border border-white/10 bg-[#17181d] p-4 sm:p-5">
          <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div class="flex rounded-xl border border-white/10 bg-white/[0.035] p-1">
              <button
                v-for="tab in tabs"
                :key="tab.value"
                type="button"
                class="rounded-lg px-4 py-2 text-sm font-bold transition"
                :class="activeTab === tab.value ? 'bg-emerald-300 text-slate-950' : 'text-white/55 hover:bg-white/[0.06] hover:text-white'"
                @click="activeTab = tab.value"
              >
                {{ tab.label }}
              </button>
            </div>

            <RouterLink
              v-if="purchasePath"
              :to="purchasePath"
              class="inline-flex items-center justify-center rounded-xl bg-emerald-400 px-5 py-3 text-sm font-black text-slate-950 transition hover:bg-emerald-300"
            >
              {{ primaryCta }}
            </RouterLink>
          </div>

          <div v-if="loading" class="flex items-center justify-center py-24">
            <div class="h-8 w-8 animate-spin rounded-full border-4 border-emerald-300 border-t-transparent"></div>
          </div>

          <div v-else-if="loadError" class="mt-5 rounded-2xl border border-red-300/20 bg-red-300/10 p-5 text-sm text-red-50">
            {{ loadFailedLabel }}
          </div>

          <div v-else-if="activeTab === 'recharge'" class="mt-5">
            <div v-if="rechargeProducts.length === 0" class="rounded-2xl border border-white/10 bg-white/[0.035] p-10 text-center text-white/50">
              {{ emptyRechargeLabel }}
            </div>
            <div v-else class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
              <article
                v-for="product in rechargeProducts"
                :key="product.id"
                class="flex min-h-72 flex-col rounded-2xl border border-white/10 bg-white/[0.04] p-5 transition hover:border-emerald-200/35 hover:bg-white/[0.06]"
              >
                <div class="flex items-start justify-between gap-3">
                  <div>
                    <p v-if="product.badge || product.recommended" class="mb-3 inline-flex rounded-full bg-emerald-300/15 px-3 py-1 text-xs font-bold text-emerald-100">
                      {{ product.badge || recommendedLabel }}
                    </p>
                    <h2 v-if="product.name" class="text-xl font-black text-white">{{ product.name }}</h2>
                  </div>
                  <p class="shrink-0 text-right text-3xl font-black text-emerald-200">
                    {{ formatCurrency(product.amount) }}
                  </p>
                </div>

                <p v-if="product.description" class="mt-4 text-sm leading-7 text-white/55">
                  {{ product.description }}
                </p>

                <div class="mt-5 rounded-2xl border border-white/10 bg-black/15 p-4 text-sm">
                  <div class="flex items-center justify-between gap-3">
                    <span class="text-white/45">{{ creditedBalanceLabel }}</span>
                    <span class="font-black text-white">{{ formatCredits(product.credited_amount) }}</span>
                  </div>
                </div>

                <ul v-if="product.features?.length" class="mt-5 space-y-2 text-sm text-white/60">
                  <li v-for="feature in product.features" :key="feature" class="flex gap-2">
                    <span class="mt-2 h-1.5 w-1.5 shrink-0 rounded-full bg-emerald-300"></span>
                    <span>{{ feature }}</span>
                  </li>
                </ul>

                <RouterLink
                  v-if="rechargePurchaseRoute"
                  :to="rechargePurchaseRoute"
                  class="mt-auto inline-flex items-center justify-center rounded-xl border border-white/10 px-4 py-3 text-sm font-bold text-white/75 transition hover:bg-white/[0.06] hover:text-white"
                >
                  {{ buyButton }}
                </RouterLink>
              </article>
            </div>
          </div>

          <div v-else class="mt-5">
            <div v-if="subscriptionPlans.length === 0" class="rounded-2xl border border-white/10 bg-white/[0.035] p-10 text-center text-white/50">
              {{ emptyPlansLabel }}
            </div>
            <div v-else class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
              <article
                v-for="plan in subscriptionPlans"
                :key="plan.id"
                class="flex min-h-80 flex-col rounded-2xl border border-white/10 bg-white/[0.04] p-5 transition hover:border-cyan-200/35 hover:bg-white/[0.06]"
              >
                <div class="flex items-start justify-between gap-3">
                  <div>
                    <p v-if="planSourceLabel(plan)" class="mb-3 inline-flex rounded-full bg-cyan-300/15 px-3 py-1 text-xs font-bold text-cyan-100">
                      {{ planSourceLabel(plan) }}
                    </p>
                    <h2 class="text-xl font-black text-white">{{ plan.name }}</h2>
                  </div>
                  <div class="shrink-0 text-right">
                    <p v-if="plan.original_price" class="text-sm text-white/35 line-through">
                      {{ formatCurrency(plan.original_price) }}
                    </p>
                    <p class="text-3xl font-black text-cyan-100">{{ formatCurrency(plan.price) }}</p>
                    <p class="text-xs text-white/40">{{ formatValidity(plan) }}</p>
                  </div>
                </div>

                <p v-if="plan.description" class="mt-4 text-sm leading-7 text-white/55">
                  {{ plan.description }}
                </p>

                <div class="mt-5 grid grid-cols-2 gap-3 text-sm">
                  <div v-if="plan.rate_multiplier != null" class="rounded-2xl border border-white/10 bg-black/15 p-3">
                    <p class="text-white/40">{{ rateLabel }}</p>
                    <p class="mt-1 font-black text-white">x{{ plan.rate_multiplier }}</p>
                  </div>
                  <div v-if="plan.quota_label" class="rounded-2xl border border-white/10 bg-black/15 p-3">
                    <p class="text-white/40">{{ quotaLabelText }}</p>
                    <p class="mt-1 font-black text-white">{{ plan.quota_label }}</p>
                  </div>
                </div>

                <ul v-if="plan.features?.length" class="mt-5 space-y-2 text-sm text-white/60">
                  <li v-for="feature in plan.features" :key="feature" class="flex gap-2">
                    <span class="mt-2 h-1.5 w-1.5 shrink-0 rounded-full bg-cyan-300"></span>
                    <span>{{ feature }}</span>
                  </li>
                </ul>

                <RouterLink
                  v-if="subscriptionPurchaseRoute(plan)"
                  :to="subscriptionPurchaseRoute(plan)!"
                  class="mt-auto inline-flex items-center justify-center rounded-xl border border-white/10 px-4 py-3 text-sm font-bold text-white/75 transition hover:bg-white/[0.06] hover:text-white"
                >
                  {{ buyButton }}
                </RouterLink>
              </article>
            </div>
          </div>
        </section>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import { paymentAPI } from '@/api/payment'
import { getLocale } from '@/i18n'
import { useAuthRouteDefaults } from '@/composables/useAuthRouteDefaults'
import { useAppStore } from '@/stores'
import { resolvePricingShellConfig, type PricingCopy } from '@/utils/pricingShell'
import { resolveRuntimeLanguage } from '@/utils/runtimeLocale'
import type { HomeCatalogResponse, SubscriptionPlan } from '@/types/payment'
import {
  buildPricingPurchaseRoute,
  comparePricingCatalogItems,
  formatPricingCredits,
  formatPricingCurrency,
  resolvePricingBuyButton,
  resolvePricingPlanSourceLabel,
  resolvePricingPlanValidity,
  resolvePricingPurchasePath,
  resolvePricingPromptsPath,
  resolvePricingShellGroupLabel,
  resolvePricingShellLabel,
} from './pricingRuntime'

const emptyCatalog: HomeCatalogResponse = {
  recharge_products: [],
  plans: [],
}

const appStore = useAppStore()
const { authRouteDefaults } = useAuthRouteDefaults()
const activeTab = ref<'recharge' | 'subscription'>('recharge')
const catalog = ref<HomeCatalogResponse>(emptyCatalog)
const loading = ref(false)
const loadError = ref(false)

const locale = computed<'zh' | 'en'>(() => resolveRuntimeLanguage(getLocale()))
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName)
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const shellConfig = computed(() =>
  resolvePricingShellConfig(
    appStore.cachedPublicSettings?.pricing_shell_config,
    locale.value,
  ),
)
const copy = computed<PricingCopy>(() => shellConfig.value.labels)

const pageTitle = computed(() => copy.value.title)
const pageDescription = computed(() => copy.value.description)

const rechargeProducts = computed(() =>
  [...(catalog.value.recharge_products || [])].sort(comparePricingCatalogItems),
)
const subscriptionPlans = computed(() =>
  [...(catalog.value.plans || [])].sort(comparePricingCatalogItems),
)

const tabs = computed(() => [
  { value: 'recharge' as const, label: resolvePricingShellGroupLabel(shellConfig.value, 'one-time') || copy.value.recharge },
  { value: 'subscription' as const, label: resolvePricingShellGroupLabel(shellConfig.value, 'subscription') || copy.value.subscription },
])

const eyebrowLabel = computed(() => resolvePricingShellLabel(shellConfig.value, 'eyebrow'))
const buyButton = computed(() => resolvePricingBuyButton(shellConfig.value, copy.value))
const catalogStatusLabel = computed(() => resolvePricingShellLabel(shellConfig.value, 'catalogStatus'))
const rechargeProductsLabel = computed(() => resolvePricingShellLabel(shellConfig.value, 'rechargeProducts'))
const subscriptionPlansLabel = computed(() => resolvePricingShellLabel(shellConfig.value, 'subscriptionPlans'))
const rechargeCtaLabel = computed(() => resolvePricingShellLabel(shellConfig.value, 'rechargeCta'))
const subscriptionCtaLabel = computed(() => resolvePricingShellLabel(shellConfig.value, 'subscriptionCta'))
const loadFailedLabel = computed(() => resolvePricingShellLabel(shellConfig.value, 'loadFailed'))
const emptyRechargeLabel = computed(() => resolvePricingShellLabel(shellConfig.value, 'emptyRecharge'))
const emptyPlansLabel = computed(() => resolvePricingShellLabel(shellConfig.value, 'emptyPlans'))
const recommendedLabel = computed(() => resolvePricingShellLabel(shellConfig.value, 'recommended'))
const creditedBalanceLabel = computed(() => resolvePricingShellLabel(shellConfig.value, 'creditedBalance'))
const rateLabel = computed(() => resolvePricingShellLabel(shellConfig.value, 'rate'))
const quotaLabelText = computed(() => resolvePricingShellLabel(shellConfig.value, 'quota'))
const dayLabel = computed(() => resolvePricingShellLabel(shellConfig.value, 'day'))
const daysLabel = computed(() => resolvePricingShellLabel(shellConfig.value, 'days'))
const monthLabel = computed(() => resolvePricingShellLabel(shellConfig.value, 'month'))
const primaryCta = computed(() => activeTab.value === 'subscription' ? subscriptionCtaLabel.value : rechargeCtaLabel.value)
const promptsPath = computed(() => resolvePricingPromptsPath(shellConfig.value))
const configuredPurchasePath = computed(() => resolvePricingPurchasePath(shellConfig.value))
const purchasePath = computed(() =>
  buildPricingPurchaseRoute(configuredPurchasePath.value, activeTab.value === 'subscription' ? 'subscription' : 'recharge'),
)
const rechargePurchaseRoute = computed(() => buildPricingPurchaseRoute(configuredPurchasePath.value, 'recharge'))
const pricingCurrencySymbol = computed(() => appStore.cachedPublicSettings?.pricing_currency_symbol?.trim() || '')

function formatCurrency(value: number | undefined) {
  return formatPricingCurrency(value, pricingCurrencySymbol.value)
}

function formatCredits(value: number | undefined) {
  return formatPricingCredits(value)
}

function planSourceLabel(plan: SubscriptionPlan): string {
  return resolvePricingPlanSourceLabel(plan)
}

function subscriptionPurchaseRoute(plan: SubscriptionPlan) {
  return buildPricingPurchaseRoute(configuredPurchasePath.value, 'subscription', String(plan.group_id))
}

function formatValidity(plan: SubscriptionPlan) {
  return resolvePricingPlanValidity(plan, {
    day: dayLabel.value,
    days: daysLabel.value,
    month: monthLabel.value,
  })
}

async function loadCatalog() {
  loading.value = true
  loadError.value = false
  try {
    const response = await paymentAPI.getPublicCatalog()
    catalog.value = response.data || emptyCatalog
  } catch (error) {
    console.error('[pricing] failed to load public catalog', error)
    catalog.value = emptyCatalog
    loadError.value = true
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  if (!appStore.publicSettingsLoaded) {
    void appStore.fetchPublicSettings()
  }
  void loadCatalog()
})
</script>
