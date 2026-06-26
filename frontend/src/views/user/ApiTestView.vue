<template>
  <AppLayout>
    <div class="space-y-6">
      <section class="page-hero">
        <div class="page-hero-grid">
          <div class="space-y-4">
            <div class="page-kicker">
              <Icon name="beaker" size="sm" />
              <span>{{ apiTestText('badge') }}</span>
            </div>
            <div class="space-y-3">
              <h1 class="max-w-3xl text-3xl font-bold tracking-tight text-gray-900 dark:text-white">
                {{ apiTestText('title') }}
              </h1>
              <p class="max-w-3xl text-sm leading-7 text-gray-600 dark:text-gray-300">
                {{ apiTestText('description') }}
              </p>
            </div>

            <div class="flex flex-wrap items-center gap-3">
              <router-link :to="apiTestDefaults.guidePath" class="btn btn-secondary">
                {{ apiTestText('openGuide') }}
              </router-link>
              <button class="btn btn-primary" :disabled="sending || !canSend" @click="runTest">
                {{ sending ? apiTestText('sending') : apiTestText('send') }}
              </button>
            </div>
          </div>

          <div class="grid gap-3 sm:grid-cols-3 xl:grid-cols-1">
            <div class="metric-panel">
              <div class="flex items-start gap-3">
                <div class="metric-icon">
                  <Icon name="key" size="sm" />
                </div>
                <div class="min-w-0">
                  <div class="text-xs font-semibold uppercase tracking-[0.18em] text-gray-500 dark:text-gray-400">
                    {{ apiTestText('keySelector') }}
                  </div>
                  <p class="mt-2 truncate text-sm font-semibold text-gray-900 dark:text-white">
                    {{ selectedKey?.name || apiTestText('noSelection') }}
                  </p>
                  <p class="mt-1 truncate text-xs text-gray-500 dark:text-gray-400">
                    {{ selectedKey ? maskKey(selectedKey.key) : apiTestText('noGroupAssigned') }}
                  </p>
                </div>
              </div>
            </div>

            <div class="metric-panel">
              <div class="flex items-start gap-3">
                <div class="metric-icon">
                  <Icon name="chat" size="sm" />
                </div>
                <div class="min-w-0">
                  <div class="text-xs font-semibold uppercase tracking-[0.18em] text-gray-500 dark:text-gray-400">
                    {{ apiTestText('protocol') }}
                  </div>
                  <p class="mt-2 text-sm font-semibold text-gray-900 dark:text-white">
                    {{ selectedVariantLabel }}
                  </p>
                  <p class="mt-1 truncate text-xs text-gray-500 dark:text-gray-400">
                    {{ selectedPlatformLabel }}
                  </p>
                </div>
              </div>
            </div>

            <div class="metric-panel">
              <div class="flex items-start gap-3">
                <div class="metric-icon">
                  <Icon name="sparkles" size="sm" />
                </div>
                <div class="min-w-0">
                  <div class="text-xs font-semibold uppercase tracking-[0.18em] text-gray-500 dark:text-gray-400">
                    {{ apiTestText('model') }}
                  </div>
                  <p class="mt-2 truncate text-sm font-semibold text-gray-900 dark:text-white">
                    {{ selectedModelLabel }}
                  </p>
                  <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                    {{ stream && selectedVariant?.supportsStream ? apiTestText('stream') : apiTestText('requestMeta') }}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <div v-if="loadingKeys" class="card flex items-center justify-center px-6 py-16">
        <LoadingSpinner />
      </div>

      <EmptyState
        v-else-if="keys.length === 0"
        :title="apiTestText('noKeysTitle')"
        :description="apiTestText('noKeysDescription')"
        :action-text="apiTestText('manageKeys')"
        :action-to="authRouteDefaults.apiKeysPath"
      >
        <template #icon>
          <Icon name="key" size="xl" class="text-gray-400" />
        </template>
      </EmptyState>

      <div v-else class="grid gap-6 xl:grid-cols-[380px,minmax(0,1fr)]">
        <aside class="sticky-panel space-y-4 self-start">
          <div class="surface-panel-strong space-y-4">
            <div>
              <label class="input-label mb-1.5 block">{{ apiTestText('keySelector') }}</label>
              <Select
                v-model="selectedKeyId"
                :options="keyOptions"
                :placeholder="apiTestText('keySelector')"
                searchable
              />
            </div>

            <div>
              <label class="input-label mb-1.5 block">{{ apiTestText('protocol') }}</label>
              <Select
                v-model="selectedVariantId"
                :options="variantOptions"
                :placeholder="apiTestText('protocol')"
                :disabled="variantOptions.length === 0"
              />
            </div>

            <div>
              <label class="input-label mb-1.5 block">{{ apiTestText('model') }}</label>
              <Select
                v-model="selectedModelOption"
                :options="modelOptionsForSelect"
                :placeholder="selectedVariant?.modelPlaceholder || apiTestText('modelPlaceholder')"
                :search-placeholder="apiTestText('modelSearchPlaceholder')"
                :empty-text="loadingModels ? apiTestText('loading') : apiTestText('noOptionsFound')"
                :disabled="!selectedVariant"
                searchable
              />
              <p class="input-hint mt-1.5">
                {{ loadingModels ? apiTestText('loading') : apiTestText('modelHint') }}
              </p>
            </div>

            <Input
              v-if="showCustomModelInput"
              v-model="model"
              :label="apiTestText('customModel')"
              :placeholder="selectedVariant?.modelPlaceholder || apiTestText('modelPlaceholder')"
              :hint="apiTestText('customModelHint')"
            />

            <TextArea
              v-model="prompt"
              :label="apiTestText('prompt')"
              :placeholder="apiTestText('promptPlaceholder')"
              :hint="apiTestText('promptHint')"
              rows="6"
            />

            <label
              v-if="selectedVariant?.supportsStream"
              class="surface-panel-muted flex items-start gap-3 text-sm"
            >
              <input v-model="stream" type="checkbox" class="mt-0.5 h-4 w-4 rounded border-gray-300 text-primary-500 focus-visible:ring-2 focus-visible:ring-primary-500" />
              <div>
                <div class="font-medium text-gray-900 dark:text-white">{{ apiTestText('stream') }}</div>
                <div class="mt-1 text-xs leading-5 text-gray-500 dark:text-gray-400">
                  {{ apiTestText('streamHint') }}
                </div>
              </div>
            </label>

            <div
              v-if="selectedKey && !selectedKey.group"
              class="surface-panel border-amber-200/80 bg-amber-50/90 px-4 py-3 text-sm text-amber-800 dark:border-amber-500/30 dark:bg-amber-500/10 dark:text-amber-100"
            >
              <div class="font-semibold">{{ apiTestText('unassignedTitle') }}</div>
              <div class="mt-1 text-xs leading-5">{{ apiTestText('unassignedDescription') }}</div>
            </div>

            <div
              class="surface-panel border-sky-200/80 bg-sky-50/90 px-4 py-3 text-sm text-sky-900 dark:border-sky-500/30 dark:bg-sky-500/10 dark:text-sky-100"
            >
              <div class="font-semibold">{{ apiTestText('liveBillingTitle') }}</div>
              <div class="mt-1 text-xs leading-5">{{ apiTestText('liveBillingDescription') }}</div>
            </div>

            <div class="flex flex-wrap items-center gap-3">
              <button class="btn btn-primary" :disabled="sending || !canSend" @click="runTest">
                {{ sending ? apiTestText('sending') : apiTestText('send') }}
              </button>
              <button class="btn btn-secondary" :disabled="!curlCommand" @click="copyCurlCommand">
                {{ apiTestText('copyCurl') }}
              </button>
            </div>
          </div>

          <div class="surface-panel space-y-3">
            <div class="flex items-center justify-between">
              <div>
                <div class="text-sm font-semibold text-gray-900 dark:text-white">{{ apiTestText('requestMeta') }}</div>
                <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ requestUrl }}</div>
              </div>
            </div>

            <dl class="space-y-3 text-sm">
              <div>
                <dt class="text-gray-500 dark:text-gray-400">{{ apiTestText('platform') }}</dt>
                <dd class="mt-1 font-medium text-gray-900 dark:text-white">
                  {{ selectedKey?.group ? t(`gateway.platforms.${selectedKey.group.platform}`) : apiTestText('notReady') }}
                </dd>
              </div>
              <div>
                <dt class="text-gray-500 dark:text-gray-400">{{ apiTestText('headerMode') }}</dt>
                <dd class="mt-1 break-all font-mono text-xs text-gray-900 dark:text-white">
                  {{ headerPreview }}
                </dd>
              </div>
            </dl>
          </div>
        </aside>

        <section class="space-y-4">
          <div class="surface-panel">
            <div class="mb-3 flex items-center justify-between gap-3">
              <div>
                <div class="text-sm font-semibold text-gray-900 dark:text-white">{{ apiTestText('requestPreview') }}</div>
                <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ requestUrl }}</div>
              </div>
              <button class="btn btn-secondary" :disabled="!requestBodyPreview" @click="copyRequestBody">
                {{ apiTestText('copyRequest') }}
              </button>
            </div>
            <pre class="code-surface">{{ requestBodyPreview }}</pre>
          </div>

          <div class="surface-panel min-h-[28rem]">
            <div class="mb-3 flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
              <div>
                <div class="text-sm font-semibold text-gray-900 dark:text-white">{{ apiTestText('responsePreview') }}</div>
                <div class="mt-2 flex flex-wrap items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
                  <span class="badge badge-gray">{{ apiTestText('statusCode') }} · {{ responseStatusLabel }}</span>
                  <span class="badge badge-gray">{{ apiTestText('duration') }} · {{ responseDurationLabel }}</span>
                  <span v-if="stream && selectedVariant?.supportsStream" class="badge badge-primary">
                    {{ apiTestText('stream') }}
                  </span>
                </div>
              </div>

              <button class="btn btn-secondary" :disabled="!responseText" @click="copyResponseText">
                {{ apiTestText('copyResponse') }}
              </button>
            </div>

            <div
              v-if="responsePreview"
              aria-live="polite"
              :class="[
                'mb-4 rounded-[24px] border px-4 py-3 text-sm',
                responseStatus && responseStatus >= 200 && responseStatus < 300
                  ? 'border-emerald-200 bg-emerald-50 text-emerald-900 dark:border-emerald-500/30 dark:bg-emerald-500/10 dark:text-emerald-100'
                  : 'border-amber-200 bg-amber-50 text-amber-900 dark:border-amber-500/30 dark:bg-amber-500/10 dark:text-amber-100'
              ]"
            >
              <div class="font-semibold">{{ apiTestText('responseSummary') }}</div>
              <div class="mt-1 whitespace-pre-wrap break-words text-xs leading-6">{{ responsePreview }}</div>
            </div>

            <div
              v-if="showUsageRecordNotice"
              aria-live="polite"
              class="mb-4 rounded-2xl border border-sky-200 bg-sky-50 px-4 py-3 text-sm text-sky-900 dark:border-sky-500/30 dark:bg-sky-500/10 dark:text-sky-100"
            >
              <div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
                <div>
                  <div class="font-semibold">{{ apiTestText('usageRecordTitle') }}</div>
                  <div class="mt-1 text-xs leading-5">{{ usageRecordSummary }}</div>
                </div>
                <router-link :to="usageInspectTo" class="btn btn-secondary">
                  {{ apiTestText('openUsage') }}
                </router-link>
              </div>
            </div>

            <div v-if="responseText" class="space-y-2">
              <div class="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
                {{ apiTestText('rawResponse') }}
              </div>
              <pre class="code-surface">{{ formattedResponseText }}</pre>
            </div>

            <div
              v-else
              class="rounded-2xl border border-dashed border-gray-300 px-6 py-10 text-center text-sm text-gray-500 dark:border-dark-600 dark:text-gray-400"
            >
              {{ apiTestText('responsePending') }}
            </div>
          </div>
        </section>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { resolveRuntimeLocale } from '@/utils/runtimeLocale'
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Input from '@/components/common/Input.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Select from '@/components/common/Select.vue'
import TextArea from '@/components/common/TextArea.vue'
import Icon from '@/components/icons/Icon.vue'
import { keysAPI, usageAPI } from '@/api'
import { useAuthRouteDefaults } from '@/composables/useAuthRouteDefaults'
import { useClipboard } from '@/composables/useClipboard'
import { useAppStore } from '@/stores/app'
import type { ApiKey, UsageLog } from '@/types'
import type { GatewayModelOption, GatewayVariant, GatewayVariantId } from '@/utils/gatewayDocs'
import {
  renderAPITestShellText,
  resolveAPITestShellDefaults,
  resolveAPITestShellLabels,
  type APITestLabelKey,
} from '@/utils/apiTestShell'
import {
  buildGatewayAbsoluteUrl,
  buildGatewayCurlExample,
  buildGatewayHeaders,
  buildGatewayModelsRelativePath,
  buildGatewayRelativePath,
  buildGatewayRequestBody,
  extractGatewayModelOptions,
  extractGatewayResponsePreview,
  getGatewayFallbackModelOptions,
  getGatewayBaseUrl,
  getGatewayVariantById,
  getGatewayVariantsForApiKey,
  resolveGatewayVariantOverrides
} from '@/utils/gatewayDocs'

