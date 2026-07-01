import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

const hotContentViewSource = readFileSync(resolve(process.cwd(), 'src/views/public/HotContentView.vue'), 'utf8')
const zhLocaleSource = readFileSync(resolve(process.cwd(), 'src/i18n/locales/zh.ts'), 'utf8')
const enLocaleSource = readFileSync(resolve(process.cwd(), 'src/i18n/locales/en.ts'), 'utf8')

describe('HotContentView i18n integration', () => {
  it('uses vue-i18n t() function for all UI strings', () => {
    expect(hotContentViewSource).toContain("const { t } = useI18n()")
    expect(hotContentViewSource).toContain("t('hotContent.title')")
    expect(hotContentViewSource).toContain("t('hotContent.subtitle')")
    expect(hotContentViewSource).toContain("t('hotContent.search')")
    expect(hotContentViewSource).toContain("t('hotContent.emptyItems')")
  })

  it('has no hardcoded Chinese strings in the view', () => {
    // Should not contain any Chinese characters outside of comments/imports
    const chinesePattern = /[一-鿿]/
    const matches = hotContentViewSource.match(chinesePattern)
    expect(matches).toBeNull()
  })

  it('defines hotContent section in both locale files', () => {
    expect(zhLocaleSource).toContain('hotContent: {')
    expect(zhLocaleSource).toContain("title: '热点追踪'")
    expect(zhLocaleSource).toContain("tabItems: '主热点'")

    expect(enLocaleSource).toContain('hotContent: {')
    expect(enLocaleSource).toContain("title: 'Hot Topic Tracking'")
    expect(enLocaleSource).toContain("tabItems: 'Hot Items'")
  })
})

describe('HotContentView brand navigation', () => {
  it('links the header brand to the configured home path', () => {
    expect(hotContentViewSource).toContain('PublicDarkHeader')
    expect(hotContentViewSource).toContain(':account-label="t(\'hotContent.goConsole\')"')
    expect(hotContentViewSource).toContain('container-class="max-w-6xl"')
    expect(hotContentViewSource).toContain('public-template-container')
    expect(hotContentViewSource).not.toContain('public-template-container-wide')
    expect(hotContentViewSource).not.toContain('useAuthRouteDefaults')
    expect(hotContentViewSource).not.toContain(':to="authRouteDefaults.homePath"')
    expect(hotContentViewSource).not.toContain('to="/home"')
  })

  it('shows dashboard and avatar actions in the header', () => {
    expect(hotContentViewSource).toContain("t('hotContent.goConsole')")
    expect(hotContentViewSource).not.toContain(':to="isAuthenticated ? dashboardPath : loginPath"')
    expect(hotContentViewSource).not.toContain('const dashboardPath = computed(() => resolveHomePath(authStore.isAdmin))')
    expect(hotContentViewSource).not.toContain('const avatarUrl = computed(() => authStore.user?.avatar_url?.trim() || \'\')')
    expect(hotContentViewSource).not.toContain('const userInitial = computed(() => displayName.value.charAt(0).toUpperCase() || \'U\')')
  })
})

describe('HotContentView hero summary cleanup', () => {
  it('does not render the removed stats cards or capability status panel', () => {
    expect(hotContentViewSource).toContain('itemTotal')
    expect(hotContentViewSource).not.toContain('v-for="stat in stats"')
    expect(hotContentViewSource).not.toContain('const stats = computed')
    expect(hotContentViewSource).not.toContain("t('hotContent.capabilityStatus')")
    expect(hotContentViewSource).not.toContain("t('hotContent.capabilityMain')")
    expect(hotContentViewSource).not.toContain("t('hotContent.capabilitySources'")
    expect(hotContentViewSource).not.toContain("t('hotContent.statSources')")
    expect(hotContentViewSource).not.toContain("t('hotContent.statHot')")
    expect(hotContentViewSource).not.toContain("t('hotContent.statDaily')")
    expect(hotContentViewSource).not.toContain("t('hotContent.statMp')")
  })

  it('no longer uses visible item array length for hero stats', () => {
    expect(hotContentViewSource).not.toContain('items.value.length')
    expect(hotContentViewSource).not.toContain('dailyIssues.value.length')
    expect(hotContentViewSource).not.toContain('mpEntries.value.length')
  })
})

