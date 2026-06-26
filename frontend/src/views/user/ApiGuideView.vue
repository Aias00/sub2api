<template>
  <AppLayout>
    <div class="space-y-6">
      <section class="page-hero">
        <div class="page-hero-grid">
          <div class="space-y-4">
            <div class="page-kicker">
              <Icon name="book" size="sm" />
              <span>{{ apiGuideText('badge') }}</span>
            </div>
            <div class="space-y-3">
              <h1 class="max-w-3xl text-3xl font-bold tracking-tight text-gray-900 dark:text-white">
                {{ apiGuideText('title') }}
              </h1>
              <p class="max-w-3xl text-sm leading-7 text-gray-600 dark:text-gray-300">
                {{ apiGuideText('description') }}
              </p>
            </div>

            <div class="flex flex-wrap items-center gap-3">
              <router-link :to="apiGuideDefaults.testPath" class="btn btn-primary">
                {{ apiGuideText('openTester') }}
              </router-link>
              <router-link :to="authRouteDefaults.apiKeysPath" class="btn btn-secondary">
                {{ apiGuideText('manageKeys') }}
              </router-link>
            </div>
          </div>

          <div class="grid gap-3 sm:grid-cols-3 xl:grid-cols-1">
            <div class="metric-panel">
              <div class="flex items-start gap-3">
                <div class="metric-icon">
                  <Icon name="server" size="sm" />
                </div>
                <div class="min-w-0">
                  <div class="text-xs font-semibold uppercase tracking-[0.18em] text-gray-500 dark:text-gray-400">
                    {{ apiGuideText('baseUrl') }}
                  </div>
                  <code class="mt-2 block break-all text-sm font-medium text-primary-700 dark:text-primary-200">
                    {{ gatewayBaseUrl }}
                  </code>
                </div>
              </div>
            </div>

            <div class="metric-panel">
              <div class="flex items-start gap-3">
                <div class="metric-icon">
                  <Icon name="key" size="sm" />
                </div>
                <div class="min-w-0">
                  <div class="text-xs font-semibold uppercase tracking-[0.18em] text-gray-500 dark:text-gray-400">
                    {{ apiGuideText('currentKey') }}
                  </div>
                  <p class="mt-2 text-sm font-semibold text-gray-900 dark:text-white">
                    {{ selectedKey?.name || apiGuideText('noSelection') }}
                  </p>
                  <p class="mt-1 truncate text-xs text-gray-500 dark:text-gray-400">
                    {{ selectedKeyMasked }}
                  </p>
                </div>
              </div>
            </div>

            <div class="metric-panel">
              <div class="flex items-start gap-3">
                <div class="metric-icon">
                  <Icon name="grid" size="sm" />
                </div>
                <div class="min-w-0">
                  <div class="text-xs font-semibold uppercase tracking-[0.18em] text-gray-500 dark:text-gray-400">
                    {{ apiGuideText('supportedEndpoints') }}
                  </div>
                  <p class="mt-2 text-2xl font-bold tabular-nums text-gray-900 dark:text-white">
                    {{ variants.length }}
                  </p>
                  <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                    {{ selectedKey?.group ? platformLabel(selectedKey.group.platform) : apiGuideText('noGroupAssigned') }}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <div v-if="loading" class="card flex items-center justify-center px-6 py-16">
        <LoadingSpinner />
      </div>

      <EmptyState
        v-else-if="keys.length === 0"
        :title="apiGuideText('noKeysTitle')"
        :description="apiGuideText('noKeysDescription')"
        :action-text="apiGuideText('manageKeys')"
        :action-to="authRouteDefaults.apiKeysPath"
      >
        <template #icon>
          <Icon name="key" size="xl" class="text-gray-400" />
        </template>
      </EmptyState>

      <div v-else class="grid gap-6 xl:grid-cols-[340px,minmax(0,1fr)]">
        <aside class="sticky-panel space-y-4 self-start">
          <div class="surface-panel-strong space-y-5">
            <div class="space-y-2">
              <div class="text-xs font-semibold uppercase tracking-[0.18em] text-primary-700 dark:text-primary-200">
                {{ apiGuideText('keySelector') }}
              </div>
              <p class="text-sm leading-6 text-gray-600 dark:text-gray-300">
                {{ apiGuideText('keySelectorHint') }}
              </p>
            </div>
            <label class="input-label mb-1.5 block">{{ apiGuideText('keySelector') }}</label>
            <Select
              v-model="selectedKeyId"
              :options="keyOptions"
              :placeholder="apiGuideText('keySelector')"
              searchable
            />
          </div>

          <div
            v-if="selectedKey && !selectedKey.group"
            class="surface-panel border-amber-200/80 bg-amber-50/90 text-sm text-amber-800 dark:border-amber-500/25 dark:bg-amber-500/10 dark:text-amber-100"
          >
            <div class="mb-1 font-semibold">{{ apiGuideText('unassignedTitle') }}</div>
            <p class="text-xs leading-5">{{ apiGuideText('unassignedDescription') }}</p>
          </div>

          <div
            v-else-if="selectedKey"
            class="surface-panel space-y-4"
          >
            <div class="text-sm font-semibold text-gray-800 dark:text-gray-100">
              {{ apiGuideText('keySummary') }}
            </div>
            <dl class="space-y-3 text-sm">
              <div>
                <dt class="text-gray-500 dark:text-gray-400">{{ apiGuideText('groupName') }}</dt>
                <dd class="mt-1 font-medium text-gray-900 dark:text-white">
                  {{ selectedKey.group?.name || apiGuideText('noGroupAssigned') }}
                </dd>
              </div>
              <div>
                <dt class="text-gray-500 dark:text-gray-400">{{ apiGuideText('platform') }}</dt>
                <dd class="mt-1 font-medium text-gray-900 dark:text-white">
                  {{ selectedKey.group ? platformLabel(selectedKey.group.platform) : apiGuideText('noSelection') }}
                </dd>
              </div>
              <div>
                <dt class="text-gray-500 dark:text-gray-400">{{ apiGuideText('status') }}</dt>
                <dd class="mt-1 font-medium text-gray-900 dark:text-white">
                  {{ selectedKey.status }}
                </dd>
              </div>
            </dl>
          </div>

          <div class="rounded-2xl border border-blue-200 bg-blue-50 p-4 dark:border-blue-500/30 dark:bg-blue-500/10">
            <div class="mb-2 flex items-center gap-2 text-sm font-semibold text-blue-800 dark:text-blue-100">
              <Icon name="shield" size="sm" />
              <span>{{ apiGuideText('authHeaderTitle') }}</span>
            </div>
            <p class="text-xs leading-5 text-blue-700 dark:text-blue-100/90">
              {{ apiGuideText('authHeaderDescription') }}
            </p>
            <code class="mt-3 block break-all rounded-2xl bg-white/90 px-3 py-3 text-xs text-blue-800 shadow-sm dark:bg-dark-900 dark:text-blue-100">
              {{ authHeaderPreview }}
            </code>
          </div>
        </aside>

        <section class="space-y-4">
          <div
            v-if="variants.length === 0"
            class="card rounded-2xl border border-dashed border-gray-300 px-6 py-10 text-center text-sm text-gray-500 dark:border-dark-600 dark:text-gray-400"
          >
            {{ apiGuideText('noEndpointVariants') }}
          </div>

          <article
            v-for="variant in variants"
            :key="variant.id"
            class="surface-panel overflow-hidden"
          >
            <div class="border-b border-gray-200/70 bg-white/[0.35] px-6 py-5 dark:border-white/5 dark:bg-white/[0.02]">
              <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                <div class="space-y-2">
                  <div class="flex flex-wrap items-center gap-2">
                    <div class="page-kicker">
                      <Icon :name="variant.protocol === 'google' ? 'sparkles' : variant.protocol === 'openai' ? 'cpu' : 'chat'" size="sm" />
                      <span>{{ t(`${variant.translationKey}.label`) }}</span>
                    </div>
                    <span class="badge badge-gray">{{ protocolLabel(variant.protocol) }}</span>
                    <span class="badge badge-primary">{{ headerModeLabel(variant.headerMode) }}</span>
                    <span v-if="variant.supportsStream" class="badge badge-success">{{ apiGuideText('stream') }}</span>
                  </div>
                  <div class="inline-flex items-center gap-2 rounded-full bg-white px-3 py-1 text-xs font-semibold text-gray-700 shadow-sm dark:bg-dark-900 dark:text-gray-200">
                    <Icon :name="variant.protocol === 'google' ? 'sparkles' : variant.protocol === 'openai' ? 'cpu' : 'chat'" size="sm" />
                    <span>{{ buildGatewayRelativePath(variant.id, variant.defaultModel) }}</span>
                  </div>
                  <p class="text-sm leading-6 text-gray-600 dark:text-gray-300">
                    {{ t(`${variant.translationKey}.description`) }}
                  </p>
                </div>

                <router-link
                  :to="{ path: apiGuideDefaults.testPath, query: { key: String(selectedKey?.id || ''), variant: variant.id } }"
                  class="btn btn-secondary"
                >
                  {{ apiGuideText('testThisVariant') }}
                </router-link>
              </div>
            </div>

            <div class="space-y-5 px-6 py-6">
              <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
                <div class="metric-panel">
                  <div class="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
                    {{ apiGuideText('endpoint') }}
                  </div>
                  <code class="mt-2 block break-all text-sm text-gray-900 dark:text-white">
                    {{ buildGatewayRelativePath(variant.id, variant.defaultModel) }}
                  </code>
                </div>

                <div class="metric-panel">
                  <div class="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
                    {{ apiGuideText('protocol') }}
                  </div>
                  <div class="mt-2 text-sm font-medium text-gray-900 dark:text-white">
                    {{ protocolLabel(variant.protocol) }}
                  </div>
                </div>

                <div class="metric-panel">
                  <div class="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
                    {{ apiGuideText('defaultModel') }}
                  </div>
                  <code class="mt-2 block break-all text-sm text-gray-900 dark:text-white">
                    {{ variant.defaultModel }}
                  </code>
                </div>

                <div class="metric-panel">
                  <div class="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
                    {{ apiGuideText('headerMode') }}
                  </div>
                  <code class="mt-2 block break-all text-sm text-gray-900 dark:text-white">
                    {{ headerModeLabel(variant.headerMode) }}
                  </code>
                </div>
              </div>

              <div class="space-y-3">
                <div class="flex items-center justify-between gap-3">
                  <div>
                    <div class="text-sm font-semibold text-gray-900 dark:text-white">
                      {{ apiGuideText('curlExample') }}
                    </div>
                    <div class="text-xs text-gray-500 dark:text-gray-400">
                      {{ buildGatewayAbsoluteUrl(gatewayBaseUrl, variant.id, variant.defaultModel) }}
                    </div>
                  </div>

                  <button class="btn btn-secondary" @click="copyCommand(variant.id)">
                    {{ apiGuideText('copyCurl') }}
                  </button>
                </div>

                <pre class="code-surface">{{ buildCurl(variant) }}</pre>
              </div>
            </div>
          </article>
        </section>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { resolveRuntimeLocale } from '@/utils/runtimeLocale'
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import { keysAPI } from '@/api'
import { useAppStore } from '@/stores/app'
import { useClipboard } from '@/composables/useClipboard'
import type { ApiKey, GroupPlatform } from '@/types'
import type { GatewayVariantId } from '@/utils/gatewayDocs'
import { useAuthRouteDefaults } from '@/composables/useAuthRouteDefaults'
import {
  renderAPIGuideShellText,
  resolveAPIGuideShellDefaults,
  resolveAPIGuideShellLabels,
  type APIGuideLabelKey,
} from '@/utils/apiGuideShell'
import {
  buildGatewayAbsoluteUrl,
  buildGatewayCurlExample,
  buildGatewayRelativePath,
  getGatewayBaseUrl,
  getGatewayVariantsForApiKey,
  resolveGatewayVariantOverrides
} from '@/utils/gatewayDocs'
import {
  buildApiGuideKeyOptions,
  maskApiGuideKey,
  resolveApiGuideAuthHeaderPreview,
} from './apiGuideRuntime'

