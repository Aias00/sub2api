#!/usr/bin/env node
import { writeFileSync } from 'node:fs'

const apiBase = (process.env.API_BASE || 'http://127.0.0.1:8080/api/v1').replace(/\/$/, '')
const authHeader = process.env.AUTH_HEADER || ''
const taskIdEnv = process.env.IMAGE_WORKSPACE_ACCEPTANCE_TASK_ID || ''
const reportPath = process.env.IMAGE_WORKSPACE_ACCEPTANCE_REPORT_PATH || ''
const historyPageSize = positiveInt(process.env.IMAGE_WORKSPACE_ACCEPTANCE_HISTORY_PAGE_SIZE, 50)
const usagePageSize = positiveInt(process.env.IMAGE_WORKSPACE_ACCEPTANCE_USAGE_PAGE_SIZE, 50)
const timeoutMs = positiveInt(process.env.IMAGE_WORKSPACE_ACCEPTANCE_TIMEOUT_MS, 15000)

function positiveInt(value, fallback) {
  const parsed = Number.parseInt(String(value || ''), 10)
  return Number.isFinite(parsed) && parsed > 0 ? parsed : fallback
}

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

async function fetchWithTimeout(url, options = {}) {
  const controller = new AbortController()
  const timer = setTimeout(() => controller.abort(), timeoutMs)
  try {
    return await fetch(url, { ...options, signal: controller.signal })
  } finally {
    clearTimeout(timer)
  }
}

async function api(path) {
  const response = await fetchWithTimeout(`${apiBase}${path}`, { headers: authHeaders() })
  const text = await response.text()
  let body
  try {
    body = text ? JSON.parse(text) : {}
  } catch {
    body = { raw: text }
  }
  if (!response.ok || (typeof body.code === 'number' && body.code !== 0)) {
    throw new Error(`GET ${path} failed: HTTP ${response.status} ${text.slice(0, 240)}`)
  }
  return body.data ?? body
}

async function download(path, accept = '*/*') {
  const response = await fetchWithTimeout(`${apiBase}${path}`, { headers: authHeaders({ accept }) })
  const bytes = Buffer.from(await response.arrayBuffer())
  return {
    ok: response.ok,
    status: response.status,
    bytes,
    contentType: response.headers.get('content-type') || '',
  }
}

async function probeImageURL(url) {
  const head = await probe(url, { method: 'HEAD' })
  if (head.ok) return head
  if (![0, 403, 405, 501].includes(head.status)) return head
  const get = await probe(url, {
    method: 'GET',
    headers: { range: 'bytes=0-0', accept: 'image/*,*/*' },
  })
  return get.ok ? get : head
}

async function probe(url, options = {}) {
  try {
    const response = await fetchWithTimeout(url, options)
    return {
      ok: response.status >= 200 && response.status < 300,
      status: response.status,
      contentType: response.headers.get('content-type') || '',
    }
  } catch (error) {
    return { ok: false, status: 0, error: error instanceof Error ? error.message : String(error) }
  }
}

function parseJSON(value, label) {
  if (!value) return {}
  if (typeof value === 'object') return value
  try {
    return JSON.parse(value)
  } catch (error) {
    throw new Error(`${label} is not valid JSON: ${error.message}`)
  }
}

function selectTask(tasks) {
  if (taskIdEnv) {
    const id = Number(taskIdEnv)
    assert(Number.isInteger(id) && id > 0, 'IMAGE_WORKSPACE_ACCEPTANCE_TASK_ID must be a positive integer')
    return { id }
  }
  const items = Array.isArray(tasks.items) ? tasks.items : []
  const succeeded = items.find((task) => task.status === 'succeeded')
  assert(succeeded, 'no succeeded Image Workspace task found; set IMAGE_WORKSPACE_ACCEPTANCE_TASK_ID after a real provider/R2 run')
  return succeeded
}

function isPublicHTTPSURL(value) {
  try {
    const parsed = new URL(value)
    return parsed.protocol === 'https:' && !['localhost', '127.0.0.1', '0.0.0.0', '::1'].includes(parsed.hostname)
  } catch {
    return false
  }
}

function has64HexChecksum(value) {
  return typeof value === 'string' && /^[a-f0-9]{64}$/i.test(value)
}

function assertTaskResult({ task, artifacts }) {
  const result = parseJSON(task.result_json, 'task result_json')
  assert(result && Object.keys(result).length > 0, `task ${task.id} has no result_json`)
  assert(String(result.provider || '') === String(task.provider || ''), `result provider ${result.provider || ''} does not match task provider ${task.provider || ''}`)
  assert(String(result.model || '') === String(task.model || ''), `result model ${result.model || ''} does not match task model ${task.model || ''}`)
  assert(Number(result.artifact_count || 0) === artifacts.length, `result artifact_count ${result.artifact_count || 0} does not match artifacts ${artifacts.length}`)
  assert(Number(result.cost || 0) > 0, 'result cost is zero')
  assert(['flat', 'map', 'task_estimate'].includes(String(result.cost_source || '')), `unexpected result cost_source: ${result.cost_source || ''}`)
  return result
}