const { t, locale } = useI18n()
const route = useRoute()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()
const { authRouteDefaults } = useAuthRouteDefaults()


const apiTestLabels = computed(() =>
  resolveAPITestShellLabels(
    appStore.cachedPublicSettings?.api_test_shell_config,
    resolveRuntimeLocale(locale),
  ),
)
const apiTestDefaults = computed(() =>
  resolveAPITestShellDefaults(
    appStore.cachedPublicSettings?.api_test_shell_config,
    resolveRuntimeLocale(locale),
  ),
)

const gatewayVariantOverrides = computed(() =>
  resolveGatewayVariantOverrides(
    appStore.cachedPublicSettings?.api_guide_shell_config,
    resolveRuntimeLocale(locale),
  ),
)

function apiTestText(key: APITestLabelKey, values?: Record<string, string | number>): string {
  return renderAPITestShellText(apiTestLabels.value, key, values)
}

const loadingKeys = ref(false)
const loadingModels = ref(false)
const sending = ref(false)
const keys = ref<ApiKey[]>([])
const selectedKeyId = ref<number | null>(null)
const selectedVariantId = ref<GatewayVariantId | null>(null)
const model = ref('')
const modelOptions = ref<GatewayModelOption[]>([])
const prompt = ref('')
const stream = ref(false)
const responseStatus = ref<number | null>(null)
const responseDurationMs = ref<number | null>(null)
const responseText = ref('')
const latestUsageRecord = ref<UsageLog | null>(null)
const usageRecordSyncState = ref<'idle' | 'checking' | 'found' | 'pending'>('idle')
const CUSTOM_MODEL_OPTION = '__custom_model__'
const customModelMode = ref(false)
let activeModelRequestID = 0
let activeTestAbortController: AbortController | null = null
let lastAppliedDefaultPrompt = ''