const { t, locale } = useI18n()
const route = useRoute()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()
const { authRouteDefaults } = useAuthRouteDefaults()


const apiGuideLabels = computed(() =>
  resolveAPIGuideShellLabels(
    appStore.cachedPublicSettings?.api_guide_shell_config,
    resolveRuntimeLocale(locale),
  ),
)
const apiGuideDefaults = computed(() =>
  resolveAPIGuideShellDefaults(
    appStore.cachedPublicSettings?.api_guide_shell_config,
    resolveRuntimeLocale(locale),
  ),
)

const gatewayVariantOverrides = computed(() =>
  resolveGatewayVariantOverrides(
    appStore.cachedPublicSettings?.api_guide_shell_config,
    resolveRuntimeLocale(locale),
  ),
)

function apiGuideText(key: APIGuideLabelKey): string {
  return renderAPIGuideShellText(apiGuideLabels.value, key)
}

const loading = ref(false)
const keys = ref<ApiKey[]>([])
const selectedKeyId = ref<number | null>(null)

const gatewayBaseUrl = computed(() => getGatewayBaseUrl(appStore.cachedPublicSettings?.api_base_url))

const selectedKey = computed(() => keys.value.find(key => key.id === selectedKeyId.value) ?? null)
const selectedKeyMasked = computed(() =>
  selectedKey.value ? maskApiGuideKey(selectedKey.value.key) : apiGuideText('selectKeyHint'),
)

