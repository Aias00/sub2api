<template>
  <div class="home-business-page public-template-page min-h-screen">
    <PublicDarkHeader :account-label="t('imageWorkspace.goConsole')">
      <template #actions>
        <RouterLink
          v-if="catalogPath"
          :to="catalogPath"
          class="rounded-full border border-[var(--public-border)] px-4 py-2 text-sm font-semibold text-[var(--public-body)] transition hover:bg-[var(--public-panel-soft)] hover:text-[var(--public-ink)]"
        >
          {{ workspaceShell.catalogLabel }}
        </RouterLink>
      </template>
    </PublicDarkHeader>

    <main class="public-template-main">
      <div class="public-template-container">
        <section class="rounded-2xl public-template-panel-muted p-6 sm:p-8">
          <p class="text-sm font-semibold uppercase tracking-[0.22em] text-[var(--public-muted)]">
            {{ workspaceShell.eyebrow }}
          </p>
          <h1 class="mt-4 text-4xl font-black leading-tight text-[var(--public-ink)] sm:text-5xl">
            {{ workspaceShell.title }}
          </h1>
          <p v-if="workspaceShell.heroDescription" class="mt-4 max-w-3xl text-base leading-8 text-[var(--public-body)]">
            {{ workspaceShell.heroDescription }}
          </p>
        </section>

        <section class="mt-8">
          <div class="rounded-2xl public-template-panel p-5 sm:p-6">
            <label class="block">
              <span class="text-sm font-bold text-[var(--public-body)]">{{ workspaceShell.promptLabel }}</span>
              <textarea
                v-model="prompt"
                class="mt-3 min-h-72 w-full resize-y rounded-2xl border public-template-input px-4 py-4 text-sm leading-7 text-[var(--public-ink)] outline-none transition placeholder:text-[var(--public-faint)] focus:border-violet-300/45 focus:bg-[var(--public-panel-soft)]"
                :placeholder="workspaceShell.promptPlaceholder"
              />
            </label>

            <label class="mt-5 block">
              <span class="text-sm font-bold text-[var(--public-body)]">{{ t('imageWorkspace.negativePrompt') }}</span>
              <textarea
                v-model="negativePrompt"
                class="mt-3 min-h-24 w-full resize-y rounded-2xl public-template-panel-muted px-4 py-3 text-sm leading-7 text-[var(--public-ink)] outline-none transition placeholder:text-[var(--public-faint)] focus:border-violet-300/45 focus:bg-[var(--public-panel-soft)]"
                :placeholder="t('imageWorkspace.negativePromptPlaceholder')"
              />
            </label>

            <div class="mt-5 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
              <label class="block">
                <span class="text-xs font-bold uppercase tracking-[0.14em] text-[var(--public-muted)]">{{ t('imageWorkspace.model') }}</span>
                <select v-model="model" class="mt-2 w-full rounded-xl public-template-panel-muted px-3 py-2 text-sm text-[var(--public-ink)] outline-none">
                  <option v-for="item in enabledModelConfigs" :key="item.id" :value="item.id">
                    {{ item.label || item.id }}
                  </option>
                </select>
                <p v-if="selectedModelConfig?.cost_hint" class="mt-1 text-[11px] text-[var(--public-faint)]">
                  {{ selectedModelConfig.cost_hint }}
                </p>
              </label>
              <label class="block">
                <span class="text-xs font-bold uppercase tracking-[0.14em] text-[var(--public-muted)]">{{ t('imageWorkspace.size') }}</span>
                <select v-model="size" class="mt-2 w-full rounded-xl public-template-panel-muted px-3 py-2 text-sm text-[var(--public-ink)] outline-none">
                  <option v-for="item in selectedSizeOptions" :key="item" :value="item">{{ item }}</option>
                </select>
              </label>
              <label class="block">
                <span class="text-xs font-bold uppercase tracking-[0.14em] text-[var(--public-muted)]">{{ t('imageWorkspace.quality') }}</span>
                <select v-model="quality" class="mt-2 w-full rounded-xl public-template-panel-muted px-3 py-2 text-sm text-[var(--public-ink)] outline-none">
                  <option v-for="item in selectedQualityOptions" :key="item" :value="item">{{ item }}</option>
                </select>
              </label>
              <label class="block">
                <span class="text-xs font-bold uppercase tracking-[0.14em] text-[var(--public-muted)]">{{ t('imageWorkspace.batchSize') }}</span>
                <input v-model.number="batchSize" type="number" min="1" max="4" class="mt-2 w-full rounded-xl public-template-panel-muted px-3 py-2 text-sm text-[var(--public-ink)] outline-none" />
              </label>
            </div>

            <label class="mt-5 block">
              <span class="text-xs font-bold uppercase tracking-[0.14em] text-[var(--public-muted)]">{{ t('imageWorkspace.style') }}</span>
              <input v-model="style" class="mt-2 w-full rounded-xl public-template-panel-muted px-3 py-2 text-sm text-[var(--public-ink)] outline-none" :placeholder="t('imageWorkspace.stylePlaceholder')" />
            </label>

            <div v-if="promptSafetyWarning" class="mt-4 public-template-warning px-4 py-3 text-sm leading-6">
              {{ promptSafetyWarning }}
            </div>

            <div class="mt-3 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div class="text-xs" :class="isPromptTooLong ? 'text-[var(--public-danger)]' : 'text-[var(--public-muted)]'">
                {{ promptLength }} / {{ maxPromptLength }}
                <span v-if="isPromptTooLong" class="ml-2">
                  {{ workspaceShell.promptTooLong }}
                </span>
              </div>
              <div class="flex gap-2">
                <button
                  type="button"
                  class="rounded-xl border border-[var(--public-border)] px-4 py-2 text-sm font-semibold text-[var(--public-body)] transition hover:bg-[var(--public-panel-soft)] disabled:cursor-not-allowed disabled:opacity-40"
                  :disabled="!trimmedPrompt"
                  @click="clearPrompt"
                >
                  {{ workspaceShell.clearLabel }}
                </button>
                <button
                  type="button"
                  class="public-template-button px-5 py-2 text-sm font-black"
                  :disabled="!trimmedPrompt || isPromptTooLong"
                  @click="copyPrompt"
                >
                  {{ workspaceShell.copyPromptLabel }}
                </button>
                <button
                  type="button"
                  class="public-template-button-primary px-5 py-2 text-sm font-black"
                  :disabled="!canGenerate"
                  @click="createGenerationTask"
                >
                  {{ creatingTask ? t('imageWorkspace.queuing') : t('imageWorkspace.startGenerating') }}
                </button>
              </div>
            </div>

            <div v-if="!isAuthenticated" class="mt-4 public-template-warning p-4 text-sm leading-7">
              {{ t('imageWorkspace.loginRequired') }}
            </div>
            <div v-if="message" class="mt-4 rounded-2xl border p-4 text-sm public-template-success-message">
              {{ message }}
            </div>
            <div v-if="errorMessage" class="mt-4 rounded-2xl border border-red-200/20 bg-red-300/10 p-4 text-sm text-[var(--public-danger)]">
              {{ errorMessage }}
            </div>
            <RouterLink
              v-if="catalogPath"
              :to="catalogPath"
              class="mt-5 inline-flex items-center justify-center rounded-xl border border-[var(--public-border)] px-4 py-3 text-sm font-bold text-[var(--public-body)] transition hover:bg-[var(--public-panel-soft)]"
            >
              {{ workspaceShell.backToCatalogLabel }}
            </RouterLink>
          </div>
        </section>

        <section ref="taskListSection" class="mt-8">
          <div class="mb-4 flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
            <div>
              <p class="text-sm font-semibold uppercase tracking-[0.18em] text-[var(--public-muted)]">{{ t('imageWorkspace.generationHistory') }}</p>
              <h2 class="mt-2 text-2xl font-black text-[var(--public-ink)]">{{ t('imageWorkspace.imageTasks') }}</h2>
            </div>
            <div class="flex flex-wrap items-center gap-2">
              <p class="mr-2 text-sm text-[var(--public-muted)]">
                {{ t('imageWorkspace.estimatedCost') }}：{{ estimatedCost.toFixed(2) }}
              </p>
              <button
                type="button"
                class="rounded-xl border border-[var(--public-border)] px-4 py-2 text-sm font-bold text-[var(--public-body)] transition hover:bg-[var(--public-panel-soft)] disabled:opacity-40"
                :disabled="!isAuthenticated || loadingTasks"
                @click="refreshWorkspaceData"
              >
                {{ t('imageWorkspace.refresh') }}
              </button>
            </div>
          </div>

          <div class="flex flex-wrap items-center gap-2">
            <button
              v-for="filter in taskStatusFilters"
              :key="filter.value"
              type="button"
              class="rounded-full border px-3 py-1.5 text-xs font-black transition"
              :class="taskStatusFilter === filter.value ? 'border-cyan-200/40 bg-cyan-200/15 text-[var(--public-accent-strong)]' : 'border-[var(--public-border)] text-[var(--public-muted)] hover:bg-[var(--public-panel-soft)] hover:text-[var(--public-body)]'"
              @click="setTaskStatusFilter(filter.value)"
            >
              {{ filter.label }}
            </button>
            <span class="ml-2 text-xs text-[var(--public-faint)]">{{ t('imageWorkspace.totalTasks', { count: taskTotal }) }}</span>
          </div>

          <div class="mt-4 min-w-0">
              <div class="grid min-w-0 gap-4 sm:grid-cols-2 xl:grid-cols-3">
                <article
                  v-for="task in tasks"
                  :key="task.id"
                  class="min-w-0 overflow-hidden rounded-3xl border bg-[var(--public-panel-soft)] transition duration-300 hover:-translate-y-0.5 hover:border-cyan-200/30 hover:bg-[var(--public-panel-soft)]"
                  :class="task.status === 'running' ? 'border-cyan-200/25 shadow-[0_0_32px_rgba(103,232,249,0.08)]' : 'border-[var(--public-border)]'"
                >
                  <div v-if="task.artifacts?.length" class="space-y-2">
                        <div
                          v-for="artifact in task.artifacts"
                          :key="artifact.id"
                          class="group relative w-full max-w-full overflow-hidden bg-[var(--public-canvas)] shadow-black/20"
                          :style="artifactAspectStyle(artifact)"
                        >
                          <button
                            type="button"
                            class="relative block h-full w-full overflow-hidden bg-[radial-gradient(circle_at_30%_20%,rgba(125,211,252,0.16),transparent_32%),#08090c]"
                            @click="openLightbox(task, artifact)"
                          >
                            <div
                              v-if="artifactSrc(artifact) !== '#' && !isArtifactLoaded(artifact.id) && !isArtifactFailed(artifact.id)"
                              class="absolute inset-0 animate-pulse bg-gradient-to-br from-white/[0.08] via-white/[0.025] to-cyan-200/[0.08]"
                            ></div>
                            <img
                              v-if="artifactSrc(artifact) !== '#' && !isArtifactFailed(artifact.id)"
                              :src="artifactSrc(artifact)"
                              :alt="artifact.prompt || task.prompt"
                              loading="lazy"
                              class="h-full w-full object-contain transition duration-500 group-hover:scale-[1.025]"
                              :class="isArtifactLoaded(artifact.id) ? 'opacity-100' : 'opacity-0'"
                              @load="markArtifactLoaded(artifact.id)"
                              @error="markArtifactFailed(artifact.id)"
                            />
                            <div v-else class="flex h-full min-h-28 items-center justify-center text-xs font-bold text-[var(--public-faint)]">
                              {{ t('imageWorkspace.imageLoadFailed') }}
                            </div>
                          </button>
                          <button
                            type="button"
                            class="absolute bottom-2 right-2 inline-flex h-8 w-8 items-center justify-center rounded-full border border-[var(--public-border)] bg-black/50 text-[var(--public-body)] backdrop-blur transition hover:bg-black/70 hover:text-[var(--public-ink)] disabled:cursor-wait disabled:opacity-75"
                            :title="isArtifactDownloading(artifact.id) ? t('imageWorkspace.downloading') : t('imageWorkspace.download')"
                            :aria-busy="isArtifactDownloading(artifact.id)"
                            :disabled="isArtifactDownloading(artifact.id)"
                            @click.stop="downloadArtifact(artifact)"
                          >
                            <svg v-if="isArtifactDownloading(artifact.id)" xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 animate-spin" viewBox="0 0 24 24" fill="none">
                              <circle class="opacity-25" cx="12" cy="12" r="9" stroke="currentColor" stroke-width="3" />
                              <path class="opacity-80" fill="currentColor" d="M21 12a9 9 0 00-9-9v3a6 6 0 016 6h3z" />
                            </svg>
                            <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" viewBox="0 0 20 20" fill="currentColor"><path d="M3 17a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm3.293-7.707a1 1 0 011.414 0L9 10.586V3a1 1 0 112 0v7.586l1.293-1.293a1 1 0 111.414 1.414l-3 3a1 1 0 01-1.414 0l-3-3a1 1 0 010-1.414z" /></svg>
                          </button>
                        </div>
                  </div>
                  <div v-else class="flex min-h-36 items-center justify-center bg-[var(--public-canvas)] px-4 py-8 text-center text-xs font-bold text-[var(--public-faint)]">
                    {{ task.status === 'running' ? t('imageWorkspace.workerGenerating') : formatTaskStatus(task.status) }}
                  </div>
                  <div class="space-y-2 p-4">
                    <div class="flex items-center justify-between gap-2">
                      <p class="min-w-0 truncate text-sm font-black text-[var(--public-ink)]">
                        {{ t('imageWorkspace.taskId', { id: task.id }) }}
                        <span class="ml-1 text-xs font-medium text-[var(--public-faint)]">{{ formatRelativeTime(task.created_at) }}</span>
                      </p>
                      <span class="shrink-0 rounded-full px-2.5 py-1 text-xs font-bold uppercase" :class="taskStatusClass(task.status)">{{ formatTaskStatus(task.status) }}</span>
                    </div>
                    <p class="truncate text-xs text-[var(--public-muted)]">{{ task.model }} · {{ task.size }} · {{ task.quality }} · {{ t('imageWorkspace.batchLabel') }} {{ task.batch_size }}</p>
                    <p class="line-clamp-2 text-sm leading-6 text-[var(--public-body)]">{{ task.prompt }}</p>
                    <p v-if="task.cost_estimate" class="text-xs text-[var(--public-success)]">{{ t('imageWorkspace.cost') }} {{ task.cost_estimate.toFixed(2) }} · {{ t('imageWorkspace.balanceSnapshot') }} {{ task.balance_snapshot.toFixed(2) }}</p>
                    <p v-if="task.error_message" class="text-xs text-[var(--public-danger)]">{{ formatTaskError(task.error_message) }}</p>
                    <p v-else-if="task.status === 'running'" class="inline-flex items-center gap-2 text-xs text-[var(--public-accent-strong)]">
                      <span class="h-3 w-3 animate-spin rounded-full border-2 border-cyan-200/30 border-t-cyan-100"></span>
                      {{ t('imageWorkspace.workerGenerating') }}
                    </p>
                    <p v-else-if="task.worker_lease_until" class="text-xs text-[var(--public-accent-strong)]">{{ t('imageWorkspace.workerLeaseUntil') }} {{ formatTaskTime(task.worker_lease_until) }}</p>
                    <div class="flex flex-wrap items-center gap-2 pt-1">
                      <button
                        v-if="task.status === 'queued'"
                        type="button"
                        class="rounded-xl border border-red-200/20 bg-red-300/10 px-3 py-1 text-xs font-bold text-[var(--public-danger)] transition hover:bg-red-300/20 disabled:opacity-40"
                        :disabled="cancellingTaskId === task.id"
                        @click="cancelTask(task.id)"
                      >
                        {{ cancellingTaskId === task.id ? t('imageWorkspace.cancelling') : t('imageWorkspace.cancel') }}
                      </button>
                      <button
                        v-if="canRetryTask(task)"
                        type="button"
                        class="rounded-xl border border-amber-200/20 bg-amber-300/10 px-3 py-1 text-xs font-bold text-[var(--public-warning)] transition hover:bg-amber-300/20 disabled:opacity-40"
                        :disabled="retryingTaskId === task.id"
                        @click="retryTask(task.id)"
                      >
                        {{ retryingTaskId === task.id ? t('imageWorkspace.retrying') : t('imageWorkspace.retry') }}
                      </button>
                    </div>
                  </div>
                </article>
                <p v-if="isAuthenticated && tasks.length === 0" class="rounded-2xl public-template-panel-muted p-5 text-sm text-[var(--public-muted)]">
                  {{ t('imageWorkspace.noTasks') }}
                </p>
              </div>

              <!-- Page-based pagination -->
              <div
                v-if="taskTotalPages > 1"
                class="mt-6 flex flex-col items-center gap-3 sm:flex-row sm:justify-between"
              >
                <p class="text-xs text-[var(--public-muted)]">
                  {{ t('imageWorkspace.totalTasks', { count: taskTotal }) }} · {{ t('imageWorkspace.pageInfo', { current: taskPage, total: taskTotalPages }) }}
                </p>
                <nav class="flex items-center gap-1" :aria-label="t('imageWorkspace.paginationAriaLabel')">
                  <button
                    type="button"
                    class="inline-flex h-8 items-center justify-center rounded-lg public-template-panel-muted px-2 text-sm text-[var(--public-body)] transition hover:bg-[var(--public-panel-soft)] hover:text-[var(--public-ink)] disabled:cursor-not-allowed disabled:opacity-30"
                    :disabled="taskPage <= 1 || loadingTasks"
                    @click="goToTaskPage(taskPage - 1)"
                    :aria-label="t('imageWorkspace.prevPage')"
                  >
                    ‹
                  </button>
                  <template v-for="(pageNum, index) in visiblePages" :key="`${pageNum}-${index}`">
                    <span
                      v-if="typeof pageNum === 'string'"
                      class="inline-flex h-8 w-8 items-center justify-center text-xs text-[var(--public-faint)]"
                    >
                      ...
                    </span>
                    <button
                      v-else
                      type="button"
                      class="inline-flex h-8 w-8 items-center justify-center rounded-lg text-sm font-medium transition"
                      :class="
                        pageNum === taskPage
                          ? 'border border-cyan-200/40 bg-cyan-200/15 text-[var(--public-accent-strong)]'
                          : 'public-template-panel-muted text-[var(--public-body)] hover:bg-[var(--public-panel-soft)] hover:text-[var(--public-ink)]'
                      "
                      :disabled="loadingTasks"
                      @click="goToTaskPage(pageNum)"
                    >
                      {{ pageNum }}
                    </button>
                  </template>
                  <button
                    type="button"
                    class="inline-flex h-8 items-center justify-center rounded-lg public-template-panel-muted px-2 text-sm text-[var(--public-body)] transition hover:bg-[var(--public-panel-soft)] hover:text-[var(--public-ink)] disabled:cursor-not-allowed disabled:opacity-30"
                    :disabled="taskPage >= taskTotalPages || loadingTasks"
                    @click="goToTaskPage(taskPage + 1)"
                    :aria-label="t('imageWorkspace.nextPage')"
                  >
                    ›
                  </button>
                </nav>
              </div>
            </div>
        </section>
      </div>
    </main>

    <div
      v-if="lightboxArtifact"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/90 px-4 py-6 backdrop-blur-xl"
      @click.self="closeLightbox"
      @keydown="handleLightboxKeydown"
      tabindex="0"
    >
      <button
        type="button"
        class="absolute right-4 top-4 rounded-full border public-template-input px-4 py-2 text-sm font-black text-[var(--public-body)] transition hover:bg-white/[0.12]"
        @click="closeLightbox"
      >
        {{ t('imageWorkspace.close') }}
      </button>
      <button
        v-if="canNavigateLightbox"
        type="button"
        class="absolute left-4 top-1/2 -translate-y-1/2 rounded-full border public-template-input px-4 py-3 text-2xl text-[var(--public-body)] transition hover:bg-white/[0.12] sm:block"
        @click="showPreviousArtifact"
      >
        ‹
      </button>
      <figure class="max-h-full w-full max-w-6xl">
        <div class="flex max-h-[78vh] items-center justify-center overflow-hidden rounded-3xl border border-[var(--public-border)] bg-[var(--public-canvas)] shadow-2xl shadow-cyan-950/30">
          <img
            :src="artifactSrc(lightboxArtifact)"
            :alt="lightboxArtifact.prompt || lightboxTask?.prompt || 'image artifact'"
            class="max-h-[78vh] max-w-full object-contain"
          />
        </div>
        <figcaption class="mt-4 flex flex-col gap-3 rounded-2xl border public-template-input p-4 text-sm text-[var(--public-body)] sm:flex-row sm:items-center sm:justify-between">
          <div class="min-w-0">
            <p class="font-black text-[var(--public-ink)]">{{ t('imageWorkspace.taskId', { id: lightboxTask?.id }) }} · {{ t('imageWorkspace.imageId', { id: lightboxArtifact.id }) }}</p>
            <p class="mt-1 truncate text-xs text-[var(--public-muted)]">
              {{ lightboxArtifact.mime_type || 'image' }}
              <span v-if="lightboxArtifact.width && lightboxArtifact.height"> · {{ lightboxArtifact.width }}×{{ lightboxArtifact.height }}</span>
              <span v-if="lightboxArtifact.file_size"> · {{ formatFileSize(lightboxArtifact.file_size) }}</span>
            </p>
          </div>
          <button
            type="button"
            class="public-template-button-primary min-w-28 gap-2 px-4 py-2 text-sm font-black disabled:cursor-wait disabled:opacity-75"
            :aria-busy="isArtifactDownloading(lightboxArtifact.id)"
            :disabled="isArtifactDownloading(lightboxArtifact.id)"
            @click="downloadArtifact(lightboxArtifact!)"
          >
            <svg v-if="isArtifactDownloading(lightboxArtifact.id)" xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 animate-spin" viewBox="0 0 24 24" fill="none">
              <circle class="opacity-25" cx="12" cy="12" r="9" stroke="currentColor" stroke-width="3" />
              <path class="opacity-80" fill="currentColor" d="M21 12a9 9 0 00-9-9v3a6 6 0 016 6h3z" />
            </svg>
            {{ isArtifactDownloading(lightboxArtifact.id) ? t('imageWorkspace.downloading') : t('imageWorkspace.downloadOriginal') }}
          </button>
        </figcaption>
      </figure>
      <button
        v-if="canNavigateLightbox"
        type="button"
        class="absolute right-4 top-1/2 -translate-y-1/2 rounded-full border public-template-input px-4 py-3 text-2xl text-[var(--public-body)] transition hover:bg-white/[0.12] sm:block"
        @click="showNextArtifact"
      >
        ›
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { CSSProperties } from 'vue'
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { getLocale } from '@/i18n'
import PublicDarkHeader from '@/components/layout/PublicDarkHeader.vue'
import { useClipboard } from '@/composables/useClipboard'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores'
import {
  cancelImageWorkspaceTask,
  createImageWorkspaceTask,
  downloadImageWorkspaceArtifact,
  listImageWorkspaceModels,
  listImageWorkspaceTasks,
  retryImageWorkspaceTask,
  type ImageWorkspaceArtifact,
  type ImageWorkspaceModelOption,
  type ImageWorkspaceTask,
} from '@/api/image-workspace'
import { clearImageGeneratorDraft, loadImageGeneratorDraft } from '@/utils/imageGeneratorDraft'
import {
  resolveWorkspaceShellConfig,
  resolveWorkspaceShellDefaults,
  type WorkspaceShellCopy,
} from '@/utils/imageWorkspaceShell'
import { resolveRuntimeLanguage } from '@/utils/runtimeLocale'
import { applyImageGeneratorDraft, resolveImageGeneratorCatalogPath } from './imageGeneratorRuntime'

