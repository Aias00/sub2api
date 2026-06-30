#!/usr/bin/env node

const baseURL = env('BASE_URL', 'http://127.0.0.1:8080').replace(/\/+$/, '')
const requireReady = boolEnv('REQUIRE_HOME_BUSINESS_CAPABILITIES_READY', false)
const requiredKeys = env('HOME_BUSINESS_REQUIRED_KEYS', 'wechat-export,prompt-catalog,image-workspace,hot-topics')
  .split(',')
  .map((item) => item.trim())
  .filter(Boolean)
const defaultPaths = {
  'wechat-export': '/wechat',
  'prompt-catalog': '/prompts',
  'image-workspace': '/image-generator',
  'hot-topics': '/hot',
}
const validStatuses = new Set(['available', 'in_progress', 'disabled', 'hidden'])
const pathStatus = new Map()

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
  if (!condition) throw new Error(message)
}

async function getJSON(path) {
  const response = await fetch(`${baseURL}${path}`, {
    headers: { accept: 'application/json' },
  })
  const text = await response.text()
  let payload
  try {
    payload = text ? JSON.parse(text) : {}
  } catch {
    throw new Error(`GET ${path} did not return JSON: HTTP ${response.status} ${text.slice(0, 120)}`)
  }
  if (!response.ok || (typeof payload.code === 'number' && payload.code !== 0)) {
    throw new Error(`GET ${path} failed: HTTP ${response.status} ${text.slice(0, 200)}`)
  }
  return payload.data ?? payload
}

async function probePage(path) {
  if (pathStatus.has(path)) return pathStatus.get(path)
  const response = await fetch(`${baseURL}${path}`, {
    method: 'GET',
    redirect: 'manual',
    headers: { accept: 'text/html,application/xhtml+xml' },
  })
  const ok = response.status >= 200 && response.status < 400
  pathStatus.set(path, { ok, status: response.status })
  return pathStatus.get(path)
}

function safeParseJSON(value, fallback) {
  if (typeof value !== 'string' || value.trim() === '') return fallback
  try {
    return JSON.parse(value)
  } catch {
    return fallback
  }
}

function cardsFromSettings(settings) {
  const config = safeParseJSON(settings?.home_business_shell_config, {})
  const cards = new Map()
  for (const locale of ['zh', 'en']) {
    const localizedCards = Array.isArray(config?.[locale]?.businessCards) ? config[locale].businessCards : []
    for (const card of localizedCards) {
      if (!card || typeof card !== 'object' || typeof card.key !== 'string') continue
      const previous = cards.get(card.key) || {}
      cards.set(card.key, {
        ...previous,
        key: card.key,
        title: previous.title || card.title || '',
        path: previous.path || card.path || '',
        status: previous.status || card.status || 'available',
        visible: card.visible !== false && previous.visible !== false,
        disabled: Boolean(previous.disabled || card.disabled),
      })
    }
  }
  for (const key of requiredKeys) {
    if (!cards.has(key)) {
      cards.set(key, {
        key,
        title: key,
        path: defaultPaths[key] || '',
        status: 'available',
        visible: true,
        disabled: false,
      })
    } else {
      const card = cards.get(key)
      if (!card.path && defaultPaths[key]) card.path = defaultPaths[key]
    }
  }
  return cards
}

function normalizeStatus(status) {
  return typeof status === 'string' && validStatuses.has(status) ? status : ''
}

const settings = await getJSON('/api/v1/settings/public')
const statuses = await getJSON('/api/v1/home/business-capabilities')
assert(statuses && typeof statuses === 'object' && !Array.isArray(statuses), 'business capability response must be an object')

const home = await probePage('/home')
assert(home.ok, `/home must be reachable, got HTTP ${home.status}`)

const cards = cardsFromSettings(settings)
const rows = []
const failures = []

for (const key of requiredKeys) {
  const status = statuses[key]
  if (!status) {
    failures.push(`${key}: missing runtime status`)
    continue
  }
  const runtimeStatus = normalizeStatus(status.status)
  if (!runtimeStatus) {
    failures.push(`${key}: invalid runtime status ${JSON.stringify(status.status)}`)
    continue
  }
  const card = cards.get(key) || { key, path: defaultPaths[key] || '', visible: true, disabled: false, status: 'available' }
  const configuredStatus = normalizeStatus(card.status) || 'available'
  const effectiveStatus = configuredStatus === 'available' ? runtimeStatus : configuredStatus
  const visible = card.visible !== false && effectiveStatus !== 'hidden'
  const disabled = Boolean(card.disabled || effectiveStatus === 'disabled' || effectiveStatus === 'in_progress' || !card.path)
  if (requireReady && effectiveStatus !== 'available') {
    failures.push(`${key}: strict readiness requires available status, got ${effectiveStatus}`)
  }
  let route = null
  if (visible && !disabled && effectiveStatus === 'available') {
    route = await probePage(card.path)
    if (!route.ok) {
      failures.push(`${key}: available card path ${card.path} returned HTTP ${route.status}`)
    }
  }
  if (runtimeStatus === 'available' && ['prompt-catalog', 'hot-topics', 'image-workspace'].includes(key)) {
    const count = Number(status.count || 0)
    if (!Number.isFinite(count) || count <= 0) {
      failures.push(`${key}: available runtime status must include positive count`)
    }
  }
  rows.push({
    key,
    runtime_status: runtimeStatus,
    configured_status: configuredStatus,
    effective_status: effectiveStatus,
    visible,
    disabled,
    path: card.path || '',
    route_status: route ? route.status : '',
    count: status.count || 0,
  })
}

if (failures.length > 0) {
  console.error('# Home Business Capability Smoke')
  console.error('')
  for (const row of rows) {
    console.error(`- ${row.key}: runtime=${row.runtime_status}, effective=${row.effective_status}, path=${row.path || '-'}, route=${row.route_status || '-'}`)
  }
  console.error('')
  for (const failure of failures) {
    console.error(`ERROR: ${failure}`)
  }
  process.exit(2)
}

console.log('# Home Business Capability Smoke')
console.log('')
console.log(`- Base URL: ${baseURL}`)
console.log(`- /home: HTTP ${home.status}`)
for (const row of rows) {
  console.log(`- ${row.key}: runtime=${row.runtime_status}, effective=${row.effective_status}, path=${row.path || '-'}, route=${row.route_status || '-'}, count=${row.count}`)
}
console.log('')
console.log('Home business capability smoke check complete.')
