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
}

export interface WechatJsonExportRecord {
  accountName: string
  title: string
  link: string
  html: string
  text: string
  metadata?: WechatExportMetadata
  comments?: any[]
}

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

export function exportWechatArticleAsHtml(rawHtml: string) {
  return rawHtml
}

export function exportWechatArticleAsMarkdown(rawHtml: string) {
  const turndown = new TurndownService()
  const fallback = buildWechatContentFallback(rawHtml)
  if (fallback?.contentHtml) {
    return turndown.turndown(fallback.contentHtml)
  }
  return turndown.turndown(rawHtml)
}

export function exportWechatArticleAsJson(record: WechatJsonExportRecord) {
  return record
}

export async function exportWechatArticleAsDocx(rawHtml: string) {
  return htmlDocx.asBlob(rawHtml)
}
