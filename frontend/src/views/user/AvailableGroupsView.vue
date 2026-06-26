<template>
  <AppLayout>
    <div class="space-y-6">
      <section class="card overflow-hidden">
        <div class="flex flex-col gap-6 border-b border-gray-100 px-6 py-6 dark:border-dark-700 lg:flex-row lg:items-start lg:justify-between">
          <div class="max-w-3xl">
            <div class="inline-flex items-center gap-2 rounded-full bg-primary-50 px-3 py-1 text-xs font-semibold uppercase tracking-[0.18em] text-primary-700 dark:bg-primary-900/20 dark:text-primary-300">
              <Icon name="grid" size="xs" />
              {{ availableGroupsText('title') }}
            </div>
            <h1 class="mt-3 text-2xl font-bold text-gray-900 dark:text-white">
              {{ availableGroupsText('title') }}
            </h1>
            <p class="mt-2 text-sm leading-6 text-gray-500 dark:text-dark-400">
              {{ availableGroupsText('description') }}
            </p>
          </div>

          <div class="grid gap-3 sm:grid-cols-3">
            <div class="rounded-2xl border border-gray-100 bg-gray-50/80 px-4 py-3 dark:border-dark-700 dark:bg-dark-800/60">
              <div class="text-xs font-semibold uppercase tracking-[0.16em] text-gray-500 dark:text-dark-400">
                {{ availableGroupsText('total') }}
              </div>
              <div class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">
                {{ groups.length }}
              </div>
            </div>
            <div class="rounded-2xl border border-gray-100 bg-gray-50/80 px-4 py-3 dark:border-dark-700 dark:bg-dark-800/60">
              <div class="text-xs font-semibold uppercase tracking-[0.16em] text-gray-500 dark:text-dark-400">
                {{ availableGroupsText('public') }}
              </div>
              <div class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">
                {{ publicGroups.length }}
              </div>
            </div>
            <div class="rounded-2xl border border-gray-100 bg-gray-50/80 px-4 py-3 dark:border-dark-700 dark:bg-dark-800/60">
              <div class="text-xs font-semibold uppercase tracking-[0.16em] text-gray-500 dark:text-dark-400">
                {{ availableGroupsText('memberOnly') }}
              </div>
              <div class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">
                {{ memberGroups.length }}
              </div>
            </div>
          </div>
        </div>

        <div class="px-6 py-5">
          <div class="relative w-full max-w-md">
            <Icon
              name="search"
              size="md"
              class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 dark:text-gray-500"
            />
            <input
              v-model="searchQuery"
              type="text"
              :placeholder="availableGroupsText('searchPlaceholder')"
              class="input pl-10"
            />
          </div>
        </div>
      </section>

      <div v-if="loading" class="card flex items-center justify-center py-16">
        <Icon name="refresh" size="lg" class="animate-spin text-gray-400" />
      </div>

      <EmptyState
        v-else-if="filteredGroups.length === 0"
        :title="availableGroupsText('emptyTitle')"
        :description="groups.length === 0 ? availableGroupsText('emptyDescription') : availableGroupsText('emptyFilteredDescription')"
      >
        <template #icon>
          <Icon name="grid" size="xl" class="text-gray-400" />
        </template>
      </EmptyState>

      <template v-else>
        <section v-if="filteredPublicGroups.length > 0" class="space-y-4">
          <div class="flex items-end justify-between gap-4">
            <div>
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ availableGroupsText('publicTitle') }}
              </h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
                {{ availableGroupsText('publicDescription') }}
              </p>
            </div>
            <span class="rounded-full border border-gray-200 bg-white px-3 py-1 text-xs font-medium text-gray-500 dark:border-dark-600 dark:bg-dark-800 dark:text-dark-300">
              {{ filteredPublicGroups.length }}
            </span>
          </div>

          <div class="grid gap-4 xl:grid-cols-2">
            <article
              v-for="group in filteredPublicGroups"
              :key="group.id"
              class="rounded-2xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-800"
            >
              <div class="flex items-start justify-between gap-3">
                <div class="min-w-0">
                  <GroupBadge
                    :name="group.name"
                    :platform="group.platform"
                    :subscription-type="group.subscription_type"
                    :rate-multiplier="group.rate_multiplier"
                    :user-rate-multiplier="userGroupRates[group.id] ?? null"
                    always-show-rate
                  />
                  <p v-if="group.description" class="mt-3 text-sm leading-6 text-gray-600 dark:text-dark-300">
                    {{ group.description }}
                  </p>
                </div>

                <span class="rounded-full bg-emerald-50 px-2.5 py-1 text-xs font-semibold text-emerald-700 dark:bg-emerald-900/20 dark:text-emerald-300">
                  {{ availableGroupsText('publicBadge') }}
                </span>
              </div>

              <div class="mt-4 flex flex-wrap gap-2 text-xs">
                <span class="rounded-full border border-gray-200 px-2.5 py-1 text-gray-600 dark:border-dark-600 dark:text-dark-300">
                  {{ t(`admin.groups.platforms.${group.platform}`) }}
                </span>
                <span class="rounded-full border border-gray-200 px-2.5 py-1 text-gray-600 dark:border-dark-600 dark:text-dark-300">
                  {{ subscriptionTypeLabel(group.subscription_type) }}
                </span>
                <span v-if="group.allow_image_generation" class="rounded-full border border-gray-200 px-2.5 py-1 text-gray-600 dark:border-dark-600 dark:text-dark-300">
                  {{ availableGroupsText('imageEnabledBadge') }}
                </span>
              </div>

              <dl class="mt-5 grid gap-3 text-sm sm:grid-cols-2">
                <div>
                  <dt class="text-gray-500 dark:text-dark-400">{{ availableGroupsText('rate') }}</dt>
                  <dd class="mt-1 font-medium text-gray-900 dark:text-white">
                    ×{{ userGroupRates[group.id] ?? group.rate_multiplier }}
                  </dd>
                </div>
                <div>
                  <dt class="text-gray-500 dark:text-dark-400">{{ availableGroupsText('quota') }}</dt>
                  <dd class="mt-1 font-medium text-gray-900 dark:text-white">
                    {{ quotaSummary(group) }}
                  </dd>
                </div>
              </dl>
            </article>
          </div>
        </section>

        <section v-if="filteredMemberGroups.length > 0" class="space-y-4">
          <div class="flex items-end justify-between gap-4">
            <div>
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ availableGroupsText('memberTitle') }}
              </h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
                {{ availableGroupsText('memberDescription') }}
              </p>
            </div>
            <span class="rounded-full border border-gray-200 bg-white px-3 py-1 text-xs font-medium text-gray-500 dark:border-dark-600 dark:bg-dark-800 dark:text-dark-300">
              {{ filteredMemberGroups.length }}
            </span>
          </div>

          <div class="grid gap-4 xl:grid-cols-2">
            <article
              v-for="group in filteredMemberGroups"
              :key="group.id"
              class="rounded-2xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-800"
            >
              <div class="flex items-start justify-between gap-3">
                <div class="min-w-0">
                  <GroupBadge
                    :name="group.name"
                    :platform="group.platform"
                    :subscription-type="group.subscription_type"
                    :rate-multiplier="group.rate_multiplier"
                    :user-rate-multiplier="userGroupRates[group.id] ?? null"
                    always-show-rate
                  />
                  <p v-if="group.description" class="mt-3 text-sm leading-6 text-gray-600 dark:text-dark-300">
                    {{ group.description }}
                  </p>
                </div>

                <span
                  class="rounded-full px-2.5 py-1 text-xs font-semibold"
                  :class="
                    group.subscription_type === 'subscription'
                      ? 'bg-violet-50 text-violet-700 dark:bg-violet-900/20 dark:text-violet-300'
                      : 'bg-amber-50 text-amber-700 dark:bg-amber-900/20 dark:text-amber-300'
                  "
                >
                  {{
                    group.subscription_type === 'subscription'
                      ? availableGroupsText('subscriptionBadge')
                      : availableGroupsText('exclusiveBadge')
                  }}
                </span>
              </div>

              <div class="mt-4 flex flex-wrap gap-2 text-xs">
                <span class="rounded-full border border-gray-200 px-2.5 py-1 text-gray-600 dark:border-dark-600 dark:text-dark-300">
                  {{ t(`admin.groups.platforms.${group.platform}`) }}
                </span>
                <span class="rounded-full border border-gray-200 px-2.5 py-1 text-gray-600 dark:border-dark-600 dark:text-dark-300">
                  {{ subscriptionTypeLabel(group.subscription_type) }}
                </span>
                <span
                  v-if="group.is_exclusive"
                  class="rounded-full border border-gray-200 px-2.5 py-1 text-gray-600 dark:border-dark-600 dark:text-dark-300"
                >
                  {{ availableGroupsText('exclusiveBadge') }}
                </span>
              </div>

              <dl class="mt-5 grid gap-3 text-sm sm:grid-cols-2">
                <div>
                  <dt class="text-gray-500 dark:text-dark-400">{{ availableGroupsText('rate') }}</dt>
                  <dd class="mt-1 font-medium text-gray-900 dark:text-white">
                    ×{{ userGroupRates[group.id] ?? group.rate_multiplier }}
                  </dd>
                </div>
                <div>
                  <dt class="text-gray-500 dark:text-dark-400">{{ availableGroupsText('quota') }}</dt>
                  <dd class="mt-1 font-medium text-gray-900 dark:text-white">
                    {{ quotaSummary(group) }}
                  </dd>
                </div>
              </dl>
            </article>
          </div>
        </section>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { resolveRuntimeLocale } from '@/utils/runtimeLocale'
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import userGroupsAPI from '@/api/groups'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import {
  renderAvailableGroupsShellText,
  resolveAvailableGroupsShellLabels,
  type AvailableGroupsLabelKey,
} from '@/utils/availableGroupsShell'
import type { Group, SubscriptionType } from '@/types'
import {
  filterAvailableGroupsByQuery,
  resolveAvailableGroupQuotaSummary,
  resolveAvailableGroupSubscriptionLabel,
  resolveMemberAvailableGroups,
  resolvePublicAvailableGroups,
} from './availableGroupsRuntime'

