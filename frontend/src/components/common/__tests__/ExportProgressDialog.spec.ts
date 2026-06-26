import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import ExportProgressDialog from '../ExportProgressDialog.vue'

const messages: Record<string, string> = {
  'common.exportProgress.cancelExport': 'Cancel Export',
  'common.exportProgress.estimatedTime': 'Estimated time remaining: {time}',
  'common.exportProgress.exportedCount': 'Exported {current}/{total} records',
  'common.exportProgress.exporting': 'Exporting...',
  'common.exportProgress.exportingProgress': 'Exporting data...',
}

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, unknown>) => {
      let text = messages[key] ?? key
      for (const [name, value] of Object.entries(params ?? {})) {
        text = text.replace(`{${name}}`, String(value))
      }
      return text
    },
  }),
}))

describe('ExportProgressDialog', () => {
  it('renders common export progress copy and emits cancel', async () => {
    const wrapper = mount(ExportProgressDialog, {
      props: {
        show: true,
        progress: 42.4,
        current: 21,
        total: 50,
        estimatedTime: '2m',
      },
      global: {
        stubs: {
          BaseDialog: {
            props: ['show', 'title'],
            emits: ['close'],
            template:
              '<section v-if="show"><h1>{{ title }}</h1><slot /><footer><slot name="footer" /></footer></section>',
          },
        },
      },
    })

    expect(wrapper.text()).toContain('Exporting...')
    expect(wrapper.text()).toContain('Exporting data...')
    expect(wrapper.text()).toContain('Exported 21/50 records')
    expect(wrapper.text()).toContain('Estimated time remaining: 2m')
    expect(wrapper.find('[role="progressbar"]').attributes('aria-label')).toBe(
      'Exporting data...: 42%',
    )

    await wrapper.find('button').trigger('click')
    expect(wrapper.emitted('cancel')).toHaveLength(1)
  })
})
