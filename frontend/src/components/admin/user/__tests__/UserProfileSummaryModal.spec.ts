import { describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import UserProfileSummaryModal from '../UserProfileSummaryModal.vue'
import type { AdminUser } from '@/types'
import type { UserProfileSummary } from '@/api/admin/users'

const { getUserProfileSummary } = vi.hoisted(() => ({
  getUserProfileSummary: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    users: {
      getUserProfileSummary
    }
  }
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => {
        if (key === 'admin.users.profileSummary.timelineCount') return `${params?.count} records`
        if (key.endsWith('.gateway')) return 'Gateway'
        return key
      }
    })
  }
})

const user: AdminUser = {
  id: 42,
  email: 'timeline@example.com',
  username: 'timeline',
  role: 'user',
  balance: 0,
  paid_balance: 0,
  gift_balance: 0,
  total_recharged: 0,
  concurrency: 1,
  status: 'active',
  signup_source: 'email',
  allowed_groups: [],
  balance_notify_enabled: false,
  balance_notify_threshold: null,
  balance_notify_extra_emails: [],
  created_at: '2026-07-07T00:00:00Z',
  updated_at: '2026-07-07T00:00:00Z',
  notes: ''
}

function createSummary(): UserProfileSummary {
  return {
    user: {
      id: 42,
      email: 'timeline@example.com',
      username: 'timeline',
      role: 'user',
      status: 'active',
      signup_source: 'email',
      balance: 0,
      paid_balance: 0,
      gift_balance: 0,
      total_recharged: 0,
      concurrency: 1,
      created_at: '2026-07-07T00:00:00Z',
      updated_at: '2026-07-07T00:00:00Z'
    },
    classification: { category: 'registered', label: '注册用户', confidence: 'medium', reasons: [] },
    registration: {
      registered_via: 'email',
      same_ip_signup_count_24h: 1,
      same_domain_signup_count: 1,
      email_domain: 'example.com',
      disposable_email: false
    },
    auth_identities: [],
    activity: { api_usage_count: 1, api_actual_cost: 0.12 },
    api_keys: { total_count: 1, active_count: 1 },
    payments: { order_count: 0, paid_order_count: 0, paid_amount: 0, refund_amount: 0 },
    balance: { ledger_count: 0, positive_ledger_amount: 0, net_ledger_amount: 0, redeem_count: 0, redeem_balance_amount: 0 },
    business: {
      image_task_count: 0,
      image_success_count: 0,
      image_actual_cost: 0,
      wechat_task_count: 0,
      wechat_actual_cost: 0
    },
    timeline: [
      {
        occurred_at: '2026-07-07T08:00:00Z',
        source: 'gateway',
        action: 'api_usage',
        title: 'API 调用',
        detail: 'gpt-5 · req-1',
        status: 'completed',
        amount: -0.1234,
        ip_address: '203.0.113.9',
        user_agent: 'Mozilla/5.0',
        record_id: '99'
      }
    ],
    risk_tags: []
  }
}

describe('UserProfileSummaryModal', () => {
  it('renders operation timeline records from the profile summary', async () => {
    getUserProfileSummary.mockResolvedValue(createSummary())

    const wrapper = mount(UserProfileSummaryModal, {
      props: { show: true, user },
      global: {
        stubs: {
          BaseDialog: { template: '<div><slot /></div>' }
        }
      }
    })

    await flushPromises()

    expect(getUserProfileSummary).toHaveBeenCalledWith(42)
    expect(wrapper.text()).toContain('admin.users.profileSummary.timeline')
    expect(wrapper.text()).toContain('1 records')
    expect(wrapper.text()).toContain('Gateway')
    expect(wrapper.text()).toContain('API 调用')
    expect(wrapper.text()).toContain('-$0.1234')
    expect(wrapper.text()).toContain('IP 203.0.113.9')
  })
})
