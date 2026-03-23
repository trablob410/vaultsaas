'use client'

import { useEffect, useMemo, useState } from 'react'
import { api } from '@/lib/api-client'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { SecretPolicyBindingSection } from './secret-policy-binding-section'
import { useSecretPolicyBinding } from './use-secret-policy-binding'

interface SecretPolicyBindingDialogProps {
  open: boolean
  secretId: string | null
  secretName?: string
  onClose: () => void
  onSaved?: () => void
}

export function SecretPolicyBindingDialog({
  open,
  secretId,
  secretName,
  onClose,
  onSaved,
}: SecretPolicyBindingDialogProps) {
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const {
    projectId,
    templates,
    templateId,
    templateVersion,
    templateVersions,
    loadingTemplates,
    loadingBinding,
    loadingVersions,
    loadError,
    basePolicy,
    overrideEnabled,
    overrideParams,
    warnings,
    onTemplateVersionChange,
    setOverrideEnabled,
    setOverrideParams,
    onTemplateChange,
    buildOverridePayload,
  } = useSecretPolicyBinding({ open, secretId })

  const isTemplateUnavailable = useMemo(
    () => Boolean(templateId) && !templates.some((t) => t.id === templateId),
    [templateId, templates]
  )

  useEffect(() => {
    if (!open) setError('')
  }, [open])

  async function handleSave() {
    if (!secretId || !templateId) {
      setError('Select a policy template before saving.')
      return
    }
    setSubmitting(true)
    setError('')
    try {
      await api.policies.updateBinding(secretId, {
        template_id: templateId,
        template_version: templateVersion,
        override_parameters: buildOverridePayload(),
      })
      onSaved?.()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save policy binding')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={(next) => !next && onClose()}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Manage policies</DialogTitle>
          <DialogDescription>
            Configure template, version, and overrides{secretName ? ` for ${secretName}` : ''}.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <SecretPolicyBindingSection
            projectId={projectId}
            templates={templates}
            templateId={templateId}
            templateVersion={templateVersion}
            basePolicy={basePolicy}
            overrideEnabled={overrideEnabled}
            overrideParams={overrideParams}
            warnings={warnings}
            versions={templateVersions}
            loadingTemplates={loadingTemplates}
            loadingBinding={loadingBinding}
            loadingVersions={loadingVersions}
            loadError={loadError}
            onTemplateChange={onTemplateChange}
            onTemplateVersionChange={onTemplateVersionChange}
            onOverrideEnabledChange={setOverrideEnabled}
            onOverrideParamsChange={setOverrideParams}
          />
          {error && <p className="text-sm text-destructive">{error}</p>}
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button
            type="button"
            disabled={submitting || loadingBinding || loadingTemplates || isTemplateUnavailable}
            onClick={handleSave}
          >
            {submitting ? 'Saving…' : 'Save policy'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
