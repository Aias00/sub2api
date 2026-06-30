export const wechatVerifyPageMessage = '微信返回验证页，请通过公众号同步导入'

export function isWechatVerifyPageHtml(rawHtml: string) {
  const normalized = rawHtml.toLowerCase()
  return (
    normalized.includes('secitptpage/verify') ||
    normalized.includes('mmbizwap:secitptpage/verify.html') ||
    (normalized.includes('cap_sid') && normalized.includes('poc_token') && normalized.includes('target_url'))
  )
}

export function assertWechatArticleHtml(rawHtml: string) {
  if (isWechatVerifyPageHtml(rawHtml)) {
    throw new Error(wechatVerifyPageMessage)
  }
}
