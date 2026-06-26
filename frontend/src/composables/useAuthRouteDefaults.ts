import { computed } from 'vue'

import { useAppStore } from '@/stores'
import { resolveAuthRouteDefaults, resolveRoleHomeRedirect } from '@/router/setupRedirect'

export function useAuthRouteDefaults() {
  const appStore = useAppStore()
  const defaults = computed(() => resolveAuthRouteDefaults(appStore.cachedPublicSettings?.auth_shell_config))

  function resolveHomePath(isAdmin: boolean): string {
    return resolveRoleHomeRedirect(isAdmin, defaults.value)
  }

  return {
    authRouteDefaults: defaults,
    resolveHomePath,
  }
}
