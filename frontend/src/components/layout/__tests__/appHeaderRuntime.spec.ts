import { describe, expect, it } from 'vitest'

import {
  resolveCompactUserDropdown,
  resolveHeaderDisplayName,
  resolveHeaderPageTitle,
  resolveHeaderUserInitials,
} from '../appHeaderRuntime'

describe('appHeaderRuntime', () => {
  it('resolves display name and initials from username/email', () => {
    expect(resolveHeaderDisplayName({ username: 'alice', email: 'alice@example.com' })).toBe('alice')
    expect(resolveHeaderDisplayName({ email: 'bob@example.com' })).toBe('bob')
    expect(resolveHeaderUserInitials({ username: 'alice' })).toBe('AL')
    expect(resolveHeaderUserInitials({ email: 'bob@example.com' })).toBe('BO')
  })

  it('resolves compact dropdown state', () => {
    expect(resolveCompactUserDropdown(false, false)).toBe(true)
    expect(resolveCompactUserDropdown(true, false)).toBe(false)
  })

  it('resolves page title from custom menu item or route meta', () => {
    expect(resolveHeaderPageTitle({
      routeName: 'CustomPage',
      routeCustomId: 'abc',
      publicMenuItems: [{ id: 'abc', label: 'Custom menu' }],
      adminMenuItems: [],
      isAdmin: false,
      translate: (key) => key,
    })).toBe('Custom menu')

    expect(resolveHeaderPageTitle({
      routeName: 'Usage',
      routeMetaTitleKey: 'nav.usage',
      publicMenuItems: [],
      adminMenuItems: [],
      isAdmin: false,
      translate: (key) => `translated:${key}`,
    })).toBe('translated:nav.usage')
  })
})
