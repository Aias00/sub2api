import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const styleSource = readFileSync('src/style.css', 'utf8')
const appLayoutSource = readFileSync('src/components/layout/AppLayout.vue', 'utf8')
const authLayoutSource = readFileSync('src/components/layout/AuthLayout.vue', 'utf8')
const mainSource = readFileSync('src/main.ts', 'utf8')
const publicDarkHeaderSource = readFileSync('src/components/layout/PublicDarkHeader.vue', 'utf8')
const homeViewSource = readFileSync('src/views/HomeView.vue', 'utf8')
const docsViewSource = readFileSync('src/views/public/DocsView.vue', 'utf8')
const legalViewSource = readFileSync('src/views/public/LegalDocumentView.vue', 'utf8')
const promptCatalogViewSource = readFileSync('src/views/public/PromptCatalogView.vue', 'utf8')
const weChatExportViewSource = readFileSync('src/views/public/WeChatExportView.vue', 'utf8')
const hotContentViewSource = readFileSync('src/views/public/HotContentView.vue', 'utf8')
const imageGeneratorViewSource = readFileSync('src/views/public/ImageGeneratorView.vue', 'utf8')
const keyUsageViewSource = readFileSync('src/views/KeyUsageView.vue', 'utf8')

describe('home business Vercel design skin', () => {
  it('defines the adapted Vercel DESIGN.md token surface', () => {
    expect(styleSource).toContain('--vercel-ink: #171717')
    expect(styleSource).toContain('--vercel-hairline: #ebebeb')
    expect(styleSource).toContain('--vercel-canvas-soft: #fafafa')
    expect(styleSource).toContain('--vercel-link: #0070f3')
    expect(styleSource).toContain('radial-gradient(circle at 26% 4%, rgba(0, 124, 240, 0.16), transparent 24%)')
  })

  it('keeps Vercel-inspired public pages crisp and compact', () => {
    expect(styleSource).toContain('border-radius: 12px !important')
    expect(styleSource).toContain('background-color: var(--vercel-ink) !important')
    expect(styleSource).toContain(".home-business-page a[class*='bg-[#171717]']")
    expect(styleSource).toContain("font-family: 'Geist Mono', ui-monospace")
    expect(styleSource).toContain('letter-spacing: 0 !important')
  })

  it('uses explicit token classes for home page color overrides', () => {
    expect(styleSource).toContain('.home-business-page .home-surface')
    expect(styleSource).toContain('.home-business-page .home-border')
    expect(styleSource).toContain('.home-business-page .home-ink')
    expect(styleSource).toContain('.home-business-page .home-accent')
    expect(styleSource).not.toContain(".home-business-page [class*='text-white']")
    expect(styleSource).not.toContain(".home-business-page [class*='text-cyan-']")
    expect(styleSource).not.toContain(".home-business-page [class*='bg-[#101114]']")
    expect(styleSource).not.toContain(".home-business-page [class*='border-white/10']")
  })

  it('keeps public tool pages on the dark template explicitly', () => {
    expect(styleSource).toContain('.home-business-page.public-dark-page')
    expect(styleSource).toContain("html[data-public-theme='light'] .home-business-page.public-dark-page")
    expect(styleSource).toContain('.home-business-page.public-dark-page .public-dark-header')
    expect(styleSource).toContain('.public-dark-header__theme-toggle')
    expect(promptCatalogViewSource).toContain('home-business-page public-dark-page')
    expect(weChatExportViewSource).toContain('home-business-page public-dark-page')
    expect(hotContentViewSource).toContain('home-business-page public-dark-page')
    expect(imageGeneratorViewSource).toContain('home-business-page public-dark-page')
  })

  it('supports switching the public tool theme template', () => {
    expect(mainSource).toContain('initPublicTheme()')
    expect(mainSource).not.toContain("localStorage.setItem('theme', 'light')")
    expect(publicDarkHeaderSource).toContain('togglePublicTheme')
    expect(publicDarkHeaderSource).toContain("Icon :name=\"isDarkTheme ? 'sun' : 'moon'\"")
    expect(publicDarkHeaderSource).toContain('public-dark-header__theme-toggle')
  })

  it('applies the Vercel shell to shared and standalone route entry points', () => {
    expect(appLayoutSource).toContain('vercel-app-shell')
    expect(authLayoutSource).toContain('vercel-auth-shell')
    expect(homeViewSource).toContain('home-business-page')
    expect(docsViewSource).toContain('docs-page home-business-page')
    expect(legalViewSource).toContain('home-business-page')
    expect(promptCatalogViewSource).toContain('home-business-page')
    expect(keyUsageViewSource).toContain('home-business-page')
  })
})
