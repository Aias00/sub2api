<template>
  <div class="home-business-page public-template-page" :class="props.appShell ? 'min-h-0' : 'min-h-screen'">
    <PublicDarkHeader v-if="!props.appShell" :account-label="isAuthenticated ? t('nav.dashboard') : t('common.login')" />

    <main :class="props.appShell ? 'py-0' : 'public-template-main'">
      <div class="public-template-container-wide overflow-x-hidden">
        <section>
          <div class="grid min-w-0 gap-5 xl:grid-cols-[minmax(0,1fr)_minmax(320px,460px)] xl:items-end">
            <div class="min-w-0">
              <p class="text-sm font-semibold uppercase tracking-[0.22em] text-[var(--public-muted)]">
                {{ t('taskList.breadcrumb') }}
              </p>
              <h1 class="mt-4 max-w-4xl text-4xl font-black leading-tight text-[var(--public-ink)] sm:text-5xl">
                {{ t('taskList.title') }}
              </h1>
              <p class="mt-4 max-w-3xl text-base leading-8 text-[var(--public-body)]">
                {{ t('taskList.subtitle') }}
              </p>
            </div>

            <div class="public-template-panel min-w-0 overflow-hidden p-3 sm:p-4 xl:justify-self-end">
              <div class="mb-3 flex min-w-0 items-center justify-between gap-3">
                <p class="shrink-0 text-sm font-bold text-[var(--public-accent-strong)]">{{ t('taskList.actions') }}</p>
                <p class="hidden truncate text-xs text-[var(--public-muted)] xl:block">{{ t('taskList.subtitle') }}</p>
              </div>
              <div class="grid min-w-0 gap-2 sm:grid-cols-2">
                <router-link :to="wechatPath" class="inline-flex h-12 min-w-0 items-center justify-center truncate rounded-xl border public-template-input px-4 text-sm font-bold text-[var(--public-ink)] transition hover:bg-[var(--public-panel-muted)]">
                  {{ t('taskList.newWechatTask') }}
                </router-link>
                <router-link :to="imageGeneratorPath" class="public-template-button-primary h-12 min-w-0 truncate px-4 text-sm font-black">
                  {{ t('taskList.newImageTask') }}
                </router-link>
              </div>
            </div>
          </div>
        </section>

        <section v-if="!authStore.isAuthenticated" class="mt-8 whitespace-normal rounded-2xl border border-[#ffefcf] bg-[#fff8e8] px-5 py-4 text-sm text-[#ab570a] [overflow-wrap:anywhere]">
          {{ t('taskList.loginRequired') }}
        </section>

        <section v-else class="mt-8 min-w-0 space-y-5">
          <div class="min-w-0 overflow-hidden rounded-2xl public-template-panel p-4 sm:p-5">
            <div class="grid min-w-0 gap-5 xl:grid-cols-[minmax(0,1fr)_auto] xl:items-end">
              <div class="min-w-0">
                <span class="mb-3 block text-xs font-semibold uppercase tracking-[0.16em] text-[var(--public-muted)]">{{ t('taskList.filters') }}</span>
                <div class="flex min-w-0 flex-wrap gap-2" role="tablist" :aria-label="t('taskList.filters')">
                  <button
                    v-for="option in filterOptions"
                    :key="option.value"
                    type="button"
                    class="inline-flex h-10 max-w-full items-center rounded-xl border px-3.5 text-left text-sm font-bold transition"
                    :class="filter === option.value ? 'border-cyan-300/45 bg-cyan-300/12 text-[var(--public-accent-strong)] shadow-sm' : 'border-[var(--public-border)] bg-[var(--public-panel-soft)] text-[var(--public-body)] hover:bg-[var(--public-panel-muted)]'"
                    @click="filter = option.value"
                  >
                    {{ option.label }}
                  </button>
                </div>
              </div>

              <button
                type="button"
                class="public-template-button-primary h-11 w-full px-5 text-sm font-black xl:w-[150px]"
                :disabled="loading"
                @click="loadTasks"
              >
                {{ loading ? t('taskList.refreshing') : t('taskList.refresh') }}
              </button>
            </div>

            <div class="mt-5 grid min-w-0 gap-3 sm:grid-cols-2 xl:grid-cols-4">
              <button
                v-for="summary in summaries"
                :key="summary.key"
                type="button"
                class="group min-w-0 rounded-2xl border border-[var(--public-border)] bg-[var(--public-panel-soft)] p-4 text-left transition hover:-translate-y-0.5 hover:border-cyan-300/35 hover:bg-[var(--public-panel-muted)]"
                @click="setSummaryFilter(summary.key)"
              >
                <div class="flex items-center justify-between gap-3">
                  <p class="text-xs font-semibold uppercase tracking-[0.14em] text-[var(--public-muted)]">{{ summary.label }}</p>
                  <span class="h-2 w-2 rounded-full bg-cyan-300 opacity-0 transition group-hover:opacity-100"></span>
                </div>
                <p class="mt-2 text-3xl font-black tracking-tight text-[var(--public-ink)]">{{ summary.value }}</p>
              </button>
            </div>
          </div>

          <div class="min-w-0">
            <div v-if="errorMessage" class="mb-5 rounded-2xl border border-red-300/20 bg-red-300/10 px-5 py-4 text-sm text-[var(--public-danger)]">
              {{ errorMessage }}
            </div>

            <div v-if="loading && tasks.length === 0" class="grid gap-3">
              <div v-for="i in 4" :key="i" class="h-28 animate-pulse rounded-2xl public-template-panel-muted"></div>
            </div>

            <div v-else-if="visibleTasks.length === 0" class="rounded-2xl border border-[var(--public-border)] bg-[var(--public-panel)] px-8 py-20 text-center shadow-[0_18px_50px_rgba(0,0,0,0.10)]">
              <p class="text-2xl font-bold text-[var(--public-ink)]">{{ t('taskList.emptyTitle') }}</p>
              <p class="mx-auto mt-3 max-w-xl text-sm leading-7 text-[var(--public-body)]">{{ t('taskList.emptyDescription') }}</p>
              <div class="mt-6 flex flex-wrap justify-center gap-3">
                <router-link :to="wechatPath" class="inline-flex h-11 items-center justify-center rounded-xl border public-template-input px-5 text-sm font-bold text-[var(--public-ink)] transition hover:bg-[var(--public-panel-muted)]">
                  {{ t('taskList.newWechatTask') }}
                </router-link>
                <router-link :to="imageGeneratorPath" class="public-template-button-primary h-11 px-5 text-sm font-black">
                  {{ t('taskList.newImageTask') }}
                </router-link>
              </div>
            </div>

            <div v-else class="min-w-0 overflow-hidden rounded-2xl public-template-panel shadow-[0_20px_60px_rgba(0,0,0,0.14)]">
              <div class="hidden grid-cols-[minmax(0,1.45fr)_120px_120px_160px_220px] gap-4 border-b border-[var(--public-border)] bg-[var(--public-panel-soft)] px-5 py-3 text-xs font-semibold uppercase tracking-[0.12em] text-[var(--public-muted)] lg:grid">
                <span>{{ t('taskList.task') }}</span>
                <span>{{ t('taskList.type') }}</span>
                <span>{{ t('taskList.status') }}</span>
                <span>{{ t('taskList.updated') }}</span>
                <span class="text-right">{{ t('taskList.actions') }}</span>
              </div>
              <article
                v-for="task in visibleTasks"
                :key="task.key"
                class="grid gap-4 border-b border-[var(--public-border)] bg-[var(--public-panel)] px-5 py-5 last:border-b-0 hover:bg-[var(--public-panel-soft)] lg:grid-cols-[minmax(0,1.45fr)_120px_120px_160px_220px] lg:items-center"
              >
                <div class="min-w-0">
                  <div class="flex min-w-0 items-center gap-3">
                    <span class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl border public-template-input font-mono text-xs text-[var(--public-body)]">
                      {{ task.type === 'wechat' ? 'WX' : 'AI' }}
                    </span>
                    <div class="min-w-0">
                      <h2 class="truncate text-sm font-semibold text-[var(--public-ink)]">{{ task.title }}</h2>
                      <p class="mt-1 truncate text-xs text-[var(--public-muted)]">{{ task.description }}</p>
                    </div>
                  </div>
                  <div class="mt-3 flex flex-wrap gap-2 font-mono text-xs text-[var(--public-muted)]">
                    <span class="rounded-full border public-template-input px-2.5 py-1">{{ t('taskList.created') }} {{ formatTime(task.createdAt) }}</span>
                    <span v-if="task.costLabel" class="rounded-full border public-template-input px-2.5 py-1">{{ task.costLabel }}</span>
                    <span v-if="task.errorMessage" class="max-w-full truncate rounded-full border border-red-300/20 bg-red-300/10 px-2.5 py-1 text-[var(--public-danger)]">{{ task.errorMessage }}</span>
                  </div>
                </div>

                <div class="text-sm text-[var(--public-body)]">
                  <span class="lg:hidden">{{ t('taskList.type') }}: </span>{{ task.typeLabel }}
                </div>
                <div>
                  <span class="inline-flex rounded-full px-2.5 py-1 text-xs font-semibold" :class="statusClass(task.canonicalStatus)">
                    {{ statusLabel(task.canonicalStatus) }}
                  </span>
                </div>
                <div class="font-mono text-xs text-[var(--public-muted)]">
                  <span class="lg:hidden">{{ t('taskList.updated') }}: </span>{{ formatTime(task.updatedAt) }}
                </div>
                <div class="flex flex-wrap justify-start gap-2 lg:justify-end">
                  <router-link :to="task.detailPath" class="rounded-xl border border-[var(--public-border)] px-3 py-2 text-xs font-semibold text-[var(--public-body)] transition hover:bg-[var(--public-panel-soft)]">
                    {{ t('taskList.open') }}
                  </router-link>
                  <button
                    v-if="task.canDownload"
                    type="button"
                    class="rounded-xl border border-cyan-300/20 bg-cyan-300/10 px-3 py-2 text-xs font-semibold text-[var(--public-accent-strong)] transition hover:bg-cyan-300/20 disabled:cursor-not-allowed disabled:opacity-50"
                    :disabled="busyTaskKey === task.key"
                    @click="downloadTask(task)"
                  >
                    {{ busyTaskKey === task.key ? t('taskList.downloading') : t('taskList.download') }}
                  </button>
                  <button
                    v-if="task.canRetry"
                    type="button"
                    class="rounded-xl border border-[var(--public-border)] px-3 py-2 text-xs font-semibold text-[var(--public-body)] transition hover:bg-[var(--public-panel-soft)] disabled:cursor-not-allowed disabled:opacity-50"
                    :disabled="busyTaskKey === task.key"
                    @click="retryTask(task)"
                  >
                    {{ t('taskList.retry') }}
                  </button>
                  <button
                    v-if="task.canCancel"
                    type="button"
                    class="rounded-xl border border-red-300/20 bg-red-300/10 px-3 py-2 text-xs font-semibold text-[var(--public-danger)] transition hover:bg-red-300/20 disabled:cursor-not-allowed disabled:opacity-50"
                    :disabled="busyTaskKey === task.key"
                    @click="cancelTask(task)"
                  >
                    {{ t('taskList.cancel') }}
                  </button>
                </div>
              </article>
            </div>
            <div v-if="hasMoreTasks" class="mt-4 flex flex-wrap justify-center gap-3">
              <button
                v-if="wechatPage < wechatPages"
                type="button"
                class="rounded-xl border border-[var(--public-border)] px-4 py-2 text-sm font-bold text-[var(--public-body)] transition hover:bg-[var(--public-panel-soft)] disabled:opacity-45"
                :disabled="loadingMoreType === 'wechat'"
                @click="loadMoreTasks('wechat')"
              >
                {{ loadingMoreType === 'wechat' ? t('taskList.loadingMore') : t('taskList.loadMoreWechat', { loaded: wechatLoaded, total: wechatTotal }) }}
              </button>
              <button
                v-if="imagePage < imagePages"
                type="button"
                class="rounded-xl border border-[var(--public-border)] px-4 py-2 text-sm font-bold text-[var(--public-body)] transition hover:bg-[var(--public-panel-soft)] disabled:opacity-45"
                :disabled="loadingMoreType === 'image'"
                @click="loadMoreTasks('image')"
              >
                {{ loadingMoreType === 'image' ? t('taskList.loadingMore') : t('taskList.loadMoreImage', { loaded: imageLoaded, total: imageTotal }) }}
              </button>
            </div>
          </div>
        </section>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import PublicDarkHeader from '@/components/layout/PublicDarkHeader.vue'
