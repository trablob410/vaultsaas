'use client'

import { Badge } from '@/components/ui/badge'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { PolicyParameterFields } from '@/components/policies/policy-parameter-fields'
import { POLICY_LABELS, POLICY_KEYS, mergePolicy } from '@/components/policies/policy-helpers'
import type { PolicyParameters, PolicyTemplate, PolicyTemplateVersion } from '@/types/api'

import { AlertCircle, FileWarning, Inbox } from 'lucide-react'

interface Props {
  projectId: string
  templates: PolicyTemplate[]
  templateId: string
  templateVersion: number
  basePolicy: PolicyParameters | null
  loadingTemplates: boolean
  loadingBinding: boolean
  loadingVersions: boolean
  loadError: string
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
  loadingTemplates,
  loadingBinding,
  loadingVersions,
  loadError,
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
  const isTemplateUnavailable = templateId && !selected
  const effective = basePolicy && overrideEnabled
    ? mergePolicy(basePolicy, Object.fromEntries(
      POLICY_KEYS
        .filter((k) => overrideParams[k] !== basePolicy[k])
        .map((k) => [k, overrideParams[k]])
    ) as Partial<PolicyParameters>)
    : basePolicy ?? selected?.parameters

  return (
    <div className="space-y-4 rounded-md border p-4">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
        <Label className="text-base font-semibold">Policy Template Binding</Label>
        {warnings.length > 0 && <Badge variant="outline" className="text-amber-700 border-amber-300">Weaker override detected</Badge>}
      </div>
      
      {!projectId ? (
        <div className="flex flex-col items-center justify-center py-6 text-center space-y-2">
          <AlertCircle className="h-8 w-8 text-muted-foreground" />
          <p className="text-sm text-muted-foreground">Select a current project in the Projects page to bind policies.</p>
        </div>
      ) : loadingBinding || loadingTemplates ? (
        <div className="py-6 text-center text-sm text-muted-foreground">Loading policy data…</div>
      ) : templates.length === 0 && !templateId ? (
        <div className="flex flex-col items-center justify-center py-6 text-center space-y-2 border rounded-md border-dashed">
          <Inbox className="h-8 w-8 text-muted-foreground" />
          <p className="text-sm font-medium">No templates available</p>
          <p className="text-xs text-muted-foreground">Create a policy template in this project to bind it to secrets.</p>
        </div>
      ) : (
        <div className="space-y-4">
          {loadError && (
            <p className="text-sm text-destructive">{loadError}</p>
          )}
          <div className="space-y-2">
            <Label>Select Template</Label>
            <Select value={templateId} onValueChange={onTemplateChange}>
              <SelectTrigger>
                <SelectValue placeholder="Select template" />
              </SelectTrigger>
              <SelectContent>
                {templates.map((t) => (
                  <SelectItem key={t.id} value={t.id}>{t.name} (v{t.current_version})</SelectItem>
                ))}
              </SelectContent>
            </Select>
            {isTemplateUnavailable && (
              <div className="flex items-start gap-2 mt-2 p-2 rounded bg-destructive/10 text-destructive text-sm">
                <FileWarning className="h-4 w-4 mt-0.5 shrink-0" />
                <p>The currently bound template is no longer available or was deleted.</p>
              </div>
            )}
          </div>
          
          {templateId && !isTemplateUnavailable && (
            <div className="space-y-4 pt-2 border-t">
              <div className="space-y-2">
                <Label>Template Version</Label>
                <Select value={String(templateVersion)} onValueChange={(v) => onTemplateVersionChange(Number(v))}>
                  <SelectTrigger><SelectValue placeholder="Select version" /></SelectTrigger>
                  <SelectContent>
                    {loadingVersions ? (
                      <div className="px-2 py-1.5 text-sm text-muted-foreground">Loading versions...</div>
                    ) : versions.length === 0 ? (
                      <div className="px-2 py-1.5 text-sm text-muted-foreground">No versions available</div>
                    ) : (
                      versions.map((v) => <SelectItem key={v.id} value={String(v.version)}>v{v.version}</SelectItem>)
                    )}
                  </SelectContent>
                </Select>
              </div>
              
              <div className="pt-2">
                <Label htmlFor="override-toggle" className="flex items-center gap-2 cursor-pointer">
                  <input
                    id="override-toggle"
                    type="checkbox"
                    className="rounded border-gray-300 text-primary focus:ring-primary h-4 w-4"
                    checked={overrideEnabled}
                    onChange={(e) => onOverrideEnabledChange(e.target.checked)}
                  />
                  <span>Customize parameters for this secret</span>
                </Label>
              </div>
              
              {overrideEnabled && (
                <div className="space-y-4 bg-muted/30 p-4 rounded-md border">
                  <PolicyParameterFields value={overrideParams} onChange={onOverrideParamsChange} />
                  {warnings.length > 0 && (
                    <div className="rounded-md border border-amber-300 bg-amber-50 dark:bg-amber-950/30 p-3 text-sm text-amber-800 dark:text-amber-300">
                      <strong>Warning: Weaker settings detected</strong>
                      <div className="mt-2 flex flex-wrap gap-1.5">
                        {warnings.map((w) => {
                          const key = w.replace('weaker:', '') as keyof typeof POLICY_LABELS
                          return <Badge key={w} variant="secondary" className="bg-amber-100 dark:bg-amber-900 text-amber-800 dark:text-amber-300 border-amber-300">{POLICY_LABELS[key] ?? w}</Badge>
                        })}
                      </div>
                    </div>
                  )}
                </div>
              )}
              
              {effective && (
                <Card className="mt-4 bg-muted/10">
                  <CardHeader className="pb-2">
                    <CardTitle className="text-sm font-semibold">Effective Policy Preview</CardTitle>
                  </CardHeader>
                  <CardContent className="pt-0">
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 text-sm text-muted-foreground">
                      {POLICY_KEYS.map((k) => (
                        <div key={k} className="flex items-center justify-between border rounded bg-background px-3 py-2">
                          <span>{POLICY_LABELS[k]}</span>
                          <span className="font-medium text-foreground">{String(effective[k])}</span>
                        </div>
                      ))}
                    </div>
                  </CardContent>
                </Card>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
