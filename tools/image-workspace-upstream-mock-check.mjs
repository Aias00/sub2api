#!/usr/bin/env node
import { createServer } from 'node:http'
import { existsSync, mkdtempSync, readFileSync, rmSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join, resolve } from 'node:path'
import { spawn } from 'node:child_process'

const png1x1 = 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/p9sAAAAASUVORK5CYII='
const apiKey = 'mock-image-workspace-key'
const workerToken = 'mock-worker-token'
const outputDir = mkdtempSync(join(tmpdir(), 'cloudbase-image-workspace-upstream-'))
const keepOutput = process.env.KEEP_IMAGE_WORKSPACE_MOCK_OUTPUT === '1'
const workerPath = resolve('tools/image-workspace-worker/src/worker.mjs')
const received = []

function readBody(req) {
  return new Promise((resolveBody, reject) => {
    const chunks = []
    req.on('data', (chunk) => chunks.push(chunk))
    req.on('end', () => resolveBody(Buffer.concat(chunks).toString('utf8')))
    req.on('error', reject)
  })
}

function startMockServer() {
  const server = createServer(async (req, res) => {
    try {
      if (req.method !== 'POST' || req.url !== '/v1/images/generations') {
        res.writeHead(404, { 'content-type': 'application/json' })
        res.end(JSON.stringify({ error: { message: 'not found' } }))
        return
      }
      const raw = await readBody(req)
      const body = raw ? JSON.parse(raw) : {}
      received.push({
        authorization: req.headers.authorization || '',
        content_type: req.headers['content-type'] || '',
        body,
      })
      if (req.headers.authorization !== `Bearer ${apiKey}`) {
        res.writeHead(401, { 'content-type': 'application/json' })
        res.end(JSON.stringify({ error: { message: 'invalid token' } }))
        return
      }
      res.writeHead(200, { 'content-type': 'application/json' })
      res.end(JSON.stringify({
        created: 1893456000,
        data: [
          {
            b64_json: png1x1,
            output_format: 'png',
            revised_prompt: `mock revised: ${body.prompt || ''}`,
          },
        ],
      }))
    } catch (error) {
      res.writeHead(500, { 'content-type': 'application/json' })
      res.end(JSON.stringify({ error: { message: error instanceof Error ? error.message : String(error) } }))
    }
  })

  return new Promise((resolveServer, reject) => {
    server.on('error', reject)
    server.listen(0, '127.0.0.1', () => resolveServer(server))
  })
}

function startRuntimeConfigServer(runtimeConfig) {
  const server = createServer((req, res) => {
    if (req.headers['x-image-workspace-worker-token'] !== workerToken) {
      res.writeHead(401, { 'content-type': 'application/json' })
      res.end(JSON.stringify({ code: 'IMAGE_WORKSPACE_WORKER_UNAUTHORIZED', message: 'invalid token' }))
      return
    }
    if (req.method === 'GET' && req.url === '/api/v1/image-workspace/worker/runtime-config') {
      res.writeHead(200, { 'content-type': 'application/json' })
      res.end(JSON.stringify({ data: runtimeConfig }))
      return
    }
    res.writeHead(404, { 'content-type': 'application/json' })
    res.end(JSON.stringify({ code: 'NOT_FOUND', message: `${req.method} ${req.url}` }))
  })

  return new Promise((resolveServer, reject) => {
    server.on('error', reject)
    server.listen(0, '127.0.0.1', () => resolveServer(server))
  })
}

function runWorker(upstreamURL) {
  return new Promise((resolveRun, reject) => {
    const child = spawn(process.execPath, [workerPath, '--upstream-check'], {
      cwd: process.cwd(),
      env: {
        ...process.env,
        IMAGE_WORKSPACE_API_BASE_URL: process.env.IMAGE_WORKSPACE_API_BASE_URL,
        IMAGE_WORKSPACE_WORKER_TOKEN: workerToken,
        IMAGE_WORKSPACE_UPSTREAM_API_KEY: apiKey,
        IMAGE_WORKSPACE_OUTPUT_DIR: outputDir,
        IMAGE_WORKSPACE_STORAGE_KEY_ROOT: outputDir,
        IMAGE_WORKSPACE_UPSTREAM_CHECK_PROMPT: 'Mock upstream smoke prompt',
        IMAGE_WORKSPACE_UPSTREAM_CHECK_NEGATIVE_PROMPT: 'low quality',
        IMAGE_WORKSPACE_UPSTREAM_CHECK_STYLE: 'flat vector',
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
  if (!condition) {
    throw new Error(message)
  }
}

let server
let runtimeServer
try {
  server = await startMockServer()
  const address = server.address()
  const upstreamURL = `http://127.0.0.1:${address.port}/v1/images/generations`
  runtimeServer = await startRuntimeConfigServer({
    upstream_url: upstreamURL,
    generation_timeout_ms: 120000,
    completion_cost: '0',
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
  })
  const runtimeAddress = runtimeServer.address()
  process.env.IMAGE_WORKSPACE_API_BASE_URL = `http://127.0.0.1:${runtimeAddress.port}`
  const run = await runWorker(upstreamURL)
  const result = parseWorkerJSON(run.stdout)
  assert(received.length === 1, `expected exactly one upstream request, got ${received.length}`)
  assert(received[0].authorization === `Bearer ${apiKey}`, 'worker did not send expected authorization header')
  assert(received[0].body.model === 'gpt-image-2', 'worker sent unexpected model')
  assert(received[0].body.response_format === 'b64_json', 'worker did not request b64_json output')
  assert(String(received[0].body.prompt || '').includes('Mock upstream smoke prompt'), 'worker prompt was not forwarded')
  assert(String(received[0].body.prompt || '').includes('Negative prompt: low quality'), 'negative prompt was not included')
  assert(String(received[0].body.prompt || '').includes('Style note: flat vector'), 'style note was not included')
  assert(result.ok === true, 'worker upstream check did not report ok')
  assert(result.artifact_count === 1, `expected one artifact, got ${result.artifact_count}`)
  const artifact = result.artifacts?.[0]
  assert(artifact?.storage_provider === 'local', 'artifact was not written with local storage provider')
  assert(artifact?.mime_type === 'image/png', 'artifact mime type was not inferred as image/png')
  assert(artifact?.file_size > 0, 'artifact file size was not recorded')
  assert(artifact?.checksum, 'artifact checksum was not recorded')
  assert(existsSync(artifact.storage_key), `artifact file not found: ${artifact.storage_key}`)
  const bytes = readFileSync(artifact.storage_key)
  assert(bytes.length === artifact.file_size, 'artifact file size does not match written file')

  console.log('# Image Workspace Upstream Mock Check')
  console.log('')
  console.log(`- Upstream requests: ${received.length}`)
  console.log(`- Model: ${received[0].body.model}`)
  console.log(`- Response format: ${received[0].body.response_format}`)
  console.log(`- Artifact count: ${result.artifact_count}`)
  console.log(`- Artifact file size: ${artifact.file_size}`)
  console.log(`- Output dir: ${outputDir}`)
  console.log('')
  console.log('Image Workspace upstream mock check complete.')
} finally {
  if (runtimeServer) {
    await new Promise((resolveClose) => runtimeServer.close(resolveClose))
  }
  if (server) {
    await new Promise((resolveClose) => server.close(resolveClose))
  }
  if (!keepOutput && existsSync(outputDir)) {
    rmSync(outputDir, { recursive: true, force: true })
  }
}
