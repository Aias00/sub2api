import assert from 'node:assert/strict'
import { mkdtemp, readFile, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import path from 'node:path'

import { assertWechatArticleHtml, isWechatVerifyPageHtml, wechatVerifyPageMessage } from '../core/verification'
import { runWechatExportFormats } from '../runner/export-runner'

const tinyPng = Buffer.from(
  'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAFgwJ/l0Yd4QAAAABJRU5ErkJggg==',
  'base64',
)

const richArticleHtml = String.raw`
<!doctype html>
<html>
  <body>
    <h1 id="activity-name">Rich WeChat fixture</h1>
    <div id="js_content">
      <p>正文开头</p>
      <img data-src="https://static.cloudbase.eu.org/uploads/prompts/example.jpg" alt="封面图" />
      <video title="演示视频" data-src="https://res.wx.qq.com/video.mp4"></video>
      <audio title="语音朗读"><source data-src="https://res.wx.qq.com/audio.mp3" /></audio>
      <iframe title="内嵌内容" data-src="https://mp.weixin.qq.com/mp/readtemplate?t=pages/video"></iframe>
      <mp-miniprogram
        data-miniprogram-title="小程序卡片"
        data-miniprogram-appid="wx123456"
        data-miniprogram-path="pages/index/index"
      >打开小程序</mp-miniprogram>
      <section class="vote_area" data-voteid="vote-42">你更喜欢哪种导出格式？</section>
      <section class="paywall_area" data-pay-subscribe="1">付费内容：订阅后可继续阅读</section>
      <p><a id="js_share_source" href="https://example.com/original-source">阅读原文</a></p>
      <p>联系作者：<a class="__cf_email__" data-cfemail="127d626152776a737f627e773c717d7f">[email&#160;protected]</a></p>
      <p>正文结尾</p>
    </div>
  </body>
</html>
`

const plainTextWechatHtml = String.raw`
<!doctype html>
<html>
  <body>
    <h1 id="activity-name">Plain text WeChat fixture</h1>
    <div id="js_content">第一段第一行
第一段第二行

第二段内容

第三段内容</div>
  </body>
</html>
`

async function main() {
  const verifyHtml = `<script>var PAGE_MID='mmbizwap:secitptpage/verify.html';window.cgiData={cap_sid:"sid",poc_token:"token",target_url:"/s/demo"}</script>`
  assert.equal(isWechatVerifyPageHtml(verifyHtml), true, 'WeChat verify pages should be detected')
  assert.throws(() => assertWechatArticleHtml(verifyHtml), new RegExp(wechatVerifyPageMessage), 'WeChat verify pages should fail article HTML validation')
  assert.equal(isWechatVerifyPageHtml(plainTextWechatHtml), false, 'Normal WeChat article HTML should pass verify-page detection')

  const outputDir = await mkdtemp(path.join(tmpdir(), 'wechat-fidelity-'))
  const originalFetch = globalThis.fetch
  globalThis.fetch = (async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = String(input)
    if (url === 'https://static.cloudbase.eu.org/uploads/prompts/example.jpg') {
      return new Response(tinyPng, {
        status: 200,
        headers: { 'content-type': 'image/png', 'content-length': String(tinyPng.length) },
      })
    }
    return originalFetch(input, init)
  }) as typeof fetch
  try {
    const artifacts = await runWechatExportFormats({
      outputDir,
      taskId: 'fidelity-fixture',
      formats: ['html', 'markdown'],
      includeEngagement: true,
      articles: [
        {
          accountName: '迁移验收号',
          aid: 'fixture_1',
          link: 'https://mp.weixin.qq.com/s/fidelity-fixture',
          title: 'Rich WeChat fixture',
          html: richArticleHtml,
          metadata: {
            readNum: 123,
            oldLikeNum: 45,
            likeNum: 67,
            commentNum: 8,
          },
        },
        {
          accountName: '纯文本号',
          aid: 'fixture_2',
          link: 'https://mp.weixin.qq.com/s/plain-text-fidelity-fixture',
          title: 'Plain text WeChat fixture',
          html: plainTextWechatHtml,
        },
      ],
    })

    const htmlArtifact = artifacts.find((artifact) => artifact.format === 'html')
    const markdownArtifact = artifacts.find((artifact) => artifact.format === 'markdown')
    assert(htmlArtifact, 'HTML artifact was not generated')
    assert(markdownArtifact, 'Markdown artifact was not generated')

    const html = await readFile(htmlArtifact.filePath, 'utf8')
    assert.match(html, /src="data:image\/png;base64,/, 'HTML should inline remote images as data URIs')
    assert.match(html, /data-original-src="https:\/\/static\.cloudbase\.eu\.org\/uploads\/prompts\/example\.jpg"/, 'HTML should keep original image source for traceability')
    assert.match(html, /https:\/\/res\.wx\.qq\.com\/video\.mp4/, 'HTML should keep video source')
    assert.match(html, /https:\/\/res\.wx\.qq\.com\/audio\.mp3/, 'HTML should keep audio source')
    assert.match(html, /wx123456/, 'HTML should keep mini-program metadata')
    assert.match(html, /vote-42/, 'HTML should keep poll metadata')
    assert.match(html, /付费内容：订阅后可继续阅读/, 'HTML should keep paid-content notices')
    assert.match(html, /https:\/\/example\.com\/original-source/, 'HTML should keep read-original links')
    assert.match(html, /ops@example\.com/, 'HTML should decode protected email text')
    assert.match(html, /mailto:ops@example\.com/, 'HTML should decode protected email href')

    const markdown = await readFile(markdownArtifact.filePath, 'utf8')
    assert.match(markdown, /\[WeChat video:/, 'Markdown should expose a video placeholder')
    assert.match(markdown, /\[WeChat audio:/, 'Markdown should expose an audio placeholder')
    assert.match(markdown, /\[WeChat iframe:/, 'Markdown should expose an iframe placeholder')
    assert.match(markdown, /\[WeChat mini program:/, 'Markdown should expose a mini-program placeholder')
    assert.match(markdown, /\[WeChat poll:/, 'Markdown should expose a poll placeholder')
    assert.match(markdown, /> Engagement: requested; see JSON metadata for availability\.\n\n正文开头/, 'Markdown should separate quote metadata from article body')
    assert.doesNotMatch(markdown, /> Source: https:\/\/mp\.weixin\.qq\.com\/s\/fidelity-fixture\n正文开头/, 'Markdown should not merge the first body paragraph into source quote metadata')
    assert.match(markdown, /https:\/\/res\.wx\.qq\.com\/video\.mp4/, 'Markdown video placeholder should keep source URL')
    assert.match(markdown, /wx123456/, 'Markdown mini-program placeholder should keep appid')
    assert.match(markdown, /付费内容：订阅后可继续阅读/, 'Markdown should keep paid-content notices')
    assert.match(markdown, /https:\/\/example\.com\/original-source/, 'Markdown should keep read-original links')
    assert.match(markdown, /ops@example\.com/, 'Markdown should decode protected email text')

    const plainMarkdownArtifact = artifacts.find((artifact) => artifact.format === 'markdown' && artifact.fileName.includes('Plain_text_WeChat_fixture'))
    const plainHtmlArtifact = artifacts.find((artifact) => artifact.format === 'html' && artifact.fileName.includes('Plain_text_WeChat_fixture'))
    assert(plainMarkdownArtifact, 'Plain text Markdown artifact was not generated')
    assert(plainHtmlArtifact, 'Plain text HTML artifact was not generated')
    const plainMarkdown = await readFile(plainMarkdownArtifact.filePath, 'utf8')
    const plainHtml = await readFile(plainHtmlArtifact.filePath, 'utf8')
    assert.match(plainMarkdown, /第一段第一行  \n第一段第二行\n\n第二段内容\n\n第三段内容/, 'Markdown should preserve plaintext paragraph breaks from WeChat fallback content')
    assert.match(plainHtml, /<p>第一段第一行<br ?\/?>第一段第二行<\/p>\n<p>第二段内容<\/p>\n<p>第三段内容<\/p>/, 'HTML should restore plaintext fallback content into paragraphs')

    console.log('WeChat rich export fidelity check complete.')
    console.log(`artifacts=${artifacts.length}`)
  } finally {
    globalThis.fetch = originalFetch
    await rm(outputDir, { recursive: true, force: true })
  }
}

main().catch((error) => {
  console.error(error)
  process.exit(1)
})