const appStore = useAppStore()
const authStore = useAuthStore()
const { t } = useI18n()
const { copyToClipboard } = useClipboard()

const prompt = ref('')
const negativePrompt = ref('')
const model = ref('gpt-image-2')
const size = ref('1024x1024')
const quality = ref('standard')
const style = ref('')
const batchSize = ref(1)
const draftTitle = ref('')
const tasks = ref<ImageWorkspaceTask[]>([])
const taskStatusFilter = ref('all')
const taskPage = ref(1)
const taskPageSize = 20
const taskTotalPages = ref(0)
const taskTotal = ref(0)
const taskListSection = ref<HTMLElement | null>(null)
const artifactLoadState = ref<Record<number, 'loaded' | 'failed'>>({})
const localArtifactBlobUrls = ref<Record<number, string>>({})
const downloadingArtifactIds = ref<Record<number, boolean>>({})
const lightboxTask = ref<ImageWorkspaceTask | null>(null)
const lightboxArtifact = ref<ImageWorkspaceArtifact | null>(null)
const modelConfigs = ref<ImageWorkspaceModelOption[]>([
  {
    id: 'gpt-image-2',
    label: 'GPT Image 2',
    provider: 'openai',
    default_size: '1024x1024',
    default_quality: 'standard',
    sizes: ['1024x1024', '1024x1536', '1536x1024'],
    qualities: ['standard', 'hd', 'high'],
    cost_per_image: 0.04,
    cost_hint: t('imageWorkspace.costPerImage', { cost: '0.04' }),
    enabled: true,
  },
  {
    id: 'gpt-image-1',
    label: 'GPT Image 1',
    provider: 'openai',
    default_size: '1024x1024',
    default_quality: 'standard',
    sizes: ['1024x1024', '1024x1536', '1536x1024'],
    qualities: ['standard', 'hd'],
    cost_per_image: 0.04,
    cost_hint: t('imageWorkspace.costPerImage', { cost: '0.04' }),
    enabled: true,
  },
  {
    id: 'gemini-3.1-flash-image',
    label: 'Gemini 3.1 Flash Image',
    provider: 'gemini',
    default_size: '1024x1024',
    default_quality: 'standard',
    sizes: ['1024x1024'],
    qualities: ['standard'],
    cost_per_image: 0.04,
    cost_hint: t('imageWorkspace.costPerImage', { cost: '0.04' }),
    enabled: true,
  },
])
const loadingTasks = ref(false)
const creatingTask = ref(false)
const cancellingTaskId = ref<number | null>(null)
const retryingTaskId = ref<number | null>(null)
const message = ref('')
const errorMessage = ref('')
let refreshTimer: number | undefined
let abortController: AbortController | undefined
let requestSequence = 0

