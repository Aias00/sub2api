import type { PublicSettings } from '@/types'

export const DEFAULT_PASSWORD_MIN_LENGTH = 8
export const MAX_PASSWORD_MIN_LENGTH = 128

export function normalizePasswordMinLength(value: number | null | undefined): number {
  if (!Number.isFinite(value)) {
    return DEFAULT_PASSWORD_MIN_LENGTH
  }
  const normalized = Math.trunc(Number(value))
  if (normalized < DEFAULT_PASSWORD_MIN_LENGTH) {
    return DEFAULT_PASSWORD_MIN_LENGTH
  }
  if (normalized > MAX_PASSWORD_MIN_LENGTH) {
    return MAX_PASSWORD_MIN_LENGTH
  }
  return normalized
}

export function resolvePasswordMinLength(
  settings: Pick<PublicSettings, 'password_min_length'> | null | undefined
): number {
  return normalizePasswordMinLength(settings?.password_min_length)
}
