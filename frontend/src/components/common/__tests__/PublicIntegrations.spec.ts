import { readFileSync } from 'node:fs'

import { mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'

import PublicIntegrations from '../PublicIntegrations.vue'

const componentSource = readFileSync('src/components/common/PublicIntegrations.vue', 'utf8')
const appStoreState = vi.hoisted(() => ({
  cachedPublicSettings: null as null | {
    public_integrations_enabled?: boolean
    google_analytics_id?: string
  },
}))

vi.mock('@/stores', () => ({
  useAppStore: () => appStoreState,
}))

function managedNodes() {
  return [...document.querySelectorAll('[data-sub2api-public-integration]')]
}

describe('PublicIntegrations', () => {
  beforeEach(() => {
    appStoreState.cachedPublicSettings = null
    document.head.querySelectorAll('[data-sub2api-public-integration]').forEach((node) => node.remove())
    document.body.querySelectorAll('[data-sub2api-public-integration]').forEach((node) => node.remove())
  })

  afterEach(() => {
    document.head.querySelectorAll('[data-sub2api-public-integration]').forEach((node) => node.remove())
    document.body.querySelectorAll('[data-sub2api-public-integration]').forEach((node) => node.remove())
  })

  it('injects integrations when public settings leave the switch enabled', async () => {
    appStoreState.cachedPublicSettings = {
      public_integrations_enabled: true,
      google_analytics_id: 'G-PUBLIC',
    }

    const wrapper = mount(PublicIntegrations)
    await nextTick()

    expect(document.querySelector('script#google-analytics-loader')?.getAttribute('src')).toContain('G-PUBLIC')

    wrapper.unmount()
    expect(managedNodes()).toHaveLength(0)
  })

  it('clears integrations when Sub2API public settings disable injection', async () => {
    appStoreState.cachedPublicSettings = {
      public_integrations_enabled: false,
      google_analytics_id: 'G-DISABLED',
    }

    mount(PublicIntegrations)
    await nextTick()

    expect(managedNodes()).toHaveLength(0)
  })

  it('does not depend on frontend env for integration injection', () => {
    expect(componentSource).not.toContain('VITE_PUBLIC_INTEGRATIONS_DEBUG')
    expect(componentSource).not.toContain('import.meta.env.PROD')
    expect(componentSource).toContain('public_integrations_enabled')
  })
})
