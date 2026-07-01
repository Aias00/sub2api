#!/usr/bin/env node
import { createHmac, createHash } from 'node:crypto'
import { createServer } from 'node:http'
import { execFileSync, spawn } from 'node:child_process'
import { existsSync, rmSync } from 'node:fs'
import { resolve } from 'node:path'

const realProviderMode = boolEnv('IMAGE_WORKSPACE_E2E_REAL_PROVIDER', false)
let smokeUserForCleanup = null
let localArtifactPathForCleanup = ''
let runtimeSettingsSnapshotForCleanup = null

const config = {
  baseURL: env('BASE_URL', 'http://127.0.0.1:8080').replace(/\/$/, ''),
  appContainer: env('APPDOCKER_CONTAINER', env('APP_CONTAINER', 'sub2api')),
  pgContainer: env('PGDOCKER_CONTAINER', 'sub2api-postgres'),
  pgUser: env('PGUSER', 'sub2api'),
  pgDatabase: env('PGDATABASE', 'sub2api'),
  smokeEmail: env('IMAGE_WORKSPACE_E2E_EMAIL', `iw-e2e-${Date.now()}@example.test`),
  smokePasswordHash: env(
    'IMAGE_WORKSPACE_E2E_PASSWORD_HASH',
    '$2a$10$gygTtEYFei2.4wUCc2Nsh.NDDW2mD7gLhsF2PAANKH.Q9Ba45/yAi',
  ),
  initialBalance: Number.parseFloat(env('IMAGE_WORKSPACE_E2E_INITIAL_BALANCE', '50')),
  completionCost: Number.parseFloat(env('IMAGE_WORKSPACE_E2E_COMPLETION_COST', '0.25')),
  realProvider: realProviderMode,
  upstreamURL: env('IMAGE_WORKSPACE_E2E_UPSTREAM_URL', 'https://api.openai.com/v1/images/generations'),
  upstreamAPIKey: realProviderMode
    ? env('IMAGE_WORKSPACE_UPSTREAM_API_KEY', '')
    : 'mock-image-workspace-e2e-key',
  useObjectStorage: boolEnv('IMAGE_WORKSPACE_E2E_OBJECT_STORAGE', false),
  objectStorageEndpoint: env('IMAGE_WORKSPACE_OBJECT_STORAGE_ENDPOINT', env('IMAGE_WORKSPACE_R2_ENDPOINT', '')),
  objectStorageAccessKeyID: realProviderMode
    ? env('IMAGE_WORKSPACE_OBJECT_STORAGE_ACCESS_KEY_ID', env('IMAGE_WORKSPACE_R2_ACCESS_KEY_ID', env('IMAGE_WORKSPACE_R2_ACCESS_KEY', '')))
    : 'mock-r2-access-key',
  objectStorageSecretAccessKey: realProviderMode
    ? env('IMAGE_WORKSPACE_OBJECT_STORAGE_SECRET_ACCESS_KEY', env('IMAGE_WORKSPACE_R2_SECRET_ACCESS_KEY', env('IMAGE_WORKSPACE_R2_SECRET_KEY', '')))
    : 'mock-r2-secret-key',
  objectStorageBucket: env('IMAGE_WORKSPACE_E2E_OBJECT_STORAGE_BUCKET', 'sub2api-image-workspace'),
  objectStoragePrefix: env('IMAGE_WORKSPACE_E2E_OBJECT_STORAGE_PREFIX', 'image-workspace-e2e'),
  objectStoragePublicBaseURL: env('IMAGE_WORKSPACE_E2E_OBJECT_STORAGE_PUBLIC_BASE_URL', 'https://assets.example.test/image-workspace'),
  cleanup: boolEnv('IMAGE_WORKSPACE_E2E_CLEANUP', !realProviderMode),
}

function env(name, fallback = '') {
  const value = process.env[name]
  return value === undefined || value === '' ? fallback : value
}

function boolEnv(name, fallback) {
  const value = process.env[name]
  if (value === undefined || value === '') return fallback
  return ['1', 'true', 'yes', 'on'].includes(value.toLowerCase())
}

function assert(condition, message) {
  if (!condition) {
    throw new Error(message)
  }
}

function sh(cmd, args, options = {}) {
  return execFileSync(cmd, args, {
    encoding: 'utf8',
    stdio: ['pipe', 'pipe', 'pipe'],
    ...options,
  }).trim()
}

function dockerExec(container, command, options = {}) {
  return sh('docker', ['exec', container, 'sh', '-lc', command], options)
}

function psql(sql) {
  return sh('docker', [
    'exec',
    '-i',
    config.pgContainer,
    'psql',
    '-U',
    config.pgUser,
    '-d',
    config.pgDatabase,
    '-v',
    'ON_ERROR_STOP=1',
    '-At',
    '-F',
    '\t',
  ], { input: sql })
}

