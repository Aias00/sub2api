import { readdirSync, readFileSync, statSync } from 'node:fs'
import { join, resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

const docsContentRoot = resolve(process.cwd(), 'public/docs-content')

function collectMarkdownFiles(dir: string): string[] {
  return readdirSync(dir).flatMap((entry) => {
    const path = join(dir, entry)
    return statSync(path).isDirectory() ? collectMarkdownFiles(path) : path.endsWith('.md') ? [path] : []
  })
}

describe('public docs content', () => {
  it('does not hardcode the production domain in markdown files', () => {
    const offenders = collectMarkdownFiles(docsContentRoot).filter((file) =>
      readFileSync(file, 'utf8').includes('cloudbase.eu.org')
    )

    expect(offenders).toEqual([])
  })
})
