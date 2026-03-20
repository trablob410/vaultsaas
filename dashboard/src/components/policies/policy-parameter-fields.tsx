'use client'

import type { PolicyParameters } from '@/types/api'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { POLICY_LABELS } from './policy-helpers'

interface Props {
  value: PolicyParameters
  onChange: (next: PolicyParameters) => void
  disabled?: boolean
}

function boolField(
  key: keyof PolicyParameters,
  value: PolicyParameters,
  onChange: (next: PolicyParameters) => void,
  disabled?: boolean
) {
  return (
    <label key={key} className="flex items-center justify-between gap-3 rounded-md border p-3">
      <span className="text-sm">{POLICY_LABELS[key]}</span>
      <input
        type="checkbox"
        checked={Boolean(value[key])}
        disabled={disabled}
        onChange={(e) => onChange({ ...value, [key]: e.target.checked })}
        aria-label={POLICY_LABELS[key]}
      />
    </label>
  )
}

function numberField(
  key: keyof PolicyParameters,
  value: PolicyParameters,
  onChange: (next: PolicyParameters) => void,
  disabled?: boolean,
  min = 0,
  max = 1440
) {
  return (
    <div key={key} className="space-y-1.5">
      <Label htmlFor={key}>{POLICY_LABELS[key]}</Label>
      <Input
        id={key}
        type="number"
        min={min}
        max={max}
        disabled={disabled}
        value={String(value[key] as number)}
        onChange={(e) => onChange({ ...value, [key]: Number(e.target.value) || 0 })}
      />
    </div>
  )
}

export function PolicyParameterFields({ value, onChange, disabled }: Props) {
  return (
    <div className="space-y-3">
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
        {numberField('max_duration_minutes', value, onChange, disabled, 1, 1440)}
        {numberField('min_reason_length', value, onChange, disabled, 0, 500)}
        {numberField('max_requests_per_day', value, onChange, disabled, 1, 1000)}
        {numberField('cool_down_minutes', value, onChange, disabled, 0, 1440)}
      </div>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
        {boolField('require_approval', value, onChange, disabled)}
        {boolField('allow_auto_approve', value, onChange, disabled)}
        {boolField('require_reason', value, onChange, disabled)}
        {boolField('single_use', value, onChange, disabled)}
        {boolField('notify_on_access', value, onChange, disabled)}
        {boolField('require_consent', value, onChange, disabled)}
      </div>
    </div>
  )
}
