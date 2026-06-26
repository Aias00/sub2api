import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const componentSource = readFileSync('src/components/common/SubscriptionProgressMini.vue', 'utf8')

describe('SubscriptionProgressMini route defaults', () => {
  it('uses auth route defaults for the subscriptions shortcut', () => {
    expect(componentSource).toContain('useAuthRouteDefaults')
    expect(componentSource).toContain(':to="authRouteDefaults.subscriptionsPath"')
    expect(componentSource).not.toContain('to="/subscriptions"')
  })
})
