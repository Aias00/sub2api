import type { PaymentOrder } from '@/types/payment'

const currencySymbols: Record<string, string> = {
  CNY: '¥',
  RMB: '¥',
  USD: '$',
  EUR: '€',
  GBP: '£',
  JPY: '¥',
  HKD: 'HK$',
}

export function formatPaymentCurrencyAmount(value: number | undefined, currency?: string | null): string {
  const amount = formatAmount(value)
  return `${resolvePaymentCurrencyPrefix(currency)}${amount}`
}

export function resolvePaymentCurrencyPrefix(currency?: string | null): string {
  const code = currency?.trim().toUpperCase()
  if (!code) return ''
  return currencySymbols[code] ?? `${code} `
}

export function formatOrderPayAmount(order: PaymentOrder, value = order.pay_amount): string {
  return formatPaymentCurrencyAmount(value, order.currency)
}

export function formatOrderCreditedAmount(order: PaymentOrder, value = order.amount): string {
  if (order.order_type === 'balance') {
    return formatAmount(value)
  }
  return formatPaymentCurrencyAmount(value, order.currency)
}

export function formatPublicMoneyAmount(
  value: number | null | undefined,
  currencyPrefix: string | null | undefined,
  fractionDigits = 2,
): string {
  const amount = Number.isFinite(value) ? Number(value) : 0
  const normalizedDigits = Number.isInteger(fractionDigits) && fractionDigits >= 0 ? fractionDigits : 2
  const prefix = typeof currencyPrefix === 'string' && currencyPrefix.trim() ? currencyPrefix : ''
  return `${prefix}${amount.toFixed(normalizedDigits)}`
}

function formatAmount(value: number | undefined): string {
  const amount = Number.isFinite(value) ? Number(value) : 0
  return amount.toFixed(2)
}
