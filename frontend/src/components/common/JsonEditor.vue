<template>
  <div class="json-editor-wrapper">
    <label v-if="label" class="input-label mb-1.5 flex items-center gap-2">
      <span>{{ label }}</span>
      <span
        v-if="validationState !== 'unknown'"
        class="inline-flex h-4 w-4 items-center justify-center rounded-full text-[10px] font-bold"
        :class="validationState === 'valid' ? 'bg-emerald-100 text-emerald-600 dark:bg-emerald-900/40 dark:text-emerald-400' : 'bg-red-100 text-red-600 dark:bg-red-900/40 dark:text-red-400'"
      >
        {{ validationState === 'valid' ? '✓' : '✗' }}
      </span>
    </label>
    <div
      ref="editorContainer"
      class="json-editor-container overflow-hidden rounded-xl border border-gray-200 bg-white transition-all duration-200 dark:border-dark-600 dark:bg-dark-800"
      :class="{ 'border-red-500 ring-2 ring-red-500/20': validationState === 'invalid', 'focus-within:border-primary-500 focus-within:ring-2 focus-within:ring-primary-500/30': validationState !== 'invalid', 'cursor-not-allowed opacity-60': disabled }"
      :style="{ height }"
    />
    <p v-if="error" class="input-error-text mt-1.5">{{ error }}</p>
    <p v-else-if="validationState === 'invalid' && validationMessage" class="input-error-text mt-1.5">{{ validationMessage }}</p>
    <p v-else-if="hint" class="input-hint mt-1.5">{{ hint }}</p>
  </div>
</template>

<script setup lang="ts">
import { Compartment, EditorState } from '@codemirror/state'
import { EditorView, keymap, lineNumbers, highlightActiveLine, highlightActiveLineGutter, rectangularSelection, crosshairCursor } from '@codemirror/view'
import { json, jsonParseLinter } from '@codemirror/lang-json'
import { linter, type Diagnostic } from '@codemirror/lint'
import { defaultKeymap, indentWithTab, history, historyKeymap } from '@codemirror/commands'
import {
  syntaxHighlighting,
  defaultHighlightStyle,
  bracketMatching,
  foldGutter,
  indentOnInput,
  foldKeymap,
  indentUnit,
} from '@codemirror/language'
import { closeBrackets, closeBracketsKeymap } from '@codemirror/autocomplete'
import { onMounted, onUnmounted, ref, watch } from 'vue'

interface Props {
  modelValue: string
  label?: string
  hint?: string
  error?: string
  height?: string
  disabled?: boolean
  localeEnvelope?: boolean
  readOnly?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  height: '300px',
  disabled: false,
  localeEnvelope: false,
  readOnly: false,
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'validationError', message: string | null): void
}>()

const editorContainer = ref<HTMLElement | null>(null)
const validationState = ref<'unknown' | 'valid' | 'invalid'>('unknown')
const validationMessage = ref('')

let editorView: EditorView | null = null
let isUpdatingFromProp = false
let debounceTimer: ReturnType<typeof setTimeout> | null = null

// --- Theme matching the project's .input styling ---

const lightTheme = EditorView.theme({
  '&': {
    fontSize: '0.75rem',
    lineHeight: '1.25rem',
    height: '100%',
  },
  '.cm-content': {
    fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace',
    padding: '0.625rem 0',
    caretColor: '#7c3aed',
  },
  '.cm-cursor': {
    borderLeftColor: '#7c3aed',
    borderLeftWidth: '2px',
  },
  '&.cm-focused .cm-selectionBackground, .cm-selectionBackground, .cm-content ::selection': {
    backgroundColor: 'rgba(124, 58, 237, 0.15) !important',
  },
  '.cm-gutters': {
    backgroundColor: 'transparent',
    borderRight: '1px solid rgb(229 231 235)',
    color: 'rgb(156 163 175)',
    fontSize: '0.6875rem',
    minWidth: '2.5em',
  },
  '.cm-activeLineGutter': {
    backgroundColor: 'rgb(249 250 251)',
    color: 'rgb(107 114 128)',
  },
  '.cm-activeLine': {
    backgroundColor: 'rgb(249 250 251)',
  },
  '.cm-foldGutter': {
    width: '1em',
  },
}, { dark: false })

