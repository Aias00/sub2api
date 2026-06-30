import { apiClient } from './client'
import type { Paginated } from './image-workspace'

export interface HotSource {
  id: number
  source_id: string
  adapter_kind: string
  title: string
  description: string
  enabled: boolean
  base_url: string
  seed_urls_json: string
  config_json: string
  sort_order: number
  created_at: string
  updated_at: string
}

export interface HotItem {
  id: number
  source_id: string
  external_id: string
  canonical_url: string
  title: string
  summary: string
  body: string
  quoted: string
  reason: string
  published_at?: string
  author: string
  source_name: string
  source_handle: string
  badge: string
  score: string
  content_type: string
  tags_json: string
  metrics_json: string
  raw_ref_json: string
  content_hash: string
  has_media: boolean
  status: string
  created_at: string
  updated_at: string
}

export interface HotRunEvent {
  id: number
  legacy_id: number
  run_id: string
  node: string
  message: string
  payload_json: string
  created_at: string
}

export async function listHotSources() {
  const { data } = await apiClient.get<{ items: HotSource[] }>('/hot/sources')
  return data.items
}

export async function listHotItems(params: {
  page?: number
  page_size?: number
  source_id?: string
  q?: string
} = {}) {
  const { data } = await apiClient.get<Paginated<HotItem>>('/hot/items', {
    params: {
      page: params.page ?? 1,
      page_size: params.page_size ?? 20,
      source_id: params.source_id || undefined,
      q: params.q || undefined,
    },
  })
  return data
}

export async function listHotRunEvents(params: {
  run_id: string
  page?: number
  page_size?: number
}) {
  const { data } = await apiClient.get<Paginated<HotRunEvent>>('/hot/run-events', {
    params: {
      run_id: params.run_id,
      page: params.page ?? 1,
      page_size: params.page_size ?? 20,
    },
  })
  return data
}