function assertArtifactShape(artifact, index) {
  assert(Number(artifact.id || 0) > 0, `artifact ${index + 1} has no id`)
  assert(Number(artifact.file_size || 0) > 0, `artifact ${artifact.id} has invalid file_size`)
  assert(String(artifact.mime_type || '').startsWith('image/'), `artifact ${artifact.id} has non-image mime_type ${artifact.mime_type || ''}`)
  assert(has64HexChecksum(artifact.checksum), `artifact ${artifact.id} checksum must be a SHA-256 hex digest`)
  assert(isPublicHTTPSURL(artifact.image_url), `artifact ${artifact.id} image_url is not a public HTTPS URL`)
  assert(String(artifact.storage_provider || ''), `artifact ${artifact.id} has no storage_provider`)
  assert(String(artifact.storage_key || ''), `artifact ${artifact.id} has no storage_key`)
  const metadata = parseJSON(artifact.metadata_json, `artifact ${artifact.id} metadata_json`)
  assert(Object.keys(metadata).length > 0, `artifact ${artifact.id} has no metadata_json`)
  return metadata
}

function assertUsageRecord({ usageRecord, task, artifacts }) {
  assert(usageRecord, `usage record for task ${task.id} was not found`)
  assert(String(usageRecord.billing_status || '') === 'settled', `usage billing_status is ${usageRecord?.billing_status || 'missing'}`)
  assert(Number(usageRecord.actual_cost || 0) > 0, 'usage actual_cost is zero')
  assert(Number.isFinite(Number(usageRecord.balance_snapshot)), 'usage balance_snapshot is missing')
  assert(String(usageRecord.provider || '') === String(task.provider || ''), 'usage provider does not match task provider')
  assert(String(usageRecord.model || '') === String(task.model || ''), 'usage model does not match task model')
  assert(String(usageRecord.size || '') === String(task.size || ''), 'usage size does not match task size')
  assert(String(usageRecord.quality || '') === String(task.quality || ''), 'usage quality does not match task quality')
  assert(Number(usageRecord.image_count || 0) === artifacts.length, `usage image_count ${usageRecord.image_count || 0} does not match artifacts ${artifacts.length}`)
  const metadata = parseJSON(usageRecord.metadata_json, 'usage metadata_json')
  assert(Number(metadata.artifact_count || 0) === artifacts.length, `usage metadata artifact_count ${metadata.artifact_count || 0} does not match artifacts ${artifacts.length}`)
  return metadata
}

function writeAcceptanceReport(report) {
  if (!reportPath) return
  writeFileSync(reportPath, `${JSON.stringify(report, null, 2)}\n`)
}

