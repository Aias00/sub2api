import { describe, expect, it } from 'vitest'

import { resolveSelfSidebarSections, selfSidebarItemKeys } from '@/utils/adminSidebarSchema'

describe('adminSidebarSchema self sidebar items', () => {
  it('allows dashboard tool entries in configured user sidebars', () => {
    expect(selfSidebarItemKeys).toEqual(
      expect.arrayContaining(['promptCatalog', 'imageGenerator', 'wechatExport', 'hotTopics'])
    )

    const config = JSON.stringify({
      zh: {
        defaults: {
          userSidebarSections: [
            {
              id: 'main',
              items: ['dashboard', 'promptCatalog', 'imageGenerator', 'wechatExport', 'hotTopics'],
            },
          ],
        },
      },
    })

    expect(resolveSelfSidebarSections(config, 'zh', 'userSidebarSections')).toEqual([
      {
        id: 'main',
        items: ['dashboard', 'promptCatalog', 'imageGenerator', 'wechatExport', 'hotTopics'],
      },
    ])
  })
})
