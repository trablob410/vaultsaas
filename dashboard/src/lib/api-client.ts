import type {
  Secret,
  AccessRequest,
  Credential,
  AuditLog,
  Organization,
  Workspace,
  Project,
  OrgMembership,
  ProjectMembership,
  AgentIdentity,
  AgentToken,
  ScanResult,
  ScanFinding,
  DynamicProvider,
  DynamicLease,
  PolicyParameters,
  PolicyTemplate,
  PolicyTemplateVersion,
  SecretPolicyBinding,
} from '@/types/api'

const BASE = '/api/proxy'
export class ApiError extends Error {
  constructor(message: string, public status: number) { super(message); this.name = 'ApiError' }
}

async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    ...options,
    headers: { 'Content-Type': 'application/json', ...options?.headers },
  })
  if (res.status === 401) {
    window.location.href = '/login'
    throw new Error('Unauthorized')
  }
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new ApiError((err as { error?: string }).error ?? 'Request failed', res.status)
  }
  return res.json() as Promise<T>
}

export const api = {
  secrets: {
    list: () => apiFetch<{ secrets: Secret[] }>('/secrets'),
    get: (id: string) => apiFetch<Secret>(`/secrets/${id}`),
    create: (body: Partial<Secret> & { value: string }) =>
      apiFetch<Secret>('/secrets', { method: 'POST', body: JSON.stringify(body) }),
    update: (id: string, body: Partial<Secret>) =>
      apiFetch<Secret>(`/secrets/${id}`, { method: 'PUT', body: JSON.stringify(body) }),
    delete: (id: string) => apiFetch<void>(`/secrets/${id}`, { method: 'DELETE' }),
  },
  requests: {
    list: (status?: string) =>
      apiFetch<{ requests: AccessRequest[] }>(`/access-requests${status ? `?status=${status}` : ''}`),
    create: (secretId: string, body: { reason: string; duration_minutes: number }) =>
      apiFetch<AccessRequest>(`/secrets/${secretId}/access-requests`, {
        method: 'POST',
        body: JSON.stringify(body),
      }),
    approve: (id: string, body: { reason?: string }) =>
      apiFetch<void>(`/access-requests/${id}/approve`, { method: 'POST', body: JSON.stringify(body) }),
    reject: (id: string, body: { reason: string }) =>
      apiFetch<void>(`/access-requests/${id}/reject`, { method: 'POST', body: JSON.stringify(body) }),
  },
  credentials: {
    get: (requestId: string) => apiFetch<Credential>(`/credentials/${requestId}`),
    revoke: (requestId: string) =>
      apiFetch<void>(`/credentials/${requestId}/revoke`, { method: 'POST' }),
  },
  audit: {
    list: (params?: { page?: number; limit?: number }) => {
      const q = new URLSearchParams(params as Record<string, string>).toString()
      return apiFetch<{ logs: AuditLog[] }>(`/audit/logs${q ? `?${q}` : ''}`)
    },
  },
  orgs: {
    list: async () => {
      const data = await apiFetch<{ orgs: Organization[] } | Organization[]>('/orgs')
      return Array.isArray(data) ? { orgs: data } : { orgs: data.orgs ?? [] }
    },
    create: (body: { name: string; slug: string }) =>
      apiFetch<Organization>('/orgs', { method: 'POST', body: JSON.stringify(body) }),
    getMembers: (orgId: string) =>
      apiFetch<{ members: OrgMembership[] }>(`/orgs/${orgId}/members`),
    addMember: (orgId: string, body: { user_id: string; role: string }) =>
      apiFetch<OrgMembership>(`/orgs/${orgId}/members`, { method: 'POST', body: JSON.stringify(body) }),
  },
  workspaces: {
    list: async (orgId: string) => {
      const data = await apiFetch<{ workspaces: Workspace[] } | Workspace[]>(`/orgs/${orgId}/workspaces`)
      return Array.isArray(data) ? { workspaces: data } : { workspaces: data.workspaces ?? [] }
    },
    create: (orgId: string, body: { name: string; slug: string }) =>
      apiFetch<Workspace>(`/orgs/${orgId}/workspaces`, { method: 'POST', body: JSON.stringify(body) }),
  },
  projects: {
    list: async (workspaceId: string) => {
      const data = await apiFetch<{ projects: Project[] } | Project[]>(`/workspaces/${workspaceId}/projects`)
      return Array.isArray(data) ? { projects: data } : { projects: data.projects ?? [] }
    },
    create: (workspaceId: string, body: { name: string; slug: string }) =>
      apiFetch<Project>(`/workspaces/${workspaceId}/projects`, { method: 'POST', body: JSON.stringify(body) }),
    getMembers: (projectId: string) =>
      apiFetch<{ members: ProjectMembership[] }>(`/projects/${projectId}/members`),
    addMember: (projectId: string, body: { user_id: string; role: string }) =>
      apiFetch<ProjectMembership>(`/projects/${projectId}/members`, { method: 'POST', body: JSON.stringify(body) }),
  },
  agents: {
    list: (projectId: string) =>
      apiFetch<{ agents: AgentIdentity[] }>(`/projects/${projectId}/agents`),
    create: (projectId: string, body: {
      name: string
      description?: string
      agent_type: string
      max_session_ttl: number
      allowed_scopes: string[]
    }) => apiFetch<AgentIdentity>(`/projects/${projectId}/agents`, {
      method: 'POST',
      body: JSON.stringify(body),
    }),
    get: (agentId: string) =>
      apiFetch<AgentIdentity>(`/agents/${agentId}`),
    issueToken: (agentId: string, body: { scopes: string[]; expires_in_seconds?: number }) =>
      apiFetch<AgentToken>(`/agents/${agentId}/tokens`, {
        method: 'POST',
        body: JSON.stringify(body),
      }),
    revokeToken: (agentId: string, tokenId: string) =>
      apiFetch<void>(`/agents/${agentId}/tokens/${tokenId}`, { method: 'DELETE' }),
    listTokens: (agentId: string) =>
      apiFetch<{ tokens: AgentToken[] }>(`/agents/${agentId}/tokens`),
  },
  providers: {
    list: (projectId: string) =>
      apiFetch<{ providers: DynamicProvider[] }>(`/projects/${projectId}/providers`),
    create: (projectId: string, body: { name: string; provider_type: string; config: Record<string, string> }) =>
      apiFetch<DynamicProvider>(`/projects/${projectId}/providers`, { method: 'POST', body: JSON.stringify(body) }),
    createLease: (providerId: string, body: { agent_id?: string; ttl_seconds: number }) =>
      apiFetch<DynamicLease>(`/providers/${providerId}/leases`, { method: 'POST', body: JSON.stringify(body) }),
    revokeLease: (leaseId: string) =>
      apiFetch<void>(`/leases/${leaseId}`, { method: 'DELETE' }),
    listLeases: (providerId: string) =>
      apiFetch<{ leases: DynamicLease[] }>(`/providers/${providerId}/leases`),
  },
  scans: {
    list: (projectId: string) =>
      apiFetch<{ scans: ScanResult[] }>(`/projects/${projectId}/scans`),
    getFindings: (scanId: string) =>
      apiFetch<{ findings: ScanFinding[] }>(`/scans/${scanId}/findings`),
    importFinding: (scanId: string, findingId: string, secretId: string) =>
      apiFetch<void>(`/scans/${scanId}/findings/${findingId}/import`, {
        method: 'POST',
        body: JSON.stringify({ secret_id: secretId }),
      }),
    dismissFinding: (scanId: string, findingId: string) =>
      apiFetch<void>(`/scans/${scanId}/findings/${findingId}/dismiss`, { method: 'POST' }),
  },
  policies: {
    listTemplates: (projectId: string) =>
      apiFetch<{ templates: PolicyTemplate[] }>(`/projects/${projectId}/policy-templates`),
    createTemplate: (
      projectId: string,
      body: {
        name: string
        description: string
        base_credential_type?: string
        parameters: PolicyParameters
      }
    ) => apiFetch<PolicyTemplate>(`/projects/${projectId}/policy-templates`, {
      method: 'POST',
      body: JSON.stringify(body),
    }),
    getTemplate: (templateId: string) =>
      apiFetch<PolicyTemplate>(`/policy-templates/${templateId}`),
    updateTemplate: (
      templateId: string,
      body: { parameters: PolicyParameters; change_note: string }
    ) => apiFetch<PolicyTemplateVersion>(`/policy-templates/${templateId}`, {
      method: 'PUT',
      body: JSON.stringify(body),
    }),
    cloneTemplate: (templateId: string, name: string) =>
      apiFetch<PolicyTemplate>(`/policy-templates/${templateId}/clone`, {
        method: 'POST',
        body: JSON.stringify({ name }),
      }),
    listVersions: (templateId: string) =>
      apiFetch<{ versions: PolicyTemplateVersion[] }>(`/policy-templates/${templateId}/versions`),
    getBinding: (secretId: string) =>
      apiFetch<SecretPolicyBinding>(`/secrets/${secretId}/policy-binding`),
    updateBinding: (
      secretId: string,
      body: { template_id: string; template_version: number; override_parameters?: Record<string, unknown> }
    ) => apiFetch<SecretPolicyBinding>(`/secrets/${secretId}/policy-binding`, {
      method: 'PUT',
      body: JSON.stringify(body),
    }),
  },
}