import {
  cancelImageWorkspaceTask,
  downloadImageWorkspaceArtifact,
  listImageWorkspaceTasks,
  retryImageWorkspaceTask,
  type ImageWorkspaceTask,
} from '@/api/image-workspace'
import {
  cancelWeChatExportTask,
  downloadWeChatExportTaskZip,
  listWeChatExportTasks,
  retryWeChatExportTask,
  type WeChatExportTask,
} from '@/api/wechat-export'
import { useAuthStore } from '@/stores'
import { formatCostFixed, formatDateTime } from '@/utils/format'

const props = withDefaults(defineProps<{
  appShell?: boolean
}>(), {
  appShell: false,
})

type TaskType = 'wechat' | 'image'
type TaskFilter = 'all' | TaskType | 'active' | 'done' | 'attention'
type CanonicalStatus = 'queued' | 'running' | 'succeeded' | 'partial' | 'failed' | 'cancelled'

type UnifiedTask = {
  key: string
  id: number
  type: TaskType
  typeLabel: string
  title: string
  description: string
  canonicalStatus: CanonicalStatus
  rawStatus: string
  createdAt: string
  updatedAt: string
  detailPath: string
  costLabel: string
  errorMessage: string
  canCancel: boolean
  canRetry: boolean
  canDownload: boolean
  artifactId?: number
}

