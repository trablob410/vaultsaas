import { describe, expect, it } from 'vitest'
import { computeWeakerWarnings, defaultPolicy, mergePolicy } from '@/components/policies/policy-helpers'

describe('policy helpers', () => {
  it('merges override into base policy', () => {
    const base = defaultPolicy()
    const merged = mergePolicy(base, { max_duration_minutes: 30, single_use: true })
    expect(merged.max_duration_minutes).toBe(30)
    expect(merged.single_use).toBe(true)
    expect(merged.require_approval).toBe(base.require_approval)
  })

  it('detects weaker override warning codes', () => {
    const base = defaultPolicy()
    const effective = {
      ...base,
      max_duration_minutes: base.max_duration_minutes + 30,
      require_approval: false,
      min_reason_length: 0,
    }
    const warnings = computeWeakerWarnings(base, effective)
    expect(warnings).toContain('weaker:max_duration_minutes')
    expect(warnings).toContain('weaker:require_approval')
    expect(warnings).toContain('weaker:min_reason_length')
  })
})
