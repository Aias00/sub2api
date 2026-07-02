import { afterEach, describe, expect, it } from 'vitest'

import {
  applyPublicIntegrations,
  clearPublicIntegrations
} from '@/utils/publicIntegrations'
import type { PublicSettings } from '@/types'

function managedNodes() {
  return [...document.querySelectorAll('[data-cloudbase-public-integration]')]
}

describe('publicIntegrations', () => {
  afterEach(() => {
    clearPublicIntegrations()
  })

  it('injects enabled integrations from public settings', () => {
    applyPublicIntegrations({
      google_analytics_id: 'G-CLOUDBASE',
      adsense_code: 'ca-pub-cloudbase',
      affonso_enabled: true,
      affonso_id: 'affonso-public',
      affonso_cookie_duration: '45',
      crisp_enabled: true,
      crisp_website_id: 'crisp-public'
    } as PublicSettings)

    expect(document.querySelector('script#google-analytics-loader')?.getAttribute('src')).toBe(
      'https://www.googletagmanager.com/gtag/js?id=G-CLOUDBASE'
    )
    expect(document.querySelector('meta[name="google-adsense-account"]')?.getAttribute('content')).toBe(
      'ca-pub-cloudbase'
    )
    expect(document.querySelector('script#affonso')?.getAttribute('data-affonso')).toBe(
      'affonso-public'
    )
    expect(document.querySelector('script#affonso')?.getAttribute('data-cookie_duration')).toBe('45')
    expect(document.querySelector('script#crisp')?.textContent).toContain('crisp-public')
  })

  it('does not synthesize an Affonso cookie duration fallback in the frontend', () => {
    applyPublicIntegrations({
      affonso_enabled: true,
      affonso_id: 'affonso-public'
    } as PublicSettings)

    expect(document.querySelector('script#affonso')?.getAttribute('data-cookie_duration')).toBe('')
  })

  it('replaces old integration nodes when settings change', () => {
    applyPublicIntegrations({
      google_analytics_id: 'G-OLD',
      tawk_enabled: true,
      tawk_property_id: 'old-property',
      tawk_widget_id: 'old-widget'
    } as PublicSettings)

    applyPublicIntegrations({
      google_analytics_id: 'G-NEW',
      promotekit_enabled: true,
      promotekit_id: 'promotekit-public'
    } as PublicSettings)

    expect(managedNodes().filter((node) => node.id === 'google-analytics-loader')).toHaveLength(1)
    expect(document.querySelector('script#google-analytics-loader')?.getAttribute('src')).toContain('G-NEW')
    expect(document.querySelector('script#tawk')).toBeNull()
    expect(document.querySelector('script#promotekit')?.getAttribute('data-promotekit')).toBe(
      'promotekit-public'
    )
  })

  it('clears managed nodes when integration injection is disabled', () => {
    applyPublicIntegrations({
      clarity_id: 'clarity-public'
    } as PublicSettings)

    applyPublicIntegrations({
      clarity_id: 'clarity-public'
    } as PublicSettings, { enabled: false })

    expect(managedNodes()).toHaveLength(0)
  })

  it('ignores legacy Touch integration settings', () => {
    applyPublicIntegrations({
      touch_google_analytics_id: 'G-TOUCH',
      touch_adsense_code: 'ca-pub-touch',
      touch_crisp_enabled: true,
      touch_crisp_website_id: 'crisp-touch'
    } as PublicSettings)

    expect(managedNodes()).toHaveLength(0)
  })
})
