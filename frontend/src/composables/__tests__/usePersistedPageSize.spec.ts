import { afterEach, describe, expect, it } from 'vitest'

import { getPersistedPageSize, setPersistedPageSize } from '@/composables/usePersistedPageSize'

describe('usePersistedPageSize', () => {
  afterEach(() => {
    localStorage.clear()
    delete window.__APP_CONFIG__
  })

  it('uses the system table default instead of stale localStorage state', () => {
    window.__APP_CONFIG__ = {
      table_default_page_size: 1000,
      table_page_size_options: [20, 50, 1000]
    } as any
    localStorage.setItem('table-page-size', '50')
    localStorage.setItem('table-page-size-source', 'user')

    expect(getPersistedPageSize()).toBe(1000)
  })

  it('does not persist user-selected page size locally', () => {
    setPersistedPageSize(50)

    expect(localStorage.getItem('table-page-size')).toBeNull()
  })
})
