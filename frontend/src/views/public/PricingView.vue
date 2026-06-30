<template>
  <div class="home-business-page public-template-page min-h-screen">
    <PublicDarkHeader>
      <template #actions>
        <RouterLink
          v-if="promptsPath"
          :to="promptsPath"
          class="rounded-full border border-[var(--public-border)] px-4 py-2 text-sm font-semibold text-[var(--public-body)] transition hover:bg-[var(--public-panel-soft)] hover:text-[var(--public-ink)]"
        >
          {{ copy.prompts }}
        </RouterLink>
      </template>
    </PublicDarkHeader>

    <main class="public-template-main">
      <div class="public-template-container-wide">
        <section class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_320px] lg:items-end">
          <div>
            <p class="text-sm font-semibold uppercase tracking-[0.22em] text-[var(--public-success)]/70">
              {{ eyebrowLabel }}
            </p>
            <h1 class="mt-4 max-w-4xl text-4xl font-black leading-tight text-[var(--public-ink)] sm:text-5xl">
              {{ pageTitle }}
            </h1>
            <p class="mt-4 max-w-3xl text-base leading-8 text-[var(--public-body)]">
              {{ pageDescription }}
            </p>
          </div>

          <div class="rounded-2xl public-template-panel-muted p-5">
            <p class="text-xs font-semibold uppercase tracking-[0.2em] text-[var(--public-faint)]">
              {{ catalogStatusLabel }}
            </p>
            <div class="mt-4 grid grid-cols-2 gap-4">
              <div>
                <p class="text-xs text-[var(--public-muted)]">{{ rechargeProductsLabel }}</p>
                <p class="mt-1 text-2xl font-black text-[var(--public-ink)]">{{ rechargeProducts.length }}</p>
              </div>
              <div>
                <p class="text-xs text-[var(--public-muted)]">{{ subscriptionPlansLabel }}</p>
                <p class="mt-1 text-2xl font-black text-[var(--public-ink)]">{{ subscriptionPlans.length }}</p>
              </div>
            </div>
          </div>
        </section>

        <section class="mt-8 rounded-2xl public-template-panel p-4 sm:p-5">
          <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div class="flex rounded-xl public-template-panel-muted p-1">
              <button
                v-for="tab in tabs"
                :key="tab.value"
                type="button"
                class="rounded-lg px-4 py-2 text-sm font-bold transition"
                :class="activeTab === tab.value ? 'bg-emerald-300 text-slate-950' : 'text-[var(--public-body)] hover:bg-[var(--public-panel-soft)] hover:text-[var(--public-ink)]'"
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

          <div v-else-if="loadError" class="mt-5 rounded-2xl border border-red-300/20 bg-red-300/10 p-5 text-sm text-[var(--public-danger)]">
            {{ loadFailedLabel }}
          </div>

          <div v-else-if="activeTab === 'recharge'" class="mt-5">
            <div v-if="rechargeProducts.length === 0" class="rounded-2xl public-template-panel-muted p-10 text-center text-[var(--public-muted)]">
              {{ emptyRechargeLabel }}
            </div>
            <div v-else class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
              <article
                v-for="product in rechargeProducts"
                :key="product.id"
                class="flex min-h-72 flex-col rounded-2xl border public-template-input p-5 shadow-[0_20px_60px_rgba(0,0,0,0.24)] transition hover:border-emerald-200/35 hover:bg-[var(--public-panel-soft)]"
              >
                <div class="flex items-start justify-between gap-3">
                  <div>
                    <p v-if="product.badge || product.recommended" class="mb-3 inline-flex rounded-full bg-emerald-300/15 px-3 py-1 text-xs font-bold text-[var(--public-success)]">
                      {{ product.badge || recommendedLabel }}
                    </p>
                    <h2 v-if="product.name" class="text-xl font-black text-[var(--public-ink)]">{{ product.name }}</h2>
                  </div>
                  <p class="shrink-0 text-right text-3xl font-black text-[var(--public-success)]">
                    {{ formatCurrency(product.amount) }}
                  </p>
                </div>

                <p v-if="product.description" class="mt-4 text-sm leading-7 text-[var(--public-body)]">
                  {{ product.description }}
                </p>

                <div class="mt-5 rounded-2xl border border-[var(--public-border)] bg-[var(--public-panel-muted)] p-4 text-sm">
                  <div class="flex items-center justify-between gap-3">
                    <span class="text-[var(--public-muted)]">{{ creditedBalanceLabel }}</span>
                    <span class="font-black text-[var(--public-ink)]">{{ formatCredits(product.credited_amount) }}</span>
                  </div>
                </div>

                <ul v-if="product.features?.length" class="mt-5 space-y-2 text-sm text-[var(--public-body)]">
                  <li v-for="feature in product.features" :key="feature" class="flex gap-2">
                    <span class="mt-2 h-1.5 w-1.5 shrink-0 rounded-full bg-emerald-300"></span>
                    <span>{{ feature }}</span>
                  </li>
                </ul>

                <RouterLink
                  v-if="rechargePurchaseRoute"
                  :to="rechargePurchaseRoute"
                  class="mt-auto inline-flex items-center justify-center rounded-xl border border-[var(--public-border)] px-4 py-3 text-sm font-bold text-[var(--public-body)] transition hover:bg-[var(--public-panel-soft)] hover:text-[var(--public-ink)]"
                >
                  {{ buyButton }}
                </RouterLink>
              </article>
            </div>
          </div>

          <div v-else class="mt-5">
            <div v-if="subscriptionPlans.length === 0" class="rounded-2xl public-template-panel-muted p-10 text-center text-[var(--public-muted)]">
              {{ emptyPlansLabel }}
            </div>
            <div v-else class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
              <article
                v-for="plan in subscriptionPlans"
                :key="plan.id"
                class="flex min-h-80 flex-col rounded-2xl border public-template-input p-5 shadow-[0_20px_60px_rgba(0,0,0,0.24)] transition hover:border-cyan-200/35 hover:bg-[var(--public-panel-soft)]"
              >
                <div class="flex items-start justify-between gap-3">
                  <div>
                    <p v-if="planSourceLabel(plan)" class="mb-3 inline-flex rounded-full bg-cyan-300/15 px-3 py-1 text-xs font-bold text-[var(--public-accent-strong)]">
                      {{ planSourceLabel(plan) }}
                    </p>
                    <h2 class="text-xl font-black text-[var(--public-ink)]">{{ plan.name }}</h2>
                  </div>
                  <div class="shrink-0 text-right">
                    <p v-if="plan.original_price" class="text-sm text-[var(--public-faint)] line-through">
                      {{ formatCurrency(plan.original_price) }}
                    </p>
                    <p class="text-3xl font-black text-[var(--public-accent-strong)]">{{ formatCurrency(plan.price) }}</p>
                    <p class="text-xs text-[var(--public-muted)]">{{ formatValidity(plan) }}</p>
                  </div>
                </div>

                <p v-if="plan.description" class="mt-4 text-sm leading-7 text-[var(--public-body)]">
                  {{ plan.description }}
                </p>

                <div class="mt-5 grid grid-cols-2 gap-3 text-sm">
                  <div v-if="plan.rate_multiplier != null" class="rounded-2xl border border-[var(--public-border)] bg-[var(--public-panel-muted)] p-3">
                    <p class="text-[var(--public-muted)]">{{ rateLabel }}</p>
                    <p class="mt-1 font-black text-[var(--public-ink)]">x{{ plan.rate_multiplier }}</p>
                  </div>
                  <div v-if="plan.quota_label" class="rounded-2xl border border-[var(--public-border)] bg-[var(--public-panel-muted)] p-3">
                    <p class="text-[var(--public-muted)]">{{ quotaLabelText }}</p>
                    <p class="mt-1 font-black text-[var(--public-ink)]">{{ plan.quota_label }}</p>
                  </div>
                </div>

                <ul v-if="plan.features?.length" class="mt-5 space-y-2 text-sm text-[var(--public-body)]">
                  <li v-for="feature in plan.features" :key="feature" class="flex gap-2">
                    <span class="mt-2 h-1.5 w-1.5 shrink-0 rounded-full bg-cyan-300"></span>
                    <span>{{ feature }}</span>
                  </li>
                </ul>

                <RouterLink
                  v-if="subscriptionPurchaseRoute(plan)"
                  :to="subscriptionPurchaseRoute(plan)!"
                  class="mt-auto inline-flex items-center justify-center rounded-xl border border-[var(--public-border)] px-4 py-3 text-sm font-bold text-[var(--public-body)] transition hover:bg-[var(--public-panel-soft)] hover:text-[var(--public-ink)]"
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
import PublicDarkHeader from '@/components/layout/PublicDarkHeader.vue'
import { paymentAPI } from '@/api/payment'
import { getLocale } from '@/i18n'
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
const activeTab = ref<'recharge' | 'subscription'>('recharge')
const catalog = ref<HomeCatalogResponse>(emptyCatalog)
const loading = ref(false)
const loadError = ref(false)

const locale = computed<'zh' | 'en'>(() => resolveRuntimeLanguage(getLocale()))
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
