<template>
  <div v-if="siteKey" class="turnstile-wrapper">
    <div ref="containerRef" class="turnstile-container"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { resolveCSPNonce } from '@/utils/cspNonce'

interface TurnstileRenderOptions {
  sitekey: string
  callback: (token: string) => void
  'expired-callback'?: () => void
  'error-callback'?: () => void
  'timeout-callback'?: () => void
  theme?: 'light' | 'dark' | 'auto'
  size?: 'normal' | 'compact' | 'flexible'
  retry?: 'auto' | 'never'
  'refresh-expired'?: 'auto' | 'manual' | 'never'
  'refresh-timeout'?: 'auto' | 'manual' | 'never'
}

interface TurnstileAPI {
  render: (container: HTMLElement, options: TurnstileRenderOptions) => string
  reset: (widgetId?: string) => void
  remove: (widgetId?: string) => void
}

declare global {
  interface Window {
    turnstile?: TurnstileAPI
  }
}

const props = withDefaults(
  defineProps<{
    siteKey: string
    theme?: 'light' | 'dark' | 'auto'
    size?: 'normal' | 'compact' | 'flexible'
  }>(),
  {
    theme: 'auto',
    size: 'flexible'
  }
)

const emit = defineEmits<{
  (e: 'verify', token: string): void
  (e: 'expire'): void
  (e: 'error'): void
}>()

const containerRef = ref<HTMLElement | null>(null)
const widgetId = ref<string | null>(null)
const scriptLoaded = ref(false)
const turnstileScriptSrc = 'https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit'
const turnstileScriptSelector = 'script[data-turnstile-script="true"]'

let scriptLoadPromise: Promise<void> | null = null

const loadScript = (): Promise<void> => {
  if (scriptLoadPromise) {
    return scriptLoadPromise
  }

  scriptLoadPromise = new Promise((resolve, reject) => {
    if (window.turnstile) {
      scriptLoaded.value = true
      resolve()
      return
    }

    const existingScript = document.querySelector<HTMLScriptElement>(turnstileScriptSelector)
    if (existingScript) {
      existingScript.addEventListener(
        'load',
        () => {
          scriptLoaded.value = true
          resolve()
        },
        { once: true }
      )
      existingScript.addEventListener(
        'error',
        () => {
          scriptLoadPromise = null
          reject(new Error('Failed to load Turnstile script'))
        },
        { once: true }
      )
      return
    }

    const script = document.createElement('script')
    script.src = turnstileScriptSrc
    script.async = true
    script.defer = true
    script.dataset.turnstileScript = 'true'
    const nonce = resolveCSPNonce(document)
    if (nonce) {
      script.nonce = nonce
      script.setAttribute('nonce', nonce)
    }

    script.onload = () => {
      scriptLoaded.value = true
      resolve()
    }

    script.onerror = () => {
      scriptLoadPromise = null
      reject(new Error('Failed to load Turnstile script'))
    }

    document.head.appendChild(script)
  })

  return scriptLoadPromise
}

const renderWidget = () => {
  if (!window.turnstile || !containerRef.value || !props.siteKey) {
    return
  }

  // Remove existing widget if any
  if (widgetId.value) {
    try {
      window.turnstile.remove(widgetId.value)
    } catch {
      // Ignore errors when removing
    }
    widgetId.value = null
  }

  // Clear container
  containerRef.value.innerHTML = ''

  widgetId.value = window.turnstile.render(containerRef.value, {
    sitekey: props.siteKey,
    callback: (token: string) => {
      emit('verify', token)
    },
    'expired-callback': () => {
      emit('expire')
    },
    'error-callback': () => {
      emit('error')
    },
    'timeout-callback': () => {
      emit('expire')
    },
    theme: props.theme,
    size: props.size,
    retry: 'never',
    'refresh-expired': 'manual',
    'refresh-timeout': 'manual'
  })
}

const reset = () => {
  if (window.turnstile && widgetId.value) {
    window.turnstile.reset(widgetId.value)
  }
}

// Expose reset method to parent
defineExpose({ reset })

onMounted(async () => {
  if (!props.siteKey) {
    return
  }

  try {
    await loadScript()
    renderWidget()
  } catch (error) {
    console.error('Failed to initialize Turnstile:', error)
    emit('error')
  }
})

onUnmounted(() => {
  if (window.turnstile && widgetId.value) {
    try {
      window.turnstile.remove(widgetId.value)
    } catch {
      // Ignore errors when removing
    }
  }
})

// Re-render when siteKey changes
watch(
  () => props.siteKey,
  (newKey) => {
    if (newKey && scriptLoaded.value) {
      renderWidget()
    }
  }
)
</script>

<style scoped>
.turnstile-wrapper {
  width: 100%;
}

.turnstile-container {
  width: 100%;
  min-height: 65px;
}

/* Make the Turnstile iframe fill the container width */
.turnstile-container :deep(iframe) {
  width: 100% !important;
}
</style>
