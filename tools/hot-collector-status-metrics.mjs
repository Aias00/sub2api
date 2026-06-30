#!/usr/bin/env node
import { existsSync, readFileSync, writeFileSync } from 'node:fs'

const statusPath = env('HOT_WORKER_STATUS_PATH', '/app/runtime/hot-worker-status.json')
const outputPath = env('HOT_WORKER_METRICS_PATH', '')

function env(key, fallback) {
  const value = process.env[key]
  return value === undefined || value === '' ? fallback : value
}

function assert(condition, message) {
  if (!condition) throw new Error(message)
}

function parseTime(value) {
  const timestamp = Date.parse(String(value || ''))
  return Number.isFinite(timestamp) ? timestamp : 0
}

function secondsSince(value) {
  const timestamp = parseTime(value)
  if (timestamp <= 0) return -1
  return Math.max(0, Math.floor((Date.now() - timestamp) / 1000))
}

function metricLine(name, value, labels = {}) {
  const labelEntries = Object.entries(labels).filter(([, labelValue]) => labelValue !== undefined && labelValue !== '')
  const labelText = labelEntries.length === 0
    ? ''
    : `{${labelEntries.map(([key, value]) => `${key}="${String(value).replaceAll('\\', '\\\\').replaceAll('"', '\\"')}"`).join(',')}}`
  return `${name}${labelText} ${value}`
}

function statusGaugeLines(currentStatus) {
  const statuses = ['ok', 'running', 'error', 'fatal']
  return statuses.map((status) =>
    metricLine('sub2api_hot_collector_status', currentStatus === status ? 1 : 0, { status }),
  )
}

function modeGaugeLines(currentMode) {
  const modes = ['loop', 'once']
  return modes.map((mode) =>
    metricLine('sub2api_hot_collector_mode', currentMode === mode ? 1 : 0, { mode }),
  )
}

function renderMetrics(status) {
  const updatedAge = secondsSince(status.updated_at)
  const successAge = secondsSince(status.last_finished_at)
  const failureAge = secondsSince(status.last_failed_at)
  const durationMs = Number(status.last_run_duration_ms || 0)
  const intervalMs = Number(status.interval_ms || 0)
  const maxBackoffMs = Number(status.max_backoff_ms || 0)
  const maxRuns = Number(status.max_runs || 0)
  return [
    '# HELP sub2api_hot_collector_status Hot collector worker status labels.',
    '# TYPE sub2api_hot_collector_status gauge',
    ...statusGaugeLines(String(status.status || 'unknown')),
    '# HELP sub2api_hot_collector_apply_mode Whether the worker status represents apply/import mode (1) or dry-run mode (0).',
    '# TYPE sub2api_hot_collector_apply_mode gauge',
    metricLine('sub2api_hot_collector_apply_mode', status.apply === true ? 1 : 0),
    '# HELP sub2api_hot_collector_mode Hot collector worker runtime mode labels.',
    '# TYPE sub2api_hot_collector_mode gauge',
    ...modeGaugeLines(String(status.mode || 'unknown')),
    '# HELP sub2api_hot_collector_interval_seconds Configured collector interval in seconds.',
    '# TYPE sub2api_hot_collector_interval_seconds gauge',
    metricLine('sub2api_hot_collector_interval_seconds', Math.max(0, intervalMs / 1000)),
    '# HELP sub2api_hot_collector_max_backoff_seconds Configured collector max backoff in seconds.',
    '# TYPE sub2api_hot_collector_max_backoff_seconds gauge',
    metricLine('sub2api_hot_collector_max_backoff_seconds', Math.max(0, maxBackoffMs / 1000)),
    '# HELP sub2api_hot_collector_max_runs Configured max runs, or 0 for unlimited.',
    '# TYPE sub2api_hot_collector_max_runs gauge',
    metricLine('sub2api_hot_collector_max_runs', Math.max(0, maxRuns)),
    '# HELP sub2api_hot_collector_status_age_seconds Seconds since the worker status file was updated.',
    '# TYPE sub2api_hot_collector_status_age_seconds gauge',
    metricLine('sub2api_hot_collector_status_age_seconds', updatedAge),
    '# HELP sub2api_hot_collector_last_run_duration_seconds Last collector run duration in seconds.',
    '# TYPE sub2api_hot_collector_last_run_duration_seconds gauge',
    metricLine('sub2api_hot_collector_last_run_duration_seconds', Math.max(0, durationMs / 1000)),
    '# HELP sub2api_hot_collector_last_success_age_seconds Seconds since the last successful collector run, or -1 if absent.',
    '# TYPE sub2api_hot_collector_last_success_age_seconds gauge',
    metricLine('sub2api_hot_collector_last_success_age_seconds', successAge),
    '# HELP sub2api_hot_collector_last_failure_age_seconds Seconds since the last failed collector run, or -1 if absent.',
    '# TYPE sub2api_hot_collector_last_failure_age_seconds gauge',
    metricLine('sub2api_hot_collector_last_failure_age_seconds', failureAge),
    '# HELP sub2api_hot_collector_run_count Total collector runs recorded in the status file.',
    '# TYPE sub2api_hot_collector_run_count counter',
    metricLine('sub2api_hot_collector_run_count', Number(status.run_count || 0)),
    '# HELP sub2api_hot_collector_success_count Total successful collector runs recorded in the status file.',
    '# TYPE sub2api_hot_collector_success_count counter',
    metricLine('sub2api_hot_collector_success_count', Number(status.success_count || 0)),
    '# HELP sub2api_hot_collector_failure_count Total failed collector runs recorded in the status file.',
    '# TYPE sub2api_hot_collector_failure_count counter',
    metricLine('sub2api_hot_collector_failure_count', Number(status.failure_count || 0)),
    '',
  ].join('\n')
}

function main() {
  assert(existsSync(statusPath), `HOT_WORKER_STATUS_PATH not found: ${statusPath}`)
  const status = JSON.parse(readFileSync(statusPath, 'utf8'))
  assert(status && typeof status === 'object', 'status file did not contain a JSON object')
  assert(status.updated_at, 'status file is missing updated_at')
  const metrics = renderMetrics(status)
  if (outputPath) {
    writeFileSync(outputPath, metrics)
    console.log(`hot_collector_metrics_path=${outputPath}`)
  } else {
    process.stdout.write(metrics)
  }
}

try {
  main()
} catch (error) {
  console.error(error instanceof Error ? error.message : String(error))
  process.exit(1)
}
