import * as cheerio from 'cheerio'
import * as htmlDocx from 'html-docx-js-typescript'
import TurndownService from 'turndown'
import vm from 'node:vm'

export interface WechatExportMetadata {
  readNum?: number
  oldLikeNum?: number
  shareNum?: number
  likeNum?: number
  commentNum?: number
  appmsgToken?: string
  commentId?: string
  engagementFetchStatus?: string
  engagementFetchMessage?: string
}

export interface WechatRichBlock {
  type: 'image' | 'video' | 'iframe' | 'audio' | 'link' | 'mini_program' | 'poll' | 'text'
  title?: string
  src?: string
  href?: string
  text?: string
  raw?: Record<string, string>
}

const preserveEntityCheerioOptions = { decodeEntities: false } as any

function extractWechatAssignedObjectSource(scriptContent: string, assignmentPrefix: string) {
  const prefixIndex = scriptContent.indexOf(assignmentPrefix)
  if (prefixIndex === -1) return null
  const objectStartIndex = scriptContent.indexOf('{', prefixIndex)
  if (objectStartIndex === -1) return null

  let braceDepth = 0
  let inSingleQuote = false
  let inDoubleQuote = false
  let inTemplateString = false
  let isEscaped = false

  for (let index = objectStartIndex; index < scriptContent.length; index += 1) {
    const currentChar = scriptContent[index]
    if (isEscaped) {
      isEscaped = false
      continue
    }
    if (currentChar === '\\') {
      isEscaped = true
      continue
    }
    if (!inDoubleQuote && !inTemplateString && currentChar === "'") {
      inSingleQuote = !inSingleQuote
      continue
    }
    if (!inSingleQuote && !inTemplateString && currentChar === '"') {
      inDoubleQuote = !inDoubleQuote
      continue
    }
    if (!inSingleQuote && !inDoubleQuote && currentChar === '`') {
      inTemplateString = !inTemplateString
      continue
    }
    if (inSingleQuote || inDoubleQuote || inTemplateString) continue
    if (currentChar === '{') {
      braceDepth += 1
      continue
    }
    if (currentChar === '}') {
      braceDepth -= 1
      if (braceDepth === 0) {
        return scriptContent.slice(objectStartIndex, index + 1)
      }
    }
  }
  return null
}

function parseWechatCgiDataNew(rawHtml: string) {
  const $ = cheerio.load(rawHtml)
  const targetScript = $('script[type="text/javascript"][h5only]')
    .toArray()
    .map((element) => $(element).html() || '')
    .find((content) => content.includes('window.cgiDataNew = '))
  if (!targetScript) return null
  const objectSource = extractWechatAssignedObjectSource(targetScript, 'window.cgiDataNew = ')
  if (!objectSource) return null
  return vm.runInNewContext(`(${objectSource})`, {})
}

function buildWechatContentFallback(rawHtml: string) {
  const cgiData = parseWechatCgiDataNew(rawHtml)
  if (!cgiData) return null
  const fallbackHtml =
    cgiData.content_noencode ||
    cgiData.text_page_info?.content_noencode ||
    cgiData.page_content ||
    ''
  const fallbackTitle = cgiData.title || cgiData.msg_title || ''
  return { title: String(fallbackTitle || ''), contentHtml: String(fallbackHtml || '') }
}

function extractArticleContentHtml(rawHtml: string) {
  const fallback = buildWechatContentFallback(rawHtml)
  if (fallback?.contentHtml) return normalizePlainTextContentHtml(fallback.contentHtml)
  const $ = cheerio.load(rawHtml, preserveEntityCheerioOptions)
  const content = $('#js_content')
  return normalizePlainTextContentHtml(content.length > 0 ? content.html() || '' : rawHtml)
}

function normalizePlainTextContentHtml(contentHtml: string) {
  const content = String(contentHtml || '')
  if (!content.trim()) return content
  if (!/[\r\n]/.test(content)) return content
  if (/<[a-zA-Z][\s\S]*?>/.test(content)) return content

  return content
    .replace(/\r\n?/g, '\n')
    .split(/\n{2,}/)
    .map((paragraph) => paragraph.trim())
    .filter(Boolean)
    .map((paragraph) => `<p>${paragraph.split(/\n/).map((line) => escapeHtml(line.trim())).join('<br />')}</p>`)
    .join('\n')
}

function normalizeWechatLazyMedia($: cheerio.CheerioAPI) {
  $('img').each((_index, element) => {
    const $element = $(element)
    const src =
      $element.attr('data-src') ||
      $element.attr('data-backsrc') ||
      $element.attr('data-original') ||
      $element.attr('src') ||
      ''
    if (src) {
      $element.attr('src', src)
    }
    $element.removeAttr('data-src')
    $element.removeAttr('data-backsrc')
    $element.removeAttr('data-original')
  })
  $('iframe, video, audio, source').each((_index, element) => {
    const $element = $(element)
    const src = $element.attr('data-src') || $element.attr('src') || ''
    if (src) {
      $element.attr('src', src)
    }
  })
}