const { t } = useI18n()
const authStore = useAuthStore()

const loading = ref(false)
const loadingMoreType = ref<TaskType | ''>('')
const errorMessage = ref('')
const tasks = ref<UnifiedTask[]>([])
const filter = ref<TaskFilter>('all')
const busyTaskKey = ref('')
const taskPageSize = 50
const wechatPage = ref(1)
const wechatPages = ref(0)
const wechatTotal = ref(0)
const imagePage = ref(1)
const imagePages = ref(0)
const imageTotal = ref(0)

const isAuthenticated = computed(() => authStore.isAuthenticated)
const wechatLoaded = computed(() => tasks.value.filter((task) => task.type === 'wechat').length)
const imageLoaded = computed(() => tasks.value.filter((task) => task.type === 'image').length)
const hasMoreTasks = computed(() => wechatPage.value < wechatPages.value || imagePage.value < imagePages.value)
const wechatPath = computed(() => props.appShell ? '/app/wechat' : '/wechat')
const imageGeneratorPath = computed(() => props.appShell ? '/app/image-generator' : '/image-generator')

const filterOptions = computed(() => [
  { value: 'all' as const, label: t('taskList.filterAll') },
  { value: 'active' as const, label: t('taskList.filterActive') },
  { value: 'done' as const, label: t('taskList.filterDone') },
  { value: 'attention' as const, label: t('taskList.filterAttention') },
  { value: 'wechat' as const, label: t('taskList.filterWechat') },
  { value: 'image' as const, label: t('taskList.filterImage') },
])

