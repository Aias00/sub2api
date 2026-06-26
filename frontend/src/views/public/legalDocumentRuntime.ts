import DOMPurify from 'dompurify'
import { marked } from 'marked'

import { renderLoginAgreementDocumentContent } from '@/utils/loginAgreementTemplates'
import type { LoginAgreementDocument, PublicSettings } from '@/types'

export type LegalDocumentIcon = 'document' | 'shield' | 'globe' | 'cog'

export function resolveCurrentLegalDocument(
  documents: LoginAgreementDocument[],
  documentId: string,
): LoginAgreementDocument | null {
  if (!documentId) return null
  return documents.find((doc) => doc.id === documentId) ?? null
}

export function resolveLegalDocumentIcon(title: string): LegalDocumentIcon {
  const lowerTitle = title.toLowerCase()
  if (title.includes('政策') || title.includes('隐私') || lowerTitle.includes('policy') || lowerTitle.includes('privacy')) {
    return 'shield'
  }
  if (title.includes('国家') || title.includes('地区') || lowerTitle.includes('country') || lowerTitle.includes('region')) {
    return 'globe'
  }
  if (title.includes('特定') || lowerTitle.includes('specific')) {
    return 'cog'
  }
  return 'document'
}

export function renderLegalDocumentHtml(
  document: LoginAgreementDocument | null,
  settings: PublicSettings | null,
): string {
  const content = renderLoginAgreementDocumentContent(
    document?.content_md?.trim() || '',
    {
      documentId: document?.id,
      updatedAt: settings?.login_agreement_updated_at || '',
      frontendUrl: '',
      contactInfo: settings?.contact_info || '',
    },
  )
  if (!content) {
    return ''
  }
  const html = marked.parse(content) as string
  return DOMPurify.sanitize(html)
}
