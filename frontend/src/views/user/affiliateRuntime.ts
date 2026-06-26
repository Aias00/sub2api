import type { OrderStatus } from '@/types/payment'
import type { AffiliateLabelKey } from '@/utils/affiliateShell'

export type AffiliateTextGetter = (
  key: AffiliateLabelKey,
  values?: Record<string, string | number>,
) => string

export function formatAffiliateCount(value: number): string {
  return value.toLocaleString()
}

export function formatAffiliateNullableCurrency(
  value: number | null | undefined,
  formatCurrency: (value: number) => string,
): string {
  return typeof value === 'number' ? formatCurrency(value) : '-'
}

export function formatAffiliatePaymentType(
  paymentType: string,
  hasTranslation: (key: string) => boolean,
  translate: (key: string) => string,
): string {
  const key = `payment.methods.${paymentType}`
  return hasTranslation(key) ? translate(key) : paymentType || '-'
}

export function asAffiliateOrderStatus(status: string): OrderStatus {
  return status as OrderStatus
}

export function changeAffiliatePage(
  page: number,
  setPage: (page: number) => void,
  reload: () => Promise<void> | void,
) {
  setPage(page)
  void reload()
}

export function changeAffiliatePageSize(
  pageSize: number,
  setPageSize: (size: number) => void,
  resetPage: () => void,
  reload: () => Promise<void> | void,
) {
  setPageSize(pageSize)
  resetPage()
  void reload()
}
