import type { UserAvailableChannel } from '@/api/channels'
import type { AvailableChannelsLabelKey } from '@/utils/availableChannelsShell'

export type AvailableChannelsTextGetter = (key: AvailableChannelsLabelKey) => string

export function filterAvailableChannelsByQuery(
  channels: UserAvailableChannel[],
  query: string,
): UserAvailableChannel[] {
  const normalized = query.trim().toLowerCase()
  if (!normalized) return channels
  return channels
    .map((channel) => {
      const nameHit = channel.name.toLowerCase().includes(normalized)
      const descHit = (channel.description || '').toLowerCase().includes(normalized)
      if (nameHit || descHit) return channel
      const matchingSections = channel.platforms.filter(
        (platform) =>
          platform.platform.toLowerCase().includes(normalized) ||
          platform.groups.some((group) => group.name.toLowerCase().includes(normalized)) ||
          platform.supported_models.some((model) => model.name.toLowerCase().includes(normalized)),
      )
      if (matchingSections.length === 0) return null
      return { ...channel, platforms: matchingSections }
    })
    .filter((channel): channel is UserAvailableChannel => channel !== null)
}

export function buildAvailableChannelsColumnLabels(availableChannelsText: AvailableChannelsTextGetter) {
  return {
    name: availableChannelsText('columnName'),
    description: availableChannelsText('columnDescription'),
    platform: availableChannelsText('columnPlatform'),
    groups: availableChannelsText('columnGroups'),
    supportedModels: availableChannelsText('columnSupportedModels'),
  }
}

export function buildAvailableChannelsPricingLabels(availableChannelsText: AvailableChannelsTextGetter) {
  return {
    billingMode: availableChannelsText('pricingBillingMode'),
    billingModeImage: availableChannelsText('pricingBillingModeImage'),
    billingModePerRequest: availableChannelsText('pricingBillingModePerRequest'),
    billingModeToken: availableChannelsText('pricingBillingModeToken'),
    cacheReadPrice: availableChannelsText('pricingCacheReadPrice'),
    cacheWritePrice: availableChannelsText('pricingCacheWritePrice'),
    imageOutputPrice: availableChannelsText('pricingImageOutputPrice'),
    inputPrice: availableChannelsText('pricingInputPrice'),
    intervals: availableChannelsText('pricingIntervals'),
    outputPrice: availableChannelsText('pricingOutputPrice'),
    perRequestPrice: availableChannelsText('pricingPerRequestPrice'),
    unitPerMillion: availableChannelsText('pricingUnitPerMillion'),
    unitPerRequest: availableChannelsText('pricingUnitPerRequest'),
  }
}