const runtimeSettingKeys = [
  'image_workspace_upstream_url',
  'image_workspace_generation_timeout_ms',
  'image_workspace_completion_cost',
  'image_workspace_completion_cost_map_json',
  'image_workspace_object_storage_enabled',
  'image_workspace_object_storage_provider',
  'image_workspace_object_storage_bucket',
  'image_workspace_object_storage_region',
  'image_workspace_object_storage_prefix',
  'image_workspace_object_storage_public_base_url',
]

function sqlString(value) {
  return `'${String(value).replaceAll("'", "''")}'`
}

function snapshotRuntimeSettings() {
  const rows = psql(`
SELECT key, value
FROM settings
WHERE key IN (${runtimeSettingKeys.map(sqlString).join(', ')})
ORDER BY key;
`)
  const snapshot = new Map()
  for (const row of rows.split('\n').filter(Boolean)) {
    const [key, value] = parseTabRow(row)
    snapshot.set(key, value)
  }
  return snapshot
}

function restoreRuntimeSettings(snapshot) {
  if (!(snapshot instanceof Map)) return
  psql(`
DELETE FROM settings
WHERE key IN (${runtimeSettingKeys.map(sqlString).join(', ')});
${snapshot.size > 0 ? `INSERT INTO settings (key, value, updated_at)
VALUES ${Array.from(snapshot.entries()).map(([key, value]) => `(${sqlString(key)}, ${sqlString(value)}, now())`).join(',\n')}
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = now();` : ''}
`)
}

function setImageWorkspaceRuntimeSettings({ upstreamURL, objectStorageEndpoint = '' }) {
  const values = {
    image_workspace_upstream_url: upstreamURL,
    image_workspace_generation_timeout_ms: '120000',
    image_workspace_completion_cost: String(config.completionCost),
    image_workspace_completion_cost_map_json: '{}',
    image_workspace_object_storage_enabled: config.useObjectStorage ? 'true' : 'false',
    image_workspace_object_storage_provider: 'r2',
    image_workspace_object_storage_bucket: config.useObjectStorage ? config.objectStorageBucket : '',
    image_workspace_object_storage_region: 'auto',
    image_workspace_object_storage_prefix: config.objectStoragePrefix,
    image_workspace_object_storage_public_base_url: config.useObjectStorage ? config.objectStoragePublicBaseURL : '',
  }
  psql(`
INSERT INTO settings (key, value, updated_at)
VALUES ${Object.entries(values).map(([key, value]) => `(${sqlString(key)}, ${sqlString(value)}, now())`).join(',\n')}
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = now();
`)
  if (config.useObjectStorage) {
    assert(objectStorageEndpoint || config.objectStorageEndpoint, 'object storage endpoint must be available before worker run')
  }
}

function parseTabRow(row) {
  return row.split('\t')
}

function parseJSON(value, label) {
  try {
    return JSON.parse(String(value || '{}'))
  } catch (error) {
    throw new Error(`${label} is not valid JSON: ${error instanceof Error ? error.message : String(error)}`)
  }
}

function joinPublicURL(baseURL, key) {
  return `${String(baseURL || '').replace(/\/+$/, '')}/${String(key || '').replace(/^\/+/, '')}`
}

function readJWTSecret() {
  return dockerExec(
    config.appContainer,
    "awk '/^jwt:/{flag=1;next} flag && /secret:/{print $2; exit}' /app/data/config.yaml",
  )
}

function readAppDataMount() {
  const raw = sh('docker', ['inspect', config.appContainer])
  const containers = JSON.parse(raw)
  const mount = containers?.[0]?.Mounts?.find((item) => item.Destination === '/app/data')
  assert(mount?.Source, `container ${config.appContainer} must mount /app/data`)
  return mount.Source
}

function tokenVersionFor(email, passwordHash) {
  const hash = createHash('sha256')
    .update(`${email.trim().toLowerCase()}\n${passwordHash}`)
    .digest()
    .subarray(0, 8)
  return hash.readBigUInt64BE() & 0x7fffffffffffffffn
}

function jwtForUser({ id, email, role, passwordHash }, secret) {
  const now = BigInt(Math.floor(Date.now() / 1000))
  const headerJSON = '{"alg":"HS256","typ":"JWT"}'
  const payloadJSON = [
    '{',
    `"user_id":${id},`,
    `"email":${JSON.stringify(email)},`,
    `"role":${JSON.stringify(role)},`,
    `"token_version":${tokenVersionFor(email, passwordHash).toString()},`,
    `"exp":${(now + 3600n).toString()},`,
    `"iat":${now.toString()},`,
    `"nbf":${now.toString()}`,
    '}',
  ].join('')
  const b64 = (value) => Buffer.from(value).toString('base64url')
  const body = `${b64(headerJSON)}.${b64(payloadJSON)}`
  const sig = createHmac('sha256', secret).update(body).digest('base64url')
  return `${body}.${sig}`
}