const promptLength = computed(() => prompt.value.trim().length)
const trimmedPrompt = computed(() => prompt.value.trim())
const isPromptTooLong = computed(() => promptLength.value > maxPromptLength.value)
const isAuthenticated = computed(() => authStore.isAuthenticated)
const enabledModelConfigs = computed(() => {
  const enabled = modelConfigs.value.filter((item) => item.enabled !== false)
  return enabled.length > 0 ? enabled : modelConfigs.value
})
const selectedModelConfig = computed(() =>
  enabledModelConfigs.value.find((item) => item.id === model.value) || enabledModelConfigs.value[0],
)
const selectedSizeOptions = computed(() => selectedModelConfig.value?.sizes?.length ? selectedModelConfig.value.sizes : ['1024x1024'])
const selectedQualityOptions = computed(() => selectedModelConfig.value?.qualities?.length ? selectedModelConfig.value.qualities : ['standard'])
const estimatedCost = computed(() => Math.max(0, selectedModelConfig.value?.cost_per_image || 0) * Math.max(1, batchSize.value || 1))
const currentBalance = computed(() => Number(authStore.user?.balance || 0))
const hasInsufficientBalance = computed(() => isAuthenticated.value && estimatedCost.value > 0 && currentBalance.value < estimatedCost.value)
const taskStatusFilters = computed(() => [
  { value: 'all', label: t('imageWorkspace.filterAll') },
  { value: 'queued', label: t('imageWorkspace.statusQueued') },
  { value: 'running', label: t('imageWorkspace.statusRunning') },
  { value: 'succeeded', label: t('imageWorkspace.statusSucceeded') },
  { value: 'failed', label: t('imageWorkspace.statusFailed') },
  { value: 'cancelled', label: t('imageWorkspace.statusCancelled') },
])
const visiblePages = computed(() => {
  const pages: (number | string)[] = []
  const total = taskTotalPages.value
  const maxVisible = 7
  if (total <= maxVisible) {
    for (let i = 1; i <= total; i++) pages.push(i)
  } else {
    pages.push(1)
    const start = Math.max(2, taskPage.value - 2)
    const end = Math.min(total - 1, taskPage.value + 2)
    if (start > 2) pages.push('...')
    for (let i = start; i <= end; i++) pages.push(i)
    if (end < total - 1) pages.push('...')
    pages.push(total)
  }
  return pages
})
const lightboxArtifacts = computed(() => lightboxTask.value?.artifacts ?? [])
const lightboxArtifactIndex = computed(() => {
  if (!lightboxArtifact.value) return -1
  return lightboxArtifacts.value.findIndex((artifact) => artifact.id === lightboxArtifact.value?.id)
})
const canNavigateLightbox = computed(() => lightboxArtifacts.value.length > 1 && lightboxArtifactIndex.value >= 0)
const imagePromptFilter = computed(() => {
  const raw = appStore.cachedPublicSettings?.image_prompt_filter_config
  if (!raw?.trim()) return null
  try {
    const parsed = JSON.parse(raw)
    return parsed?.enabled ? parsed : null
  } catch {
    return null
  }
})
const promptSafetyWarning = computed(() => {
  const filter = imagePromptFilter.value
  if (!filter) return ''
  const text = `${prompt.value}\n${negativePrompt.value}\n${style.value}`.toLowerCase()
  const explicitKeywords = filter.explicit_keywords || []
  const youthContextKeywords = filter.youth_context_keywords || []
  const hasExplicit = explicitKeywords.some((kw: string) => wordMatch(text, kw.toLowerCase()))
  const hasYouthContext = youthContextKeywords.some((kw: string) => wordMatch(text, kw.toLowerCase()))
  if (hasExplicit && hasYouthContext) return filter.warning_message || ''
  if (hasExplicit && wordMatch(text, 'young')) return filter.youth_warning_message || ''
  return ''
})
const canGenerate = computed(() =>
  isAuthenticated.value &&
  Boolean(trimmedPrompt.value) &&
  !isPromptTooLong.value &&
  !promptSafetyWarning.value &&
  !creatingTask.value &&
  !hasInsufficientBalance.value,
)

