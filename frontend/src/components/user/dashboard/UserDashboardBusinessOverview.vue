<template>
  <section class="card overflow-hidden">
    <div class="border-b border-gray-100 px-6 py-5 dark:border-dark-700">
      <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
        <div>
          <p class="text-xs font-semibold uppercase tracking-[0.18em] text-gray-500 dark:text-dark-400">
            {{ copy.eyebrow }}
          </p>
          <h2 class="mt-2 text-xl font-bold text-gray-900 dark:text-white">
            {{ copy.title }}
          </h2>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
            {{ copy.description }}
          </p>
        </div>
        <button
          type="button"
          class="inline-flex h-10 items-center justify-center gap-2 rounded-xl border border-gray-200 px-4 text-sm font-semibold text-gray-700 transition hover:border-primary-200 hover:bg-primary-50 hover:text-primary-700 dark:border-dark-700 dark:text-dark-300 dark:hover:border-primary-800 dark:hover:bg-primary-950/30 dark:hover:text-primary-300"
          :disabled="loading"
          @click="loadBusinessData"
        >
          <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
          {{ copy.refresh }}
        </button>
      </div>
    </div>

    <div class="grid gap-4 p-4 lg:grid-cols-3">
      <RouterLink
        v-for="capability in capabilityCards"
        :key="capability.key"
        :to="capability.path"
        class="group rounded-2xl border border-gray-100 bg-gray-50/70 p-5 transition hover:-translate-y-0.5 hover:border-primary-200 hover:bg-white hover:shadow-lg hover:shadow-primary-100/60 dark:border-dark-700 dark:bg-dark-800/45 dark:hover:border-primary-800 dark:hover:bg-dark-800 dark:hover:shadow-black/20"
      >
        <div class="flex items-start justify-between gap-4">
          <div
            class="flex h-12 w-12 items-center justify-center rounded-xl transition group-hover:scale-105"
            :class="capability.iconClass"
          >
            <Icon :name="capability.icon" size="lg" />
          </div>
          <Icon name="arrowRight" size="sm" class="mt-1 text-gray-400 transition group-hover:translate-x-0.5 group-hover:text-primary-500 dark:text-dark-500" />
        </div>
        <p class="mt-4 text-sm font-semibold text-gray-900 dark:text-white">{{ capability.title }}</p>
        <p class="mt-1 min-h-10 text-xs leading-5 text-gray-500 dark:text-dark-400">{{ capability.description }}</p>
        <div class="mt-4 flex items-end justify-between gap-3">
          <div>
            <p class="text-2xl font-black text-gray-950 dark:text-white">{{ capability.total }}</p>
            <p class="text-xs text-gray-500 dark:text-dark-400">{{ capability.totalLabel }}</p>
          </div>
          <span
            class="rounded-full px-2.5 py-1 text-xs font-semibold"
            :class="capability.badgeClass"
          >
            {{ capability.badge }}
          </span>
        </div>
      </RouterLink>
    </div>

    <div class="grid gap-4 border-t border-gray-100 p-4 dark:border-dark-700 xl:grid-cols-2">
      <BusinessTaskList
        :title="copy.wechatRecent"
        :empty-text="copy.emptyWechat"
        :tasks="wechatRecentRows"
        :view-all-path="paths.tasks"
        :view-all-label="copy.viewAll"
      />
      <BusinessTaskList
        :title="copy.imageRecent"
        :empty-text="copy.emptyImage"
        :tasks="imageRecentRows"
        :view-all-path="paths.image"
        :view-all-label="copy.openImageWorkspace"
      />
    </div>

    <div
      v-if="errorMessage"
      class="border-t border-amber-100 bg-amber-50 px-6 py-3 text-sm text-amber-700 dark:border-amber-900/40 dark:bg-amber-950/20 dark:text-amber-300"
    >
      {{ errorMessage }}
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { listImageWorkspaceTasks, type ImageWorkspaceTask } from '@/api/image-workspace'
import { listWeChatExportTasks, type WeChatExportTask } from '@/api/wechat-export'

type BusinessTaskRow = {
  id: number
  title: string
  subtitle: string
  status: string
  statusClass: string
  createdAt: string
}

const paths = {
  wechat: '/app/wechat',
  image: '/app/image-generator',
  tasks: '/app/tasks',
}

const { locale } = useI18n()
const loading = ref(false)
const errorMessage = ref('')
const wechatTasks = ref<WeChatExportTask[]>([])
const imageTasks = ref<ImageWorkspaceTask[]>([])

