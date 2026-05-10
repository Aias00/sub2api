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
  it('keeps the source template explicitly admin-only', () => {
    expect(componentSource).toContain("const showGithubLink = computed(() => user.value?.role === 'admin')")
    expect(componentSource).toContain('v-if="showGithubLink"')
  })

  it('keeps the embedded frontend dist aligned with the admin-only rule', () => {
    expect(appLayoutChunk).toBeTruthy()
    expect(appLayoutSource).toContain('href:"https://github.com/Wei-Shaw/sub2api"')
    expect(appLayoutSource).not.toContain('}),e("a",{href:"https://github.com/Wei-Shaw/sub2api"')
  })
})

describe('AppHeader regular user dropdown shortcuts', () => {
  it('keeps duplicate profile and api key shortcuts out of the regular user dropdown', () => {
    expect(componentSource).toContain(
      "const showDropdownAccountLinks = computed(() => user.value?.role === 'admin')"
    )
    expect(componentSource).toContain('<template v-if="showDropdownAccountLinks">')
  })

  it('collapses the regular user dropdown when only logout remains', () => {
    expect(componentSource).toContain(
      'const showDropdownPrimaryActions = computed('
    )
    expect(componentSource).toContain(
      "const compactUserDropdown = computed(() => {"
    )
    expect(componentSource).toContain('<div v-if="showDropdownPrimaryActions" class="py-1">')
    expect(componentSource).toContain("compactUserDropdown\n                    ? 'px-4 py-3'")
    expect(componentSource).toContain("compactUserDropdown\n                    ? 'py-1'")
  })

  it('uses an icon-only docs shortcut and removes the desktop balance pill', () => {
    expect(componentSource).toContain(":aria-label=\"t('nav.docs')\"")
    expect(componentSource).toContain("group-hover:opacity-100")
    expect(componentSource).not.toContain("`${{ user.balance?.toFixed(2) || '0.00' }}`")
    expect(componentSource).not.toContain("class=\"hidden items-center gap-2 rounded-xl bg-primary-50 px-3 py-1.5 dark:bg-primary-900/20 sm:flex\"")
  })
})
