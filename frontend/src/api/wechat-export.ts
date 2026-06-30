import { apiClient } from './client'

export type WeChatExportFormat = 'html' | 'markdown'

export interface WeChatSession {
  id?: number
  status: string
  login_token?: string
  login_account_name?: string
  expires_at?: string
  updated_at?: string
}

export interface WeChatQRCodeSessionResponse {
  session: WeChatSession
  qrcode_url: string
}

export interface WeChatAccount {
  id: number
  fakeid: string
  nickname: string
  alias: string
  avatar: string
  description: string
  is_active: boolean
  last_synced_at?: string
}

export interface WeChatAccountSyncResult {
  fakeid: string
  synced_count: number
  page_count: number
  total_count: number
  has_more: boolean
}

export interface WeChatArticle {
  id: number
  account_fakeid: string
  title: string
  link: string
  source_type: string
  content_status: string
  author?: string
  publish_at?: string
  created_at: string
}

export interface WeChatExportTask {
  id: number
  status: string
  article_ids: number[]
  formats: WeChatExportFormat[]
  selected_article_count: number
  successful_article_count: number
  failed_article_count: number
  result_manifest_json?: string
  error_message: string
  worker_lease_until?: string
  expires_at?: string
  created_at: string
  updated_at: string
}

export interface WeChatExportWorkerStatus {
  health: 'idle' | 'waiting' | 'active' | 'attention' | string
  message: string
  total_count: number
  queued_count: number
  running_count: number
  stale_running_count: number
  failed_count: number
  completed_count: number
  cancelled_count: number
  last_task_updated_at?: string
  oldest_queued_at?: string
  last_task_age_seconds?: number
  oldest_queued_seconds?: number
  attention_reasons?: string[]
}

export interface WeChatExportArtifact {
  id: number
  task_id: number
  format: WeChatExportFormat
  file_name: string
  file_size: number
  download_url: string
}

export interface WeChatExportTaskLog {
  id: number
  task_id: number
  user_id: number
  event: string
  status: string
  message: string
  meta_json: string
  created_at: string
}

export interface Paginated<T> {
  items: T[]
  total: number
  page: number
  page_size: number
  pages: number
}

export async function getWeChatSession() {
  const { data } = await apiClient.get<WeChatSession>('/wechat/session')
  return data
}

export async function createWeChatQRCodeSession() {
  const { data } = await apiClient.post<WeChatQRCodeSessionResponse>('/wechat/session/qrcode')
  return data
}

export async function pollWeChatSession(sessionId: number) {
  const { data } = await apiClient.get<WeChatSession>(`/wechat/session/poll/${sessionId}`)
  return data
}

export async function logoutWeChatSession() {
  const { data } = await apiClient.post<{ ok: boolean }>('/wechat/session/logout')
  return data
}

export async function validateWeChatSession() {
  const { data } = await apiClient.post<WeChatSession>('/wechat/session/validate')
  return data
}

export async function searchWeChatAccounts(q = '', remote = false) {
  const { data } = await apiClient.get<{ items: WeChatAccount[]; query: string }>('/wechat/accounts/search', {
    params: { q, limit: remote ? 5 : 20, remote },
  })
  return data.items
}

export async function bindWeChatAccount(payload: {
  fakeid: string
  nickname?: string
  alias?: string
  avatar?: string
  description?: string
}) {
  const { data } = await apiClient.post<{ account: WeChatAccount; sync_required: boolean }>('/wechat/accounts/bind', payload)
  return data
}

export async function syncWeChatAccount(fakeid: string, beginFrom?: number) {
  const params: Record<string, any> = {}
  if (beginFrom !== undefined && beginFrom > 0) {
    params.begin = beginFrom
  }
  const { data } = await apiClient.post<{ account: WeChatAccount; status: string; result: WeChatAccountSyncResult }>(
    `/wechat/accounts/${encodeURIComponent(fakeid)}/sync`,
    null,
    { params }
  )
  return data
}

export async function importWeChatArticleLink(link: string) {
  const { data } = await apiClient.post<WeChatArticle>('/wechat/articles/import-link', { link })
  return data
}

