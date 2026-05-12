import { describe, expect, it } from 'vitest'

import { shouldAutoStartAdminOnboarding } from '../useOnboardingTour'

describe('shouldAutoStartAdminOnboarding', () => {
  it('allows admin onboarding auto-start on dashboard entry points only', () => {
    expect(shouldAutoStartAdminOnboarding('/dashboard')).toBe(true)
    expect(shouldAutoStartAdminOnboarding('/admin/dashboard')).toBe(true)
    expect(shouldAutoStartAdminOnboarding('/dashboard/')).toBe(true)
  })

  it('blocks admin onboarding auto-start on other admin pages', () => {
    expect(shouldAutoStartAdminOnboarding('/admin/settings')).toBe(false)
    expect(shouldAutoStartAdminOnboarding('/admin/orders/plans')).toBe(false)
    expect(shouldAutoStartAdminOnboarding('/admin/users')).toBe(false)
  })
})
