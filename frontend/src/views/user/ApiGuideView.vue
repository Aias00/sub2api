<template>
  <AppLayout>
    <div class="space-y-6">
      <section class="card overflow-hidden">
        <div class="border-b border-gray-100 px-6 py-6 dark:border-dark-700">
          <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
            <div class="space-y-2">
              <div class="inline-flex items-center gap-2 rounded-full bg-primary-50 px-3 py-1 text-xs font-semibold text-primary-700 dark:bg-primary-500/10 dark:text-primary-200">
                <Icon name="book" size="sm" />
                <span>{{ t('apiGuide.badge') }}</span>
              </div>
              <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
                {{ t('apiGuide.title') }}
              </h1>
              <p class="max-w-3xl text-sm leading-6 text-gray-600 dark:text-gray-300">
                {{ t('apiGuide.description') }}
              </p>
            </div>

            <div class="flex flex-wrap items-center gap-3">
              <router-link to="/gateway-test" class="btn btn-primary">
                {{ t('apiGuide.openTester') }}
              </router-link>
              <router-link to="/keys" class="btn btn-secondary">
                {{ t('apiGuide.manageKeys') }}
              </router-link>
            </div>
          </div>
        </div>

        <div class="grid gap-4 px-6 py-6 md:grid-cols-3">
          <div class="rounded-2xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-600 dark:bg-dark-800/60">
            <div class="mb-2 flex items-center gap-2 text-sm font-semibold text-gray-700 dark:text-gray-200">
              <Icon name="server" size="sm" />
              <span>{{ t('apiGuide.baseUrl') }}</span>
            </div>
            <code class="break-all text-sm text-primary-700 dark:text-primary-200">{{ gatewayBaseUrl }}</code>
          </div>

          <div class="rounded-2xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-600 dark:bg-dark-800/60">
            <div class="mb-2 flex items-center gap-2 text-sm font-semibold text-gray-700 dark:text-gray-200">
              <Icon name="key" size="sm" />
              <span>{{ t('apiGuide.currentKey') }}</span>
            </div>
            <p class="text-sm font-medium text-gray-900 dark:text-white">
              {{ selectedKey?.name || t('apiGuide.noSelection') }}
            </p>
            <p class="mt-1 break-all text-xs text-gray-500 dark:text-gray-400">
              {{ selectedKey ? maskKey(selectedKey.key) : t('apiGuide.selectKeyHint') }}
            </p>
          </div>

          <div class="rounded-2xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-600 dark:bg-dark-800/60">
            <div class="mb-2 flex items-center gap-2 text-sm font-semibold text-gray-700 dark:text-gray-200">
              <Icon name="grid" size="sm" />
              <span>{{ t('apiGuide.supportedEndpoints') }}</span>
            </div>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ variants.length }}</p>
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
              {{ selectedKey?.group ? platformLabel(selectedKey.group.platform) : t('apiGuide.noGroupAssigned') }}
            </p>
          </div>
        </div>
      </section>

      <div v-if="loading" class="card flex items-center justify-center px-6 py-16">
        <LoadingSpinner />
      </div>

      <EmptyState
        v-else-if="keys.length === 0"
        :title="t('apiGuide.noKeysTitle')"
        :description="t('apiGuide.noKeysDescription')"
        :action-text="t('apiGuide.manageKeys')"
        action-to="/keys"
      >
        <template #icon>
          <Icon name="key" size="xl" class="text-gray-400" />
        </template>
      </EmptyState>

      <div v-else class="grid gap-6 xl:grid-cols-[320px,minmax(0,1fr)]">
        <aside class="card space-y-5 p-5">
          <div>
            <label class="input-label mb-1.5 block">{{ t('apiGuide.keySelector') }}</label>
            <Select
              v-model="selectedKeyId"
              :options="keyOptions"
              :placeholder="t('apiGuide.keySelector')"
              searchable
            />
            <p class="mt-2 text-xs leading-5 text-gray-500 dark:text-gray-400">
              {{ t('apiGuide.keySelectorHint') }}
            </p>
          </div>

          <div
            v-if="selectedKey && !selectedKey.group"
            class="rounded-2xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-800 dark:border-amber-500/40 dark:bg-amber-500/10 dark:text-amber-100"
          >
            <div class="mb-1 font-semibold">{{ t('apiGuide.unassignedTitle') }}</div>
            <p class="text-xs leading-5">{{ t('apiGuide.unassignedDescription') }}</p>
          </div>

          <div
            v-else-if="selectedKey"
            class="rounded-2xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-600 dark:bg-dark-800/60"
          >
            <div class="mb-2 text-sm font-semibold text-gray-800 dark:text-gray-100">
              {{ t('apiGuide.keySummary') }}
            </div>
            <dl class="space-y-3 text-sm">
              <div>
                <dt class="text-gray-500 dark:text-gray-400">{{ t('apiGuide.groupName') }}</dt>
                <dd class="mt-1 font-medium text-gray-900 dark:text-white">
                  {{ selectedKey.group?.name || t('apiGuide.noGroupAssigned') }}
                </dd>
              </div>
              <div>
                <dt class="text-gray-500 dark:text-gray-400">{{ t('apiGuide.platform') }}</dt>
                <dd class="mt-1 font-medium text-gray-900 dark:text-white">
                  {{ selectedKey.group ? platformLabel(selectedKey.group.platform) : t('apiGuide.noSelection') }}
                </dd>
              </div>
              <div>
                <dt class="text-gray-500 dark:text-gray-400">{{ t('common.status') }}</dt>
                <dd class="mt-1 font-medium text-gray-900 dark:text-white">
                  {{ selectedKey.status }}
                </dd>
              </div>
            </dl>
          </div>

          <div class="rounded-2xl border border-blue-200 bg-blue-50 p-4 dark:border-blue-500/30 dark:bg-blue-500/10">
            <div class="mb-2 flex items-center gap-2 text-sm font-semibold text-blue-800 dark:text-blue-100">
              <Icon name="shield" size="sm" />
              <span>{{ t('apiGuide.authHeaderTitle') }}</span>
            </div>
            <p class="text-xs leading-5 text-blue-700 dark:text-blue-100/90">
              {{ t('apiGuide.authHeaderDescription') }}
            </p>
            <code class="mt-3 block break-all rounded-xl bg-white px-3 py-2 text-xs text-blue-800 shadow-sm dark:bg-dark-900 dark:text-blue-100">
              {{ authHeaderPreview }}
            </code>
          </div>
        </aside>

        <section class="space-y-4">
          <div
            v-if="variants.length === 0"
            class="card rounded-2xl border border-dashed border-gray-300 px-6 py-10 text-center text-sm text-gray-500 dark:border-dark-600 dark:text-gray-400"
          >
            {{ t('apiGuide.noEndpointVariants') }}
          </div>

          <article
            v-for="variant in variants"
            :key="variant.id"
            class="card overflow-hidden rounded-3xl border border-gray-200 dark:border-dark-600"
          >
            <div class="border-b border-gray-100 bg-gray-50/80 px-6 py-5 dark:border-dark-700 dark:bg-dark-800/60">
              <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                <div class="space-y-2">
                  <div class="inline-flex items-center gap-2 rounded-full bg-white px-3 py-1 text-xs font-semibold text-gray-700 shadow-sm dark:bg-dark-900 dark:text-gray-200">
                    <Icon :name="variant.protocol === 'google' ? 'sparkles' : variant.protocol === 'openai' ? 'cpu' : 'chat'" size="sm" />
                    <span>{{ t(`${variant.translationKey}.label`) }}</span>
                  </div>
                  <p class="text-sm leading-6 text-gray-600 dark:text-gray-300">
                    {{ t(`${variant.translationKey}.description`) }}
                  </p>
                </div>

                <router-link
                  :to="{ path: '/gateway-test', query: { key: String(selectedKey?.id || ''), variant: variant.id } }"
                  class="btn btn-secondary"
                >
                  {{ t('apiGuide.testThisVariant') }}
                </router-link>
              </div>
            </div>

            <div class="space-y-5 px-6 py-6">
              <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
                <div class="rounded-2xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-600 dark:bg-dark-800/60">
                  <div class="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
                    {{ t('apiGuide.endpoint') }}
                  </div>
                  <code class="mt-2 block break-all text-sm text-gray-900 dark:text-white">
                    {{ buildGatewayRelativePath(variant.id, variant.defaultModel) }}
                  </code>
                </div>

                <div class="rounded-2xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-600 dark:bg-dark-800/60">
                  <div class="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
                    {{ t('apiGuide.protocol') }}
                  </div>
                  <div class="mt-2 text-sm font-medium text-gray-900 dark:text-white">
                    {{ protocolLabel(variant.protocol) }}
                  </div>
                </div>

                <div class="rounded-2xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-600 dark:bg-dark-800/60">
                  <div class="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
                    {{ t('apiGuide.defaultModel') }}
                  </div>
                  <code class="mt-2 block break-all text-sm text-gray-900 dark:text-white">
                    {{ variant.defaultModel }}
                  </code>
                </div>

                <div class="rounded-2xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-600 dark:bg-dark-800/60">
                  <div class="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
                    {{ t('apiGuide.headerMode') }}
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
                      {{ t('apiGuide.curlExample') }}
                    </div>
                    <div class="text-xs text-gray-500 dark:text-gray-400">
                      {{ buildGatewayAbsoluteUrl(gatewayBaseUrl, variant.id, variant.defaultModel) }}
                    </div>
                  </div>

                  <button class="btn btn-secondary" @click="copyCommand(variant.id)">
                    {{ t('apiGuide.copyCurl') }}
                  </button>
                </div>

                <pre class="overflow-x-auto rounded-2xl bg-gray-950 p-4 text-xs leading-6 text-gray-100">{{ buildCurl(variant) }}</pre>
              </div>
            </div>
          </article>
        </section>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
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
import {
  buildGatewayAbsoluteUrl,
  buildGatewayCurlExample,
  buildGatewayRelativePath,
  getGatewayBaseUrl,
  getGatewayVariantsForApiKey
} from '@/utils/gatewayDocs'

