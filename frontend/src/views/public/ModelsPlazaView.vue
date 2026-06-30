<template>
  <div class="home-business-page public-template-page min-h-screen">
    <div class="relative overflow-hidden border-b border-[var(--public-border)]">
      <div class="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_top_left,rgba(59,130,246,0.18),transparent_30%),radial-gradient(circle_at_top_right,rgba(16,185,129,0.16),transparent_26%),linear-gradient(180deg,rgba(255,255,255,0.02),transparent_45%)]"></div>

      <PublicDarkHeader :account-label="isAuthenticated ? copy.dashboard : copy.login">
        <template #actions>
          <DocsLink
            :doc-url="docUrl"
            class="hidden rounded-full border border-[var(--public-border)] px-4 py-2 text-sm font-medium text-[var(--public-body)] transition hover:border-[var(--public-border-strong)] hover:text-[var(--public-ink)] sm:inline-flex"
          >
            {{ copy.viewDocs }}
          </DocsLink>
        </template>
      </PublicDarkHeader>

      <section class="relative z-10 px-6 pb-16 pt-10 sm:pb-20 sm:pt-14">
        <div class="mx-auto max-w-5xl text-center">
          <div class="inline-flex items-center rounded-full border border-[var(--public-border)] bg-[var(--public-panel-soft)] px-4 py-2 text-sm font-medium text-[var(--public-body)] backdrop-blur">
            {{ copy.badge }}
          </div>
          <h1 class="mt-6 text-balance text-5xl font-black tracking-tight text-[var(--public-ink)] sm:text-6xl">
            {{ copy.title }}
          </h1>
          <p class="mx-auto mt-5 max-w-3xl text-balance text-base leading-8 text-[var(--public-body)] sm:text-lg">
            {{ copy.description }}
          </p>
        </div>
      </section>
    </div>

    <main class="public-template-main">
      <div class="public-template-container-wide">
        <div v-if="loading" class="flex justify-center py-20">
          <div class="h-10 w-10 animate-spin rounded-full border-2 border-[var(--public-border-strong)] border-t-white"></div>
        </div>

        <div
          v-else-if="items.length === 0"
          class="rounded-[32px] public-template-panel-muted px-8 py-16 text-center"
        >
          <h2 class="text-2xl font-bold text-[var(--public-ink)]">
            {{ copy.emptyTitle }}
          </h2>
          <p class="mx-auto mt-4 max-w-2xl text-sm leading-7 text-[var(--public-body)] sm:text-base">
            {{ copy.emptyDescription }}
          </p>
        </div>

        <div v-else class="grid gap-6 lg:grid-cols-[260px_minmax(0,1fr)] lg:gap-8">
          <aside class="lg:sticky lg:top-6 lg:self-start">
            <div class="rounded-3xl public-template-panel p-4 shadow-[0_20px_50px_rgba(0,0,0,0.22)] sm:p-5">
              <div>
                <p class="text-xs font-semibold uppercase tracking-[0.24em] text-[var(--public-faint)]">
                  {{ copy.quickFind }}
                </p>
                <label class="mt-3 block">
                  <span class="sr-only">{{ copy.searchLabel }}</span>
                  <input
                    v-model.trim="searchQuery"
                    type="search"
                    class="w-full rounded-2xl border public-template-input px-4 py-3 text-sm text-[var(--public-ink)] outline-none transition placeholder:text-[var(--public-faint)] focus:border-[var(--public-border-strong)] focus:bg-[var(--public-panel-soft)]"
                    :placeholder="copy.searchPlaceholder"
                  />
                </label>
              </div>

              <div class="mt-6">
                <p class="text-xs font-semibold uppercase tracking-[0.24em] text-[var(--public-faint)]">
                  {{ copy.groupsTitle }}
                </p>
                <div class="mt-3 flex gap-2 overflow-x-auto pb-1 lg:block lg:space-y-2 lg:overflow-visible lg:pb-0">
                  <button
                    v-for="group in groupOptions"
                    :key="group.key"
                    type="button"
                    class="flex min-w-max items-center justify-between gap-4 rounded-2xl px-4 py-3 text-left text-sm transition lg:w-full lg:min-w-0"
                    :class="
                      activeGroup === group.key
                        ? 'border border-[var(--public-border)] bg-white text-slate-950 shadow-[0_12px_30px_rgba(255,255,255,0.08)]'
                        : 'border border-[var(--public-border)] bg-[var(--public-panel-muted)] text-[var(--public-body)] hover:border-[var(--public-border)] hover:bg-[var(--public-panel-soft)] hover:text-[var(--public-ink)]'
                    "
                    @click="activeGroup = group.key"
                  >
                    <span class="font-medium">{{ group.label }}</span>
                    <span
                      class="rounded-full px-2 py-1 text-xs"
                      :class="activeGroup === group.key ? 'bg-slate-950/10 text-slate-700' : 'bg-[var(--public-panel-soft)] text-[var(--public-muted)]'"
                    >
                      {{ group.count }}
                    </span>
                  </button>
                </div>
              </div>
            </div>
          </aside>

          <section class="min-w-0">
            <div class="mb-6 flex flex-col gap-3 rounded-3xl public-template-panel-muted px-5 py-4 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <p class="text-sm font-semibold text-[var(--public-ink)]">
                  {{ activeGroupLabel }}
                </p>
                <p class="mt-1 text-sm text-[var(--public-muted)]">
                  {{
                    searchQuery
                      ? formatModelsPlazaTemplate(copy.currentSearch, { query: searchQuery })
                      : copy.browseHint
                  }}
                </p>
              </div>
              <div class="inline-flex items-center rounded-full border public-template-input px-4 py-2 text-sm text-[var(--public-body)]">
                {{ copy.results }} · {{ filteredItems.length }}
              </div>
            </div>

            <div
              v-if="filteredItems.length === 0"
              class="rounded-[32px] public-template-panel-muted px-8 py-16 text-center"
            >
              <h2 class="text-2xl font-bold text-[var(--public-ink)]">
                {{ copy.emptyFilteredTitle }}
              </h2>
              <p class="mx-auto mt-4 max-w-2xl text-sm leading-7 text-[var(--public-body)] sm:text-base">
                {{ copy.emptyFilteredDescription }}
              </p>
            </div>

            <div v-else class="grid gap-6 xl:grid-cols-2">
              <article
                v-for="item in filteredItems"
                :key="item.id"
                class="rounded-3xl public-template-panel p-5 shadow-[0_20px_60px_rgba(0,0,0,0.28)] sm:p-7"
              >
                <div class="flex items-start justify-between gap-4">
                  <div class="flex min-w-0 items-start gap-4">
                    <div
                      class="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl text-lg font-black text-[var(--public-ink)] sm:h-14 sm:w-14 sm:text-xl"
                      :class="providerIconClass(item.provider)"
                    >
                      {{ providerInitial(item.provider) }}
                    </div>
                    <div class="min-w-0">
                      <div class="flex flex-wrap items-center gap-2">
                        <h2 class="max-w-full break-words text-2xl font-black leading-tight text-[var(--public-ink)] sm:text-[2rem]">
                          {{ item.title }}
                        </h2>
                        <span
                          v-if="item.badge"
                          class="rounded-full border border-[var(--public-border)] bg-[var(--public-panel-soft)] px-3 py-1 text-xs font-semibold uppercase tracking-[0.18em] text-[var(--public-body)]"
                        >
                          {{ item.badge }}
                        </span>
                      </div>
                      <p
                        v-if="item.description"
                        class="mt-3 max-w-2xl text-sm leading-7 text-[var(--public-body)] sm:text-base"
                      >
                        {{ item.description }}
                      </p>
                    </div>
                  </div>

                  <button
                    v-if="item.model_ids.length > 0"
                    type="button"
                    class="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-xl border border-[var(--public-border)] text-[var(--public-body)] transition hover:border-[var(--public-border-strong)] hover:text-[var(--public-ink)]"
                    :title="copy.copyModelIds"
                    @click="copyToClipboard(item.model_ids.join('\n'), copy.modelIdsCopied)"
                  >
                    <Icon name="copy" size="sm" />
                  </button>
                </div>

                <div class="mt-8 grid gap-2 text-sm text-[var(--public-body)] sm:grid-cols-2 sm:text-base">
                  <p v-if="item.input_price" class="rounded-2xl border border-[var(--public-border)] bg-[var(--public-panel-muted)] px-3 py-2">{{ copy.inputPrice }} {{ item.input_price }}</p>
                  <p v-if="item.output_price" class="rounded-2xl border border-[var(--public-border)] bg-[var(--public-panel-muted)] px-3 py-2">{{ copy.outputPrice }} {{ item.output_price }}</p>
                  <p v-if="item.cache_read_price" class="rounded-2xl border border-[var(--public-border)] bg-[var(--public-panel-muted)] px-3 py-2 text-[var(--public-body)]">{{ copy.cacheReadPrice }} {{ item.cache_read_price }}</p>
                  <p v-if="item.cache_write_price" class="rounded-2xl border border-[var(--public-border)] bg-[var(--public-panel-muted)] px-3 py-2 text-[var(--public-body)]">{{ copy.cacheWritePrice }} {{ item.cache_write_price }}</p>
                </div>

                <div class="mt-8 flex flex-wrap gap-2">
                  <span
                    v-for="tag in item.capability_tags"
                    :key="tag"
                    class="rounded-full border border-[var(--public-border)] bg-[var(--public-panel-soft)] px-3 py-1 text-xs font-medium text-[var(--public-body)]"
                  >
                    {{ tag }}
                  </span>
                </div>

                <div class="mt-8 flex items-center justify-between gap-3">
                  <span
                    v-if="item.billing_badge"
                    class="rounded-full bg-[var(--public-accent-soft)] px-4 py-2 text-sm font-semibold text-[var(--public-accent-strong)]"
                  >
                    {{ item.billing_badge }}
                  </span>
                  <div v-else></div>
                  <p v-if="item.model_ids.length > 0" class="text-xs text-[var(--public-faint)]">
                    {{ copy.modelIdsConfigured }} · {{ item.model_ids.length }}
                  </p>
                </div>
              </article>
            </div>
          </section>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import DocsLink from '@/components/common/DocsLink.vue'
