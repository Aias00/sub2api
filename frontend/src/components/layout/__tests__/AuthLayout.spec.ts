import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

const source = readFileSync(resolve(process.cwd(), 'src/components/layout/AuthLayout.vue'), 'utf8')

describe('AuthLayout brand block', () => {
  it('shows the site name without the subtitle line', () => {
    expect(source).toContain('{{ siteName }}')
    expect(source).not.toContain('{{ siteSubtitle }}')
    expect(source).not.toContain('Subscription to API Conversion Platform')
  })
})
