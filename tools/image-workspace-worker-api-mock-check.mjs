#!/usr/bin/env node
import { createServer } from 'node:http'
import { mkdtempSync, rmSync, existsSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join, resolve } from 'node:path'
import { spawn } from 'node:child_process'

const png1x1 = 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/p9sAAAAASUVORK5CYII='
const workerToken = 'mock-worker-token'
const upstreamKey = 'mock-upstream-key'
const outputDir = mkdtempSync(join(tmpdir(), 'sub2api-image-worker-api-'))
const keepOutput = process.env.KEEP_IMAGE_WORKSPACE_MOCK_OUTPUT === '1'
const workerPath = resolve('tools/image-workspace-worker/src/worker.mjs')
const upstreamRequests = []
const completedTasks = []
const failedTasks = []
let claimCount = 0
let runtimeConfig = null

function readBody(req) {
  return new Promise((resolveBody, reject) => {
    const chunks = []
    req.on('data', (chunk) => chunks.push(chunk))
    req.on('end', () => {
      const raw = Buffer.concat(chunks).toString('utf8')
      resolveBody(raw ? JSON.parse(raw) : {})
    })
    req.on('error', reject)
  })
}

function json(res, status, body) {
  res.writeHead(status, { 'content-type': 'application/json' })
  res.end(JSON.stringify(body))
}

function startUpstreamServer() {
  const server = createServer(async (req, res) => {
    if (req.method !== 'POST' || req.url !== '/v1/images/generations') {
      json(res, 404, { error: { message: 'not found' } })
      return
    }
    const body = await readBody(req)
    upstreamRequests.push({ headers: req.headers, body })
    json(res, 200, {
      created: 1893456000,
      data: [{ b64_json: png1x1, output_format: 'png', revised_prompt: `api mock: ${body.prompt || ''}` }],
    })
  })
  return listen(server)
}

function startBackendServer() {
  const server = createServer(async (req, res) => {
    if (req.headers['x-image-workspace-worker-token'] !== workerToken) {
      json(res, 401, { code: 'IMAGE_WORKSPACE_WORKER_UNAUTHORIZED', message: 'invalid token' })
      return
    }
    if (req.method === 'GET' && req.url === '/api/v1/image-workspace/worker/runtime-config') {
      json(res, 200, { data: runtimeConfig })
      return
    }
    if (req.method === 'POST' && req.url === '/api/v1/image-workspace/worker/tasks/claim') {
      claimCount += 1
      await readBody(req)
      json(res, 200, {
        data: {
          task: {
            id: 101,
            user_id: 42,
            status: 'running',
            prompt: 'Mock backend task prompt',
            negative_prompt: 'watermark',
            model: 'gpt-image-2',
            provider: 'openai',
            size: '1024x1024',
            quality: 'standard',
            style: 'editorial',
            batch_size: 1,
            cost_estimate: 0,
            balance_snapshot: 10,
            result_json: '{}',
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
          },
        },
      })
      return
    }
    if (req.method === 'POST' && req.url === '/api/v1/image-workspace/worker/tasks/101/complete') {
      const body = await readBody(req)
      completedTasks.push(body)
      json(res, 200, { data: { id: 101, status: 'succeeded' } })
      return
    }
    if (req.method === 'POST' && req.url === '/api/v1/image-workspace/worker/tasks/101/fail') {
      failedTasks.push(await readBody(req))
      json(res, 200, { data: { id: 101, status: 'failed' } })
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

function runWorker(baseURL, upstreamURL) {
  return new Promise((resolveRun, reject) => {
    const child = spawn(process.execPath, [workerPath, '--once'], {
      cwd: process.cwd(),
      env: {
        ...process.env,
        IMAGE_WORKSPACE_API_BASE_URL: baseURL,
        IMAGE_WORKSPACE_WORKER_TOKEN: workerToken,
        IMAGE_WORKSPACE_UPSTREAM_API_KEY: upstreamKey,
        IMAGE_WORKSPACE_OUTPUT_DIR: outputDir,
        IMAGE_WORKSPACE_STORAGE_KEY_ROOT: outputDir,
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

function assert(condition, message) {
  if (!condition) throw new Error(message)
}

let backendServer
let upstreamServer
try {
  backendServer = await startBackendServer()
  upstreamServer = await startUpstreamServer()
  const backendAddr = backendServer.address()
  const upstreamAddr = upstreamServer.address()
  runtimeConfig = {
    upstream_url: `http://127.0.0.1:${upstreamAddr.port}/v1/images/generations`,
    generation_timeout_ms: 120000,
    completion_cost: '0.25',
    completion_cost_map_json: '{}',
    prompt_safety_enabled: true,
    assume_worker_ready: false,
    object_storage: {
      enabled: false,
      provider: 'r2',
      bucket: '',
      region: 'auto',
      key_prefix: 'image-workspace',
      public_base_url: '',
    },
    media_cdn_base_url: '',
  }
  await runWorker(
    `http://127.0.0.1:${backendAddr.port}`,
    runtimeConfig.upstream_url,
  )

  assert(claimCount === 1, `expected one claim request, got ${claimCount}`)
  assert(upstreamRequests.length === 1, `expected one upstream request, got ${upstreamRequests.length}`)
  assert(completedTasks.length === 1, `expected one complete request, got ${completedTasks.length}`)
  assert(failedTasks.length === 0, `expected no fail requests, got ${failedTasks.length}`)
  assert(upstreamRequests[0].headers.authorization === `Bearer ${upstreamKey}`, 'missing upstream authorization header')
  assert(upstreamRequests[0].body.prompt.includes('Mock backend task prompt'), 'task prompt was not forwarded upstream')
  assert(upstreamRequests[0].body.prompt.includes('Negative prompt: watermark'), 'negative prompt was not forwarded upstream')
  assert(upstreamRequests[0].body.prompt.includes('Style note: editorial'), 'style note was not forwarded upstream')
  const complete = completedTasks[0]
  assert(Array.isArray(complete.artifacts) && complete.artifacts.length === 1, 'complete payload must include one artifact')
  assert(complete.artifacts[0].storage_provider === 'local', 'complete artifact storage provider mismatch')
  assert(complete.artifacts[0].file_size > 0, 'complete artifact missing file size')
  assert(complete.artifacts[0].checksum, 'complete artifact missing checksum')
  assert(existsSync(complete.artifacts[0].storage_key), `artifact file missing: ${complete.artifacts[0].storage_key}`)
  assert(complete.cost === 0.25, `expected complete cost from worker flat cost, got ${complete.cost}`)
  const result = JSON.parse(complete.result_json)
  assert(result.artifact_count === 1, 'result_json missing artifact_count')
  assert(result.cost_source === 'flat', `expected flat cost source, got ${result.cost_source}`)

  console.log('# Image Workspace Worker API Mock Check')
  console.log('')
  console.log(`- Claim requests: ${claimCount}`)
  console.log(`- Upstream requests: ${upstreamRequests.length}`)
  console.log(`- Complete requests: ${completedTasks.length}`)
  console.log(`- Artifact file size: ${complete.artifacts[0].file_size}`)
  console.log(`- Completion cost: ${complete.cost}`)
  console.log(`- Completion cost source: ${result.cost_source}`)
  console.log('')
  console.log('Image Workspace worker API mock check complete.')
} finally {
  if (backendServer) {
    await new Promise((resolveClose) => backendServer.close(resolveClose))
  }
  if (upstreamServer) {
    await new Promise((resolveClose) => upstreamServer.close(resolveClose))
  }
  if (!keepOutput && existsSync(outputDir)) {
    rmSync(outputDir, { recursive: true, force: true })
  }
}
