import type { ApiKey } from '@/types'

export function maskApiGuideKey(key: string): string {
  if (key.length <= 14) return key
  return `${key.slice(0, 8)}...${key.slice(-4)}`
}

export function buildApiGuideKeyOptions(
  keys: ApiKey[],
  noGroupAssignedLabel: string,
): Array<{ value: number; label: string; description: string }> {
  return keys.map((key) => ({
    value: key.id,
    label: `${key.name} · ${maskApiGuideKey(key.key)}`,
    description: key.group?.name || noGroupAssignedLabel,
  }))
}

export function resolveApiGuideAuthHeaderPreview(
  hasGoogHeaderVariant: boolean,
): string {
  return hasGoogHeaderVariant
    ? 'x-goog-api-key: <API_KEY>'
    : 'Authorization: Bearer <API_KEY>'
}