const visibleTasks = computed(() => tasks.value.filter((task) => {
  if (filter.value === 'all') return true
  if (filter.value === 'wechat' || filter.value === 'image') return task.type === filter.value
  if (filter.value === 'active') return task.canonicalStatus === 'queued' || task.canonicalStatus === 'running'
  if (filter.value === 'done') return task.canonicalStatus === 'succeeded' || task.canonicalStatus === 'partial'
  if (filter.value === 'attention') return task.canonicalStatus === 'failed' || task.canonicalStatus === 'cancelled' || task.canonicalStatus === 'partial'
  return true
}))

const summaries = computed(() => [
  { key: 'all', label: t('taskList.summaryAll'), value: tasks.value.length },
  { key: 'active', label: t('taskList.summaryActive'), value: tasks.value.filter((task) => task.canonicalStatus === 'queued' || task.canonicalStatus === 'running').length },
  { key: 'done', label: t('taskList.summaryDone'), value: tasks.value.filter((task) => task.canonicalStatus === 'succeeded' || task.canonicalStatus === 'partial').length },
  { key: 'attention', label: t('taskList.summaryAttention'), value: tasks.value.filter((task) => task.canonicalStatus === 'failed' || task.canonicalStatus === 'cancelled' || task.canonicalStatus === 'partial').length },
])

function setSummaryFilter(value: string) {
  if (value === 'all' || value === 'active' || value === 'done' || value === 'attention') {
    filter.value = value
  }
}

onMounted(() => {
  if (authStore.isAuthenticated) {
    loadTasks()
  }
})

