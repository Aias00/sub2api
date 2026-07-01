<template>
  <AppLayout>
    <div class="space-y-5">
      <div class="overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900">
        <div class="flex flex-col gap-4 p-6 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <p class="text-sm font-semibold uppercase tracking-[0.2em] text-primary-600 dark:text-primary-300">
              Worker Control
            </p>
            <h1 class="mt-2 text-3xl font-black tracking-tight text-gray-950 dark:text-white sm:text-4xl">
              Worker 管理
            </h1>
            <p class="mt-2 max-w-3xl text-sm leading-6 text-gray-500 dark:text-dark-300">
              查看业务 Worker 与热点采集 Worker 的运行状态，执行重启、上线、下线，并获取部署命令。
            </p>
          </div>
          <button class="btn btn-secondary" type="button" :disabled="loading" @click="loadWorkers">
            <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            刷新
          </button>
        </div>
      </div>

      <div
        v-if="management && !management.enabled"
        class="rounded-2xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-800 dark:border-amber-900/60 dark:bg-amber-950/30 dark:text-amber-200"
      >
        {{ management.reason || 'Worker 管理动作未启用。' }}
      </div>

      <div v-if="error" class="rounded-2xl border border-red-200 bg-red-50 p-4 text-sm text-red-700 dark:border-red-900/60 dark:bg-red-950/30 dark:text-red-200">
        {{ error }}
      </div>

      <div class="grid gap-5 xl:grid-cols-2">
        <article
          v-for="node in workerNodes"
          :key="node.id"
          class="overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900"
        >
          <div class="border-b border-gray-100 p-5 dark:border-dark-700">
            <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
              <div>
                <p class="text-base font-black text-gray-950 dark:text-white">{{ nodeTitle(node.id) }}</p>
                <p class="mt-1 text-xs text-gray-500 dark:text-dark-300">
                  容器：{{ node.container_name || '-' }} · 状态：{{ node.container_state || 'unknown' }}
                </p>
              </div>
              <span :class="nodeHealthClass(node.health)">{{ workerHealthLabel(node.health) }}</span>
            </div>

            <div class="mt-4 flex flex-wrap gap-2">
              <button class="btn btn-secondary btn-sm" type="button" @click="copyDeployCommand(node)">
                <Icon name="copy" size="sm" />
                部署命令
              </button>
              <button class="btn btn-secondary btn-sm" type="button" :disabled="!canManage(node) || actionLoading === `${node.id}:restart`" @click="runAction(node, 'restart')">
                <Icon name="refresh" size="sm" :class="actionLoading === `${node.id}:restart` ? 'animate-spin' : ''" />
                重启
              </button>
              <button class="btn btn-secondary btn-sm" type="button" :disabled="!canManage(node) || actionLoading === `${node.id}:start`" @click="runAction(node, 'start')">
                <Icon name="play" size="sm" />
                上线
              </button>
              <button class="btn btn-danger btn-sm" type="button" :disabled="!canManage(node) || actionLoading === `${node.id}:stop`" @click="runAction(node, 'stop')">
                下线
              </button>
            </div>

            <p v-if="!canManage(node)" class="mt-3 text-xs text-amber-700 dark:text-amber-300">
              {{ node.management_reason || management?.reason || '该节点暂不可由页面管理。' }}
            </p>
          </div>

          <div class="divide-y divide-gray-100 dark:divide-dark-700">
            <section v-for="worker in node.workers" :key="worker.id" class="p-5">
              <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <div>
                  <p class="text-sm font-bold text-gray-950 dark:text-white">{{ worker.name }}</p>
                  <p class="mt-1 text-xs text-gray-500 dark:text-dark-300">{{ worker.message || workerStatusMessage(worker.health) }}</p>
                </div>
                <span :class="workerHealthClass(worker.health)">{{ workerHealthLabel(worker.health) }}</span>
              </div>

              <dl class="mt-4 grid grid-cols-4 gap-2 text-center">
                <div class="rounded-xl bg-gray-50 p-2 dark:bg-dark-800">
                  <dt class="text-[11px] text-gray-500 dark:text-dark-300">队列</dt>
                  <dd class="mt-1 text-base font-black text-gray-950 dark:text-white">{{ formatNumber(worker.queue) }}</dd>
                </div>
                <div class="rounded-xl bg-gray-50 p-2 dark:bg-dark-800">
                  <dt class="text-[11px] text-gray-500 dark:text-dark-300">运行</dt>
                  <dd class="mt-1 text-base font-black text-gray-950 dark:text-white">{{ formatNumber(worker.running) }}</dd>
                </div>
                <div class="rounded-xl bg-gray-50 p-2 dark:bg-dark-800">
                  <dt class="text-[11px] text-gray-500 dark:text-dark-300">失败</dt>
                  <dd class="mt-1 text-base font-black text-gray-950 dark:text-white">{{ formatNumber(worker.failed) }}</dd>
                </div>
                <div class="rounded-xl bg-gray-50 p-2 dark:bg-dark-800">
                  <dt class="text-[11px] text-gray-500 dark:text-dark-300">卡死</dt>
                  <dd class="mt-1 text-base font-black text-gray-950 dark:text-white">{{ formatNumber(worker.stale) }}</dd>
                </div>
              </dl>

              <div class="mt-3 space-y-1 text-xs text-gray-500 dark:text-dark-300">
                <p>总量：{{ formatNumber(worker.total) }} · 成功：{{ formatNumber(worker.succeeded) }} · 最后更新：{{ formatTime(worker.last_updated_at) }}</p>
                <p v-if="worker.status_path" class="truncate" :title="worker.status_path">状态文件：{{ worker.status_path }}</p>
                <p v-if="worker.attention_reasons?.length" class="text-amber-700 dark:text-amber-300">
                  注意：{{ worker.attention_reasons.join(', ') }}
                </p>
              </div>
            </section>
          </div>
        </article>
      </div>

      <div v-if="!loading && workerNodes.length === 0" class="rounded-2xl border border-dashed border-gray-200 p-8 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-dark-300">
        暂无 Worker 节点数据
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { adminAPI } from '@/api/admin'
import type { RuntimeWorkerAction, RuntimeWorkerStatus, RuntimeWorkersResponse } from '@/api/admin/settings'
import { useAppStore } from '@/stores'
import { extractApiErrorMessage } from '@/utils/apiError'

