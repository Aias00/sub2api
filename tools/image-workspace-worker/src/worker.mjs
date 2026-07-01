#!/usr/bin/env node
import { createHash, createHmac } from 'node:crypto'
import { mkdir, writeFile } from 'node:fs/promises'
import { join, posix } from 'node:path'

let config = {
  baseURL: env('IMAGE_WORKSPACE_API_BASE_URL', 'http://127.0.0.1:8080'),
  workerToken: env('IMAGE_WORKSPACE_WORKER_TOKEN', ''),
  upstreamURL: '',
  upstreamAPIKey: env('IMAGE_WORKSPACE_UPSTREAM_API_KEY', ''),
  outputDir: env('IMAGE_WORKSPACE_OUTPUT_DIR', join(process.cwd(), 'runtime', 'image-workspace')),
  storageKeyRoot: env('IMAGE_WORKSPACE_STORAGE_KEY_ROOT', ''),
  publicArtifactBaseURL: env('IMAGE_WORKSPACE_PUBLIC_ARTIFACT_BASE_URL', ''),
  publicStorageKeyRoot: env('IMAGE_WORKSPACE_PUBLIC_STORAGE_KEY_ROOT', ''),
  intervalMs: intEnv('IMAGE_WORKSPACE_WORKER_INTERVAL_MS', 5000),
  maxBackoffMs: intEnv('IMAGE_WORKSPACE_WORKER_MAX_BACKOFF_MS', 60000),
  leaseSeconds: intEnv('IMAGE_WORKSPACE_WORKER_LEASE_SECONDS', 300),
  runtimeConfigRefreshMs: intEnv('IMAGE_WORKSPACE_RUNTIME_CONFIG_REFRESH_MS', 60000),
  requestTimeoutMs: intEnv('IMAGE_WORKSPACE_GENERATION_TIMEOUT_MS', 120000),
  generationRetries: intEnv('IMAGE_WORKSPACE_GENERATION_RETRIES', 2),
  generationRetryBaseMs: intEnv('IMAGE_WORKSPACE_GENERATION_RETRY_BASE_MS', 3000),
  diagnosticBodyPreviewBytes: intEnv('IMAGE_WORKSPACE_UPSTREAM_DIAGNOSTIC_BODY_PREVIEW_BYTES', 4096),
  completionCost: 0,
  completionCostMap: {},
  objectStorage: {
    enabled: false,
    provider: 'r2',
    endpoint: env('IMAGE_WORKSPACE_OBJECT_STORAGE_ENDPOINT', env('IMAGE_WORKSPACE_R2_ENDPOINT', '')),
    accountID: env('IMAGE_WORKSPACE_R2_ACCOUNT_ID', ''),
    accessKeyID: env('IMAGE_WORKSPACE_OBJECT_STORAGE_ACCESS_KEY_ID', env('IMAGE_WORKSPACE_R2_ACCESS_KEY_ID', env('IMAGE_WORKSPACE_R2_ACCESS_KEY', ''))),
    secretAccessKey: env('IMAGE_WORKSPACE_OBJECT_STORAGE_SECRET_ACCESS_KEY', env('IMAGE_WORKSPACE_R2_SECRET_ACCESS_KEY', env('IMAGE_WORKSPACE_R2_SECRET_KEY', ''))),
    bucket: '',
    region: 'auto',
    keyPrefix: 'image-workspace',
    publicBaseURL: '',
    cacheControl: env('IMAGE_WORKSPACE_OBJECT_STORAGE_CACHE_CONTROL', 'public, max-age=31536000, immutable'),
  },
}

let nextRuntimeConfigRefreshAt = 0
let runtimeConfigLoaded = false

function env(key, fallback) {
  const value = process.env[key]
  return value === undefined || value === '' ? fallback : value
}

function intEnv(key, fallback) {
  const value = Number.parseInt(process.env[key] || '', 10)
  return Number.isFinite(value) && value > 0 ? value : fallback
}

