import { describe, it, expect } from 'vitest'
import { cn, formatDate } from '@/lib/utils'

describe('cn', () => {
  it('merges class names', () => {
    expect(cn('foo', 'bar')).toBe('foo bar')
  })

  it('handles conditional classes', () => {
    expect(cn('base', false && 'ignored', 'included')).toBe('base included')
  })

  it('deduplicates tailwind conflicts', () => {
    // tailwind-merge resolves conflicts by keeping the last value
    const result = cn('p-2', 'p-4')
    expect(result).toBe('p-4')
  })

  it('handles undefined and null inputs', () => {
    expect(cn('foo', undefined, null, 'bar')).toBe('foo bar')
  })

  it('handles empty string', () => {
    expect(cn('')).toBe('')
  })

  it('handles no arguments', () => {
    expect(cn()).toBe('')
  })
})

describe('formatDate', () => {
  it('returns a non-empty string for valid ISO date', () => {
    const result = formatDate('2024-01-15T10:30:00Z')
    expect(typeof result).toBe('string')
    expect(result.length).toBeGreaterThan(0)
  })

  it('includes year in output', () => {
    const result = formatDate('2024-01-15T10:30:00Z')
    expect(result).toContain('2024')
  })
})
