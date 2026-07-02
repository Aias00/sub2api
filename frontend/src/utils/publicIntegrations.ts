import type { PublicSettings } from '@/types'

const INTEGRATION_ATTR = 'data-cloudbase-public-integration'

type NodeTarget = 'head' | 'body'

type IntegrationNode = {
  id: string
  target: NodeTarget
  create: () => HTMLElement
}

function text(value: unknown): string {
  return typeof value === 'string' ? value.trim() : ''
}

function enabled(value: unknown): boolean {
  if (typeof value === 'boolean') return value
  if (typeof value === 'string') return value.trim().toLowerCase() === 'true'
  return false
}

function script(id: string, attrs: Record<string, string | boolean | number>): HTMLScriptElement {
  const element = document.createElement('script')
  element.id = id
  for (const [key, value] of Object.entries(attrs)) {
    if (typeof value === 'boolean') {
      if (value) element.setAttribute(key, '')
      continue
    }
    element.setAttribute(key, String(value))
  }
  return element
}

function inlineScript(id: string, code: string): HTMLScriptElement {
  const element = document.createElement('script')
  element.id = id
  element.textContent = code
  return element
}

function meta(name: string, content: string): HTMLMetaElement {
  const element = document.createElement('meta')
  element.name = name
  element.content = content
  return element
}

function mark(node: HTMLElement, id: string): HTMLElement {
  node.setAttribute(INTEGRATION_ATTR, id)
  return node
}

export function clearPublicIntegrations(doc: Document = document): void {
  doc.querySelectorAll(`[${INTEGRATION_ATTR}]`).forEach((node) => node.remove())
}

