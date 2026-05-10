<template>
  <button
    type="button"
    :class="[
      'group relative flex h-full flex-col overflow-hidden rounded-3xl border text-left transition-all duration-200',
      selected
        ? 'border-amber-400 bg-white shadow-[0_18px_50px_rgba(15,23,42,0.12)] dark:border-amber-400/70 dark:bg-slate-900'
        : 'border-gray-200 bg-white hover:-translate-y-0.5 hover:border-gray-300 hover:shadow-lg dark:border-dark-700 dark:bg-dark-900/60 dark:hover:border-dark-500'
    ]"
    @click="emit('select', product)"
  >
    <div
      v-if="product.recommended || product.badge"
      :class="[
        'px-4 py-2 text-center text-xs font-semibold tracking-[0.18em]',
        selected
          ? 'bg-amber-500 text-white'
          : 'bg-amber-100 text-amber-700 dark:bg-amber-500/15 dark:text-amber-200'
      ]"
    >
      {{ product.badge || t('payment.rechargeProducts.recommended') }}
    </div>

    <div class="flex flex-1 flex-col p-5">
      <div class="space-y-1">
        <h3 class="text-[1.75rem] font-semibold tracking-tight text-gray-950 dark:text-white">
          {{ product.name }}
        </h3>
        <p class="text-sm text-gray-500 dark:text-gray-400">
          {{ product.description }}
        </p>
      </div>

      <div class="mt-5">
        <div class="flex items-end gap-1 text-gray-900 dark:text-white">
          <span class="text-lg text-gray-400 dark:text-gray-500">¥</span>
          <span class="text-4xl font-black tracking-tight">{{ product.amount }}</span>
        </div>
        <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">
          {{ t('payment.rechargeProducts.creditLine', { amount: product.credited_amount.toFixed(2) }) }}
        </p>
      </div>

      <div class="mt-5 space-y-2">
        <div
          v-for="feature in product.features"
          :key="feature"
          class="flex items-start gap-2 text-sm text-gray-600 dark:text-gray-300"
        >
          <svg class="mt-0.5 h-4 w-4 flex-shrink-0 text-gray-400 dark:text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
          </svg>
          <span>{{ feature }}</span>
        </div>
      </div>

      <div class="flex-1" />

      <div
        :class="[
          'mt-6 rounded-2xl px-4 py-3 text-center text-sm font-semibold transition-colors',
          selected
            ? 'bg-amber-500 text-white'
            : 'bg-slate-900 text-white dark:bg-slate-800'
        ]"
      >
        {{ t('payment.rechargeProducts.cta') }}
      </div>
    </div>
  </button>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { RechargeProduct } from '@/types/payment'

defineProps<{
  product: RechargeProduct
  selected?: boolean
}>()

const emit = defineEmits<{
  select: [product: RechargeProduct]
}>()

const { t } = useI18n()
</script>
