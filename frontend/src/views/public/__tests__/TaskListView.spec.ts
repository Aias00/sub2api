import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const taskListViewSource = readFileSync('src/views/public/TaskListView.vue', 'utf8')
const wechatApiSource = readFileSync('src/api/wechat-export.ts', 'utf8')

describe('TaskListView', () => {
  it('aggregates existing WeChat export and image workspace task APIs', () => {
    expect(taskListViewSource).toContain('listWeChatExportTasks({ page: 1, page_size: taskPageSize })')
    expect(taskListViewSource).toContain('listImageWorkspaceTasks({ page: 1, page_size: taskPageSize })')
    expect(taskListViewSource).toContain('mapWeChatTask')
    expect(taskListViewSource).toContain('mapImageTask')
  })

  it('supports loading additional pages for each task source', () => {
    expect(taskListViewSource).toContain('async function loadMoreTasks(type: TaskType)')
    expect(taskListViewSource).toContain("loadMoreTasks('wechat')")
    expect(taskListViewSource).toContain("loadMoreTasks('image')")
    expect(taskListViewSource).toContain('wechatPage.value < wechatPages.value')
    expect(taskListViewSource).toContain('imagePage.value < imagePages.value')
  })

  it('keeps task operations delegated to their original feature APIs', () => {
    expect(taskListViewSource).toContain('cancelWeChatExportTask(task.id)')
    expect(taskListViewSource).toContain('retryWeChatExportTask(task.id)')
    expect(taskListViewSource).toContain('downloadWeChatExportTaskZip(task.id)')
    expect(taskListViewSource).toContain('cancelImageWorkspaceTask(task.id)')
    expect(taskListViewSource).toContain('retryImageWorkspaceTask(task.id)')
    expect(taskListViewSource).toContain('downloadImageWorkspaceArtifact(task.artifactId)')
  })

  it('shows a busy label while a task download is preparing', () => {
    expect(taskListViewSource).toContain("busyTaskKey === task.key ? t('taskList.downloading') : t('taskList.download')")
  })

  it('uses the prompt catalog style public navigation and hero layout', () => {
    expect(taskListViewSource).toContain('PublicDarkHeader')
    expect(taskListViewSource).toContain(':account-label="isAuthenticated ? t(\'nav.dashboard\') : t(\'common.login\')"')
    expect(taskListViewSource).toContain('lg:grid-cols-[minmax(0,1fr)_minmax(360px,520px)]')
  })

  it('uses Vercel-inspired task page chrome', () => {
    expect(taskListViewSource).toContain('home-business-page min-h-screen bg-[#101114] text-white')
    expect(taskListViewSource).toContain('rounded-2xl border border-white/10 bg-[#17181d]')
    expect(taskListViewSource).toContain('lg:grid-cols-[300px_minmax(0,1fr)]')
    expect(taskListViewSource).toContain('font-mono text-xs text-white/38')
    expect(taskListViewSource).toContain('grid-cols-[minmax(0,1fr)_140px_140px_170px_190px]')
  })

  it('preserves the original WeChat task API default pagination', () => {
    expect(wechatApiSource).toContain('listWeChatExportTasks(params: { page?: number; page_size?: number } = {})')
    expect(wechatApiSource).toContain('page: params.page ?? 1')
    expect(wechatApiSource).toContain('page_size: params.page_size ?? 20')
  })

  it('keeps WeChat article loading paginated instead of fetching every page concurrently', () => {
    expect(wechatApiSource).toContain('listWeChatArticles(params: { page?: number; page_size?: number } = {})')
    expect(wechatApiSource).toContain('page: params.page ?? 1')
    expect(wechatApiSource).toContain('page_size: params.page_size ?? 100')
    expect(wechatApiSource).not.toContain('Promise.all(')
    expect(wechatApiSource).not.toContain('total: items.length, page: 1, page_size: items.length, pages: 1')
  })
})
