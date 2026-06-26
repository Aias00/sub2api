export function resolveLocalizedShellLabels<K extends string>(
  raw: string | undefined,
  runtimeLocale: string,
  allowedKeys: readonly K[],
): Record<K, string> {
  const empty = emptyLabels(allowedKeys)
  if (!raw?.trim()) return empty
  try {
    const parsed = JSON.parse(raw) as unknown
    if (!isRecord(parsed)) return empty

    const localized = pickLocalizedRecord(parsed, runtimeLocale)
    if (!localized) return empty

    const labels = isRecord(localized.labels)
      ? readAllowedLabels(localized.labels, allowedKeys)
      : {}

    return {
      ...empty,
      ...labels,
    }
  } catch {
    return empty
  }
}

function emptyLabels<K extends string>(allowedKeys: readonly K[]): Record<K, string> {
  return Object.fromEntries(allowedKeys.map((key) => [key, ''])) as Record<K, string>
}

function pickLocalizedRecord(value: Record<string, unknown>, runtimeLocale: string): Record<string, unknown> | null {
  const normalizedLocale = runtimeLocale.toLowerCase()
  const baseLocale = normalizedLocale.split('-')[0]
  const localeKeys = [normalizedLocale, baseLocale, 'en', 'zh']

  for (const key of localeKeys) {
    const localized = value[key]
    if (isRecord(localized)) return localized
  }

  return value
}

function readAllowedLabels<K extends string>(
  labels: Record<string, unknown>,
  allowedKeys: readonly K[],
): Partial<Record<K, string>> {
  const result: Partial<Record<K, string>> = {}
  for (const key of allowedKeys) {
    const value = labels[key]
    if (typeof value === 'string') {
      result[key] = value
    }
  }
  return result
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}
