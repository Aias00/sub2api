import { describe, expect, it } from 'vitest'

import {
  renderLegalDocumentHtml,
  resolveCurrentLegalDocument,
  resolveLegalDocumentIcon,
} from '../legalDocumentRuntime'
import type { LoginAgreementDocument, PublicSettings } from '@/types'

describe('legalDocumentRuntime', () => {
  it('resolves current document safely', () => {
    const docs: LoginAgreementDocument[] = [
      { id: 'terms', title: 'Terms', content_md: '' },
      { id: 'privacy', title: 'Privacy Policy', content_md: '' },
    ]

    expect(resolveCurrentLegalDocument(docs, 'privacy')?.title).toBe('Privacy Policy')
    expect(resolveCurrentLegalDocument(docs, '')).toBeNull()
  })

  it('resolves legal document icon from title semantics', () => {
    expect(resolveLegalDocumentIcon('Privacy Policy')).toBe('shield')
    expect(resolveLegalDocumentIcon('Country Addendum')).toBe('globe')
    expect(resolveLegalDocumentIcon('Specific Terms')).toBe('cog')
    expect(resolveLegalDocumentIcon('General Terms')).toBe('document')
  })

  it('renders sanitized legal document html', () => {
    const document: LoginAgreementDocument = {
      id: 'terms',
      title: 'Terms',
      content_md: '# Hello\n\n<script>alert(1)</script>\n\nContact us.',
    }
    const settings = {
      login_agreement_updated_at: '2026-06-18',
      contact_info: 'support@example.com',
    } as PublicSettings

    const html = renderLegalDocumentHtml(document, settings)
    expect(html).toContain('<h1')
    expect(html).toContain('Hello')
    expect(html).not.toContain('<script>')
  })
})
