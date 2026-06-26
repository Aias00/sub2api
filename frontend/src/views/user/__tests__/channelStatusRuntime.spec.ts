import { describe, expect, it } from 'vitest'

import {
  buildChannelStatusDetailCache,
  resolveChannelStatusDetailTitle,
  resolveChannelStatusOverallStatus,
  shouldEnsureChannelStatusDetails,
} from '../channelStatusRuntime'

describe('channelStatusRuntime', () => {
  it('resolves overall status from monitor states', () => {
    expect(resolveChannelStatusOverallStatus([])).toBe('operational')
    expect(resolveChannelStatusOverallStatus([{ primary_status: 'operational' }] as any)).toBe('operational')
    expect(resolveChannelStatusOverallStatus([{ primary_status: 'failed' }] as any)).toBe('degraded')
  })

  it('resolves detail title and window detail loading rule', () => {
    expect(resolveChannelStatusDetailTitle({ name: 'Monitor A' } as any, 'fallback')).toBe('Monitor A')
    expect(resolveChannelStatusDetailTitle(null, 'fallback')).toBe('fallback')
    expect(shouldEnsureChannelStatusDetails('7d')).toBe(false)
    expect(shouldEnsureChannelStatusDetails('15d')).toBe(true)
  })

  it('builds detail cache immutably', () => {
    expect(buildChannelStatusDetailCache({}, 1, { id: 1 } as any)).toEqual({ 1: { id: 1 } })
  })
})