function floatEnv(key, fallback) {
  const value = Number.parseFloat(process.env[key] || '')
  return Number.isFinite(value) && value >= 0 ? value : fallback
}

function boolEnv(key, fallback) {
  const value = process.env[key]
  if (value === undefined || value === '') return fallback
  return ['1', 'true', 'yes', 'on'].includes(value.toLowerCase())
}

function parseObjectJSON(value, fallback) {
  if (!value) return fallback
  if (typeof value === 'object' && !Array.isArray(value)) return value
  try {
    const parsed = JSON.parse(String(value))
    return parsed && typeof parsed === 'object' && !Array.isArray(parsed) ? parsed : fallback
  } catch {
    return fallback
  }
}

function trimSlashes(value) {
  return String(value || '').replace(/^\/+|\/+$/g, '')
}

function upstreamURLDiagnostics() {
  try {
    const url = new URL(config.upstreamURL)
    return {
      upstream_host: url.host,
      upstream_path: url.pathname,
    }
  } catch {
    return {
      upstream_host: '',
      upstream_path: '',
    }
  }
}

function redactDiagnosticText(value) {
  let text = String(value || '')
  if (config.upstreamAPIKey) {
    text = text.replaceAll(config.upstreamAPIKey, '[redacted]')
  }
  return text
    .replace(/Bearer\s+[A-Za-z0-9._~+/=-]+/gi, 'Bearer [redacted]')
    .replace(/sk-[A-Za-z0-9_-]{8,}/g, 'sk-[redacted]')
    .slice(0, config.diagnosticBodyPreviewBytes)
}

function bodyShape(body) {
  if (body === null) return 'null'
  if (Array.isArray(body)) return `array:${body.length}`
  if (typeof body !== 'object') return typeof body
  const keys = Object.keys(body).slice(0, 20)
  const dataShape = Array.isArray(body.data) ? `data:${body.data.length}` : `data:${typeof body.data}`
  return `${dataShape};keys:${keys.join(',')}`
}

function buildUpstreamDiagnostics(task, response, text, body) {
  return {
    provider: task.provider || 'openai',
    model: task.model || '',
    ...upstreamURLDiagnostics(),
    upstream_status: response.status,
    upstream_status_text: response.statusText || '',
    upstream_content_type: response.headers.get('content-type') || '',
    upstream_request_id: response.headers.get('x-request-id') || response.headers.get('request-id') || response.headers.get('cf-ray') || '',
    upstream_body_shape: bodyShape(body),
    upstream_body_preview: redactDiagnosticText(text),
    captured_at: new Date().toISOString(),
  }
}

function errorWithUpstreamDiagnostics(message, diagnostics) {
  const error = new Error(message)
  error.upstreamDiagnostics = diagnostics
  return error
}

function apiURL(path) {
  return `${config.baseURL.replace(/\/+$/, '')}/api/v1${path}`
}

function workerHeaders(extra = {}) {
  const headers = { ...extra }
  if (config.workerToken) {
    headers['X-Image-Workspace-Worker-Token'] = config.workerToken
  }
  return headers
}

async function apiFetch(path, options = {}) {
  const response = await fetch(apiURL(path), {
    ...options,
    headers: workerHeaders({
      'content-type': 'application/json',
      ...(options.headers || {}),
    }),
  })
  const text = await response.text()
  let body
  try {
    body = text ? JSON.parse(text) : {}
  } catch {
    body = { message: text }
  }
  if (!response.ok || body.code) {
    throw new Error(body.message || `API ${response.status}`)
  }
  return body.data
}

