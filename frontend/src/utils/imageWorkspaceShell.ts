export interface WorkspaceShellCopy {
  catalogLabel: string
  eyebrow: string
  title: string
  heroDescription: string
  draftImported: string
  draftImportedDescription: string
  promptLabel: string
  promptPlaceholder: string
  promptTooLong: string
  clearLabel: string
  copyPromptLabel: string
  copySuccessMessage: string
  copyEmptyError: string
  workspaceTitle: string
  workspaceDescription: string
  workspaceStatus: string
  backToCatalogLabel: string
}

export interface WorkspaceShellDefaults {
  catalogPath: string
  maxPromptLength: number
}

export const DEFAULT_WORKSPACE_CATALOG_PATH = ''
export const DEFAULT_WORKSPACE_MAX_PROMPT_LENGTH = 2000

const workspaceShellKeys: Array<keyof WorkspaceShellCopy> = [
  'catalogLabel',
  'eyebrow',
  'title',
  'heroDescription',
  'draftImported',
  'draftImportedDescription',
  'promptLabel',
  'promptPlaceholder',
  'promptTooLong',
  'clearLabel',
  'copyPromptLabel',
  'copySuccessMessage',
  'copyEmptyError',
  'workspaceTitle',
  'workspaceDescription',
  'workspaceStatus',
  'backToCatalogLabel',
]

export function resolveWorkspaceShellConfig(
  raw: string | undefined,
  activeLocale: 'zh' | 'en',
): Partial<WorkspaceShellCopy> {
  if (!raw?.trim()) return {}
  try {
    const parsed = JSON.parse(raw) as unknown
    if (!isRecord(parsed)) return {}
    const scoped = parsed[activeLocale] || parsed.default || parsed
    if (!isRecord(scoped)) return {}

    return readWorkspaceShellCopy(scoped)
  } catch {
    return {}
  }
}

export function resolveWorkspaceShellDefaults(
  raw: string | undefined,
  activeLocale: 'zh' | 'en',
): WorkspaceShellDefaults {
  const scoped = readWorkspaceShellScope(raw, activeLocale)
  if (!scoped || !isRecord(scoped.defaults)) {
    return {
      catalogPath: DEFAULT_WORKSPACE_CATALOG_PATH,
      maxPromptLength: DEFAULT_WORKSPACE_MAX_PROMPT_LENGTH,
    }
  }
  const catalogPath = readInternalPath(scoped.defaults.catalogPath)
  const maxPromptLength = readPositiveInteger(scoped.defaults.maxPromptLength)
  return {
    catalogPath: catalogPath || DEFAULT_WORKSPACE_CATALOG_PATH,
    maxPromptLength: maxPromptLength || DEFAULT_WORKSPACE_MAX_PROMPT_LENGTH,
  }
}

export function formatWorkspaceShellTemplate(template: string, values: Record<string, string>): string {
  return template.replace(/\{(\w+)\}/g, (_match, key: string) => values[key] ?? '')
}

function readWorkspaceShellCopy(value: Record<string, unknown>): Partial<WorkspaceShellCopy> {
  const copy: Partial<WorkspaceShellCopy> = {}
  for (const key of workspaceShellKeys) {
    const label = readString(value[key])
    if (label) copy[key] = label
  }
  return copy
}

function readWorkspaceShellScope(raw: string | undefined, activeLocale: 'zh' | 'en'): Record<string, unknown> | null {
  if (!raw?.trim()) return null
  try {
    const parsed = JSON.parse(raw) as unknown
    if (!isRecord(parsed)) return null
    const scoped = parsed[activeLocale] || parsed.default || parsed
    return isRecord(scoped) ? scoped : null
  } catch {
    return null
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

function readString(value: unknown): string | undefined {
  return typeof value === 'string' ? value : undefined
}

function readInternalPath(value: unknown): string | undefined {
  const path = readString(value)?.trim()
  if (!path || !path.startsWith('/') || path.startsWith('//')) return undefined
  if (path.includes('://') || path.includes('\n') || path.includes('\r')) return undefined
  return path
}

function readPositiveInteger(value: unknown): number | undefined {
  if (typeof value !== 'number' || !Number.isInteger(value) || value < 1) {
    return undefined
  }
  return value
}
