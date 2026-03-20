'use client'

import { useEffect, useState } from 'react'
import { api } from '@/lib/api-client'
import { computeWeakerWarnings, defaultPolicy, mergePolicy, POLICY_KEYS } from '@/components/policies/policy-helpers'
import type { PolicyParameters, PolicyTemplate, PolicyTemplateVersion, Secret } from '@/types/api'

export function useSecretPolicyBinding(open: boolean, secret?: Secret | null) {
  const [projectId, setProjectId] = useState('')
  const [templates, setTemplates] = useState<PolicyTemplate[]>([])
  const [templateId, setTemplateId] = useState('')
  const [templateVersion, setTemplateVersion] = useState(1)
  const [templateVersions, setTemplateVersions] = useState<PolicyTemplateVersion[]>([])
  const [overrideEnabled, setOverrideEnabled] = useState(false)
  const [overrideParams, setOverrideParams] = useState(defaultPolicy())
  const [warnings, setWarnings] = useState<string[]>([])

  const selectedTemplate = templates.find((t) => t.id === templateId)
  const versionBase = templateVersions.find((v) => v.version === templateVersion)?.parameters
  const basePolicy: PolicyParameters | null = versionBase ?? selectedTemplate?.parameters ?? null

  function resetPolicyState() {
    setTemplateId('')
    setTemplateVersion(1)
    setTemplateVersions([])
    setOverrideEnabled(false)
    setOverrideParams(defaultPolicy())
    setWarnings([])
  }

  async function loadTemplateVersions(id: string) {
    try {
      const res = await api.policies.listVersions(id)
      const list = (res.versions ?? []).sort((a, b) => b.version - a.version)
      setTemplateVersions(list)
    } catch {
      setTemplateVersions([])
    }
  }

  useEffect(() => {
    if (!open) return
    const pid = localStorage.getItem('valt_current_project') ?? ''
    setProjectId(pid)
    if (!pid) return
    api.policies.listTemplates(pid).then((res) => setTemplates(res.templates ?? [])).catch(() => setTemplates([]))
  }, [open])

  useEffect(() => {
    if (!open) return
    resetPolicyState()
    if (!secret) return
    api.policies.getBinding(secret.id).then((binding) => {
      if (!binding.template) return
      setTemplateId(binding.template.id)
      setTemplateVersion(binding.template_version ?? 1)
      setOverrideParams(binding.effective_policy)
      setOverrideEnabled(Object.keys(binding.override_parameters ?? {}).length > 0)
      setWarnings(binding.override_warnings ?? [])
      loadTemplateVersions(binding.template.id)
    }).catch(() => undefined)
  }, [secret, open])

  useEffect(() => {
    if (!selectedTemplate || !overrideEnabled) {
      setWarnings([])
      return
    }
    if (!basePolicy) return
    setWarnings(computeWeakerWarnings(basePolicy, overrideParams))
  }, [selectedTemplate, basePolicy, overrideParams, overrideEnabled])

  function onTemplateChange(id: string) {
    setTemplateId(id)
    const t = templates.find((item) => item.id === id)
    if (!t) return
    setTemplateVersion(t.current_version)
    setOverrideParams(mergePolicy(t.parameters, {}))
    setOverrideEnabled(false)
    setWarnings([])
    loadTemplateVersions(id)
  }

  function buildOverridePayload() {
    if (!basePolicy || !overrideEnabled) return {}
    const partial: Record<string, unknown> = {}
    for (const key of POLICY_KEYS) {
      if (overrideParams[key] !== basePolicy[key]) partial[key] = overrideParams[key]
    }
    return partial
  }

  function onTemplateVersionChange(version: number) {
    setTemplateVersion(version)
    if (!basePolicy || overrideEnabled) return
    const chosen = templateVersions.find((v) => v.version === version)
    if (chosen) setOverrideParams(chosen.parameters)
  }

  return {
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
  }
}
