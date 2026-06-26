#!/usr/bin/env node

import { mkdir, stat } from 'node:fs/promises'
import path from 'node:path'

import { runWechatExportFormats, WechatExportFormat, WechatExportRunnerArticle } from '../runner/export-runner'

interface ApiEnvelope<T> {
  code: number
  message: string
  data?: T
}

interface WechatTask {
  ID?: number
  id?: number
  Formats?: WechatExportFormat[]
  formats?: WechatExportFormat[]
}

interface WechatArticle {
  ID?: number
  id?: number
  AccountFakeID?: string
  account_fakeid?: string
  Title?: string
  title?: string
  Author?: string
  author?: string
  Link?: string
  link?: string
  MetadataJSON?: string
  metadata_json?: string
}

interface ClaimResponse {
  task: WechatTask | null
  articles: WechatArticle[]
}

const baseURL = (process.env.SUB2API_BASE_URL || 'http://127.0.0.1:8080/api/v1').replace(/\/+$/, '')
const workerToken = process.env.WECHAT_EXPORT_WORKER_TOKEN || ''
const outputRoot = process.env.WECHAT_EXPORT_OUTPUT_DIR || path.resolve(process.cwd(), 'runtime/wechat-export')

function apiHeaders() {
  const headers: Record<string, string> = {
    'content-type': 'application/json',
  }
  if (workerToken) {
    headers['x-wechat-worker-token'] = workerToken
  }
  return headers
}

async function apiPost<T>(pathName: string, body: unknown) {
  const response = await fetch(`${baseURL}${pathName}`, {
    method: 'POST',
    headers: apiHeaders(),
    body: JSON.stringify(body),
  })
  const contentType = response.headers.get('content-type') || ''
  if (!contentType.includes('application/json')) {
    const text = await response.text()
    throw new Error(`${pathName} failed: ${response.status} ${text.slice(0, 160)}`)
  }
  const envelope = (await response.json()) as ApiEnvelope<T>
  if (!response.ok || envelope.code !== 0) {
    throw new Error(`${pathName} failed: ${response.status} ${envelope.message || response.statusText}`)
  }
  return envelope.data as T
}

function taskIdOf(task: WechatTask) {
  return Number(task.id ?? task.ID ?? 0)
}

function formatsOf(task: WechatTask): WechatExportFormat[] {
  const formats = task.formats ?? task.Formats ?? []
  const normalized = formats.filter((format): format is WechatExportFormat =>
    format === 'html' || format === 'markdown' || format === 'json',
  )
  return normalized.length > 0 ? normalized : ['html', 'markdown', 'json']
}

function articleField(article: WechatArticle, lower: keyof WechatArticle, upper: keyof WechatArticle) {
  return String(article[lower] ?? article[upper] ?? '').trim()
}

function articleMetadata(article: WechatArticle) {
  const raw = articleField(article, 'metadata_json', 'MetadataJSON')
  if (!raw) return undefined
  try {
    return JSON.parse(raw)
  } catch {
    return undefined
  }
}

async function fetchArticleHtml(link: string, title: string) {
  if (!link) {
    return `<article><h1>${escapeHtml(title || 'Untitled')}</h1></article>`
  }
  try {
    const response = await fetch(link, {
      headers: {
        'user-agent':
          'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125 Safari/537.36',
      },
    })
    if (response.ok) {
      return await response.text()
    }
  } catch (error) {
    console.warn('[wechat-worker] article fetch failed, using fallback html', { link, error })
  }
  return `<article><h1>${escapeHtml(title || link)}</h1><p><a href="${escapeHtml(link)}">${escapeHtml(link)}</a></p></article>`
}

function escapeHtml(value: string) {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}

async function buildRunnerArticles(articles: WechatArticle[]): Promise<WechatExportRunnerArticle[]> {
  const output: WechatExportRunnerArticle[] = []
  for (const article of articles) {
    const link = articleField(article, 'link', 'Link')
    const title = articleField(article, 'title', 'Title') || link || `article-${article.id ?? article.ID ?? 'unknown'}`
    const accountName = articleField(article, 'account_fakeid', 'AccountFakeID')
    const html = await fetchArticleHtml(link, title)
    output.push({
      accountName,
      aid: String(article.id ?? article.ID ?? ''),
      link,
      title,
      html,
      metadata: articleMetadata(article),
    })
  }
  return output
}

async function runOnce() {
  const claim = await apiPost<ClaimResponse>('/wechat/worker/tasks/claim', { lease_seconds: 300 })
  if (!claim.task) {
    console.log('[wechat-worker] no queued tasks')
    return
  }
  const taskId = taskIdOf(claim.task)
  try {
    const taskOutputDir = path.join(outputRoot, String(taskId))
    await mkdir(taskOutputDir, { recursive: true })
    const generated = await runWechatExportFormats({
      outputDir: taskOutputDir,
      taskId: String(taskId),
      formats: formatsOf(claim.task),
      articles: await buildRunnerArticles(claim.articles || []),
    })
    const artifacts = await Promise.all(
      generated.map(async (item) => {
        const fileStat = await stat(item.filePath)
        return {
          format: item.format,
          storage_provider: 'local',
          storage_key: item.filePath,
          file_name: item.fileName,
          file_size: fileStat.size,
        }
      }),
    )
    await apiPost(`/wechat/worker/tasks/${taskId}/complete`, {
      artifacts,
      result_manifest_json: JSON.stringify({ artifacts }),
    })
    console.log('[wechat-worker] completed task', taskId)
  } catch (error) {
    await apiPost(`/wechat/worker/tasks/${taskId}/fail`, {
      message: error instanceof Error ? error.message : String(error),
    })
    throw error
  }
}

async function main() {
  const once = process.argv.includes('--once')
  const intervalMs = Number.parseInt(
    process.env.WECHAT_EXPORT_WORKER_INTERVAL_MS || '5000',
    10,
  )

  do {
    try {
      await runOnce()
    } catch (error) {
      console.error('[wechat-worker] task run failed', error)
    }

    if (once) {
      break
    }

    await new Promise((resolve) => setTimeout(resolve, intervalMs))
  } while (true)
}

main().catch((error) => {
  console.error('[wechat-worker] fatal error', error)
  process.exit(1)
})
