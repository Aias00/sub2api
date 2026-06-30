#!/usr/bin/env node
import { createHash, randomUUID } from 'node:crypto'
import { existsSync, mkdirSync, readFileSync, writeFileSync } from 'node:fs'
import { dirname } from 'node:path'
import { spawnSync } from 'node:child_process'

const config = {
  databaseURL: env('DATABASE_URL', ''),
  statusPath: env('HOT_RSS_WORKER_STATUS_PATH', env('HOT_WORKER_STATUS_PATH', '/app/runtime/hot-worker-status.json')),
  intervalMs: intEnv('HOT_RSS_COLLECT_INTERVAL_MS', 30 * 60 * 1000),
  maxBackoffMs: intEnv('HOT_RSS_COLLECT_MAX_BACKOFF_MS', 10 * 60 * 1000),
  collectOnStart: boolEnv('HOT_RSS_COLLECT_ON_START', true),
  maxRuns: intEnv('HOT_RSS_COLLECT_MAX_RUNS', 0),
  once: process.argv.includes('--once') || boolEnv('HOT_RSS_COLLECT_ONCE', false),
  healthcheck: process.argv.includes('--healthcheck'),
  limitPerSource: intEnv('HOT_RSS_LIMIT_PER_SOURCE', 10),
  requestTimeoutMs: intEnv('HOT_RSS_REQUEST_TIMEOUT_MS', 20 * 1000),
  healthMaxAgeMs: intEnv('HOT_RSS_WORKER_HEALTH_MAX_AGE_MS', 0),
}

function env(key, fallback) {
  const value = process.env[key]
  return value === undefined || value === '' ? fallback : value
}

function intEnv(key, fallback) {
  const value = Number.parseInt(process.env[key] || '', 10)
  return Number.isFinite(value) && value > 0 ? value : fallback
}

function boolEnv(key, fallback) {
  const value = process.env[key]
  if (value === undefined || value === '') return fallback
  return ['1', 'true', 'yes', 'on'].includes(value.toLowerCase())
}

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

function validateConfig() {
  if (!config.databaseURL) {
    throw new Error('DATABASE_URL is required')
  }
}

function statusMaxAgeMs() {
  if (config.healthMaxAgeMs > 0) return config.healthMaxAgeMs
  return Math.max(config.intervalMs * 2, config.intervalMs + config.maxBackoffMs)
}

function readStatus() {
  if (!existsSync(config.statusPath)) return {}
  try {
    return JSON.parse(readFileSync(config.statusPath, 'utf8'))
  } catch {
    return {}
  }
}

function writeStatus(status, extra = {}) {
  const payload = {
    status,
    apply: true,
    storage: 'postgres',
    mode: config.once ? 'once' : 'loop',
    collect_on_start: config.collectOnStart,
    interval_ms: config.intervalMs,
    max_backoff_ms: config.maxBackoffMs,
    max_runs: config.maxRuns,
    limit_per_source: config.limitPerSource,
    updated_at: new Date().toISOString(),
    ...extra,
  }
  mkdirSync(dirname(config.statusPath), { recursive: true })
  writeFileSync(config.statusPath, `${JSON.stringify(payload, null, 2)}\n`)
}

function runHealthcheck() {
  validateConfig()
  if (!existsSync(config.statusPath)) {
    throw new Error(`HOT_RSS_WORKER_STATUS_PATH not found: ${config.statusPath}`)
  }
  const status = JSON.parse(readFileSync(config.statusPath, 'utf8'))
  if (status.status !== 'ok' && status.status !== 'running') {
    throw new Error(`hot rss worker unhealthy status: ${status.status || 'unknown'}`)
  }
  const updatedAt = Date.parse(status.updated_at || '')
  if (!Number.isFinite(updatedAt)) {
    throw new Error('hot rss worker status has invalid updated_at')
  }
  const ageMs = Date.now() - updatedAt
  if (ageMs > statusMaxAgeMs()) {
    throw new Error(`hot rss worker status is stale: ${ageMs}ms > ${statusMaxAgeMs()}ms`)
  }
  console.log(`[hot-rss-worker] healthcheck ok status=${status.status} age_ms=${ageMs}`)
}

