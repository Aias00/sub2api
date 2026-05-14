<template>
  <div :class="rootClasses">
    <!-- Icon -->
    <div
      :class="iconWrapClasses"
    >
      <slot name="icon">
        <component v-if="icon" :is="icon" :class="iconClasses" aria-hidden="true" />
        <svg
          v-else
          :class="iconClasses"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
          stroke-width="1.5"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"
          />
        </svg>
      </slot>
    </div>

    <!-- Title -->
    <h3 class="empty-state-title">
      {{ displayTitle }}
    </h3>

    <!-- Description -->
    <p v-if="displayDescription" class="empty-state-description">
      {{ displayDescription }}
    </p>

    <!-- Action -->
    <div v-if="actionText || $slots.action" class="mt-6">
      <slot name="action">
        <component
          :is="actionTo ? 'RouterLink' : 'button'"
          v-if="actionText"
          :to="actionTo"
          @click="!actionTo && $emit('action')"
          class="btn btn-primary"
        >
          <Icon v-if="actionIcon" name="plus" size="md" class="mr-2" />
          {{ actionText }}
        </component>
      </slot>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Component } from 'vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

interface Props {
  icon?: Component | string
  title?: string
  description?: string
  actionText?: string
  actionTo?: string | object
  actionIcon?: boolean
  message?: string
  variant?: 'plain' | 'panel'
  size?: 'sm' | 'md' | 'lg'
}

const props = withDefaults(defineProps<Props>(), {
  description: '',
  actionIcon: true,
  variant: 'plain',
  size: 'md'
})

const displayTitle = computed(() => props.title || props.message || t('common.noData'))
const displayDescription = computed(() => props.description || (props.title ? props.message || '' : ''))

const rootClasses = computed(() => [
  'empty-state',
  props.variant === 'panel' &&
    'mx-auto w-full max-w-2xl rounded-[2rem] border border-gray-200/80 bg-white/85 shadow-[0_24px_80px_rgba(17,24,39,0.08)] backdrop-blur-sm dark:border-dark-700 dark:bg-dark-800/80',
  props.size === 'sm' && 'py-8',
  props.size === 'md' && 'py-10',
  props.size === 'lg' && 'px-6 py-12 sm:px-10 sm:py-14'
])

const iconWrapClasses = computed(() => [
  'mb-5 flex items-center justify-center rounded-3xl border border-gray-200/80 bg-white text-gray-300 shadow-sm dark:border-dark-700 dark:bg-dark-800 dark:text-dark-500',
  props.size === 'sm' ? 'h-14 w-14' : props.size === 'lg' ? 'h-20 w-20' : 'h-16 w-16'
])

const iconClasses = computed(() => [
  'empty-state-icon',
  props.size === 'sm' ? 'h-7 w-7' : props.size === 'lg' ? 'h-10 w-10' : 'h-8 w-8'
])

defineEmits(['action'])
</script>
