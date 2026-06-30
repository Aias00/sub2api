#!/usr/bin/env node
import { writeFileSync } from 'node:fs'

const apiBase = (process.env.API_BASE || 'http://127.0.0.1:8080/api/v1').replace(/\/$/, '')
const authHeader = process.env.AUTH_HEADER || ''
const taskIdEnv = process.env.WECHAT_EXPORT_ACCEPTANCE_TASK_ID || ''
const reportPath = process.env.WECHAT_EXPORT_ACCEPTANCE_REPORT_PATH || ''
const expectedFormats = new Set((process.env.WECHAT_EXPORT_ACCEPTANCE_FORMATS || 'html,markdown,json')
  .split(',')
  .map((item) => item.trim())
  .filter(Boolean))

function assert(condition, message) {
  if (!condition) throw new Error(message)
}

function authHeaders(extra = {}) {
  const headers = { accept: 'application/json', ...extra }
  const index = authHeader.indexOf(':')
  if (index > 0) {
    headers[authHeader.slice(0, index).trim()] = authHeader.slice(index + 1).trim()
  }
  return headers
}

async function api(path) {
  const response = await fetch(`${apiBase}${path}`, { headers: authHeaders() })
  const text = await response.text()
  let body
  try {
    body = text ? JSON.parse(text) : {}
  } catch {
    body = { raw: text }
  }
  if (!response.ok || (typeof body.code === 'number' && body.code !== 0)) {
    throw new Error(`GET ${path} failed: HTTP ${response.status} ${text}`)
  }
  return body.data
}

async function download(path, accept = '*/*') {
  const response = await fetch(`${apiBase}${path}`, { headers: authHeaders({ accept }) })
  const bytes = Buffer.from(await response.arrayBuffer())
  if (!response.ok) {
    throw new Error(`download ${path} failed: HTTP ${response.status} ${bytes.toString('utf8').slice(0, 200)}`)
  }
  return { response, bytes }
}

function selectTask(tasks) {
  if (taskIdEnv) {
    const id = Number(taskIdEnv)
    assert(Number.isInteger(id) && id > 0, 'WECHAT_EXPORT_ACCEPTANCE_TASK_ID must be a positive integer')
    return { id }
  }
  const items = Array.isArray(tasks.items) ? tasks.items : []
  const completed = items.find((task) => ['completed', 'completed_with_errors'].includes(task.status))
  assert(completed, 'no completed WeChat export task found; set WECHAT_EXPORT_ACCEPTANCE_TASK_ID after a real run')
  return completed
}

function parseManifest(task) {
  if (!task.result_manifest_json) return {}
  if (typeof task.result_manifest_json === 'object') return task.result_manifest_json
  try {
    return JSON.parse(task.result_manifest_json)
  } catch {
    return {}
  }
}

function hasMeaningfulText(value, minLength = 12) {
  return typeof value === 'string' && value.replace(/\s+/g, '').length >= minLength
}

function writeAcceptanceReport(report) {
  if (!reportPath) return
  writeFileSync(reportPath, `${JSON.stringify(report, null, 2)}\n`)
}

