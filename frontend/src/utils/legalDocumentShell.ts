import { resolveLocalizedShellLabels } from './localizedShell'

export type LegalDocumentCopy = {
  login: string
  agreementLabel: string
  loadFailedTitle: string
  loadFailedDescription: string
  missingTitle: string
  missingDescription: string
  updatedAt: string
  emptyContent: string
}

const legalDocumentCopyKeys = [
  'login',
  'agreementLabel',
  'loadFailedTitle',
  'loadFailedDescription',
  'missingTitle',
  'missingDescription',
  'updatedAt',
  'emptyContent',
] as const satisfies readonly (keyof LegalDocumentCopy)[]

export function resolveLegalDocumentCopy(raw: string | undefined, runtimeLocale: string): LegalDocumentCopy {
  return resolveLocalizedShellLabels(raw, runtimeLocale, legalDocumentCopyKeys)
}

export function formatLegalDocumentTemplate(template: string, values: Record<string, string>): string {
  return template.replace(/\{(\w+)\}/g, (_match, key: string) => values[key] ?? '')
}