async function loadRuntimeConfig({ required = false } = {}) {
  try {
    const runtime = await apiFetch('/image-workspace/worker/runtime-config', { method: 'GET' })
    if (!runtime || typeof runtime !== 'object') {
      throw new Error('runtime config response is empty')
    }
    if (typeof runtime.upstream_url === 'string' && runtime.upstream_url.trim()) {
      config.upstreamURL = runtime.upstream_url.trim()
    }
    if (Number.isFinite(runtime.generation_timeout_ms) && runtime.generation_timeout_ms > 0) {
      config.requestTimeoutMs = runtime.generation_timeout_ms
    }
    if (runtime.completion_cost !== undefined && runtime.completion_cost !== null) {
      const completionCost = Number.parseFloat(String(runtime.completion_cost))
      if (Number.isFinite(completionCost) && completionCost >= 0) {
        config.completionCost = completionCost
      }
    }
    if (runtime.completion_cost_map_json !== undefined && runtime.completion_cost_map_json !== null) {
      config.completionCostMap = parseObjectJSON(runtime.completion_cost_map_json, config.completionCostMap)
    }
    if (typeof runtime.prompt_safety_enabled === 'boolean') {
      process.env.IMAGE_WORKSPACE_PROMPT_SAFETY_ENABLED = String(runtime.prompt_safety_enabled)
    }
    const objectStorage = runtime.object_storage
    if (objectStorage && typeof objectStorage === 'object') {
      if (typeof objectStorage.enabled === 'boolean') config.objectStorage.enabled = objectStorage.enabled
      if (typeof objectStorage.provider === 'string' && objectStorage.provider.trim()) config.objectStorage.provider = objectStorage.provider.trim()
      if (typeof objectStorage.bucket === 'string') config.objectStorage.bucket = objectStorage.bucket.trim()
      if (typeof objectStorage.region === 'string' && objectStorage.region.trim()) config.objectStorage.region = objectStorage.region.trim()
      if (typeof objectStorage.key_prefix === 'string' && objectStorage.key_prefix.trim()) config.objectStorage.keyPrefix = trimSlashes(objectStorage.key_prefix)
      if (typeof objectStorage.public_base_url === 'string') config.objectStorage.publicBaseURL = objectStorage.public_base_url.trim()
    }
    runtimeConfigLoaded = true
    console.log('[image-workspace-worker] loaded runtime config')
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error)
    if (required) {
      throw new Error(`[image-workspace-worker] failed to load required runtime config: ${message}`)
    }
    console.warn('[image-workspace-worker] failed to refresh runtime config; keeping existing config', message)
  }
}

async function refreshRuntimeConfigIfDue(force = false, options = {}) {
  const now = Date.now()
  if (!force && now < nextRuntimeConfigRefreshAt) {
    return
  }
  nextRuntimeConfigRefreshAt = now + config.runtimeConfigRefreshMs
  await loadRuntimeConfig(options)
}

async function claimTask() {
  const data = await apiFetch('/image-workspace/worker/tasks/claim', {
    method: 'POST',
    body: JSON.stringify({ lease_seconds: config.leaseSeconds }),
  })
  return data?.task || null
}

function buildGenerationBody(task) {
  const promptParts = [task.prompt]
  if (task.negative_prompt) {
    promptParts.push(`\nNegative prompt: ${task.negative_prompt}`)
  }
  if (task.style) {
    promptParts.push(`\nStyle note: ${task.style}`)
  }
  return {
    model: task.model || 'gpt-image-2',
    prompt: promptParts.join(''),
    size: task.size || '1024x1024',
    quality: task.quality || 'standard',
    n: Math.max(1, Math.min(Number(task.batch_size || 1), 4)),
    response_format: 'b64_json',
  }
}

function validatePromptSafety(task) {
  if (!boolEnv('IMAGE_WORKSPACE_PROMPT_SAFETY_ENABLED', true)) {
    return
  }
  const text = [
    task.prompt || '',
    task.negative_prompt || '',
    task.style || '',
  ].join('\n').toLowerCase()
  const explicitTerms = [
    'panty',
    'panties',
    'underwear',
    'lingerie',
    'crotch',
    'open crotch',
    'legs spread',
    'provocative',
    'seductive',
    'bra',
    'nude',
    'sex',
  ]
  const youthContext = [
    'school uniform',
    'student',
    'teen',
    'teenager',
    'teenage',
    'young girl',
    'underage',
    'minor',
  ]
  const hasExplicitTerm = explicitTerms.some((term) => wordMatch(text, term))
  const hasYouthContext = youthContext.some((term) => wordMatch(text, term))
  if (hasExplicitTerm && hasYouthContext) {
    throw new Error('提示词包含露骨性内容或校园/未成年语境的性化描写，已在本地安全检查中拒绝。请改为非露骨、非性化的创意描述后重试。')
  }
  if (hasExplicitTerm && wordMatch(text, 'young')) {
    throw new Error('提示词包含年轻人物的露骨性化描写，已在本地安全检查中拒绝。请移除露骨身体部位、内衣展示或挑逗姿势等描述。')
  }
}