async function loadTasks() {
  loading.value = true
  errorMessage.value = ''
  try {
    const [wechat, image] = await Promise.all([
      listWeChatExportTasks({ page: 1, page_size: taskPageSize }),
      listImageWorkspaceTasks({ page: 1, page_size: taskPageSize }),
    ])
    wechatPage.value = wechat.page || 1
    wechatPages.value = wechat.pages || 0
    wechatTotal.value = wechat.total || wechat.items.length
    imagePage.value = image.page || 1
    imagePages.value = image.pages || 0
    imageTotal.value = image.total || image.items.length
    tasks.value = [
      ...wechat.items.map(mapWeChatTask),
      ...image.items.map(mapImageTask),
    ].sort((a, b) => new Date(b.updatedAt || b.createdAt).getTime() - new Date(a.updatedAt || a.createdAt).getTime())
  } catch (error) {
    errorMessage.value = extractErrorMessage(error) || t('taskList.loadFailed')
  } finally {
    loading.value = false
  }
}

async function loadMoreTasks(type: TaskType) {
  if (loadingMoreType.value) return
  const nextPage = type === 'wechat' ? wechatPage.value + 1 : imagePage.value + 1
  const totalPages = type === 'wechat' ? wechatPages.value : imagePages.value
  if (nextPage > totalPages) return
  loadingMoreType.value = type
  errorMessage.value = ''
  try {
    const result = type === 'wechat'
      ? await listWeChatExportTasks({ page: nextPage, page_size: taskPageSize })
      : await listImageWorkspaceTasks({ page: nextPage, page_size: taskPageSize })
    if (type === 'wechat') {
      wechatPage.value = result.page || nextPage
      wechatPages.value = result.pages || wechatPages.value
      wechatTotal.value = result.total || wechatTotal.value
    } else {
      imagePage.value = result.page || nextPage
      imagePages.value = result.pages || imagePages.value
      imageTotal.value = result.total || imageTotal.value
    }
    const mapped = type === 'wechat'
      ? result.items.map((task) => mapWeChatTask(task as WeChatExportTask))
      : result.items.map((task) => mapImageTask(task as ImageWorkspaceTask))
    const existingKeys = new Set(tasks.value.map((task) => task.key))
    tasks.value = [
      ...tasks.value,
      ...mapped.filter((task) => !existingKeys.has(task.key)),
    ].sort((a, b) => new Date(b.updatedAt || b.createdAt).getTime() - new Date(a.updatedAt || a.createdAt).getTime())
  } catch (error) {
    errorMessage.value = extractErrorMessage(error) || t('taskList.loadFailed')
  } finally {
    loadingMoreType.value = ''
  }
}

function mapWeChatTask(task: WeChatExportTask): UnifiedTask {
  const canonicalStatus = mapWeChatStatus(task.status)
  const formatLabel = task.formats?.length ? task.formats.join(' + ').toUpperCase() : t('taskList.wechatFormatFallback')
  return {
    key: `wechat:${task.id}`,
    id: task.id,
    type: 'wechat',
    typeLabel: t('taskList.typeWechat'),
    title: t('taskList.wechatTitle', { id: task.id }),
    description: t('taskList.wechatDescription', {
      count: task.selected_article_count ?? task.article_ids?.length ?? 0,
      formats: formatLabel,
    }),
    canonicalStatus,
    rawStatus: task.status,
    createdAt: task.created_at,
    updatedAt: task.updated_at,
    detailPath: wechatPath.value,
    costLabel: '',
    errorMessage: task.error_message || '',
    canCancel: task.status === 'queued' || task.status === 'running',
    canRetry: ['failed', 'completed_with_errors', 'cancelled', 'completed'].includes(task.status),
    canDownload: task.status === 'completed' || task.status === 'completed_with_errors',
  }
}

function mapImageTask(task: ImageWorkspaceTask): UnifiedTask {
  const canonicalStatus = mapImageStatus(task.status)
  const firstArtifact = task.artifacts?.[0]
  return {
    key: `image:${task.id}`,
    id: task.id,
    type: 'image',
    typeLabel: t('taskList.typeImage'),
    title: t('taskList.imageTitle', { id: task.id }),
    description: t('taskList.imageDescription', {
      model: task.model || t('taskList.imageModelFallback'),
      size: task.size || t('taskList.imageSizeFallback'),
      count: task.batch_size || task.artifacts?.length || 1,
    }),
    canonicalStatus,
    rawStatus: task.status,
    createdAt: task.created_at,
    updatedAt: task.updated_at,
    detailPath: imageGeneratorPath.value,
    costLabel: task.cost_estimate > 0 ? t('taskList.cost', { cost: formatCostFixed(task.cost_estimate) }) : '',
    errorMessage: task.error_message || '',
    canCancel: task.status === 'queued' || (task.status === 'running' && isExpired(task.worker_lease_until)),
    canRetry: task.status === 'failed' || task.status === 'cancelled',
    canDownload: Boolean(firstArtifact?.id),
    artifactId: firstArtifact?.id,
  }
}

