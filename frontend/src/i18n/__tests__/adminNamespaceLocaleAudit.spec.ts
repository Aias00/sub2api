import { readdirSync, readFileSync, statSync } from 'node:fs'
import { join } from 'node:path'

import { describe, expect, it } from 'vitest'

import en from '../locales/en'
import zh from '../locales/zh'

const TARGET_PREFIXES = [
  'admin.backup.',
  'admin.channelMonitor.',
  'admin.channels.',
  'admin.dashboard.',
  'admin.groups.',
  'admin.ops.',
  'admin.proxies.',
  'admin.riskControl.',
  'admin.redeem.',
  'admin.subscriptions.',
  'admin.usage.',
] as const

function flatten(obj: unknown, prefix = '', out: Record<string, unknown> = {}) {
  if (!obj || typeof obj !== 'object' || Array.isArray(obj)) return out
  for (const [key, value] of Object.entries(obj)) {
    const next = prefix ? `${prefix}.${key}` : key
    if (value && typeof value === 'object' && !Array.isArray(value)) {
      flatten(value, next, out)
    } else {
      out[next] = value
    }
  }
  return out
}

function walkFiles(dir: string, out: string[] = []) {
  for (const name of readdirSync(dir)) {
    const full = join(dir, name)
    const stat = statSync(full)
    if (stat.isDirectory()) {
      if (['i18n', '__tests__', 'node_modules', 'dist'].includes(name)) continue
      walkFiles(full, out)
      continue
    }
    if (/\.(vue|ts)$/.test(name)) out.push(full)
  }
  return out
}

function collectReferencedKeys() {
  const files = walkFiles(join(process.cwd(), 'src'))
  const re = /(?:\$t|\bt|\btm)\(\s*['"]([A-Za-z0-9_.-]+)['"]/g
  const keys = new Set<string>()

  for (const file of files) {
    const source = readFileSync(file, 'utf8')
    let match: RegExpExecArray | null
    while ((match = re.exec(source))) {
      const key = match[1]
      if (!key.endsWith('.') && TARGET_PREFIXES.some((prefix) => key.startsWith(prefix))) {
        keys.add(key)
      }
    }
  }

  return [...keys].sort()
}

describe('admin namespace locale coverage', () => {
  it('contains every referenced risk control, redeem, and subscriptions key in zh', () => {
    const zhKeys = new Set(Object.keys(flatten(zh)))
    const missing = collectReferencedKeys().filter((key) => !zhKeys.has(key))
    expect(missing).toEqual([])
  })

  it('does not leave Chinese copy in the English locale for those namespaces', () => {
    const enEntries = Object.entries(flatten(en))
    const leaks = enEntries.filter(
      ([key, value]) =>
        TARGET_PREFIXES.some((prefix) => key.startsWith(prefix)) &&
        typeof value === 'string' &&
        /[\u4e00-\u9fff]/.test(value)
    )
    expect(leaks).toEqual([])
  })
})
