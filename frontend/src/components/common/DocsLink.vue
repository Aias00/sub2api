<template>
  <RouterLink v-if="link.internal" :to="link.to" v-bind="$attrs">
    <slot />
  </RouterLink>
  <a v-else :href="link.href" target="_blank" rel="noopener noreferrer" v-bind="$attrs">
    <slot />
  </a>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import { resolveDocsLink } from '@/utils/docs'

defineOptions({
  inheritAttrs: false,
})

const props = withDefaults(
  defineProps<{
    docUrl?: string
  }>(),
  {
    docUrl: '',
  }
)

const link = computed(() =>
  resolveDocsLink(props.docUrl, typeof window !== 'undefined' ? window.location.origin : '')
)
</script>