const { t, locale } = useI18n()
const appStore = useAppStore()


const availableGroupsLabels = computed(() =>
  resolveAvailableGroupsShellLabels(
    appStore.cachedPublicSettings?.available_groups_shell_config,
    resolveRuntimeLocale(locale),
  ),
)

function availableGroupsText(key: AvailableGroupsLabelKey, values?: Record<string, string | number>): string {
  return renderAvailableGroupsShellText(availableGroupsLabels.value, key, values)
}

const loading = ref(false)
const searchQuery = ref('')
const groups = ref<Group[]>([])
const userGroupRates = ref<Record<number, number>>({})

const filteredGroups = computed(() => filterAvailableGroupsByQuery(groups.value, searchQuery.value))

const publicGroups = computed(() => resolvePublicAvailableGroups(groups.value))

const memberGroups = computed(() => resolveMemberAvailableGroups(groups.value))

const filteredPublicGroups = computed(() => resolvePublicAvailableGroups(filteredGroups.value))

const filteredMemberGroups = computed(() => resolveMemberAvailableGroups(filteredGroups.value))

function subscriptionTypeLabel(type: SubscriptionType): string {
  return resolveAvailableGroupSubscriptionLabel(type, availableGroupsText)
}

function quotaSummary(group: Group): string {
  return resolveAvailableGroupQuotaSummary(group, availableGroupsText)
}

async function loadGroups() {
  loading.value = true
  try {
    const [list, rates] = await Promise.all([
      userGroupsAPI.getAvailable(),
      userGroupsAPI.getUserGroupRates().catch(() => ({} as Record<number, number>)),
    ])
    groups.value = list
    userGroupRates.value = rates
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, availableGroupsText('loadFailed')))
  } finally {
    loading.value = false
  }
}

onMounted(loadGroups)
</script>
