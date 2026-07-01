import type { PromptCatalogFacet } from '@/api/prompts'
import type { PromptCatalogCopy, PromptCatalogDefaults } from '@/utils/promptCatalogShell'

export type PromptCatalogFilters = {
  search: string
  category: string
  hasImage: boolean
}

export function applyPromptCatalogDefaults(
  filters: PromptCatalogFilters,
  defaults: PromptCatalogDefaults | undefined,
) {
  filters.hasImage = Boolean(defaults?.hasImage)
}

export function resolvePromptCatalogPageTitle(copy: PromptCatalogCopy, _sourceType: string): string {
  return copy.caseTitle || copy.title
}

export function resolvePromptCatalogPageDescription(copy: PromptCatalogCopy, _sourceType: string): string {
  return copy.caseDescription || copy.description
}

export function resolvePromptCatalogPageSize(defaults: PromptCatalogDefaults | undefined): number | undefined {
  return defaults?.pageSize
}

export function resolvePromptCatalogSortBy(defaults: PromptCatalogDefaults | undefined): string | undefined {
  return defaults?.sortBy
}

export function resolvePromptCatalogSortOrder(defaults: PromptCatalogDefaults | undefined): string | undefined {
  return defaults?.sortOrder
}

export function resolvePromptCatalogGeneratorPath(defaults: PromptCatalogDefaults | undefined): string | undefined {
  return defaults?.generatorPath || '/image-generator'
}

export function resolvePromptCatalogGeneratorDraftSource(
  defaults: PromptCatalogDefaults | undefined,
): string | undefined {
  return defaults?.generatorDraftSource
}

export function resolvePromptCatalogImportXAuto(defaults: PromptCatalogDefaults | undefined): boolean {
  return defaults?.importXAuto ?? true
}

export function buildPromptCatalogListParams(filters: PromptCatalogFilters, defaults: PromptCatalogDefaults | undefined, page: number) {
  return {
    page,
    page_size: resolvePromptCatalogPageSize(defaults),
    category: filters.category || undefined,
    search: filters.search || undefined,
    has_image: filters.hasImage ? true : undefined,
    sort_by: resolvePromptCatalogSortBy(defaults),
    sort_order: resolvePromptCatalogSortOrder(defaults),
  }
}

export function formatPromptCatalogDate(value: string | null | undefined, locale: string): string {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  return new Intl.DateTimeFormat(locale, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  }).format(date)
}

