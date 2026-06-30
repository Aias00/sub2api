import * as cheerio from 'cheerio'
import vm from 'node:vm'

export interface WechatArticleSummary {
  title: string
  author: string
  accountName: string
  accountAlias: string
  accountAvatar: string
  accountDescription: string
  publishAt: string
  fakeid: string
  isOriginal: boolean
  isPaySubscribe: boolean
  cover: string
  digest: string
  metadataSeed: Record<string, unknown>
}

function extractAssignedObjectSource(scriptContent: string, assignmentPrefix: string) {
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
      if (braceDepth === 0) return scriptContent.slice(objectStartIndex, index + 1)
    }
  }

  return null
}

function parseCgiDataNew(html: string) {
  const $ = cheerio.load(html)
  const targetScript = $('script[type="text/javascript"][h5only]')
    .toArray()
    .map((element) => $(element).html() || '')
    .find((content) => content.includes('window.cgiDataNew = '))
  if (!targetScript) return null
  const objectSource = extractAssignedObjectSource(targetScript, 'window.cgiDataNew = ')
  if (!objectSource) return null
  try {
    return vm.runInNewContext(`(${objectSource})`, {})
  } catch {
    return null
  }
}

function extractBizFromHtml(html: string) {
  const patterns = [
    /window\.biz\s*=\s*["']([^"']+)["']/,
    /var biz = ["']([^"']+)["']/,
    /"biz"\s*:\s*["']([^"']+)["']/,
    /__biz=([^&"'<>]+)/,
  ]
  for (const pattern of patterns) {
    const match = html.match(pattern)
    if (match?.[1]) return match[1]
  }
  return ''
}

function normalizeBiz(value: string) {
  const normalized = String(value || '').trim()
  if (!normalized || normalized.includes('${') || normalized.includes('window.')) return ''
  return normalized
}

function normalizeUnixTimestamp(value: unknown) {
  const numeric = Number.parseInt(String(value || ''), 10)
  if (!Number.isFinite(numeric) || numeric < 1_000_000_000) return ''
  return new Date(numeric * 1000).toISOString()
}

function numericSeed(value: unknown, fallback = 0) {
  const numeric = Number.parseInt(String(value || ''), 10)
  return Number.isFinite(numeric) ? numeric : fallback
}

function extractNumericByPatterns(html: string, patterns: RegExp[]) {
  for (const pattern of patterns) {
    const match = html.match(pattern)
    if (!match?.[1]) continue
    const numeric = Number.parseInt(match[1], 10)
    if (Number.isFinite(numeric)) return numeric
  }
  return undefined
}

function extractStringByPatterns(html: string, patterns: RegExp[]) {
  for (const pattern of patterns) {
    const match = html.match(pattern)
    if (match?.[1]) return match[1]
  }
  return ''
}

export function extractWechatArticleSummaryFromHtml(html: string): WechatArticleSummary {
  const $ = cheerio.load(html)
  const cgiData = parseCgiDataNew(html) || {}

  const domTitle =
    $('#activity-name').text().trim() ||
    $('meta[property="og:title"]').attr('content') ||
    ''
  const domAuthor =
    $('#js_author_name_text').text().trim() ||
    $('#js_author_name').text().trim() ||
    ''
  const domAccountName = $('#js_name').text().trim() || domAuthor
  const domAccountAlias = $('mp-common-profile').attr('data-alias') || ''
  const domAccountAvatar =
    $('mp-common-profile').attr('data-headimg') ||
    String(cgiData.round_head_img || cgiData.hd_head_img || '')
  const domAccountDescription = $('mp-common-profile').attr('data-signature') || ''
  const domDigest = $('meta[property="og:description"]').attr('content') || ''
  const domCover =
    $('#js_cover').attr('data-src') ||
    $('meta[property="og:image"]').attr('content') ||
    ''

  const appmsgid = numericSeed(cgiData.appmsgid || cgiData.mid)
  const itemidx = numericSeed(cgiData.itemidx || cgiData.idx, 1)
  const createTime = numericSeed(cgiData.create_time || cgiData.publish_time)
  const fakeid = normalizeBiz(String(cgiData.biz || extractBizFromHtml(html) || ''))
  const readNum = extractNumericByPatterns(html, [
    /["']read_num["']\s*:\s*(\d+)/,
    /read_num\s*=\s*["']?(\d+)/,
    /readNum\s*[:=]\s*["']?(\d+)/,
  ])
  const oldLikeNum = extractNumericByPatterns(html, [
    /["']old_like_num["']\s*:\s*(\d+)/,
    /old_like_num\s*=\s*["']?(\d+)/,
    /oldLikeNum\s*[:=]\s*["']?(\d+)/,
  ])
  const likeNum = extractNumericByPatterns(html, [
    /["']like_num["']\s*:\s*(\d+)/,
    /like_num\s*=\s*["']?(\d+)/,
    /likeNum\s*[:=]\s*["']?(\d+)/,
  ])
  const commentNum = extractNumericByPatterns(html, [
    /["']comment_count["']\s*:\s*(\d+)/,
    /comment_count\s*=\s*["']?(\d+)/,
    /commentNum\s*[:=]\s*["']?(\d+)/,
  ])
  const appmsgToken = extractStringByPatterns(html, [
    /window\.appmsg_token\s*=\s*["']([^"']+)["']/,
    /appmsg_token\s*[:=]\s*["']([^"']+)["']/,
  ])
  const commentId = extractStringByPatterns(html, [
    /window\.comment_id\s*=\s*["']?([^"';\s]+)["']?/,
    /comment_id\s*[:=]\s*["']?([^"',;\s}]+)["']?/,
  ])

  return {
    title: String(cgiData.title || cgiData.msg_title || domTitle || ''),
    author: String(cgiData.author || cgiData.nick_name || domAuthor || ''),
    accountName: String(cgiData.nick_name || domAccountName || ''),
    accountAlias: domAccountAlias,
    accountAvatar: domAccountAvatar,
    accountDescription: domAccountDescription,
    publishAt: normalizeUnixTimestamp(cgiData.create_time || cgiData.publish_time),
    fakeid,
    isOriginal: Number(cgiData.copyright_stat || 0) === 11,
    isPaySubscribe: Number(cgiData.is_pay_subscribe || 0) === 1,
    cover: domCover,
    digest: String(cgiData.digest || domDigest || ''),
    metadataSeed: {
      aid: appmsgid > 0 ? `${appmsgid}_${itemidx || 1}` : '',
      appmsgid,
      itemidx,
      createTime,
      readNum,
      oldLikeNum,
      likeNum,
      commentNum,
      appmsgToken,
      commentId,
      source: 'worker_html_parse',
      accountName: String(cgiData.nick_name || domAccountName || ''),
    },
  }
}