const workspaceShell = computed<WorkspaceShellCopy>(() =>
  cleanWorkspaceShellCopy(
    resolveWorkspaceShellConfig(
      appStore.cachedPublicSettings?.workspace_shell_config,
      resolveRuntimeLanguage(getLocale()),
    ),
  ),
)
const workspaceShellDefaults = computed(() =>
  resolveWorkspaceShellDefaults(
    appStore.cachedPublicSettings?.workspace_shell_config,
    resolveRuntimeLanguage(getLocale()),
  ),
)
const catalogPath = computed(() => resolveImageGeneratorCatalogPath(workspaceShellDefaults.value.catalogPath))
const maxPromptLength = computed(() => workspaceShellDefaults.value.maxPromptLength)

function cleanWorkspaceShellCopy(copy: WorkspaceShellCopy): WorkspaceShellCopy {
  return {
    ...copy,
    workspaceDescription: copy.workspaceDescription
      .replace('、参数模板', '')
      .replace('参数模板、', '')
      .replace('参数模板', '')
      .replace('、余额预授权', '')
      .replace('余额预授权、', '')
      .replace('余额预授权', '')
      .replace('prompt templates, ', '')
      .replace(', prompt templates', '')
      .replace('prompt templates', '')
      .replace('balance pre-authorization, ', '')
      .replace(', balance pre-authorization', '')
      .replace('balance pre-authorization', ''),
  }
}

