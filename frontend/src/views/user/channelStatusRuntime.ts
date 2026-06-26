import type { UserMonitorDetail, UserMonitorView } from '@/api/channelMonitor'
import { STATUS_OPERATIONAL } from '@/constants/channelMonitor'
import type { MonitorWindow, OverallStatus } from '@/components/user/monitor/MonitorHero.vue'

export function resolveChannelStatusOverallStatus(items: UserMonitorView[]): OverallStatus {
  if (items.length === 0) return 'operational'
  for (const item of items) {
    if (item.primary_status === 'failed' || item.primary_status === 'error') return 'degraded'
    if (item.primary_status !== STATUS_OPERATIONAL) return 'degraded'
  }
  return 'operational'
}

export function resolveChannelStatusDetailTitle(
  detailTarget: UserMonitorView | null,
  fallbackTitle: string,
) {
  return detailTarget?.name || fallbackTitle
}

export function shouldEnsureChannelStatusDetails(window: MonitorWindow) {
  return window !== '7d'
}

export function buildChannelStatusDetailCache(
  cache: Record<number, UserMonitorDetail>,
  id: number,
  detail: UserMonitorDetail,
) {
  return {
    ...cache,
    [id]: detail,
  }
}