/**
 * Match a keyword as a whole word (word-boundary) rather than a substring.
 * Prevents false positives like "vibrant" matching "bra" or "library" matching "bra".
 */
function wordMatch(text, term) {
  const escaped = term.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  return new RegExp(`\\b${escaped}\\b`).test(text)
}

async function generateImages(task) {
  if (!config.upstreamAPIKey) {
    throw new Error('IMAGE_WORKSPACE_UPSTREAM_API_KEY is required')
  }
  validatePromptSafety(task)
  const controller = new AbortController()
  const timeout = setTimeout(() => controller.abort(new Error('generation timeout')), config.requestTimeoutMs)
  try {
    const response = await fetch(config.upstreamURL, {
      method: 'POST',
      headers: {
        authorization: `Bearer ${config.upstreamAPIKey}`,
        'content-type': 'application/json',
      },
      body: JSON.stringify(buildGenerationBody(task)),
      signal: controller.signal,
    })
    const text = await response.text()
    let body
    try {
      body = text ? JSON.parse(text) : {}
    } catch {
      body = { raw: text }
    }
    const diagnostics = buildUpstreamDiagnostics(task, response, text, body)
    if (!response.ok) {
      const message = body?.error?.message || body?.message || `upstream ${response.status}`
      if (response.status === 404) {
        throw errorWithUpstreamDiagnostics(`生图上游返回 404：当前后台 Runtime Settings 未提供 OpenAI 图片生成端点，或该服务不支持所选模型 ${task.model || 'default'}。请确认上游是否支持 /v1/images/generations。`, diagnostics)
      }
      if (response.status === 401 || response.status === 403) {
        throw errorWithUpstreamDiagnostics('生图上游鉴权失败：请检查 IMAGE_WORKSPACE_UPSTREAM_API_KEY 是否有效，或该服务是否使用非 Bearer 鉴权。', diagnostics)
      }
      throw errorWithUpstreamDiagnostics(message, diagnostics)
    }
    const items = Array.isArray(body.data) ? body.data : []
    if (items.length === 0) {
      throw errorWithUpstreamDiagnostics('upstream returned no images', diagnostics)
    }
    return { body, items }
  } finally {
    clearTimeout(timeout)
  }
}

function inferMime(item) {
  if (item.mime_type) return item.mime_type
  if (item.output_format === 'jpeg' || item.output_format === 'jpg') return 'image/jpeg'
  if (item.output_format === 'webp') return 'image/webp'
  return 'image/png'
}

function extensionForMime(mime) {
  if (mime === 'image/jpeg') return 'jpg'
  if (mime === 'image/webp') return 'webp'
  return 'png'
}

