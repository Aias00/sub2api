import { describe, expect, it } from 'vitest'
import {
  DEFAULT_WORKSPACE_CATALOG_PATH,
  DEFAULT_WORKSPACE_MAX_PROMPT_LENGTH,
  formatWorkspaceShellTemplate,
  resolveWorkspaceShellConfig,
  resolveWorkspaceShellDefaults,
  type WorkspaceShellCopy,
} from '../imageWorkspaceShell'

const configuredShell: WorkspaceShellCopy = {
  catalogLabel: 'Prompt catalog',
  eyebrow: 'Prompt Workspace',
  title: 'AI Image Workspace',
  heroDescription: 'Prepare prompts.',
  draftImported: 'Imported "{title}"',
  draftImportedDescription: 'Prompt is ready.',
  promptLabel: 'Prompt',
  promptPlaceholder: 'Enter a prompt',
  promptTooLong: 'Prompt is too long',
  clearLabel: 'Clear',
  copyPromptLabel: 'Copy prompt',
  copySuccessMessage: 'Prompt copied',
  copyEmptyError: 'Enter a prompt first',
  workspaceTitle: 'Workspace status',
  workspaceDescription: 'Description',
  workspaceStatus: 'Status',
  backToCatalogLabel: 'Back to catalog',
}

describe('resolveWorkspaceShellConfig', () => {
  it('resolves locale-scoped root fields', () => {
    const shell = resolveWorkspaceShellConfig(
      JSON.stringify({
        en: {
          title: 'Configured workspace',
          copyPromptLabel: 'Configured copy',
          ignored: 'ignored',
        },
      }),
      'en',
    )

    expect(shell.title).toBe('Configured workspace')
    expect(shell.copyPromptLabel).toBe('Configured copy')
    expect(shell.promptLabel).toBeUndefined()
  })

  it('falls back to default scope when locale scope is missing', () => {
    const shell = resolveWorkspaceShellConfig(
      JSON.stringify({
        default: {
          promptPlaceholder: 'Default placeholder',
        },
      }),
      'zh',
    )

    expect(shell.promptPlaceholder).toBe('Default placeholder')
  })

  it('returns empty partial copy for missing or invalid public settings config', () => {
    expect(resolveWorkspaceShellConfig(undefined, 'en')).toEqual({})
    expect(resolveWorkspaceShellConfig('{bad json', 'en')).toEqual({})
  })

  it('does not synthesize blank labels that override caller defaults', () => {
    const shell = resolveWorkspaceShellConfig(JSON.stringify({ en: { title: configuredShell.title } }), 'en')

    expect(shell.title).toBe(configuredShell.title)
    expect(shell.catalogLabel).toBeUndefined()
    expect(shell.clearLabel).toBeUndefined()
    expect(shell.copyPromptLabel).toBeUndefined()
  })

  it('formats shell templates without page-local interpolation helpers', () => {
    expect(formatWorkspaceShellTemplate('Imported "{title}"', { title: 'Toy Portrait' })).toBe('Imported "Toy Portrait"')
    expect(formatWorkspaceShellTemplate('Imported "{missing}"', {})).toBe('Imported ""')
  })

  it('resolves configured workspace defaults and falls back to the shared catalog path', () => {
    expect(resolveWorkspaceShellDefaults(JSON.stringify({
      en: {
        defaults: {
          catalogPath: '/configured-prompts',
          maxPromptLength: 1200,
        },
      },
    }), 'en')).toEqual({
      catalogPath: '/configured-prompts',
      maxPromptLength: 1200,
    })

    expect(resolveWorkspaceShellDefaults(undefined, 'en')).toEqual({
      catalogPath: DEFAULT_WORKSPACE_CATALOG_PATH,
      maxPromptLength: DEFAULT_WORKSPACE_MAX_PROMPT_LENGTH,
    })

    expect(resolveWorkspaceShellDefaults(JSON.stringify({
      en: {
        defaults: {
          catalogPath: 'https://evil.example/prompts',
          maxPromptLength: 0,
        },
      },
    }), 'en')).toEqual({
      catalogPath: DEFAULT_WORKSPACE_CATALOG_PATH,
      maxPromptLength: DEFAULT_WORKSPACE_MAX_PROMPT_LENGTH,
    })
  })
})
