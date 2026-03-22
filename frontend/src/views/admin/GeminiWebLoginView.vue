<template>
  <div class="space-y-6">
    <section class="rounded-3xl border border-blue-200/70 bg-gradient-to-br from-blue-50 via-white to-cyan-50 p-6 shadow-sm dark:border-blue-900/50 dark:from-dark-900 dark:via-dark-950 dark:to-dark-900">
      <div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div class="max-w-3xl space-y-3">
          <p class="text-xs font-semibold uppercase tracking-[0.24em] text-blue-600 dark:text-blue-300">
            Gemini Command Gateway
          </p>
          <h1 class="text-3xl font-semibold text-gray-900 dark:text-white">
            {{ t('admin.geminiWebLogin.title') }}
          </h1>
          <p class="text-sm leading-6 text-gray-600 dark:text-gray-300">
            {{ t('admin.geminiWebLogin.description') }}
          </p>
        </div>
        <div class="flex flex-wrap gap-3">
          <button
            class="inline-flex items-center rounded-xl border border-gray-200 bg-white px-4 py-2 text-sm font-medium text-gray-700 transition hover:border-blue-300 hover:text-blue-700 dark:border-dark-700 dark:bg-dark-900 dark:text-dark-200 dark:hover:border-blue-500 dark:hover:text-blue-300"
            :disabled="loading"
            @click="loadAccounts"
          >
            {{ t('admin.geminiWebLogin.refreshAll') }}
          </button>
          <router-link
            to="/admin/accounts"
            class="inline-flex items-center rounded-xl bg-gray-900 px-4 py-2 text-sm font-medium text-white transition hover:bg-gray-800 dark:bg-white dark:text-gray-900 dark:hover:bg-gray-200"
          >
            {{ t('admin.geminiWebLogin.manageAccounts') }}
          </router-link>
        </div>
      </div>
    </section>

    <section class="rounded-2xl border border-amber-200/70 bg-amber-50/80 p-5 dark:border-amber-900/40 dark:bg-amber-950/30">
      <h2 class="text-sm font-semibold text-amber-900 dark:text-amber-200">
        {{ t('admin.geminiWebLogin.workflowTitle') }}
      </h2>
      <ol class="mt-3 space-y-2 text-sm text-amber-800 dark:text-amber-100">
        <li>1. {{ t('admin.geminiWebLogin.workflowStepCreate') }}</li>
        <li>2. {{ t('admin.geminiWebLogin.workflowStepOpen') }}</li>
        <li>3. {{ t('admin.geminiWebLogin.workflowStepImport') }}</li>
      </ol>
    </section>

    <section v-if="loading" class="rounded-2xl border border-gray-200 bg-white p-8 text-sm text-gray-500 dark:border-dark-700 dark:bg-dark-900 dark:text-dark-300">
      {{ t('common.loading') }}
    </section>

    <section
      v-else-if="accounts.length === 0"
      class="rounded-2xl border border-dashed border-gray-300 bg-white p-8 text-center dark:border-dark-700 dark:bg-dark-900"
    >
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
        {{ t('admin.geminiWebLogin.emptyTitle') }}
      </h2>
      <p class="mx-auto mt-2 max-w-2xl text-sm text-gray-600 dark:text-dark-300">
        {{ t('admin.geminiWebLogin.emptyDescription') }}
      </p>
      <router-link
        to="/admin/accounts"
        class="mt-5 inline-flex items-center rounded-xl bg-blue-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-blue-500"
      >
        {{ t('admin.geminiWebLogin.goCreateAccount') }}
      </router-link>
    </section>

    <section v-else class="grid gap-5">
      <article
        v-for="account in accounts"
        :key="account.id"
        class="rounded-3xl border border-gray-200 bg-white p-6 shadow-sm dark:border-dark-700 dark:bg-dark-900"
      >
        <div class="flex flex-col gap-5 lg:flex-row lg:items-start lg:justify-between">
          <div class="space-y-3">
            <div class="flex flex-wrap items-center gap-3">
              <h2 class="text-xl font-semibold text-gray-900 dark:text-white">
                {{ account.name }}
              </h2>
              <span
                class="inline-flex items-center rounded-full px-3 py-1 text-xs font-semibold"
                :class="statusBadgeClass(sessionMap[account.id]?.status)"
              >
                {{ sessionMap[account.id]?.status || 'idle' }}
              </span>
            </div>
            <p class="text-sm text-gray-600 dark:text-dark-300">
              {{ sessionMap[account.id]?.message || t('admin.geminiWebLogin.notStarted') }}
            </p>
            <div class="grid gap-2 text-xs text-gray-500 dark:text-dark-400">
              <div>
                <span class="font-medium text-gray-700 dark:text-dark-200">{{ t('admin.geminiWebLogin.gatewayUrl') }}:</span>
                {{ sessionMap[account.id]?.gateway_url || t('common.notSet') }}
              </div>
              <div>
                <span class="font-medium text-gray-700 dark:text-dark-200">{{ t('admin.geminiWebLogin.loginId') }}:</span>
                {{ sessionMap[account.id]?.login_id || '-' }}
              </div>
              <div>
                <span class="font-medium text-gray-700 dark:text-dark-200">{{ t('admin.geminiWebLogin.updatedAt') }}:</span>
                {{ formatTimestamp(sessionMap[account.id]?.updated_at) }}
              </div>
              <div>
                <span class="font-medium text-gray-700 dark:text-dark-200">{{ t('admin.geminiWebLogin.loginMode') }}:</span>
                {{ sessionMap[account.id]?.login_mode || 'auto' }}
              </div>
            </div>
          </div>

          <div class="flex flex-wrap gap-3">
            <button
              class="rounded-xl bg-blue-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-blue-500 disabled:cursor-not-allowed disabled:opacity-60"
              :disabled="busyIds.has(account.id)"
              @click="startLogin(account.id)"
            >
              {{ t('admin.geminiWebLogin.startLogin') }}
            </button>
            <button
              class="rounded-xl border border-gray-200 px-4 py-2 text-sm font-medium text-gray-700 transition hover:border-blue-300 hover:text-blue-700 disabled:cursor-not-allowed disabled:opacity-60 dark:border-dark-700 dark:text-dark-200 dark:hover:border-blue-500 dark:hover:text-blue-300"
              :disabled="busyIds.has(account.id)"
              @click="refreshStatus(account.id)"
            >
              {{ t('admin.geminiWebLogin.refreshStatus') }}
            </button>
            <button
              class="rounded-xl border border-gray-200 px-4 py-2 text-sm font-medium text-gray-700 transition hover:border-blue-300 hover:text-blue-700 disabled:cursor-not-allowed disabled:opacity-60 dark:border-dark-700 dark:text-dark-200 dark:hover:border-blue-500 dark:hover:text-blue-300"
              :disabled="!sessionMap[account.id]?.login_url"
              @click="openLoginUrl(account.id)"
            >
              {{ t('admin.geminiWebLogin.openLoginPage') }}
            </button>
          </div>
        </div>

        <div class="mt-6 grid gap-3">
          <label class="text-sm font-medium text-gray-900 dark:text-white">
            {{ t('admin.geminiWebLogin.cookiesLabel') }}
          </label>
          <textarea
            v-model="cookiesInputs[account.id]"
            rows="8"
            class="w-full rounded-2xl border border-gray-200 bg-gray-50 px-4 py-3 text-sm text-gray-800 outline-none transition focus:border-blue-400 focus:bg-white dark:border-dark-700 dark:bg-dark-950 dark:text-dark-100 dark:focus:border-blue-500"
            :placeholder="t('admin.geminiWebLogin.cookiesPlaceholder')"
          ></textarea>
          <div class="flex flex-wrap items-center justify-between gap-3">
            <p class="text-xs text-gray-500 dark:text-dark-400">
              {{ t('admin.geminiWebLogin.cookiesHint') }}
            </p>
            <button
              class="rounded-xl bg-gray-900 px-4 py-2 text-sm font-medium text-white transition hover:bg-gray-800 disabled:cursor-not-allowed disabled:opacity-60 dark:bg-white dark:text-gray-900 dark:hover:bg-gray-200"
              :disabled="busyIds.has(account.id) || !cookiesInputs[account.id]?.trim()"
              @click="importCookies(account.id)"
            >
              {{ t('admin.geminiWebLogin.importCookies') }}
            </button>
          </div>
        </div>
      </article>
    </section>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api'
