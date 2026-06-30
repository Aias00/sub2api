import { apiClient } from './client'

export type HomeBusinessCapabilityRuntimeStatus = 'available' | 'in_progress' | 'disabled' | 'hidden'

export interface HomeBusinessCapabilityStatus {
  status: HomeBusinessCapabilityRuntimeStatus
  statusLabel?: string
  status_label?: string
  message?: string
  count?: number
}

export type HomeBusinessCapabilityStatusMap = Record<string, HomeBusinessCapabilityStatus>

export const HOME_BUSINESS_CAPABILITY_STATUS_UNAVAILABLE: HomeBusinessCapabilityStatusMap = {
  'wechat-export': {
    status: 'in_progress',
    message: 'Capability status is currently unavailable.',
  },
  'image-workspace': {
    status: 'in_progress',
    message: 'Capability status is currently unavailable.',
  },
  'hot-topics': {
    status: 'in_progress',
    message: 'Capability status is currently unavailable.',
  },
}

export async function probeHomeBusinessCapabilities(): Promise<HomeBusinessCapabilityStatusMap> {
  const { data } = await apiClient.get<HomeBusinessCapabilityStatusMap>('/home/business-capabilities')
  return Object.fromEntries(
    Object.entries(data || {}).map(([key, value]) => [
      key,
      value
        ? {
            ...value,
            statusLabel: value.statusLabel || value.status_label,
          }
        : value,
    ]),
  )
}
