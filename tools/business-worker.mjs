#!/usr/bin/env node
import { spawn, spawnSync } from 'node:child_process'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const root = dirname(fileURLToPath(import.meta.url))
const once = process.argv.includes('--once')
const healthcheck = process.argv.includes('--healthcheck')

function env(key, fallback = '') {
  const value = process.env[key]
  return value === undefined || value === '' ? fallback : value
}

function boolEnv(key, fallback = false) {
  const value = process.env[key]
  if (value === undefined || value === '') return fallback
  return ['1', 'true', 'yes', 'on'].includes(value.trim().toLowerCase())
}

function workerEnabled(modeKey, configured) {
  const mode = env(modeKey, 'auto').trim().toLowerCase()
  if (['0', 'false', 'no', 'off', 'disabled'].includes(mode)) return false
  if (['1', 'true', 'yes', 'on', 'enabled'].includes(mode)) return true
  return configured
}

const workerSpecs = [
  {
    name: 'wechat-export',
    enabled: workerEnabled(
      'BUSINESS_WORKER_ENABLE_WECHAT_EXPORT',
      Boolean(env('WECHAT_EXPORT_WORKER_TOKEN')) || boolEnv('WECHAT_EXPORT_ALLOW_PRIVATE_WORKER_WITHOUT_TOKEN'),
    ),
    command: 'npm',
    args: ['--prefix', join(root, 'wechat-worker'), 'run', 'worker', ...(once ? ['--', '--once'] : [])],
    healthArgs: ['--prefix', join(root, 'wechat-worker'), 'run', 'worker', '--', '--healthcheck'],
  },
  {
    name: 'image-workspace',
    enabled: workerEnabled(
      'BUSINESS_WORKER_ENABLE_IMAGE_WORKSPACE',
      Boolean(env('IMAGE_WORKSPACE_WORKER_TOKEN')) || boolEnv('IMAGE_WORKSPACE_ALLOW_PRIVATE_WORKER_WITHOUT_TOKEN'),
    ),
    command: 'node',
    args: [join(root, 'image-workspace-worker', 'src', 'worker.mjs'), ...(once ? ['--once'] : [])],
    healthArgs: [join(root, 'image-workspace-worker', 'src', 'worker.mjs'), '--healthcheck'],
  },
]

const workers = workerSpecs.filter((worker) => worker.enabled)

if (workers.length === 0) {
  console.error('[business-worker] no child workers enabled; set BUSINESS_WORKER_ENABLE_WECHAT_EXPORT or BUSINESS_WORKER_ENABLE_IMAGE_WORKSPACE, or configure the matching worker token/private fallback')
  process.exit(1)
}

if (healthcheck) {
  for (const worker of workers) {
    const result = spawnSync(worker.command, worker.healthArgs, {
      env: process.env,
      stdio: 'inherit',
      timeout: 30_000,
    })
    if (result.status !== 0) {
      process.exit(result.status || 1)
    }
  }
  console.log('[business-worker] healthcheck ok')
  process.exit(0)
}

const children = new Map()
let shuttingDown = false

function stopAll(signal = 'SIGTERM') {
  shuttingDown = true
  for (const child of children.values()) {
    if (!child.killed) {
      child.kill(signal)
    }
  }
}

for (const worker of workers) {
  const child = spawn(worker.command, worker.args, {
    env: process.env,
    stdio: 'inherit',
  })
  children.set(worker.name, child)
  console.log(`[business-worker] started ${worker.name} pid=${child.pid}`)

  child.on('exit', (code, signal) => {
    children.delete(worker.name)
    console.log(`[business-worker] ${worker.name} exited code=${code ?? 'null'} signal=${signal ?? 'null'}`)
    if (shuttingDown) {
      if (children.size === 0) process.exit(0)
      return
    }
    stopAll()
    process.exit(code || 1)
  })
}

process.on('SIGTERM', () => stopAll('SIGTERM'))
process.on('SIGINT', () => stopAll('SIGINT'))
