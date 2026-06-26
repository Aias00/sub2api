import { describe, expect, it } from 'vitest'
import {
  PAYMENT_CURRENCY_OPTIONS,
  PROVIDER_CONFIG_FIELDS,
  buildProviderCallbackPaths,
} from '@/components/payment/providerConfig'

function findField(providerKey: string, key: string) {
  const fields = PROVIDER_CONFIG_FIELDS[providerKey] || []
  return fields.find(field => field.key === key)
}

describe('PROVIDER_CONFIG_FIELDS.wxpay', () => {
  it('keeps admin form validation aligned with backend-required credentials', () => {
    expect(findField('wxpay', 'publicKeyId')?.optional).toBeFalsy()
    expect(findField('wxpay', 'certSerial')?.optional).toBeFalsy()
  })

  it('only keeps the simplified visible credential set in the admin form', () => {
    expect(findField('wxpay', 'mpAppId')).toBeUndefined()
    expect(findField('wxpay', 'h5AppName')).toBeUndefined()
    expect(findField('wxpay', 'h5AppUrl')).toBeUndefined()
  })
})

describe('PROVIDER_CONFIG_FIELDS.airwallex', () => {
  it('adds currency config without a frontend default', () => {
    const currency = findField('airwallex', 'currency')

    expect('defaultValue' in (currency || {})).toBe(false)
    expect(currency?.hintKey).toBe('admin.settings.payment.field_paymentCurrencyHint')
    expect(currency?.options).toBe(PAYMENT_CURRENCY_OPTIONS)
  })

  it('marks accountId as optional and explains when it can be left blank', () => {
    const accountId = findField('airwallex', 'accountId')

    expect(accountId?.optional).toBe(true)
    expect(accountId?.clearable).toBe(true)
    expect(accountId?.hintKey).toBe('admin.settings.payment.field_accountIdHint')
  })

  it('explains that apiBase must match the Airwallex key environment', () => {
    expect(findField('airwallex', 'apiBase')?.hintKey).toBe('admin.settings.payment.field_airwallexApiBaseHint')
  })
})

describe('PROVIDER_CONFIG_FIELDS.stripe', () => {
  it('adds currency config without a frontend default', () => {
    const currency = findField('stripe', 'currency')

    expect('defaultValue' in (currency || {})).toBe(false)
    expect(currency?.hintKey).toBe('admin.settings.payment.field_paymentCurrencyHint')
    expect(currency?.options).toBe(PAYMENT_CURRENCY_OPTIONS)
  })
})

describe('buildProviderCallbackPaths', () => {
  it('uses the configured payment result path for provider return URLs', () => {
    const paths = buildProviderCallbackPaths('/configured-payment-result')

    expect(paths.easypay.returnUrl).toBe('/configured-payment-result')
    expect(paths.alipay.returnUrl).toBe('/configured-payment-result')
    expect(paths.creem.returnUrl).toBe('/configured-payment-result')
    expect(paths.waffo.returnUrl).toBe('/configured-payment-result')
    expect(paths.wxpay.returnUrl).toBeUndefined()
  })

  it('falls back to the built-in payment result path for unsafe return paths', () => {
    expect(buildProviderCallbackPaths('https://evil.example/result').alipay.returnUrl).toBe('/payment/result')
    expect(buildProviderCallbackPaths('//evil.example/result').alipay.returnUrl).toBe('/payment/result')
  })
})