function buildStandaloneHtml({
  title,
  accountName,
  link,
  contentHtml,
  includeEngagement,
}: {
  title?: string
  accountName?: string
  link?: string
  contentHtml: string
  includeEngagement?: boolean
}) {
  const $ = cheerio.load(`<article class="wechat-export-article">${contentHtml}</article>`, preserveEntityCheerioOptions)
  normalizeWechatLazyMedia($)
  decodeCloudflareProtectedEmails($)
  const engagementNote = includeEngagement
    ? '<p class="wechat-export-note">Engagement metrics were requested. If they are absent from JSON metadata, the current session/article HTML did not expose them.</p>'
    : ''
  return `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>${escapeHtml(title || 'WeChat article')}</title>
  <style>
    body { margin: 0; background: #f6f2ea; color: #1f2933; font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; }
    main { max-width: 760px; margin: 0 auto; padding: 40px 20px 64px; background: #fff; min-height: 100vh; }
    header { border-bottom: 1px solid #ece7df; margin-bottom: 28px; padding-bottom: 20px; }
    h1 { font-size: 30px; line-height: 1.25; margin: 0 0 12px; }
    .meta, .wechat-export-note { color: #687280; font-size: 13px; line-height: 1.7; }
    .wechat-export-article { font-size: 16px; line-height: 1.85; overflow-wrap: anywhere; }
    img, video, iframe { max-width: 100%; }
    pre { white-space: pre-wrap; }
  </style>
</head>
<body>
  <main>
    <header>
      <h1>${escapeHtml(title || 'Untitled')}</h1>
      <p class="meta">${escapeHtml(accountName || '')}${link ? ` · <a href="${escapeHtml(link)}">${escapeHtml(link)}</a>` : ''}</p>
      ${engagementNote}
    </header>
    ${$.html('article')}
  </main>
</body>
</html>`
}

