import { readFileSync, readdirSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const testsDir = dirname(fileURLToPath(import.meta.url))
const componentPath = resolve(testsDir, '../AppHeader.vue')
const componentSource = readFileSync(componentPath, 'utf8')
const distAssetsDir = resolve(testsDir, '../../../../../backend/internal/web/dist/assets')
const appLayoutChunk = readdirSync(distAssetsDir).find((name) =>
  name.startsWith('AppLayout.vue_vue_type_script_setup_true_lang-') && name.endsWith('.js')
)
const appLayoutSource = appLayoutChunk
  ? readFileSync(resolve(distAssetsDir, appLayoutChunk), 'utf8')
  : ''

describe('AppHeader GitHub dropdown visibility', () => {
  it('keeps the source template GitHub shortcut disabled', () => {
    expect(componentSource).toContain('const showGithubLink = computed(() => false)')
    expect(componentSource).toContain('v-if="showGithubLink"')
  })

  it('keeps the embedded frontend dist aligned with the admin-only rule', () => {
    expect(appLayoutChunk).toBeTruthy()
    expect(appLayoutSource).toContain('href:"https://github.com/Wei-Shaw/sub2api"')
    expect(appLayoutSource).not.toContain('}),e("a",{href:"https://github.com/Wei-Shaw/sub2api"')
  })
})

describe('AppHeader regular user dropdown shortcuts', () => {
  it('keeps duplicate profile and api key shortcuts disabled in the dropdown', () => {
    expect(componentSource).toContain(
      'const showDropdownAccountLinks = computed(() => false)'
    )
    expect(componentSource).toContain('<template v-if="showDropdownAccountLinks">')
  })

  it('keeps the unified task list shortcut in the regular user dropdown', () => {
    expect(componentSource).toContain(
      'const showDropdownPrimaryActions = computed('
    )
    expect(componentSource).toContain('<router-link to="/tasks" @click="closeDropdown" class="dropdown-item">')
    expect(componentSource).toContain("{{ t('nav.myTasks') }}")
    expect(componentSource).toContain('const showDropdownTaskLink = computed(() => true)')
    expect(componentSource).toContain('showDropdownTaskLink.value || showDropdownAccountLinks.value || showGithubLink.value')
    expect(componentSource).toContain("from './appHeaderRuntime'")
    expect(componentSource).toContain('resolveCompactUserDropdown')
    expect(componentSource).toContain('<div v-if="showDropdownPrimaryActions" class="py-1">')
    expect(componentSource).toContain("compactUserDropdown\n                    ? 'px-4 py-3'")
    expect(componentSource).toContain("compactUserDropdown\n                    ? 'py-1'")
  })

  it('uses an icon-only docs shortcut and removes the desktop balance pill', () => {
    expect(componentSource).toContain(":aria-label=\"t('nav.docs')\"")
    expect(componentSource).toContain("group-hover:opacity-100")
    expect(componentSource).toContain('whitespace-nowrap rounded-lg bg-gray-950')
    expect(componentSource).not.toContain("`${{ user.balance?.toFixed(2) || '0.00' }}`")
    expect(componentSource).not.toContain("class=\"hidden items-center gap-2 rounded-xl bg-primary-50 px-3 py-1.5 dark:bg-primary-900/20 sm:flex\"")
  })

  it('does not render a secondary page description line under the title', () => {
    expect(componentSource).not.toContain('pageDescription')
    expect(componentSource).not.toContain("class=\"text-xs text-gray-500 dark:text-slate-300/90\"")
  })

  it('delegates header display-name and title shaping to shared runtime helpers', () => {
    expect(componentSource).toContain('resolveHeaderDisplayName')
    expect(componentSource).toContain('resolveHeaderUserInitials')
    expect(componentSource).toContain('resolveHeaderPageTitle')
    expect(componentSource).not.toContain("return user.value.username || user.value.email?.split('@')[0] || ''")
  })

  it('uses auth route defaults for logout redirect instead of a local login path', () => {
    expect(componentSource).toContain('useAuthRouteDefaults')
    expect(componentSource).toContain('router.push(authRouteDefaults.value.loginPath)')
    expect(componentSource).not.toContain("router.push('/login')")
  })

  it('uses auth route defaults for account dropdown shortcuts', () => {
    expect(componentSource).toContain(':to="authRouteDefaults.profilePath"')
    expect(componentSource).toContain(':to="authRouteDefaults.apiKeysPath"')
    expect(componentSource).not.toContain('to="/profile"')
    expect(componentSource).not.toContain('to="/keys"')
  })
})
