<template>
  <div class="min-h-screen bg-[#101114] text-white">
    <header class="border-b border-white/10 px-6 py-5">
      <nav class="mx-auto flex max-w-6xl items-center justify-between gap-4">
        <RouterLink :to="authRouteDefaults.homePath" class="flex min-w-0 items-center gap-3">
          <div v-if="siteLogo" class="h-9 w-9 shrink-0 overflow-hidden rounded-xl border border-white/10 bg-white/5">
            <img :src="siteLogo" alt="Logo" class="h-full w-full object-contain" />
          </div>
          <span class="truncate text-sm font-semibold text-white">{{ siteName }}</span>
        </RouterLink>

        <div class="flex items-center gap-3">
          <LocaleSwitcher />
          <RouterLink
            :to="isAuthenticated ? dashboardPath : loginPath"
            class="inline-flex items-center rounded-full border border-white/10 bg-white px-4 py-2 text-sm font-semibold text-slate-950 transition hover:bg-white/90"
          >
            {{ isAuthenticated ? 'Dashboard' : 'Log in' }}
          </RouterLink>
        </div>
      </nav>
    </header>

    <main class="px-6 py-10 sm:py-12">
      <div class="mx-auto max-w-6xl">
        <div class="flex flex-col gap-3 border-b border-white/10 pb-8">
          <p class="text-sm font-semibold uppercase tracking-[0.22em] text-cyan-200/70">WeChat export</p>
          <h1 class="max-w-4xl text-4xl font-black tracking-tight text-white sm:text-5xl">
            Export articles into HTML, Markdown, and JSON artifacts
          </h1>
          <p class="max-w-3xl text-sm leading-7 text-white/60 sm:text-base">
            Import public article links, create an export task, then let the local worker generate downloadable files.
          </p>
        </div>

        <div
          v-if="!isAuthenticated"
          class="mt-8 rounded-2xl border border-amber-300/25 bg-amber-300/10 p-5 text-sm leading-7 text-amber-50"
        >
          Log in before using WeChat export. The API stores sessions, tasks, and artifacts under your account.
        </div>

        <div class="mt-8 grid gap-6 xl:grid-cols-[minmax(0,1fr)_390px]">
          <section class="rounded-2xl border border-white/10 bg-[#17181d] p-5 sm:p-6">
            <div class="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
              <div>
                <h2 class="text-2xl font-black text-white">Article intake</h2>
                <p class="mt-2 text-sm leading-6 text-white/55">Paste one public WeChat article link per import.</p>
              </div>
              <button
                type="button"
                class="rounded-full border border-white/10 px-4 py-2 text-sm font-semibold text-white/70 transition hover:bg-white/[0.06] disabled:cursor-not-allowed disabled:opacity-45"
                :disabled="!isAuthenticated || loading"
                @click="refreshAll"
              >
                Refresh
              </button>
            </div>

            <form class="mt-6 flex flex-col gap-3 sm:flex-row" @submit.prevent="handleImport">
              <input
                v-model="articleLink"
                type="url"
                placeholder="https://mp.weixin.qq.com/s/..."
                class="min-h-12 flex-1 rounded-xl border border-white/10 bg-black/20 px-4 text-sm text-white outline-none transition placeholder:text-white/30 focus:border-cyan-200/50"
                :disabled="!isAuthenticated || importing"
              />
              <button
                type="submit"
                class="min-h-12 rounded-xl bg-cyan-200 px-5 text-sm font-black text-slate-950 transition hover:bg-cyan-100 disabled:cursor-not-allowed disabled:opacity-45"
                :disabled="!isAuthenticated || importing || !articleLink.trim()"
              >
                {{ importing ? 'Importing' : 'Import link' }}
              </button>
            </form>

            <div class="mt-6 rounded-2xl border border-white/10 bg-black/20">
              <div class="grid grid-cols-[44px_minmax(0,1fr)] border-b border-white/10 px-4 py-3 text-xs font-semibold uppercase tracking-[0.16em] text-white/35">
                <span></span>
                <span>Imported articles</span>
              </div>

              <div v-if="articles.length === 0" class="px-4 py-10 text-center text-sm text-white/45">
                No article links imported yet.
              </div>

              <label
                v-for="article in articles"
                :key="article.id"
                class="grid cursor-pointer grid-cols-[44px_minmax(0,1fr)] gap-2 border-b border-white/10 px-4 py-4 last:border-b-0 hover:bg-white/[0.03]"
              >
                <input
                  v-model="selectedArticleIds"
                  type="checkbox"
                  class="mt-1 h-4 w-4 accent-cyan-200"
                  :value="article.id"
                />
                <span class="min-w-0">
                  <span class="block truncate text-sm font-semibold text-white">{{ article.title || article.link }}</span>
                  <span class="mt-1 block truncate text-xs text-white/45">{{ article.link }}</span>
                </span>
              </label>
            </div>
          </section>

          <aside class="rounded-2xl border border-white/10 bg-white/[0.03] p-5 sm:p-6">
            <h2 class="text-2xl font-black text-white">Create task</h2>
            <p class="mt-2 text-sm leading-6 text-white/55">The Node worker will pick queued tasks and write artifacts back.</p>

            <div class="mt-6 space-y-3">
              <label
                v-for="format in availableFormats"
                :key="format"
                class="flex items-center justify-between rounded-xl border border-white/10 bg-black/20 px-4 py-3"
              >
                <span class="text-sm font-semibold uppercase text-white/75">{{ format }}</span>
                <input v-model="formats" type="checkbox" class="h-4 w-4 accent-cyan-200" :value="format" />
              </label>
            </div>

            <label class="mt-4 flex items-center justify-between rounded-xl border border-white/10 bg-black/20 px-4 py-3">
              <span class="text-sm font-semibold text-white/75">Engagement data</span>
              <input v-model="includeEngagement" type="checkbox" class="h-4 w-4 accent-cyan-200" />
            </label>

            <button
              type="button"
              class="mt-6 min-h-12 w-full rounded-xl bg-white px-5 text-sm font-black text-slate-950 transition hover:bg-white/90 disabled:cursor-not-allowed disabled:opacity-45"
              :disabled="!isAuthenticated || creating || selectedArticleIds.length === 0 || formats.length === 0"
              @click="handleCreateTask"
            >
              {{ creating ? 'Creating task' : `Create export task (${selectedArticleIds.length})` }}
            </button>

            <p v-if="message" class="mt-4 rounded-xl border border-white/10 bg-black/20 px-4 py-3 text-sm text-white/65">
              {{ message }}
            </p>
            <p v-if="errorMessage" class="mt-4 rounded-xl border border-red-300/20 bg-red-300/10 px-4 py-3 text-sm text-red-100">
              {{ errorMessage }}
            </p>
          </aside>
        </div>

        <section class="mt-6 rounded-2xl border border-white/10 bg-[#17181d] p-5 sm:p-6">
          <div class="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
            <div>
              <h2 class="text-2xl font-black text-white">Tasks and artifacts</h2>
              <p class="mt-2 text-sm leading-6 text-white/55">Completed tasks expose generated files through authorized downloads.</p>
            </div>
          </div>

          <div v-if="tasks.length === 0" class="mt-6 rounded-2xl border border-white/10 bg-black/20 px-4 py-10 text-center text-sm text-white/45">
            No export tasks yet.
          </div>

          <div v-else class="mt-6 space-y-3">
            <article v-for="task in tasks" :key="task.id" class="rounded-2xl border border-white/10 bg-black/20 p-4">
              <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <div>
                  <p class="text-sm font-black text-white">Task #{{ task.id }}</p>
                  <p class="mt-1 text-xs uppercase tracking-[0.16em] text-white/35">{{ task.status }}</p>
                </div>
                <div class="flex flex-wrap gap-2">
                  <span v-for="format in task.formats" :key="format" class="rounded-full bg-white/10 px-3 py-1 text-xs font-semibold uppercase text-white/60">
                    {{ format }}
                  </span>
                </div>
              </div>

              <p v-if="task.error_message" class="mt-3 text-sm text-red-100">{{ task.error_message }}</p>

              <div class="mt-4 flex flex-wrap gap-2">
                <button
                  type="button"
                  class="rounded-full border border-white/10 px-4 py-2 text-xs font-semibold text-white/70 transition hover:bg-white/[0.06]"
                  @click="loadArtifacts(task.id)"
                >
                  Load artifacts
                </button>
                <a
                  v-for="artifact in artifactsByTask[task.id] || []"
                  :key="artifact.id"
                  class="rounded-full border border-cyan-200/30 bg-cyan-200/10 px-4 py-2 text-xs font-semibold uppercase text-cyan-100 transition hover:bg-cyan-200/20"
                  :href="artifactDownloadPath(artifact.id)"
                  target="_blank"
                  rel="noreferrer"
                >
                  {{ artifact.format }} · {{ artifact.file_name }}
                </a>
              </div>
            </article>
          </div>
        </section>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import { buildApiUrl } from '@/api/client'
