<template>
  <AppLayout>
    <div class="mx-auto max-w-4xl space-y-6">
      <section class="rounded-2xl border border-gray-200 bg-white p-6 shadow-sm dark:border-dark-700 dark:bg-dark-900 sm:p-8">
        <div class="flex flex-col gap-6 lg:flex-row lg:items-start lg:justify-between">
          <div>
            <p class="text-sm font-semibold uppercase tracking-[0.2em] text-primary-600 dark:text-primary-300">
              {{ copy.eyebrow }}
            </p>
            <h1 class="mt-3 text-3xl font-black text-gray-950 dark:text-white sm:text-4xl">
              {{ pageTitle }}
            </h1>
            <p class="mt-3 max-w-2xl text-sm leading-7 text-gray-500 dark:text-dark-300">
              {{ pageDescription }}
            </p>
          </div>

          <div class="flex flex-wrap gap-3">
            <RouterLink v-if="purchaseRoute" :to="purchaseRoute" class="btn btn-primary">
              {{ purchaseLabel }}
            </RouterLink>
            <RouterLink v-if="ordersPath" :to="ordersPath" class="btn btn-secondary">
              {{ copy.orders }}
            </RouterLink>
          </div>
        </div>
      </section>

      <section class="grid gap-5 md:grid-cols-2">
        <div class="rounded-2xl border border-gray-200 bg-white p-6 shadow-sm dark:border-dark-700 dark:bg-dark-900">
          <p class="text-sm font-medium text-gray-500 dark:text-dark-300">
            {{ copy.credits }}
          </p>
          <p class="mt-3 text-5xl font-black text-gray-950 dark:text-white">
            {{ formattedCredits }}
          </p>
          <p class="mt-3 text-sm text-gray-500 dark:text-dark-300">
            {{ balanceLabel }}
          </p>
        </div>

        <div class="rounded-2xl border border-gray-200 bg-white p-6 shadow-sm dark:border-dark-700 dark:bg-dark-900">
          <p class="text-sm font-medium text-gray-500 dark:text-dark-300">
            {{ copy.cloudbaseBalance }}
          </p>
          <p class="mt-3 text-5xl font-black text-primary-600 dark:text-primary-300">
            {{ formattedBalanceAmount }}
          </p>
          <p class="mt-3 text-sm text-gray-500 dark:text-dark-300">
            {{ conversionLabel }}
          </p>
        </div>
      </section>

      <section class="rounded-2xl border border-gray-200 bg-white p-6 shadow-sm dark:border-dark-700 dark:bg-dark-900">
        <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h2 class="text-lg font-bold text-gray-950 dark:text-white">
              {{ actionsTitle }}
            </h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-300">
              {{ actionsDescription }}
            </p>
          </div>
          <div class="flex flex-wrap gap-3">
            <RouterLink v-if="rechargeRoute" :to="rechargeRoute" class="btn btn-primary">
              {{ rechargeLabel }}
            </RouterLink>
            <RouterLink to="/settings/credits/ledger" class="btn btn-secondary">
              {{ t('credits.ledger.title') }}
            </RouterLink>
            <RouterLink v-if="ordersPath" :to="ordersPath" class="btn btn-secondary">
              {{ ordersLabel }}
            </RouterLink>
          </div>
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import { getLocale } from '@/i18n'
import { useI18n } from 'vue-i18n'
import { useAppStore, useAuthStore } from '@/stores'
import { resolveCreditsShellConfig, type CreditsCopy } from '@/utils/creditsShell'
import { formatPublicMoneyAmount } from '@/utils/paymentCurrency'
import { resolveRuntimeLanguage } from '@/utils/runtimeLocale'
import {
  buildCreditsPurchaseRoute,
  parseCreditsPerBalance,
  renderCreditsBalanceLabel,
  renderCreditsPerBalance,
  resolveCreditsActionsDescription,
  resolveCreditsActionsTitle,
  resolveCreditsOrdersLabel,
  resolveCreditsOrdersPath,
  resolveCreditsPurchasePath,
  resolveCreditsRechargeLabel,
} from './creditsRuntime'

const appStore = useAppStore()
const authStore = useAuthStore()
const { t } = useI18n()

const locale = computed<'zh' | 'en'>(() => resolveRuntimeLanguage(getLocale()))
const user = computed(() => authStore.user)
const balance = computed(() => Number(user.value?.balance || 0))
const formattedBalance = computed(() => balance.value.toFixed(2))
const formattedCredits = computed(() => formattedBalance.value)
const currencyPrefix = computed(() => appStore.cachedPublicSettings?.pricing_currency_symbol || '')
const formattedBalanceAmount = computed(() => formatPublicMoneyAmount(balance.value, currencyPrefix.value))
const creditsPerBalance = computed(() => parseCreditsPerBalance(appStore.cachedPublicSettings?.credits_per_balance))
const shellConfig = computed(() =>
  resolveCreditsShellConfig(
    appStore.cachedPublicSettings?.credits_shell_config,
    locale.value,
  ),
)
const copy = computed<CreditsCopy>(() => shellConfig.value.labels)

const pageTitle = computed(() => copy.value.title)
const pageDescription = computed(() => copy.value.description)
const purchaseLabel = computed(() => copy.value.purchase)
const balanceLabel = computed(() => renderCreditsBalanceLabel(copy.value.balanceLabel, formattedBalance.value))
const conversionLabel = computed(() =>
  renderCreditsPerBalance(shellConfig.value.conversion?.trim() || copy.value.conversion, creditsPerBalance.value),
)
const actionsTitle = computed(() => resolveCreditsActionsTitle(shellConfig.value, copy.value))
const actionsDescription = computed(() => resolveCreditsActionsDescription(shellConfig.value, copy.value))
const rechargeLabel = computed(() => resolveCreditsRechargeLabel(shellConfig.value, copy.value))
const ordersLabel = computed(() => resolveCreditsOrdersLabel(shellConfig.value, copy.value))
const configuredPurchasePath = computed(() => resolveCreditsPurchasePath(shellConfig.value))
const ordersPath = computed(() => resolveCreditsOrdersPath(shellConfig.value))
const purchaseRoute = computed(() => buildCreditsPurchaseRoute(configuredPurchasePath.value))
const rechargeRoute = computed(() => buildCreditsPurchaseRoute(configuredPurchasePath.value, 'recharge'))

onMounted(() => {
  if (!appStore.publicSettingsLoaded) {
    void appStore.fetchPublicSettings()
  }
  void authStore.refreshUser().catch((error) => {
    console.error('[credits] failed to refresh user', error)
  })
})
</script>
