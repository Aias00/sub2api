import { describe, expect, it } from 'vitest'
import {
  DEFAULT_API_GUIDE_API_KEY_PAGE_SIZE,
  DEFAULT_API_GUIDE_MAX_TOKENS,
  DEFAULT_API_GUIDE_PROMPT,
  DEFAULT_API_GUIDE_TEST_PATH,
  apiGuideLabelKeys,
  renderAPIGuideShellText,
  resolveAPIGuideShellDefaults,
  resolveAPIGuideShellLabels,
} from '../apiGuideShell'

describe('api guide shell helpers', () => {
  it('resolves api guide labels from localized shell config', () => {
    const labels = resolveAPIGuideShellLabels(
      JSON.stringify({
        en: {
          labels: {
            title: 'API Guide',
            stream: 'Stream',
            copyCurlSuccess: 'Copied',
            ignored: 'ignored',
          },
        },
      }),
      'en-US',
    )

    expect(apiGuideLabelKeys).toContain('status')
    expect(apiGuideLabelKeys).toContain('stream')
    expect(apiGuideLabelKeys).toContain('loadKeysFailed')
    expect(labels.title).toBe('API Guide')
    expect(labels.stream).toBe('Stream')
    expect(labels.copyCurlSuccess).toBe('Copied')
    expect(labels.badge).toBeUndefined()
  })

  it('renders empty text for missing api guide labels', () => {
    expect(renderAPIGuideShellText({ title: 'Guide' }, 'title')).toBe('Guide')
    expect(renderAPIGuideShellText({}, 'loadKeysFailed')).toBe('')
  })

  it('resolves tester path defaults from localized shell config', () => {
    expect(resolveAPIGuideShellDefaults(JSON.stringify({
      en: {
        defaults: {
          testPath: '/configured-gateway-test',
          defaultPrompt: 'Configured guide prompt',
          maxTokens: 512,
          apiKeyPageSize: 33,
        },
      },
    }), 'en-US')).toEqual({
      testPath: '/configured-gateway-test',
      defaultPrompt: 'Configured guide prompt',
      maxTokens: 512,
      apiKeyPageSize: 33,
    })

    expect(resolveAPIGuideShellDefaults(undefined, 'en')).toEqual({
      testPath: DEFAULT_API_GUIDE_TEST_PATH,
      defaultPrompt: DEFAULT_API_GUIDE_PROMPT,
      maxTokens: DEFAULT_API_GUIDE_MAX_TOKENS,
      apiKeyPageSize: DEFAULT_API_GUIDE_API_KEY_PAGE_SIZE,
    })
    expect(resolveAPIGuideShellDefaults(JSON.stringify({
      en: {
        defaults: {
          testPath: 'https://evil.example/gateway-test',
          defaultPrompt: '   ',
          maxTokens: 0,
          apiKeyPageSize: 1001,
        },
      },
    }), 'en')).toEqual({
      testPath: DEFAULT_API_GUIDE_TEST_PATH,
      defaultPrompt: DEFAULT_API_GUIDE_PROMPT,
      maxTokens: DEFAULT_API_GUIDE_MAX_TOKENS,
      apiKeyPageSize: DEFAULT_API_GUIDE_API_KEY_PAGE_SIZE,
    })
  })
})