const darkTheme = EditorView.theme({
  '&': {
    color: 'rgb(243 244 246)',
  },
  '.cm-content': {
    caretColor: '#a78bfa',
  },
  '.cm-cursor': {
    borderLeftColor: '#a78bfa',
  },
  '&.cm-focused .cm-selectionBackground, .cm-selectionBackground, .cm-content ::selection': {
    backgroundColor: 'rgba(167, 139, 250, 0.2) !important',
  },
  '.cm-gutters': {
    backgroundColor: 'transparent',
    borderRightColor: 'rgb(55 65 81)',
    color: 'rgb(107 114 128)',
  },
  '.cm-activeLineGutter': {
    backgroundColor: 'rgb(31 41 55)',
    color: 'rgb(156 163 175)',
  },
  '.cm-activeLine': {
    backgroundColor: 'rgb(31 41 55)',
  },
}, { dark: true })

const themeCompartment = new Compartment()

function isDarkMode(): boolean {
  return document.documentElement.classList.contains('dark')
}

// --- Locale envelope linter ---

function localeEnvelopeLinter() {
  return linter((view): Diagnostic[] => {
    const doc = view.state.doc.toString()
    if (!doc.trim()) return []
    try {
      const parsed = JSON.parse(doc)
      if (typeof parsed !== 'object' || parsed === null || Array.isArray(parsed)) {
        return [{
          from: 0,
          to: doc.length,
          severity: 'error',
          message: 'Expected a JSON object with locale keys (e.g. {"en": {...}, "zh": {...}})',
        }]
      }
      const keys = Object.keys(parsed)
      if (keys.length === 0) {
        return [{
          from: 0,
          to: 1,
          severity: 'warning',
          message: 'No locale keys found. Add at least "en" or "zh".',
        }]
      }
      const invalidKeys = keys.filter(k => typeof parsed[k] !== 'object' || parsed[k] === null || Array.isArray(parsed[k]))
      if (invalidKeys.length > 0) {
        const docText = doc
        return invalidKeys.map(k => {
          const idx = docText.indexOf(`"${k}"`)
          return {
            from: Math.max(0, idx),
            to: Math.max(0, idx + k.length + 2),
            severity: 'error' as const,
            message: `Locale key "${k}" must be an object, got ${Array.isArray(parsed[k]) ? 'array' : typeof parsed[k]}`,
          }
        })
      }
      return []
    } catch {
      // JSON parse errors are already caught by jsonParseLinter
      return []
    }
  })
}

// --- Validation state computation ---

function computeValidation(doc: string): { state: 'unknown' | 'valid' | 'invalid'; message: string } {
  if (!doc.trim()) return { state: 'unknown', message: '' }
  try {
    JSON.parse(doc)
    return { state: 'valid', message: '' }
  } catch (e) {
    return { state: 'invalid', message: (e as Error).message.replace(/^JSON\.parse:\s*/, '') }
  }
}

// --- Create editor ---

