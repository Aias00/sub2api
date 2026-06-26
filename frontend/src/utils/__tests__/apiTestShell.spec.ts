import { describe, expect, it } from 'vitest'
import {
  DEFAULT_API_TEST_API_KEY_PAGE_SIZE,
  DEFAULT_API_TEST_GUIDE_PATH,
  DEFAULT_API_TEST_MAX_TOKENS,
  DEFAULT_API_TEST_USAGE_SYNC_PAGE_SIZE,
  apiTestLabelKeys,
  renderAPITestShellText,
  resolveAPITestShellDefaults,
  resolveAPITestShellLabels,
} from '../apiTestShell'
import { DEFAULT_GATEWAY_TEST_PROMPT } from '../gatewayDocs'

describe('api test shell helpers', () => {
  it('resolves api test labels from localized shell config', () => {
    const labels = resolveAPITestShellLabels(
      JSON.stringify({
        en: {
          labels: {
            title: 'API Test',
            usageRecordFound: 'Found {id}',
            ignored: 'ignored',
          },
        },
      }),
      'en-US',
    )

    expect(apiTestLabelKeys).toContain('loading')
    expect(apiTestLabelKeys).toContain('noOptionsFound')
    expect(apiTestLabelKeys).toContain('usageRecordIdle')
    expect(apiTestLabelKeys).toContain('unknownError')
    expect(labels.title).toBe('API Test')
    expect(labels.usageRecordFound).toBe('Found {id}')
    expect(labels.badge).toBeUndefined()
  })

  it('renders api test labels with placeholders', () => {
    const labels = {
      usageRecordFound: 'Found {id}',
      title: 'API Test',
    }

    expect(renderAPITestShellText(labels, 'usageRecordFound', { id: 42 })).toBe('Found 42')
    expect(renderAPITestShellText(labels, 'title')).toBe('API Test')
    expect(renderAPITestShellText(labels, 'loadKeysFailed')).toBe('')
  })

  it('resolves guide path defaults from localized shell config', () => {
    expect(resolveAPITestShellDefaults(JSON.stringify({
      en: {
        defaults: {
          guidePath: '/configured-gateway-guide',
          defaultPrompt: 'Configured test prompt',
          maxTokens: 512,
          apiKeyPageSize: 33,
          usageSyncPageSize: 7,
        },
      },
    }), 'en-US')).toEqual({
      guidePath: '/configured-gateway-guide',
      defaultPrompt: 'Configured test prompt',
      maxTokens: 512,
      apiKeyPageSize: 33,
      usageSyncPageSize: 7,
    })

    expect(resolveAPITestShellDefaults(undefined, 'en')).toEqual({
      guidePath: DEFAULT_API_TEST_GUIDE_PATH,
      defaultPrompt: DEFAULT_GATEWAY_TEST_PROMPT,
      maxTokens: DEFAULT_API_TEST_MAX_TOKENS,
      apiKeyPageSize: DEFAULT_API_TEST_API_KEY_PAGE_SIZE,
      usageSyncPageSize: DEFAULT_API_TEST_USAGE_SYNC_PAGE_SIZE,
    })
    expect(resolveAPITestShellDefaults(JSON.stringify({
      en: {
        defaults: {
          guidePath: '//evil.example/gateway-guide',
          defaultPrompt: '   ',
          maxTokens: 0,
          apiKeyPageSize: 1001,
          usageSyncPageSize: 0,
        },
      },
    }), 'en')).toEqual({
      guidePath: DEFAULT_API_TEST_GUIDE_PATH,
      defaultPrompt: DEFAULT_GATEWAY_TEST_PROMPT,
      maxTokens: DEFAULT_API_TEST_MAX_TOKENS,
      apiKeyPageSize: DEFAULT_API_TEST_API_KEY_PAGE_SIZE,
      usageSyncPageSize: DEFAULT_API_TEST_USAGE_SYNC_PAGE_SIZE,
    })
  })
})
