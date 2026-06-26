import { mkdir, writeFile } from 'node:fs/promises'
import path from 'node:path'

import {
  exportWechatArticleAsHtml,
  exportWechatArticleAsJson,
  exportWechatArticleAsMarkdown,
  exportWechatArticleAsText,
  WechatExportMetadata,
} from '../core/exporter'

export type WechatExportFormat = 'html' | 'markdown' | 'json'

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

function getWechatTaskArchiveBaseName(taskId: string) {
  return sanitizeWechatExportFileName(`wechat-export-${taskId}`)
}

export async function runWechatExportFormats({
  outputDir,
  taskId,
  formats,
  articles,
}: {
  outputDir: string
  taskId: string
  formats: WechatExportFormat[]
  articles: WechatExportRunnerArticle[]
}) {
  await mkdir(outputDir, { recursive: true })

  const artifacts: WechatGeneratedArtifact[] = []
  const archiveBaseName = getWechatTaskArchiveBaseName(taskId)

  for (const article of articles) {
    const articleBaseName = sanitizeWechatExportFileName(article.title)

    if (formats.includes('html')) {
      const fileName = `${articleBaseName}.html`
      const filePath = path.join(outputDir, fileName)
      await writeFile(filePath, exportWechatArticleAsHtml(article.html), 'utf8')
      artifacts.push({ format: 'html', fileName, filePath })
    }

    if (formats.includes('markdown')) {
      const fileName = `${articleBaseName}.md`
      const filePath = path.join(outputDir, fileName)
      await writeFile(filePath, exportWechatArticleAsMarkdown(article.html), 'utf8')
      artifacts.push({ format: 'markdown', fileName, filePath })
    }
  }

  if (formats.includes('json')) {
    const fileName = `${archiveBaseName}.json`
    const filePath = path.join(outputDir, fileName)
    const jsonPayload = articles.map((article) =>
      exportWechatArticleAsJson({
        accountName: article.accountName,
        title: article.title,
        link: article.link,
        html: article.html,
        text: exportWechatArticleAsText(article.html),
        metadata: article.metadata,
        comments: article.comments,
      }),
    )
    await writeFile(filePath, JSON.stringify(jsonPayload, null, 2), 'utf8')
    artifacts.push({ format: 'json', fileName, filePath })
  }

  return artifacts
}
