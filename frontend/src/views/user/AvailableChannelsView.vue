<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-col justify-between gap-4 lg:flex-row lg:items-start">
          <div class="flex flex-1 flex-wrap items-center gap-3">
            <div class="relative w-full sm:w-80">
              <Icon
                name="search"
                size="md"
                class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 dark:text-gray-500"
              />
              <input
                v-model="searchQuery"
                type="text"
                :placeholder="availableChannelsText('searchPlaceholder')"
                class="input pl-10"
              />
            </div>
          </div>

          <div class="flex w-full flex-shrink-0 flex-wrap items-center justify-end gap-3 lg:w-auto">
            <button
              @click="loadChannels"
              :disabled="loading"
              class="btn btn-secondary"
              :title="availableChannelsText('refreshTitle')"
            >
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            </button>
          </div>
        </div>
      </template>

      <template #table>
        <AvailableChannelsTable
          :columns="columnLabels"
          :rows="filteredChannels"
          :loading="loading"
          :user-group-rates="userGroupRates"
          :pricing-labels="pricingLabels"
          :no-pricing-label="availableChannelsText('noPricing')"
          :no-models-label="availableChannelsText('noModels')"
          :empty-label="availableChannelsText('empty')"
          :exclusive-label="availableChannelsText('exclusive')"
          :exclusive-tooltip-label="availableChannelsText('exclusiveTooltip')"
          :public-label="availableChannelsText('public')"
          :public-tooltip-label="availableChannelsText('publicTooltip')"
        />
      </template>
    </TablePageLayout>
  </AppLayout>
</template>

<script setup lang="ts">
import { resolveRuntimeLocale } from '@/utils/runtimeLocale'
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import AvailableChannelsTable from '@/components/channels/AvailableChannelsTable.vue'
import userChannelsAPI, { type UserAvailableChannel } from '@/api/channels'
import userGroupsAPI from '@/api/groups'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import {
  resolveConfiguredAvailableChannelsShellLabels,
  type AvailableChannelsLabelKey,
} from '@/utils/availableChannelsShell'
import {
  buildAvailableChannelsColumnLabels,
  buildAvailableChannelsPricingLabels,
  filterAvailableChannelsByQuery,
} from './availableChannelsRuntime'

const { locale } = useI18n()
const appStore = useAppStore()


const availableChannelsShellLabels = computed(() =>
  resolveConfiguredAvailableChannelsShellLabels(
    appStore.cachedPublicSettings?.available_channels_shell_config,
    resolveRuntimeLocale(locale),
  ),
)

function availableChannelsText(key: AvailableChannelsLabelKey): string {
  return availableChannelsShellLabels.value[key]
}

const channels = ref<UserAvailableChannel[]>([])
const userGroupRates = ref<Record<number, number>>({})
const loading = ref(false)
const searchQuery = ref('')

const columnLabels = computed(() => buildAvailableChannelsColumnLabels(availableChannelsText))

const pricingLabels = computed(() => buildAvailableChannelsPricingLabels(availableChannelsText))

/**
 * 搜索过滤：
 * - 命中渠道名/描述 → 整个渠道（所有 platforms）都保留
 * - 否则按 platform/group/model 维度在 sections 里过滤，保留有匹配的 section
 * - 所有 sections 都不匹配时，渠道本身被过滤掉
 */
const filteredChannels = computed(() => filterAvailableChannelsByQuery(channels.value, searchQuery.value))

async function loadChannels() {
  loading.value = true
  try {
    // 渠道列表和用户专属倍率并发拉取。专属倍率失败不阻塞渠道展示——
    // 失败时只是无法渲染专属倍率角标，降级为仅显示默认倍率。
    const [list, rates] = await Promise.all([
      userChannelsAPI.getAvailable(),
      userGroupsAPI.getUserGroupRates().catch((err: unknown) => {
        console.error('Failed to load user group rates:', err)
        return {} as Record<number, number>
      }),
    ])
    channels.value = list
    userGroupRates.value = rates
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, availableChannelsText('loadError')))
  } finally {
    loading.value = false
  }
}

onMounted(loadChannels)
</script>
