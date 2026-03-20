'use client'

import { Badge } from '@/components/ui/badge'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { PolicyParameterFields } from '@/components/policies/policy-parameter-fields'
import { POLICY_LABELS, POLICY_KEYS, mergePolicy } from '@/components/policies/policy-helpers'
import type { PolicyParameters, PolicyTemplate, PolicyTemplateVersion } from '@/types/api'

interface Props {
  projectId: string
  templates: PolicyTemplate[]
  templateId: string
  templateVersion: number
  basePolicy: PolicyParameters | null
  overrideEnabled: boolean
  overrideParams: PolicyParameters
  warnings: string[]
  versions: PolicyTemplateVersion[]
  onTemplateChange: (templateId: string) => void
  onTemplateVersionChange: (version: number) => void
  onOverrideEnabledChange: (enabled: boolean) => void
  onOverrideParamsChange: (params: PolicyParameters) => void
}

export function SecretPolicyBindingSection({
  projectId,
  templates,
  templateId,
  templateVersion,
  basePolicy,
  overrideEnabled,
  overrideParams,
  warnings,
  versions,
  onTemplateChange,
  onTemplateVersionChange,
  onOverrideEnabledChange,
  onOverrideParamsChange,
}: Props) {
  const selected = templates.find((t) => t.id === templateId)
  const effective = basePolicy && overrideEnabled
    ? mergePolicy(basePolicy, Object.fromEntries(
      POLICY_KEYS
        .filter((k) => overrideParams[k] !== basePolicy[k])
        .map((k) => [k, overrideParams[k]])
    ) as Partial<PolicyParameters>)
    : basePolicy ?? selected?.parameters

  return (
    <div className="space-y-2 rounded-md border p-3">
      <div className="flex items-center justify-between">
        <Label>Policy template</Label>
        {warnings.length > 0 && <Badge variant="outline" className="text-amber-700 border-amber-300">Weaker override</Badge>}
      </div>
      {!projectId ? (
        <p className="text-xs text-muted-foreground">Select current project in Projects page to bind policies.</p>
      ) : (
        <>
          <Select value={templateId} onValueChange={onTemplateChange}>
            <SelectTrigger><SelectValue placeholder="Select template" /></SelectTrigger>
            <SelectContent>
              {templates.map((t) => (
                <SelectItem key={t.id} value={t.id}>{t.name} (v{t.current_version})</SelectItem>
              ))}
            </SelectContent>
          </Select>
          {templateId && (
            <div className="space-y-2">
              <div className="space-y-1.5">
                <Label>Template version</Label>
                <Select value={String(templateVersion)} onValueChange={(v) => onTemplateVersionChange(Number(v))}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    {versions.map((v) => <SelectItem key={v.id} value={String(v.version)}>v{v.version}</SelectItem>)}
                  </SelectContent>
                </Select>
              </div>
              <Label htmlFor="override-toggle" className="flex items-center gap-2">
                <input
                  id="override-toggle"
                  type="checkbox"
                  checked={overrideEnabled}
                  onChange={(e) => onOverrideEnabledChange(e.target.checked)}
                />
                Customize for this secret
              </Label>
              {overrideEnabled && (
                <div className="space-y-2">
                  <PolicyParameterFields value={overrideParams} onChange={onOverrideParamsChange} />
                  {warnings.length > 0 && (
                    <div className="rounded-md border border-amber-300 bg-amber-50 dark:bg-amber-950/30 p-2 text-xs text-amber-800 dark:text-amber-300">
                      Weaker settings detected:
                      <div className="mt-1 flex flex-wrap gap-1">
                        {warnings.map((w) => {
                          const key = w.replace('weaker:', '') as keyof typeof POLICY_LABELS
                          return <Badge key={w} variant="outline" className="border-amber-300">{POLICY_LABELS[key] ?? w}</Badge>
                        })}
                      </div>
                    </div>
                  )}
                </div>
              )}
              {effective && (
                <Card>
                  <CardHeader className="pb-2"><CardTitle className="text-sm">Effective policy preview</CardTitle></CardHeader>
                  <CardContent className="pt-0 grid grid-cols-1 sm:grid-cols-2 gap-1 text-xs text-muted-foreground">
                    {POLICY_KEYS.map((k) => (
                      <div key={k} className="flex items-center justify-between border rounded px-2 py-1">
                        <span>{POLICY_LABELS[k]}</span>
                        <span className="font-medium text-foreground">{String(effective[k])}</span>
                      </div>
                    ))}
                  </CardContent>
                </Card>
              )}
            </div>
          )}
        </>
      )}
    </div>
  )
}
