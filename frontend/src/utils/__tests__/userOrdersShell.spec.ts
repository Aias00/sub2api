import { describe, expect, it } from 'vitest'
import { renderUserOrdersShellText, resolveUserOrdersShellLabels, userOrdersLabelKeys } from '../userOrdersShell'

describe('user orders shell helpers', () => {
  it('resolves order labels from payment shell config', () => {
    const labels = resolveUserOrdersShellLabels(
      JSON.stringify({
        zh: {
          labels: {
            orderId: '订单 ID',
            statusCompleted: '已完成',
            ignored: 'ignored',
          },
        },
      }),
      'zh-CN',
    )

    expect(userOrdersLabelKeys).toContain('statusRefundFailed')
    expect(labels.orderId).toBe('订单 ID')
    expect(labels.statusCompleted).toBe('已完成')
    expect(labels.refresh).toBeUndefined()
  })

  it('renders empty text for missing order labels', () => {
    const labels = resolveUserOrdersShellLabels(undefined, 'en')
    expect(renderUserOrdersShellText(labels, 'orderId')).toBe('')
  })
})
