import { apiClient } from './client'

export type WeChatExportFormat = 'html' | 'markdown' | 'json'

export interface WeChatArticle {
  id: number
  title: string
  link: string
  source_type: string
  content_status: string
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
  error_message: string
  created_at: string
  updated_at: string
}

export interface WeChatExportArtifact {
  id: number
  task_id: number
  format: WeChatExportFormat
  file_name: string
  file_size: number
  download_url: string
}

export interface Paginated<T> {
  items: T[]
  total: number
  page: number
  page_size: number
  pages: number
}

export async function importWeChatArticleLink(link: string) {
  const { data } = await apiClient.post<WeChatArticle>('/wechat/articles/import-link', { link })
  return data
}

export async function listWeChatArticles() {
  const { data } = await apiClient.get<Paginated<WeChatArticle>>('/wechat/articles', {
    params: { page: 1, page_size: 50 },
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

export async function listWeChatExportTasks() {
  const { data } = await apiClient.get<Paginated<WeChatExportTask>>('/wechat/tasks', {
    params: { page: 1, page_size: 20 },
  })
  return data
}

export async function listWeChatExportArtifacts(taskId: number) {
  const { data } = await apiClient.get<{ items: WeChatExportArtifact[] }>(`/wechat/tasks/${taskId}/artifacts`)
  return data.items
}
