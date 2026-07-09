<template>
  <div class="locale-envelope-editor space-y-1.5">
    <label v-if="label" class="input-label mb-1.5 flex items-center gap-2">
      <span>{{ label }}</span>
      <span
        v-if="overallValidationState === 'valid'"
        class="inline-flex h-4 w-4 items-center justify-center rounded-full bg-emerald-100 text-[10px] font-bold text-emerald-600 dark:bg-emerald-900/40 dark:text-emerald-400"
      >✓</span>
      <span
        v-else-if="overallValidationState === 'invalid'"
        class="inline-flex h-4 w-4 items-center justify-center rounded-full bg-red-100 text-[10px] font-bold text-red-600 dark:bg-red-900/40 dark:text-red-400"
      >✗</span>
    </label>

    <!-- Fallback raw editor when envelope is malformed -->
    <JsonEditor
      v-if="!isEnvelope"
      :model-value="modelValue"
      :hint="hint"
      :error="error"
      :height="height"
      :disabled="disabled"
      :locale-envelope="false"
      @update:model-value="$emit('update:modelValue', $event)"
      @validation-error="onValidationError(null, $event)"
    />

    <!-- Tabbed editor for valid envelope structure -->
    <template v-else>
      <div class="flex items-center gap-1 border-b border-gray-200 dark:border-dark-600">
        <button
          v-for="locale in localeKeys"
          :key="locale"
          type="button"
          class="relative px-3 py-1.5 text-sm font-medium transition-colors duration-150"
          :class="[
            activeTab === locale
              ? 'text-primary-600 dark:text-primary-400'
              : 'text-gray-500 hover:text-gray-700 dark:text-dark-400 dark:hover:text-dark-200',
          ]"
          @click="activeTab = locale"
        >
          {{ locale.toUpperCase() }}
          <span
            v-if="tabValidation[locale] === 'invalid'"
            class="absolute -right-0.5 -top-0.5 h-2 w-2 rounded-full bg-red-500"
          />
          <span
            v-if="activeTab === locale"
            class="absolute bottom-0 left-0 right-0 h-0.5 bg-primary-600 dark:bg-primary-400"
          />
        </button>
      </div>

      <JsonEditor
        :key="activeTab"
        :model-value="activeTabContent"
        :hint="activeTab === 'en' ? hint : undefined"
        :error="error"
        :height="height"
        :disabled="disabled"
        @update:model-value="onTabContentChange(activeTab, $event)"
        @validation-error="onValidationError(activeTab, $event)"
      />
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import JsonEditor from './JsonEditor.vue'

interface Props {
  modelValue: string
  label?: string
  hint?: string
  error?: string
  height?: string
  disabled?: boolean
  localeKeys?: string[]
}

const props = withDefaults(defineProps<Props>(), {
  height: '300px',
  disabled: false,
  localeKeys: () => ['en', 'zh'],
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'validationError', message: string | null): void
}>()

const activeTab = ref(props.localeKeys[0] || 'en')
const tabValidation = reactive<Record<string, 'unknown' | 'valid' | 'invalid'>>({})

// Parse the envelope JSON
function parseEnvelope(value: string): Record<string, unknown> | null {
  if (!value || !value.trim()) return {}
  try {
    const parsed = JSON.parse(value)
    if (typeof parsed === 'object' && parsed !== null && !Array.isArray(parsed)) {
      return parsed as Record<string, unknown>
    }
    return null // Not an object = malformed envelope
  } catch {
    return null // Invalid JSON = malformed envelope
  }
}

// Whether the current value is a valid envelope (object or empty)
const isEnvelope = computed(() => {
  const parsed = parseEnvelope(props.modelValue)
  // Empty string -> treat as envelope (will initialize with empty locale objects)
  // Valid object -> envelope
  // Non-object or invalid JSON -> not envelope, show raw editor
  return parsed !== null
})

// Get the content for the active tab
const activeTabContent = computed(() => {
  const parsed = parseEnvelope(props.modelValue)
  if (!parsed) return props.modelValue

  const localeValue = parsed[activeTab.value]
  if (localeValue === undefined || localeValue === null) return '{}'
  if (typeof localeValue === 'string') return localeValue
  return JSON.stringify(localeValue, null, 2)
})

// Overall validation state
const overallValidationState = computed<'unknown' | 'valid' | 'invalid'>(() => {
  const states = Object.values(tabValidation)
  if (states.some(s => s === 'invalid')) return 'invalid'
  if (states.length > 0 && states.every(s => s === 'valid')) return 'valid'
  return 'unknown'
})

// When a tab's content changes, merge it back into the envelope
function onTabContentChange(locale: string, content: string) {
  const parsed = parseEnvelope(props.modelValue) || {}
  let localeValue: unknown
  try {
    localeValue = JSON.parse(content)
  } catch {
    // Content is not valid JSON yet, store as raw string
    localeValue = content
  }
  parsed[locale] = localeValue
  emit('update:modelValue', JSON.stringify(parsed, null, 2))
}

function onValidationError(locale: string | null, message: string | null) {
  if (locale) {
    tabValidation[locale] = message ? 'invalid' : 'valid'
  }
  emit('validationError', message)
}

// Initialize empty envelope when modelValue is empty
onMounted(() => {
  if (!props.modelValue || !props.modelValue.trim()) {
    const envelope: Record<string, string> = {}
    for (const key of props.localeKeys) {
      envelope[key] = '{}'
    }
    emit('update:modelValue', JSON.stringify(envelope, null, 2))
  }
  // Initialize tab validation
  for (const key of props.localeKeys) {
    tabValidation[key] = 'unknown'
  }
})

// Reset active tab if localeKeys change
watch(() => props.localeKeys, (newKeys) => {
  if (!newKeys.includes(activeTab.value)) {
    activeTab.value = newKeys[0] || 'en'
  }
})
</script>
