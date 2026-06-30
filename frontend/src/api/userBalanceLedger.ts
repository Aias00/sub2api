import { apiClient } from './client'

// 流水类型枚举
export type BalanceLedgerEntryType =
  | 'recharge'
  | 'api_usage'
  | 'image_workspace'
  | 'wechat_export'
  | 'redeem'
  | 'admin_adjustment'
  | 'affiliate_transfer'
  | 'refund'
  | 'promo_bonus'
  | 'oauth_bind_bonus'
  | 'expiry'
  | 'correction'

// 来源类型枚举
export type BalanceLedgerSourceType =
  | 'payment_order'
  | 'usage_log'
  | 'redeem_code'
  | 'admin_action'
  | 'affiliate_ledger'
  | 'refund'
  | 'promo_code_usage'
  | 'oauth_binding'
  | 'image_workspace_record'
  | 'wechat_export_task'
  | 'system_correction'

// 流水记录
export interface BalanceLedgerEntry {
  id: number
  entry_type: BalanceLedgerEntryType
  amount: number
  balance_before: number | null
  balance_after: number | null
  source_type: BalanceLedgerSourceType
  source_id: number | null
  description: string
  metadata: Record<string, unknown> | null
  created_at: string
}

// 查询过滤器
export interface BalanceLedgerFilter {
  page?: number
  page_size?: number
  entry_types?: BalanceLedgerEntryType[]
  start_at?: string
  end_at?: string
}

// 流水列表响应
export interface BalanceLedgerListResponse {
  entries: BalanceLedgerEntry[]
  total: number
  page: number
  page_size: number
}

// 流水类型显示名称（中文）
export const entryTypeDisplayName: Record<BalanceLedgerEntryType, string> = {
  recharge: '充值',
  api_usage: 'API调用扣费',
  image_workspace: '图片生成扣费',
  wechat_export: '微信导出扣费',
  redeem: '兑换码兑换',
  admin_adjustment: '管理员调整',
  affiliate_transfer: '返利转入',
  refund: '退款扣减',
  promo_bonus: '优惠码奖励',
  oauth_bind_bonus: '绑定奖励',
  expiry: '过期清零',
  correction: '系统纠正',
}

// 流水类型图标名称
export const entryTypeIcon: Record<BalanceLedgerEntryType, string> = {
  recharge: 'dollar-up',
  api_usage: 'zap',
  image_workspace: 'image',
  wechat_export: 'wechat',
  redeem: 'gift',
  admin_adjustment: 'settings',
  affiliate_transfer: 'users',
  refund: 'dollar-down',
  promo_bonus: 'tag',
  oauth_bind_bonus: 'link',
  expiry: 'clock',
  correction: 'wrench',
}

// 流水类型颜色（CSS class）
export const entryTypeColor: Record<BalanceLedgerEntryType, string> = {
  recharge: 'green',
  api_usage: 'red',
  image_workspace: 'red',
  wechat_export: 'red',
  redeem: 'blue',
  admin_adjustment: 'yellow',
  affiliate_transfer: 'purple',
  refund: 'red',
  promo_bonus: 'green',
  oauth_bind_bonus: 'green',
  expiry: 'gray',
  correction: 'orange',
}

// 获取用户余额流水
export async function getUserBalanceLedger(
  filter: BalanceLedgerFilter = {},
): Promise<BalanceLedgerListResponse> {
  const params: Record<string, string | number> = {}

  if (filter.page) {
    params.page = filter.page
  }
  if (filter.page_size) {
    params.page_size = filter.page_size
  }
  if (filter.entry_types && filter.entry_types.length > 0) {
    params.entry_types = filter.entry_types.join(',')
  }
  if (filter.start_at) {
    params.start_at = filter.start_at
  }
  if (filter.end_at) {
    params.end_at = filter.end_at
  }

  const response = await apiClient.get<BalanceLedgerListResponse>('/user/balance-ledger', { params })
  return response.data
}

// 格式化金额（正数带+，负数保持负号）
export function formatLedgerAmount(amount: number): string {
  const prefix = amount >= 0 ? '+' : ''
  return `${prefix}${amount.toFixed(2)}`
}

// 格式化余额（null 显示为"历史数据"）
export function formatBalanceSnapshot(value: number | null): string {
  if (value === null) {
    return '历史数据'
  }
  return value.toFixed(2)
}