/**
 * Admin Users API endpoints
 * Handles user management for administrators
 */

import { apiClient } from '../client'
import type { AdminUser, UpdateUserRequest, PaginatedResponse, ApiKey } from '@/types'

export type AdminUserResourceID = string | number

export function userResourcePath(id: AdminUserResourceID): string {
  return `/admin/users/${encodeURIComponent(String(id))}`
}

export interface AdminBindAuthIdentityChannelRequest {
  channel: string
  channel_app_id: string
  channel_subject: string
  metadata?: Record<string, unknown> | null
}

export interface AdminBindAuthIdentityRequest {
  provider_type: string
  provider_key: string
  provider_subject: string
  issuer?: string | null
  metadata?: Record<string, unknown> | null
  channel?: AdminBindAuthIdentityChannelRequest
}

export interface AdminBoundAuthIdentityChannel {
  channel: string
  channel_app_id: string
  channel_subject: string
  metadata: Record<string, unknown> | null
  created_at: string
  updated_at: string
}

export interface AdminBoundAuthIdentity {
  user_id: number
  provider_type: string
  provider_key: string
  provider_subject: string
  verified_at?: string | null
  issuer?: string | null
  metadata: Record<string, unknown> | null
  created_at: string
  updated_at: string
  channel?: AdminBoundAuthIdentityChannel | null
}

export interface SignupGrantRiskClaimRecord {
  id: number
  user_id?: number | null
  user_public_id?: string
  email: string
  email_domain: string
  ip_address: string
  email_hash: string
  email_domain_hash: string
  ip_hash: string
  device_hash: string
  signup_source: string
  provider_type: string
  provider_subject: string
  provider_subject_hash: string
  decision: string
  reason: string
  grant_balance: number
  created_at: string
  updated_at: string
}

export interface SignupGrantRiskClaimsResponse {
  items: SignupGrantRiskClaimRecord[]
  total: number
  page: number
  size: number
}

export interface SignupGrantRiskClaimFilters {
  decision?: string
  user_id?: number | string
  subject_type?: 'email' | 'email_domain' | 'ip' | 'oauth_identity' | 'device' | ''
  subject?: string
  reason?: string
}

export interface SignupGrantRiskUserSummary {
  user_id: number
  has_claim: boolean
  decision?: string
  reason?: string
  grant_balance?: number
  created_at?: string | null
  updated_at?: string | null
}

export interface SignupGrantRiskOverrideRequest {
  subject_type: 'email' | 'email_domain' | 'ip' | 'oauth_identity' | 'device'
  subject: string
  action: 'allow' | 'block'
  reason?: string
}

export interface SignupGrantRiskOverrideRecord {
  id: number
  subject_type: string
  subject_value: string
  subject_hash: string
  action: string
  reason: string
  created_by?: number | null
  expires_at?: string | null
  created_at: string
  updated_at: string
}

export interface SignupGrantRiskOverridesResponse {
  items: SignupGrantRiskOverrideRecord[]
  total: number
  page: number
  size: number
}

export interface SignupGrantRiskOverrideFilters {
  subject_type?: 'email' | 'email_domain' | 'ip' | 'oauth_identity' | 'device' | ''
  action?: 'allow' | 'block' | ''
  subject?: string
}

export interface SignupGrantAdminAuditLog {
  id: number
  operation: string
  target_user_id?: number | null
  target_user_public_id?: string
  subject_type: string
  subject_value: string
  subject_hash: string
  action: string
  amount: number
  reason: string
  admin_id?: number | null
  metadata?: Record<string, unknown>
  created_at: string
}

export interface SignupGrantAdminAuditLogsResponse {
  items: SignupGrantAdminAuditLog[]
  total: number
  page: number
  size: number
}

export interface UserProfileSummary {
  user: UserProfileSummaryUser
  classification: UserProfileClassification
  registration: UserProfileRegistrationSummary
  auth_identities: UserProfileAuthIdentitySummary[]
  activity: UserProfileActivitySummary
  api_keys: UserProfileAPIKeySummary
  payments: UserProfilePaymentSummary
  balance: UserProfileBalanceSummary
  business: UserProfileBusinessSummary
  timeline: UserProfileTimelineEvent[]
  risk_tags: UserProfileRiskTag[]
}