import {
  createWeChatExportTask,
  importWeChatArticleLink,
  listWeChatArticles,
  listWeChatExportArtifacts,
  listWeChatExportTasks,
  type WeChatArticle,
  type WeChatExportArtifact,
  type WeChatExportFormat,
  type WeChatExportTask,
} from '@/api/wechat-export'
import { useAuthStore, useAppStore } from '@/stores'
import { useAuthRouteDefaults } from '@/composables/useAuthRouteDefaults'

const authStore = useAuthStore()
const appStore = useAppStore()
const { authRouteDefaults, resolveHomePath } = useAuthRouteDefaults()

const availableFormats: WeChatExportFormat[] = ['html', 'markdown', 'json']

const articleLink = ref('')
const articles = ref<WeChatArticle[]>([])
const tasks = ref<WeChatExportTask[]>([])
const selectedArticleIds = ref<number[]>([])
const formats = ref<WeChatExportFormat[]>(['html', 'markdown', 'json'])
const includeEngagement = ref(false)
const artifactsByTask = ref<Record<number, WeChatExportArtifact[]>>({})
const loading = ref(false)
const importing = ref(false)
const creating = ref(false)
const message = ref('')
const errorMessage = ref('')

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName)
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => resolveHomePath(isAdmin.value))
const loginPath = computed(() => authRouteDefaults.value.loginPath)

