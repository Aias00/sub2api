import { describe, expect, it } from 'vitest'
import { resolveDocumentTitle, resolveRouteDocumentTitle } from '@/router/title'

describe('resolveDocumentTitle', () => {
  it('路由存在标题时，使用“路由标题 - 站点名”格式', () => {
    expect(resolveDocumentTitle('Usage Records', 'My Site')).toBe('Usage Records - My Site')
  })

  it('路由无标题时，回退到站点名', () => {
    expect(resolveDocumentTitle(undefined, 'My Site')).toBe('My Site')
  })

  it('站点名为空时，不补本地默认品牌', () => {
    expect(resolveDocumentTitle('Dashboard', '')).toBe('Dashboard')
    expect(resolveDocumentTitle(undefined, '   ')).toBe('')
  })

  it('站点名变更时仅影响后续路由标题计算', () => {
    const before = resolveDocumentTitle('Admin Dashboard', 'Alpha')
    const after = resolveDocumentTitle('Admin Dashboard', 'Beta')

    expect(before).toBe('Admin Dashboard - Alpha')
    expect(after).toBe('Admin Dashboard - Beta')
  })

  it('user shell-backed routes do not depend on local i18n title metadata', () => {
    const routerSource = readFileSync('src/router/index.ts', 'utf8')

    expect(routerSource).not.toContain("titleKey: 'dashboard.title'")
    expect(routerSource).not.toContain("descriptionKey: 'dashboard.welcomeMessage'")
    expect(routerSource).not.toContain("titleKey: 'keys.title'")
    expect(routerSource).not.toContain("descriptionKey: 'keys.description'")
    expect(routerSource).not.toContain("titleKey: 'usage.title'")
    expect(routerSource).not.toContain("descriptionKey: 'usage.description'")
  })
})

describe('resolveRouteDocumentTitle', () => {
  it('自定义页面菜单加载后，使用菜单名称作为标题', () => {
    const route = {
      name: 'CustomPage',
      params: { id: 'scheduler' },
      meta: {
        title: 'Custom Page'
      }
    }

    expect(resolveRouteDocumentTitle(route, 'EzouAPI')).toBe('Custom Page - EzouAPI')
    expect(resolveRouteDocumentTitle(route, 'EzouAPI', [
      {
        id: 'scheduler',
        label: '账号调度器',
        icon_svg: '',
        url: 'https://example.com',
        visibility: 'admin',
        sort_order: 0
      }
    ])).toBe('账号调度器 - EzouAPI')
  })
})