import { useAppStore } from '@/stores'
import type { Account, GeminiWebSessionResponse } from '@/types'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const accounts = ref<Account[]>([])
const sessionMap = reactive<Record<number, GeminiWebSessionResponse | null>>({})
const cookiesInputs = reactive<Record<number, string>>({})
const busyIds = ref<Set<number>>(new Set())
const pollTimers = new Map<number, number>()

function setBusy(id: number, busy: boolean) {
  const next = new Set(busyIds.value)
  if (busy) {
    next.add(id)
  } else {
    next.delete(id)
  }
  busyIds.value = next
}

function statusBadgeClass(status?: string) {
  switch ((status || '').toLowerCase()) {
    case 'ready':
      return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-300'
    case 'error':
      return 'bg-rose-100 text-rose-700 dark:bg-rose-950/40 dark:text-rose-300'
    case 'pending':
      return 'bg-amber-100 text-amber-700 dark:bg-amber-950/40 dark:text-amber-300'
    case 'waiting_import':
      return 'bg-blue-100 text-blue-700 dark:bg-blue-950/40 dark:text-blue-300'
    default:
      return 'bg-gray-100 text-gray-700 dark:bg-dark-800 dark:text-dark-200'
  }
}

function formatTimestamp(value?: string) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

function clearStatusPoll(id?: number) {
  if (typeof id === 'number') {
    const timer = pollTimers.get(id)
    if (timer) {
      window.clearTimeout(timer)
      pollTimers.delete(id)
    }
    return
  }

  for (const timer of pollTimers.values()) {
    window.clearTimeout(timer)
  }
  pollTimers.clear()
}

