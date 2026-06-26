import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { readFileSync } from 'node:fs'
import AffiliateView from '../AffiliateView.vue'

const affiliateViewSource = readFileSync('src/views/user/AffiliateView.vue', 'utf8')

const getAffiliateDetail = vi.hoisted(() => vi.fn())
const getAffiliateRebates = vi.hoisted(() => vi.fn())
const getAffiliateTransfers = vi.hoisted(() => vi.fn())
const transferAffiliateQuota = vi.hoisted(() => vi.fn())
const copyToClipboard = vi.hoisted(() => vi.fn())
const refreshUser = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())
const showSuccess = vi.hoisted(() => vi.fn())

const appStoreState = vi.hoisted(() => ({
  cachedPublicSettings: {
    affiliate_shell_config: JSON.stringify({
      en: {
        labels: {
          rebateRate: 'Configured rebate rate',
          rebateRateHint: 'Configured rebate hint',
          invitedUsers: 'Configured invited users',
          availableQuota: 'Configured available quota',
          totalQuota: 'Configured total quota',
          frozenQuota: 'Configured frozen',
          title: 'Configured affiliate title',
          description: 'Configured affiliate description',
          yourCode: 'Configured code title',
          copyCode: 'Configured copy code',
          inviteLink: 'Configured invite link',
          copyLink: 'Configured copy link',
          tipsTitle: 'Configured tips',
          tipShare: 'Configured share tip',
          tipRebate: 'Configured rebate {rate}',
          tipTransfer: 'Configured transfer tip',
          tipFreeze: 'Configured freeze tip',
          transferTitle: 'Configured transfer title',
          transferDescription: 'Configured transfer description',
          transferButton: 'Configured transfer button',
          transferEmpty: 'Configured transfer empty',
          transferSuccess: 'Configured transferred {amount}',
          inviteesTitle: 'Configured invitees',
          inviteesEmpty: 'Configured invitees empty',
          emailColumn: 'Configured email column',
          usernameColumn: 'Configured username column',
          rebateColumn: 'Configured rebate column',
          joinedAtColumn: 'Configured joined column',
          rebatesTitle: 'Configured rebates',
          rebatesEmpty: 'Configured rebates empty',
          inviteeColumn: 'Configured invitee column',
          orderAmountColumn: 'Configured order amount column',
          payAmountColumn: 'Configured pay amount column',
          rebateAmountColumn: 'Configured rebate amount column',
          paymentTypeColumn: 'Configured payment type column',
          orderStatusColumn: 'Configured order status column',
          createdAtColumn: 'Configured created column',
          transfersTitle: 'Configured transfers',
          transfersEmpty: 'Configured transfers empty',
          amountColumn: 'Configured amount column',
          balanceAfterColumn: 'Configured balance after column',
          availableQuotaAfterColumn: 'Configured available after column',
          frozenQuotaAfterColumn: 'Configured frozen after column',
          historyQuotaAfterColumn: 'Configured history after column',
          transferredAtColumn: 'Configured transferred at column',
          codeCopied: 'Configured code copied',
          linkCopied: 'Configured link copied',
        },
      },
    }),
  },
  showError,
  showSuccess,
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, string | number>) => {
        if (!params) return key
        return Object.entries(params).reduce(
          (text, [name, value]) => text.split(`{${name}}`).join(String(value)),
          key,
        )
      },
      te: () => false,
      locale: { value: 'en-US' },
    }),
  }
})

vi.mock('@/api/user', () => ({
  default: {
    getAffiliateDetail,
    getAffiliateRebates,
    getAffiliateTransfers,
    transferAffiliateQuota,
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => appStoreState,
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    refreshUser,
  }),
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({ copyToClipboard }),
}))

