export function resolveShellLabelOverrides<K extends string>(
  raw: string | undefined,
  runtimeLocale: string,
  allowedKeys: readonly K[],
): Partial<Record<K, string>> {
  if (!raw) return {}
  try {
    const parsed = JSON.parse(raw) as Record<string, { labels?: Record<string, unknown> } | undefined>
    const normalizedLocale = runtimeLocale.toLowerCase()
    const localeKeys = [normalizedLocale, normalizedLocale.split('-')[0], 'en', 'zh']
    for (const key of localeKeys) {
      const labels = parsed[key]?.labels
      if (!labels) continue
      const result: Partial<Record<K, string>> = {}
      for (const labelKey of allowedKeys) {
        const value = labels[labelKey]
        if (typeof value === 'string') {
          result[labelKey] = value
        }
      }
      return result
    }
  } catch {
    return {}
  }
  return {}
}