function scheduleStatusPoll(id: number) {
  clearStatusPoll(id)
  const status = (sessionMap[id]?.status || '').toLowerCase()
  if (status !== 'pending') {
    return
  }

  const timer = window.setTimeout(async () => {
    await refreshStatus(id, false)
  }, 3000)
  pollTimers.set(id, timer)
}

async function loadAccounts() {
  loading.value = true
  try {
    const result = await adminAPI.accounts.list(1, 100, { type: 'gemini-web' })
    accounts.value = result.items
    await Promise.all(accounts.value.map((account) => refreshStatus(account.id, false)))
  } catch (error) {
    console.error('Failed to load Gemini Web accounts:', error)
    appStore.showError(t('admin.geminiWebLogin.loadFailed'))
  } finally {
    loading.value = false
  }
}

async function refreshStatus(id: number, withToast: boolean = false) {
  setBusy(id, true)
  try {
    const session = await adminAPI.accounts.getGeminiWebLoginStatus(id)
    sessionMap[id] = session
    scheduleStatusPoll(id)
    if (withToast) {
      appStore.showSuccess(t('admin.geminiWebLogin.statusRefreshed'))
    }
  } catch (error) {
    console.error(`Failed to refresh Gemini Web status for account ${id}:`, error)
    appStore.showError(t('admin.geminiWebLogin.statusFailed'))
  } finally {
    setBusy(id, false)
  }
}

async function startLogin(id: number) {
  const loginWindow = window.open('about:blank', '_blank')
  setBusy(id, true)
  try {
    const session = await adminAPI.accounts.startGeminiWebLogin(id, { login_mode: 'remote' })
    sessionMap[id] = session
    scheduleStatusPoll(id)
    appStore.showSuccess(t('admin.geminiWebLogin.startSuccess'))
    if (session.login_url) {
      if (loginWindow) {
        loginWindow.location.href = session.login_url
      } else {
        window.open(session.login_url, '_blank', 'noopener,noreferrer')
      }
    } else {
      loginWindow?.close()
    }
  } catch (error) {
    loginWindow?.close()
    console.error(`Failed to start Gemini Web login for account ${id}:`, error)
    appStore.showError(t('admin.geminiWebLogin.startFailed'))
  } finally {
    setBusy(id, false)
  }
}

function openLoginUrl(id: number) {
  const loginUrl = sessionMap[id]?.login_url
  if (!loginUrl) return
  window.open(loginUrl, '_blank', 'noopener,noreferrer')
}

async function importCookies(id: number) {
  const cookiesJSON = cookiesInputs[id]?.trim()
  if (!cookiesJSON) {
    appStore.showError(t('admin.geminiWebLogin.cookiesRequired'))
    return
  }

  setBusy(id, true)
  try {
    const session = await adminAPI.accounts.importGeminiWebCookies(id, cookiesJSON)
    sessionMap[id] = session
    clearStatusPoll(id)
    appStore.showSuccess(t('admin.geminiWebLogin.importSuccess'))
    await refreshStatus(id)
  } catch (error) {
    console.error(`Failed to import Gemini Web cookies for account ${id}:`, error)
    appStore.showError(t('admin.geminiWebLogin.importFailed'))
  } finally {
    setBusy(id, false)
  }
}

onMounted(() => {
  loadAccounts()
})

onBeforeUnmount(() => {
  clearStatusPoll()
})
</script>