describe('AffiliateView', () => {
  beforeEach(() => {
    getAffiliateDetail.mockReset().mockResolvedValue({
      user_id: 1,
      aff_code: 'AFF123',
      inviter_id: null,
      aff_count: 2,
      aff_quota: 10,
      aff_frozen_quota: 5,
      aff_history_quota: 30,
      effective_rebate_rate_percent: 12.5,
      invitees: [
        {
          user_id: 2,
          email: 'invitee@example.com',
          username: 'invitee',
          created_at: '2026-06-19T00:00:00Z',
          total_rebate: 3,
        },
      ],
    })
    getAffiliateRebates.mockReset().mockResolvedValue({
      items: [
        {
          order_id: 10,
          out_trade_no: 'order-10',
          inviter_id: 1,
          inviter_email: 'owner@example.com',
          inviter_username: 'owner',
          invitee_id: 2,
          invitee_email: 'invitee@example.com',
          invitee_username: 'invitee',
          order_amount: 20,
          pay_amount: 20,
          rebate_amount: 2.5,
          payment_type: 'stripe',
          order_status: 'paid',
          created_at: '2026-06-19T00:00:00Z',
        },
      ],
      total: 1,
    })
    getAffiliateTransfers.mockReset().mockResolvedValue({
      items: [
        {
          ledger_id: 11,
          user_id: 1,
          user_email: 'owner@example.com',
          username: 'owner',
          amount: 6,
          balance_after: 9,
          available_quota_after: 4,
          frozen_quota_after: 5,
          history_quota_after: 30,
          snapshot_available: true,
          created_at: '2026-06-19T00:00:00Z',
        },
      ],
      total: 1,
    })
    transferAffiliateQuota.mockReset().mockResolvedValue({
      transferred_quota: 10,
      balance: 20,
    })
    copyToClipboard.mockReset().mockResolvedValue(undefined)
    refreshUser.mockReset().mockResolvedValue(undefined)
    showError.mockReset()
    showSuccess.mockReset()
  })

  it('renders affiliate shell labels from public settings', async () => {
    const wrapper = mount(AffiliateView, {
      global: {
        stubs: {
          AppLayout: { template: '<main><slot /></main>' },
          Icon: { template: '<i />' },
          Pagination: { template: '<nav />' },
          OrderStatusBadge: { props: ['status'], template: '<span>{{ status }}</span>' },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('Configured rebate rate')
    expect(wrapper.text()).toContain('Configured rebate hint')
    expect(wrapper.text()).toContain('Configured invited users')
    expect(wrapper.text()).toContain('Configured available quota')
    expect(wrapper.text()).toContain('Configured frozen')
    expect(wrapper.text()).toContain('Configured affiliate title')
    expect(wrapper.text()).toContain('Configured affiliate description')
    expect(wrapper.text()).toContain('Configured copy code')
    expect(wrapper.text()).toContain('Configured invite link')
    expect(wrapper.text()).toContain('Configured rebate 12.5%')
    expect(wrapper.text()).toContain('Configured transfer title')
    expect(wrapper.text()).toContain('Configured transfer button')
    expect(wrapper.text()).toContain('Configured invitees')
    expect(wrapper.text()).toContain('Configured email column')
    expect(wrapper.text()).toContain('Configured rebates')
    expect(wrapper.text()).toContain('Configured payment type column')
    expect(wrapper.text()).toContain('Configured transfers')
    expect(wrapper.text()).toContain('Configured history after column')
    expect(wrapper.text()).toContain('invitee@example.com')
  })

  it('uses configured labels for clipboard and transfer success', async () => {
    const wrapper = mount(AffiliateView, {
      global: {
        stubs: {
          AppLayout: { template: '<main><slot /></main>' },
          Icon: { template: '<i />' },
          Pagination: { template: '<nav />' },
          OrderStatusBadge: { props: ['status'], template: '<span>{{ status }}</span>' },
        },
      },
    })

    await flushPromises()
    await wrapper.get('button.btn-secondary').trigger('click')
    await wrapper.find('button.btn-primary').trigger('click')
    await flushPromises()

    expect(copyToClipboard).toHaveBeenCalledWith('AFF123', 'Configured code copied')
    expect(transferAffiliateQuota).toHaveBeenCalled()
    expect(showSuccess).toHaveBeenCalledWith('Configured transferred $10.00')
    expect(refreshUser).toHaveBeenCalled()
  })

  it('does not keep affiliate shell i18n fallback keys in the view bootstrap layer', () => {
    expect(affiliateViewSource).not.toContain('affiliateFallbackKeys')
    expect(affiliateViewSource).not.toContain('affiliateLabels.value[key] || key')
    expect(affiliateViewSource).not.toContain('affiliate.title')
    expect(affiliateViewSource).not.toContain('affiliate.transfer.success')
    expect(affiliateViewSource).toContain("from './affiliateRuntime'")
    expect(affiliateViewSource).toContain('formatAffiliateNullableCurrency')
    expect(affiliateViewSource).toContain('changeAffiliatePageSize')
    expect(affiliateViewSource).not.toContain('const affiliateLabelKeys')
    expect(affiliateViewSource).not.toContain('resolveShellLabelOverrides(')
    expect(affiliateViewSource).toContain('resolveAffiliateShellLabels')
    expect(affiliateViewSource).toContain('renderAffiliateShellText')
  })

  it('does not keep the legacy affiliate locale section in frontend bundles or route meta', () => {
    const zhLocaleSource = readFileSync('src/i18n/locales/zh.ts', 'utf8')
    const enLocaleSource = readFileSync('src/i18n/locales/en.ts', 'utf8')
    const routerSource = readFileSync('src/router/index.ts', 'utf8')

    for (const source of [zhLocaleSource, enLocaleSource]) {
      expect(source).not.toContain('\n  affiliate: {')
    }
    expect(routerSource).not.toContain("titleKey: 'affiliate.title'")
    expect(routerSource).not.toContain("descriptionKey: 'affiliate.description'")
  })
})