async function main() {
  assert(authHeader, 'AUTH_HEADER is required for Image Workspace acceptance')

  const tasks = await api(`/image-workspace/tasks?page=1&page_size=${historyPageSize}`)
  const selected = selectTask(tasks)
  const task = await api(`/image-workspace/tasks/${selected.id}`)
  assert(task.status === 'succeeded', `task ${task.id} status is ${task.status}`)
  assert(Number(task.cost_estimate || 0) > 0, `task ${task.id} cost_estimate is zero`)
  assert(Number.isFinite(Number(task.balance_snapshot)), `task ${task.id} balance_snapshot is missing`)
  assert(Number(task.batch_size || 0) > 0, `task ${task.id} batch_size is invalid`)

  const historyItems = Array.isArray(tasks.items) ? tasks.items : []
  assert(historyItems.some((item) => String(item.id) === String(task.id)), `task ${task.id} was not found in task history page`)

  const modelsData = await api('/image-workspace/models')
  const models = Array.isArray(modelsData.models) ? modelsData.models : []
  const model = models.find((item) => String(item.id || '') === String(task.model || ''))
  assert(model, `task model ${task.model || 'missing'} was not found in Image Workspace models`)
  assert(model.enabled !== false, `task model ${task.model} is disabled`)

  const templatesData = await api('/image-workspace/templates')
  const templates = Array.isArray(templatesData.items) ? templatesData.items : []
  const templateID = Number(task.template_id || 0)
  if (templateID > 0) {
    assert(templates.some((item) => Number(item.id || 0) === templateID), `task template ${templateID} was not found`)
  }

  const artifacts = Array.isArray(task.artifacts) ? task.artifacts : []
  assert(artifacts.length > 0, `task ${task.id} has no artifacts`)
  const result = assertTaskResult({ task, artifacts })

  const artifactURLChecks = []
  const localDownloadChecks = []
  const artifactMetadata = []
  for (const [index, artifact] of artifacts.entries()) {
    artifactMetadata.push(assertArtifactShape(artifact, index))
    const urlCheck = await probeImageURL(artifact.image_url)
    artifactURLChecks.push({ id: artifact.id, url: artifact.image_url, ...urlCheck })
    assert(urlCheck.ok, `artifact ${artifact.id} public URL is not reachable: HTTP ${urlCheck.status}${urlCheck.error ? ` ${urlCheck.error}` : ''}`)

    const downloadResult = await download(`/image-workspace/artifacts/${artifact.id}/download`, 'image/*,*/*')
    localDownloadChecks.push({
      id: artifact.id,
      ok: downloadResult.ok,
      status: downloadResult.status,
      bytes: downloadResult.bytes.length,
      contentType: downloadResult.contentType,
    })
  }

  const usageData = await api(`/image-workspace/usage-records?page=1&page_size=${usagePageSize}`)
  const usageItems = Array.isArray(usageData.items) ? usageData.items : []
  const usageRecord = usageItems.find((item) => String(item.task_id || '') === String(task.id))
  const usageMetadata = assertUsageRecord({ usageRecord, task, artifacts })

  const storageProviders = [...new Set(artifacts.map((artifact) => String(artifact.storage_provider || '')).filter(Boolean))]
  const localDownloadSucceeded = localDownloadChecks.filter((item) => item.ok && item.bytes > 0).length
  const report = {
    schema: 'sub2api-image-workspace-acceptance/v1',
    status: 'passed',
    generated_at: new Date().toISOString(),
    api_base: apiBase,
    task: {
      id: task.id,
      status: task.status,
      provider: task.provider,
      model: task.model,
      size: task.size,
      quality: task.quality,
      style: task.style,
      batch_size: task.batch_size,
      template_id: templateID > 0 ? templateID : null,
      cost_estimate: task.cost_estimate,
      balance_snapshot: task.balance_snapshot,
    },
    model: {
      id: model.id,
      label: model.label || '',
      provider: model.provider || '',
      enabled: model.enabled !== false,
    },
    result: {
      provider: result.provider,
      model: result.model,
      artifact_count: result.artifact_count,
      cost: result.cost,
      cost_source: result.cost_source,
    },
    artifacts: artifacts.map((artifact) => ({
      id: artifact.id,
      storage_provider: artifact.storage_provider,
      storage_key: artifact.storage_key,
      image_url: artifact.image_url,
      mime_type: artifact.mime_type,
      file_size: artifact.file_size,
      checksum: artifact.checksum,
    })),
    artifact_url_checks: artifactURLChecks,
    local_download_checks: localDownloadChecks,
    usage: {
      id: usageRecord.id,
      task_id: usageRecord.task_id,
      provider: usageRecord.provider,
      model: usageRecord.model,
      size: usageRecord.size,
      quality: usageRecord.quality,
      image_count: usageRecord.image_count,
      reserved_cost: usageRecord.reserved_cost,
      actual_cost: usageRecord.actual_cost,
      balance_snapshot: usageRecord.balance_snapshot,
      billing_status: usageRecord.billing_status,
      metadata_artifact_count: usageMetadata.artifact_count,
    },
    checks: {
      task_history_found: true,
      task_model_enabled: model.enabled !== false,
      task_template_found: true,
      public_artifact_url_count: artifactURLChecks.length,
      public_artifact_url_reachable_count: artifactURLChecks.filter((item) => item.ok).length,
      local_download_succeeded_count: localDownloadSucceeded,
      artifact_metadata_count: artifactMetadata.length,
      usage_record_found: true,
      usage_settled: usageRecord.billing_status === 'settled',
    },
  }
  writeAcceptanceReport(report)

  console.log('# Image Workspace Completed Task Acceptance')
  console.log(`- Task ID: ${task.id}`)
  console.log(`- Status: ${task.status}`)
  console.log(`- Model: ${task.provider}/${task.model} (${model.label || model.id})`)
  console.log(`- Size/quality/style: ${task.size}/${task.quality}/${task.style || ''}`)
  console.log(`- Artifacts: ${artifacts.length} (${storageProviders.join(', ')})`)
  console.log(`- Public artifact URL checks: ${artifactURLChecks.length}`)
  console.log(`- Local download checks succeeded: ${localDownloadSucceeded}/${localDownloadChecks.length}`)
  console.log(`- Usage billing: ${usageRecord.billing_status}, actual_cost=${usageRecord.actual_cost}, balance_snapshot=${usageRecord.balance_snapshot}`)
  console.log(`- Result cost: ${result.cost}, cost_source=${result.cost_source}`)
  console.log(`- Artifact metadata records: ${artifactMetadata.length}`)
  console.log(`- Usage metadata artifact_count: ${usageMetadata.artifact_count}`)
  if (reportPath) console.log(`- Report: ${reportPath}`)
  console.log('Image Workspace completed task acceptance complete.')
}

main().catch((error) => {
  console.error(error instanceof Error ? error.stack || error.message : String(error))
  process.exit(1)
})
