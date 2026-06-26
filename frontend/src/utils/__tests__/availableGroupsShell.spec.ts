import { describe, expect, it } from 'vitest'
import {
  availableGroupsLabelKeys,
  renderAvailableGroupsShellText,
  resolveAvailableGroupsShellLabels,
} from '../availableGroupsShell'

describe('available groups shell helpers', () => {
  it('resolves available groups labels from localized shell config', () => {
    const labels = resolveAvailableGroupsShellLabels(
      JSON.stringify({
        zh: {
          labels: {
            title: '可用分组',
            dailyLimit: '每日 {amount}',
            ignored: 'ignored',
          },
        },
      }),
      'zh-CN',
    )

    expect(availableGroupsLabelKeys).toContain('dailyLimit')
    expect(labels.title).toBe('可用分组')
    expect(labels.dailyLimit).toBe('每日 {amount}')
    expect(labels.loadFailed).toBeUndefined()
  })

  it('renders available groups labels with placeholders', () => {
    const labels = {
      dailyLimit: 'Daily {amount}',
      unlimited: 'Unlimited',
    }

    expect(renderAvailableGroupsShellText(labels, 'dailyLimit', { amount: '20.00' })).toBe('Daily 20.00')
    expect(renderAvailableGroupsShellText(labels, 'unlimited')).toBe('Unlimited')
    expect(renderAvailableGroupsShellText(labels, 'loadFailed')).toBe('')
  })
})