const gatewayBaseUrl = computed(() => getGatewayBaseUrl(appStore.cachedPublicSettings?.api_base_url))

const selectedKey = computed(() => keys.value.find(key => key.id === selectedKeyId.value) ?? null)
const availableVariants = computed(() => getGatewayVariantsForApiKey(selectedKey.value, gatewayVariantOverrides.value))
const selectedVariant = computed<GatewayVariant | null>(() => {
  if (!selectedVariantId.value) return null
  return availableVariants.value.find(variant => variant.id === selectedVariantId.value) ?? null
})
const selectedVariantLabel = computed(() => {
  return selectedVariant.value ? t(`${selectedVariant.value.translationKey}.label`) : apiTestText('notReady')
})
const selectedPlatformLabel = computed(() => {
  return selectedKey.value?.group ? t(`gateway.platforms.${selectedKey.value.group.platform}`) : apiTestText('notReady')
})
const selectedModelLabel = computed(() => {
  const selectedModel = model.value.trim()
  if (selectedModel) return selectedModel
  return selectedVariant.value?.defaultModel || '-'
})

const keyOptions = computed(() => keys.value.map(key => ({
  value: key.id,
  label: `${key.name} · ${maskKey(key.key)}`,
  description: key.group?.name || apiTestText('noGroupAssigned')
})))