export function buildPublicIntegrationNodes(settings: PublicSettings): IntegrationNode[] {
  const nodes: IntegrationNode[] = []

  const gaId = text(settings.google_analytics_id)
  if (gaId) {
    nodes.push({
      id: 'google-analytics-loader',
      target: 'head',
      create: () => script('google-analytics-loader', {
        async: true,
        src: `https://www.googletagmanager.com/gtag/js?id=${gaId}`
      })
    })
    nodes.push({
      id: 'google-analytics-init',
      target: 'head',
      create: () => inlineScript('google-analytics-init', `
        window.dataLayer = window.dataLayer || [];
        function gtag(){dataLayer.push(arguments);}
        gtag('js', new Date());
        gtag('config', '${gaId}');
      `)
    })
  }

  const clarityId = text(settings.clarity_id)
  if (clarityId) {
    nodes.push({
      id: 'clarity',
      target: 'head',
      create: () => inlineScript('clarity', `
        (function(c,l,a,r,i,t,y){
          c[a]=c[a]||function(){(c[a].q=c[a].q||[]).push(arguments)};
          t=l.createElement(r);t.async=1;t.src="https://www.clarity.ms/tag/"+i;
          y=l.getElementsByTagName(r)[0];y.parentNode.insertBefore(t,y);
        })(window, document, "clarity", "script", "${clarityId}");
      `)
    })
  }

  const plausibleDomain = text(settings.plausible_domain)
  const plausibleSrc = text(settings.plausible_src)
  if (plausibleDomain && plausibleSrc) {
    nodes.push({
      id: 'plausible-init',
      target: 'head',
      create: () => inlineScript('plausible-init', `
        window.plausible = window.plausible || function() {
          (window.plausible.q = window.plausible.q || []).push(arguments)
        }
      `)
    })
    nodes.push({
      id: 'plausible-loader',
      target: 'head',
      create: () => script('plausible-loader', {
        defer: true,
        async: true,
        src: plausibleSrc,
        'data-domain': plausibleDomain
      })
    })
  }

  const openPanelClientId = text(settings.openpanel_client_id)
  if (openPanelClientId) {
    nodes.push({
      id: 'openpanel-init',
      target: 'head',
      create: () => inlineScript('openpanel-init', `
        window.op = window.op || function(...args){(window.op.q = window.op.q || []).push(args);};
        window.op('init', {
          clientId: '${openPanelClientId}',
          trackScreenViews: true,
          trackOutgoingLinks: true,
          trackAttributes: true
        });
      `)
    })
    nodes.push({
      id: 'openpanel-loader',
      target: 'head',
      create: () => script('openpanel-loader', {
        defer: true,
        async: true,
        src: 'https://openpanel.dev/op1.js'
      })
    })
  }

  if (enabled(settings.vercel_analytics_enabled)) {
    nodes.push({
      id: 'vercel-analytics',
      target: 'body',
      create: () => script('vercel-analytics', {
        defer: true,
        src: '/_vercel/insights/script.js'
      })
    })
  }

  const adsenseCode = text(settings.adsense_code)
  if (adsenseCode) {
    nodes.push({
      id: 'google-adsense-account',
      target: 'head',
      create: () => meta('google-adsense-account', adsenseCode)
    })
    nodes.push({
      id: 'google-adsense-loader',
      target: 'head',
      create: () => script('google-adsense-loader', {
        async: true,
        src: `https://pagead2.googlesyndication.com/pagead/js/adsbygoogle.js?client=${adsenseCode}`,
        crossorigin: 'anonymous'
      })
    })
  }

  const affonsoId = text(settings.affonso_id)
  if (enabled(settings.affonso_enabled) && affonsoId) {
    nodes.push({
      id: 'affonso',
      target: 'head',
      create: () => script('affonso', {
        async: true,
        defer: true,
        src: 'https://affonso.io/js/pixel.min.js',
        'data-affonso': affonsoId,
        'data-cookie_duration': text(settings.affonso_cookie_duration)
      })
    })
  }

  const promoteKitId = text(settings.promotekit_id)
  if (enabled(settings.promotekit_enabled) && promoteKitId) {
    nodes.push({
      id: 'promotekit',
      target: 'head',
      create: () => script('promotekit', {
        async: true,
        defer: true,
        src: 'https://cdn.promotekit.com/promotekit.js',
        'data-promotekit': promoteKitId
      })
    })
  }

  const crispWebsiteId = text(settings.crisp_website_id)
  if (enabled(settings.crisp_enabled) && crispWebsiteId) {
    nodes.push({
      id: 'crisp',
      target: 'head',
      create: () => inlineScript('crisp', `
        window.$crisp = [];
        window.CRISP_WEBSITE_ID = "${crispWebsiteId}";
        (function(){
          var d = document;
          var s = d.createElement("script");
          s.src = "https://client.crisp.chat/l.js";
          s.async = 1;
          d.getElementsByTagName("head")[0].appendChild(s);
        })();
      `)
    })
  }

  const tawkPropertyId = text(settings.tawk_property_id)
  const tawkWidgetId = text(settings.tawk_widget_id)
  if (enabled(settings.tawk_enabled) && tawkPropertyId && tawkWidgetId) {
    nodes.push({
      id: 'tawk',
      target: 'head',
      create: () => inlineScript('tawk', `
        var Tawk_API = Tawk_API || {}, Tawk_LoadStart = new Date();
        (function(){
          var s1 = document.createElement("script"), s0 = document.getElementsByTagName("script")[0];
          s1.async = true;
          s1.src = "https://embed.tawk.to/${tawkPropertyId}/${tawkWidgetId}";
          s1.charset = "UTF-8";
          s1.setAttribute("crossorigin", "*");
          s0.parentNode.insertBefore(s1, s0);
        })();
      `)
    })
  }

  return nodes
}

export function applyPublicIntegrations(
  settings: PublicSettings | null | undefined,
  options: { enabled?: boolean; doc?: Document } = {}
): void {
  const doc = options.doc ?? document
  clearPublicIntegrations(doc)

  if (!settings || options.enabled === false) {
    return
  }

  for (const node of buildPublicIntegrationNodes(settings)) {
    const target = node.target === 'body' ? doc.body : doc.head
    target.appendChild(mark(node.create(), node.id))
  }
}
