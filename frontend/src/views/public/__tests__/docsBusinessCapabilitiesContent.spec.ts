import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

function readDocsContent(path: string) {
  return readFileSync(resolve(process.cwd(), '../frontend/public/docs-content', path), 'utf8')
}

describe('business capabilities docs content', () => {
  it('exposes the Chinese business capabilities page from the docs sidebar and overview', () => {
    const sidebar = readDocsContent('_sidebar.md')
    const overview = readDocsContent('README.md')
    const page = readDocsContent('business-capabilities.md')

    expect(sidebar).toContain('[业务能力](business-capabilities)')
    expect(sidebar).toContain('\n* [业务能力](business-capabilities)\n  * [能力总览]')
    expect(sidebar).not.toContain('  * [业务能力](business-capabilities)')
    expect(sidebar).toContain('[微信导出](business-capabilities?id=微信导出)')
    expect(sidebar).toContain('[热点追踪](business-capabilities?id=热点追踪)')
    expect(sidebar).toContain('[图片提示词](business-capabilities?id=图片提示词)')
    expect(sidebar).toContain('[生图工作台](business-capabilities?id=生图工作台)')
    expect(sidebar).toContain('[我的任务](business-capabilities?id=我的任务)')
    expect(sidebar).toContain('[推荐工作流](business-capabilities?id=推荐工作流)')
    expect(overview).toContain('[业务能力使用指南](business-capabilities)')
    expect(page).toContain('# 业务能力使用指南')
    expect(page).toContain('[微信导出](/wechat)')
    expect(page).toContain('[热点追踪](/hot)')
    expect(page).toContain('[图片提示词](/prompts)')
    expect(page).toContain('[生图工作台](/image-generator)')
    expect(page).toContain('[我的任务](/tasks)')
  })

  it('exposes the English business capabilities page from the docs sidebar and overview', () => {
    const sidebar = readDocsContent('en/_sidebar.md')
    const overview = readDocsContent('en/README.md')
    const page = readDocsContent('en/business-capabilities.md')

    expect(sidebar).toContain('[Business Capabilities](business-capabilities)')
    expect(sidebar).toContain('\n* [Business Capabilities](business-capabilities)\n  * [Capability Overview]')
    expect(sidebar).not.toContain('  * [Business Capabilities](business-capabilities)')
    expect(sidebar).toContain('[WeChat Export](business-capabilities?id=wechat-export)')
    expect(sidebar).toContain('[Hot Topics](business-capabilities?id=hot-topics)')
    expect(sidebar).toContain('[Image Prompts](business-capabilities?id=image-prompts)')
    expect(sidebar).toContain('[Image Workspace](business-capabilities?id=image-workspace)')
    expect(sidebar).toContain('[My Tasks](business-capabilities?id=my-tasks)')
    expect(sidebar).toContain('[Recommended Workflows](business-capabilities?id=recommended-workflows)')
    expect(overview).toContain('[Business Capabilities Guide](business-capabilities)')
    expect(page).toContain('# Business Capabilities Guide')
    expect(page).toContain('[WeChat Export](/wechat)')
    expect(page).toContain('[Hot Topics](/hot)')
    expect(page).toContain('[Image Prompts](/prompts)')
    expect(page).toContain('[Image Workspace](/image-generator)')
    expect(page).toContain('[My Tasks](/tasks)')
  })
})