function escapeHtml(value: string) {
  return String(value || '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}

function getElementMediaSrc($element: cheerio.Cheerio<any>) {
  return (
    $element.attr('src') ||
    $element.attr('data-src') ||
    $element.find('source').first().attr('src') ||
    ''
  )
}

function replaceElementWithMarkdownPlaceholder(
  $element: cheerio.Cheerio<any>,
  label: string,
  details: string[],
) {
  const compactDetails = details.map((detail) => detail.trim()).filter(Boolean).join(' · ')
  const placeholder = compactDetails ? `[WeChat ${label}: ${compactDetails}]` : `[WeChat ${label}]`
  $element.replaceWith(`<p>${escapeHtml(placeholder)}</p>`)
}

function addWechatRichBlockMarkdownPlaceholders($: cheerio.CheerioAPI) {
  $('video').each((_index, element) => {
    const $element = $(element)
    replaceElementWithMarkdownPlaceholder($element, 'video', [
      $element.attr('title') || '',
      getElementMediaSrc($element),
    ])
  })
  $('audio').each((_index, element) => {
    const $element = $(element)
    replaceElementWithMarkdownPlaceholder($element, 'audio', [
      $element.attr('title') || '',
      getElementMediaSrc($element),
    ])
  })
  $('iframe').each((_index, element) => {
    const $element = $(element)
    replaceElementWithMarkdownPlaceholder($element, 'iframe', [
      $element.attr('title') || '',
      getElementMediaSrc($element),
    ])
  })
  $('mp-miniprogram, .weapp_display_element, [data-miniprogram-appid]').each((_index, element) => {
    const $element = $(element)
    replaceElementWithMarkdownPlaceholder($element, 'mini program', [
      $element.attr('data-miniprogram-title') || $element.text().trim(),
      $element.attr('data-miniprogram-appid') || '',
      $element.attr('data-miniprogram-path') || '',
    ])
  })
  $('[id*="vote"], [class*="vote"], [data-voteid]').each((_index, element) => {
    const $element = $(element)
    replaceElementWithMarkdownPlaceholder($element, 'poll', [
      $element.attr('data-voteid') || '',
      $element.text().trim().slice(0, 160),
    ])
  })
}

function decodeCloudflareProtectedEmail(encodedValue: string) {
  if (!encodedValue || encodedValue.length < 2 || encodedValue.length % 2 !== 0) return ''
  const xorKey = Number.parseInt(encodedValue.slice(0, 2), 16)
  if (!Number.isFinite(xorKey)) return ''
  let decodedEmail = ''
  for (let index = 2; index < encodedValue.length; index += 2) {
    const encodedByte = Number.parseInt(encodedValue.slice(index, index + 2), 16)
    if (!Number.isFinite(encodedByte)) return ''
    decodedEmail += String.fromCharCode(encodedByte ^ xorKey)
  }
  return decodedEmail
}

function decodeCloudflareProtectedEmails($: cheerio.CheerioAPI) {
  $('a.__cf_email__, span.__cf_email__').each((_index, element) => {
    const $element = $(element)
    const encodedValue =
      $element.attr('data-cfemail') ||
      ($element.attr('href')?.match(/#([0-9a-fA-F]+)/)?.[1] ?? '')
    const decodedEmail = decodeCloudflareProtectedEmail(encodedValue)
    if (!decodedEmail) return
    $element.text(decodedEmail)
    if ($element.is('a')) {
      $element.attr('href', `mailto:${decodedEmail}`)
    }
    $element.removeClass('__cf_email__')
    $element.removeAttr('data-cfemail')
  })
}

export function exportWechatArticleAsText(rawHtml: string) {
  const fallback = buildWechatContentFallback(rawHtml)
  if (fallback?.contentHtml) {
    const $fallback = cheerio.load(`<div>${fallback.contentHtml}</div>`)
    decodeCloudflareProtectedEmails($fallback)
    return $fallback.text().trim()
  }
  const $ = cheerio.load(rawHtml)
  decodeCloudflareProtectedEmails($)
  return $('#js_content').text().trim() || $.root().text().trim()
}

export function getWechatArticleContentHtml(rawHtml: string) {
  return extractArticleContentHtml(rawHtml)
}

export function exportWechatArticleContentAsHtml(contentHtml: string, options: {
  title?: string
  accountName?: string
  link?: string
  metadata?: WechatExportMetadata
  includeEngagement?: boolean
} = {}) {
  return buildStandaloneHtml({
    title: options.title,
    accountName: options.accountName,
    link: options.link,
    contentHtml,
    includeEngagement: options.includeEngagement,
  })
}

export function exportWechatArticleAsHtml(rawHtml: string, options: {
  title?: string
  accountName?: string
  link?: string
  metadata?: WechatExportMetadata
  includeEngagement?: boolean
} = {}) {
  return exportWechatArticleContentAsHtml(getWechatArticleContentHtml(rawHtml), options)
}

export function exportWechatArticleContentAsMarkdown(contentHtml: string, options: {
  title?: string
  accountName?: string
  link?: string
  includeEngagement?: boolean
} = {}) {
  const turndown = new TurndownService()
  const $ = cheerio.load(`<article>${contentHtml}</article>`, preserveEntityCheerioOptions)
  normalizeWechatLazyMedia($)
  decodeCloudflareProtectedEmails($)
  addWechatRichBlockMarkdownPlaceholders($)
  const metadataLines = [
    options.accountName ? `> ${options.accountName}` : '',
    options.link ? `> Source: ${options.link}` : '',
    options.includeEngagement ? '> Engagement: requested; see JSON metadata for availability.' : '',
  ].filter(Boolean)
  const header = [
    `# ${options.title || 'Untitled'}`,
    ...metadataLines,
  ].join('\n')
  return `${header}\n\n${turndown.turndown($.html('article'))}`.trim()
}

export function exportWechatArticleAsMarkdown(rawHtml: string, options: {
  title?: string
  accountName?: string
  link?: string
  includeEngagement?: boolean
} = {}) {
  return exportWechatArticleContentAsMarkdown(getWechatArticleContentHtml(rawHtml), options)
}

export function exportWechatArticleRichBlocks(rawHtml: string): WechatRichBlock[] {
  const $ = cheerio.load(`<article>${extractArticleContentHtml(rawHtml)}</article>`, preserveEntityCheerioOptions)
  normalizeWechatLazyMedia($)
  const blocks: WechatRichBlock[] = []
  $('img').each((_index, element) => {
    blocks.push({ type: 'image', src: $(element).attr('src') || '', raw: { alt: $(element).attr('alt') || '' } })
  })
  $('video').each((_index, element) => {
    blocks.push({ type: 'video', src: getElementMediaSrc($(element)) })
  })
  $('iframe').each((_index, element) => {
    blocks.push({ type: 'iframe', src: getElementMediaSrc($(element)), title: $(element).attr('title') || '' })
  })
  $('audio').each((_index, element) => {
    blocks.push({ type: 'audio', src: getElementMediaSrc($(element)) })
  })
  $('a').each((_index, element) => {
    const href = $(element).attr('href') || ''
    const text = $(element).text().trim()
    if (href || text) blocks.push({ type: 'link', href, text })
  })
  $('mp-miniprogram, .weapp_display_element, [data-miniprogram-appid]').each((_index, element) => {
    const $element = $(element)
    blocks.push({
      type: 'mini_program',
      title: $element.attr('data-miniprogram-title') || $element.text().trim(),
      raw: {
        appid: $element.attr('data-miniprogram-appid') || '',
        path: $element.attr('data-miniprogram-path') || '',
      },
    })
  })
  $('[id*="vote"], [class*="vote"], [data-voteid]').each((_index, element) => {
    blocks.push({ type: 'poll', title: $(element).text().trim().slice(0, 160) })
  })
  return blocks
}

export async function exportWechatArticleAsDocx(rawHtml: string) {
  return htmlDocx.asBlob(rawHtml)
}
