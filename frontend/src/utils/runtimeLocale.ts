export function resolveRuntimeLocale(locale: unknown): string {
  if (typeof locale === 'string') return locale
  if (locale && typeof locale === 'object' && 'value' in locale) {
    return String((locale as { value?: unknown }).value || '')
  }
  return ''
}

export function resolveRuntimeLanguage(locale: unknown): 'zh' | 'en' {
  return resolveRuntimeLocale(locale).toLowerCase().startsWith('zh') ? 'zh' : 'en'
}
