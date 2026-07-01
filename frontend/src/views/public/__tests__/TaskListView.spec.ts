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
    expect(taskListViewSource).toContain('xl:grid-cols-[minmax(0,1fr)_minmax(320px,460px)]')
  })

  it('keeps task controls compact above the task results', () => {
    expect(taskListViewSource).toContain('mt-8 min-w-0 space-y-5')
    expect(taskListViewSource).toContain('xl:grid-cols-[minmax(0,1fr)_auto]')
    expect(taskListViewSource).toContain('mt-5 grid min-w-0 gap-3 sm:grid-cols-2 xl:grid-cols-4')
    expect(taskListViewSource).toContain('@click="setSummaryFilter(summary.key)"')
    expect(taskListViewSource).not.toContain('lg:grid-cols-[300px_minmax(0,1fr)]')
    expect(taskListViewSource).not.toContain('lg:sticky lg:top-6')
  })

  it('prevents wide task controls from creating horizontal page overflow', () => {
    expect(taskListViewSource).toContain('public-template-container-wide overflow-x-hidden')
    expect(taskListViewSource).toContain('public-template-panel min-w-0 overflow-hidden')
    expect(taskListViewSource).toContain('grid min-w-0 gap-2 sm:grid-cols-2')
    expect(taskListViewSource).toContain('class="public-template-button-primary h-11 w-full')
  })

  it('uses a full-width task table and useful empty-state actions', () => {
    expect(taskListViewSource).toContain('grid-cols-[minmax(0,1.45fr)_120px_120px_160px_220px]')
    expect(taskListViewSource).toContain(':to="wechatPath" class="inline-flex h-11')
    expect(taskListViewSource).toContain(':to="imageGeneratorPath" class="public-template-button-primary h-11')
    expect(taskListViewSource).toContain("const wechatPath = computed(() => props.appShell ? '/app/wechat' : '/wechat')")
    expect(taskListViewSource).toContain("const imageGeneratorPath = computed(() => props.appShell ? '/app/image-generator' : '/image-generator')")
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
