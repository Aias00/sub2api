import { readFileSync } from 'node:fs'
import { describe, expect, it, beforeEach, afterEach, vi } from 'vitest'

const opsSource = readFileSync('src/api/admin/ops.ts', 'utf8')

describe('admin ops websocket runtime URL', () => {
  beforeEach(() => {
    vi.resetModules()
    delete window.__APP_CONFIG__
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('uses the same-origin Sub2API base path by default', async () => {
    const { resolveOpsWebSocketURL } = await import('../ops')
    const expectedOrigin = new URL(window.location.origin)
    expectedOrigin.protocol = expectedOrigin.protocol === 'https:' ? 'wss:' : 'ws:'

    expect(resolveOpsWebSocketURL('/admin/ops/ws/qps')).toBe(
      `${expectedOrigin.origin}/api/v1/admin/ops/ws/qps`,
    )
  })

  it('follows injected public api_base_url for remote Sub2API runtimes', async () => {
    window.__APP_CONFIG__ = {
      api_base_url: 'https://api.example.com/sub2api/v1/',
    } as typeof window.__APP_CONFIG__
    const { resolveOpsWebSocketURL } = await import('../ops')

    expect(resolveOpsWebSocketURL('/admin/ops/ws/qps')).toBe(
      'wss://api.example.com/sub2api/v1/admin/ops/ws/qps',
    )
  })

  it('keeps explicit websocket host overrides for tests and special deployments', async () => {
    const { resolveOpsWebSocketURL } = await import('../ops')
    const expectedProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'

    expect(resolveOpsWebSocketURL('/admin/ops/ws/qps', 'ops.example.com')).toBe(
      `${expectedProtocol}//ops.example.com/api/v1/admin/ops/ws/qps`,
    )
    expect(resolveOpsWebSocketURL('/admin/ops/ws/qps', 'https://ops.example.com')).toBe(
      'wss://ops.example.com/api/v1/admin/ops/ws/qps',
    )
  })

  it('does not use frontend build-time websocket env fallback', () => {
    expect(opsSource).not.toContain('VITE_WS_BASE_URL')
    expect(opsSource).toContain('resolveApiBaseUrl')
  })
})