const copy = computed(() => {
  const zh = String(locale.value || '').toLowerCase().startsWith('zh')
  return zh
    ? {
        eyebrow: 'Business workspace',
        title: '业务能力与任务',
        description: '集中查看微信导出、生图工作台和异步任务状态。',
        refresh: '刷新',
        wechatTitle: '微信导出',
        wechatDescription: '公众号文章导出、格式产物和 worker 处理状态。',
        imageTitle: '生图记录',
        imageDescription: '图片生成任务、产物数量和最近生成状态。',
        taskTitle: '统一任务',
        taskDescription: '跨业务查看微信导出和生图异步任务。',
        totalTasks: '任务',
        active: '处理中',
        completed: '已完成',
        failed: '需关注',
        wechatRecent: '最近微信导出',
        imageRecent: '最近生图记录',
        emptyWechat: '暂无微信导出任务。',
        emptyImage: '暂无生图任务。',
        viewAll: '查看全部',
        openImageWorkspace: '打开生图',
      }
    : {
        eyebrow: 'Business workspace',
        title: 'Business Capabilities & Tasks',
        description: 'Track WeChat exports, image generation, and async task flow in one place.',
        refresh: 'Refresh',
        wechatTitle: 'WeChat Export',
        wechatDescription: 'Article exports, generated formats, and worker task status.',
        imageTitle: 'Image Generation',
        imageDescription: 'Generation tasks, produced artifacts, and recent image status.',
        taskTitle: 'Unified Tasks',
        taskDescription: 'Review async tasks across WeChat export and image workspace.',
        totalTasks: 'tasks',
        active: 'active',
        completed: 'completed',
        failed: 'attention',
        wechatRecent: 'Recent WeChat exports',
        imageRecent: 'Recent image tasks',
        emptyWechat: 'No WeChat export tasks yet.',
        emptyImage: 'No image generation tasks yet.',
        viewAll: 'View all',
        openImageWorkspace: 'Open workspace',
      }
})

const wechatActiveCount = computed(() => countByStatuses(wechatTasks.value, ['queued', 'running', 'uploading']))
const wechatCompletedCount = computed(() => countByStatuses(wechatTasks.value, ['completed', 'completed_with_errors']))
const wechatFailedCount = computed(() => countByStatuses(wechatTasks.value, ['failed']))
const imageActiveCount = computed(() => countByStatuses(imageTasks.value, ['queued', 'running']))
const imageCompletedCount = computed(() => countByStatuses(imageTasks.value, ['succeeded']))
const imageFailedCount = computed(() => countByStatuses(imageTasks.value, ['failed']))

const capabilityCards = computed(() => [
  {
    key: 'wechat',
    path: paths.wechat,
    icon: 'document' as const,
    iconClass: 'bg-sky-100 text-sky-600 dark:bg-sky-950/40 dark:text-sky-300',
    badgeClass: wechatFailedCount.value > 0
      ? 'bg-amber-100 text-amber-700 dark:bg-amber-950/40 dark:text-amber-300'
      : 'bg-sky-100 text-sky-700 dark:bg-sky-950/40 dark:text-sky-300',
    title: copy.value.wechatTitle,
    description: copy.value.wechatDescription,
    total: String(wechatTasks.value.length),
    totalLabel: copy.value.totalTasks,
    badge: wechatActiveCount.value > 0 ? `${wechatActiveCount.value} ${copy.value.active}` : `${wechatCompletedCount.value} ${copy.value.completed}`,
  },
  {
    key: 'image',
    path: paths.image,
    icon: 'sparkles' as const,
    iconClass: 'bg-emerald-100 text-emerald-600 dark:bg-emerald-950/40 dark:text-emerald-300',
    badgeClass: imageFailedCount.value > 0
      ? 'bg-amber-100 text-amber-700 dark:bg-amber-950/40 dark:text-amber-300'
      : 'bg-emerald-100 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-300',
    title: copy.value.imageTitle,
    description: copy.value.imageDescription,
    total: String(imageTasks.value.length),
    totalLabel: copy.value.totalTasks,
    badge: imageActiveCount.value > 0 ? `${imageActiveCount.value} ${copy.value.active}` : `${imageCompletedCount.value} ${copy.value.completed}`,
  },
  {
    key: 'tasks',
    path: paths.tasks,
    icon: 'clipboard' as const,
    iconClass: 'bg-violet-100 text-violet-600 dark:bg-violet-950/40 dark:text-violet-300',
    badgeClass: wechatFailedCount.value + imageFailedCount.value > 0
      ? 'bg-amber-100 text-amber-700 dark:bg-amber-950/40 dark:text-amber-300'
      : 'bg-violet-100 text-violet-700 dark:bg-violet-950/40 dark:text-violet-300',
    title: copy.value.taskTitle,
    description: copy.value.taskDescription,
    total: String(wechatTasks.value.length + imageTasks.value.length),
    totalLabel: copy.value.totalTasks,
    badge: wechatFailedCount.value + imageFailedCount.value > 0
      ? `${wechatFailedCount.value + imageFailedCount.value} ${copy.value.failed}`
      : `${wechatActiveCount.value + imageActiveCount.value} ${copy.value.active}`,
  },
])

