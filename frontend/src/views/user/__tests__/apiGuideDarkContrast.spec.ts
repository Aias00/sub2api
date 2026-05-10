import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

const styleSource = readFileSync(resolve(process.cwd(), 'src/style.css'), 'utf8')
const apiGuideSource = readFileSync(resolve(process.cwd(), 'src/views/user/ApiGuideView.vue'), 'utf8')

describe('API guide dark contrast tokens', () => {
  it('defines shared page surface tokens in source styles', () => {
    expect(styleSource).toContain('.page-hero {')
    expect(styleSource).toContain('.page-kicker {')
    expect(styleSource).toContain('.metric-panel {')
    expect(styleSource).toContain('.surface-panel {')
    expect(styleSource).toContain('.surface-panel-strong {')
    expect(styleSource).toContain('.surface-panel-muted {')
  })

  it('gives dark mode stronger foreground/background separation for guide surfaces', () => {
    expect(styleSource).toContain('.dark .page-hero {')
    expect(styleSource).toContain('.dark .page-kicker {')
    expect(styleSource).toContain('.dark .metric-panel {')
    expect(styleSource).toContain('.dark .surface-panel {')
    expect(styleSource).toContain('.dark .surface-panel-strong {')
    expect(styleSource).toContain('.dark .surface-panel-muted {')
  })

  it('still uses the shared surface tokens on the API guide page', () => {
    expect(apiGuideSource).toContain('class="page-hero"')
    expect(apiGuideSource).toContain('class="page-kicker"')
    expect(apiGuideSource).toContain('class="metric-panel"')
    expect(apiGuideSource).toContain('class="surface-panel-strong')
    expect(apiGuideSource).toContain('class="surface-panel')
  })
})
