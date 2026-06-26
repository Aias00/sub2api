<template>
  <AppLayout>
    <div class="mx-auto max-w-lg space-y-6 py-8">
      <div v-if="loading" class="flex items-center justify-center py-20">
        <div class="h-8 w-8 animate-spin rounded-full border-4 border-emerald-500 border-t-transparent"></div>
      </div>

      <div v-else-if="errorMessage" class="card p-8 text-center">
        <div class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-red-100 dark:bg-red-900/30">
          <Icon name="exclamationCircle" size="xl" class="text-red-500" />
        </div>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{{ paymentText('airwallexLoadFailed') }}</h3>
        <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">{{ errorMessage }}</p>
        <button class="btn btn-primary mt-6" @click="router.push(authRouteDefaults.purchasePath)">{{ paymentText('backToRecharge') }}</button>
      </div>

      <div v-else class="card p-6">
        <div class="flex flex-col items-center space-y-4 py-4">
          <div class="h-10 w-10 animate-spin rounded-full border-4 border-emerald-500 border-t-transparent"></div>
          <p class="text-sm text-gray-500 dark:text-gray-400">{{ paymentText('payInNewWindowHint') }}</p>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { resolveRuntimeLanguage, resolveRuntimeLocale } from '@/utils/runtimeLocale'
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores'
import {
  type PaymentRecoverySnapshot,
} from '@/components/payment/paymentFlow'
import { normalizePaymentCountryCode, normalizePaymentCurrency } from '@/components/payment/currency'
import {
  renderAirwallexPaymentText,
  resolveAirwallexPaymentLabels,
  type AirwallexPaymentLabelKey,
} from '@/utils/paymentShell'
import { useAuthRouteDefaults } from '@/composables/useAuthRouteDefaults'
import {
  buildAirwallexSuccessUrl,
  restoreAirwallexPaymentSnapshot,
} from './airwallexPaymentRuntime'

const { locale } = useI18n()
const route = useRoute()
const router = useRouter()
const appStore = useAppStore()
const { authRouteDefaults } = useAuthRouteDefaults()

const loading = ref(true)
const errorMessage = ref('')


const airwallexPaymentLabels = computed(() =>
  resolveAirwallexPaymentLabels(
    appStore.cachedPublicSettings?.payment_shell_config,
    resolveRuntimeLocale(locale),
  ),
)

function paymentText(key: AirwallexPaymentLabelKey): string {
  return renderAirwallexPaymentText(airwallexPaymentLabels.value, key)
}

function buildSuccessUrl(snapshot: PaymentRecoverySnapshot): string {
  return buildAirwallexSuccessUrl(
    authRouteDefaults.value.paymentResultPath,
    route.query,
    snapshot,
    window.location.origin,
  )
}

function restoreAirwallexSnapshot(): PaymentRecoverySnapshot | null {
  return restoreAirwallexPaymentSnapshot(
    typeof window === 'undefined' ? null : window.localStorage,
    route.query,
  )
}

onMounted(async () => {
  const snapshot = restoreAirwallexSnapshot()
  const checkoutLocale = resolveRuntimeLanguage(locale)

  if (!snapshot) {
    loading.value = false
    errorMessage.value = paymentText('airwallexMissingParams')
    return
  }

  try {
    const airwallex = await import('@airwallex/components-sdk')
    const result = await airwallex.init({
      env: snapshot.paymentEnv === 'prod' ? 'prod' : 'demo',
      enabledElements: ['payments'],
      locale: checkoutLocale,
    })

    loading.value = false
    const checkoutOptions = {
      intent_id: snapshot.intentId,
      client_secret: snapshot.clientSecret,
      currency: normalizePaymentCurrency(snapshot.currency),
      country_code: normalizePaymentCountryCode(snapshot.countryCode),
      successUrl: buildSuccessUrl(snapshot),
    }
    if (!result.payments) {
      throw new Error(paymentText('airwallexLoadFailed'))
    }
    const redirectResult = result.payments.redirectToCheckout(checkoutOptions)

    if (typeof redirectResult === 'string' && redirectResult) {
      window.location.assign(redirectResult)
    }
  } catch (err: unknown) {
    loading.value = false
    errorMessage.value = err instanceof Error && err.message
      ? err.message
      : paymentText('airwallexLoadFailed')
  }
})
</script>