function assertArtifactContentQuality({ task, artifactTextsByFormat }) {
  const htmlTexts = artifactTextsByFormat.get('html') || []
  if (expectedFormats.has('html')) {
    assert(htmlTexts.length > 0, `task ${task.id} has no downloaded HTML artifact text`)
    for (const [index, html] of htmlTexts.entries()) {
      assert(/<!doctype html>|<html/i.test(html), `HTML artifact ${index + 1} does not look like standalone HTML`)
      assert(/wechat-export-article/.test(html), `HTML artifact ${index + 1} is missing exported article wrapper`)
      assert(/<h1[\s>]/i.test(html), `HTML artifact ${index + 1} is missing article title heading`)
      assert(/mp\.weixin\.qq\.com|__biz=|wechat/i.test(html), `HTML artifact ${index + 1} is missing source/account trace`)
    }
  }

  const markdownTexts = artifactTextsByFormat.get('markdown') || []
  if (expectedFormats.has('markdown')) {
    assert(markdownTexts.length > 0, `task ${task.id} has no downloaded Markdown artifact text`)
    for (const [index, markdown] of markdownTexts.entries()) {
      assert(/^#\s+\S+/m.test(markdown), `Markdown artifact ${index + 1} is missing title heading`)
      assert(/mp\.weixin\.qq\.com|Source:|来源|WeChat/i.test(markdown), `Markdown artifact ${index + 1} is missing source/account trace`)
      assert(hasMeaningfulText(markdown, 40), `Markdown artifact ${index + 1} has too little body text`)
    }
  }

  const jsonTexts = artifactTextsByFormat.get('json') || []
  if (expectedFormats.has('json')) {
    assert(jsonTexts.length > 0, `task ${task.id} has no downloaded JSON artifact text`)
    for (const [index, jsonText] of jsonTexts.entries()) {
      let records
      try {
        records = JSON.parse(jsonText)
      } catch (error) {
        throw new Error(`JSON artifact ${index + 1} is not valid JSON: ${error.message}`)
      }
      assert(Array.isArray(records), `JSON artifact ${index + 1} must be an array of article records`)
      assert(records.length > 0, `JSON artifact ${index + 1} has no article records`)
      for (const [recordIndex, record] of records.entries()) {
        assert(hasMeaningfulText(record.title, 2), `JSON record ${recordIndex + 1} is missing title`)
        assert(/^https:\/\/mp\.weixin\.qq\.com\//.test(record.link || ''), `JSON record ${recordIndex + 1} is missing mp.weixin.qq.com link`)
        assert(hasMeaningfulText(record.html, 20), `JSON record ${recordIndex + 1} is missing HTML body`)
        assert(hasMeaningfulText(record.text, 20), `JSON record ${recordIndex + 1} is missing extracted text`)
        assert(Array.isArray(record.richBlocks), `JSON record ${recordIndex + 1} is missing richBlocks array`)
        if (task.include_engagement || task.includeEngagement) {
          assert(record.engagement && typeof record.engagement.available === 'boolean', `JSON record ${recordIndex + 1} is missing engagement availability metadata`)
        }
      }
    }
  }
}

async function main() {
  assert(authHeader, 'AUTH_HEADER is required for WeChat export acceptance')

  const tasks = await api('/wechat/tasks?page=1&page_size=20')
  const selected = selectTask(tasks)
  const task = await api(`/wechat/tasks/${selected.id}`)
  assert(['completed', 'completed_with_errors'].includes(task.status), `task ${task.id} status is ${task.status}`)
  assert(Number(task.selected_article_count || 0) > 0, `task ${task.id} has no selected articles`)

  const artifactsData = await api(`/wechat/tasks/${task.id}/artifacts`)
  const artifacts = Array.isArray(artifactsData.items) ? artifactsData.items : []
  assert(artifacts.length > 0, `task ${task.id} has no artifacts`)

  const artifactFormats = new Set(artifacts.map((artifact) => artifact.format))
  for (const format of expectedFormats) {
    assert(artifactFormats.has(format), `task ${task.id} is missing ${format} artifact`)
  }

  const logsData = await api(`/wechat/tasks/${task.id}/logs`)
  const logs = Array.isArray(logsData.items) ? logsData.items : []
  assert(logs.length > 0, `task ${task.id} has no persisted task logs`)
  const events = new Set(logs.map((log) => log.event))
  for (const event of ['task_created', 'task_claimed', 'task_completed']) {
    assert(events.has(event), `task ${task.id} is missing persisted log event ${event}`)
  }
  assert([...events].some((event) => event.startsWith('article_')), `task ${task.id} has no article-level log events`)

  const manifest = parseManifest(task)
  assert(Object.keys(manifest).length > 0, `task ${task.id} has no result manifest`)
  const manifestText = JSON.stringify(manifest)
  assert(/engagement/i.test(manifestText), `task ${task.id} manifest does not mention engagement status/metadata`)

  let downloadedBytes = 0
  const artifactTextsByFormat = new Map()
  for (const artifact of artifacts) {
    assert(Number(artifact.file_size || 0) > 0, `artifact ${artifact.id} has invalid file_size`)
    const { bytes } = await download(`/wechat/artifacts/${artifact.id}/download`, '*/*')
    assert(bytes.length > 0, `artifact ${artifact.id} download is empty`)
    if (['html', 'markdown', 'json'].includes(artifact.format)) {
      const values = artifactTextsByFormat.get(artifact.format) || []
      values.push(bytes.toString('utf8'))
      artifactTextsByFormat.set(artifact.format, values)
    }
    downloadedBytes += bytes.length
  }

  assertArtifactContentQuality({ task, artifactTextsByFormat })

  const zip = await download(`/wechat/tasks/${task.id}/artifacts.zip`, 'application/zip')
  assert(zip.bytes.length > 4, `task ${task.id} zip download is empty`)
  assert(zip.bytes[0] === 0x50 && zip.bytes[1] === 0x4b, `task ${task.id} zip does not start with PK magic`)
  const report = {
    schema: 'sub2api-wechat-export-acceptance/v1',
    status: 'passed',
    generated_at: new Date().toISOString(),
    api_base: apiBase,
    task: {
      id: task.id,
      status: task.status,
      selected_article_count: task.selected_article_count,
      successful_article_count: task.successful_article_count,
      failed_article_count: task.failed_article_count,
      include_engagement: Boolean(task.include_engagement || task.includeEngagement),
    },
    artifacts: artifacts.map((artifact) => ({
      id: artifact.id,
      format: artifact.format,
      file_name: artifact.file_name,
      file_size: artifact.file_size,
      checksum: artifact.checksum,
      download_url: artifact.download_url,
    })),
    artifact_formats: [...artifactFormats].sort(),
    expected_formats: [...expectedFormats].sort(),
    logs: {
      count: logs.length,
      required_events: ['task_created', 'task_claimed', 'task_completed'],
      present_required_events: ['task_created', 'task_claimed', 'task_completed'].filter((event) => events.has(event)),
      article_event_count: logs.filter((log) => String(log.event || '').startsWith('article_')).length,
    },
    result_manifest: {
      present: Object.keys(manifest).length > 0,
      mentions_engagement: /engagement/i.test(manifestText),
    },
    downloads: {
      artifact_bytes: downloadedBytes,
      zip_bytes: zip.bytes.length,
      zip_has_pk_magic: zip.bytes[0] === 0x50 && zip.bytes[1] === 0x4b,
    },
    content_quality: {
      html_checked: expectedFormats.has('html'),
      markdown_checked: expectedFormats.has('markdown'),
      json_checked: expectedFormats.has('json'),
      passed: true,
    },
  }
  writeAcceptanceReport(report)

  console.log('# WeChat Export Completed Task Acceptance')
  console.log(`- Task ID: ${task.id}`)
  console.log(`- Status: ${task.status}`)
  console.log(`- Selected articles: ${task.selected_article_count}`)
  console.log(`- Artifacts: ${artifacts.length} (${[...artifactFormats].join(', ')})`)
  console.log(`- Persisted logs: ${logs.length}`)
  console.log(`- Downloaded artifact bytes: ${downloadedBytes}`)
  console.log(`- ZIP bytes: ${zip.bytes.length}`)
  console.log('- Artifact content quality: title/source/body/json-rich-block checks passed')
  if (reportPath) console.log(`- Report: ${reportPath}`)
  console.log('WeChat export completed task acceptance complete.')
}

main().catch((error) => {
  console.error(error instanceof Error ? error.stack || error.message : String(error))
  process.exit(1)
})
