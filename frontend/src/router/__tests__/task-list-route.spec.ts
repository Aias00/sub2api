import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const routerSource = readFileSync('src/router/index.ts', 'utf8')

describe('task list route', () => {
  it('registers a standalone task list entry without replacing original pages', () => {
    expect(routerSource).toContain("path: '/tasks'")
    expect(routerSource).toContain("name: 'TaskList'")
    expect(routerSource).toContain("component: () => import('@/views/public/TaskListView.vue')")
    expect(routerSource).toContain("titleKey: 'nav.myTasks'")
    expect(routerSource).not.toContain("title: 'My Tasks'")
    expect(routerSource).toContain("path: '/wechat'")
    expect(routerSource).toContain("path: '/image-generator'")
  })

  it('allows public business routes in backend mode', () => {
    expect(routerSource).toContain("'/prompts'")
    expect(routerSource).toContain("'/image-generator'")
    expect(routerSource).toContain("'/tasks'")
    expect(routerSource).toContain("'/hot'")
    expect(routerSource).toContain("'/wechat'")
    expect(routerSource).toContain("'/wechat-export'")
  })
})