type WorkerNode = RuntimeWorkerStatus & {
  workers: RuntimeWorkerStatus[]
}

const appStore = useAppStore()
const loading = ref(false)
const error = ref('')
const workers = ref<RuntimeWorkerStatus[]>([])
const management = ref<RuntimeWorkersResponse['management']>()
const actionLoading = ref('')

const workerNodes = computed<WorkerNode[]>(() => {
  const byNode = new Map<string, WorkerNode>()
  for (const worker of workers.value) {
    const nodeID = worker.node_id || worker.id
    let node = byNode.get(nodeID)
    if (!node) {
      node = {
        ...worker,
        id: nodeID,
        name: nodeTitle(nodeID),
        health: worker.health,
        workers: [],
      }
      byNode.set(nodeID, node)
    }
    node.workers.push(worker)
    node.queue = (node.queue || 0) + (worker.queue || 0)
    node.running = (node.running || 0) + (worker.running || 0)
    node.failed = (node.failed || 0) + (worker.failed || 0)
    node.stale = (node.stale || 0) + (worker.stale || 0)
    node.total = (node.total || 0) + (worker.total || 0)
    node.succeeded = (node.succeeded || 0) + (worker.succeeded || 0)
    if (healthRank(worker.health) > healthRank(node.health)) {
      node.health = worker.health
    }
  }
  return [...byNode.values()]
})

