import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { readFileSync } from 'node:fs'
import LegalDocumentView from '../LegalDocumentView.vue'

const legalDocumentViewSource = readFileSync('src/views/public/LegalDocumentView.vue', 'utf8')

const currentLocale = vi.hoisted(() => ({ value: 'en' }))
const { fetchPublicSettings, appStoreState } = vi.hoisted(() => ({
  fetchPublicSettings: vi.fn(),
  appStoreState: {
    publicSettingsLoaded: true,
    cachedPublicSettings: null as any,
  },
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    ...appStoreState,
    fetchPublicSettings,
  }),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => ({
      params: { documentId: 'terms' },
    }),
  }
})

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      locale: currentLocale,
    }),
  }
})

describe('LegalDocumentView', () => {
  beforeEach(() => {
    fetchPublicSettings.mockReset()
    currentLocale.value = 'en'
    appStoreState.publicSettingsLoaded = true
    appStoreState.cachedPublicSettings = null
  })

  it('renders legal document shell labels from public settings', async () => {
    appStoreState.cachedPublicSettings = {
      site_name: 'cloudbase',
      site_logo: '',
      login_agreement_updated_at: '2026-06-18',
      auth_shell_config: JSON.stringify({
        en: {
          defaults: {
            loginPath: '/configured-login',
          },
        },
      }),
      legal_document_shell_config: JSON.stringify({
        en: {
          labels: {
            login: 'Configured login',
            agreementLabel: 'Configured agreement',
            updatedAt: 'Configured updated {date}',
            emptyContent: 'Configured empty',
          },
        },
      }),
      login_agreement_documents: [
        {
          id: 'terms',
          title: 'Configured Terms',
          content_md: '',
        },
      ],
    }

    const wrapper = mount(LegalDocumentView, {
      global: {
        stubs: {
          RouterLink: {
            props: ['to'],
            template: '<a :href="typeof to === \'string\' ? to : to.path"><slot /></a>',
          },
          Icon: { template: '<i />' },
        },
      },
    })
    await flushPromises()

    expect(wrapper.text()).toContain('cloudbase')
    expect(wrapper.text()).toContain('Configured login')
    expect(wrapper.text()).toContain('Configured agreement')
    expect(wrapper.text()).toContain('Configured Terms')
    expect(wrapper.text()).toContain('Configured updated 2026-06-18')
    expect(wrapper.text()).toContain('Configured empty')
  })

  it('fetches public settings through the app store when cache is cold', async () => {
    appStoreState.publicSettingsLoaded = false
    fetchPublicSettings.mockResolvedValue({})

    mount(LegalDocumentView, {
      global: {
        stubs: {
          RouterLink: true,
          Icon: true,
        },
      },
    })
    await flushPromises()

    expect(fetchPublicSettings).toHaveBeenCalledTimes(1)
  })

  it('does not keep locale-specific legal document fallback copy in the view bootstrap layer', () => {
    expect(legalDocumentViewSource).toContain('useAppStore')
    expect(legalDocumentViewSource).toContain('useAuthRouteDefaults')
    expect(legalDocumentViewSource).not.toContain('getPublicSettings')
    expect(legalDocumentViewSource).toContain(':to="authRouteDefaults.homePath"')
    expect(legalDocumentViewSource).not.toContain('to="/home"')
    expect(legalDocumentViewSource).toContain(':to="loginPath"')
    expect(legalDocumentViewSource).not.toContain('to="/login"')
    expect(legalDocumentViewSource).not.toContain('EMPTY_LEGAL_DOCUMENT_COPY')
    expect(legalDocumentViewSource).not.toContain('DEFAULT_LEGAL_DOCUMENT_COPY')
    expect(legalDocumentViewSource).not.toContain("login: '登录'")
    expect(legalDocumentViewSource).not.toContain("agreementLabel: '登录条款'")
    expect(legalDocumentViewSource).not.toContain("missingTitle: '文档不存在'")
    expect(legalDocumentViewSource).not.toContain("login: 'Log in'")
    expect(legalDocumentViewSource).not.toContain("agreementLabel: 'Login agreement'")
    expect(legalDocumentViewSource).not.toContain("missingTitle: 'Document not found'")
    expect(legalDocumentViewSource).not.toContain('type LegalDocumentCopy =')
    expect(legalDocumentViewSource).not.toContain('const legalDocumentCopyKeys')
    expect(legalDocumentViewSource).not.toContain('function formatTemplate')
    expect(legalDocumentViewSource).not.toContain('resolveLocalizedShellLabels(')
    expect(legalDocumentViewSource).toContain('resolveLegalDocumentCopy')
    expect(legalDocumentViewSource).toContain('formatLegalDocumentTemplate')
    expect(legalDocumentViewSource).toContain("from './legalDocumentRuntime'")
    expect(legalDocumentViewSource).toContain('resolveCurrentLegalDocument')
    expect(legalDocumentViewSource).toContain('resolveLegalDocumentIcon')
    expect(legalDocumentViewSource).toContain('renderLegalDocumentHtml')
  })
})