async function writeArtifacts(task, items) {
  const taskDir = join(config.outputDir, String(task.user_id), String(task.id))
  if (!config.objectStorage.enabled) {
    await mkdir(taskDir, { recursive: true })
  }
  const artifacts = []
  for (const [index, item] of items.entries()) {
    const mimeType = inferMime(item)
    const b64 = item.b64_json || item.base64 || item.image_base64 || ''
    const url = item.url || ''
    let storageKey = ''
    let checksum = ''
    let fileSize = 0
    let imageURL = url
    if (b64) {
      const bytes = Buffer.from(b64, 'base64')
      checksum = createHash('sha256').update(bytes).digest('hex')
      fileSize = bytes.length
      const filename = `image-${index + 1}.${extensionForMime(mimeType)}`
      if (config.objectStorage.enabled) {
        const uploaded = await uploadImageArtifact(task, filename, bytes, mimeType)
        storageKey = uploaded.storageKey
        imageURL = uploaded.imageURL
      } else {
        const filePath = join(taskDir, filename)
        await writeFile(filePath, bytes)
        storageKey = config.storageKeyRoot
          ? posix.join(config.storageKeyRoot, String(task.user_id), String(task.id), filename)
          : filePath
        imageURL = publicArtifactURL(task, filename) || `data:${mimeType};base64,${b64}`
      }
    }
    artifacts.push({
      storage_provider: b64 ? artifactStorageProvider() : 'upstream',
      storage_key: storageKey,
      image_url: imageURL,
      prompt: item.revised_prompt || task.prompt,
      mime_type: mimeType,
      width: 0,
      height: 0,
      file_size: fileSize,
      checksum,
      metadata_json: JSON.stringify({
        upstream_url: config.upstreamURL,
        model: task.model,
        revised_prompt: item.revised_prompt || '',
      }),
    })
  }
  return artifacts
}

function artifactStorageProvider() {
  return config.objectStorage.enabled ? config.objectStorage.provider || 'object_storage' : 'local'
}

function publicArtifactURL(task, filename) {
  if (!config.publicArtifactBaseURL) {
    return ''
  }
  const publicKey = posix.join(config.publicStorageKeyRoot, String(task.user_id), String(task.id), filename)
  return `${config.publicArtifactBaseURL.replace(/\/+$/, '')}/${publicKey.replace(/^\/+/, '')}`
}

async function uploadImageArtifact(task, filename, bytes, mimeType) {
  const storage = resolveObjectStorageConfig()
  const objectKey = posix.join(storage.keyPrefix, String(task.user_id), String(task.id), filename)
  const url = new URL(`${storage.endpoint.replace(/\/+$/, '')}/${storage.bucket}/${objectKey}`)
  const payloadHash = sha256Hex(bytes)
  const headers = signedS3Headers({
    method: 'PUT',
    url,
    bodyHash: payloadHash,
    contentType: mimeType,
    cacheControl: storage.cacheControl,
    accessKeyID: storage.accessKeyID,
    secretAccessKey: storage.secretAccessKey,
    region: storage.region,
  })
  const response = await fetch(url, {
    method: 'PUT',
    headers,
    body: bytes,
  })
  if (!response.ok) {
    const text = await response.text().catch(() => '')
    throw new Error(`object storage upload failed: ${response.status} ${response.statusText}${text ? ` ${text.slice(0, 200)}` : ''}`)
  }
  return {
    storageKey: objectKey,
    imageURL: objectPublicURL(storage, objectKey) || url.toString(),
  }
}

function resolveObjectStorageConfig() {
  const storage = { ...config.objectStorage }
  if (!storage.endpoint) {
    if (!storage.accountID) {
      throw new Error('IMAGE_WORKSPACE_OBJECT_STORAGE_ENDPOINT or IMAGE_WORKSPACE_R2_ACCOUNT_ID is required')
    }
    storage.endpoint = `https://${storage.accountID}.r2.cloudflarestorage.com`
  }
  for (const [key, value] of Object.entries({
    accessKeyID: storage.accessKeyID,
    secretAccessKey: storage.secretAccessKey,
    bucket: storage.bucket,
  })) {
    if (!value) {
      throw new Error(`IMAGE_WORKSPACE_OBJECT_STORAGE_${key} is required`)
    }
  }
  return storage
}

function objectPublicURL(storage, objectKey) {
  if (!storage.publicBaseURL) return ''
  return `${storage.publicBaseURL.replace(/\/+$/, '')}/${objectKey.replace(/^\/+/, '')}`
}

