import * as cheerio from 'cheerio'
import { mkdir, writeFile } from 'node:fs/promises'
import path from 'node:path'

import {
  exportWechatArticleContentAsHtml,
  exportWechatArticleContentAsMarkdown,
  getWechatArticleContentHtml,
  WechatExportMetadata,
} from '../core/exporter'

export type WechatExportFormat = 'html' | 'markdown'

export interface WechatExportRunnerArticle {
  accountName: string
  aid: string
  link: string
  title: string
  html: string
  metadata?: WechatExportMetadata
  comments?: any[]
}

export interface WechatGeneratedArtifact {
  format: WechatExportFormat
  fileName: string
  filePath: string
}

function sanitizeWechatExportFileName(fileName: string) {
  return fileName
    .replace(/[<>:"/\\|?*\u0000-\u001F]/g, '_')
    .replace(/\s+/g, '_')
    .slice(0, 120)
}

const inlineImageTimeoutMs = Number(process.env.WECHAT_EXPORT_INLINE_IMAGE_TIMEOUT_MS || 15000)
const inlineImageMaxBytes = Number(process.env.WECHAT_EXPORT_INLINE_IMAGE_MAX_BYTES || 15 * 1024 * 1024)

function getWechatImageSource($element: cheerio.Cheerio<any>) {
  return (
    $element.attr('data-src') ||
    $element.attr('data-backsrc') ||
    $element.attr('data-original') ||
    $element.attr('src') ||
    ''
  )
}

function isRemoteImageURL(rawURL: string) {
  try {
    const parsed = new URL(rawURL)
    return parsed.protocol === 'https:' || parsed.protocol === 'http:'
  } catch {
    return false
  }
}

async function fetchImageDataURI(rawURL: string) {
  const controller = new AbortController()
  const timeout = setTimeout(() => controller.abort(), inlineImageTimeoutMs)
  try {
    const response = await fetch(rawURL, {
      signal: controller.signal,
      headers: {
        'user-agent':
          'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125 Safari/537.36',
        accept: 'image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8',
      },
    })
    if (!response.ok) return null
    const contentType = response.headers.get('content-type') || ''
    if (!contentType.toLowerCase().startsWith('image/')) return null
    const contentLength = Number(response.headers.get('content-length') || 0)
    if (contentLength > inlineImageMaxBytes) return null
    const bytes = Buffer.from(await response.arrayBuffer())
    if (bytes.length > inlineImageMaxBytes) return null
    return `data:${contentType.split(';')[0]};base64,${bytes.toString('base64')}`
  } catch (error) {
    console.warn('[wechat-worker] image inline failed', { url: rawURL, error })
    return null
  } finally {
    clearTimeout(timeout)
  }
}

async function inlineRemoteImages(contentHtml: string) {
  const $ = cheerio.load(`<article>${contentHtml}</article>`, { decodeEntities: false } as any)
  const images = $('img').toArray()
  for (const element of images) {
    const $element = $(element)
    const src = getWechatImageSource($element).replace(/&amp;/g, '&').trim()
    if (!src || !isRemoteImageURL(src)) continue
    const dataURI = await fetchImageDataURI(src)
    if (!dataURI) {
      $element.attr('src', src)
      continue
    }
    $element.attr('src', dataURI)
    $element.attr('data-original-src', src)
    $element.removeAttr('data-src')
    $element.removeAttr('data-backsrc')
    $element.removeAttr('data-original')
  }
  return $.html('article')
}

export async function runWechatExportFormats({
  outputDir,
  taskId,
  formats,
  includeEngagement,
  articles,
}: {
  outputDir: string
  taskId: string
  formats: WechatExportFormat[]
  includeEngagement?: boolean
  articles: WechatExportRunnerArticle[]
}) {
  await mkdir(outputDir, { recursive: true })

  const artifacts: WechatGeneratedArtifact[] = []

  for (const article of articles) {
    const articleBaseName = sanitizeWechatExportFileName(article.title)
    const contentHtml = getWechatArticleContentHtml(article.html)
    const standaloneContentHtml = formats.includes('html') || formats.includes('markdown')
      ? await inlineRemoteImages(contentHtml)
      : contentHtml

    if (formats.includes('html')) {
      const fileName = `${articleBaseName}.html`
      const filePath = path.join(outputDir, fileName)
      await writeFile(filePath, exportWechatArticleContentAsHtml(standaloneContentHtml, {
        title: article.title,
        accountName: article.accountName,
        link: article.link,
        metadata: article.metadata,
        includeEngagement,
      }), 'utf8')
      artifacts.push({ format: 'html', fileName, filePath })
    }

    if (formats.includes('markdown')) {
      const fileName = `${articleBaseName}.md`
      const filePath = path.join(outputDir, fileName)
      await writeFile(filePath, exportWechatArticleContentAsMarkdown(standaloneContentHtml, {
        title: article.title,
        accountName: article.accountName,
        link: article.link,
        includeEngagement,
      }), 'utf8')
      artifacts.push({ format: 'markdown', fileName, filePath })
    }
  }

  return artifacts
}
