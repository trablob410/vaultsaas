import { describe, it, expect, vi, beforeEach } from 'vitest'
import { api } from '@/lib/api-client'

const mockFetch = vi.fn()
vi.stubGlobal('fetch', mockFetch)

describe('api client', () => {
  beforeEach(() => {
    mockFetch.mockClear()
  })

  describe('secrets.list', () => {
    it('calls correct proxy URL with JSON content-type', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({ secrets: [] }),
      })
      await api.secrets.list()
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/proxy/secrets',
        expect.objectContaining({
          headers: expect.objectContaining({ 'Content-Type': 'application/json' }),
        })
      )
    })

    it('returns data on success', async () => {
      const mockSecrets = [{ id: '1', name: 'test' }]
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({ secrets: mockSecrets }),
      })
      const result = await api.secrets.list()
      expect(result.secrets).toEqual(mockSecrets)
    })

    it('throws error message on non-ok response', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
        json: async () => ({ error: 'Internal server error' }),
      })
      await expect(api.secrets.list()).rejects.toThrow('Internal server error')
    })

    it('throws fallback message when error body has no error field', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 503,
        statusText: 'Service Unavailable',
        json: async () => ({}),
      })
      await expect(api.secrets.list()).rejects.toThrow('Request failed')
    })
  })

  describe('secrets.get', () => {
    it('calls correct URL with secret id', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({ id: 'abc', name: 'my-secret' }),
      })
      await api.secrets.get('abc')
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/proxy/secrets/abc',
        expect.any(Object)
      )
    })
  })

  describe('secrets.create', () => {
    it('sends POST with body', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 201,
        json: async () => ({ id: 'new', name: 'key', value: 'val' }),
      })
      await api.secrets.create({ name: 'key', value: 'val' })
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/proxy/secrets',
        expect.objectContaining({ method: 'POST' })
      )
    })
  })

  describe('secrets.delete', () => {
    it('sends DELETE to correct URL', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 204,
        json: async () => undefined,
      })
      await api.secrets.delete('xyz')
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/proxy/secrets/xyz',
        expect.objectContaining({ method: 'DELETE' })
      )
    })
  })

  describe('requests.list', () => {
    it('appends status query param when provided', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({ requests: [] }),
      })
      await api.requests.list('pending')
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/proxy/access-requests?status=pending',
        expect.any(Object)
      )
    })

    it('omits query param when status not provided', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({ requests: [] }),
      })
      await api.requests.list()
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/proxy/access-requests',
        expect.any(Object)
      )
    })
  })

  describe('audit.list', () => {
    it('builds query string from params', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({ logs: [] }),
      })
      await api.audit.list({ page: 2, limit: 10 })
      const calledUrl = mockFetch.mock.calls[0][0] as string
      expect(calledUrl).toContain('/api/proxy/audit/logs')
      expect(calledUrl).toContain('page=2')
      expect(calledUrl).toContain('limit=10')
    })

    it('omits query string when no params', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({ logs: [] }),
      })
      await api.audit.list()
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/proxy/audit/logs',
        expect.any(Object)
      )
    })
  })

  describe('policies API', () => {
    it('calls list templates endpoint', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({ templates: [] }),
      })
      await api.policies.listTemplates('p1')
      expect(mockFetch).toHaveBeenCalledWith('/api/proxy/projects/p1/policy-templates', expect.any(Object))
    })

    it('calls update binding endpoint with put payload', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({ secret_id: 's1' }),
      })
      await api.policies.updateBinding('s1', {
        template_id: 't1',
        template_version: 2,
        override_parameters: { max_duration_minutes: 30 },
      })
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/proxy/secrets/s1/policy-binding',
        expect.objectContaining({ method: 'PUT' })
      )
    })
  })

  describe('workspace/project list response shapes', () => {
    it('normalizes array response for workspaces.list', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => [{ id: 'w1', name: 'ws' }],
      })

      const result = await api.workspaces.list('o1')
      expect(result.workspaces).toHaveLength(1)
      expect(result.workspaces[0].id).toBe('w1')
    })

    it('normalizes object response for workspaces.list', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({ workspaces: [{ id: 'w2', name: 'ws2' }] }),
      })

      const result = await api.workspaces.list('o1')
      expect(result.workspaces).toHaveLength(1)
      expect(result.workspaces[0].id).toBe('w2')
    })

    it('normalizes array response for projects.list', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => [{ id: 'p1', name: 'proj' }],
      })

      const result = await api.projects.list('w1')
      expect(result.projects).toHaveLength(1)
      expect(result.projects[0].id).toBe('p1')
    })

    it('normalizes array response for agents.list', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => [{ id: 'a1', name: 'agent-1' }],
      })

      const result = await api.agents.list('p1')
      expect(result.agents).toHaveLength(1)
      expect(result.agents[0].id).toBe('a1')
    })

    it('normalizes object response for agents.list', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({ agents: [{ id: 'a2', name: 'agent-2' }] }),
      })

      const result = await api.agents.list('p1')
      expect(result.agents).toHaveLength(1)
      expect(result.agents[0].id).toBe('a2')
    })

    it('normalizes array response for agents.listTokens', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => [{ id: 't1', agent_id: 'a1' }],
      })

      const result = await api.agents.listTokens('a1')
      expect(result.tokens).toHaveLength(1)
      expect(result.tokens[0].id).toBe('t1')
    })
  })
})