function psql(sql, args = []) {
  const result = spawnSync('psql', [
    config.databaseURL,
    '-X',
    '-v',
    'ON_ERROR_STOP=1',
    '-q',
    ...args,
  ], {
    input: sql,
    encoding: 'utf8',
    maxBuffer: 64 * 1024 * 1024,
  })
  if (result.status !== 0) {
    if (result.error) {
      throw result.error
    }
    throw new Error(result.stderr || `psql exited with ${result.status}`)
  }
  return result.stdout
}

function psqlJSON(sql) {
  const stdout = psql(sql, ['-At'])
  const trimmed = stdout.trim()
  return trimmed ? JSON.parse(trimmed) : []
}

function psqlExec(sql) {
  psql(sql)
}

function sqlJSON(value, fallback) {
  const raw = value === null || value === undefined || value === '' ? fallback : value
  const json = typeof raw === 'string' ? raw : JSON.stringify(raw)
  try {
    JSON.parse(json)
    return `${sqlString(json)}::jsonb`
  } catch {
    return `${sqlString(fallback)}::jsonb`
  }
}

function sqlString(value) {
  if (value === null || value === undefined) return 'NULL'
  return `'${String(value).replaceAll("'", "''")}'`
}

function xmlDecode(value) {
  return String(value || '')
    .replaceAll('<![CDATA[', '')
    .replaceAll(']]>', '')
    .replaceAll('&amp;', '&')
    .replaceAll('&lt;', '<')
    .replaceAll('&gt;', '>')
    .replaceAll('&quot;', '"')
    .replaceAll('&apos;', "'")
    .replaceAll('&#39;', "'")
    .replaceAll(/\s+/g, ' ')
    .trim()
}

function stripTags(value) {
  return xmlDecode(String(value || '').replaceAll(/<[^>]+>/g, ' '))
}

function blocks(xml, tag) {
  return [...String(xml || '').matchAll(new RegExp(`<${tag}\\b[^>]*>([\\s\\S]*?)<\\/${tag}>`, 'gi'))].map((match) => match[1])
}

function tagValue(block, tag) {
  const match = String(block || '').match(new RegExp(`<${tag}\\b[^>]*>([\\s\\S]*?)<\\/${tag}>`, 'i'))
  return match ? xmlDecode(match[1]) : ''
}

function tagText(block, tag) {
  return stripTags(tagValue(block, tag))
}

function firstTagText(block, tags) {
  for (const tag of tags) {
    const value = tagText(block, tag)
    if (value) return value
  }
  return ''
}

function linkValue(block) {
  const atomLink = String(block || '').match(/<link\b[^>]*href=["']([^"']+)["'][^>]*>/i)
  if (atomLink?.[1]) return xmlDecode(atomLink[1])
  return tagText(block, 'link')
}

function categoryValues(block) {
  const values = blocks(block, 'category').map(stripTags).filter(Boolean)
  for (const match of String(block || '').matchAll(/<category\b[^>]*term=["']([^"']+)["'][^>]*>/gi)) {
    values.push(xmlDecode(match[1]))
  }
  return [...new Set(values)].slice(0, 8)
}

function normalizeDate(value) {
  if (!value) return ''
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? '' : date.toISOString()
}

function scoreForItem({ publishedAt, tags, title }, sourceConfig) {
  const category = sourceConfig.category || 'blog'
  const ageMs = publishedAt ? Date.now() - new Date(publishedAt).getTime() : Number.POSITIVE_INFINITY
  const ageDays = Number.isFinite(ageMs) ? ageMs / 86400000 : 999
  let score = category === 'official' ? 70 : 60
  if (ageDays <= 1) score += 25
  else if (ageDays <= 3) score += 15
  score += Math.min(tags.length * 2, 10)
  if (/\b(ai|agent|model|llm|copilot|image|video|search|reasoning)\b/i.test(title || '')) score += 8
  return Math.max(0, Math.min(100, Math.round(score)))
}

function sourceCategoryLabel(sourceConfig) {
  return sourceConfig.category === 'official' ? '官方信源' : '博客文章'
}

function sourceDisplayName(source) {
  return String(source.title || source.source_id || '').trim()
}

