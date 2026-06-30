import { apiClient } from './client'

export interface ImageWorkspaceArtifact {
  id: number
  task_id: number
  user_id?: number
  image_url: string
  storage_provider: string
  storage_key: string
  prompt: string
  mime_type: string
  width: number
  height: number
  file_size: number
  checksum?: string
  metadata_json?: string
  created_at: string
}

export interface ImageWorkspaceTask {
  id: number
  status: string
  prompt: string
  negative_prompt: string
  model: string
  provider: string
  size: string
  quality: string
  style: string
  seed?: number
  batch_size: number
  template_id?: number
  worker_lease_until?: string
  cost_estimate: number
  balance_snapshot: number
  error_message: string
  result_json: string
  artifacts?: ImageWorkspaceArtifact[]
  created_at: string
  updated_at: string
}

export interface ImageWorkspaceTemplate {
  id: number
  title: string
  description: string
  prompt: string
  negative_prompt: string
  model: string
  size: string
  quality: string
  style: string
  is_default: boolean
  created_at: string
  updated_at: string
}

export interface ImageWorkspaceUsageRecord {
  id: number
  task_id: number
  user_id?: number
  provider: string
  model: string
  size: string
  quality: string
  image_count: number
  reserved_cost: number
  actual_cost: number
  balance_snapshot: number
  billing_status: string
  metadata_json?: string
  created_at: string
  updated_at: string
}

export interface ImageWorkspaceModelOption {
  id: string
  label: string
  provider: string
  default_size: string
  default_quality: string
  sizes: string[]
  qualities: string[]
  cost_per_image: number
  cost_hint?: string
  enabled: boolean
}

export interface Paginated<T> {
  items: T[]
  total: number
  page: number
  page_size: number
  pages: number
}

export async function createImageWorkspaceTask(payload: {
  prompt: string
  negative_prompt?: string
  model?: string
  provider?: string
  size?: string
  quality?: string
  style?: string
  seed?: number | null
  batch_size?: number
  template_id?: number | null
}) {
  const { data } = await apiClient.post<ImageWorkspaceTask>('/image-workspace/tasks', payload)
  return data
}

export async function listImageWorkspaceTasks(params: { page?: number; page_size?: number; status?: string } = {}) {
  const { data } = await apiClient.get<Paginated<ImageWorkspaceTask>>('/image-workspace/tasks', {
    params: {
      page: params.page ?? 1,
      page_size: params.page_size ?? 20,
      ...(params.status ? { status: params.status } : {}),
    },
  })
  return data
}

export async function listImageWorkspaceModels() {
  const { data } = await apiClient.get<{ models: ImageWorkspaceModelOption[] }>('/image-workspace/models')
  return data.models
}

export async function getImageWorkspaceTask(taskID: number) {
  const { data } = await apiClient.get<ImageWorkspaceTask>(`/image-workspace/tasks/${taskID}`)
  return data
}

export async function listImageWorkspaceTemplates() {
  const { data } = await apiClient.get<{ items: ImageWorkspaceTemplate[] }>('/image-workspace/templates')
  return data.items
}

export async function listImageWorkspaceUsageRecords(params: { page?: number; page_size?: number } = {}) {
  const { data } = await apiClient.get<Paginated<ImageWorkspaceUsageRecord>>('/image-workspace/usage-records', {
    params: {
      page: params.page ?? 1,
      page_size: params.page_size ?? 10,
    },
  })
  return data
}

export async function saveImageWorkspaceTemplate(payload: {
  id?: number
  title: string
  description?: string
  prompt: string
  negative_prompt?: string
  model?: string
  size?: string
  quality?: string
  style?: string
  is_default?: boolean
}) {
  const { data } = await apiClient.post<ImageWorkspaceTemplate>('/image-workspace/templates', payload)
  return data
}

export async function cancelImageWorkspaceTask(taskID: number) {
  const { data } = await apiClient.post<ImageWorkspaceTask>(`/image-workspace/tasks/${taskID}/cancel`)
  return data
}

export async function retryImageWorkspaceTask(taskID: number) {
  const { data } = await apiClient.post<ImageWorkspaceTask>(`/image-workspace/tasks/${taskID}/retry`)
  return data
}

export async function downloadImageWorkspaceArtifact(artifactId: number) {
  const { data } = await apiClient.get<Blob>(`/image-workspace/artifacts/${artifactId}/download`, {
    responseType: 'blob',
  })
  return data
}

export async function deleteImageWorkspaceTemplate(templateId: number) {
  const { data } = await apiClient.delete<{ deleted: boolean }>(`/image-workspace/templates/${templateId}`)
  return data
}
