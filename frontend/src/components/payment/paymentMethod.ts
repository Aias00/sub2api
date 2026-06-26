import {
  normalizeVisibleMethod,
  type StripeVisibleMethod,
  type VisiblePaymentMethod,
} from './paymentFlow'

const STRIPE_PAYMENT_METHOD_COLORS: Record<StripeVisibleMethod, string> = {
  alipay: '#00AEEF',
  wechat_pay: '#07C160',
}

const UNKNOWN_STRIPE_PAYMENT_METHOD_COLOR = '#635bff'

export function resolveVisiblePaymentMethod(method?: string | null): VisiblePaymentMethod | '' {
  return normalizeVisibleMethod(String(method || ''))
}

export function normalizeStripePaymentMethod(method?: string | null): StripeVisibleMethod | '' {
  const normalized = String(method || '').trim()
  return normalized === 'wechat_pay' || normalized === 'alipay' ? normalized : ''
}

export function getStripePaymentMethodColor(method?: string | null): string {
  const normalized = String(method || '').trim()
  if (normalized === 'wechat_pay' || normalized === 'alipay') {
    return STRIPE_PAYMENT_METHOD_COLORS[normalized]
  }
  return UNKNOWN_STRIPE_PAYMENT_METHOD_COLOR
}
