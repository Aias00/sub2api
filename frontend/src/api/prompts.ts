import { apiClient } from './client'

export interface PromptCatalogFacet {
  value: string
  label?: string
  count: number
  display_label: string
}

export interface PromptCatalogSummary {
  total: number
  case_count: number
  template_count: number
  source_count: number
  category_count: number
  sources: PromptCatalogFacet[]
  categories: PromptCatalogFacet[]
  template_groups: PromptCatalogFacet[]
}

export interface PromptCatalogItem {
  id: string
  title: string
  prompt: string
  prompt_preview: string
  category: string
  tags: string[]
  display_tags: string[]
  model_tags: string[]
  all_tags: string[]
  visible_tags: string[]
  source_url?: string
  image_url?: string
  primary_image_url: string
  image_urls: string[]
  image_original_url?: string
  image_preview_url?: string
  image_thumb_url?: string
  source_project: string
  source_type: string
  source_label?: string
  source_display_label: string
  github_url?: string
  prompt_char_count: number
  featured: boolean
  styles: string[]
  scenes: string[]
  import_source: string
  status: string
  imported_at?: string | null
  created_at: string
  updated_at: string
}

export interface PromptCatalogListParams {
  page?: number
  page_size?: number
  source_type?: 'case' | 'template' | string
  source_project?: string
  category?: string
  search?: string
  featured?: boolean
  has_image?: boolean
  sort_by?: 'imported_at' | 'title' | 'created_at' | 'updated_at' | string
  sort_order?: 'asc' | 'desc' | string
}

export interface PromptCatalogListResponse {
  items: PromptCatalogItem[]
  total: number
  page: number
  page_size: number
  pages: number
  summary: PromptCatalogSummary
}

export interface PromptTwitterImportRequest {
  url: string
  prompt?: string
  title?: string
  category?: string
  image_urls?: string[]
  x_auto?: boolean
}

export interface PromptTwitterImportResponse {
  item: PromptCatalogItem
  image_urls: string[]
  uploaded_urls: string[]
  warnings: string[]
}

export const promptsAPI = {
  listCases(params: PromptCatalogListParams = {}) {
    return apiClient.get<PromptCatalogListResponse>('/prompts/cases', { params })
  },

  getCase(id: string) {
    return apiClient.get<PromptCatalogItem>(`/prompts/cases/${encodeURIComponent(id)}`)
  },

  importTwitter(request: PromptTwitterImportRequest) {
    return apiClient.post<PromptTwitterImportResponse>('/admin/prompts/import-twitter', request)
  },
}
