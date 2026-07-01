<template>
  <AppLayout content-container-class="max-w-6xl" header-container-class="max-w-6xl">
    <component :is="activeView" app-shell />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import TaskListView from '@/views/public/TaskListView.vue'
import PromptCatalogView from '@/views/public/PromptCatalogView.vue'
import ImageGeneratorView from '@/views/public/ImageGeneratorView.vue'
import WeChatExportView from '@/views/public/WeChatExportView.vue'
import HotContentView from '@/views/public/HotContentView.vue'

const route = useRoute()

const viewMap = {
  tasks: TaskListView,
  prompts: PromptCatalogView,
  imageGenerator: ImageGeneratorView,
  wechat: WeChatExportView,
  hot: HotContentView,
}

const activeView = computed(() => {
  const key = route.meta.toolView
  return typeof key === 'string' && key in viewMap
    ? viewMap[key as keyof typeof viewMap]
    : TaskListView
})
</script>