function readDraft() {
  try {
    const draft = loadImageGeneratorDraft()
    const resolved = applyImageGeneratorDraft(draft, maxPromptLength.value)
    prompt.value = resolved.prompt
    draftTitle.value = resolved.title
  } catch {
    // Ignore malformed drafts; users can still type a prompt.
  }
}

function clearPrompt() {
  prompt.value = ''
  negativePrompt.value = ''
  style.value = ''
  draftTitle.value = ''
  try {
    clearImageGeneratorDraft()
  } catch {
    // Ignore storage failures.
  }
}

function setError(error: unknown) {
  errorMessage.value = error instanceof Error ? formatTaskError(error.message) : t('imageWorkspace.errorRequestFailed')
}

function cancelPendingRefresh() {
  if (abortController) {
    abortController.abort()
    abortController = undefined
  }
}

async function refreshWorkspaceData() {
  if (!isAuthenticated.value) return
  cancelPendingRefresh()
  requestSequence++
  const currentSequence = requestSequence
  loadingTasks.value = true
  try {
    const taskResult = await listImageWorkspaceTasks(taskListParams(1))
    // Discard stale response if a newer request has been issued
    if (currentSequence !== requestSequence) return
    tasks.value = taskResult.items
    void loadLocalArtifactBlobs(taskResult.items)
    taskTotalPages.value = taskResult.pages || 0
    taskTotal.value = taskResult.total || 0
  } catch (error) {
    if (currentSequence !== requestSequence) return // Ignore stale error
    setError(error)
  } finally {
    if (currentSequence === requestSequence) {
      loadingTasks.value = false
    }
  }
}

