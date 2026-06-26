import { describe, expect, it } from 'vitest'
import { customPageLabelKeys, renderCustomPageShellText, resolveCustomPageShellLabels } from '../customPageShell'

describe('custom page shell helpers', () => {
  it('resolves custom page labels from localized shell config', () => {
    const labels = resolveCustomPageShellLabels(
      JSON.stringify({
        zh: {
          labels: {
            notFoundTitle: '不存在',
            openInNewTab: '新窗口',
            ignored: 'ignored',
          },
        },
      }),
      'zh-CN',
    )

    expect(customPageLabelKeys).toContain('markdownLoadFailed')
    expect(labels.notFoundTitle).toBe('不存在')
    expect(labels.openInNewTab).toBe('新窗口')
    expect(labels.copyCode).toBeUndefined()
  })

  it('renders empty text for missing custom page labels', () => {
    expect(renderCustomPageShellText({ tocTitle: '目录' }, 'tocTitle')).toBe('目录')
    expect(renderCustomPageShellText({}, 'copyCodeFailed')).toBe('')
  })
})