function setError(error: unknown) {
  errorMessage.value = error instanceof Error ? error.message : 'Request failed'
}

async function refreshAll() {
  if (!isAuthenticated.value) return
  loading.value = true
  errorMessage.value = ''
  try {
    const [articleResult, taskResult] = await Promise.all([
      listWeChatArticles(),
      listWeChatExportTasks(),
    ])
    articles.value = articleResult.items
    tasks.value = taskResult.items
  } catch (error) {
    setError(error)
  } finally {
    loading.value = false
  }
}

async function handleImport() {
  importing.value = true
  message.value = ''
  errorMessage.value = ''
  try {
    const article = await importWeChatArticleLink(articleLink.value)
    articleLink.value = ''
    message.value = 'Article link imported.'
    await refreshAll()
    if (!selectedArticleIds.value.includes(article.id)) {
      selectedArticleIds.value.push(article.id)
    }
  } catch (error) {
    setError(error)
  } finally {
    importing.value = false
  }
}

async function handleCreateTask() {
  creating.value = true
  message.value = ''
  errorMessage.value = ''
  try {
    const task = await createWeChatExportTask({
      article_ids: selectedArticleIds.value,
      formats: formats.value,
      include_engagement: includeEngagement.value,
    })
    message.value = `Task #${task.id} queued. Start the worker to generate artifacts.`
    selectedArticleIds.value = []
    await refreshAll()
  } catch (error) {
    setError(error)
  } finally {
    creating.value = false
  }
}

async function loadArtifacts(taskId: number) {
  errorMessage.value = ''
  try {
    artifactsByTask.value = {
      ...artifactsByTask.value,
      [taskId]: await listWeChatExportArtifacts(taskId),
    }
  } catch (error) {
    setError(error)
  }
}

function artifactDownloadPath(artifactId: number) {
  return buildApiUrl(`/wechat/artifacts/${artifactId}/download`, appStore.cachedPublicSettings)
}

onMounted(() => {
  void refreshAll()
})
</script>
