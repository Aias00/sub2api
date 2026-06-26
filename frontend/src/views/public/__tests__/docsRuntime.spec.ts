import { describe, expect, it } from 'vitest'

import {
  buildDocsSearchNamespace,
  getDocsHashPath,
  normalizeDocsNamespacePart,
  resolveInitialDocsHash,
  withDocsContentVersion,
} from '../docsRuntime'

describe('docsRuntime', () => {
  it('normalizes namespace parts and builds namespace', () => {
    expect(normalizeDocsNamespacePart(' Cloud Base ')).toBe('cloud-base')
    expect(buildDocsSearchNamespace(['Cloud Base', 'zh-CN', 'v1.0.0'])).toBe('cloud-base-zh-cn-v1-0-0')
  })

  it('adds docs content version only to docs hashes', () => {
    expect(withDocsContentVersion('#/guide', 'v1')).toBe('#/guide?_docs_v=v1')
    expect(withDocsContentVersion('#/guide?x=1', 'v1')).toBe('#/guide?x=1&_docs_v=v1')
    expect(withDocsContentVersion('/docs', 'v1')).toBe('#/docs?_docs_v=v1')
  })

  it('resolves hash path and initial docs hash safely', () => {
    expect(getDocsHashPath('#/guide/intro?x=1')).toBe('/guide/intro')
    expect(resolveInitialDocsHash('/docs', '#/guide', '/fallback')).toBe('#/guide')
    expect(resolveInitialDocsHash('/pricing', '#/guide', '/fallback')).toBe('/fallback')
  })
})
