#!/usr/bin/env node
import { createHash } from 'node:crypto'
import { createServer } from 'node:http'
import { mkdtempSync, rmSync, existsSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join, resolve } from 'node:path'
import { spawn } from 'node:child_process'

const png1x1 = 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/p9sAAAAASUVORK5CYII='
const apiKey = 'mock-image-workspace-key'
const accessKeyID = 'mock-r2-access-key'
const secretAccessKey = 'mock-r2-secret-key'
const bucket = 'sub2api-image-workspace'
const keyPrefix = 'image-workspace-e2e'
const publicBaseURL = 'https://assets.example.test/image-workspace'
const workerToken = 'mock-worker-token'
const outputDir = mkdtempSync(join(tmpdir(), 'sub2api-image-object-storage-'))
const keepOutput = process.env.KEEP_IMAGE_WORKSPACE_MOCK_OUTPUT === '1'
const workerPath = resolve('tools/image-workspace-worker/src/worker.mjs')
const upstreamRequests = []
const storageUploads = []

function readRawBody(req) {
  return new Promise((resolveBody, reject) => {
    const chunks = []
    req.on('data', (chunk) => chunks.push(chunk))
    req.on('end', () => resolveBody(Buffer.concat(chunks)))
    req.on('error', reject)
  })
}

function json(res, status, body) {
  res.writeHead(status, { 'content-type': 'application/json' })
  res.end(JSON.stringify(body))
}

function startUpstreamServer() {
  const server = createServer(async (req, res) => {
    try {
      if (req.method !== 'POST' || req.url !== '/v1/images/generations') {
        json(res, 404, { error: { message: 'not found' } })
        return
      }
      const raw = await readRawBody(req)
      const body = raw.length > 0 ? JSON.parse(raw.toString('utf8')) : {}
      upstreamRequests.push({ headers: req.headers, body })
      if (req.headers.authorization !== `Bearer ${apiKey}`) {
        json(res, 401, { error: { message: 'invalid token' } })
        return
      }
      json(res, 200, {
        created: 1893456000,
        data: [{
          b64_json: png1x1,
          output_format: 'png',
          revised_prompt: `object storage mock: ${body.prompt || ''}`,
        }],
      })
    } catch (error) {
      json(res, 500, { error: { message: error instanceof Error ? error.message : String(error) } })
    }
  })
  return listen(server)
}

function startObjectStorageServer() {
  const server = createServer(async (req, res) => {
    try {
      if (req.method !== 'PUT') {
        json(res, 405, { error: 'method not allowed' })
        return
      }
      const bytes = await readRawBody(req)
      const upload = {
        method: req.method,
        url: req.url,
        headers: req.headers,
        bytes,
        checksum: createHash('sha256').update(bytes).digest('hex'),
      }
      storageUploads.push(upload)
      if (!String(req.url || '').startsWith(`/${bucket}/${keyPrefix}/1/1/image-1.png`)) {
        json(res, 400, { error: `unexpected object path: ${req.url}` })
        return
      }
      if (!String(req.headers.authorization || '').startsWith(`AWS4-HMAC-SHA256 Credential=${accessKeyID}/`)) {
        json(res, 401, { error: 'missing signed authorization' })
        return
      }
      if (!req.headers['x-amz-content-sha256']) {
        json(res, 400, { error: 'missing payload hash' })
        return
      }
      res.writeHead(200, { etag: `"${upload.checksum}"` })
      res.end('')
    } catch (error) {
      json(res, 500, { error: error instanceof Error ? error.message : String(error) })
    }
  })
  return listen(server)
}

function startRuntimeConfigServer(runtimeConfig) {
  const server = createServer((req, res) => {
    if (req.headers['x-image-workspace-worker-token'] !== workerToken) {
      json(res, 401, { code: 'IMAGE_WORKSPACE_WORKER_UNAUTHORIZED', message: 'invalid token' })
      return
    }
    if (req.method === 'GET' && req.url === '/api/v1/image-workspace/worker/runtime-config') {
      json(res, 200, { data: runtimeConfig })
      return
    }
    json(res, 404, { code: 'NOT_FOUND', message: `${req.method} ${req.url}` })
  })
  return listen(server)
}

function listen(server) {
  return new Promise((resolveListen, reject) => {
    server.on('error', reject)
    server.listen(0, '127.0.0.1', () => resolveListen(server))
  })
}