function signedS3Headers({ method, url, bodyHash, contentType, cacheControl, accessKeyID, secretAccessKey, region }) {
  const now = new Date()
  const amzDate = now.toISOString().replace(/[:-]|\.\d{3}/g, '')
  const dateStamp = amzDate.slice(0, 8)
  const service = 's3'
  const host = url.host
  const headers = {
    'cache-control': cacheControl,
    'content-type': contentType,
    host,
    'x-amz-content-sha256': bodyHash,
    'x-amz-date': amzDate,
  }
  const signedHeaders = Object.keys(headers).sort().join(';')
  const canonicalHeaders = Object.keys(headers)
    .sort()
    .map((key) => `${key}:${headers[key]}\n`)
    .join('')
  const canonicalRequest = [
    method,
    awsEncodePath(url.pathname),
    url.searchParams.toString(),
    canonicalHeaders,
    signedHeaders,
    bodyHash,
  ].join('\n')
  const credentialScope = `${dateStamp}/${region}/${service}/aws4_request`
  const stringToSign = [
    'AWS4-HMAC-SHA256',
    amzDate,
    credentialScope,
    sha256Hex(canonicalRequest),
  ].join('\n')
  const signingKey = hmac(`AWS4${secretAccessKey}`, dateStamp)
  const regionKey = hmac(signingKey, region)
  const serviceKey = hmac(regionKey, service)
  const requestKey = hmac(serviceKey, 'aws4_request')
  const signature = hmacHex(requestKey, stringToSign)
  return {
    'Cache-Control': cacheControl,
    'Content-Type': contentType,
    Host: host,
    'X-Amz-Content-Sha256': bodyHash,
    'X-Amz-Date': amzDate,
    Authorization: `AWS4-HMAC-SHA256 Credential=${accessKeyID}/${credentialScope}, SignedHeaders=${signedHeaders}, Signature=${signature}`,
  }
}

