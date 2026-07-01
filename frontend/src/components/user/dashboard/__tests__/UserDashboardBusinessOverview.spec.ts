import { describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import UserDashboardBusinessOverview from '../UserDashboardBusinessOverview.vue'

const listWeChatExportTasks = vi.hoisted(() => vi.fn())
const listImageWorkspaceTasks = vi.hoisted(() => vi.fn())

vi.mock('@/api/wechat-export', () => ({
  listWeChatExportTasks,
}))

vi.mock('@/api/image-workspace', () => ({
  listImageWorkspaceTasks,
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      locale: { value: 'zh-CN' },
    }),
  }
})

describe('UserDashboardBusinessOverview', () => {
  it('renders business capability cards and recent WeChat/image tasks', async () => {
    listWeChatExportTasks.mockResolvedValue({
      items: [
        {
          id: 23,
          status: 'completed',
          formats: ['html', 'markdown'],
          selected_article_count: 2,
          successful_article_count: 2,
          failed_article_count: 0,
          created_at: '2026-07-01T08:00:00Z',
          updated_at: '2026-07-01T08:10:00Z',
        },
      ],
    })
    listImageWorkspaceTasks.mockResolvedValue({
      items: [
        {
          id: 36,
          status: 'succeeded',
          model: 'gpt-image-2',
          size: '1024x1024',
          batch_size: 1,
          created_at: '2026-07-01T08:20:00Z',
          updated_at: '2026-07-01T08:25:00Z',
        },
      ],
    })

    const wrapper = mount(UserDashboardBusinessOverview, {
      global: {
        stubs: {
          Icon: { template: '<i />' },
          RouterLink: {
            props: ['to'],
            template: '<a :href="to"><slot /></a>',
          },
        },
      },
    })

    await flushPromises()

    expect(listWeChatExportTasks).toHaveBeenCalledWith({ page: 1, page_size: 5 })
    expect(listImageWorkspaceTasks).toHaveBeenCalledWith({ page: 1, page_size: 5 })
    expect(wrapper.text()).toContain('业务能力与任务')
    expect(wrapper.text()).toContain('微信导出')
    expect(wrapper.text()).toContain('生图记录')
    expect(wrapper.text()).toContain('统一任务')
    expect(wrapper.text()).toContain('#23 · HTML + MARKDOWN')
    expect(wrapper.text()).toContain('#36 · gpt-image-2')
    expect(wrapper.find('a[href="/app/wechat"]').exists()).toBe(true)
    expect(wrapper.find('a[href="/app/image-generator"]').exists()).toBe(true)
    expect(wrapper.find('a[href="/app/tasks"]').exists()).toBe(true)
  })
})
