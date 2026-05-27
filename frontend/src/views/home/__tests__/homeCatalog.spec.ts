import { describe, expect, it } from 'vitest'
import type { HomeCatalogResponse } from '@/types/payment'
import {
  buildHomeModelFamilies,
} from '../homeCatalog'

function fixture(): HomeCatalogResponse {
  return {
    recharge_products: [
      {
        id: 'starter',
        name: '入门包',
        description: '适合快速体验',
        amount: 30,
        credited_amount: 45,
        badge: '推荐',
        recommended: true,
        features: ['45 credits'],
        sort_order: 1,
      },
    ],
    plans: [
      {
        id: 1,
        group_id: 11,
        group_platform: 'anthropic',
        group_name: 'Claude',
        supported_model_scopes: ['Claude Opus 4.6', 'Claude Sonnet 4.6'],
        name: 'Claude 开发包',
        description: '复杂推理',
        price: 59,
        validity_days: 30,
        validity_unit: 'day',
        features: ['priority'],
        for_sale: true,
        sort_order: 1,
      },
      {
        id: 2,
        group_id: 12,
        group_platform: 'openai',
        group_name: 'GPT',
        supported_model_scopes: ['GPT-5.4', 'GPT-5.3 Codex'],
        name: 'GPT 开发包',
        description: '代码生成',
        price: 49,
        validity_days: 30,
        validity_unit: 'day',
        features: ['daily coding'],
        for_sale: true,
        sort_order: 2,
      },
      {
        id: 3,
        group_id: 13,
        group_platform: 'gemini',
        group_name: 'Gemini',
        supported_model_scopes: ['Gemini 2.5 Pro'],
        name: 'Gemini 开发包',
        description: '视觉理解',
        price: 39,
        validity_days: 30,
        validity_unit: 'day',
        features: ['multimodal'],
        for_sale: true,
        sort_order: 3,
      },
    ],
  }
}

describe('homeCatalog helpers', () => {
  it('builds only the publicly visible model families from current plans', () => {
    const families = buildHomeModelFamilies(fixture())

    expect(families.map((item) => item.key)).toEqual(['claude', 'gpt'])
    expect(families[0].models).toContain('Claude Opus 4.6')
    expect(families[1].models).toContain('GPT-5.4')
  })

  it('filters out unsupported/unknown platforms from the model matrix', () => {
    const catalog = fixture()
    catalog.plans.push({
      id: 99,
      group_id: 99,
      group_platform: 'unknown',
      group_name: 'Unknown',
      supported_model_scopes: ['Mystery'],
      name: 'Unknown',
      description: '',
      price: 1,
      validity_days: 30,
      validity_unit: 'day',
      features: [],
      for_sale: true,
      sort_order: 99,
    })

    const families = buildHomeModelFamilies(catalog)
    expect(families.map((item) => item.key)).not.toContain('unknown')
  })
})
