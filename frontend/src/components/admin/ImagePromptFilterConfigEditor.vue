<template>
  <div class="image-prompt-filter-editor space-y-1.5">
    <label v-if="label" class="input-label mb-1.5 flex items-center justify-between">
      <span>{{ label }}</span>
      <div class="flex items-center gap-2">
        <button
          type="button"
          class="rounded-lg px-2.5 py-1 text-xs font-medium transition-colors duration-150"
          :class="[
            mode === 'form'
              ? 'bg-primary-100 text-primary-700 dark:bg-primary-900/40 dark:text-primary-300'
              : 'text-gray-500 hover:bg-gray-100 dark:text-dark-400 dark:hover:bg-dark-800',
          ]"
          @click="mode = 'form'"
        >
          {{ t('admin.settings.runtime.editorModeForm') }}
        </button>
        <button
          type="button"
          class="rounded-lg px-2.5 py-1 text-xs font-medium transition-colors duration-150"
          :class="[
            mode === 'json'
              ? 'bg-primary-100 text-primary-700 dark:bg-primary-900/40 dark:text-primary-300'
              : 'text-gray-500 hover:bg-gray-100 dark:text-dark-400 dark:hover:bg-dark-800',
          ]"
          @click="mode = 'json'"
        >
          {{ t('admin.settings.runtime.editorModeJson') }}
        </button>
      </div>
    </label>

    <!-- JSON mode -->
    <JsonEditor
      v-if="mode === 'json'"
      :model-value="modelValue"
      :hint="hint"
      :error="parseError || error"
      height="260px"
      :disabled="disabled"
      @update:model-value="onJsonChange($event)"
      @validation-error="onJsonValidationError"
    />

    <!-- Form mode -->
    <div v-else class="space-y-4 rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-600 dark:bg-dark-800">
      <div v-if="parseError" class="rounded-lg bg-red-50 px-3 py-2 text-xs text-red-600 dark:bg-red-900/20 dark:text-red-400">
        {{ parseError }}
      </div>

      <div class="flex items-start justify-between gap-3 rounded-2xl bg-gray-50 p-3 dark:bg-dark-800/70">
        <div>
          <p class="text-sm font-medium text-gray-800 dark:text-dark-200">{{ t('admin.settings.runtime.imageFilterEnabled') }}</p>
        </div>
        <Toggle :model-value="formData.enabled" @update:model-value="updateField('enabled', $event)" />
      </div>

      <div class="space-y-1.5">
        <label class="block text-sm font-medium text-gray-700 dark:text-dark-200">
          {{ t('admin.settings.runtime.imageFilterExplicitKeywords') }}
        </label>
        <textarea
          :value="keywordsToArrayString(formData.explicit_keywords)"
          rows="3"
          class="input w-full resize-y font-mono text-xs leading-5"
          :placeholder="t('admin.settings.runtime.imageFilterExplicitKeywordsPlaceholder')"
          :disabled="disabled"
          @input="onKeywordsChange('explicit_keywords', ($event.target as HTMLTextAreaElement).value)"
        />
        <p class="input-hint">{{ t('admin.settings.runtime.imageFilterKeywordsHint') }}</p>
      </div>

      <div class="space-y-1.5">
        <label class="block text-sm font-medium text-gray-700 dark:text-dark-200">
          {{ t('admin.settings.runtime.imageFilterYouthKeywords') }}
        </label>
        <textarea
          :value="keywordsToArrayString(formData.youth_context_keywords)"
          rows="3"
          class="input w-full resize-y font-mono text-xs leading-5"
          :placeholder="t('admin.settings.runtime.imageFilterYouthKeywordsPlaceholder')"
          :disabled="disabled"
          @input="onKeywordsChange('youth_context_keywords', ($event.target as HTMLTextAreaElement).value)"
        />
      </div>

      <div class="space-y-1.5">
        <label class="block text-sm font-medium text-gray-700 dark:text-dark-200">
          {{ t('admin.settings.runtime.imageFilterWarningMessage') }}
        </label>
        <input
          :value="formData.warning_message"
          type="text"
          class="input"
          :placeholder="t('admin.settings.runtime.imageFilterWarningPlaceholder')"
          :disabled="disabled"
          @input="updateField('warning_message', ($event.target as HTMLInputElement).value)"
        />
      </div>

      <div class="space-y-1.5">
        <label class="block text-sm font-medium text-gray-700 dark:text-dark-200">
          {{ t('admin.settings.runtime.imageFilterYouthWarningMessage') }}
        </label>
        <input
          :value="formData.youth_warning_message"
          type="text"
          class="input"
          :placeholder="t('admin.settings.runtime.imageFilterYouthWarningPlaceholder')"
          :disabled="disabled"
          @input="updateField('youth_warning_message', ($event.target as HTMLInputElement).value)"
        />
      </div>
    </div>

    <p v-if="hint && mode === 'form'" class="input-hint mt-1">{{ hint }}</p>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import JsonEditor from '@/components/common/JsonEditor.vue'