function createEditor() {
  if (!editorContainer.value) return

  const extensions = [
    lineNumbers(),
    highlightActiveLineGutter(),
    history(),
    foldGutter(),
    indentOnInput(),
    bracketMatching(),
    closeBrackets(),
    indentUnit.of('  '),
    json(),
    linter(jsonParseLinter(), { delay: 400 }),
    highlightActiveLine(),
    rectangularSelection(),
    crosshairCursor(),
    keymap.of([
      ...closeBracketsKeymap,
      ...defaultKeymap,
      ...searchKeymap,
      ...historyKeymap,
      ...foldKeymap,
      indentWithTab,
    ]),
    syntaxHighlighting(defaultHighlightStyle, { fallback: true }),
    themeCompartment.of(isDarkMode() ? darkTheme : lightTheme),
    EditorView.lineWrapping,
    EditorView.updateListener.of((update) => {
      if (update.docChanged && !isUpdatingFromProp) {
        const doc = update.state.doc.toString()
        const { state, message } = computeValidation(doc)
        validationState.value = state
        validationMessage.value = message
        emit('validationError', state === 'invalid' ? message : null)

        if (debounceTimer) clearTimeout(debounceTimer)
        debounceTimer = setTimeout(() => {
          emit('update:modelValue', doc)
        }, 300)
      }
    }),
    EditorView.domEventHandlers({
      blur: () => {
        // Auto-format on blur if valid
        if (!editorView) return
        const doc = editorView.state.doc.toString()
        if (!doc.trim()) return
        try {
          const parsed = JSON.parse(doc)
          const formatted = JSON.stringify(parsed, null, 2)
          if (formatted !== doc) {
            isUpdatingFromProp = true
            editorView.dispatch({
              changes: { from: 0, to: editorView.state.doc.length, insert: formatted },
            })
            isUpdatingFromProp = false
            emit('update:modelValue', formatted)
          }
        } catch {
          // Not valid JSON, don't format
        }
      },
    }),
  ]

  if (props.readOnly || props.disabled) {
    extensions.push(EditorState.readOnly.of(true))
  }

  if (props.localeEnvelope) {
    extensions.push(localeEnvelopeLinter())
  }

  const state = EditorState.create({
    doc: props.modelValue || '',
    extensions,
  })

  editorView = new EditorView({
    state,
    parent: editorContainer.value,
  })

  // Compute initial validation
  const { state: vState, message: vMessage } = computeValidation(props.modelValue || '')
  validationState.value = vState
  validationMessage.value = vMessage
}

// Search keymap (minimal, to avoid importing @codemirror/search for just keymap bindings)
const searchKeymap: readonly import('@codemirror/view').KeyBinding[] = []

// --- Watch for prop changes ---

watch(() => props.modelValue, (newVal) => {
  if (!editorView) return
  const currentDoc = editorView.state.doc.toString()
  if (newVal === currentDoc) return

  isUpdatingFromProp = true
  editorView.dispatch({
    changes: { from: 0, to: editorView.state.doc.length, insert: newVal || '' },
  })
  isUpdatingFromProp = false

  const { state, message } = computeValidation(newVal || '')
  validationState.value = state
  validationMessage.value = message
})

// --- Watch for dark mode changes ---

let darkModeObserver: MutationObserver | null = null

function startDarkModeObserver() {
  darkModeObserver = new MutationObserver(() => {
    if (!editorView) return
    const dark = isDarkMode()
    editorView.dispatch({
      effects: themeCompartment.reconfigure(dark ? darkTheme : lightTheme),
    })
  })
  darkModeObserver.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ['class'],
  })
}

// --- Lifecycle ---

onMounted(() => {
  createEditor()
  startDarkModeObserver()
})

onUnmounted(() => {
  if (debounceTimer) clearTimeout(debounceTimer)
  if (darkModeObserver) darkModeObserver.disconnect()
  if (editorView) {
    editorView.destroy()
    editorView = null
  }
})
</script>

<style>
.json-editor-container .cm-editor {
  height: 100%;
  border-radius: 0.75rem;
  overflow: hidden;
}

.json-editor-container .cm-editor.cm-focused {
  outline: none;
}

.json-editor-container .cm-scroller {
  overflow: auto;
  border-radius: 0.75rem;
}

.json-editor-container .cm-tooltip {
  border-radius: 0.5rem;
  font-size: 0.75rem;
}

.json-editor-container .cm-tooltip-autocomplete ul li {
  font-size: 0.75rem;
}

.json-editor-container .cm-lint-marker {
  font-size: 0.625rem;
}

.json-editor-container .cm-lintRange-error {
  background-image: none;
  text-decoration: wavy underline rgb(239 68 68 / 0.5);
}

.json-editor-container .cm-lintRange-warning {
  background-image: none;
  text-decoration: wavy underline rgb(245 158 11 / 0.5);
}

.json-editor-container .cm-foldPlaceholder {
  background-color: rgb(243 232 255);
  border: none;
  color: rgb(124 58 237);
  font-size: 0.6875rem;
  padding: 0 0.25rem;
  border-radius: 0.25rem;
}

.dark .json-editor-container .cm-foldPlaceholder {
  background-color: rgb(88 28 135 / 0.4);
  color: rgb(196 181 253);
}
</style>
