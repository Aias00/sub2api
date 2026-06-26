import { describe, expect, it } from 'vitest'

import {
  filterAvailableGroupsByQuery,
  resolveAvailableGroupQuotaSummary,
  resolveAvailableGroupSubscriptionLabel,
  resolveMemberAvailableGroups,
  resolvePublicAvailableGroups,
} from '../availableGroupsRuntime'

describe('availableGroupsRuntime', () => {
  const groups = [
    {
      name: 'Claude Public',
      description: 'Public Claude access',
      platform: 'anthropic',
      subscription_type: 'standard',
      is_exclusive: false,
      daily_limit_usd: null,
      weekly_limit_usd: null,
      monthly_limit_usd: null,
    },
    {
      name: 'GPT Pro',
      description: 'Subscription GPT access',
      platform: 'openai',
      subscription_type: 'subscription',
      is_exclusive: false,
      daily_limit_usd: 20,
      weekly_limit_usd: null,
      monthly_limit_usd: null,
    },
  ] as any[]

  const text = (key: any, values?: Record<string, string | number>) =>
    values?.amount !== undefined ? `${key}:${values.amount}` : `label:${key}`

  it('filters groups by search and splits public/member groups', () => {
    expect(filterAvailableGroupsByQuery(groups as any, 'gpt')).toHaveLength(1)
    expect(resolvePublicAvailableGroups(groups as any)).toHaveLength(1)
    expect(resolveMemberAvailableGroups(groups as any)).toHaveLength(1)
  })

  it('renders subscription labels and quota summaries', () => {
    expect(resolveAvailableGroupSubscriptionLabel('subscription' as any, text)).toBe('label:subscriptionBadge')
    expect(resolveAvailableGroupQuotaSummary(groups[1] as any, text)).toBe('dailyLimit:20.00')
    expect(resolveAvailableGroupQuotaSummary(groups[0] as any, text)).toBe('label:unlimited')
  })
})