async function api(path, { method = 'GET', token = '', body } = {}) {
  const headers = { accept: 'application/json' }
  if (token) headers.authorization = `Bearer ${token}`
  if (body !== undefined) headers['content-type'] = 'application/json'
  const response = await fetch(`${config.baseURL}/api/v1${path}`, {
    method,
    headers,
    body: body === undefined ? undefined : JSON.stringify(body),
  })
  const text = await response.text()
  let payload
  try {
    payload = text ? JSON.parse(text) : {}
  } catch {
    payload = { raw: text }
  }
  if (!response.ok || (typeof payload.code === 'number' && payload.code !== 0)) {
    throw new Error(`${method} ${path} failed: HTTP ${response.status} ${text}`)
  }
  return payload.data
}

async function assertPublicImageURLReachable(url) {
  const head = await fetch(url, { method: 'HEAD' })
  if (head.ok) return
  const range = await fetch(url, {
    headers: { range: 'bytes=0-0' },
  })
  assert(range.ok, `public artifact URL failed: HEAD ${head.status}, GET ${range.status}`)
}

function createSmokeUser() {
  const row = psql(`
INSERT INTO users (email, username, password_hash, role, balance, concurrency, status, signup_source, created_at, updated_at)
VALUES (${sqlString(config.smokeEmail)}, 'Image Workspace E2E', ${sqlString(config.smokePasswordHash)}, 'user', ${config.initialBalance}, 5, 'active', 'email', now(), now())
ON CONFLICT (email) WHERE deleted_at IS NULL AND signup_source <> 'touch'
DO UPDATE SET password_hash = EXCLUDED.password_hash, balance = EXCLUDED.balance, status = 'active', updated_at = now()
RETURNING id, email, role, password_hash, balance;
`)
  const [id, email, role, passwordHash, balance] = parseTabRow(row)
  return { id: Number(id), email, role, passwordHash, balance: Number.parseFloat(balance) }
}

function cleanupSmokeUser(user) {
  if (!config.cleanup || !user?.id || !user?.email) return
  psql(`
DELETE FROM users
WHERE id = ${user.id}
  AND email = ${sqlString(user.email)}
  AND email LIKE 'iw-e2e-%@example.test';
`)
}

function cleanupAllSmokeUsers() {
  if (!config.cleanup) return
  psql(`
DELETE FROM users
WHERE email LIKE 'iw-e2e-%@example.test';
`)
}

function startMockUpstream() {
  const png1x1 = 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/p9sAAAAASUVORK5CYII='
  const requests = []
  const server = createServer(async (req, res) => {
    if (req.method !== 'POST' || req.url !== '/v1/images/generations') {
      res.writeHead(404, { 'content-type': 'application/json' })
      res.end(JSON.stringify({ error: { message: 'not found' } }))
      return
    }
    const chunks = []
    for await (const chunk of req) chunks.push(chunk)
    const body = JSON.parse(Buffer.concat(chunks).toString('utf8') || '{}')
    requests.push({ authorization: req.headers.authorization || '', body })
    if (req.headers.authorization !== `Bearer ${config.upstreamAPIKey}`) {
      res.writeHead(401, { 'content-type': 'application/json' })
      res.end(JSON.stringify({ error: { message: 'invalid token' } }))
      return
    }
    res.writeHead(200, { 'content-type': 'application/json' })
    res.end(JSON.stringify({
      created: 1893456000,
      data: [{
        b64_json: png1x1,
        output_format: 'png',
        revised_prompt: `mock revised: ${body.prompt || ''}`,
      }],
    }))
  })
  return new Promise((resolveListen, reject) => {
    server.on('error', reject)
    server.listen(0, '127.0.0.1', () => resolveListen({ server, requests }))
  })
}

function startFailingMockUpstream() {
  const requests = []
  const server = createServer(async (req, res) => {
    if (req.method !== 'POST' || req.url !== '/v1/images/generations') {
      res.writeHead(404, { 'content-type': 'application/json' })
      res.end(JSON.stringify({ error: { message: 'not found' } }))
      return
    }
    const chunks = []
    for await (const chunk of req) chunks.push(chunk)
    const body = JSON.parse(Buffer.concat(chunks).toString('utf8') || '{}')
    requests.push({ authorization: req.headers.authorization || '', body })
    res.writeHead(500, { 'content-type': 'application/json' })
    res.end(JSON.stringify({ error: { message: 'mock upstream failure for refund verification' } }))
  })
  return new Promise((resolveListen, reject) => {
    server.on('error', reject)
    server.listen(0, '127.0.0.1', () => resolveListen({ server, requests }))
  })
}