function runWorker(baseURL, storageEndpoint) {
  return new Promise((resolveRun, reject) => {
    const child = spawn(process.execPath, [workerPath, '--upstream-check'], {
      cwd: process.cwd(),
      env: {
        ...process.env,
        IMAGE_WORKSPACE_API_BASE_URL: baseURL,
        IMAGE_WORKSPACE_WORKER_TOKEN: workerToken,
        IMAGE_WORKSPACE_UPSTREAM_API_KEY: apiKey,
        IMAGE_WORKSPACE_OUTPUT_DIR: outputDir,
        IMAGE_WORKSPACE_STORAGE_KEY_ROOT: outputDir,
        IMAGE_WORKSPACE_OBJECT_STORAGE_ENDPOINT: storageEndpoint,
        IMAGE_WORKSPACE_OBJECT_STORAGE_ACCESS_KEY_ID: accessKeyID,
        IMAGE_WORKSPACE_OBJECT_STORAGE_SECRET_ACCESS_KEY: secretAccessKey,
        IMAGE_WORKSPACE_OBJECT_STORAGE_CACHE_CONTROL: 'public, max-age=60',
        IMAGE_WORKSPACE_UPSTREAM_CHECK_PROMPT: 'Mock object storage prompt',
        IMAGE_WORKSPACE_UPSTREAM_CHECK_NEGATIVE_PROMPT: 'watermark',
        IMAGE_WORKSPACE_UPSTREAM_CHECK_STYLE: 'clean product render',
      },
      stdio: ['ignore', 'pipe', 'pipe'],
    })
    let stdout = ''
    let stderr = ''
    child.stdout.on('data', (chunk) => {
      stdout += chunk.toString('utf8')
    })
    child.stderr.on('data', (chunk) => {
      stderr += chunk.toString('utf8')
    })
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

function parseWorkerJSON(stdout) {
  const start = stdout.indexOf('{')
  if (start < 0) {
    throw new Error(`worker did not print JSON: ${stdout}`)
  }
  return JSON.parse(stdout.slice(start))
}

function assert(condition, message) {
  if (!condition) throw new Error(message)
}

let upstreamServer
let storageServer
let runtimeServer
try {
  upstreamServer = await startUpstreamServer()
  storageServer = await startObjectStorageServer()
  const upstreamAddr = upstreamServer.address()
  const storageAddr = storageServer.address()
  runtimeServer = await startRuntimeConfigServer({
    upstream_url: `http://127.0.0.1:${upstreamAddr.port}/v1/images/generations`,
    generation_timeout_ms: 120000,
    completion_cost: '0',
    completion_cost_map_json: '{}',
    prompt_safety_enabled: true,
    assume_worker_ready: false,
    object_storage: {
      enabled: true,
      provider: 'r2',
      bucket,
      region: 'auto',
      key_prefix: keyPrefix,
      public_base_url: publicBaseURL,
    },
    media_cdn_base_url: '',
  })
  const runtimeAddr = runtimeServer.address()
  const run = await runWorker(`http://127.0.0.1:${runtimeAddr.port}`, `http://127.0.0.1:${storageAddr.port}`)
  const result = parseWorkerJSON(run.stdout)

  assert(upstreamRequests.length === 1, `expected one upstream request, got ${upstreamRequests.length}`)
  assert(storageUploads.length === 1, `expected one object upload, got ${storageUploads.length}`)
  const upload = storageUploads[0]
  const expectedHash = createHash('sha256').update(upload.bytes).digest('hex')
  assert(upload.headers['x-amz-content-sha256'] === expectedHash, 'signed payload hash does not match uploaded bytes')
  assert(upload.headers['content-type'] === 'image/png', 'object upload content-type mismatch')
  assert(upload.headers['cache-control'] === 'public, max-age=60', 'object upload cache-control mismatch')
  assert(upload.bytes.length > 0, 'object upload body was empty')
  assert(result.ok === true, 'worker object storage check did not report ok')
  assert(result.artifact_count === 1, `expected one artifact, got ${result.artifact_count}`)
  const artifact = result.artifacts?.[0]
  assert(artifact?.storage_provider === 'r2', 'artifact storage provider mismatch')
  assert(artifact?.storage_key === `${keyPrefix}/1/1/image-1.png`, `artifact storage key mismatch: ${artifact?.storage_key}`)
  assert(artifact?.image_url === `${publicBaseURL}/${keyPrefix}/1/1/image-1.png`, `artifact public URL mismatch: ${artifact?.image_url}`)
  assert(artifact?.checksum === expectedHash, 'artifact checksum does not match uploaded bytes')
  assert(artifact?.file_size === upload.bytes.length, 'artifact file size does not match uploaded bytes')

  console.log('# Image Workspace Object Storage Mock Check')
  console.log('')
  console.log(`- Upstream requests: ${upstreamRequests.length}`)
  console.log(`- Object uploads: ${storageUploads.length}`)
  console.log(`- Object path: ${upload.url}`)
  console.log(`- Storage provider: ${artifact.storage_provider}`)
  console.log(`- Storage key: ${artifact.storage_key}`)
  console.log(`- Public URL: ${artifact.image_url}`)
  console.log(`- Uploaded bytes: ${upload.bytes.length}`)
  console.log('')
  console.log('Image Workspace object storage mock check complete.')
} finally {
  if (runtimeServer) {
    await new Promise((resolveClose) => runtimeServer.close(resolveClose))
  }
  if (upstreamServer) {
    await new Promise((resolveClose) => upstreamServer.close(resolveClose))
  }
  if (storageServer) {
    await new Promise((resolveClose) => storageServer.close(resolveClose))
  }
  if (!keepOutput && existsSync(outputDir)) {
    rmSync(outputDir, { recursive: true, force: true })
  }
}
