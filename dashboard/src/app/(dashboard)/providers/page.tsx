'use client'

import { useState, useEffect, useCallback } from 'react'
import { api } from '@/lib/api-client'
import type { DynamicProvider } from '@/types/api'

const PROVIDER_TYPES = ['postgres'] as const

const CONFIG_FIELDS: Record<string, { label: string; placeholder: string; secret?: boolean }[]> = {
  postgres: [
    { label: 'Host', placeholder: 'localhost' },
    { label: 'Port', placeholder: '5432' },
    { label: 'Database', placeholder: 'mydb' },
    { label: 'Admin User', placeholder: 'postgres' },
    { label: 'Admin Password', placeholder: '••••••••', secret: true },
    { label: 'SSL Mode', placeholder: 'require' },
  ],
}

const CONFIG_KEYS: Record<string, string[]> = {
  postgres: ['host', 'port', 'database', 'admin_user', 'admin_password', 'ssl_mode'],
}

const DEFAULT_PROJECT_ID = ''

export default function ProvidersPage() {
  const [providers, setProviders] = useState<DynamicProvider[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showForm, setShowForm] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  // Form state
  const [name, setName] = useState('')
  const [providerType, setProviderType] = useState<string>('postgres')
  const [configValues, setConfigValues] = useState<Record<string, string>>({})
  const [projectId] = useState(DEFAULT_PROJECT_ID)

  const fetchProviders = useCallback(async () => {
    if (!projectId) {
      setLoading(false)
      return
    }
    try {
      const res = await api.providers.list(projectId)
      setProviders(res.providers ?? [])
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load providers')
    } finally {
      setLoading(false)
    }
  }, [projectId])

  useEffect(() => {
    fetchProviders()
  }, [fetchProviders])

  const handleConfigChange = (key: string, value: string) => {
    setConfigValues(prev => ({ ...prev, [key]: value }))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!projectId) {
      setError('Project ID is required. Set it in URL or settings.')
      return
    }
    setSubmitting(true)
    setError(null)
    try {
      const keys = CONFIG_KEYS[providerType] ?? []
      const config: Record<string, string> = {}
      keys.forEach(k => { if (configValues[k]) config[k] = configValues[k] })
      await api.providers.create(projectId, { name, provider_type: providerType, config })
      setShowForm(false)
      setName('')
      setConfigValues({})
      await fetchProviders()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create provider')
    } finally {
      setSubmitting(false)
    }
  }

  const fields = CONFIG_FIELDS[providerType] ?? []
  const keys = CONFIG_KEYS[providerType] ?? []

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Dynamic Providers</h1>
          <p className="text-sm text-muted-foreground mt-1">
            On-demand credential generation from configured backends.
          </p>
        </div>
        <button
          onClick={() => setShowForm(v => !v)}
          className="px-4 py-2 bg-primary text-primary-foreground text-sm rounded-md hover:bg-primary/90 transition-colors"
        >
          {showForm ? 'Cancel' : 'Add Provider'}
        </button>
      </div>

      {error && (
        <div className="rounded-md border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error}
        </div>
      )}

      {showForm && (
        <form onSubmit={handleSubmit} className="rounded-lg border bg-card p-5 space-y-4">
          <h2 className="text-base font-medium">New Provider</h2>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1">
              <label className="text-sm font-medium">Name</label>
              <input
                required
                value={name}
                onChange={e => setName(e.target.value)}
                placeholder="my-postgres"
                className="w-full rounded-md border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
              />
            </div>
            <div className="space-y-1">
              <label className="text-sm font-medium">Type</label>
              <select
                value={providerType}
                onChange={e => setProviderType(e.target.value)}
                className="w-full rounded-md border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
              >
                {PROVIDER_TYPES.map(t => (
                  <option key={t} value={t}>{t}</option>
                ))}
              </select>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            {fields.map((f, i) => (
              <div key={keys[i]} className="space-y-1">
                <label className="text-sm font-medium">{f.label}</label>
                <input
                  type={f.secret ? 'password' : 'text'}
                  value={configValues[keys[i]] ?? ''}
                  onChange={e => handleConfigChange(keys[i], e.target.value)}
                  placeholder={f.placeholder}
                  className="w-full rounded-md border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                />
              </div>
            ))}
          </div>

          <div className="flex justify-end gap-2">
            <button
              type="button"
              onClick={() => setShowForm(false)}
              className="px-4 py-2 text-sm rounded-md border hover:bg-accent transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={submitting}
              className="px-4 py-2 bg-primary text-primary-foreground text-sm rounded-md hover:bg-primary/90 disabled:opacity-50 transition-colors"
            >
              {submitting ? 'Creating…' : 'Create Provider'}
            </button>
          </div>
        </form>
      )}

      {loading ? (
        <p className="text-sm text-muted-foreground">Loading providers…</p>
      ) : providers.length === 0 ? (
        <div className="rounded-lg border bg-card p-10 text-center text-sm text-muted-foreground">
          No providers configured. Add one to start generating dynamic credentials.
        </div>
      ) : (
        <div className="rounded-lg border bg-card overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b bg-muted/40">
                <th className="px-4 py-3 text-left font-medium">Name</th>
                <th className="px-4 py-3 text-left font-medium">Type</th>
                <th className="px-4 py-3 text-left font-medium">Status</th>
                <th className="px-4 py-3 text-left font-medium">Created</th>
                <th className="px-4 py-3 text-left font-medium">Leases</th>
              </tr>
            </thead>
            <tbody>
              {providers.map(p => (
                <tr key={p.id} className="border-b last:border-0 hover:bg-muted/20 transition-colors">
                  <td className="px-4 py-3 font-medium">{p.name}</td>
                  <td className="px-4 py-3">
                    <span className="inline-flex items-center rounded-full border px-2 py-0.5 text-xs font-medium">
                      {p.provider_type}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
                      p.status === 'active'
                        ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
                        : 'bg-muted text-muted-foreground'
                    }`}>
                      {p.status}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-muted-foreground font-mono text-xs">
                    {new Date(p.created_at).toLocaleDateString()}
                  </td>
                  <td className="px-4 py-3">
                    <a
                      href={`/providers/${p.id}/leases`}
                      className="text-primary hover:underline text-xs"
                    >
                      View Leases
                    </a>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
