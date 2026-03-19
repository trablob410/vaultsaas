'use client'

import { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'

export interface CustomPolicy {
  approvers?: string[]
  max_duration_minutes?: number
  auto_approve?: boolean
  block?: boolean
  require_reason?: boolean
  min_reason_length?: number
  business_hours?: string
  escalate_after_minutes?: number
  escalate_to_user_id?: string
}

interface PolicyEditorProps {
  initial: CustomPolicy
  onSave: (policy: CustomPolicy) => Promise<void>
  tierMaxDuration?: number
}

export function PolicyEditor({ initial, onSave, tierMaxDuration }: PolicyEditorProps) {
  const [policy, setPolicy] = useState<CustomPolicy>(initial)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [saved, setSaved] = useState(false)

  async function handleSave() {
    setSaving(true)
    setError(null)
    setSaved(false)
    try {
      await onSave(policy)
      setSaved(true)
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Failed to save policy')
    } finally {
      setSaving(false)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Access Policy</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center gap-3">
          <input
            id="block"
            type="checkbox"
            className="h-4 w-4"
            checked={policy.block ?? false}
            onChange={(e) => setPolicy({ ...policy, block: e.target.checked })}
          />
          <Label htmlFor="block">Block auto-approve (always require human review)</Label>
        </div>

        <div className="flex items-center gap-3">
          <input
            id="require_reason"
            type="checkbox"
            className="h-4 w-4"
            checked={policy.require_reason ?? false}
            onChange={(e) => setPolicy({ ...policy, require_reason: e.target.checked })}
          />
          <Label htmlFor="require_reason">Require reason from requester</Label>
        </div>

        {policy.require_reason && (
          <div className="space-y-1">
            <Label htmlFor="min_reason_length">Min reason length (characters)</Label>
            <Input
              id="min_reason_length"
              type="number"
              min={0}
              value={policy.min_reason_length ?? 0}
              onChange={(e) => setPolicy({ ...policy, min_reason_length: Number(e.target.value) })}
            />
          </div>
        )}

        <div className="space-y-1">
          <Label htmlFor="max_duration">
            Max duration (minutes)
            {tierMaxDuration ? (
              <span className="ml-1 text-xs text-muted-foreground">tier default: {tierMaxDuration}</span>
            ) : null}
          </Label>
          <Input
            id="max_duration"
            type="number"
            min={0}
            placeholder="0 = use tier default"
            value={policy.max_duration_minutes ?? ''}
            onChange={(e) =>
              setPolicy({ ...policy, max_duration_minutes: Number(e.target.value) || undefined })
            }
          />
        </div>

        <div className="space-y-1">
          <Label htmlFor="business_hours">Business hours restriction</Label>
          <Input
            id="business_hours"
            placeholder="e.g. 09:00-18:00 Mon-Fri Asia/Bangkok"
            value={policy.business_hours ?? ''}
            onChange={(e) => setPolicy({ ...policy, business_hours: e.target.value || undefined })}
          />
          <p className="text-xs text-muted-foreground">Leave empty for no restriction.</p>
        </div>

        {error && <p className="text-sm text-destructive">{error}</p>}
        {saved && <p className="text-sm text-green-600">Policy saved.</p>}

        <Button onClick={handleSave} disabled={saving}>
          {saving ? 'Saving…' : 'Save Policy'}
        </Button>
      </CardContent>
    </Card>
  )
}
