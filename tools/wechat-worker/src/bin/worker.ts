#!/usr/bin/env node

import { mkdir, stat } from 'node:fs/promises'
import { createHash } from 'node:crypto'
import { createReadStream } from 'node:fs'
import path from 'node:path'

import { extractWechatArticleSummaryFromHtml, WechatArticleSummary } from '../core/html'
import { assertWechatArticleHtml, wechatVerifyPageMessage } from '../core/verification'
import { runWechatExportFormats, WechatExportFormat, WechatExportRunnerArticle } from '../runner/export-runner'

interface ApiEnvelope<T> {
  code: number
  message: string
  data?: T
}

interface WechatTask {
  ID?: number
  id?: number
  UserID?: number
  user_id?: number
  Formats?: WechatExportFormat[]
  formats?: WechatExportFormat[]
  IncludeEngagement?: boolean
  include_engagement?: boolean
}

interface WechatArticle {
  ID?: number
  id?: number
  UserID?: number
  user_id?: number
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

interface BuiltArticle {
  runnerArticle: WechatExportRunnerArticle
  rawArticle: WechatArticle
  summary: WechatArticleSummary
}

interface TaskLogPayload {
  event: string
  status?: string
  message?: string
  meta?: Record<string, unknown>
}

interface BuiltArticlesResult {
  articles: BuiltArticle[]
  failedCount: number
  failures: Array<{ link: string; message: string }>
}

interface WechatEngagementResult {
  readNum?: number
  oldLikeNum?: number
  shareNum?: number
  likeNum?: number
  commentNum?: number
  engagementFetchStatus: 'skipped' | 'fetched' | 'unavailable' | 'failed'
  engagementFetchMessage?: string
}

interface BackendWechatEngagementResult {
  read_num?: number
  old_like_num?: number
  share_num?: number
  like_num?: number
  comment_num?: number
  status?: 'skipped' | 'fetched' | 'unavailable' | 'failed'
  message?: string
}

interface ClaimResponse {
  task: WechatTask | null
  articles: WechatArticle[]
  lease_token?: string
  leaseToken?: string
}

const baseURL = (process.env.CLOUDBASE_BASE_URL || 'http://127.0.0.1:8080/api/v1').replace(/\/+$/, '')
const workerToken = process.env.WECHAT_EXPORT_WORKER_TOKEN || ''
const outputRoot = process.env.WECHAT_EXPORT_OUTPUT_DIR || path.resolve(process.cwd(), 'runtime/wechat-export')
const storageKeyRoot = process.env.WECHAT_EXPORT_STORAGE_KEY_ROOT || outputRoot
let workerConcurrency = clampInteger(process.env.WECHAT_EXPORT_WORKER_CONCURRENCY || '1', 1, 8)
let workerLeaseSeconds = clampInteger(process.env.WECHAT_EXPORT_WORKER_LEASE_SECONDS || '300', 60, 3600)
let fetchRetries = clampInteger(process.env.WECHAT_EXPORT_FETCH_RETRIES || '2', 0, 5)
let fetchTimeoutMs = clampInteger(process.env.WECHAT_EXPORT_FETCH_TIMEOUT_MS || '20000', 1000, 120000)
let idleIntervalMs = clampInteger(process.env.WECHAT_EXPORT_WORKER_INTERVAL_MS || '5000', 1000, 300000)
let maxBackoffMs = clampInteger(process.env.WECHAT_EXPORT_WORKER_MAX_BACKOFF_MS || '60000', idleIntervalMs, 300000)
let runtimeConfigRefreshMs = clampInteger(process.env.WECHAT_EXPORT_RUNTIME_CONFIG_REFRESH_MS || '60000', 10000, 3600000)
let nextRuntimeConfigRefreshAt = 0

function clampInteger(value: string, min: number, max: number) {
  const parsed = Number.parseInt(value, 10)
  if (!Number.isFinite(parsed)) return min
  return Math.min(Math.max(parsed, min), max)
}

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

async function apiGet<T>(pathName: string) {
  const response = await fetch(`${baseURL}${pathName}`, {
    method: 'GET',
    headers: apiHeaders(),
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

async function loadRuntimeConfig() {
  try {
    const runtime = await apiGet<{
      fetch_retries?: number
      fetch_timeout_ms?: number
      worker_concurrency?: number
      worker_interval_ms?: number
      worker_lease_seconds?: number
      worker_max_backoff_ms?: number
    }>('/wechat/worker/runtime-config')
    fetchRetries = clampInteger(String(runtime.fetch_retries ?? fetchRetries), 0, 5)
    fetchTimeoutMs = clampInteger(String(runtime.fetch_timeout_ms ?? fetchTimeoutMs), 1000, 120000)
    workerConcurrency = clampInteger(String(runtime.worker_concurrency ?? workerConcurrency), 1, 8)
    workerLeaseSeconds = clampInteger(String(runtime.worker_lease_seconds ?? workerLeaseSeconds), 60, 3600)
    idleIntervalMs = clampInteger(String(runtime.worker_interval_ms ?? idleIntervalMs), 1000, 300000)
    maxBackoffMs = clampInteger(String(runtime.worker_max_backoff_ms ?? maxBackoffMs), idleIntervalMs, 300000)
    console.log('[wechat-worker] loaded runtime config')
  } catch (error) {
    console.warn('[wechat-worker] failed to load runtime config; using env fallback', error instanceof Error ? error.message : String(error))
  }
}

async function refreshRuntimeConfigIfDue(force = false) {
  const now = Date.now()
  if (!force && now < nextRuntimeConfigRefreshAt) {
    return
  }
  nextRuntimeConfigRefreshAt = now + runtimeConfigRefreshMs
  await loadRuntimeConfig()
}

async function logTaskEvent(taskId: number, leaseToken: string, payload: TaskLogPayload) {
  try {
    await apiPost(`/wechat/worker/tasks/${taskId}/logs`, {
      lease_token: leaseToken,
      event: payload.event,
      status: payload.status || 'running',
      message: payload.message || '',
      meta_json: JSON.stringify(payload.meta || {}),
    })
  } catch (error) {
    console.warn('[wechat-worker] task log write failed', { taskId, event: payload.event, error })
  }
}

function taskIdOf(task: WechatTask) {
  return Number(task.id ?? task.ID ?? 0)
}

function leaseTokenOf(claim: ClaimResponse) {
  return String(claim.lease_token ?? claim.leaseToken ?? '').trim()
}

function formatsOf(task: WechatTask): WechatExportFormat[] {
  const formats = task.formats ?? task.Formats ?? []
  const normalized = formats.filter((format): format is WechatExportFormat =>
    format === 'html' || format === 'markdown',
  )
  return normalized.length > 0 ? normalized : ['html', 'markdown']
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

function firstNumber(...values: unknown[]) {
  for (const value of values) {
    const parsed = Number.parseInt(String(value ?? ''), 10)
    if (Number.isFinite(parsed)) return parsed
  }
  return undefined
}

function articleURLParams(link: string) {
  try {
    const parsed = new URL(link)
    return parsed.searchParams
  } catch {
    return new URLSearchParams()
  }
}

async function fetchEngagementMetadata(
  article: WechatArticle,
  link: string,
  metadata: Record<string, unknown>,
  enabled: boolean,
): Promise<WechatEngagementResult> {
  if (!enabled) {
    return { engagementFetchStatus: 'skipped' }
  }
  const backendResult = await fetchBackendEngagementMetadata(article, link, metadata)
  if (backendResult?.engagementFetchStatus === 'fetched') {
    return backendResult
  }
  const directResult = await fetchDirectEngagementMetadata(link, metadata)
  if (directResult.engagementFetchStatus === 'fetched') {
    return directResult
  }
  return backendResult || directResult
}

async function fetchBackendEngagementMetadata(
  article: WechatArticle,
  link: string,
  metadata: Record<string, unknown>,
): Promise<WechatEngagementResult | null> {
  const articleId = Number(article.id ?? article.ID ?? 0)
  if (!articleId) return null
  try {
    const result = await apiPost<BackendWechatEngagementResult>(`/wechat/worker/articles/${articleId}/engagement`, {
      user_id: Number(article.user_id ?? article.UserID ?? 0),
      link,
      metadata_json: JSON.stringify(metadata),
    })
    return {
      readNum: firstNumber(result.read_num),
      oldLikeNum: firstNumber(result.old_like_num),
      shareNum: firstNumber(result.share_num),
      likeNum: firstNumber(result.like_num),
      commentNum: firstNumber(result.comment_num),
      engagementFetchStatus: result.status || 'unavailable',
      engagementFetchMessage: result.message || '',
    }
  } catch (error) {
    console.warn('[wechat-worker] backend engagement fetch failed, falling back to direct fetch', { articleId, error })
    return null
  }
}

async function fetchDirectEngagementMetadata(link: string, metadata: Record<string, unknown>): Promise<WechatEngagementResult> {
  const params = articleURLParams(link)
  const biz = String(metadata.biz || metadata.fakeid || params.get('__biz') || '').trim()
  const mid = String(metadata.mid || metadata.appmsgid || params.get('mid') || '').trim()
  const idx = String(metadata.idx || metadata.itemidx || params.get('idx') || '1').trim()
  const sn = String(metadata.sn || params.get('sn') || '').trim()
  const appmsgToken = String(metadata.appmsgToken || metadata.appmsg_token || params.get('appmsg_token') || '').trim()
  const existing = {
    readNum: firstNumber(metadata.readNum, metadata.read_num),
    oldLikeNum: firstNumber(metadata.oldLikeNum, metadata.old_like_num),
    shareNum: firstNumber(metadata.shareNum, metadata.share_num),
    likeNum: firstNumber(metadata.likeNum, metadata.like_num),
    commentNum: firstNumber(metadata.commentNum, metadata.comment_count),
  }
  if (!biz || !mid || !idx || !sn || !appmsgToken) {
    return {
      ...existing,
      engagementFetchStatus: Object.values(existing).some((value) => typeof value === 'number') ? 'unavailable' : 'unavailable',
      engagementFetchMessage: 'Missing __biz/mid/idx/sn/appmsg_token required by getappmsgext.',
    }
  }
  const endpoint = new URL('https://mp.weixin.qq.com/mp/getappmsgext')
  endpoint.searchParams.set('__biz', biz)
  endpoint.searchParams.set('mid', mid)
  endpoint.searchParams.set('idx', idx)
  endpoint.searchParams.set('sn', sn)
  endpoint.searchParams.set('appmsg_token', appmsgToken)
  endpoint.searchParams.set('x5', '0')
  endpoint.searchParams.set('f', 'json')
  try {
    const controller = new AbortController()
    const timeout = setTimeout(() => controller.abort(), fetchTimeoutMs)
    const response = await fetch(endpoint, {
      method: 'POST',
      signal: controller.signal,
      headers: {
        'content-type': 'application/x-www-form-urlencoded; charset=UTF-8',
        'user-agent':
          'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125 Safari/537.36',
        referer: link,
      },
      body: new URLSearchParams({
        is_only_read: '1',
        is_temp_url: '0',
        appmsg_type: '9',
      }),
    })
    clearTimeout(timeout)
    if (!response.ok) {
      return { ...existing, engagementFetchStatus: 'failed', engagementFetchMessage: `getappmsgext HTTP ${response.status}` }
      }
      const payload = await response.json() as any
    const appmsgstat = payload.appmsgstat || {}
    return {
      readNum: firstNumber(appmsgstat.read_num, payload.read_num, existing.readNum),
      oldLikeNum: firstNumber(appmsgstat.old_like_num, payload.old_like_num, existing.oldLikeNum),
      shareNum: firstNumber(appmsgstat.share_num, payload.share_num, existing.shareNum),
      likeNum: firstNumber(appmsgstat.like_num, payload.like_num, existing.likeNum),
      commentNum: firstNumber(appmsgstat.comment_count, payload.comment_count, existing.commentNum),
      engagementFetchStatus: 'fetched',
      engagementFetchMessage: payload.base_resp?.ret ? `ret=${payload.base_resp.ret} msg=${payload.base_resp.err_msg || ''}` : '',
    }
  } catch (error) {
    return {
      ...existing,
      engagementFetchStatus: 'failed',
      engagementFetchMessage: error instanceof Error ? error.message : String(error),
    }
  }
}

async function fetchArticleHtml(link: string, title: string) {
  if (!link) {
    return `<article><h1>${escapeHtml(title || 'Untitled')}</h1></article>`
  }
  if (!isAllowedWechatArticleURL(link)) {
    throw new Error(`non-wechat article url rejected: ${link}`)
  }
  for (let attempt = 0; attempt <= fetchRetries; attempt += 1) {
    const controller = new AbortController()
    const timeout = setTimeout(() => controller.abort(), fetchTimeoutMs)
    try {
      const response = await fetch(link, {
        signal: controller.signal,
        headers: {
          'user-agent':
            'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125 Safari/537.36',
          accept: 'text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8',
        },
      })
      if (response.ok) {
        const html = await response.text()
        assertWechatArticleHtml(html)
        return html
      }
      if (response.status >= 400 && response.status < 500) {
        break
      }
    } catch (error) {
      if (error instanceof Error && error.message === wechatVerifyPageMessage) {
        throw error
      }
      console.warn('[wechat-worker] article fetch attempt failed', { link, attempt: attempt + 1, error })
    } finally {
      clearTimeout(timeout)
    }
    if (attempt < fetchRetries) {
      await sleep(Math.min(1000 * 2 ** attempt, 8000))
    }
  }
  console.warn('[wechat-worker] article fetch failed, using fallback html', { link })
  return `<article><h1>${escapeHtml(title || link)}</h1><p><a href="${escapeHtml(link)}">${escapeHtml(link)}</a></p></article>`
}

function isAllowedWechatArticleURL(rawURL: string) {
  try {
    const parsed = new URL(rawURL)
    if (parsed.protocol !== 'https:' || parsed.hostname.toLowerCase() !== 'mp.weixin.qq.com') {
      return false
    }
    const pathName = parsed.pathname.replace(/\/+$/, '')
    return pathName === '/s' || pathName.startsWith('/s/') || pathName.startsWith('/mp/appmsg')
  } catch {
    return false
  }
}

async function enrichArticle(article: WechatArticle, summary: WechatArticleSummary, fallbackUserId: number) {
  const articleId = Number(article.id ?? article.ID ?? 0)
  const userId = Number(article.user_id ?? article.UserID ?? fallbackUserId)
  if (!articleId || !userId) return

  try {
    await apiPost(`/wechat/worker/articles/${articleId}/enrich`, {
      user_id: userId,
      account_fakeid: summary.fakeid,
      account_name: summary.accountName,
      account_alias: summary.accountAlias,
      account_avatar: summary.accountAvatar,
      account_description: summary.accountDescription,
      title: summary.title,
      author: summary.author,
      cover: summary.cover,
      digest: summary.digest,
      publish_at: summary.publishAt || undefined,
      is_original: summary.isOriginal,
      is_pay_subscribe: summary.isPaySubscribe,
      content_status: 'fetched',
      metadata_json: JSON.stringify(summary.metadataSeed || {}),
    })
  } catch (error) {
    console.warn('[wechat-worker] article enrichment failed', { articleId, error })
  }
}

function escapeHtml(value: string) {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}

async function buildRunnerArticles(taskId: number, leaseToken: string, articles: WechatArticle[], fallbackUserId: number, includeEngagement: boolean): Promise<BuiltArticlesResult> {
  const output: BuiltArticle[] = []
  const failures: Array<{ link: string; message: string }> = []
  for (const article of articles) {
    const articleId = Number(article.id ?? article.ID ?? 0)
    const link = articleField(article, 'link', 'Link')
    const fallbackTitle = articleField(article, 'title', 'Title') || link || `article-${articleId || 'unknown'}`
    try {
      await logTaskEvent(taskId, leaseToken, {
        event: 'article_fetch_started',
        message: `Fetching article: ${fallbackTitle}`,
        meta: { article_id: articleId, link, title: fallbackTitle },
      })
      const html = await fetchArticleHtml(link, fallbackTitle)
      const summary = extractWechatArticleSummaryFromHtml(html)
      const title = summary.title || fallbackTitle
      const accountName = summary.accountName || articleField(article, 'account_fakeid', 'AccountFakeID')
      await logTaskEvent(taskId, leaseToken, {
        event: 'article_fetched',
        message: `Fetched article HTML: ${title}`,
        meta: {
          article_id: articleId,
          link,
          title,
          account_name: accountName,
          html_bytes: Buffer.byteLength(html, 'utf8'),
          fallback_html: html.includes('<p><a href='),
        },
      })
      const metadata = {
        ...(articleMetadata(article) || {}),
        ...(summary.metadataSeed || {}),
      }
      const engagement = await fetchEngagementMetadata(article, link, metadata, includeEngagement)
      await logTaskEvent(taskId, leaseToken, {
        event: 'article_engagement_checked',
        message: engagement.engagementFetchStatus === 'fetched'
          ? `Fetched engagement metrics for ${title}.`
          : `Engagement metrics ${engagement.engagementFetchStatus}.`,
        meta: {
          article_id: articleId,
          link,
          title,
          engagement_status: engagement.engagementFetchStatus,
          engagement_message: engagement.engagementFetchMessage || '',
          has_read_num: typeof engagement.readNum === 'number',
          has_like_num: typeof engagement.likeNum === 'number',
          requested: includeEngagement,
        },
      })
      await enrichArticle(article, summary, fallbackUserId)
      await logTaskEvent(taskId, leaseToken, {
        event: 'article_enriched',
        message: `Saved parsed metadata for ${title}.`,
        meta: {
          article_id: articleId,
          link,
          title,
          account_name: accountName,
          content_status: 'fetched',
        },
      })
      output.push({
        rawArticle: article,
        summary,
        runnerArticle: {
          accountName,
          aid: String(article.id ?? article.ID ?? ''),
          link,
          title,
          html,
          metadata: {
            ...metadata,
            ...engagement,
          },
        },
      })
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error)
      failures.push({ link, message })
      await logTaskEvent(taskId, leaseToken, {
        event: 'article_failed',
        status: 'failed',
        message,
        meta: { article_id: articleId, link, title: fallbackTitle },
      })
      console.warn('[wechat-worker] article build failed', { link, error: message })
    }
  }
  return { articles: output, failedCount: failures.length, failures }
}

async function sha256File(filePath: string) {
  return new Promise<string>((resolve, reject) => {
    const hash = createHash('sha256')
    const stream = createReadStream(filePath)
    stream.on('error', reject)
    stream.on('data', (chunk) => hash.update(chunk))
    stream.on('end', () => resolve(hash.digest('hex')))
  })
}

function storageKeyForFile(filePath: string) {
  const relativePath = path.relative(outputRoot, filePath)
  if (relativePath.startsWith('..') || path.isAbsolute(relativePath)) {
    throw new Error(`generated artifact escaped output root: ${filePath}`)
  }
  return path.join(storageKeyRoot, relativePath)
}

async function runOnce() {
  const claim = await apiPost<ClaimResponse>('/wechat/worker/tasks/claim', { lease_seconds: workerLeaseSeconds })
  if (!claim.task) {
    console.log('[wechat-worker] no queued tasks')
    return false
  }
  const taskId = taskIdOf(claim.task)
  const leaseToken = leaseTokenOf(claim)
  if (!leaseToken) {
    throw new Error(`claimed task ${taskId || 'unknown'} without lease_token`)
  }
  try {
    const taskOutputDir = path.join(outputRoot, String(taskId))
    await mkdir(taskOutputDir, { recursive: true })
    const includeEngagement = Boolean(claim.task.include_engagement ?? claim.task.IncludeEngagement)
    const built = await buildRunnerArticles(taskId, leaseToken, claim.articles || [], Number(claim.task.user_id ?? claim.task.UserID ?? 0), includeEngagement)
    if (built.articles.length === 0) {
      throw new Error(`all ${built.failedCount} article(s) failed before export`)
    }
    const generated = await runWechatExportFormats({
      outputDir: taskOutputDir,
      taskId: String(taskId),
      formats: formatsOf(claim.task),
      includeEngagement,
      articles: built.articles.map((article) => article.runnerArticle),
    })
    await logTaskEvent(taskId, leaseToken, {
      event: 'artifacts_generated',
      message: `Generated ${generated.length} artifact file(s).`,
      meta: {
        artifact_count: generated.length,
        formats: formatsOf(claim.task),
        failed_article_count: built.failedCount,
      },
    })
    const artifacts = await Promise.all(
      generated.map(async (item) => {
        const fileStat = await stat(item.filePath)
        return {
          format: item.format,
          storage_provider: 'local',
          storage_key: storageKeyForFile(item.filePath),
          file_name: item.fileName,
          file_size: fileStat.size,
          checksum: await sha256File(item.filePath),
        }
      }),
    )
    await apiPost(`/wechat/worker/tasks/${taskId}/complete`, {
      lease_token: leaseToken,
      artifacts,
      failed_article_count: built.failedCount,
      result_manifest_json: JSON.stringify({
        artifacts,
        failed_articles: built.failures,
        worker: {
          concurrency: workerConcurrency,
          fetch_retries: fetchRetries,
          include_engagement: includeEngagement,
        },
      }),
    })
    console.log('[wechat-worker] completed task', taskId)
    return true
  } catch (error) {
    await apiPost(`/wechat/worker/tasks/${taskId}/fail`, {
      lease_token: leaseToken,
      message: error instanceof Error ? error.message : String(error),
    })
    throw error
  }
}

async function main() {
  await refreshRuntimeConfigIfDue(true)
  if (process.argv.includes('--healthcheck')) {
    await apiGet('/wechat/worker/health')
    console.log('[wechat-worker] healthcheck ok')
    return
  }

  const once = process.argv.includes('--once')
  let backoffMs = idleIntervalMs

  do {
    let processed = false
    try {
      await refreshRuntimeConfigIfDue()
      if (once || workerConcurrency === 1) {
        processed = await runOnce()
      } else {
        const results = await Promise.allSettled(Array.from({ length: workerConcurrency }, () => runOnce()))
        processed = results.some((result) => result.status === 'fulfilled' && result.value)
        for (const result of results) {
          if (result.status === 'rejected') {
            console.error('[wechat-worker] concurrent task run failed', result.reason)
          }
        }
      }
    } catch (error) {
      console.error('[wechat-worker] task run failed', error)
    }

    if (once) {
      break
    }

    backoffMs = processed ? idleIntervalMs : Math.min(backoffMs * 2, maxBackoffMs)
    await sleep(backoffMs)
  } while (true)
}

function sleep(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

main().catch((error) => {
  console.error('[wechat-worker] fatal error', error)
  process.exit(1)
})