const variants = computed(() => getGatewayVariantsForApiKey(selectedKey.value, gatewayVariantOverrides.value))

const keyOptions = computed(() =>
  buildApiGuideKeyOptions(keys.value, apiGuideText('noGroupAssigned')),
)

const authHeaderPreview = computed(() => {
  const firstGoogleVariant = variants.value.find(variant => variant.headerMode === 'x-goog-api-key')
  return resolveApiGuideAuthHeaderPreview(Boolean(firstGoogleVariant))
})

function platformLabel(platform: GroupPlatform): string {
  return t(`gateway.platforms.${platform}`)
}

function protocolLabel(protocol: 'anthropic' | 'openai' | 'google'): string {
  return t(`gateway.protocols.${protocol}`)
}

function headerModeLabel(mode: 'bearer' | 'x-goog-api-key'): string {
  return t(`gateway.headerModes.${mode}`)
}

function buildCurl(variant: { id: GatewayVariantId; defaultModel: string }): string {
  if (!selectedKey.value) return ''
  return buildGatewayCurlExample(
    gatewayBaseUrl.value,
    selectedKey.value.key,
    variant.id,
    variant.defaultModel,
    apiGuideDefaults.value.defaultPrompt,
    false,
    { maxTokens: apiGuideDefaults.value.maxTokens },
  )
}

async function copyCommand(variantId: GatewayVariantId) {
  const variant = variants.value.find(item => item.id === variantId)
  if (!selectedKey.value || !variant) return
  await copyToClipboard(buildCurl(variant), apiGuideText('copyCurlSuccess'))
}

async function loadKeys() {
  loading.value = true
  try {
    const response = await keysAPI.list(1, apiGuideDefaults.value.apiKeyPageSize)
    keys.value = response.items

    const queryKeyId = Number(route.query.key)
    if (Number.isFinite(queryKeyId) && keys.value.some(key => key.id === queryKeyId)) {
      selectedKeyId.value = queryKeyId
      return
    }

    const firstActive = keys.value.find(key => key.status === 'active')
    selectedKeyId.value = firstActive?.id ?? keys.value[0]?.id ?? null
  } catch (error) {
    console.error('Failed to load API keys for guide:', error)
    appStore.showError(apiGuideText('loadKeysFailed'))
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadKeys()
})
</script>
