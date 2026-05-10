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
    expect(appLayoutSource).toMatch(
      /role==="admin"\?\(o\(\),s\("a",\{[^}]*href:"https:\/\/github\.com\/Wei-Shaw\/sub2api"/
    )
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
})
