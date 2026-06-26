export function buildWechatGatewayHeaders(cookieHeader?: string): Headers {
  const headers = new Headers({
    Referer: 'https://mp.weixin.qq.com/',
    Origin: 'https://mp.weixin.qq.com',
    'User-Agent':
      'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36',
    'Accept-Encoding': 'identity',
  })

  if (cookieHeader) {
    headers.set('Cookie', cookieHeader)
  }

  return headers
}

export async function requestWechatGateway({
  method,
  endpoint,
  query,
  body,
  cookieHeader,
}: {
  method: 'GET' | 'POST'
  endpoint: string
  query?: Record<string, string | number | undefined>
  body?: Record<string, string | number | undefined>
  cookieHeader?: string
}) {
  const url = new URL(endpoint)
  if (query) {
    for (const [key, value] of Object.entries(query)) {
      if (value === undefined || value === null) continue
      url.searchParams.set(key, String(value))
    }
  }

  return fetch(url, {
    method,
    headers: buildWechatGatewayHeaders(cookieHeader),
    body:
      method === 'POST' && body
        ? new URLSearchParams(
            Object.entries(body)
              .filter(([, value]) => value !== undefined && value !== null)
              .map(([key, value]) => [key, String(value)]),
          ).toString()
        : undefined,
    redirect: 'follow',
  })
}
