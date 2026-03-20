'use client'

import { useState, useEffect } from 'react'
import { api } from '@/lib/api-client'
import { CREDENTIAL_TYPES } from '@/lib/constants'
import type { Secret } from '@/types/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { SecretPolicyBindingSection } from './secret-policy-binding-section'
import { useSecretPolicyBinding } from './use-secret-policy-binding'

interface Props {
  open: boolean
  secret?: Secret | null
  onClose: () => void
  onSuccess: () => void
}

export default function SecretForm({ open, secret, onClose, onSuccess }: Props) {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [credentialType, setCredentialType] = useState('')
  const [source, setSource] = useState('')
  const [value, setValue] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const {
    projectId,
    templates,
    templateId,
    templateVersion,
    templateVersions,
    basePolicy,
    overrideEnabled,
    overrideParams,
    warnings,
    onTemplateVersionChange,
    setOverrideEnabled,
    setOverrideParams,
    onTemplateChange,
    buildOverridePayload,
  } = useSecretPolicyBinding(open, secret)

  useEffect(() => {
    if (secret) {
      setName(secret.name)
      setDescription(secret.description ?? '')
      setCredentialType(secret.credential_type)
      setSource(secret.source ?? '')
      setValue('')
    } else {
      setName(''); setDescription(''); setCredentialType(''); setSource(''); setValue('')
    }
    setError('')
  }, [secret, open])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      if (secret) {
        await api.secrets.update(secret.id, { name, description, credential_type: credentialType, source })
        if (templateId) {
          await api.policies.updateBinding(secret.id, {
            template_id: templateId,
            template_version: templateVersion,
            override_parameters: buildOverridePayload(),
          })
        }
      } else {
        const created = await api.secrets.create({ name, description, credential_type: credentialType, source, value })
        if (templateId) {
          await api.policies.updateBinding(created.id, {
            template_id: templateId,
            template_version: templateVersion,
            override_parameters: buildOverridePayload(),
          })
        }
      }
      onSuccess()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Something went wrong')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{secret ? 'Edit Secret' : 'New Secret'}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4 mt-2">
          <div className="space-y-1.5">
            <Label htmlFor="name">Name *</Label>
            <Input id="name" value={name} onChange={(e) => setName(e.target.value)} required />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="desc">Description</Label>
            <Input id="desc" value={description} onChange={(e) => setDescription(e.target.value)} />
          </div>
          <div className="space-y-1.5">
            <Label>Type</Label>
            <Select value={credentialType} onValueChange={setCredentialType}>
              <SelectTrigger><SelectValue placeholder="Select type" /></SelectTrigger>
              <SelectContent>
                {CREDENTIAL_TYPES.map((t) => (
                  <SelectItem key={t.value} value={t.value}>{t.label}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="source">Source</Label>
            <Input id="source" value={source} onChange={(e) => setSource(e.target.value)} />
          </div>

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
            onTemplateChange={onTemplateChange}
            onTemplateVersionChange={onTemplateVersionChange}
            onOverrideEnabledChange={setOverrideEnabled}
            onOverrideParamsChange={setOverrideParams}
          />

          {!secret && (
            <div className="space-y-1.5">
              <Label htmlFor="value">Value *</Label>
              <Textarea id="value" value={value} onChange={(e) => setValue(e.target.value)} required rows={3} />
            </div>
          )}
          {error && <p className="text-sm text-destructive">{error}</p>}
          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose}>Cancel</Button>
            <Button type="submit" disabled={loading}>{loading ? 'Saving…' : 'Save'}</Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