async function goToTaskPage(page: number) {
  if (!isAuthenticated.value || loadingTasks.value) return
  if (page < 1 || page > taskTotalPages.value || page === taskPage.value) return
  cancelPendingRefresh()
  requestSequence++
  const currentSequence = requestSequence
  loadingTasks.value = true
  try {
    const taskResult = await listImageWorkspaceTasks(taskListParams(page))
    if (currentSequence !== requestSequence) return
    tasks.value = taskResult.items
    void loadLocalArtifactBlobs(taskResult.items)
    taskPage.value = page
    taskTotalPages.value = taskResult.pages || 0
    taskTotal.value = taskResult.total || 0
    taskListSection.value?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  } catch (error) {
    if (currentSequence !== requestSequence) return
    setError(error)
  } finally {
    if (currentSequence === requestSequence) {
      loadingTasks.value = false
    }
  }
}

async function reloadCurrentPage() {
  if (!isAuthenticated.value || loadingTasks.value) return
  cancelPendingRefresh()
  requestSequence++
  const currentSequence = requestSequence
  loadingTasks.value = true
  try {
    const taskResult = await listImageWorkspaceTasks(taskListParams(taskPage.value))
    if (currentSequence !== requestSequence) return
    tasks.value = taskResult.items
    void loadLocalArtifactBlobs(taskResult.items)
    taskTotalPages.value = taskResult.pages || 0
    taskTotal.value = taskResult.total || 0
  } catch (error) {
    if (currentSequence !== requestSequence) return
    setError(error)
  } finally {
    if (currentSequence === requestSequence) {
      loadingTasks.value = false
    }
  }
}

function taskListParams(page: number) {
  return {
    page,
    page_size: taskPageSize,
    ...(taskStatusFilter.value === 'all' ? {} : { status: taskStatusFilter.value }),
  }
}

function setTaskStatusFilter(value: string) {
  if (taskStatusFilter.value === value) return
  taskStatusFilter.value = value
  void refreshWorkspaceData()
}

async function loadModelConfigs() {
  if (!isAuthenticated.value) return
  try {
    const models = await listImageWorkspaceModels()
    if (models.length > 0) {
      modelConfigs.value = models
      ensureModelSelection()
    }
  } catch {
    // Keep built-in defaults when model config is unavailable.
  }
}