function awsEncodePath(pathname) {
  return pathname
    .split('/')
    .map((segment) => encodeURIComponent(decodeURIComponent(segment)).replace(/[!'()*]/g, (char) => `%${char.charCodeAt(0).toString(16).toUpperCase()}`))
    .join('/')
}

function sha256Hex(value) {
  return createHash('sha256').update(value).digest('hex')
}

function hmac(key, value) {
  return createHmac('sha256', key).update(value).digest()
}

function hmacHex(key, value) {
  return createHmac('sha256', key).update(value).digest('hex')
}

function completionCostForTask(task, artifactCount) {
  const mapCost = lookupCompletionUnitCost(task)
  if (mapCost !== null) {
    return mapCost * Math.max(1, artifactCount)
  }
  if (config.completionCost > 0) {
    return config.completionCost
  }
  const taskEstimate = Number.parseFloat(task.cost_estimate)
  if (Number.isFinite(taskEstimate) && taskEstimate >= 0) {
    return taskEstimate
  }
  return 0
}

function completionCostSource(task) {
  if (lookupCompletionUnitCost(task) !== null) return 'map'
  if (config.completionCost > 0) return 'flat'
  const taskEstimate = Number.parseFloat(task.cost_estimate)
  if (Number.isFinite(taskEstimate) && taskEstimate >= 0) return 'task_estimate'
  return 'none'
}

function lookupCompletionUnitCost(task) {
  const provider = task.provider || 'openai'
  const model = task.model || 'gpt-image-2'
  const size = task.size || '1024x1024'
  const quality = task.quality || 'standard'
  const keys = [
    `${provider}:${model}:${size}:${quality}`,
    `${model}:${size}:${quality}`,
    `${model}:${quality}`,
    model,
    'default',
  ]
  for (const key of keys) {
    const value = Number.parseFloat(config.completionCostMap[key])
    if (Number.isFinite(value) && value >= 0) {
      return value
    }
  }
  return null
}

async function completeTask(task, upstreamBody, artifacts) {
  const cost = completionCostForTask(task, artifacts.length)
  await apiFetch(`/image-workspace/worker/tasks/${task.id}/complete`, {
    method: 'POST',
    body: JSON.stringify({
      artifacts,
      result_json: JSON.stringify({
        provider: task.provider || 'openai',
        model: task.model,
        upstream_created: upstreamBody.created || null,
        artifact_count: artifacts.length,
        cost,
        cost_source: completionCostSource(task),
      }),
      cost,
    }),
  })
}

function failureResultJSON(task, error) {
  const message = error instanceof Error ? error.message : String(error)
  const failure = {
    message,
    kind: error?.upstreamDiagnostics ? 'upstream_response' : 'worker_error',
    captured_at: new Date().toISOString(),
  }
  if (error?.upstreamDiagnostics) {
    failure.upstream_response = error.upstreamDiagnostics
  }
  return JSON.stringify({
    provider: task.provider || 'openai',
    model: task.model || '',
    failure,
  })
}

async function failTask(task, error) {
  const message = error instanceof Error ? error.message : String(error)
  await apiFetch(`/image-workspace/worker/tasks/${task.id}/fail`, {
    method: 'POST',
    body: JSON.stringify({
      message,
      result_json: failureResultJSON(task, error),
    }),
  })
}

function isRetryableError(error) {
  if (!(error instanceof Error)) return true
  const msg = error.message || ''
  const cause = error.cause
  // Network-level errors: fetch failed, ECONNREFUSED, ECONNRESET, ETIMEDOUT, ENOTFOUND, etc.
  if (msg === 'fetch failed') return true
  if (cause && cause.code === 'ECONNREFUSED') return true
  if (cause && cause.code === 'ECONNRESET') return true
  if (cause && cause.code === 'ETIMEDOUT') return true
  if (cause && cause.code === 'ENOTFOUND') return true
  if (cause && cause.code === 'EAI_AGAIN') return true
  if (msg.includes('ECONNREFUSED')) return true
  if (msg.includes('ECONNRESET')) return true
  if (msg.includes('ETIMEDOUT')) return true
  if (msg.includes('ENOTFOUND')) return true
  if (msg.includes('fetch failed')) return true
  // Upstream 5xx or 429 are transient
  if (/upstream [5]\d{2}/.test(msg)) return true
  if (msg.includes('429')) return true
  // Generation timeout — upstream may be slow but recoverable
  if (msg === 'generation timeout') return true
  // Auth/config errors are NOT retryable
  if (msg.includes('IMAGE_WORKSPACE_UPSTREAM_API_KEY')) return false
  if (msg.includes('鉴权失败')) return false
  if (/upstream 40[134]/.test(msg)) return false
  // Content safety rejections are NOT retryable
  if (msg.includes('安全检查')) return false
  // Upstream 404 (wrong endpoint) is NOT retryable
  if (msg.includes('upstream 404') || msg.includes('上游返回 404')) return false
  // Default: retry on unknown errors (safer for transient issues)
  return true
}

async function processTask(task) {
  const maxAttempts = 1 + Math.max(0, config.generationRetries)
  let lastError
  for (let attempt = 1; attempt <= maxAttempts; attempt++) {
    try {
      const { body, items } = await generateImages(task)
      const artifacts = await writeArtifacts(task, items)
      await completeTask(task, body, artifacts)
      console.log(`Completed image workspace task ${task.id} with ${artifacts.length} artifact(s)`)
      return
    } catch (error) {
      lastError = error
      const retryable = isRetryableError(error)
      if (!retryable || attempt === maxAttempts) {
        console.error(`Failed image workspace task ${task.id} (attempt ${attempt}/${maxAttempts}${retryable ? ', retries exhausted' : ', non-retryable'}):`, error)
        break
      }
      const delay = config.generationRetryBaseMs * 2 ** (attempt - 1)
      console.warn(`Retryable error on image workspace task ${task.id} (attempt ${attempt}/${maxAttempts}), retrying in ${delay}ms:`, error.message || error)
      await sleep(delay)
    }
  }
  console.error(`Failed image workspace task ${task.id} after ${maxAttempts} attempt(s):`, lastError)
  await failTask(task, lastError)
}

async function loop() {
  let backoff = config.intervalMs
  for (;;) {
    try {
      await refreshRuntimeConfigIfDue()
      const task = await claimTask()
      if (task) {
        backoff = config.intervalMs
        await processTask(task)
      } else {
        await sleep(backoff)
      }
    } catch (error) {
      console.error('Worker loop error:', error)
      await sleep(backoff)
      backoff = Math.min(backoff * 2, config.maxBackoffMs)
    }
  }
}

async function runOnce() {
  const task = await claimTask()
  if (!task) {
    console.log('No queued image workspace task')
    return
  }
  await processTask(task)
}

async function runHealthcheck() {
  await apiFetch('/image-workspace/worker/health', { method: 'GET' })
  console.log(JSON.stringify({
    ok: true,
    api_base_url: config.baseURL,
    object_storage_enabled: config.objectStorage.enabled,
    upstream_configured: Boolean(config.upstreamAPIKey),
  }))
}

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

async function runStorageCheck() {
  if (!config.objectStorage.enabled) {
    console.log('object storage disabled')
    return
  }
  const storage = resolveObjectStorageConfig()
  const url = new URL(`${storage.endpoint.replace(/\/+$/, '')}/${storage.bucket}/${posix.join(storage.keyPrefix, 'healthcheck.txt')}`)
  const body = Buffer.from('image-workspace-storage-check')
  const headers = signedS3Headers({
    method: 'PUT',
    url,
    bodyHash: sha256Hex(body),
    contentType: 'text/plain',
    cacheControl: storage.cacheControl,
    accessKeyID: storage.accessKeyID,
    secretAccessKey: storage.secretAccessKey,
    region: storage.region,
  })
  const required = ['Authorization', 'X-Amz-Date', 'X-Amz-Content-Sha256', 'Content-Type', 'Cache-Control']
  for (const key of required) {
    if (!headers[key]) {
      throw new Error(`storage check missing signed header: ${key}`)
    }
  }
  console.log(JSON.stringify({
    ok: true,
    provider: storage.provider,
    bucket: storage.bucket,
    endpoint_host: url.host,
    key_prefix: storage.keyPrefix,
    public_base_url_configured: Boolean(storage.publicBaseURL),
  }))
}

async function runUpstreamCheck() {
  const task = {
    id: 1,
    user_id: 1,
    prompt: env('IMAGE_WORKSPACE_UPSTREAM_CHECK_PROMPT', 'Image workspace upstream check prompt'),
    negative_prompt: env('IMAGE_WORKSPACE_UPSTREAM_CHECK_NEGATIVE_PROMPT', 'watermark'),
    style: env('IMAGE_WORKSPACE_UPSTREAM_CHECK_STYLE', 'minimal geometric icon'),
    provider: env('IMAGE_WORKSPACE_UPSTREAM_CHECK_PROVIDER', 'openai'),
    model: env('IMAGE_WORKSPACE_UPSTREAM_CHECK_MODEL', 'gpt-image-2'),
    size: env('IMAGE_WORKSPACE_UPSTREAM_CHECK_SIZE', '1024x1024'),
    quality: env('IMAGE_WORKSPACE_UPSTREAM_CHECK_QUALITY', 'standard'),
    batch_size: intEnv('IMAGE_WORKSPACE_UPSTREAM_CHECK_BATCH_SIZE', 1),
    cost_estimate: floatEnv('IMAGE_WORKSPACE_UPSTREAM_CHECK_COST_ESTIMATE', 0),
  }
  const { body, items } = await generateImages(task)
  const artifacts = await writeArtifacts(task, items)
  console.log(JSON.stringify({
    ok: true,
    upstream_url: config.upstreamURL,
    upstream_created: body.created || null,
    artifact_count: artifacts.length,
    artifacts,
  }, null, 2))
}

async function main() {
  await refreshRuntimeConfigIfDue(true, { required: true })
  if (process.argv.includes('--storage-check')) {
    await runStorageCheck()
    return
  }
  if (process.argv.includes('--upstream-check')) {
    await runUpstreamCheck()
    return
  }
  if (process.argv.includes('--healthcheck')) {
    await runHealthcheck()
    return
  }
  if (process.argv.includes('--once')) {
    await runOnce()
    return
  }
  await loop()
}

void main()
