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
} from '@/types/api'

const BASE = '/api/proxy'

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
    throw new Error((err as { error?: string }).error ?? 'Request failed')
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
    list: () => apiFetch<{ orgs: Organization[] }>('/orgs'),
    create: (body: { name: string; slug: string }) =>
      apiFetch<Organization>('/orgs', { method: 'POST', body: JSON.stringify(body) }),
    getMembers: (orgId: string) =>
      apiFetch<{ members: OrgMembership[] }>(`/orgs/${orgId}/members`),
    addMember: (orgId: string, body: { user_id: string; role: string }) =>
      apiFetch<OrgMembership>(`/orgs/${orgId}/members`, { method: 'POST', body: JSON.stringify(body) }),
  },
  workspaces: {
    list: (orgId: string) =>
      apiFetch<{ workspaces: Workspace[] }>(`/orgs/${orgId}/workspaces`),
    create: (orgId: string, body: { name: string; slug: string }) =>
      apiFetch<Workspace>(`/orgs/${orgId}/workspaces`, { method: 'POST', body: JSON.stringify(body) }),
  },
  projects: {
    list: (workspaceId: string) =>
      apiFetch<{ projects: Project[] }>(`/workspaces/${workspaceId}/projects`),
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
}