function mapWeChatStatus(status: string): CanonicalStatus {
  if (status === 'completed') return 'succeeded'
  if (status === 'completed_with_errors') return 'partial'
  if (status === 'failed') return 'failed'
  if (status === 'cancelled') return 'cancelled'
  if (status === 'running' || status === 'uploading') return 'running'
  return 'queued'
}

function mapImageStatus(status: string): CanonicalStatus {
  if (status === 'succeeded') return 'succeeded'
  if (status === 'failed') return 'failed'
  if (status === 'cancelled') return 'cancelled'
  if (status === 'running') return 'running'
  return 'queued'
}

function statusLabel(status: CanonicalStatus): string {
  const labels: Record<CanonicalStatus, string> = {
    queued: t('taskList.statusQueued'),
    running: t('taskList.statusRunning'),
    succeeded: t('taskList.statusSucceeded'),
    partial: t('taskList.statusPartial'),
    failed: t('taskList.statusFailed'),
    cancelled: t('taskList.statusCancelled'),
  }
  return labels[status]
}

function statusClass(status: CanonicalStatus): string {
  if (status === 'succeeded') return 'border border-[#d3e5ff] bg-[#f5faff] text-[#0070f3]'
  if (status === 'partial') return 'border border-[#ffefcf] bg-[#fff8e8] text-[#ab570a]'
  if (status === 'failed') return 'border border-[#f7d4d6] bg-[#fff5f5] text-[#c50000]'
  if (status === 'cancelled') return 'border border-[#ebebeb] bg-[#f5f5f5] text-[#4d4d4d]'
  if (status === 'running') return 'border border-[#d3e5ff] bg-[#f5faff] text-[#0070f3]'
  return 'border border-[#ebebeb] bg-white text-[#4d4d4d]'
}

async function cancelTask(task: UnifiedTask) {
  await withBusyTask(task, async () => {
    if (task.type === 'wechat') {
      await cancelWeChatExportTask(task.id)
    } else {
      await cancelImageWorkspaceTask(task.id)
    }
    await loadTasks()
  })
}

async function retryTask(task: UnifiedTask) {
  await withBusyTask(task, async () => {
    if (task.type === 'wechat') {
      await retryWeChatExportTask(task.id)
    } else {
      await retryImageWorkspaceTask(task.id)
    }
    await loadTasks()
  })
}

async function downloadTask(task: UnifiedTask) {
  await withBusyTask(task, async () => {
    const blob = task.type === 'wechat'
      ? await downloadWeChatExportTaskZip(task.id)
      : task.artifactId
        ? await downloadImageWorkspaceArtifact(task.artifactId)
        : null
    if (!blob) return
    saveBlob(blob, task.type === 'wechat' ? `wechat-export-task-${task.id}.zip` : `image-task-${task.id}.png`)
  })
}

async function withBusyTask(task: UnifiedTask, action: () => Promise<void>) {
  busyTaskKey.value = task.key
  errorMessage.value = ''
  try {
    await action()
  } catch (error) {
    errorMessage.value = extractErrorMessage(error) || t('taskList.actionFailed')
  } finally {
    busyTaskKey.value = ''
  }
}

function formatTime(value: string): string {
  return formatDateTime(value)
}

function isExpired(value?: string): boolean {
  if (!value) return true
  const time = new Date(value).getTime()
  return Number.isFinite(time) && time <= Date.now()
}

function saveBlob(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  link.remove()
  URL.revokeObjectURL(url)
}

function extractErrorMessage(error: unknown): string {
  const responseMessage = (error as { response?: { data?: { message?: unknown } } })?.response?.data?.message
  if (typeof responseMessage === 'string') return responseMessage
  const message = (error as { message?: unknown })?.message
  return typeof message === 'string' ? message : ''
}
</script>