async function createGenerationTask() {
  if (!canGenerate.value) return
  creatingTask.value = true
  message.value = ''
  errorMessage.value = ''
  try {
    const task = await createImageWorkspaceTask({
      prompt: trimmedPrompt.value,
      negative_prompt: negativePrompt.value,
      model: model.value,
      provider: selectedModelConfig.value?.provider || 'openai',
      size: size.value,
      quality: quality.value,
      style: style.value,
      batch_size: batchSize.value,
    })
    message.value = t('imageWorkspace.taskQueued', { id: task.id })
    await authStore.refreshUser().catch(() => {
      // The task was created; stale balance is non-fatal and will refresh later.
    })
    await refreshWorkspaceData()
  } catch (error) {
    setError(error)
  } finally {
    creatingTask.value = false
  }
}

async function cancelTask(taskId: number) {
  cancellingTaskId.value = taskId
  errorMessage.value = ''
  try {
    await cancelImageWorkspaceTask(taskId)
    await refreshWorkspaceData()
  } catch (error) {
    setError(error)
  } finally {
    cancellingTaskId.value = null
  }
}

async function retryTask(taskId: number) {
  retryingTaskId.value = taskId
  errorMessage.value = ''
  try {
    await retryImageWorkspaceTask(taskId)
    message.value = t('imageWorkspace.taskRetried')
    await refreshWorkspaceData()
  } catch (error) {
    setError(error)
  } finally {
    retryingTaskId.value = null
  }
}

function ensureModelSelection() {
  const selected = selectedModelConfig.value
  if (!selected) return
  if (!enabledModelConfigs.value.some((item) => item.id === model.value)) {
    model.value = selected.id
  }
  if (!selectedSizeOptions.value.includes(size.value)) {
    size.value = selected.default_size || selectedSizeOptions.value[0] || '1024x1024'
  }
  if (!selectedQualityOptions.value.includes(quality.value)) {
    quality.value = selected.default_quality || selectedQualityOptions.value[0] || 'standard'
  }
}

function taskStatusClass(status: string) {
  if (status === 'succeeded') return 'bg-emerald-200/15 text-[var(--public-success)]'
  if (status === 'failed') return 'bg-red-200/15 text-[var(--public-danger)]'
  if (status === 'running') return 'bg-cyan-200/15 text-[var(--public-accent-strong)]'
  return 'bg-white/10 text-[var(--public-body)]'
}

function formatTaskStatus(status: string) {
  const labels: Record<string, string> = {
    queued: t('imageWorkspace.statusQueued'),
    running: t('imageWorkspace.statusRunning'),
    succeeded: t('imageWorkspace.statusSucceeded'),
    failed: t('imageWorkspace.statusFailed'),
    cancelled: t('imageWorkspace.statusCancelled'),
  }
  return labels[status] || status || ''
}

function formatTaskError(message: string) {
  if (!message) return ''
  if (message === 'upstream 404') {
    return t('imageWorkspace.errorUpstream404')
  }
  if (message.includes('upstream 404')) {
    return message.replace('upstream 404', t('imageWorkspace.errorUpstream404'))
  }
  if (message === 'IMAGE_WORKSPACE_UPSTREAM_API_KEY is required') {
    return t('imageWorkspace.errorUpstreamApiKeyMissing')
  }
  if (message.toLowerCase().includes('unauthorized') || message.includes('401') || message.includes('IMAGE_WORKSPACE_UPSTREAM_API_KEY')) {
    return t('imageWorkspace.errorUpstreamAuth')
  }
  return message
}

function canRetryTask(task: ImageWorkspaceTask) {
  if (task.status !== 'failed' && task.status !== 'cancelled') return false
  return !isNonRetryableImageFailure(task.error_message, task.result_json)
}

function isNonRetryableImageFailure(message = '', resultJSON = '') {
  const normalized = `${message} ${resultJSON}`.trim().toLowerCase()
  if (!normalized) return false
  return [
    'safety_error',
    'policy_violation',
    'content_policy',
    'content policy',
    'moderation',
    'moderation_blocked',
    'blocked by policy',
    'policy blocked',
    'safety system',
    'safety filter',
    'safety violation',
    'violates policy',
    'violated policy',
    'flagged',
    'responsible ai',
    'risk policy',
    '违规',
    '违反',
    '安全策略',
    '内容安全',
    '风控',
  ].some((marker) => normalized.includes(marker))
}

function formatTaskTime(value: string) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

/**
 * Match a keyword as a whole word (word-boundary) rather than a substring.
 * Prevents false positives like "vibrant" matching "bra" or "library" matching "bra".
 */
function wordMatch(text: string, term: string): boolean {
  const escaped = term.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  return new RegExp(`\\b${escaped}\\b`).test(text)
}

function formatRelativeTime(value: string) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  const diffSeconds = Math.round((date.getTime() - Date.now()) / 1000)
  const divisions = [
    { amount: 60, unit: 'second' },
    { amount: 60, unit: 'minute' },
    { amount: 24, unit: 'hour' },
    { amount: 7, unit: 'day' },
    { amount: 4.345, unit: 'week' },
    { amount: 12, unit: 'month' },
    { amount: Number.POSITIVE_INFINITY, unit: 'year' },
  ] as const
  let duration = diffSeconds
  for (const division of divisions) {
    if (Math.abs(duration) < division.amount) {
      return new Intl.RelativeTimeFormat(resolveRuntimeLanguage(getLocale()) === 'zh' ? 'zh-CN' : 'en', { numeric: 'auto' }).format(
        Math.round(duration),
        division.unit,
      )
    }
    duration /= division.amount
  }
  return ''
}

function artifactSrc(artifact: ImageWorkspaceArtifact) {
  if (artifact.id in localArtifactBlobUrls.value) {
    return localArtifactBlobUrls.value[artifact.id]
  }
  return artifact.image_url || '#'
}

async function loadLocalArtifactBlobs(items: ImageWorkspaceTask[]) {
  // Collect current artifact IDs to track which URLs should remain
  const currentArtifactIds = new Set<number>()
  items.flatMap((task) => (task.artifacts ?? [])).forEach((a) => currentArtifactIds.add(a.id))

  // Revoke blob URLs for artifacts no longer in the current page
  for (const [id, url] of Object.entries(localArtifactBlobUrls.value)) {
    if (!currentArtifactIds.has(Number(id))) {
      URL.revokeObjectURL(url)
      delete localArtifactBlobUrls.value[Number(id)]
    }
  }

  // Load new local artifacts that haven't been fetched yet
  const localArtifacts = items.flatMap((task) =>
    (task.artifacts ?? []).filter((a) => a.storage_provider === 'local' && !(a.id in localArtifactBlobUrls.value)),
  )
  await Promise.all(
    localArtifacts.map(async (artifact) => {
      try {
        const blob = await downloadImageWorkspaceArtifact(artifact.id)
        localArtifactBlobUrls.value[artifact.id] = URL.createObjectURL(blob)
      } catch {
        // Local artifact fetch failed; will fall back to broken image or error state
      }
    }),
  )
}

