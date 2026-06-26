import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

const logoSurfaces = [
  'src/views/HomeView.vue',
  'src/views/KeyUsageView.vue',
  'src/views/public/ImageGeneratorView.vue',
  'src/views/public/ModelsPlazaView.vue',
  'src/views/public/LegalDocumentView.vue',
  'src/views/public/PricingView.vue',
  'src/views/public/PromptCatalogView.vue',
  'src/components/layout/AppSidebar.vue',
  'src/components/layout/AuthLayout.vue',
] as const

describe('site logo runtime fallback', () => {
  it('does not render a bundled favicon as the configured site logo fallback', () => {
    for (const file of logoSurfaces) {
      const source = readFileSync(resolve(process.cwd(), file), 'utf8')

      expect(source, file).not.toContain("siteLogo || '/favicon.svg'")
      expect(source, file).not.toContain('siteLogo || "/favicon.svg"')
      expect(source, file).not.toContain(':src="siteLogo ||')
    }
  })
})
