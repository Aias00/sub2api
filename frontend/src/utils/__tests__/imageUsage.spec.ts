import { describe, expect, it } from 'vitest'

import {
  formatImageBillingSize,
  formatImageInputSize,
  formatImageOutputSize,
  formatImageSizeBreakdown,
  formatImageSizeSource,
} from '../imageUsage'

const messages: Record<string, string> = {
  'common.imageUsage.sizeLegacyUnstandardized': 'legacy unstandardized',
  'common.imageUsage.sizeNotRecorded': 'not recorded',
  'common.imageUsage.sizeSourceInput': 'request input',
  'common.imageUsage.sizeSourceLegacy': 'legacy record',
  'common.imageUsage.sizeSourceMissing': 'missing',
  'common.imageUsage.sizeSourceOutput': 'upstream output',
  'common.imageUsage.sizeUnknown': 'unknown',
}

const t = (key: string) => messages[key] ?? key

describe('imageUsage', () => {
  it('formats billing size with common image usage labels', () => {
    expect(formatImageBillingSize({ image_size: null } as never, t)).toBe('not recorded')
    expect(formatImageBillingSize({ image_size: '2K' } as never, t)).toBe('2K')
    expect(formatImageBillingSize({ image_size: 'custom' } as never, t)).toBe(
      'legacy unstandardized: custom',
    )
  })

  it('formats image input/output sizes and source labels', () => {
    expect(formatImageInputSize({ image_input_size: '' } as never, t)).toBe('unknown')
    expect(formatImageOutputSize({ image_output_size: '4K' } as never, t)).toBe('4K')
    expect(formatImageSizeSource({ image_size_source: 'output' } as never, t)).toBe(
      'upstream output',
    )
    expect(formatImageSizeSource({ image_size_source: 'input' } as never, t)).toBe(
      'request input',
    )
    expect(formatImageSizeSource({ image_size: '1K', image_size_source: null } as never, t)).toBe(
      'legacy record',
    )
    expect(formatImageSizeSource({ image_size: null, image_size_source: null } as never, t)).toBe(
      'missing',
    )
  })

  it('formats image size breakdown in stable tier order', () => {
    expect(
      formatImageSizeBreakdown({
        image_size_breakdown: {
          '4K': 2,
          '1K': 1,
          '2K': 0,
        },
      } as never),
    ).toBe('1K x 1, 4K x 2')
  })
})
