'use client'

import { api } from '@/lib/api-client'
import { PolicyEditor, type CustomPolicy } from '@/components/policy-editor'

interface SecretPolicySectionProps {
  secretId: string
  initialPolicy: CustomPolicy
}

export function SecretPolicySection({ secretId, initialPolicy }: SecretPolicySectionProps) {
  return (
    <PolicyEditor
      initial={initialPolicy}
      tierMaxDuration={480}
      onSave={(p) => api.policy.putSecret(secretId, p as Record<string, unknown>).then(() => {})}
    />
  )
}