const variantOptions = computed(() => availableVariants.value.map(variant => ({
  value: variant.id,
  label: t(`${variant.translationKey}.label`),
  description: t(`${variant.translationKey}.description`)
})))

const availableModelIDs = computed(() => new Set(modelOptions.value.map(option => option.id)))
const modelOptionsForSelect = computed(() => [
  ...modelOptions.value.map(option => ({
    value: option.id,
    label: option.label,
    description: option.description
  })),
  {
    value: CUSTOM_MODEL_OPTION,
    label: apiTestText('customModelOption'),
    description: apiTestText('customModelOptionHint')
  }
])
const selectedModelOption = computed<string | null>({
  get() {
    if (customModelMode.value) {
      return CUSTOM_MODEL_OPTION
    }
    const selectedModel = model.value.trim()
    if (!selectedModel) return null
    return availableModelIDs.value.has(selectedModel) ? selectedModel : CUSTOM_MODEL_OPTION
  },
  set(value) {
    if (!value) {
      customModelMode.value = false
      model.value = ''
      return
    }
    if (value === CUSTOM_MODEL_OPTION) {
      customModelMode.value = true
      if (availableModelIDs.value.has(model.value.trim())) {
        model.value = ''
      }
      return
    }
    customModelMode.value = false
    model.value = value
  }
})
const showCustomModelInput = computed(() => {
  return customModelMode.value || (!!model.value.trim() && !availableModelIDs.value.has(model.value.trim()))
})

