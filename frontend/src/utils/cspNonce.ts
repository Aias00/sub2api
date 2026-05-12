export function resolveCSPNonce(doc: Document = document): string {
  for (const script of Array.from(doc.scripts)) {
    if (script.nonce) {
      return script.nonce
    }
  }
  const metaNonce = doc.querySelector('meta[name="csp-nonce"]')
  const content = metaNonce?.getAttribute('content')?.trim() || ''
  return content
}
