import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

import {
  apiKeysShellLabelKeys,
  renderApiKeysShellText,
  resolveApiKeysShellLabels,
  type ApiKeysShellLabels,
} from '../apiKeysShell'

describe('apiKeysShell', () => {
  it('resolves API keys shell labels and filters unknown keys', () => {
    const labels = resolveApiKeysShellLabels(
      JSON.stringify({
        zh: {
          labels: {
            createKey: '配置创建 Key',
            ignored: '不应出现',
          },
        },
      }),
      'zh',
    )

    expect(labels.createKey).toBe('配置创建 Key')
    expect(labels).not.toHaveProperty('ignored')
  })

  it('renders placeholder values from configured labels', () => {
    const labels: ApiKeysShellLabels = { expiresInDays: '{days} 天后过期' }

    expect(renderApiKeysShellText(labels, 'expiresInDays', { days: 30 })).toBe(
      '30 天后过期',
    )
  })

  it('centralizes the API keys shell schema', () => {
    expect(apiKeysShellLabelKeys).toContain('createKey')
    expect(apiKeysShellLabelKeys).toContain('ccsClientSelectClaudeCode')
    expect(apiKeysShellLabelKeys).toContain('resetRateLimitUsage')
    expect(apiKeysShellLabelKeys).toContain('endpointTitle')
    expect(apiKeysShellLabelKeys).toContain('endpointClickToCopy')
    expect(apiKeysShellLabelKeys).toContain('useKeyModalTitle')
    expect(apiKeysShellLabelKeys).toContain('useKeyModalCliCodexCliWs')
  })

  it('keeps endpoint popover copy delegated to shell labels', () => {
    const source = readFileSync('src/components/keys/EndpointPopover.vue', 'utf8')

    expect(source).toContain("from '@/utils/apiKeysShell'")
    expect(source).toContain('renderApiKeysShellText(')
    expect(source).toContain('shellLabels?: ApiKeysShellLabels')
    expect(source).not.toContain("from 'vue-i18n'")
    expect(source).not.toContain('keys.endpoints.')
  })

  it('keeps API Keys modal copy delegated to shell labels', () => {
    const source = readFileSync('src/components/keys/UseKeyModal.vue', 'utf8')

    expect(source).toContain("from '@/utils/apiKeysShell'")
    expect(source).toContain('renderApiKeysShellText(')
    expect(source).not.toContain('USE_KEY_MODAL_I18N_KEYS')
    expect(source).not.toContain("from 'vue-i18n'")
    expect(source).not.toContain('shellLabels?: Record<string, string>')
  })

  it('keeps KeysView shell labels typed through apiKeysShell', () => {
    const source = readFileSync('src/views/user/KeysView.vue', 'utf8')

    expect(source).toContain('type ApiKeysShellLabels')
    expect(source).toContain('type ApiKeysShellLabelKey')
    expect(source).not.toContain('computed<Record<string, string>>')
    expect(source).not.toContain('const apiKeysText = (key: string')
  })
})