function startMockObjectStorage() {
  const uploads = []
  const server = createServer(async (req, res) => {
    if (req.method !== 'PUT') {
      res.writeHead(405, { 'content-type': 'application/json' })
      res.end(JSON.stringify({ error: 'method not allowed' }))
      return
    }
    const chunks = []
    for await (const chunk of req) chunks.push(chunk)
    const bytes = Buffer.concat(chunks)
    const checksum = createHash('sha256').update(bytes).digest('hex')
    uploads.push({
      url: req.url,
      headers: req.headers,
      bytes,
      checksum,
    })
    if (!String(req.headers.authorization || '').startsWith(`AWS4-HMAC-SHA256 Credential=${config.objectStorageAccessKeyID}/`)) {
      res.writeHead(401, { 'content-type': 'application/json' })
      res.end(JSON.stringify({ error: 'missing signed authorization' }))
      return
    }
    if (req.headers['x-amz-content-sha256'] !== checksum) {
      res.writeHead(400, { 'content-type': 'application/json' })
      res.end(JSON.stringify({ error: 'payload hash mismatch' }))
      return
    }
    res.writeHead(200, { etag: `"${checksum}"` })
    res.end('')
  })
  return new Promise((resolveListen, reject) => {
    server.on('error', reject)
    server.listen(0, '127.0.0.1', () => resolveListen({ server, uploads }))
  })
}

function runWorker({ appDataMount, objectStorageEndpoint = '' }) {
  return new Promise((resolveRun, reject) => {
    const child = spawn(process.execPath, [resolve('tools/image-workspace-worker/src/worker.mjs'), '--once'], {
      cwd: process.cwd(),
      env: {
        ...process.env,
        IMAGE_WORKSPACE_API_BASE_URL: config.baseURL,
        IMAGE_WORKSPACE_UPSTREAM_API_KEY: config.upstreamAPIKey,
        IMAGE_WORKSPACE_OUTPUT_DIR: `${appDataMount}/image-workspace`,
        IMAGE_WORKSPACE_STORAGE_KEY_ROOT: '/app/data/image-workspace',
        IMAGE_WORKSPACE_OBJECT_STORAGE_ENDPOINT: objectStorageEndpoint,
        IMAGE_WORKSPACE_OBJECT_STORAGE_ACCESS_KEY_ID: config.objectStorageAccessKeyID,
        IMAGE_WORKSPACE_OBJECT_STORAGE_SECRET_ACCESS_KEY: config.objectStorageSecretAccessKey,
      },
      stdio: ['ignore', 'pipe', 'pipe'],
    })
    let stdout = ''
    let stderr = ''
    child.stdout.on('data', (chunk) => { stdout += chunk.toString('utf8') })
    child.stderr.on('data', (chunk) => { stderr += chunk.toString('utf8') })
    child.on('error', reject)
    child.on('close', (code) => {
      if (code !== 0) {
        reject(new Error(`worker exited ${code}\n${stderr}`))
        return
      }
      resolveRun({ stdout, stderr })
    })
  })
}

