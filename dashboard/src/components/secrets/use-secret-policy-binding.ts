'use client'

import { useEffect, useRef, useState } from 'react'
import { api } from '@/lib/api-client'
import { computeWeakerWarnings, defaultPolicy, mergePolicy, POLICY_KEYS } from '@/components/policies/policy-helpers'
import type { PolicyParameters, PolicyTemplate, PolicyTemplateVersion } from '@/types/api'

interface UseSecretPolicyBindingParams {
  open: boolean
  secretId?: string | null
}

export function useSecretPolicyBinding({ open, secretId }: UseSecretPolicyBindingParams) {
  const [projectId, setProjectId] = useState('')
  const [templates, setTemplates] = useState<PolicyTemplate[]>([])
  const [templateId, setTemplateId] = useState('')
  const [templateVersion, setTemplateVersion] = useState(1)
  const [templateVersions, setTemplateVersions] = useState<PolicyTemplateVersion[]>([])
  const [loadingTemplates, setLoadingTemplates] = useState(false)
  const [loadingBinding, setLoadingBinding] = useState(false)
  const [loadingVersions, setLoadingVersions] = useState(false)
  const [loadError, setLoadError] = useState('')
  const [overrideEnabled, setOverrideEnabled] = useState(false)
  const [overrideParams, setOverrideParams] = useState(defaultPolicy())
  const [warnings, setWarnings] = useState<string[]>([])
  const templateReqRef = useRef(0)
  const bindingReqRef = useRef(0)

  const selectedTemplate = templates.find((t) => t.id === templateId)
  const versionBase = templateVersions.find((v) => v.version === templateVersion)?.parameters
  const basePolicy: PolicyParameters | null = versionBase ?? selectedTemplate?.parameters ?? null

  function resetPolicyState() {
    setTemplateId('')
    setTemplateVersion(1)
    setTemplateVersions([])
    setLoadError('')
    setOverrideEnabled(false)
    setOverrideParams(defaultPolicy())
    setWarnings([])
  }

  async function loadTemplateVersions(id: string) {
    const reqId = ++templateReqRef.current
    setLoadingVersions(true)
    setLoadError('')
    try {
      const res = await api.policies.listVersions(id)
      if (reqId !== templateReqRef.current) return
      const list = (res.versions ?? []).sort((a, b) => b.version - a.version)
      setTemplateVersions(list)
    } catch {
      if (reqId !== templateReqRef.current) return
      setTemplateVersions([])
      setLoadError('Failed to load template versions')
    } finally {
      if (reqId !== templateReqRef.current) return
      setLoadingVersions(false)
    }
  }

  useEffect(() => {
    if (!open) return
    const pid = localStorage.getItem('valt_current_project') ?? ''
    setProjectId(pid)
    if (!pid) {
      setTemplates([])
      setLoadingTemplates(false)
      return
    }
    setLoadingTemplates(true)
    setLoadError('')
    api.policies.listTemplates(pid)
      .then((res) => {
        setTemplates(res.templates ?? [])
      })
      .catch(() => {
        setTemplates([])
        setLoadError('Failed to load policy templates')
      })
      .finally(() => setLoadingTemplates(false))
  }, [open])

  useEffect(() => {
    if (!open) return
    const reqId = ++bindingReqRef.current
    resetPolicyState()
    if (!secretId) {
      setLoadingBinding(false)
      return
    }
    setLoadingBinding(true)
    setLoadError('')
    api.policies.getBinding(secretId)
      .then((binding) => {
        if (reqId !== bindingReqRef.current) return
        if (!binding.template) return
        setTemplateId(binding.template.id)
        setTemplateVersion(binding.template_version ?? 1)
        setOverrideParams(binding.effective_policy)
        setOverrideEnabled(Object.keys(binding.override_parameters ?? {}).length > 0)
        setWarnings(binding.override_warnings ?? [])
        loadTemplateVersions(binding.template.id)
      })
      .catch(() => {
        if (reqId !== bindingReqRef.current) return
        setLoadError('Failed to load current policy binding')
      })
      .finally(() => {
        if (reqId !== bindingReqRef.current) return
        setLoadingBinding(false)
      })
  }, [secretId, open])

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
    if (!Number.isFinite(version) || version < 1) return
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
  }
}
