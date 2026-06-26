import { describe, expect, it, vi, beforeEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { readFileSync } from 'node:fs'

import UserBalanceHistoryModal from '../UserBalanceHistoryModal.vue'

const modalSource = readFileSync('src/components/admin/user/UserBalanceHistoryModal.vue', 'utf8')

const getUserBalanceHistory = vi.hoisted(() => vi.fn())
const appStoreState = vi.hoisted(() => ({
  cachedPublicSettings: {
    redeem_shell_config: JSON.stringify({
      en: {
        labels: {
          adminAdjustment: 'Configured admin adjustment',
          balanceAddedRedeem: 'Configured redeem balance',
          balanceAddedAffiliate: 'Configured affiliate balance',
          balanceAddedAdmin: 'Configured admin balance',
          balanceDeductedAdmin: 'Configured admin deduction',
          concurrencyAddedRedeem: 'Configured redeem concurrency',
          concurrencyAddedAdmin: 'Configured admin concurrency',
          concurrencyReducedAdmin: 'Configured admin concurrency reduction',
          subscriptionAssigned: 'Configured subscription',
          unknown: 'Configured unknown',
        },
      },
    }),
  },
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    users: {
      getUserBalanceHistory,
    },
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => appStoreState,
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, values?: Record<string, string | number>) => {
        if (!values) return key
        return Object.entries(values).reduce(
          (text, [name, value]) => text.split(`{${name}}`).join(String(value)),
          key,
        )
      },
      locale: { value: 'en-US' },
    }),
  }
})

describe('UserBalanceHistoryModal', () => {
  beforeEach(() => {
    getUserBalanceHistory.mockReset().mockResolvedValue({
      items: [
        {
          id: 1,
          type: 'affiliate_balance',
          value: 2,
          used_at: '2026-06-19T00:00:00Z',
          created_at: '2026-06-19T00:00:00Z',
          code: 'ABCDEFGH1234',
        },
        {
          id: 2,
          type: 'admin_balance',
          value: 5,
          used_at: '2026-06-19T00:00:00Z',
          created_at: '2026-06-19T00:00:00Z',
          code: 'ADMIN123456',
        },
      ],
      total: 2,
      total_recharged: 7,
    })
  })

  it('renders balance history type labels from redeem shell config', async () => {
    const wrapper = mount(UserBalanceHistoryModal, {
      props: {
        show: false,
        user: {
          id: 7,
          email: 'user@example.com',
          username: 'demo',
          balance: 9,
          created_at: '2026-01-01T00:00:00Z',
        },
      },
      global: {
        stubs: {
          BaseDialog: {
            props: ['show', 'title'],
            template: '<section v-if="show"><h1>{{ title }}</h1><slot /></section>',
          },
          Select: true,
          Icon: true,
        },
      },
    })

    await wrapper.setProps({ show: true })
    await flushPromises()

    expect(wrapper.text()).toContain('Configured affiliate balance')
    expect(wrapper.text()).toContain('Configured admin balance')
    expect(wrapper.text()).toContain('Configured admin adjustment')
  })

  it('does not read redeem labels from the local i18n namespace', () => {
    expect(modalSource).not.toContain("t('redeem.")
    expect(modalSource).toContain('resolveRedeemShellLabels')
    expect(modalSource).toContain('renderRedeemShellText')
  })
})