export interface UserProfileSummaryUser {
  id: number
  email: string
  username: string
  notes?: string
  role: string
  status: string
  signup_source: string
  balance: number
  paid_balance: number
  gift_balance: number
  total_recharged: number
  concurrency: number
  created_at: string
  updated_at: string
  last_login_at?: string | null
  last_active_at?: string | null
  last_used_at?: string | null
  deleted_at?: string | null
}

export interface UserProfileClassification {
  category: string
  label: string
  confidence: string
  reasons: string[]
}

export interface UserProfileRegistrationSummary {
  registered_via: string
  registration_ip?: string
  user_agent?: string
  accept_language?: string
  device_fingerprint?: string
  header_snapshot?: Record<string, string>
  nearby_auth_event?: string
  nearby_auth_status?: string
  nearby_auth_at?: string | null
  same_ip_signup_count_24h: number
  same_domain_signup_count: number
  email_domain: string
  disposable_email: boolean
}

export interface UserProfileAuthIdentitySummary {
  provider_type: string
  provider_key: string
  provider_subject: string
  verified_at?: string | null
  created_at: string
}

export interface UserProfileActivitySummary {
  api_usage_count: number
  api_actual_cost: number
  first_api_usage_at?: string | null
  last_api_usage_at?: string | null
  last_http_at?: string | null
}

export interface UserProfileAPIKeySummary {
  total_count: number
  active_count: number
  first_created_at?: string | null
  last_created_at?: string | null
}

export interface UserProfilePaymentSummary {
  order_count: number
  paid_order_count: number
  paid_amount: number
  refund_amount: number
  last_order_at?: string | null
}

export interface UserProfileBalanceSummary {
  ledger_count: number
  positive_ledger_amount: number
  net_ledger_amount: number
  redeem_count: number
  redeem_balance_amount: number
}

export interface UserProfileBusinessSummary {
  image_task_count: number
  image_success_count: number
  image_actual_cost: number
  first_image_task_at?: string | null
  last_image_task_at?: string | null
  wechat_task_count: number
  wechat_actual_cost: number
  first_wechat_task_at?: string | null
  last_wechat_task_at?: string | null
}

export interface UserProfileRiskTag {
  key: string
  label: string
  severity: 'info' | 'warning' | 'danger' | string
  detail: string
}

export interface UserProfileTimelineEvent {
  occurred_at: string
  source: string
  action: string
  title: string
  detail?: string
  status?: string
  amount?: number | null
  ip_address?: string
  user_agent?: string
  record_id?: string
}

export interface UserProfileInsights {
  generated_at: string
  classification: UserInsightCount[]
  signup_sources: UserInsightCount[]
  registration_ips: UserInsightDimension[]
  user_agents: UserInsightDimension[]
  languages: UserInsightDimension[]
  funnel: UserInsightFunnelStep[]
  risk_samples: UserInsightRiskSample[]
}

export interface UserInsightCount {
  key: string
  label: string
  count: number
}

export interface UserInsightDimension {
  value: string
  count: number
  last_seen?: string | null
}

export interface UserInsightFunnelStep {
  key: string
  label: string
  count: number
  conversion: number
}

export interface UserInsightRiskSample {
  user_id: number
  email: string
  username: string
  label: string
  reason: string
  severity: string
  registration_ip?: string
  created_at: string
  last_active_at?: string | null
}

/**
 * List all users with pagination
 * @param page - Page number (default: 1)
 * @param pageSize - Items per page (default: 20)
 * @param filters - Optional filters (status, role, search, attributes)
 * @param options - Optional request options (signal)
 * @returns Paginated list of users
 */