function sourceHandle(source, feedURL) {
  const config = sourceConfig(source)
  const configured = String(config.handle || config.source_handle || '').trim()
  if (configured) return configured
  try {
    return new URL(source.base_url || feedURL || '').hostname.replace(/^www\./, '')
  } catch {
    return String(source.source_id || '').trim()
  }
}

function sourceConfig(row) {
  if (row.config_json && typeof row.config_json === 'object') return row.config_json
  try {
    return JSON.parse(row.config_json || '{}')
  } catch {
    return {}
  }
}

function seedURLs(row) {
  if (Array.isArray(row.seed_urls_json)) return row.seed_urls_json.filter(Boolean)
  try {
    const urls = JSON.parse(row.seed_urls_json || '[]')
    return Array.isArray(urls) ? urls.filter(Boolean) : []
  } catch {
    return []
  }
}

async function fetchText(url) {
  const controller = new AbortController()
  const timeout = setTimeout(() => controller.abort(), config.requestTimeoutMs)
  try {
    const response = await fetch(url, {
      signal: controller.signal,
      headers: {
        accept: 'application/rss+xml, application/atom+xml, application/xml, text/xml, */*',
        'user-agent': 'sub2api-hot-rss-collector/1.0',
      },
    })
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`)
    }
    return await response.text()
  } finally {
    clearTimeout(timeout)
  }
}

function parseFeed(xml, source, url) {
  const entries = [...blocks(xml, 'item'), ...blocks(xml, 'entry')]
  return entries.slice(0, config.limitPerSource).map((entry) => {
    const title = firstTagText(entry, ['title'])
    const link = linkValue(entry)
    const summary = firstTagText(entry, ['description', 'summary', 'content', 'content:encoded'])
    const publishedAt = normalizeDate(firstTagText(entry, ['pubDate', 'published', 'updated', 'dc:date']))
    const author = firstTagText(entry, ['author', 'dc:creator', 'name'])
    const guid = firstTagText(entry, ['guid', 'id']) || link || title
    const tags = categoryValues(entry)
    const hashInput = [source.source_id, guid, link, title, summary].join('\n')
    const contentHash = createHash('sha256').update(hashInput).digest('hex')
    const externalID = `${url}::${source.source_id}:${guid || contentHash}`
    const sourceConfigValue = sourceConfig(source)
    const hotScore = scoreForItem({ publishedAt, tags, title }, sourceConfigValue)
    const reason = publishedAt && Date.now() - new Date(publishedAt).getTime() <= 86400000
      ? `${sourceCategoryLabel(sourceConfigValue)} · 当日热点`
      : `${sourceCategoryLabel(sourceConfigValue)} · RSS 采集`
    const metrics = {
      reason,
      hot_score: hotScore,
      source_category: sourceConfigValue.category || 'blog',
      source_category_label: sourceCategoryLabel(sourceConfigValue),
    }
    return {
      source_id: source.source_id,
      external_id: externalID,
      canonical_url: link,
      title,
      summary,
      body: summary,
      reason,
      published_at: publishedAt,
      author,
      source_name: sourceDisplayName(source),
      source_handle: sourceHandle(source, url),
      badge: sourceCategoryLabel(sourceConfigValue),
      score: String(hotScore),
      content_type: 'article',
      tags_json: JSON.stringify(tags),
      metrics_json: JSON.stringify(metrics),
      raw_ref_json: JSON.stringify({ feed_url: url, guid }),
      content_hash: contentHash,
    }
  }).filter((item) => item.title && item.external_id)
}

function loadSources() {
  return psqlJSON(`
    SELECT COALESCE(json_agg(row_to_json(source_rows)), '[]'::json)
    FROM (
    SELECT source_id, adapter_kind, title, description, enabled, base_url, seed_urls_json, config_json, created_at, updated_at
    FROM hot_sources
    WHERE enabled = TRUE AND adapter_kind = 'rss-generic'
    ORDER BY sort_order ASC, source_id
    ) AS source_rows
  `)
}

function writeRun({ runID, sourceID, status, summary, error, startedAt, finishedAt }) {
  const now = new Date().toISOString()
  const sql = `
INSERT INTO hot_runs (run_id, source_id, status, request_json, summary_json, error_message, created_at, updated_at, started_at, finished_at)
VALUES (${sqlString(runID)}, ${sqlString(sourceID)}, ${sqlString(status)}, ${sqlJSON({ source_id: sourceID, dry_run: false, limit: config.limitPerSource }, '{}')}, ${sqlJSON(summary, '{}')}, ${sqlString(error || '')}, ${sqlString(startedAt)}, ${sqlString(now)}, ${sqlString(startedAt)}, ${finishedAt ? sqlString(finishedAt) : 'NULL'})
ON CONFLICT(run_id) DO UPDATE SET status=excluded.status, summary_json=excluded.summary_json, error_message=excluded.error_message, updated_at=excluded.updated_at, finished_at=excluded.finished_at;
INSERT INTO hot_run_events (run_id, legacy_id, node, message, payload_json, created_at)
VALUES (${sqlString(runID)}, 1, 'rss-worker', ${sqlString(status === 'completed' ? 'RSS source collected' : 'RSS source failed')}, ${sqlJSON(summary, '{}')}, ${sqlString(now)})
ON CONFLICT(run_id, legacy_id) DO UPDATE SET node=excluded.node, message=excluded.message, payload_json=excluded.payload_json, created_at=excluded.created_at;
`
  psqlExec(sql)
}

function persistItems(items) {
  if (items.length === 0) return
  const now = new Date().toISOString()
  const lines = ['BEGIN;']
  for (const item of items) {
    lines.push(`
INSERT INTO hot_items (
  source_id, external_id, canonical_url, title, summary, body, reason, published_at,
  author, source_name, source_handle, badge, score, content_type, tags_json,
  metrics_json, raw_ref_json, content_hash, has_media, created_at, updated_at
) VALUES (
  ${sqlString(item.source_id)}, ${sqlString(item.external_id)}, ${sqlString(item.canonical_url || '')},
  ${sqlString(item.title)}, ${sqlString(item.summary || '')}, ${sqlString(item.body || '')}, ${sqlString(item.reason || '')},
  ${item.published_at ? sqlString(item.published_at) : 'NULL'}, ${sqlString(item.author || '')},
  ${sqlString(item.source_name || item.source_id)}, ${sqlString(item.source_handle || '')}, ${sqlString(item.badge || '')},
  ${sqlString(item.score || '')}, ${sqlString(item.content_type || 'article')}, ${sqlJSON(item.tags_json, '[]')},
  ${sqlJSON(item.metrics_json, '{}')}, ${sqlJSON(item.raw_ref_json, '{}')}, ${sqlString(item.content_hash || '')},
  FALSE, ${sqlString(now)}, ${sqlString(now)}
)
ON CONFLICT(source_id, external_id) DO UPDATE SET
  canonical_url=excluded.canonical_url,
  title=excluded.title,
  summary=excluded.summary,
  body=excluded.body,
  reason=excluded.reason,
  published_at=excluded.published_at,
  author=excluded.author,
  source_name=excluded.source_name,
  source_handle=excluded.source_handle,
  badge=excluded.badge,
  score=excluded.score,
  content_type=excluded.content_type,
  tags_json=excluded.tags_json,
  metrics_json=excluded.metrics_json,
  raw_ref_json=excluded.raw_ref_json,
  content_hash=excluded.content_hash,
  has_media=excluded.has_media,
  updated_at=excluded.updated_at;`)
  }
  lines.push('COMMIT;')
  psqlExec(lines.join('\n'))
}

async function collectSource(source) {
  const runID = `rss-${source.source_id}-${randomUUID()}`
  const startedAt = new Date().toISOString()
  const urls = seedURLs(source)
  try {
    const itemGroups = []
    const errors = []
    for (const url of urls) {
      try {
        const xml = await fetchText(url)
        itemGroups.push(parseFeed(xml, source, url))
      } catch (error) {
        errors.push(`${url}: ${error instanceof Error ? error.message : String(error)}`)
      }
    }
    const items = itemGroups.flat()
    persistItems(items)
    const finishedAt = new Date().toISOString()
    const summary = {
      discovered_count: items.length,
      hydrated_count: items.length,
      normalized_count: items.length,
      persisted_count: items.length,
      error_count: errors.length,
      errors: errors.slice(0, 3),
    }
    writeRun({ runID, sourceID: source.source_id, status: 'completed', summary, startedAt, finishedAt })
    psqlExec(`
INSERT INTO hot_checkpoints (source_id, checkpoint_json, updated_at)
VALUES (${sqlString(source.source_id)}, ${sqlJSON({ last_run_id: runID, last_collected_at: finishedAt, errors: errors.slice(0, 3) }, '{}')}, ${sqlString(finishedAt)})
ON CONFLICT(source_id) DO UPDATE SET checkpoint_json=excluded.checkpoint_json, updated_at=excluded.updated_at;
`)
    return { source_id: source.source_id, status: 'completed', summary }
  } catch (error) {
    const finishedAt = new Date().toISOString()
    const message = error instanceof Error ? error.message : String(error)
    writeRun({
      runID,
      sourceID: source.source_id,
      status: 'failed',
      summary: { discovered_count: 0, persisted_count: 0 },
      error: message,
      startedAt,
      finishedAt,
    })
    return { source_id: source.source_id, status: 'failed', error: message }
  }
}

async function runCollect() {
  validateConfig()
  const previousStatus = readStatus()
  const startedAt = new Date()
  writeStatus('running', {
    last_started_at: startedAt.toISOString(),
    run_count: Number(previousStatus.run_count || 0),
    success_count: Number(previousStatus.success_count || 0),
    failure_count: Number(previousStatus.failure_count || 0),
  })
  const sources = loadSources()
  console.log(`[hot-rss-worker] collect started sources=${sources.length}`)
  const results = []
  for (const source of sources) {
    results.push(await collectSource(source))
  }
  const failed = results.filter((result) => result.status !== 'completed')
  const finishedAt = new Date()
  const extra = {
    last_started_at: startedAt.toISOString(),
    last_finished_at: finishedAt.toISOString(),
    last_run_duration_ms: finishedAt.getTime() - startedAt.getTime(),
    source_count: sources.length,
    item_count: results.reduce((total, result) => total + Number(result.summary?.persisted_count || 0), 0),
    failed_source_count: failed.length,
    recent_results: results.slice(-10),
    run_count: Number(previousStatus.run_count || 0) + 1,
    success_count: Number(previousStatus.success_count || 0) + (failed.length === 0 ? 1 : 0),
    failure_count: Number(previousStatus.failure_count || 0) + (failed.length > 0 ? 1 : 0),
  }
  writeStatus(failed.length === 0 ? 'ok' : 'error', failed.length === 0 ? extra : { ...extra, error_message: `${failed.length} sources failed` })
  if (failed.length > 0) {
    throw new Error(`${failed.length} sources failed`)
  }
  console.log(`[hot-rss-worker] collect finished items=${extra.item_count}`)
}

async function loop() {
  let backoff = config.intervalMs
  let completedRuns = 0
  if (config.collectOnStart || config.once) {
    await runCollect()
    completedRuns += 1
  }
  if (config.once || (config.maxRuns > 0 && completedRuns >= config.maxRuns)) return
  for (;;) {
    await sleep(backoff)
    try {
      await runCollect()
      completedRuns += 1
      backoff = config.intervalMs
      if (config.maxRuns > 0 && completedRuns >= config.maxRuns) return
    } catch (error) {
      console.error('[hot-rss-worker] collect failed:', error)
      backoff = Math.min(backoff * 2, config.maxBackoffMs)
    }
  }
}

if (config.healthcheck) {
  try {
    runHealthcheck()
    process.exit(0)
  } catch (error) {
    console.error('[hot-rss-worker] healthcheck failed:', error)
    process.exit(1)
  }
}

void loop().catch((error) => {
  try {
    writeStatus('fatal', {
      last_failed_at: new Date().toISOString(),
      error_message: error instanceof Error ? error.message : String(error),
    })
  } catch {
    // Preserve the original fatal error.
  }
  console.error('[hot-rss-worker] fatal:', error)
  process.exit(1)
})