const requestUrl = computed(() => {
  if (!selectedVariant.value) return '-'
  return buildGatewayAbsoluteUrl(gatewayBaseUrl.value, selectedVariant.value.id, model.value)
})

const requestBodyPreview = computed(() => {
  if (!selectedVariant.value) return ''
  return JSON.stringify(
    buildGatewayRequestBody(
      selectedVariant.value.id,
      model.value,
      prompt.value,
      stream.value,
      { maxTokens: apiTestDefaults.value.maxTokens },
    ),
    null,
    2
  )
})

const curlCommand = computed(() => {
  if (!selectedKey.value || !selectedVariant.value) return ''
  return buildGatewayCurlExample(
    gatewayBaseUrl.value,
    selectedKey.value.key,
    selectedVariant.value.id,
    model.value,
    prompt.value,
    stream.value,
    { maxTokens: apiTestDefaults.value.maxTokens }
  )
})

const headerPreview = computed(() => {
  if (!selectedVariant.value) return '-'
  return selectedVariant.value.headerMode === 'x-goog-api-key'
    ? 'x-goog-api-key: <API_KEY>'
    : 'Authorization: Bearer <API_KEY>'
})

const responsePreview = computed(() => {
  if (!selectedVariant.value || !responseText.value) return ''
  return extractGatewayResponsePreview(selectedVariant.value.id, responseText.value)
})

const formattedResponseText = computed(() => {
  if (!responseText.value.trim()) return ''
  try {
    return JSON.stringify(JSON.parse(responseText.value), null, 2)
  } catch {
    return responseText.value
  }
})

const responseStatusLabel = computed(() => responseStatus.value === null ? '-' : String(responseStatus.value))
const responseDurationLabel = computed(() => responseDurationMs.value === null ? '-' : `${responseDurationMs.value} ms`)
const todayDate = computed(() => formatLocalDate(new Date()))
const usageInspectTo = computed(() => ({
  path: authRouteDefaults.value.usagePath,
  query: selectedKey.value
    ? {
        api_key_id: String(selectedKey.value.id),
        start_date: todayDate.value,
        end_date: todayDate.value
      }
    : {}
}))
const showUsageRecordNotice = computed(() => {
  return responseStatus.value !== null && responseStatus.value >= 200 && responseStatus.value < 300
})
const usageRecordSummary = computed(() => {
  if (usageRecordSyncState.value === 'checking') {
    return apiTestText('usageRecordSyncing')
  }
  if (usageRecordSyncState.value === 'found' && latestUsageRecord.value) {
    return apiTestText('usageRecordFound', {
      time: formatUsageRecordTime(latestUsageRecord.value.created_at),
      cost: latestUsageRecord.value.actual_cost.toFixed(4),
      tokens: totalUsageTokens(latestUsageRecord.value).toLocaleString()
    })
  }
  if (usageRecordSyncState.value === 'pending') {
    return apiTestText('usageRecordPending')
  }
  return apiTestText('usageRecordIdle')
})