describe('HotContentView loading state', () => {
  it('has a loading spinner UI', () => {
    expect(hotContentViewSource).toContain('v-if="loading"')
    expect(hotContentViewSource).toContain('animate-spin')
    expect(hotContentViewSource).toContain("t('hotContent.loading')")
  })

  it('disables buttons during loading', () => {
    expect(hotContentViewSource).toContain(':disabled="loading"')
  })

  it('hides content sections when loading', () => {
    expect(hotContentViewSource).toContain('v-if="!loading"')
    expect(hotContentViewSource).not.toContain('v-if="!loading && activeTab === \'items\'"')
    expect(hotContentViewSource).not.toContain('v-else-if="!loading && activeTab === \'daily\'"')
    expect(hotContentViewSource).not.toContain('v-else-if="!loading && activeTab === \'mp\'"')
  })

  it('adapts loading spinner color for light/dark tabs', () => {
    expect(hotContentViewSource).not.toContain('isLightTab')
    expect(hotContentViewSource).toContain('text-white/40')
  })
})

describe('HotContentView events tab search fix', () => {
  it('always shows the hot stream search bar', () => {
    expect(hotContentViewSource).toContain('class="flex flex-col gap-3 sm:flex-row"')
    expect(hotContentViewSource).not.toContain('v-if="activeTab !== \'events\'"')
  })

  it('does not render the run event input in the page template', () => {
    expect(hotContentViewSource).not.toContain('v-model="runID"')
    expect(hotContentViewSource).not.toContain("t('hotContent.eventsPlaceholder')")
    expect(hotContentViewSource).not.toContain("t('hotContent.loadEvents')")
  })
})

describe('HotContentView hot stream only layout', () => {
  it('removes the left tab navigation and keeps the content full width', () => {
    expect(hotContentViewSource).toContain('<!-- Hot stream -->')
    expect(hotContentViewSource).toContain('<div class="mt-10">')
    expect(hotContentViewSource).not.toContain('lg:grid-cols-[240px_minmax(0,1fr)]')
    expect(hotContentViewSource).not.toContain('v-for="tab in tabs"')
    expect(hotContentViewSource).not.toContain('@click="activeTab = tab.id"')
  })
})

describe('HotContentView EmptyState theme adaptation', () => {
  it('uses the dark EmptyState for the hot stream', () => {
    expect(hotContentViewSource).toContain('EmptyState v-if="items.length === 0" :text="t(\'hotContent.emptyItems\')"')
    expect(hotContentViewSource).not.toContain('EmptyState v-if="dailyIssues.length === 0"')
    expect(hotContentViewSource).not.toContain('EmptyState v-if="mpEntries.length === 0"')
  })

  it('EmptyState component supports theme prop', () => {
    expect(hotContentViewSource).toContain('theme: {')
    expect(hotContentViewSource).toContain('type: String as () => \'dark\' | \'light\'')
    expect(hotContentViewSource).toContain('props.theme === \'light\'')
  })
})

describe('HotContentView deterministic date formatting', () => {
  it('uses manual date formatting instead of toLocaleString', () => {
    expect(hotContentViewSource).toContain('function formatDate')
    expect(hotContentViewSource).toContain('date.getFullYear()')
    expect(hotContentViewSource).toContain('date.getMonth()')
    expect(hotContentViewSource).toContain('String(date.getDate()).padStart')
    expect(hotContentViewSource).not.toContain('toLocaleString()')
  })
})

describe('HotContentView parallel first screen loading', () => {
  it('uses Promise.all for parallel source and content loading', () => {
    expect(hotContentViewSource).toContain('Promise.all([loadSources(), refreshActiveTab()])')
  })
})

describe('HotContentView error banner theme adaptation', () => {
  it('uses the hot stream dark error banner', () => {
    expect(hotContentViewSource).toContain('border-red-400/30 bg-red-900/20 p-4 text-sm text-red-300')
    expect(hotContentViewSource).not.toContain('isLightTab')
  })
})

