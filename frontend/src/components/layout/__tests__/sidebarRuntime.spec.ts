import { describe, expect, it } from 'vitest'

import {
  applySidebarFeatureFlags,
  buildSidebarSections,
  buildSidebarVisibleItemMap,
} from '../sidebarRuntime'

describe('sidebarRuntime', () => {
  it('filters sidebar items by feature flag recursively', () => {
    const items = applySidebarFeatureFlags([
      { path: '/a', label: 'A', icon: null },
      { path: '/b', label: 'B', icon: null, featureFlag: () => false },
      {
        path: '/c',
        label: 'C',
        icon: null,
        children: [
          { path: '/c/1', label: 'C1', icon: null, featureFlag: () => false },
          { path: '/c/2', label: 'C2', icon: null },
        ],
      },
    ])

    expect(items.map((item) => item.path)).toEqual(['/a', '/c'])
    expect(items[1].children?.map((child) => child.path)).toEqual(['/c/2'])
  })

  it('builds visible item maps and merges remaining/custom sections', () => {
    const visibleMap = buildSidebarVisibleItemMap(
      {
        a: { path: '/a', label: 'A', icon: null },
        b: { path: '/b', label: 'B', icon: null, hideInSimpleMode: true },
      },
      true,
    )

    expect(Object.keys(visibleMap)).toEqual(['a'])

    const sections = buildSidebarSections(
      [{ id: 'main', items: ['a'] }],
      [{ id: 'fallback', items: ['a', 'b'] }],
      visibleMap,
      [{ path: '/custom', label: 'Custom', icon: null }],
      'more',
      'custom',
    )

    expect(sections.map((section) => section.id)).toEqual(['main', 'custom'])
  })
})