const canSend = computed(() => {
  return !!selectedKey.value?.group && !!selectedVariant.value && !!model.value.trim()
})

watch(
  () => apiTestDefaults.value.defaultPrompt,
  (defaultPrompt) => {
    if (!prompt.value.trim() || prompt.value === lastAppliedDefaultPrompt) {
      prompt.value = defaultPrompt
    }
    lastAppliedDefaultPrompt = defaultPrompt
  },
  { immediate: true }
)

function formatLocalDate(date: Date): string {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

function maskKey(key: string): string {
  if (key.length <= 14) return key
  return `${key.slice(0, 8)}...${key.slice(-4)}`
}

function resetResponse() {
  responseStatus.value = null
  responseDurationMs.value = null
  responseText.value = ''
  latestUsageRecord.value = null
  usageRecordSyncState.value = 'idle'
}

function applyVariantDefaults(variantId: GatewayVariantId | null) {
  if (!variantId) {
    model.value = ''
    customModelMode.value = false
    stream.value = false
    return
  }
  const variant = getGatewayVariantById(variantId, gatewayVariantOverrides.value)
  model.value = variant.defaultModel
  customModelMode.value = false
  if (!variant.supportsStream) {
    stream.value = false
  }
}

async function loadModelOptions() {
  const key = selectedKey.value
  const variant = selectedVariant.value

  if (!key || !variant) {
    modelOptions.value = []
    loadingModels.value = false
    return
  }

  const requestID = ++activeModelRequestID
  const fallbackOptions = getGatewayFallbackModelOptions(variant.id, gatewayVariantOverrides.value)
  modelOptions.value = fallbackOptions
  loadingModels.value = true

  try {
    const response = await fetch(buildGatewayModelsRelativePath(variant.id), {
      method: 'GET',
      headers: buildGatewayHeaders(key.key, variant.id)
    })
    const responseText = await response.text()
    const options = response.ok
      ? extractGatewayModelOptions(variant.id, responseText, gatewayVariantOverrides.value)
      : fallbackOptions

    if (requestID !== activeModelRequestID) return
    modelOptions.value = options.length > 0 ? options : fallbackOptions
  } catch (error) {
    if (requestID !== activeModelRequestID) return
    console.error('Failed to load gateway models for tester:', error)
    modelOptions.value = fallbackOptions
  } finally {
    if (requestID === activeModelRequestID) {
      loadingModels.value = false
    }
  }
}

async function copyCurlCommand() {
  if (!curlCommand.value) return
  await copyToClipboard(curlCommand.value, apiTestText('copyCurlSuccess'))
}

async function copyRequestBody() {
  if (!requestBodyPreview.value) return
  await copyToClipboard(requestBodyPreview.value, apiTestText('copyRequestSuccess'))
}

async function copyResponseText() {
  if (!formattedResponseText.value) return
  await copyToClipboard(formattedResponseText.value, apiTestText('copyResponseSuccess'))
}

async function loadKeys() {
  loadingKeys.value = true
  try {
    const response = await keysAPI.list(1, apiTestDefaults.value.apiKeyPageSize)
    keys.value = response.items

    const queryKeyId = Number(route.query.key)
    if (Number.isFinite(queryKeyId) && keys.value.some(key => key.id === queryKeyId)) {
      selectedKeyId.value = queryKeyId
    } else {
      const firstActive = keys.value.find(key => key.status === 'active')
      selectedKeyId.value = firstActive?.id ?? keys.value[0]?.id ?? null
    }
  } catch (error) {
    console.error('Failed to load API keys for tester:', error)
    appStore.showError(apiTestText('loadKeysFailed'))
  } finally {
    loadingKeys.value = false
  }
}

async function runTest() {
  if (!selectedKey.value || !selectedVariant.value) return

  sending.value = true
  resetResponse()
  activeTestAbortController?.abort()
  activeTestAbortController = new AbortController()

  const requestPath = buildGatewayRelativePath(selectedVariant.value.id, model.value)
  const headers = buildGatewayHeaders(selectedKey.value.key, selectedVariant.value.id)
  const body = buildGatewayRequestBody(
    selectedVariant.value.id,
    model.value,
    prompt.value,
    stream.value,
    { maxTokens: apiTestDefaults.value.maxTokens },
  )
  const start = performance.now()
  const startedAt = Date.now()

  try {
    const response = await fetch(requestPath, {
      method: 'POST',
      headers,
      body: JSON.stringify(body),
      signal: activeTestAbortController.signal
    })

    responseStatus.value = response.status
    responseText.value = stream.value
      ? await readStreamingResponse(response, (chunk) => {
          responseText.value += chunk
        })
      : await response.text()

    if (response.ok) {
      await syncLatestUsageRecord(startedAt)
    }
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') {
      return
    }
    const message = error instanceof Error ? error.message : apiTestText('unknownError')
    responseText.value = message
    appStore.showError(message)
  } finally {
    responseDurationMs.value = Math.round(performance.now() - start)
    sending.value = false
    activeTestAbortController = null
  }
}

