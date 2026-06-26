export function normalizeUsageServiceTier(serviceTier?: string | null): string | null {
  const value = serviceTier?.trim().toLowerCase()
  if (!value) return null
  if (value === 'fast') return 'priority'
  if (value === 'default' || value === 'standard') return 'standard'
  if (value === 'priority' || value === 'flex') return value
  return value
}

export function formatUsageServiceTier(serviceTier?: string | null): string {
  const normalized = normalizeUsageServiceTier(serviceTier)
  if (!normalized) return 'standard'
  return normalized
}

export function getUsageServiceTierLabel(
  serviceTier: string | null | undefined,
  translate: (key: string) => string,
): string {
  const tier = formatUsageServiceTier(serviceTier)
  if (tier === 'priority') return translate('common.serviceTier.priority')
  if (tier === 'flex') return translate('common.serviceTier.flex')
  if (tier === 'standard') return translate('common.serviceTier.standard')
  return tier
}