export async function list(
  page: number = 1,
  pageSize: number = 20,
  filters?: {
    status?: 'active' | 'disabled'
    role?: 'admin' | 'user'
    search?: string
    group_name?: string         // fuzzy filter by allowed group name
    api_key_group_id?: number   // filter users by the group their API keys are bound to
    attributes?: Record<number, string>  // attributeId -> value
    include_subscriptions?: boolean
    sort_by?: string
    sort_order?: 'asc' | 'desc'
  },
  options?: {
    signal?: AbortSignal
  }
): Promise<PaginatedResponse<AdminUser>> {
  // Build params with attribute filters in attr[id]=value format
  const params: Record<string, any> = {
    page,
    page_size: pageSize,
    status: filters?.status,
    role: filters?.role,
    search: filters?.search,
    group_name: filters?.group_name,
    api_key_group_id: filters?.api_key_group_id,
    include_subscriptions: filters?.include_subscriptions,
    sort_by: filters?.sort_by,
    sort_order: filters?.sort_order
  }

  // Add attribute filters as attr[id]=value
  if (filters?.attributes) {
    for (const [attrId, value] of Object.entries(filters.attributes)) {
      if (value) {
        params[`attr[${attrId}]`] = value
      }
    }
  }
  const { data } = await apiClient.get<PaginatedResponse<AdminUser>>('/admin/users', {
    params,
    signal: options?.signal
  })
  return data
}

/**
 * Get user by ID
 * @param id - User ID
 * @param includeDeleted - Whether to include soft-deleted users
 * @returns User details
 */
export async function getById(id: AdminUserResourceID, includeDeleted = false): Promise<AdminUser> {
  const { data } = await apiClient.get<AdminUser>(userResourcePath(id), {
    params: includeDeleted ? { include_deleted: true } : undefined
  })
  return data
}

export async function getUserProfileSummary(id: AdminUserResourceID): Promise<UserProfileSummary> {
  const { data } = await apiClient.get<UserProfileSummary>(`${userResourcePath(id)}/profile-summary`)
  return data
}

export async function getUserProfileInsights(limit = 10): Promise<UserProfileInsights> {
  const { data } = await apiClient.get<UserProfileInsights>('/admin/users/profile-insights', {
    params: { limit }
  })
  return data
}

/**
 * Create new user
 * @param userData - User data (email, password, etc.)
 * @returns Created user
 */
export async function create(userData: {
  email: string
  password: string
  username?: string
  notes?: string
  balance?: number
  concurrency?: number
  rpm_limit?: number
  allowed_groups?: number[] | null
}): Promise<AdminUser> {
  const { data } = await apiClient.post<AdminUser>('/admin/users', userData)
  return data
}

/**
 * Update user
 * @param id - User ID
 * @param updates - Fields to update
 * @returns Updated user
 */
export async function update(id: AdminUserResourceID, updates: UpdateUserRequest): Promise<AdminUser> {
  const { data } = await apiClient.put<AdminUser>(userResourcePath(id), updates)
  return data
}

/**
 * Delete user
 * @param id - User ID
 * @returns Success confirmation
 */
export async function deleteUser(id: AdminUserResourceID): Promise<{ message: string }> {
  const { data } = await apiClient.delete<{ message: string }>(userResourcePath(id))
  return data
}

/**
 * Update user balance
 * @param id - User ID
 * @param balance - New balance
 * @param operation - Operation type ('set', 'add', 'subtract')
 * @param notes - Optional notes for the balance adjustment
 * @returns Updated user
 */
export async function updateBalance(
  id: AdminUserResourceID,
  balance: number,
  operation: 'set' | 'add' | 'subtract' = 'set',
  notes?: string
): Promise<AdminUser> {
  const { data } = await apiClient.post<AdminUser>(`${userResourcePath(id)}/balance`, {
    balance,
    operation,
    notes: notes || ''
  })
  return data
}

export async function listSignupGrantRiskClaims(
  page: number = 1,
  pageSize: number = 20,
  filters?: SignupGrantRiskClaimFilters
): Promise<SignupGrantRiskClaimsResponse> {
  const { data } = await apiClient.get<SignupGrantRiskClaimsResponse>(
    '/admin/users/signup-grant-risk/claims',
    { params: { page, page_size: pageSize, ...filters } }
  )
  return data
}

export async function getSignupGrantRiskSummary(id: AdminUserResourceID): Promise<SignupGrantRiskUserSummary> {
  const { data } = await apiClient.get<SignupGrantRiskUserSummary>(
    `${userResourcePath(id)}/signup-grant/summary`
  )
  return data
}