async function readStreamingResponse(
  response: Response,
  onChunk: (chunk: string) => void
): Promise<string> {
  const reader = response.body?.getReader()
  if (!reader) {
    const text = await response.text()
    onChunk(text)
    return text
  }

  const decoder = new TextDecoder()
  let combined = ''

  while (true) {
    const { done, value } = await reader.read()
    if (done) break

    const chunk = decoder.decode(value, { stream: true })
    if (!chunk) continue
    combined += chunk
    onChunk(chunk)
  }

  const lastChunk = decoder.decode()
  if (lastChunk) {
    combined += lastChunk
    onChunk(lastChunk)
  }

  return combined
}

function totalUsageTokens(record: UsageLog): number {
  return record.input_tokens + record.output_tokens + record.cache_creation_tokens + record.cache_read_tokens
}

function formatUsageRecordTime(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  return date.toLocaleString()
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => window.setTimeout(resolve, ms))
}

async function syncLatestUsageRecord(startedAt: number) {
  if (!selectedKey.value) return

  usageRecordSyncState.value = 'checking'
  latestUsageRecord.value = null

  const threshold = startedAt - 5000
  const day = formatLocalDate(new Date(startedAt))

  for (let attempt = 0; attempt < 6; attempt += 1) {
    try {
      const response = await usageAPI.query({
        page: 1,
        page_size: apiTestDefaults.value.usageSyncPageSize,
        api_key_id: selectedKey.value.id,
        start_date: day,
        end_date: day
      })

      const match = response.items.find((item) => {
        const createdAt = Date.parse(item.created_at)
        return Number.isFinite(createdAt) && createdAt >= threshold
      })

      if (match) {
        latestUsageRecord.value = match
        usageRecordSyncState.value = 'found'
        return
      }
    } catch (error) {
      console.error('Failed to sync gateway test usage record:', error)
      break
    }

    if (attempt < 5) {
      await sleep(400)
    }
  }

  usageRecordSyncState.value = 'pending'
}

watch(selectedKey, (key) => {
  const queryVariant = typeof route.query.variant === 'string' ? route.query.variant as GatewayVariantId : null
  const variants = getGatewayVariantsForApiKey(key)
  if (queryVariant && variants.some(variant => variant.id === queryVariant)) {
    selectedVariantId.value = queryVariant
    applyVariantDefaults(queryVariant)
    return
  }

  if (!selectedVariantId.value || !variants.some(variant => variant.id === selectedVariantId.value)) {
    selectedVariantId.value = variants[0]?.id ?? null
    applyVariantDefaults(selectedVariantId.value)
  }
})

watch(selectedVariantId, (variantId, previousId) => {
  if (variantId === previousId) return
  applyVariantDefaults(variantId)
  resetResponse()
})

watch([selectedKeyId, selectedVariantId], () => {
  loadModelOptions()
})

onMounted(() => {
  loadKeys()
})

onUnmounted(() => {
  activeTestAbortController?.abort()
})
</script>
