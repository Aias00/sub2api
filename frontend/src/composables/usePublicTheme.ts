import { computed, ref } from 'vue'

export type PublicTheme = 'light' | 'dark'

const PUBLIC_THEME_STORAGE_KEY = 'public-theme'
const DEFAULT_PUBLIC_THEME: PublicTheme = 'dark'

function normalizePublicTheme(value: string | null | undefined): PublicTheme {
  return value === 'light' || value === 'dark' ? value : DEFAULT_PUBLIC_THEME
}

function canUseDocument() {
  return typeof document !== 'undefined'
}

function canUseLocalStorage() {
  return typeof localStorage !== 'undefined'
}

export function getStoredPublicTheme(): PublicTheme {
  if (!canUseLocalStorage()) return DEFAULT_PUBLIC_THEME
  try {
    return normalizePublicTheme(localStorage.getItem(PUBLIC_THEME_STORAGE_KEY))
  } catch {
    return DEFAULT_PUBLIC_THEME
  }
}

export function applyPublicTheme(theme: PublicTheme) {
  if (canUseDocument()) {
    document.documentElement.dataset.publicTheme = theme
    document.documentElement.classList.toggle('dark', theme === 'dark')
  }
  if (canUseLocalStorage()) {
    try {
      localStorage.setItem(PUBLIC_THEME_STORAGE_KEY, theme)
    } catch {
      // Ignore storage failures; the active document still receives the theme.
    }
  }
}

export function initPublicTheme() {
  applyPublicTheme(getStoredPublicTheme())
}

export function usePublicTheme() {
  const theme = ref<PublicTheme>(getStoredPublicTheme())
  applyPublicTheme(theme.value)

  const isDarkTheme = computed(() => theme.value === 'dark')
  const nextTheme = computed<PublicTheme>(() => (isDarkTheme.value ? 'light' : 'dark'))

  function setPublicTheme(next: PublicTheme) {
    theme.value = next
    applyPublicTheme(next)
  }

  function togglePublicTheme() {
    setPublicTheme(nextTheme.value)
  }

  return {
    theme,
    isDarkTheme,
    nextTheme,
    setPublicTheme,
    togglePublicTheme,
  }
}
