import { describe, expect, it } from 'vitest'

import en from '../locales/en'
import zh from '../locales/zh'

describe('usage service tier locale keys', () => {
  it('contains zh labels for service tier tooltip', () => {
    expect(zh.common.serviceTier.label).toBe('服务档位')
    expect(zh.common.serviceTier.priority).toBe('Fast')
    expect(zh.common.serviceTier.flex).toBe('Flex')
    expect(zh.common.serviceTier.standard).toBe('Standard')
  })

  it('contains en labels for service tier tooltip', () => {
    expect(en.common.serviceTier.label).toBe('Service tier')
    expect(en.common.serviceTier.priority).toBe('Fast')
    expect(en.common.serviceTier.flex).toBe('Flex')
    expect(en.common.serviceTier.standard).toBe('Standard')
  })
})
