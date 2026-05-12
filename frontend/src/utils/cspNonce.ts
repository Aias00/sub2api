export function resolveCSPNonce(doc: Document = document): string {
  const nonceScript = doc.querySelector('script[nonce]') as HTMLScriptElement | null
  if (nonceScript?.nonce) {
    return nonceScript.nonce
  }
  const metaNonce = doc.querySelector('meta[name="csp-nonce"]')
  const content = metaNonce?.getAttribute('content')?.trim() || ''
  return content
}
