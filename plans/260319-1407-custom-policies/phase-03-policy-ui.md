# Phase 03: Dashboard Policy UI

**Priority:** P2 — requires Phase 02 complete
**Status:** COMPLETED — 2026-03-19

## UX Design

```
Secret Detail Page (/secrets/[id])
  └─▶ "Policy" tab / drawer
       ├── Approvers: user picker (multi-select from project members)
       ├── Max Duration: number input (minutes) with tier cap shown as hint
       ├── Require Reason: toggle
       ├── Min Reason Length: number input (shown if require_reason=true)
       ├── Auto Approve: toggle (override tier auto-approve)
       ├── Block All Auto-Approve: toggle
       └── [Save Policy] button

Project Settings (/projects/[id]/settings → Policy tab)
  └─▶ Same fields but apply as defaults for all secrets in project
```

## Files

| File | Change |
|------|--------|
| `dashboard/src/lib/api/policy.ts` | Create — API client for policy endpoints |
| `dashboard/src/components/policy-editor.tsx` | Create — reusable policy form component |
| `dashboard/src/app/(dashboard)/secrets/[id]/page.tsx` | Modify — add Policy tab |
| `dashboard/src/app/(dashboard)/projects/[id]/settings/page.tsx` | Create or Modify — add Policy tab |

---

## Task 1: API client

**Files:**
- Create: `dashboard/src/lib/api/policy.ts`

- [ ] Write:

```typescript
import { apiFetch } from '@/lib/api-client'

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

export async function getSecretPolicy(secretId: string): Promise<CustomPolicy> {
  return apiFetch(`/secrets/${secretId}/policy`)
}

export async function putSecretPolicy(secretId: string, policy: CustomPolicy): Promise<CustomPolicy> {
  return apiFetch(`/secrets/${secretId}/policy`, {
    method: 'PUT',
    body: JSON.stringify(policy),
  })
}

export async function getProjectPolicy(projectId: string): Promise<CustomPolicy> {
  return apiFetch(`/projects/${projectId}/policy`)
}

export async function putProjectPolicy(projectId: string, policy: CustomPolicy): Promise<CustomPolicy> {
  return apiFetch(`/projects/${projectId}/policy`, {
    method: 'PUT',
    body: JSON.stringify(policy),
  })
}
```

- [ ] Commit:
```bash
git add dashboard/src/lib/api/policy.ts
git commit -m "feat(dashboard): add policy API client"
```

---

## Task 2: PolicyEditor component

**Files:**
- Create: `dashboard/src/components/policy-editor.tsx`

- [ ] Write reusable policy form (shadcn/ui: Card, Switch, Input, Label, Button):

```tsx
'use client'

import { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Switch } from '@/components/ui/switch'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import type { CustomPolicy } from '@/lib/api/policy'

interface PolicyEditorProps {
  initial: CustomPolicy
  onSave: (policy: CustomPolicy) => Promise<void>
  tierMaxDuration?: number   // hint from tier default
}

export function PolicyEditor({ initial, onSave, tierMaxDuration }: PolicyEditorProps) {
  const [policy, setPolicy] = useState<CustomPolicy>(initial)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleSave() {
    setSaving(true)
    setError(null)
    try {
      await onSave(policy)
    } catch (e: any) {
      setError(e.message ?? 'Failed to save policy')
    } finally {
      setSaving(false)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Access Policy</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center justify-between">
          <Label>Block auto-approve (always require human)</Label>
          <Switch
            checked={policy.block ?? false}
            onCheckedChange={(v) => setPolicy({ ...policy, block: v })}
          />
        </div>

        <div className="flex items-center justify-between">
          <Label>Require reason</Label>
          <Switch
            checked={policy.require_reason ?? false}
            onCheckedChange={(v) => setPolicy({ ...policy, require_reason: v })}
          />
        </div>

        {policy.require_reason && (
          <div className="space-y-1">
            <Label>Min reason length (chars)</Label>
            <Input
              type="number"
              min={0}
              value={policy.min_reason_length ?? 0}
              onChange={(e) => setPolicy({ ...policy, min_reason_length: Number(e.target.value) })}
            />
          </div>
        )}

        <div className="space-y-1">
          <Label>
            Max duration (minutes)
            {tierMaxDuration ? ` — tier default: ${tierMaxDuration}` : ''}
          </Label>
          <Input
            type="number"
            min={0}
            placeholder="0 = use tier default"
            value={policy.max_duration_minutes ?? ''}
            onChange={(e) => setPolicy({ ...policy, max_duration_minutes: Number(e.target.value) || undefined })}
          />
        </div>

        <div className="space-y-1">
          <Label>Business hours (e.g. 09:00-18:00 Mon-Fri Asia/Bangkok)</Label>
          <Input
            placeholder="Leave empty for no restriction"
            value={policy.business_hours ?? ''}
            onChange={(e) => setPolicy({ ...policy, business_hours: e.target.value || undefined })}
          />
        </div>

        {error && <p className="text-sm text-destructive">{error}</p>}

        <Button onClick={handleSave} disabled={saving}>
          {saving ? 'Saving…' : 'Save Policy'}
        </Button>
      </CardContent>
    </Card>
  )
}
```

- [ ] Run: `cd dashboard && npm run lint`

- [ ] Commit:
```bash
git add dashboard/src/components/policy-editor.tsx
git commit -m "feat(dashboard): add reusable PolicyEditor component"
```

---

## Task 3: Wire into Secret detail page

**Files:**
- Modify: `dashboard/src/app/(dashboard)/secrets/[id]/page.tsx`

- [ ] Add "Policy" tab to the existing Tabs component on the secret detail page:

```tsx
import { getSecretPolicy, putSecretPolicy } from '@/lib/api/policy'
import { PolicyEditor } from '@/components/policy-editor'

// In the server component — fetch policy alongside secret:
const policy = await getSecretPolicy(params.id)

// Add tab:
<TabsTrigger value="policy">Policy</TabsTrigger>

<TabsContent value="policy">
  <PolicyEditor
    initial={policy}
    tierMaxDuration={480}
    onSave={(p) => putSecretPolicy(params.id, p)}
  />
</TabsContent>
```

- [ ] Run: `cd dashboard && npm run lint && npm test`

- [ ] Commit:
```bash
git add dashboard/src/app/\(dashboard\)/secrets/\[id\]/page.tsx
git commit -m "feat(dashboard): add Policy tab to secret detail page"
```

---

## Task 4: Wire into Project settings

**Files:**
- Create/Modify: `dashboard/src/app/(dashboard)/projects/[id]/settings/page.tsx`

- [ ] Add project policy tab (same `PolicyEditor`, wired to `getProjectPolicy`/`putProjectPolicy`)

- [ ] Run: `cd dashboard && npm run lint && npm test`

- [ ] Commit:
```bash
git add dashboard/src/app/\(dashboard\)/projects/
git commit -m "feat(dashboard): add Policy settings tab to project settings page"
```

---

## Success Criteria
- Secret detail page shows Policy tab with current settings
- Toggling "Block auto-approve" and saving → new requests require human approval
- Project policy saved → new secrets in project inherit it
- Invalid business hours format → backend returns 400, UI shows error