const zhPromptCatalogFacetLabels: Record<string, string> = {
  gpt4o_image_prompts: 'GPT-4o 提示库',
  awesome_gpt_image_2: 'Awesome GPT Image 2',
  awesome_gpt_image_2_api_and_prompts: 'GPT Image 2 API 案例库',
  awesome_nano_banana_pro_prompts: 'Nano Banana Pro 提示库',
  GPT_Image2_Skill: 'GPT Image2 Skill',
  'nano-banana-pro': 'Nano Banana Pro',
  'openai-image': 'OpenAI Image',
  'gpt-image-2': 'GPT Image 2',
  'gpt-4o-image': 'GPT-4o Image',
  grok: 'Grok',
  '3d': '3D',
  animal: '动物',
  'Anime & Manga': '动漫与漫画',
  architecture: '建筑',
  'Architecture & Spaces': '建筑与空间',
  'Architecture & Interior': '建筑与室内',
  branding: '品牌',
  'Brand Systems & Identity': '品牌系统与视觉识别',
  'Brand & Logos': '品牌与标志',
  'Beauty & Lifestyle': '美妆与生活方式',
  'Character Design': '角色设计',
  cartoon: '卡通',
  character: '角色',
  Character: '角色',
  'Character Design Cases': '角色设计案例',
  Characters: '人物',
  'Characters & People': '人物与角色',
  Charts: '图表',
  'Charts & Infographics': '图表与信息可视化',
  'Cinematic & Animation': '影视与动画',
  'Cinematic Film References': '电影风格参考',
  'Comic / Storyboard': '漫画与分镜',
  'Comparison & Community Examples': '对比与社区案例',
  clay: '黏土',
  Classical: '古风',
  'Data Visualization': '数据可视化',
  Documents: '文档',
  'Documents & Publishing': '文档与出版物',
  Commerce: '商业',
  creative: '创意',
  'data-viz': '数据可视化',
  'Ad Creative Cases': '广告创意案例',
  'E-commerce Cases': '电商案例',
  'E-commerce Main Image': '电商主图',
  'Edit Endpoint Showcase': '编辑端点示例',
  Education: '教育',
  emoji: '表情',
  Events: '活动',
  'Events & Experience': '活动与体验',
  fantasy: '奇幻',
  fashion: '时尚',
  'Fashion Editorial': '时尚大片',
  felt: '毛毡',
  'Fine Art Painting': '艺术绘画',
  food: '美食',
  futuristic: '未来感',
  Game: '游戏',
  'Game Asset': '游戏资产',
  Gaming: '游戏',
  gaming: '游戏',
  History: '历史',
  'History & Classical Themes': '历史与古风题材',
  'Infographics & Field Guides': '信息图与图鉴',
  'Ink & Chinese': '水墨与东方',
  illustration: '插画',
  Illustration: '插画',
  'Illustration & Art': '插画与艺术',
  infographic: '信息图',
  Infographic: '信息图',
  'Infographic / Edu Visual': '信息图与教育视觉',
  interior: '室内',
  Isometric: '等距视角',
  landscape: '风景',
  logo: 'Logo',
  'More Illustration Styles': '更多插画风格',
  minimalist: '极简',
  nature: '自然',
  neon: '霓虹',
  'Official OpenAI Cookbook Examples': 'OpenAI Cookbook 官方示例',
  'Other Use Cases': '其他应用场景',
  'paper-craft': '纸艺',
  photography: '摄影',
  Photography: '摄影',
  'Photography & Realism': '摄影与写实',
  pixel: '像素',
  'Pixel Art': '像素艺术',
  portrait: '肖像',
  'Portrait & Photography Cases': '肖像与摄影案例',
  poster: '海报',
  Poster: '海报',
  'Poster & Illustration Cases': '海报与插画案例',
  'Posters & Typography': '海报与排版',
  product: '产品',
  Product: '产品',
  Products: '产品',
  'Products & E-commerce': '商品与电商',
  'Product & Food': '产品与美食',
  'Product Marketing': '产品营销',
  'Profile / Avatar': '头像与个人形象',
  'Research Paper Figures': '论文图表',
  retro: '复古',
  'Retro & Cyberpunk': '复古与赛博朋克',
  sculpture: '雕塑',
  'Scientific & Educational': '科学与教育',
  'Screen Photography': '屏幕摄影',
  Scenes: '场景',
  'Scenes & Storytelling': '场景与叙事',
  Social: '社交',
  'Social Media Post': '社媒帖子',
  Story: '叙事',
  'Tattoo Design': '纹身设计',
  Tech: '科技',
  'Technical Illustration': '技术插画',
  toy: '玩具',
  Travel: '旅行',
  typography: '字体',
  'Typography & Posters': '字体与海报',
  UI: '界面',
  'UI & Interfaces': '界面与交互',
  'UI & Social Media Mockup Cases': '界面与社媒样机案例',
  'UI/UX Mockups': 'UI/UX 样机',
  'Twitter Imports': 'Twitter 导入',
  Watercolor: '水彩',
  vehicle: '交通工具',
  'YouTube Thumbnail': 'YouTube 缩略图',
  general: '未分类',
}

export function resolvePromptCatalogValueLabel(value: string | null | undefined, locale: 'zh' | 'en' = 'en'): string {
  const normalized = String(value || '').trim()
  if (!normalized) return ''
  if (locale === 'zh') {
    const zhLabel = zhPromptCatalogFacetLabels[normalized]
    if (zhLabel) return zhLabel
  }
  return normalized
}

export function resolvePromptCatalogFacetLabel(facet: PromptCatalogFacet, locale: 'zh' | 'en' = 'en'): string {
  const valueLabel = resolvePromptCatalogValueLabel(facet.value, locale)
  if (valueLabel && valueLabel !== facet.value) {
    return valueLabel
  }
  const displayLabel = String(facet.display_label || facet.label || '').trim()
  if (!displayLabel || displayLabel === facet.value) {
    return valueLabel
  }
  if (locale === 'en' && containsCjk(displayLabel)) {
    return valueLabel
  }
  return displayLabel
}

function containsCjk(value: string): boolean {
  return /[\u3400-\u9fff]/.test(value)
}

export function buildPromptCatalogImportSuccessMessage(label: string, title: string): string {
  return `${label}: ${title}`
}
