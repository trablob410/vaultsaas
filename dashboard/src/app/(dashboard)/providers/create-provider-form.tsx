'use client'

import { useState } from 'react'
import { api } from '@/lib/api-client'

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

interface Props {
  projectId: string
  onCreated: () => void
  onCancel: () => void
}

export function CreateProviderForm({ projectId, onCreated, onCancel }: Props) {
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [name, setName] = useState('')
  const [providerType, setProviderType] = useState<string>('postgres')
  const [configValues, setConfigValues] = useState<Record<string, string>>({})

  const fields = CONFIG_FIELDS[providerType] ?? []
  const keys = CONFIG_KEYS[providerType] ?? []

  const handleConfigChange = (key: string, value: string) => {
    setConfigValues(prev => ({ ...prev, [key]: value }))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    setError(null)
    try {
      const config: Record<string, string> = {}
      keys.forEach(k => { if (configValues[k]) config[k] = configValues[k] })
      await api.providers.create(projectId, { name, provider_type: providerType, config })
      setName('')
      setConfigValues({})
      onCreated()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create provider')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="rounded-lg border bg-card p-5 space-y-4">
      <h2 className="text-base font-medium">New Provider</h2>
      {error && (
        <div className="rounded-md border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error}
        </div>
      )}
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
          onClick={onCancel}
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
  )
}
