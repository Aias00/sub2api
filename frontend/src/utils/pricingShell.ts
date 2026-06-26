export type PricingCopy = {
  prompts: string
  eyebrow: string
  title: string
  description: string
  catalogStatus: string
  rechargeProducts: string
  subscriptionPlans: string
  recharge: string
  subscription: string
  buy: string
  rechargeCta: string
  subscriptionCta: string
  loadFailed: string
  emptyRecharge: string
  emptyPlans: string
  recommended: string
  creditedBalance: string
  rate: string
  quota: string
  unlimited: string
  day: string
  days: string
  month: string
}

export type PricingShellConfig = {
  button?: {
    title?: string
  }
  defaults?: {
    promptsPath?: string
    purchasePath?: string
  }
  labels: PricingCopy
  groups?: Array<{
    name?: string
    title?: string
  }>
}

const pricingLabelKeys: Array<keyof PricingCopy> = [
  'prompts',
  'eyebrow',
  'title',
  'description',
  'catalogStatus',
  'rechargeProducts',
  'subscriptionPlans',
  'recharge',
  'subscription',
  'buy',
  'rechargeCta',
  'subscriptionCta',
  'loadFailed',
  'emptyRecharge',
  'emptyPlans',
  'recommended',
  'creditedBalance',
  'rate',
  'quota',
  'unlimited',
  'day',
  'days',
  'month',
]

export function resolvePricingShellConfig(
  raw: string | undefined,
  selectedLocale: 'zh' | 'en',
): PricingShellConfig {
  if (!raw?.trim()) return { labels: emptyPricingCopy() }
  try {
    const parsed = JSON.parse(raw) as unknown
    if (!isRecord(parsed)) {
      return { labels: emptyPricingCopy() }
    }
    const scoped = selectedLocale === 'zh' ? parsed.zh : parsed.en
    const value = isRecord(scoped) ? scoped : parsed
    const labels = isRecord(value.labels) ? readPricingLabels(value.labels) : {}

    return {
      button: isRecord(value.button) ? { title: readString(value.button.title) } : undefined,
      defaults: isRecord(value.defaults)
        ? {
            promptsPath: readInternalPath(value.defaults.promptsPath),
            purchasePath: readInternalPath(value.defaults.purchasePath),
          }
        : undefined,
      labels: {
        ...emptyPricingCopy(),
        ...labels,
      },
      groups: Array.isArray(value.groups)
        ? value.groups
          .filter(isRecord)
          .map((group) => ({
            name: readString(group.name),
            title: readString(group.title),
          }))
        : undefined,
    }
  } catch {
    return { labels: emptyPricingCopy() }
  }
}

function emptyPricingCopy(): PricingCopy {
  return Object.fromEntries(pricingLabelKeys.map((key) => [key, ''])) as PricingCopy
}

function readPricingLabels(value: Record<string, unknown>): Partial<PricingCopy> {
  const copy: Partial<PricingCopy> = {}
  for (const key of pricingLabelKeys) {
    const label = readString(value[key])
    if (label) copy[key] = label
  }
  return copy
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

function readString(value: unknown) {
  return typeof value === 'string' ? value : undefined
}

function readInternalPath(value: unknown): string | undefined {
  const path = readString(value)?.trim()
  if (!path || !path.startsWith('/') || path.startsWith('//')) return undefined
  if (path.includes('://') || path.includes('\n') || path.includes('\r')) return undefined
  return path
}