const wechatRecentRows = computed<BusinessTaskRow[]>(() =>
  wechatTasks.value.slice(0, 5).map((task) => ({
    id: task.id,
    title: `#${task.id} · ${task.formats.join(' + ').toUpperCase() || 'EXPORT'}`,
    subtitle: `${task.selected_article_count} article${task.selected_article_count === 1 ? '' : 's'} · ${task.successful_article_count}/${task.selected_article_count || 0}`,
    status: formatStatus(task.status),
    statusClass: statusClass(task.status),
    createdAt: formatDate(task.updated_at || task.created_at),
  })),
)

const imageRecentRows = computed<BusinessTaskRow[]>(() =>
  imageTasks.value.slice(0, 5).map((task) => ({
    id: task.id,
    title: `#${task.id} · ${task.model || 'image'}`,
    subtitle: `${task.size || '-'} · ${task.batch_size || 1} image${task.batch_size === 1 ? '' : 's'}`,
    status: formatStatus(task.status),
    statusClass: statusClass(task.status),
    createdAt: formatDate(task.updated_at || task.created_at),
  })),
)

async function loadBusinessData() {
  loading.value = true
  errorMessage.value = ''
  try {
    const [wechatResult, imageResult] = await Promise.allSettled([
      listWeChatExportTasks({ page: 1, page_size: 5 }),
      listImageWorkspaceTasks({ page: 1, page_size: 5 }),
    ])
    if (wechatResult.status === 'fulfilled') {
      wechatTasks.value = wechatResult.value.items || []
    }
    if (imageResult.status === 'fulfilled') {
      imageTasks.value = imageResult.value.items || []
    }
    if (wechatResult.status === 'rejected' || imageResult.status === 'rejected') {
      errorMessage.value = copy.value.description
    }
  } finally {
    loading.value = false
  }
}

function countByStatuses<T extends { status: string }>(items: T[], statuses: string[]) {
  return items.filter((item) => statuses.includes(item.status)).length
}

function formatStatus(status: string) {
  const normalized = status.replace(/_/g, ' ')
  return normalized ? normalized[0].toUpperCase() + normalized.slice(1) : '-'
}

function statusClass(status: string) {
  if (['completed', 'completed_with_errors', 'succeeded'].includes(status)) {
    return 'bg-emerald-50 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-300'
  }
  if (['queued', 'running', 'uploading'].includes(status)) {
    return 'bg-sky-50 text-sky-700 dark:bg-sky-950/40 dark:text-sky-300'
  }
  if (['failed', 'cancelled'].includes(status)) {
    return 'bg-rose-50 text-rose-700 dark:bg-rose-950/40 dark:text-rose-300'
  }
  return 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-dark-300'
}

function formatDate(value?: string) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '-'
  return new Intl.DateTimeFormat(locale.value, {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}

const BusinessTaskList = defineComponent({
  name: 'BusinessTaskList',
  props: {
    title: { type: String, required: true },
    emptyText: { type: String, required: true },
    viewAllPath: { type: String, required: true },
    viewAllLabel: { type: String, required: true },
    tasks: { type: Array as () => BusinessTaskRow[], required: true },
  },
  setup(props) {
    return () => h('div', { class: 'rounded-2xl border border-gray-100 bg-white p-4 dark:border-dark-700 dark:bg-dark-800/50' }, [
      h('div', { class: 'mb-3 flex items-center justify-between gap-3' }, [
        h('h3', { class: 'text-sm font-bold text-gray-900 dark:text-white' }, props.title),
        h(RouterLink, { to: props.viewAllPath, class: 'text-xs font-semibold text-primary-600 hover:text-primary-700 dark:text-primary-400' }, () => props.viewAllLabel),
      ]),
      props.tasks.length === 0
        ? h('p', { class: 'rounded-xl bg-gray-50 px-4 py-6 text-center text-sm text-gray-500 dark:bg-dark-900/40 dark:text-dark-400' }, props.emptyText)
        : h('div', { class: 'space-y-2' }, props.tasks.map((task) =>
          h('div', { key: task.id, class: 'flex items-center justify-between gap-3 rounded-xl bg-gray-50 px-3 py-3 dark:bg-dark-900/40' }, [
            h('div', { class: 'min-w-0' }, [
              h('p', { class: 'truncate text-sm font-semibold text-gray-900 dark:text-white' }, task.title),
              h('p', { class: 'mt-0.5 truncate text-xs text-gray-500 dark:text-dark-400' }, task.subtitle),
            ]),
            h('div', { class: 'shrink-0 text-right' }, [
              h('span', { class: ['inline-flex rounded-full px-2.5 py-1 text-xs font-semibold', task.statusClass] }, task.status),
              h('p', { class: 'mt-1 text-xs text-gray-400 dark:text-dark-500' }, task.createdAt),
            ]),
          ]),
        )),
    ])
  },
})

onMounted(() => {
  void loadBusinessData()
})
</script>