const { t } = useI18n()
const route = useRoute()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const loading = ref(false)
const keys = ref<ApiKey[]>([])
const selectedKeyId = ref<number | null>(null)

const gatewayBaseUrl = computed(() => getGatewayBaseUrl(appStore.cachedPublicSettings?.api_base_url))

const selectedKey = computed(() => keys.value.find(key => key.id === selectedKeyId.value) ?? null)

const variants = computed(() => getGatewayVariantsForApiKey(selectedKey.value))

const keyOptions = computed(() => keys.value.map(key => ({
  value: key.id,
  label: `${key.name} · ${maskKey(key.key)}`,
  description: key.group?.name || t('apiGuide.noGroupAssigned')
})))

const authHeaderPreview = computed(() => {
  const firstGoogleVariant = variants.value.find(variant => variant.headerMode === 'x-goog-api-key')
  return firstGoogleVariant
    ? 'x-goog-api-key: <API_KEY>'
    : 'Authorization: Bearer <API_KEY>'
})

function maskKey(key: string): string {
  if (key.length <= 14) return key
  return `${key.slice(0, 8)}...${key.slice(-4)}`
}

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
    t('apiGuide.defaultPrompt'),
    false
  )
}

async function copyCommand(variantId: GatewayVariantId) {
  const variant = variants.value.find(item => item.id === variantId)
  if (!selectedKey.value || !variant) return
  await copyToClipboard(buildCurl(variant), t('apiGuide.copyCurlSuccess'))
}

async function loadKeys() {
  loading.value = true
  try {
    const response = await keysAPI.list(1, 100)
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
    appStore.showError(t('keys.failedToLoad'))
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadKeys()
})
</script>