export async function manualGrantSignupGiftBalance(
  id: AdminUserResourceID,
  amount: number,
  reason?: string
): Promise<AdminUser> {
  const { data } = await apiClient.post<AdminUser>(
    `${userResourcePath(id)}/signup-grant/manual-grant`,
    { amount, reason: reason || '' }
  )
  return data
}

export async function upsertSignupGrantRiskOverride(
  input: SignupGrantRiskOverrideRequest
): Promise<{ message: string }> {
  const { data } = await apiClient.post<{ message: string }>(
    '/admin/users/signup-grant-risk/overrides',
    input
  )
  return data
}

export async function deleteSignupGrantRiskOverride(id: number): Promise<{ message: string }> {
  const { data } = await apiClient.delete<{ message: string }>(
    `/admin/users/signup-grant-risk/overrides/${id}`
  )
  return data
}

export async function listSignupGrantRiskOverrides(
  page: number = 1,
  pageSize: number = 20,
  filters?: SignupGrantRiskOverrideFilters
): Promise<SignupGrantRiskOverridesResponse> {
  const { data } = await apiClient.get<SignupGrantRiskOverridesResponse>(
    '/admin/users/signup-grant-risk/overrides',
    { params: { page, page_size: pageSize, ...filters } }
  )
  return data
}

export async function listSignupGrantAdminAuditLogs(
  page: number = 1,
  pageSize: number = 20,
  filters?: {
    operation?: string
    admin_id?: number | string
    target_user_id?: number | string
  }
): Promise<SignupGrantAdminAuditLogsResponse> {
  const { data } = await apiClient.get<SignupGrantAdminAuditLogsResponse>(
    '/admin/users/signup-grant-risk/audit-logs',
    { params: { page, page_size: pageSize, ...filters } }
  )
  return data
}

/**
 * Update user concurrency
 * @param id - User ID
 * @param concurrency - New concurrency limit
 * @returns Updated user
 */
export async function updateConcurrency(id: AdminUserResourceID, concurrency: number): Promise<AdminUser> {
  return update(id, { concurrency })
}

/**
 * Toggle user status
 * @param id - User ID
 * @param status - New status
 * @returns Updated user
 */
export async function toggleStatus(id: AdminUserResourceID, status: 'active' | 'disabled'): Promise<AdminUser> {
  return update(id, { status })
}

/**
 * Get user's API keys
 * @param id - User ID
 * @returns List of user's API keys
 */
export async function getUserApiKeys(id: AdminUserResourceID): Promise<PaginatedResponse<ApiKey>> {
  const { data } = await apiClient.get<PaginatedResponse<ApiKey>>(`${userResourcePath(id)}/api-keys`)
  return data
}

/**
 * Get user's usage statistics
 * @param id - User ID
 * @param period - Time period
 * @returns User usage statistics
 */
export async function getUserUsageStats(
  id: AdminUserResourceID,
  period: string = 'month'
): Promise<{
  total_requests: number
  total_cost: number
  total_tokens: number
}> {
  const { data } = await apiClient.get<{
    total_requests: number
    total_cost: number
    total_tokens: number
  }>(`${userResourcePath(id)}/usage`, {
    params: { period }
  })
  return data
}

/**
 * Balance history item returned from the API
 */
export interface BalanceHistoryItem {
  id: number
  code: string
  type: string
  value: number
  status: string
  used_by: number | null
  used_at: string | null
  created_at: string
  group_id: number | null
  validity_days: number
  notes: string
  user?: { id: number; email: string } | null
  group?: { id: number; name: string } | null
}

// Balance history response extends pagination with total_recharged summary
export interface BalanceHistoryResponse extends PaginatedResponse<BalanceHistoryItem> {
  total_recharged: number
}

/**
 * Get user's balance/concurrency change history
 * @param id - User ID
 * @param page - Page number
 * @param pageSize - Items per page
 * @param type - Optional type filter (balance, affiliate_balance, admin_balance, concurrency, admin_concurrency, subscription)
 * @returns Paginated balance history with total_recharged
 */
export async function getUserBalanceHistory(
  id: AdminUserResourceID,
  page: number = 1,
  pageSize: number = 20,
  type?: string
): Promise<BalanceHistoryResponse> {
  const params: Record<string, any> = { page, page_size: pageSize }
  if (type) params.type = type
  const { data } = await apiClient.get<BalanceHistoryResponse>(
    `${userResourcePath(id)}/balance-history`,
    { params }
  )
  return data
}