import Toggle from '@/components/common/Toggle.vue'
import type { ImagePromptFilterConfig } from '@/api/admin/settings'

interface Props {
  modelValue: string
  label?: string
  hint?: string
  error?: string
  disabled?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const { t } = useI18n()
const mode = ref<'form' | 'json'>('form')
const parseError = ref('')

const defaultFormData: ImagePromptFilterConfig = {
  enabled: false,
  explicit_keywords: [],
  youth_context_keywords: [],
  warning_message: '',
  youth_warning_message: '',
}

const formData = reactive<ImagePromptFilterConfig>({ ...defaultFormData })

// Parse the modelValue into formData when it changes
function parseConfig(value: string): ImagePromptFilterConfig | null {
  if (!value || !value.trim()) return { ...defaultFormData }
  try {
    const parsed = JSON.parse(value)
    return {
      enabled: typeof parsed.enabled === 'boolean' ? parsed.enabled : false,
      explicit_keywords: Array.isArray(parsed.explicit_keywords) ? parsed.explicit_keywords.filter((k: unknown) => typeof k === 'string') : [],
      youth_context_keywords: Array.isArray(parsed.youth_context_keywords) ? parsed.youth_context_keywords.filter((k: unknown) => typeof k === 'string') : [],
      warning_message: typeof parsed.warning_message === 'string' ? parsed.warning_message : '',
      youth_warning_message: typeof parsed.youth_warning_message === 'string' ? parsed.youth_warning_message : '',
    }
  } catch {
    return null
  }
}

function syncFormData(value: string) {
  const parsed = parseConfig(value)
  if (parsed) {
    parseError.value = ''
    Object.assign(formData, parsed)
  } else if (value && value.trim()) {
    parseError.value = t('admin.settings.runtime.imageFilterParseError')
  } else {
    parseError.value = ''
    Object.assign(formData, defaultFormData)
  }
}

// Initialize from prop
syncFormData(props.modelValue)

watch(() => props.modelValue, (newVal) => {
  syncFormData(newVal)
})

// Convert keywords array to line-separated string for textarea display
function keywordsToArrayString(keywords: string[]): string {
  return keywords.join('\n')
}

// Convert line-separated textarea string back to keywords array
function stringToKeywordsArray(value: string): string[] {
  return value.split('\n').map(k => k.trim()).filter(k => k.length > 0)
}

// Update a single field and emit the new JSON
function updateField(field: keyof ImagePromptFilterConfig, value: unknown) {
  (formData as Record<string, unknown>)[field] = value
  emitConfig()
}

// Handle keywords textarea change
function onKeywordsChange(field: 'explicit_keywords' | 'youth_context_keywords', value: string) {
  formData[field] = stringToKeywordsArray(value)
  emitConfig()
}

// Emit the current formData as JSON
function emitConfig() {
  parseError.value = ''
  emit('update:modelValue', JSON.stringify(formData, null, 2))
}

// Handle JSON editor change
function onJsonChange(value: string) {
  emit('update:modelValue', value)
  // Try to sync form data from JSON
  const parsed = parseConfig(value)
  if (parsed) {
    parseError.value = ''
    Object.assign(formData, parsed)
  } else if (value && value.trim()) {
    parseError.value = t('admin.settings.runtime.imageFilterParseError')
  }
}

function onJsonValidationError(message: string | null) {
  parseError.value = message || ''
}
</script>
