/**
 * Admin Channel Monitor API endpoints
 * Handles channel monitor (uptime/health) management for administrators
 */

import { apiClient } from '../client'

export type Provider = 'openai' | 'anthropic' | 'gemini'
export type MonitorStatus = 'operational' | 'degraded' | 'failed' | 'error'
export type MonitorHealthStatus = 'unknown' | 'healthy' | 'degraded' | 'unhealthy'
export type MonitorErrorCategory =
  | ''
  | 'auth'
  | 'rate_limit'
  | 'quota'
  | 'server'
  | 'network'
  | 'timeout'
  | 'challenge'
  | 'slow'
  | 'empty_response'
  | 'invalid_request'
  | 'unknown'
export type BodyOverrideMode = 'off' | 'merge' | 'replace'

export interface ChannelMonitor {
  id: number
  name: string
  provider: Provider
  endpoint: string
  api_key_masked: string
  /**
   * True when the stored encrypted API key cannot be decrypted (e.g. the
   * encryption key has changed). Admin must re-edit the monitor to provide
   * a fresh key. Backend skips checks for these monitors.
   */
  api_key_decrypt_failed?: boolean
  primary_model: string
  extra_models: string[]
  group_name: string
  enabled: boolean
  auto_disabled: boolean
  auto_disabled_at: string | null
  auto_disabled_reason: string
  auto_recovered_at: string | null
  health_status: MonitorHealthStatus
  interval_seconds: number
  last_checked_at: string | null
  created_by: number
  created_at: string
  updated_at: string
  /** Latest status of the primary model (empty when no history yet) */
  primary_status: MonitorStatus | ''
  /** Latest latency of the primary model in ms (null when no history yet) */
  primary_latency_ms: number | null
  /** Primary model 7-day availability percentage (0-100) */
  availability_7d: number
  /** Latest status per extra model (used for hover tooltip) */
  extra_models_status: ExtraModelStatus[]
  /** 请求自定义快照字段（高级设置） */
  template_id: number | null
  extra_headers: Record<string, string>
  body_override_mode: BodyOverrideMode
  body_override: Record<string, unknown> | null
}

export interface ExtraModelStatus {
  model: string
  status: MonitorStatus | ''
  latency_ms: number | null
}

export interface ListParams {
  page?: number
  page_size?: number
  provider?: Provider
  enabled?: boolean
  search?: string
}

export interface ListResponse {
  items: ChannelMonitor[]
  total: number
  page: number
  page_size: number
  pages: number
}

export interface CreateParams {
  name: string
  provider: Provider
  endpoint: string
  api_key: string
  primary_model: string
  extra_models?: string[]
  group_name?: string
  enabled?: boolean
  interval_seconds: number
  template_id?: number | null
  extra_headers?: Record<string, string>
  body_override_mode?: BodyOverrideMode
  body_override?: Record<string, unknown> | null
}

// Update request: api_key 空串 = 不修改；clear_template=true 时把 template_id 置空
export type UpdateParams = Partial<CreateParams> & {
  clear_template?: boolean
}

export interface CheckResult {
  model: string
  status: MonitorStatus
  latency_ms: number | null
  ping_latency_ms: number | null
  error_category: MonitorErrorCategory
  message: string
  checked_at: string
}

export interface ErrorCategoryCount {
  category: MonitorErrorCategory
  count: number
}

export interface HealthSnapshot {
  monitor_id: number
  health_status: MonitorHealthStatus
  auto_disabled: boolean
  auto_disabled_at: string | null
  auto_disabled_reason: string
  auto_recovered_at: string | null
  window_minutes: number
  total_checks: number
  successful_checks: number
  success_rate_pct: number
  avg_latency_ms: number | null
  consecutive_failed_runs: number
  consecutive_successful_runs: number
  top_error_categories: ErrorCategoryCount[]
  latest_status: MonitorStatus | ''
  latest_error_category: MonitorErrorCategory
  latest_message: string
  latest_checked_at: string | null
}

export interface RunNowResponse {
  results: CheckResult[]
  health?: HealthSnapshot
}

export interface HistoryItem {
  id: number
  model: string
  status: MonitorStatus
  latency_ms: number | null
  ping_latency_ms: number | null
  error_category: MonitorErrorCategory
  message: string
  checked_at: string
}

export interface HistoryParams {
  model?: string
  limit?: number
}

export interface HistoryResponse {
  items: HistoryItem[]
}

/**
 * List channel monitors with pagination and filters
 */
export async function list(
  params: ListParams = {},
  options?: { signal?: AbortSignal }
): Promise<ListResponse> {
  const { data } = await apiClient.get<ListResponse>('/admin/channel-monitors', {
    params,
    signal: options?.signal,
  })
  return data
}

/**
 * Get a channel monitor by ID
 */
export async function get(id: number): Promise<ChannelMonitor> {
  const { data } = await apiClient.get<ChannelMonitor>(`/admin/channel-monitors/${id}`)
  return data
}

/**
 * Create a new channel monitor
 */
export async function create(params: CreateParams): Promise<ChannelMonitor> {
  const { data } = await apiClient.post<ChannelMonitor>('/admin/channel-monitors', params)
  return data
}

/**
 * Update an existing channel monitor.
 * api_key field: empty string means "do not modify".
 */
export async function update(id: number, params: UpdateParams): Promise<ChannelMonitor> {
  const { data } = await apiClient.put<ChannelMonitor>(`/admin/channel-monitors/${id}`, params)
  return data
}

/**
 * Delete a channel monitor
 */
export async function del(id: number): Promise<void> {
  await apiClient.delete(`/admin/channel-monitors/${id}`)
}

/**
 * Trigger an immediate manual check for a channel monitor.
 * Returns the latest check results for primary + extra models.
 */
export async function runNow(id: number): Promise<RunNowResponse> {
  const { data } = await apiClient.post<RunNowResponse>(`/admin/channel-monitors/${id}/run`)
  return data
}

/**
 * Get monitor health snapshot for ops troubleshooting.
 */
export async function health(id: number): Promise<HealthSnapshot> {
  const { data } = await apiClient.get<HealthSnapshot>(`/admin/channel-monitors/${id}/health`)
  return data
}

/**
 * List historical check results for a monitor.
 */
export async function listHistory(
  id: number,
  params: HistoryParams = {}
): Promise<HistoryResponse> {
  const { data } = await apiClient.get<HistoryResponse>(
    `/admin/channel-monitors/${id}/history`,
    { params }
  )
  return data
}

export const channelMonitorAPI = {
  list,
  get,
  create,
  update,
  del,
  runNow,
  health,
  listHistory,
}

export default channelMonitorAPI
