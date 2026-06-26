import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

const channelsViewSource = readFileSync(resolve(process.cwd(), 'src/views/admin/ChannelsView.vue'), 'utf8')

describe('ChannelsView source defaults', () => {
  it('does not synthesize missing pricing platforms as Anthropic', () => {
    expect(channelsViewSource).not.toContain("p.platform || 'anthropic'")
    expect(channelsViewSource).not.toContain('p.platform || "anthropic"')
    expect(channelsViewSource).toContain('p.platform === platform')
  })
})