export async function listWeChatArticles(params: { page?: number; page_size?: number } = {}) {
  const { data } = await apiClient.get<Paginated<WeChatArticle>>('/wechat/articles', {
    params: {
      page: params.page ?? 1,
      page_size: params.page_size ?? 100,
    },
  })
  return data
}

export async function createWeChatExportTask(payload: {
  article_ids: number[]
  formats: WeChatExportFormat[]
  include_engagement: boolean
}) {
  const { data } = await apiClient.post<WeChatExportTask>('/wechat/tasks', payload)
  return data
}

export async function quoteWeChatExportTask(payload: {
  article_ids: number[]
  formats: WeChatExportFormat[]
  include_engagement: boolean
}) {
  const { data } = await apiClient.post<{
    article_count: number
    formats: WeChatExportFormat[]
    include_engagement: boolean
    estimated_credits: number
  }>('/wechat/tasks/quote', payload)
  return data
}

export async function listWeChatExportTasks(params: { page?: number; page_size?: number } = {}) {
  const { data } = await apiClient.get<Paginated<WeChatExportTask>>('/wechat/tasks', {
    params: {
      page: params.page ?? 1,
      page_size: params.page_size ?? 20,
    },
  })
  return data
}

export async function getWeChatExportWorkerStatus() {
  const { data } = await apiClient.get<WeChatExportWorkerStatus>('/wechat/worker/status')
  return data
}

export async function cancelWeChatExportTask(taskId: number) {
  const { data } = await apiClient.post<WeChatExportTask>(`/wechat/tasks/${taskId}/cancel`)
  return data
}

export async function retryWeChatExportTask(taskId: number) {
  const { data } = await apiClient.post<WeChatExportTask>(`/wechat/tasks/${taskId}/retry`)
  return data
}

export async function listWeChatExportTaskLogs(taskId: number) {
  const { data } = await apiClient.get<{ items: WeChatExportTaskLog[] }>(`/wechat/tasks/${taskId}/logs`)
  return data.items
}

export async function listWeChatExportArtifacts(taskId: number) {
  const { data } = await apiClient.get<{ items: WeChatExportArtifact[] }>(`/wechat/tasks/${taskId}/artifacts`)
  return data.items
}

export async function downloadWeChatExportArtifact(artifactId: number) {
  return downloadWeChatExportBlob(`/wechat/artifacts/${artifactId}/download`)
}

export async function downloadWeChatExportTaskZip(taskId: number) {
  return downloadWeChatExportBlob(`/wechat/tasks/${taskId}/artifacts.zip`)
}

async function downloadWeChatExportBlob(path: string) {
  try {
    const token = typeof window === 'undefined' ? '' : localStorage.getItem('auth_token')
    const { data } = await apiClient.get<Blob>(path, {
      responseType: 'blob',
      headers: token ? { Authorization: `Bearer ${token}` } : undefined,
    })
    return data
  } catch (error) {
    throw await normalizeWeChatExportDownloadError(error)
  }
}

async function normalizeWeChatExportDownloadError(error: unknown) {
  const response = (error as { response?: { status?: number; data?: unknown }; status?: number; message?: string })?.response
  const status = response?.status ?? (error as { status?: number })?.status
  const data = response?.data
  if (data instanceof Blob) {
    const text = await data.text().catch(() => '')
    if (text) {
      try {
        const parsed = JSON.parse(text) as { message?: string; code?: string | number }
        return {
          status,
          code: parsed.code,
          message: parsed.message || downloadFallbackMessage(status),
        }
      } catch {
        return { status, message: text }
      }
    }
  }
  return {
    status,
    message: (error as { message?: string })?.message || downloadFallbackMessage(status),
  }
}

function downloadFallbackMessage(status?: number) {
  if (status === 401) return '登录已失效，请重新登录后再下载导出产物。'
  if (status === 403) return '没有权限下载这个导出产物。'
  if (status === 404) return '导出产物不存在或已过期。'
  return '导出产物下载失败，请稍后重试。'
}
