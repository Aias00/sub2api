import { describe, expect, it } from 'vitest'

import {
  buildAvailableChannelsColumnLabels,
  buildAvailableChannelsPricingLabels,
  filterAvailableChannelsByQuery,
} from '../availableChannelsRuntime'

describe('availableChannelsRuntime', () => {
  const channels = [
    {
      name: 'Claude Channel',
      description: 'Anthropic models',
      platforms: [
        {
          platform: 'anthropic',
          groups: [{ name: 'Claude Public' }],
          supported_models: [{ name: 'claude-opus' }],
        },
      ],
    },
    {
      name: 'GPT Channel',
      description: 'OpenAI models',
      platforms: [
        {
          platform: 'openai',
          groups: [],
          supported_models: [{ name: 'gpt-5' }],
        },
      ],
    },
  ] as any[]

  const text = (key: any) => `label:${key}`

  it('filters channels by top-level fields and nested platform/group/model matches', () => {
    expect(filterAvailableChannelsByQuery(channels as any, 'gpt')).toHaveLength(1)
    expect(filterAvailableChannelsByQuery(channels as any, 'claude-opus')).toHaveLength(1)
  })

  it('builds column and pricing labels from shell text', () => {
    expect(buildAvailableChannelsColumnLabels(text)).toEqual({
      name: 'label:columnName',
      description: 'label:columnDescription',
      platform: 'label:columnPlatform',
      groups: 'label:columnGroups',
      supportedModels: 'label:columnSupportedModels',
    })
    expect(buildAvailableChannelsPricingLabels(text).billingMode).toBe('label:pricingBillingMode')
    expect(buildAvailableChannelsPricingLabels(text).unitPerMillion).toBe('label:pricingUnitPerMillion')
  })
})