async function loadWorkers() {
  loading.value = true
  error.value = ''
  try {
    const result = await adminAPI.settings.getRuntimeWorkers()
    workers.value = Array.isArray(result.workers) ? result.workers : []
    management.value = result.management
  } catch (err) {
    workers.value = []
    error.value = extractApiErrorMessage(err, 'Worker 状态加载失败')
  } finally {
    loading.value = false
  }
}

function canManage(node: WorkerNode): boolean {
  return node.manageable === true && management.value?.enabled === true
}

async function runAction(node: WorkerNode, action: RuntimeWorkerAction) {
  if (!canManage(node)) return
  const key = `${node.id}:${action}`
  actionLoading.value = key
  try {
    await adminAPI.settings.manageRuntimeWorker(node.id, action)
    appStore.showSuccess(`Worker ${actionLabel(action)} 已执行`)
    await loadWorkers()
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, `Worker ${actionLabel(action)} 失败`))
  } finally {
    actionLoading.value = ''
  }
}

async function copyDeployCommand(node: WorkerNode) {
  const command = node.deploy_command || node.workers.find(item => item.deploy_command)?.deploy_command || ''
  if (!command) {
    appStore.showError('没有可用部署命令')
    return
  }
  await navigator.clipboard.writeText(command)
  appStore.showSuccess('部署命令已复制')
}

function nodeTitle(id: string): string {
  switch (id) {
    case 'business-worker':
      return '业务 Worker 节点'
    case 'hot-collector':
      return '热点采集 Worker 节点'
    default:
      return id
  }
}

function actionLabel(action: RuntimeWorkerAction): string {
  switch (action) {
    case 'restart':
      return '重启'
    case 'start':
      return '上线'
    case 'stop':
      return '下线'
    case 'deploy':
      return '部署'
    default:
      return action
  }
}

function healthRank(health: string): number {
  switch ((health || '').toLowerCase()) {
    case 'attention':
      return 5
    case 'not_configured':
      return 4
    case 'waiting':
      return 3
    case 'active':
      return 2
    case 'idle':
      return 1
    default:
      return 0
  }
}

function workerHealthLabel(health: string): string {
  switch ((health || '').toLowerCase()) {
    case 'active':
      return '运行中'
    case 'idle':
      return '空闲'
    case 'waiting':
      return '等待'
    case 'attention':
      return '需处理'
    case 'not_configured':
      return '未配置'
    default:
      return '未知'
  }
}

function workerStatusMessage(health: string): string {
  switch ((health || '').toLowerCase()) {
    case 'active':
      return 'Worker 正在处理或最近有任务活动。'
    case 'idle':
      return 'Worker 当前无待处理任务。'
    case 'waiting':
      return '队列有任务等待 Worker 消费。'
    case 'attention':
      return '存在失败、卡死或配置异常，需要检查。'
    case 'not_configured':
      return '运行时配置不完整。'
    default:
      return '状态暂不可判断。'
  }
}

function workerHealthClass(health: string): string {
  const base = 'shrink-0 rounded-full px-2.5 py-1 text-xs font-bold'
  switch ((health || '').toLowerCase()) {
    case 'active':
      return `${base} bg-emerald-100 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300`
    case 'idle':
      return `${base} bg-sky-100 text-sky-700 dark:bg-sky-500/15 dark:text-sky-300`
    case 'waiting':
      return `${base} bg-amber-100 text-amber-700 dark:bg-amber-500/15 dark:text-amber-300`
    case 'attention':
      return `${base} bg-red-100 text-red-700 dark:bg-red-500/15 dark:text-red-300`
    case 'not_configured':
      return `${base} bg-gray-200 text-gray-700 dark:bg-dark-700 dark:text-dark-200`
    default:
      return `${base} bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-dark-300`
  }
}

function nodeHealthClass(health: string): string {
  return workerHealthClass(health)
}

function formatNumber(value: number | null | undefined): string {
  return typeof value === 'number' && Number.isFinite(value) ? String(value) : '-'
}

function formatTime(value: string | null | undefined): string {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

onMounted(() => {
  void loadWorkers()
})
</script>
