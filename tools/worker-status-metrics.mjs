#!/usr/bin/env node
import { existsSync, readFileSync, writeFileSync } from 'node:fs'

const service = env('WORKER_STATUS_SERVICE', '').toLowerCase()
const inputPath = env('WORKER_STATUS_INPUT_PATH', '')
const inputURL = env('WORKER_STATUS_URL', '')
const outputPath = env('WORKER_STATUS_METRICS_PATH', '')
const authHeader = env('WORKER_STATUS_AUTH_HEADER', '')
const workerToken = env('WORKER_STATUS_WORKER_TOKEN', '')
const workerTokenHeader = env('WORKER_STATUS_WORKER_TOKEN_HEADER', '')

const serviceConfigs = {
  wechat: {
    prefix: 'cloudbase_wechat_export',
    healthMetric: 'cloudbase_wechat_export_worker_health',
    healthValues: ['idle', 'waiting', 'active', 'attention'],
    countFields: [
      ['total_count', 'total_count'],
      ['queued_count', 'queued_count'],
      ['running_count', 'running_count'],
      ['stale_running_count', 'stale_running_count'],
      ['failed_count', 'failed_count'],
      ['completed_count', 'completed_count'],
      ['cancelled_count', 'cancelled_count'],
      ['oldest_queued_seconds', 'oldest_queued_seconds'],
      ['last_task_age_seconds', 'last_task_age_seconds'],
    ],
  },
  image: {
    prefix: 'cloudbase_image_workspace',
    healthMetric: 'cloudbase_image_workspace_worker_health',
    healthValues: ['idle', 'waiting', 'active', 'attention'],
    countFields: [
      ['total_count', 'total_count'],
      ['queued_count', 'queued_count'],
      ['running_count', 'running_count'],
      ['stale_running_count', 'stale_running_count'],
      ['failed_count', 'failed_count'],
      ['succeeded_count', 'succeeded_count'],
      ['cancelled_count', 'cancelled_count'],
      ['artifact_count', 'artifact_count'],
      ['oldest_queued_seconds', 'oldest_queued_seconds'],
      ['last_task_age_seconds', 'last_task_age_seconds'],
    ],
  },
  'image-workspace': null,
}
serviceConfigs['image-workspace'] = serviceConfigs.image

function env(key, fallback) {
  const value = process.env[key]
  return value === undefined || value === '' ? fallback : value
}

function assert(condition, message) {
  if (!condition) throw new Error(message)
}

function metricName(config, suffix) {
  return `${config.prefix}_${suffix}`
}

function labelValue(value) {
  return String(value).replaceAll('\\', '\\\\').replaceAll('"', '\\"')
}

function metricLine(name, value, labels = {}) {
  const labelEntries = Object.entries(labels).filter(([, label]) => label !== undefined && label !== '')
  const labelText = labelEntries.length === 0
    ? ''
    : `{${labelEntries.map(([key, label]) => `${key}="${labelValue(label)}"`).join(',')}}`
  return `${name}${labelText} ${value}`
}

function numberValue(value) {
  if (value === null || value === undefined || value === '') return 0
  const number = Number(value)
  return Number.isFinite(number) ? number : 0
}

function renderMetrics(status, config) {
  const health = String(status.health || 'idle')
  const attentionReasons = Array.isArray(status.attention_reasons) ? status.attention_reasons : []
  const lines = [
    `# HELP ${config.healthMetric} Worker health state exported from the Cloudbase status API.`,
    `# TYPE ${config.healthMetric} gauge`,
    ...config.healthValues.map((value) => metricLine(config.healthMetric, health === value ? 1 : 0, { health: value })),
  ]

  for (const [jsonField, suffix] of config.countFields) {
    const name = metricName(config, suffix)
    lines.push(`# HELP ${name} Worker status field ${jsonField}.`)
    lines.push(`# TYPE ${name} gauge`)
    lines.push(metricLine(name, numberValue(status[jsonField])))
  }

  const attentionName = metricName(config, 'attention_reason')
  lines.push(`# HELP ${attentionName} Attention reasons exposed by the worker status API.`)
  lines.push(`# TYPE ${attentionName} gauge`)
  if (attentionReasons.length === 0) {
    lines.push(metricLine(attentionName, 0, { reason: 'none' }))
  } else {
    for (const reason of attentionReasons) {
      lines.push(metricLine(attentionName, 1, { reason }))
    }
  }
  lines.push('')
  return lines.join('\n')
}

async function readStatusJSON() {
  if (inputPath) {
    assert(existsSync(inputPath), `WORKER_STATUS_INPUT_PATH not found: ${inputPath}`)
    return JSON.parse(readFileSync(inputPath, 'utf8'))
  }
  assert(inputURL, 'Set WORKER_STATUS_INPUT_PATH or WORKER_STATUS_URL')
  const headers = { accept: 'application/json' }
  if (authHeader) {
    const separator = authHeader.includes(':') ? ':' : ' '
    const [name, ...parts] = authHeader.split(separator)
    if (name && parts.length > 0) headers[name.trim()] = parts.join(separator).trim()
  }
  if (workerToken && workerTokenHeader) {
    headers[workerTokenHeader] = workerToken
  }
  const response = await fetch(inputURL, { headers })
  assert(response.ok, `GET ${inputURL} failed: HTTP ${response.status}`)
  return response.json()
}

async function main() {
  const config = serviceConfigs[service]
  assert(config, 'WORKER_STATUS_SERVICE must be one of: wechat, image, image-workspace')
  const status = await readStatusJSON()
  assert(status && typeof status === 'object' && !Array.isArray(status), 'worker status must be a JSON object')
  const metrics = renderMetrics(status, config)
  if (outputPath) {
    writeFileSync(outputPath, metrics)
    console.log(`worker_status_metrics_path=${outputPath}`)
  } else {
    process.stdout.write(metrics)
  }
}

try {
  await main()
} catch (error) {
  console.error(error instanceof Error ? error.message : String(error))
  process.exit(1)
}
