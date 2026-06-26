export type CreditsCopy = {
  eyebrow: string
  title: string
  description: string
  purchase: string
  orders: string
  credits: string
  sub2apiBalance: string
  conversion: string
  balanceLabel: string
  actionsTitle: string
  actionsDescription: string
  recharge: string
  viewOrders: string
}

export type CreditsShellConfig = {
  labels: CreditsCopy
  defaults?: {
    purchasePath?: string
    ordersPath?: string
  }
  actions?: {
    title?: string
    description?: string
  }
  buttons?: {
    recharge?: string
    orders?: string
  }
  conversion?: string
}

const creditsLabelKeys: Array<keyof CreditsCopy> = [
  'eyebrow',
  'title',
  'description',
  'purchase',
  'orders',
  'credits',
  'sub2apiBalance',
  'conversion',
  'balanceLabel',
  'actionsTitle',
  'actionsDescription',
  'recharge',
  'viewOrders',
]

export function resolveCreditsShellConfig(
  raw: string | undefined,
  selectedLocale: 'zh' | 'en',
): CreditsShellConfig {
  if (!raw?.trim()) {
    return { labels: emptyCreditsCopy() }
  }
  try {
    const parsed = JSON.parse(raw) as unknown
    if (!isRecord(parsed)) {
      return { labels: emptyCreditsCopy() }
    }
    const localized = parsed[selectedLocale] ?? parsed.en ?? parsed.zh ?? parsed
    if (!isRecord(localized)) {
      return { labels: emptyCreditsCopy() }
    }
    const labels = isRecord(localized.labels) ? readCreditsLabels(localized.labels) : {}

    return {
      labels: {
        ...emptyCreditsCopy(),
        ...labels,
      },
      defaults: isRecord(localized.defaults)
        ? {
            purchasePath: readInternalPath(localized.defaults.purchasePath),
            ordersPath: readInternalPath(localized.defaults.ordersPath),
          }
        : undefined,
      actions: isRecord(localized.actions)
        ? {
            title: readString(localized.actions.title),
            description: readString(localized.actions.description),
          }
        : undefined,
      buttons: isRecord(localized.buttons)
        ? {
            recharge: readString(localized.buttons.recharge),
            orders: readString(localized.buttons.orders),
          }
        : undefined,
      conversion: readString(localized.conversion),
    }
  } catch {
    return { labels: emptyCreditsCopy() }
  }
}

function emptyCreditsCopy(): CreditsCopy {
  return Object.fromEntries(creditsLabelKeys.map((key) => [key, ''])) as CreditsCopy
}

function readCreditsLabels(labels: Record<string, unknown>): Partial<CreditsCopy> {
  const copy: Partial<CreditsCopy> = {}
  for (const key of creditsLabelKeys) {
    const label = readString(labels[key])
    if (label) copy[key] = label
  }
  return copy
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

function readString(value: unknown): string | undefined {
  return typeof value === 'string' ? value : undefined
}

function readInternalPath(value: unknown): string | undefined {
  const path = readString(value)?.trim()
  if (!path || !path.startsWith('/') || path.startsWith('//')) return undefined
  if (path.includes('://') || path.includes('\n') || path.includes('\r')) return undefined
  return path
}
