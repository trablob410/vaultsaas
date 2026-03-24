import type { PolicyParameters } from '@/types/api'

export const POLICY_KEYS: Array<keyof PolicyParameters> = [
  'max_duration_minutes',
  'require_approval',
  'allow_auto_approve',
  'require_reason',
  'min_reason_length',
  'max_requests_per_day',
  'cool_down_minutes',
  'single_use',
  'notify_on_access',
  'require_consent',
]

export const POLICY_LABELS: Record<keyof PolicyParameters, string> = {
  max_duration_minutes: 'Max duration (minutes)',
  require_approval: 'Require approval',
  allow_auto_approve: 'Allow auto approve',
  require_reason: 'Require reason',
  min_reason_length: 'Min reason length',
  max_requests_per_day: 'Max requests/day',
  cool_down_minutes: 'Cool down (minutes)',
  single_use: 'Single use credential',
  notify_on_access: 'Notify on access',
  require_consent: 'Require consent',
}

export function defaultPolicy(): PolicyParameters {
  return {
    max_duration_minutes: 60,
    require_approval: true,
    allow_auto_approve: false,
    require_reason: true,
    min_reason_length: 20,
    max_requests_per_day: 20,
    cool_down_minutes: 5,
    single_use: false,
    notify_on_access: true,
    require_consent: false,
  }
}

export function mergePolicy(base: PolicyParameters, override: Partial<PolicyParameters>): PolicyParameters {
  return { ...base, ...override }
}

export function computeWeakerWarnings(base: PolicyParameters, effective: PolicyParameters): string[] {
  const warnings: string[] = []
  if (effective.max_duration_minutes > base.max_duration_minutes) warnings.push('weaker:max_duration_minutes')
  if (base.require_approval && !effective.require_approval) warnings.push('weaker:require_approval')
  if (!base.allow_auto_approve && effective.allow_auto_approve) warnings.push('weaker:allow_auto_approve')
  if (base.require_reason && !effective.require_reason) warnings.push('weaker:require_reason')
  if (effective.min_reason_length < base.min_reason_length) warnings.push('weaker:min_reason_length')
  if (effective.max_requests_per_day > base.max_requests_per_day) warnings.push('weaker:max_requests_per_day')
  if (effective.cool_down_minutes < base.cool_down_minutes) warnings.push('weaker:cool_down_minutes')
  if (base.single_use && !effective.single_use) warnings.push('weaker:single_use')
  if (base.notify_on_access && !effective.notify_on_access) warnings.push('weaker:notify_on_access')
  if (base.require_consent && !effective.require_consent) warnings.push('weaker:require_consent')
  return warnings
}
