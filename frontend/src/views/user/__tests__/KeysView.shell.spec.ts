import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

const source = readFileSync(resolve(process.cwd(), 'src/views/user/KeysView.vue'), 'utf8')
const zhLocaleSource = readFileSync(resolve(process.cwd(), 'src/i18n/locales/zh.ts'), 'utf8')
const enLocaleSource = readFileSync(resolve(process.cwd(), 'src/i18n/locales/en.ts'), 'utf8')

describe('KeysView shell config', () => {
  it('uses the API keys shell helper instead of a page-local parser or schema', () => {
    expect(source).toContain("from '@/utils/apiKeysShell'")
    expect(source).toContain('resolveApiKeysShellLabels(')
    expect(source).toContain('renderApiKeysShellText(')
    expect(source).toContain('type ApiKeysShellLabels')
    expect(source).toContain('type ApiKeysShellLabelKey')
    expect(source).not.toContain("import { resolveShellLabelOverrides } from '@/utils/shellLabelOverrides'")
    expect(source).not.toContain('resolveShellLabelOverrides(')
    expect(source).not.toContain('const apiKeysShellLabelKeys')
    expect(source).not.toContain('function readLocalizedShellLabels')
    expect(source).not.toContain('function isRecord')
  })

  it('does not keep API Keys shell i18n fallback keys in the view bootstrap layer', () => {
    expect(source).not.toContain('API_KEYS_I18N_KEYS')
    expect(source).not.toContain('apiKeysShellLabels.value[key] || key')
    expect(source).not.toContain('keys.failedToLoad')
    expect(source).not.toContain('common.actions')
  })

  it('does not keep the legacy API keys locale section in frontend bundles', () => {
    expect(zhLocaleSource).not.toContain('\n  keys: {')
    expect(enLocaleSource).not.toContain('\n  keys: {')
  })

  it('passes API Keys shell labels to endpoint popover instead of localizing it separately', () => {
    expect(source).toContain(':shell-labels="apiKeysShellLabels"')
    expect(source).not.toContain('keys.endpoints.')
  })

  it('does not invent a local provider name for CCS import links', () => {
    expect(source).not.toContain("|| 'cloudbase'")
    expect(source).not.toContain('|| "cloudbase"')
    expect(source).toContain("publicSettings.value?.site_name?.trim() || ''")
  })

  it('does not invent an Anthropic platform for CCS import links', () => {
    expect(source).not.toContain("row.group?.platform || 'anthropic'")
    expect(source).not.toContain('row.group?.platform || "anthropic"')
  })

  it('does not invent a USD usage unit for CCS import links', () => {
    expect(source).not.toContain('response?.quota?.unit ?? "USD"')
    expect(source).not.toContain("response?.quota?.unit ?? 'USD'")
  })

  it('does not hard-code a dollar prefix for quota and usage amounts', () => {
    expect(source).not.toContain('${{')
    expect(source).not.toContain('>${{')
    expect(source).not.toContain('>$</span>')
    expect(source).toContain('pricing_currency_symbol')
    expect(source).toContain('formatApiKeyCost(')
    expect(source).toContain('apiKeyCurrencyPrefix')
  })
})
