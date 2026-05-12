import { describe, expect, it } from 'vitest'

import { resolveCSPNonce } from '../cspNonce'

describe('resolveCSPNonce', () => {
  it('returns the nonce from the first script tag that exposes it', () => {
    const doc = document.implementation.createHTMLDocument('nonce')
    const script = doc.createElement('script')
    script.nonce = 'nonce-123'
    doc.head.appendChild(script)

    expect(resolveCSPNonce(doc)).toBe('nonce-123')
  })

  it('falls back to a csp-nonce meta tag when no script nonce exists', () => {
    const doc = document.implementation.createHTMLDocument('nonce')
    const meta = doc.createElement('meta')
    meta.setAttribute('name', 'csp-nonce')
    meta.setAttribute('content', 'meta-nonce')
    doc.head.appendChild(meta)

    expect(resolveCSPNonce(doc)).toBe('meta-nonce')
  })

  it('returns an empty string when the document has no nonce source', () => {
    const doc = document.implementation.createHTMLDocument('nonce')
    expect(resolveCSPNonce(doc)).toBe('')
  })
})
