import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import { getPublicSettings } from '@/api/auth'
import { DEFAULT_AUTH_BIND_REDIRECT_PATH, DEFAULT_AUTH_REDIRECT_PATH } from '@/utils/authRedirect'
import { resolveAuthRouteDefaultsFromShellDefaults } from '@/router/setupRedirect'
import {
  renderAuthShellText,
  resolveAuthShellConfig,
  type AuthShellConfig,
  type AuthShellLabelKey,
  type AuthShellLabels,
} from '@/utils/authShell'
import { resolveRuntimeLocale } from '@/utils/runtimeLocale'

export function useAuthShellText() {
  const { locale } = useI18n()
  const authShellLabels = ref<AuthShellLabels>({})
  const authShellDefaults = ref<AuthShellConfig['defaults']>({})
  const authRouteDefaults = computed(() => resolveAuthRouteDefaultsFromShellDefaults(authShellDefaults.value))
  const defaultRedirectPath = computed(
    () => authShellDefaults.value.defaultRedirectPath || DEFAULT_AUTH_REDIRECT_PATH,
  )
  const defaultBindRedirectPath = computed(
    () => authShellDefaults.value.bindRedirectPath || DEFAULT_AUTH_BIND_REDIRECT_PATH,
  )

  function authText(key: AuthShellLabelKey, params: Record<string, string | number> = {}): string {
    return renderAuthShellText(authShellLabels.value, key, params)
  }

  function applyAuthShellConfig(rawAuthShellConfig?: string): AuthShellConfig['defaults'] {
    const config = resolveAuthShellConfig(rawAuthShellConfig, resolveRuntimeLocale(locale))
    authShellLabels.value = config.labels
    authShellDefaults.value = config.defaults
    return config.defaults
  }

  async function loadAuthShellLabels(): Promise<void> {
    await loadAuthShellConfig()
  }

  async function loadAuthShellConfig(): Promise<AuthShellConfig['defaults']> {
    try {
      const settings = await getPublicSettings()
      return applyAuthShellConfig(settings.auth_shell_config)
    } catch (error) {
      console.error('Failed to load auth shell settings:', error)
      authShellDefaults.value = {}
      authShellLabels.value = {}
      return {}
    }
  }

  return {
    authText,
    authShellLabels,
    applyAuthShellConfig,
    authShellDefaults,
    authRouteDefaults,
    defaultRedirectPath,
    defaultBindRedirectPath,
    loadAuthShellConfig,
    loadAuthShellLabels,
  }
}