/**
 * Replace user's exclusive group
 * @param userId - User ID
 * @param oldGroupId - Current group ID to replace
 * @param newGroupId - New group ID to replace with
 * @returns Number of migrated keys
 */
export async function replaceGroup(
  userId: AdminUserResourceID,
  oldGroupId: number,
  newGroupId: number
): Promise<{ migrated_keys: number }> {
  const { data } = await apiClient.post<{ migrated_keys: number }>(
    `${userResourcePath(userId)}/replace-group`,
    { old_group_id: oldGroupId, new_group_id: newGroupId }
  )
  return data
}

export async function bindUserAuthIdentity(
  userId: AdminUserResourceID,
  input: AdminBindAuthIdentityRequest
): Promise<AdminBoundAuthIdentity> {
  const { data } = await apiClient.post<AdminBoundAuthIdentity>(
    `${userResourcePath(userId)}/auth-identities`,
    input
  )
  return data
}

/**
 * Platform quota types
 */
export type PlatformQuotaPlatform = 'anthropic' | 'openai' | 'gemini' | 'antigravity' | 'grok'
export type PlatformQuotaWindow = 'daily' | 'weekly' | 'monthly'

export interface PlatformQuotaItem {
  platform: PlatformQuotaPlatform
  daily_limit_usd: number | null
  weekly_limit_usd: number | null
  monthly_limit_usd: number | null
  daily_usage_usd: number
  weekly_usage_usd: number
  monthly_usage_usd: number
  daily_window_start?: string | null
  weekly_window_start?: string | null
  monthly_window_start?: string | null
  daily_window_resets_at?: string | null
  weekly_window_resets_at?: string | null
  monthly_window_resets_at?: string | null
}

export interface PlatformQuotaUpdateItem {
  platform: PlatformQuotaPlatform
  daily_limit_usd: number | null
  weekly_limit_usd: number | null
  monthly_limit_usd: number | null
}

export interface PlatformQuotasResponse {
  platform_quotas: PlatformQuotaItem[]
}

/**
 * Get user's platform quotas
 */
export async function getPlatformQuotas(id: AdminUserResourceID): Promise<PlatformQuotasResponse> {
  const { data } = await apiClient.get<PlatformQuotasResponse>(
    `${userResourcePath(id)}/platform-quotas`
  )
  return data
}

/**
 * Replace user's platform quotas (全量替换)
 */
export async function updatePlatformQuotas(
  id: AdminUserResourceID,
  quotas: PlatformQuotaUpdateItem[]
): Promise<PlatformQuotasResponse> {
  const { data } = await apiClient.put<PlatformQuotasResponse>(
    `${userResourcePath(id)}/platform-quotas`,
    { quotas }
  )
  return data
}

/**
 * Reset a single (platform, window) usage immediately
 */
export async function resetPlatformQuotaWindow(
  id: AdminUserResourceID,
  platform: PlatformQuotaPlatform,
  window: PlatformQuotaWindow
): Promise<PlatformQuotasResponse> {
  const { data } = await apiClient.post<PlatformQuotasResponse>(
    `${userResourcePath(id)}/platform-quotas/reset`,
    { platform, window }
  )
  return data
}

export const usersAPI = {
  list,
  getById,
  getUserProfileSummary,
  getUserProfileInsights,
  create,
  update,
  delete: deleteUser,
  updateBalance,
  listSignupGrantRiskClaims,
  getSignupGrantRiskSummary,
  manualGrantSignupGiftBalance,
  upsertSignupGrantRiskOverride,
  deleteSignupGrantRiskOverride,
  listSignupGrantRiskOverrides,
  listSignupGrantAdminAuditLogs,
  updateConcurrency,
  toggleStatus,
  getUserApiKeys,
  getUserUsageStats,
  getUserBalanceHistory,
  replaceGroup,
  bindUserAuthIdentity,
  getPlatformQuotas,
  updatePlatformQuotas,
  resetPlatformQuotaWindow,
}

export default usersAPI