async function main() {
  assert(Number.isFinite(config.initialBalance) && config.initialBalance > 0, 'initial balance must be positive')
  assert(Number.isFinite(config.completionCost) && config.completionCost > 0, 'completion cost must be positive')
  if (config.realProvider) {
    assert(config.upstreamAPIKey, 'IMAGE_WORKSPACE_UPSTREAM_API_KEY is required when IMAGE_WORKSPACE_E2E_REAL_PROVIDER=1')
    assert(config.useObjectStorage, 'IMAGE_WORKSPACE_E2E_OBJECT_STORAGE=true is required when IMAGE_WORKSPACE_E2E_REAL_PROVIDER=1')
    assert(config.objectStorageEndpoint, 'IMAGE_WORKSPACE_OBJECT_STORAGE_ENDPOINT or IMAGE_WORKSPACE_R2_ENDPOINT is required when IMAGE_WORKSPACE_E2E_REAL_PROVIDER=1')
    assert(config.objectStorageAccessKeyID, 'IMAGE_WORKSPACE_OBJECT_STORAGE_ACCESS_KEY_ID or R2 alias is required when IMAGE_WORKSPACE_E2E_REAL_PROVIDER=1')
    assert(config.objectStorageSecretAccessKey, 'IMAGE_WORKSPACE_OBJECT_STORAGE_SECRET_ACCESS_KEY or R2 alias is required when IMAGE_WORKSPACE_E2E_REAL_PROVIDER=1')
    assert(config.objectStorageBucket, 'IMAGE_WORKSPACE_E2E_OBJECT_STORAGE_BUCKET is required when IMAGE_WORKSPACE_E2E_REAL_PROVIDER=1')
    assert(config.objectStoragePublicBaseURL, 'IMAGE_WORKSPACE_E2E_OBJECT_STORAGE_PUBLIC_BASE_URL is required when IMAGE_WORKSPACE_E2E_REAL_PROVIDER=1')
  }

  const runtimeSettingsSnapshot = snapshotRuntimeSettings()
  runtimeSettingsSnapshotForCleanup = runtimeSettingsSnapshot
  cleanupAllSmokeUsers()
  const user = createSmokeUser()
  smokeUserForCleanup = user
  const token = jwtForUser(user, readJWTSecret())
  const appDataMount = readAppDataMount()

  const models = await api('/image-workspace/models', { token })
  assert(Array.isArray(models.models) && models.models.length > 0, 'models API returned no models')
  const selectedModel = models.models.find((item) => item.id === 'gpt-image-2')
  assert(selectedModel, 'models API does not include gpt-image-2')
  assert(selectedModel.enabled !== false, 'gpt-image-2 is disabled')
  assert(selectedModel.provider === 'openai', `unexpected gpt-image-2 provider ${selectedModel.provider}`)
  assert(Array.isArray(selectedModel.sizes) && selectedModel.sizes.includes('1024x1024'), 'gpt-image-2 sizes missing 1024x1024')
  assert(Array.isArray(selectedModel.qualities) && selectedModel.qualities.includes('standard'), 'gpt-image-2 qualities missing standard')

  const template = await api('/image-workspace/templates', {
    method: 'POST',
    token,
    body: {
      title: `E2E template ${Date.now()}`,
      description: 'Local Image Workspace E2E template',
      prompt: 'A polished AI image generation workstation',
      negative_prompt: 'watermark, blurry',
      model: 'gpt-image-2',
      size: '1024x1024',
      quality: 'standard',
      style: 'clean editorial',
      is_default: true,
    },
  })
  assert(template.id > 0, 'template was not created')
  assert(template.model === 'gpt-image-2', `template model mismatch: ${template.model}`)
  assert(template.size === '1024x1024', `template size mismatch: ${template.size}`)
  assert(template.quality === 'standard', `template quality mismatch: ${template.quality}`)
  assert(template.negative_prompt === 'watermark, blurry', 'template negative prompt was not persisted')
  assert(template.is_default === true, 'template default flag was not persisted')

  const templates = await api('/image-workspace/templates', { token })
  const listedTemplate = templates.items?.find((item) => item.id === template.id)
  assert(listedTemplate, 'template list does not include created template')
  assert(listedTemplate.prompt === template.prompt, 'template list prompt mismatch')
  assert(listedTemplate.is_default === true, 'template list default flag mismatch')

  const task = await api('/image-workspace/tasks', {
    method: 'POST',
    token,
    body: {
      prompt: 'A production-ready AI capability dashboard with artifact previews and a cost ledger',
      negative_prompt: 'watermark, blurry, text artifacts',
      model: 'gpt-image-2',
      provider: 'openai',
      size: '1024x1024',
      quality: 'standard',
      style: 'clean editorial',
      batch_size: 1,
      template_id: template.id,
    },
  })
  assert(task.id > 0 && task.status === 'queued', 'task was not queued')
  assert(task.model === 'gpt-image-2', `queued task model mismatch: ${task.model}`)
  assert(task.provider === 'openai', `queued task provider mismatch: ${task.provider}`)
  assert(task.size === '1024x1024', `queued task size mismatch: ${task.size}`)
  assert(task.quality === 'standard', `queued task quality mismatch: ${task.quality}`)
  assert(task.template_id === template.id, `queued task template mismatch: ${task.template_id}`)

  let upstream
  let objectStorage
  let localArtifactPath = ''
  try {
    if (!config.realProvider) {
      upstream = await startMockUpstream()
    }
    if (config.useObjectStorage && !config.realProvider) {
      objectStorage = await startMockObjectStorage()
    }
    const address = upstream?.server?.address()
    const upstreamURL = config.realProvider ? config.upstreamURL : `http://127.0.0.1:${address.port}/v1/images/generations`
    const storageAddress = objectStorage?.server?.address()
    const objectStorageEndpoint = config.realProvider ? config.objectStorageEndpoint : (storageAddress ? `http://127.0.0.1:${storageAddress.port}` : '')
    setImageWorkspaceRuntimeSettings({ upstreamURL, objectStorageEndpoint })
    const run = await runWorker({ appDataMount, objectStorageEndpoint })
    assert(run.stdout.includes(`Completed image workspace task ${task.id}`), `worker did not complete task ${task.id}: ${run.stdout}${run.stderr}`)
    if (!config.realProvider) {
      assert(upstream.requests.length === 1, `expected one upstream request, got ${upstream.requests.length}`)
      assert(upstream.requests[0].body.model === 'gpt-image-2', 'worker sent unexpected model')
      assert(String(upstream.requests[0].body.prompt).includes('Negative prompt:'), 'worker did not include negative prompt')
    }
    if (config.useObjectStorage && !config.realProvider) {
      assert(objectStorage.uploads.length === 1, `expected one object upload, got ${objectStorage.uploads.length}`)
      const upload = objectStorage.uploads[0]
      assert(upload.url === `/${config.objectStorageBucket}/${config.objectStoragePrefix}/${user.id}/${task.id}/image-1.png`, `unexpected object path ${upload.url}`)
      assert(upload.headers['content-type'] === 'image/png', 'object upload content-type mismatch')
      assert(upload.bytes.length > 0, 'object upload was empty')
    }
  } finally {
    if (upstream?.server) {
      await new Promise((resolveClose) => upstream.server.close(resolveClose))
    }
    if (objectStorage?.server) {
      await new Promise((resolveClose) => objectStorage.server.close(resolveClose))
    }
  }

  const detail = await api(`/image-workspace/tasks/${task.id}`, { token })
  assert(detail.status === 'succeeded', `task status is ${detail.status}`)
  assert(detail.model === 'gpt-image-2', `task detail model mismatch: ${detail.model}`)
  assert(detail.provider === 'openai', `task detail provider mismatch: ${detail.provider}`)
  assert(detail.size === '1024x1024', `task detail size mismatch: ${detail.size}`)
  assert(detail.quality === 'standard', `task detail quality mismatch: ${detail.quality}`)
  assert(detail.style === 'clean editorial', `task detail style mismatch: ${detail.style}`)
  assert(detail.template_id === template.id, `task detail template mismatch: ${detail.template_id}`)
  assert(detail.cost_estimate === config.completionCost, `expected cost ${config.completionCost}, got ${detail.cost_estimate}`)
  assert(detail.balance_snapshot === config.initialBalance - config.completionCost, `unexpected balance snapshot ${detail.balance_snapshot}`)
  const result = parseJSON(detail.result_json, 'task result_json')
  assert(result.provider === 'openai', `result provider mismatch: ${result.provider}`)
  assert(result.model === 'gpt-image-2', `result model mismatch: ${result.model}`)
  assert(Number(result.artifact_count || 0) === 1, `result artifact_count mismatch: ${result.artifact_count}`)
  assert(result.cost === config.completionCost, `result cost mismatch: ${result.cost}`)
  assert(['flat', 'map', 'task_estimate'].includes(String(result.cost_source || '')), `unexpected result cost_source: ${result.cost_source}`)
  assert(Array.isArray(detail.artifacts) && detail.artifacts.length === 1, 'task detail missing artifact')
  const artifact = detail.artifacts[0]
  const expectedProvider = config.useObjectStorage ? 'r2' : 'local'
  assert(artifact.storage_provider === expectedProvider, `unexpected storage provider ${artifact.storage_provider}`)
  assert(artifact.file_size > 0, 'artifact file size missing')
  assert(artifact.mime_type === 'image/png', `unexpected artifact mime type ${artifact.mime_type}`)
  assert(/^[a-f0-9]{64}$/.test(artifact.checksum || ''), 'artifact checksum missing or invalid')
  const artifactMetadata = parseJSON(artifact.metadata_json, 'artifact metadata_json')
  assert(artifactMetadata.model === 'gpt-image-2', `artifact metadata model mismatch: ${artifactMetadata.model}`)
  assert(String(artifactMetadata.revised_prompt || '').includes('mock revised:') || config.realProvider, 'artifact metadata missing revised prompt')

  if (config.useObjectStorage) {
    const expectedKey = `${config.objectStoragePrefix}/${user.id}/${task.id}/image-1.png`
    const expectedURL = joinPublicURL(config.objectStoragePublicBaseURL, expectedKey)
    assert(artifact.storage_key === expectedKey, `unexpected object storage key ${artifact.storage_key}`)
    assert(artifact.image_url === expectedURL, `unexpected object storage public URL ${artifact.image_url}`)
    if (config.realProvider) {
      await assertPublicImageURLReachable(artifact.image_url)
    }
  } else {
    const artifactPath = artifact.storage_key.replace('/app/data', appDataMount)
    localArtifactPath = artifactPath
    localArtifactPathForCleanup = artifactPath
    assert(existsSync(artifactPath), `artifact file does not exist: ${artifactPath}`)

    const download = await fetch(`${config.baseURL}/api/v1/image-workspace/artifacts/${artifact.id}/download`, {
      headers: { authorization: `Bearer ${token}` },
    })
    assert(download.ok, `artifact download failed: HTTP ${download.status}`)
    const bytes = Buffer.from(await download.arrayBuffer())
    assert(bytes.length === artifact.file_size, `downloaded ${bytes.length} bytes, expected ${artifact.file_size}`)
  }

  const list = await api('/image-workspace/tasks?page=1&page_size=20', { token })
  assert(list.total >= 1, 'task history is empty')
  const historyTask = list.items.find((item) => item.id === task.id)
  assert(historyTask && historyTask.status === 'succeeded', 'task history does not include succeeded e2e task')
  assert(historyTask.model === detail.model, 'task history model mismatch')
  assert(historyTask.template_id === template.id, 'task history template mismatch')

  const usage = await api('/image-workspace/usage-records?page=1&page_size=20', { token })
  const usageRecord = usage.items?.find((item) => item.task_id === task.id)
  assert(usageRecord, 'usage API does not include e2e task')
  assert(usageRecord.provider === 'openai', `usage provider mismatch: ${usageRecord.provider}`)
  assert(usageRecord.model === 'gpt-image-2', `usage model mismatch: ${usageRecord.model}`)
  assert(usageRecord.size === '1024x1024', `usage size mismatch: ${usageRecord.size}`)
  assert(usageRecord.quality === 'standard', `usage quality mismatch: ${usageRecord.quality}`)
  assert(usageRecord.image_count === 1, `usage image_count mismatch: ${usageRecord.image_count}`)
  assert(usageRecord.actual_cost === config.completionCost, `usage actual_cost mismatch: ${usageRecord.actual_cost}`)
  assert(usageRecord.balance_snapshot === config.initialBalance - config.completionCost, `usage balance snapshot mismatch: ${usageRecord.balance_snapshot}`)
  assert(usageRecord.billing_status === 'settled', `usage billing status mismatch: ${usageRecord.billing_status}`)
  const usageMetadata = parseJSON(usageRecord.metadata_json, 'usage metadata_json')
  assert(Number(usageMetadata.artifact_count || 0) === 1, `usage metadata artifact_count mismatch: ${usageMetadata.artifact_count}`)

  const dbRows = psql(`
SELECT
  (SELECT balance FROM users WHERE id = ${user.id})::text,
  (SELECT actual_cost FROM image_workspace_usage_records WHERE task_id = ${task.id})::text,
  (SELECT billing_status FROM image_workspace_usage_records WHERE task_id = ${task.id}),
  (SELECT COUNT(*) FROM image_workspace_artifacts WHERE task_id = ${task.id})::text;
`)
  const [balance, actualCost, billingStatus, artifactCount] = parseTabRow(dbRows)
  assert(Number.parseFloat(balance) === config.initialBalance - config.completionCost, `unexpected DB balance ${balance}`)
  assert(Number.parseFloat(actualCost) === config.completionCost, `unexpected DB actual cost ${actualCost}`)
  assert(billingStatus === 'settled', `unexpected billing status ${billingStatus}`)
  assert(Number(artifactCount) === 1, `unexpected artifact count ${artifactCount}`)

  let failedTask = null
  let refundedUsageRecord = null
  if (!config.realProvider) {
    const refundReserveCost = 0.5
    failedTask = await api('/image-workspace/tasks', {
      method: 'POST',
      token,
      body: {
        prompt: 'This task intentionally fails to prove failed image generation refunds the reservation',
        negative_prompt: 'watermark',
        model: 'gpt-image-2',
        provider: 'openai',
        size: '1024x1024',
        quality: 'standard',
        style: 'refund audit',
        batch_size: 1,
      },
    })
    assert(failedTask.id > 0 && failedTask.status === 'queued', 'failed-refund task was not queued')
    const reservedBalanceRow = psql(`
WITH debited AS (
  UPDATE users
  SET balance = balance - ${refundReserveCost},
      updated_at = now()
  WHERE id = ${user.id}
  RETURNING balance
)
UPDATE image_workspace_tasks
SET cost_estimate = ${refundReserveCost},
    balance_snapshot = (SELECT balance FROM debited),
    updated_at = now()
WHERE id = ${failedTask.id}
RETURNING balance_snapshot::text;
`)
    assert(Number.parseFloat(reservedBalanceRow) === config.initialBalance - config.completionCost - refundReserveCost, `unexpected reserved balance ${reservedBalanceRow}`)

    let failingUpstream
    try {
      failingUpstream = await startFailingMockUpstream()
      const failingAddress = failingUpstream.server.address()
      setImageWorkspaceRuntimeSettings({
        upstreamURL: `http://127.0.0.1:${failingAddress.port}/v1/images/generations`,
      })
      const run = await runWorker({
        appDataMount,
      })
      assert(
        `${run.stdout}\n${run.stderr}`.includes(`Failed image workspace task ${failedTask.id}`),
        `worker did not report failed task ${failedTask.id}: ${run.stdout}${run.stderr}`,
      )
      assert(failingUpstream.requests.length === 1, `expected one failing upstream request, got ${failingUpstream.requests.length}`)
    } finally {
      if (failingUpstream?.server) {
        await new Promise((resolveClose) => failingUpstream.server.close(resolveClose))
      }
    }

    const failedDetail = await api(`/image-workspace/tasks/${failedTask.id}`, { token })
    assert(failedDetail.status === 'failed', `failed-refund task status is ${failedDetail.status}`)
    assert(String(failedDetail.error_message || '').includes('mock upstream failure'), `failed task error_message mismatch: ${failedDetail.error_message || ''}`)
    assert(failedDetail.balance_snapshot === config.initialBalance - config.completionCost, `failed task refund balance snapshot mismatch: ${failedDetail.balance_snapshot}`)

    const refundedUsage = await api('/image-workspace/usage-records?page=1&page_size=20', { token })
    refundedUsageRecord = refundedUsage.items?.find((item) => item.task_id === failedTask.id)
    assert(refundedUsageRecord, 'usage API does not include refunded failed task')
    assert(refundedUsageRecord.billing_status === 'refunded', `refunded usage billing status mismatch: ${refundedUsageRecord.billing_status}`)
    assert(refundedUsageRecord.actual_cost === 0, `refunded usage actual_cost mismatch: ${refundedUsageRecord.actual_cost}`)
    assert(refundedUsageRecord.reserved_cost === refundReserveCost, `refunded usage reserved_cost mismatch: ${refundedUsageRecord.reserved_cost}`)
    assert(refundedUsageRecord.balance_snapshot === config.initialBalance - config.completionCost, `refunded usage balance snapshot mismatch: ${refundedUsageRecord.balance_snapshot}`)
    const refundedMetadata = parseJSON(refundedUsageRecord.metadata_json, 'refunded usage metadata_json')
    assert(refundedMetadata.status === 'failed', `refunded usage metadata status mismatch: ${refundedMetadata.status}`)

    const refundDBRows = psql(`
SELECT
  (SELECT balance FROM users WHERE id = ${user.id})::text,
  (SELECT actual_cost FROM image_workspace_usage_records WHERE task_id = ${failedTask.id})::text,
  (SELECT reserved_cost FROM image_workspace_usage_records WHERE task_id = ${failedTask.id})::text,
  (SELECT billing_status FROM image_workspace_usage_records WHERE task_id = ${failedTask.id});
`)
    const [refundBalance, refundActualCost, refundReservedCost, refundBillingStatus] = parseTabRow(refundDBRows)
    assert(Number.parseFloat(refundBalance) === config.initialBalance - config.completionCost, `unexpected refund DB balance ${refundBalance}`)
    assert(Number.parseFloat(refundActualCost) === 0, `unexpected refund DB actual cost ${refundActualCost}`)
    assert(Number.parseFloat(refundReservedCost) === refundReserveCost, `unexpected refund DB reserved cost ${refundReservedCost}`)
    assert(refundBillingStatus === 'refunded', `unexpected refund billing status ${refundBillingStatus}`)
  }

  console.log('# Image Workspace E2E')
  console.log('')
  console.log(`- Base URL: ${config.baseURL}`)
  console.log(`- Provider mode: ${config.realProvider ? 'real' : 'mock'}`)
  console.log(`- Smoke user: ${user.email} (${user.id})`)
  console.log(`- Models: ${models.models.length}`)
  console.log(`- Template ID: ${template.id}`)
  console.log(`- Task ID: ${task.id}`)
  console.log(`- Artifact ID: ${artifact.id}`)
  console.log(`- Artifact storage: ${artifact.storage_provider}`)
  console.log(`- Artifact URL: ${artifact.image_url}`)
  console.log(`- Artifact bytes: ${artifact.file_size}`)
  console.log(`- Balance: ${config.initialBalance} -> ${balance}`)
  console.log(`- Actual cost: ${actualCost}`)
  console.log(`- Billing status: ${billingStatus}`)
  if (failedTask && refundedUsageRecord) {
    console.log(`- Refunded task ID: ${failedTask.id}`)
    console.log(`- Refunded billing status: ${refundedUsageRecord.billing_status}`)
    console.log(`- Refunded balance: ${refundedUsageRecord.balance_snapshot}`)
  }
  console.log(`- Cleanup: ${config.cleanup ? 'enabled' : 'disabled'}`)
  console.log('')
  console.log('Image Workspace local E2E complete.')

  if (config.cleanup) {
    cleanupSmokeUser(user)
    if (localArtifactPath) {
      rmSync(localArtifactPath, { force: true })
    }
  }
  restoreRuntimeSettings(runtimeSettingsSnapshot)
  runtimeSettingsSnapshotForCleanup = null
}

main().catch((error) => {
  try {
    cleanupSmokeUser(smokeUserForCleanup)
    if (localArtifactPathForCleanup) {
      rmSync(localArtifactPathForCleanup, { force: true })
    }
    restoreRuntimeSettings(runtimeSettingsSnapshotForCleanup)
  } catch {
    // Best-effort cleanup only; keep the original failure visible.
  }
  console.error(error instanceof Error ? error.stack || error.message : String(error))
  process.exit(1)
})
