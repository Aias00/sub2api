import { describe, expect, it } from 'vitest'

import {
  filterModelsPlazaItems,
  resolveModelsPlazaActiveGroupLabel,
  resolveModelsPlazaGroupOptions,
  resolveVisibleModelsPlazaItems,
} from '../modelsPlazaRuntime'
import { MODEL_PLAZA_ALL_GROUP_KEY } from '@/utils/modelPlazaDisplay'
import type { ModelPlazaItem } from '@/types'

const items: ModelPlazaItem[] = [
  {
    id: 'claude',
    provider: 'anthropic',
    title: 'Claude Opus',
    badge: '旗舰',
    description: '复杂推理',
    capability_tags: ['复杂推理'],
    model_ids: ['claude-opus'],
    input_price: '',
    output_price: '',
    cache_read_price: '',
    cache_write_price: '',
    billing_badge: '',
    visible: true,
    sort_order: 20,
  },
  {
    id: 'gpt',
    provider: 'openai',
    title: 'GPT-5',
    badge: '编码',
    description: 'Agent',
    capability_tags: ['Agent'],
    model_ids: ['gpt-5'],
    input_price: '',
    output_price: '',
    cache_read_price: '',
    cache_write_price: '',
    billing_badge: '',
    visible: true,
    sort_order: 10,
  },
  {
    id: 'hidden',
    provider: 'openai',
    title: 'Hidden',
    badge: '',
    description: '',
    capability_tags: [],
    model_ids: [],
    input_price: '',
    output_price: '',
    cache_read_price: '',
    cache_write_price: '',
    billing_badge: '',
    visible: false,
    sort_order: 30,
  },
]

describe('modelsPlazaRuntime', () => {
  it('filters visible items and keeps sort order', () => {
    const visible = resolveVisibleModelsPlazaItems(items)
    expect(visible.map((item) => item.id)).toEqual(['gpt', 'claude'])
  })

  it('builds group options and active label from provider groups', () => {
    const visible = resolveVisibleModelsPlazaItems(items)
    const groups = resolveModelsPlazaGroupOptions(visible, {
      groupAll: '全部',
      groupOther: '其他',
    })
    expect(groups[0].key).toBe(MODEL_PLAZA_ALL_GROUP_KEY)
    expect(resolveModelsPlazaActiveGroupLabel(groups, 'gpt', { groupAll: '全部' })).toBe('GPT')
  })

  it('filters by active group and search query', () => {
    const visible = resolveVisibleModelsPlazaItems(items)
    expect(filterModelsPlazaItems(visible, 'gpt', '').map((item) => item.id)).toEqual(['gpt'])
    expect(filterModelsPlazaItems(visible, MODEL_PLAZA_ALL_GROUP_KEY, 'agent').map((item) => item.id)).toEqual(['gpt'])
  })
})