describe('HotContentView cleanup', () => {
  it('removed unused selectedSource ref', () => {
    expect(hotContentViewSource).not.toContain('selectedSource')
  })

  it('does not keep hidden daily/mp/feed API loaders in the view', () => {
    expect(hotContentViewSource).not.toContain('listHotDailyIssues')
    expect(hotContentViewSource).not.toContain('getHotDailyIssue')
    expect(hotContentViewSource).not.toContain('listHotMPEntries')
    expect(hotContentViewSource).not.toContain('listHotRunEvents')
    expect(hotContentViewSource).not.toContain('activeTab')
  })
})

describe('HotContentView pagination', () => {
  it('has item pagination state and retained loader page refs', () => {
    expect(hotContentViewSource).toContain('itemPage')
    expect(hotContentViewSource).toContain('itemTotalPages')
    expect(hotContentViewSource).not.toContain('dailyPage')
    expect(hotContentViewSource).not.toContain('mpPage')
    expect(hotContentViewSource).not.toContain('dailyTotalPages')
    expect(hotContentViewSource).not.toContain('mpTotalPages')
  })

  it('has main item pagination navigation function', () => {
    expect(hotContentViewSource).toContain('goToItemPage')
    expect(hotContentViewSource).not.toContain('goToDailyPage')
    expect(hotContentViewSource).not.toContain('goToMPPage')
  })

  it('has visible pages helper function', () => {
    expect(hotContentViewSource).toContain('getVisiblePages')
  })

  it('resets page on search', () => {
    expect(hotContentViewSource).toContain('searchAndRefresh')
    expect(hotContentViewSource).toContain('itemPage.value = 1')
    expect(hotContentViewSource).not.toContain('dailyPage.value = 1')
    expect(hotContentViewSource).not.toContain('mpPage.value = 1')
  })

  it('uses page refs in load functions', () => {
    expect(hotContentViewSource).toContain('page: itemPage.value')
    expect(hotContentViewSource).not.toContain('page: dailyPage.value')
    expect(hotContentViewSource).not.toContain('page: mpPage.value')
  })

  it('has pagination i18n keys in locales', () => {
    expect(zhLocaleSource).toContain('paginationPrev: \'上一页\'')
    expect(zhLocaleSource).toContain('paginationNext: \'下一页\'')
    expect(zhLocaleSource).toContain('paginationInfo:')
    expect(enLocaleSource).toContain('paginationPrev: \'Previous\'')
    expect(enLocaleSource).toContain('paginationNext: \'Next\'')
    expect(enLocaleSource).toContain('paginationInfo:')
  })

  it('renders pagination UI for items tab', () => {
    expect(hotContentViewSource).toContain('v-if="itemTotalPages > 1"')
    expect(hotContentViewSource).toContain('goToItemPage(itemPage - 1)')
    expect(hotContentViewSource).toContain('goToItemPage(itemPage + 1)')
  })

  it('does not render stale daily tab pagination UI', () => {
    expect(hotContentViewSource).not.toContain('v-if="dailyTotalPages > 1"')
    expect(hotContentViewSource).not.toContain('goToDailyPage(dailyPage - 1)')
    expect(hotContentViewSource).not.toContain('goToDailyPage(dailyPage + 1)')
  })

  it('does not render stale mp tab pagination UI', () => {
    expect(hotContentViewSource).not.toContain('v-if="mpTotalPages > 1"')
    expect(hotContentViewSource).not.toContain('goToMPPage(mpPage - 1)')
    expect(hotContentViewSource).not.toContain('goToMPPage(mpPage + 1)')
  })

  it('pagination adapts to dark theme for items tab', () => {
    expect(hotContentViewSource).toContain('border-white/10')
    expect(hotContentViewSource).toContain('text-white/60')
    expect(hotContentViewSource).toContain('border-cyan-200/40 bg-cyan-200/15 text-cyan-50')
  })

  it('does not retain light-theme pagination from removed daily/mp tab panels', () => {
    expect(hotContentViewSource).not.toContain('bg-white/70')
    expect(hotContentViewSource).not.toContain('border-[#b35f25] bg-[#f5b85a]/20 text-[#9d602a]')
  })
})
