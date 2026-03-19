'use client'

import { api } from '@/lib/api-client'
import { PolicyEditor, type CustomPolicy } from '@/components/policy-editor'

interface ProjectPolicySectionProps {
  projectId: string
  initialPolicy: CustomPolicy
}

export function ProjectPolicySection({ projectId, initialPolicy }: ProjectPolicySectionProps) {
  return (
    <PolicyEditor
      initial={initialPolicy}
      tierMaxDuration={480}
      onSave={(p) => api.policy.putProject(projectId, p as Record<string, unknown>).then(() => {})}
    />
  )
}
