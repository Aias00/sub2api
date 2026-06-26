<template>
  <div class="min-h-screen bg-[#101114] text-white">
    <header class="border-b border-white/10 bg-[#15171d] px-6 py-5">
      <nav class="mx-auto flex max-w-7xl items-center justify-between gap-4">
        <RouterLink :to="authRouteDefaults.homePath" class="flex min-w-0 items-center gap-3">
          <div v-if="siteLogo" class="h-9 w-9 shrink-0 overflow-hidden rounded-xl border border-white/10 bg-white/5">
            <img :src="siteLogo" alt="Logo" class="h-full w-full object-contain" />
          </div>
          <span class="truncate text-sm font-semibold text-white">{{ siteName }}</span>
        </RouterLink>

        <div class="flex items-center gap-3">
          <LocaleSwitcher />
          <RouterLink
            :to="catalogPath"
            class="rounded-full border border-white/10 px-4 py-2 text-sm font-semibold text-white/70 transition hover:bg-white/[0.06] hover:text-white"
          >
            {{ workspaceShell.catalogLabel }}
          </RouterLink>
        </div>
      </nav>
    </header>

    <main class="px-6 py-10 sm:py-14">
      <div class="mx-auto max-w-6xl">
        <section class="rounded-2xl border border-white/10 bg-white/[0.035] p-6 sm:p-8">
          <p class="text-sm font-semibold uppercase tracking-[0.22em] text-violet-200/75">
            {{ workspaceShell.eyebrow }}
          </p>
          <h1 class="mt-4 text-4xl font-black leading-tight text-white sm:text-5xl">
            {{ workspaceShell.title }}
          </h1>
          <p class="mt-4 max-w-3xl text-base leading-8 text-white/60">
            {{ workspaceShell.heroDescription || workspaceShell.workspaceDescription }}
          </p>
        </section>

        <section class="mt-8 grid gap-6 lg:grid-cols-[minmax(0,1.15fr)_360px]">
          <div class="rounded-2xl border border-white/10 bg-[#17181d] p-5 sm:p-6">
            <div
              v-if="draftTitle"
              class="mb-5 rounded-2xl border border-violet-300/20 bg-violet-300/10 px-4 py-3 text-sm text-violet-50"
            >
              <p class="font-bold">{{ formatWorkspaceShellTemplate(workspaceShell.draftImported, { title: draftTitle }) }}</p>
              <p class="mt-1 text-violet-50/65">
                {{ workspaceShell.draftImportedDescription }}
              </p>
            </div>

            <label class="block">
              <span class="text-sm font-bold text-white/75">{{ workspaceShell.promptLabel }}</span>
              <textarea
                v-model="prompt"
                class="mt-3 min-h-72 w-full resize-y rounded-2xl border border-white/10 bg-white/[0.045] px-4 py-4 text-sm leading-7 text-white outline-none transition placeholder:text-white/30 focus:border-violet-300/45 focus:bg-white/[0.065]"
                :placeholder="workspaceShell.promptPlaceholder"
              />
            </label>

            <div class="mt-3 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div class="text-xs" :class="isPromptTooLong ? 'text-red-200' : 'text-white/40'">
                {{ promptLength }} / {{ maxPromptLength }}
                <span v-if="isPromptTooLong" class="ml-2">
                  {{ workspaceShell.promptTooLong }}
                </span>
              </div>
              <div class="flex gap-2">
                <button
                  type="button"
                  class="rounded-xl border border-white/10 px-4 py-2 text-sm font-semibold text-white/70 transition hover:bg-white/[0.06] disabled:cursor-not-allowed disabled:opacity-40"
                  :disabled="!trimmedPrompt"
                  @click="clearPrompt"
                >
                  {{ workspaceShell.clearLabel }}
                </button>
                <button
                  type="button"
                  class="rounded-xl bg-violet-500 px-5 py-2 text-sm font-black text-white transition hover:bg-violet-400 disabled:cursor-not-allowed disabled:opacity-50"
                  :disabled="!trimmedPrompt || isPromptTooLong"
                  @click="copyPrompt"
                >
                  {{ workspaceShell.copyPromptLabel }}
                </button>
              </div>
            </div>
          </div>

          <aside class="rounded-2xl border border-white/10 bg-[#17181d] p-5 sm:p-6">
            <h2 class="text-xl font-black text-white">
              {{ workspaceShell.workspaceTitle }}
            </h2>
            <p class="mt-4 text-sm leading-7 text-white/58">
              {{ workspaceShell.workspaceDescription }}
            </p>
            <div class="mt-5 rounded-2xl border border-white/10 bg-white/[0.035] p-4 text-sm leading-7 text-white/62">
              {{ workspaceShell.workspaceStatus }}
            </div>
            <RouterLink
              :to="catalogPath"
              class="mt-5 inline-flex w-full items-center justify-center rounded-xl border border-white/10 px-4 py-3 text-sm font-bold text-white/70 transition hover:bg-white/[0.06]"
            >
              {{ workspaceShell.backToCatalogLabel }}
            </RouterLink>
          </aside>
        </section>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { getLocale } from '@/i18n'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import { useClipboard } from '@/composables/useClipboard'
import { useAuthRouteDefaults } from '@/composables/useAuthRouteDefaults'
import { useAppStore } from '@/stores'
import { clearImageGeneratorDraft, loadImageGeneratorDraft } from '@/utils/imageGeneratorDraft'
import {
  formatWorkspaceShellTemplate,
  resolveWorkspaceShellConfig,
  resolveWorkspaceShellDefaults,
  type WorkspaceShellCopy,
} from '@/utils/imageWorkspaceShell'
import { resolveRuntimeLanguage } from '@/utils/runtimeLocale'
import { applyImageGeneratorDraft, resolveImageGeneratorCatalogPath } from './imageGeneratorRuntime'

const appStore = useAppStore()
const { copyToClipboard } = useClipboard()
const { authRouteDefaults } = useAuthRouteDefaults()

const prompt = ref('')
const draftTitle = ref('')

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName)
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const promptLength = computed(() => prompt.value.trim().length)
const trimmedPrompt = computed(() => prompt.value.trim())
const isPromptTooLong = computed(() => promptLength.value > maxPromptLength.value)

const workspaceShell = computed<WorkspaceShellCopy>(() =>
  resolveWorkspaceShellConfig(
    appStore.cachedPublicSettings?.workspace_shell_config,
    resolveRuntimeLanguage(getLocale()),
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
  draftTitle.value = ''
  try {
    clearImageGeneratorDraft()
  } catch {
    // Ignore storage failures.
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
})
</script>