import PublicDarkHeader from '@/components/layout/PublicDarkHeader.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAuthStore, useAppStore } from '@/stores'
import { useClipboard } from '@/composables/useClipboard'
import {
  MODEL_PLAZA_ALL_GROUP_KEY,
  formatModelsPlazaTemplate,
  resolveModelPlazaProviderIconClass,
  resolveModelPlazaProviderInitial,
  resolveModelsPlazaCopy,
} from '@/utils/modelPlazaDisplay'
import { resolveRuntimeLocale } from '@/utils/runtimeLocale'
import {
  filterModelsPlazaItems,
  resolveModelsPlazaActiveGroupLabel,
  resolveModelsPlazaGroupOptions,
  resolveVisibleModelsPlazaItems,
} from './modelsPlazaRuntime'

const { locale } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const loading = ref(false)
const searchQuery = ref('')
const activeGroup = ref(MODEL_PLAZA_ALL_GROUP_KEY)

const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const isAuthenticated = computed(() => authStore.isAuthenticated)

const copy = computed(() =>
  resolveModelsPlazaCopy(
    appStore.cachedPublicSettings?.model_plaza_shell_config,
    resolveRuntimeLocale(locale),
  ),
)

const items = computed(() =>
  resolveVisibleModelsPlazaItems(appStore.cachedPublicSettings?.model_plaza_items || []),
)

const groupOptions = computed(() => resolveModelsPlazaGroupOptions(items.value, copy.value))

const activeGroupLabel = computed(() =>
  resolveModelsPlazaActiveGroupLabel(groupOptions.value, activeGroup.value, copy.value),
)

const filteredItems = computed(() =>
  filterModelsPlazaItems(items.value, activeGroup.value, searchQuery.value),
)

function providerInitial(provider: string): string {
  return resolveModelPlazaProviderInitial(provider)
}

function providerIconClass(provider: string): string {
  return resolveModelPlazaProviderIconClass(provider)
}

onMounted(async () => {
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) {
    loading.value = true
    try {
      await appStore.fetchPublicSettings()
    } finally {
      loading.value = false
    }
  }
})
</script>