async function downloadArtifact(artifact: ImageWorkspaceArtifact) {
  if (isArtifactDownloading(artifact.id)) return
  downloadingArtifactIds.value = { ...downloadingArtifactIds.value, [artifact.id]: true }
  const filename = artifactDownloadName(null, artifact)
  try {
    const cachedBlobUrl = localArtifactBlobUrls.value[artifact.id]
    const blobUrl = cachedBlobUrl || URL.createObjectURL(await downloadImageWorkspaceArtifact(artifact.id))
    const anchor = document.createElement('a')
    anchor.href = blobUrl
    anchor.download = filename
    document.body.appendChild(anchor)
    anchor.click()
    document.body.removeChild(anchor)
    if (!cachedBlobUrl) {
      URL.revokeObjectURL(blobUrl)
    }
  } catch {
    const url = artifact.image_url || artifact.storage_key
    if (url) window.open(url, '_blank')
  } finally {
    const next = { ...downloadingArtifactIds.value }
    delete next[artifact.id]
    downloadingArtifactIds.value = next
  }
}

function isArtifactDownloading(id: number) {
  return Boolean(downloadingArtifactIds.value[id])
}

function artifactAspectStyle(artifact: ImageWorkspaceArtifact): CSSProperties {
  if (artifact.width > 0 && artifact.height > 0) {
    return { aspectRatio: `${artifact.width} / ${artifact.height}`, width: '100%', maxWidth: '100%' }
  }
  return { aspectRatio: '4 / 3', width: '100%', maxWidth: '100%' }
}

function formatFileSize(value: number) {
  if (!Number.isFinite(value) || value <= 0) return ''
  if (value < 1024 * 1024) return `${Math.max(1, Math.round(value / 1024))} KB`
  return `${(value / 1024 / 1024).toFixed(1)} MB`
}

function artifactDownloadName(task: ImageWorkspaceTask | null | undefined, artifact: ImageWorkspaceArtifact) {
  const extension = artifact.mime_type?.includes('png') ? 'png' : artifact.mime_type?.includes('webp') ? 'webp' : 'jpg'
  return `image-task-${task?.id ?? artifact.task_id}-${artifact.id}.${extension}`
}

function markArtifactLoaded(id: number) {
  artifactLoadState.value = { ...artifactLoadState.value, [id]: 'loaded' }
}

function markArtifactFailed(id: number) {
  artifactLoadState.value = { ...artifactLoadState.value, [id]: 'failed' }
}

function isArtifactLoaded(id: number) {
  return artifactLoadState.value[id] === 'loaded'
}

function isArtifactFailed(id: number) {
  return artifactLoadState.value[id] === 'failed'
}

function openLightbox(task: ImageWorkspaceTask, artifact: ImageWorkspaceArtifact) {
  if (artifactSrc(artifact) === '#' || isArtifactFailed(artifact.id)) return
  lightboxTask.value = task
  lightboxArtifact.value = artifact
}

function closeLightbox() {
  lightboxTask.value = null
  lightboxArtifact.value = null
}

function showPreviousArtifact() {
  if (!canNavigateLightbox.value) return
  const artifacts = lightboxArtifacts.value
  const nextIndex = (lightboxArtifactIndex.value - 1 + artifacts.length) % artifacts.length
  lightboxArtifact.value = artifacts[nextIndex]
}

function showNextArtifact() {
  if (!canNavigateLightbox.value) return
  const artifacts = lightboxArtifacts.value
  const nextIndex = (lightboxArtifactIndex.value + 1) % artifacts.length
  lightboxArtifact.value = artifacts[nextIndex]
}

function startTaskAutoRefresh() {
  if (refreshTimer) {
    window.clearInterval(refreshTimer)
  }
  refreshTimer = window.setInterval(() => {
    if (!isAuthenticated.value) return
    if (document.visibilityState !== 'visible') return // Pause when tab hidden
    if (!tasks.value.some((task) => task.status === 'queued' || task.status === 'running')) return
    void reloadCurrentPage()
  }, 5000)
}

function handleVisibilityChange() {
  if (document.visibilityState === 'visible') {
    // Resume refresh and immediately check for updates
    if (tasks.value.some((task) => task.status === 'queued' || task.status === 'running')) {
      void reloadCurrentPage()
    }
  }
}

function handleLightboxKeydown(event: KeyboardEvent) {
  if (!lightboxArtifact.value) return
  if (event.key === 'ArrowLeft') {
    showPreviousArtifact()
  } else if (event.key === 'ArrowRight') {
    showNextArtifact()
  } else if (event.key === 'Escape') {
    closeLightbox()
  }
}

function copyPrompt() {
  if (!trimmedPrompt.value) {
    appStore.showError(workspaceShell.value.copyEmptyError)
    return
  }
  void copyToClipboard(trimmedPrompt.value, workspaceShell.value.copySuccessMessage)
}

onMounted(async () => {
  if (!appStore.publicSettingsLoaded) {
    await appStore.fetchPublicSettings()
  }
  readDraft()
  await loadModelConfigs()
  await refreshWorkspaceData()
  startTaskAutoRefresh()
  document.addEventListener('visibilitychange', handleVisibilityChange)
})

watch(model, () => {
  ensureModelSelection()
})

onBeforeUnmount(() => {
  if (refreshTimer) {
    window.clearInterval(refreshTimer)
  }
  document.removeEventListener('visibilitychange', handleVisibilityChange)
  for (const url of Object.values(localArtifactBlobUrls.value)) {
    URL.revokeObjectURL(url)
  }
})
</script>
