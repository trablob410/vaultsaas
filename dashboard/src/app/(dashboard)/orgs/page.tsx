'use client'

import { useEffect, useState } from 'react'
import { api } from '@/lib/api-client'
import type { Organization } from '@/types/api'
import { Building2, Plus } from 'lucide-react'

export default function OrgsPage() {
  const [orgs, setOrgs] = useState<Organization[]>([])
  const [loading, setLoading] = useState(true)
  const [creating, setCreating] = useState(false)
  const [form, setForm] = useState({ name: '', slug: '' })
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  function loadOrgs() {
    setLoading(true)
    api.orgs.list()
      .then(r => setOrgs(r.orgs ?? []))
      .catch(e => setError((e as Error).message))
      .finally(() => setLoading(false))
  }

  useEffect(() => { loadOrgs() }, [])

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    setSubmitting(true)
    setError(null)
    try {
      await api.orgs.create(form)
      setCreating(false)
      setForm({ name: '', slug: '' })
      loadOrgs()
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setSubmitting(false)
    }
  }

  function handleNameChange(e: React.ChangeEvent<HTMLInputElement>) {
    const name = e.target.value
    const slug = name.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '')
    setForm({ name, slug })
  }

  return (
    <div className="p-6 max-w-3xl">
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-3">
          <Building2 className="w-5 h-5 text-muted-foreground" />
          <h1 className="text-xl font-semibold">Organizations</h1>
        </div>
        {!creating && (
          <button
            onClick={() => setCreating(true)}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-md bg-primary text-primary-foreground text-sm font-medium hover:bg-primary/90 transition-colors"
          >
            <Plus className="w-4 h-4" />
            New Organization
          </button>
        )}
      </div>

      {error && (
        <p className="mb-4 text-sm text-destructive">{error}</p>
      )}

      {creating && (
        <form onSubmit={handleCreate} className="mb-6 p-4 border rounded-lg bg-card space-y-3">
          <h2 className="text-sm font-medium">Create Organization</h2>
          <div>
            <label htmlFor="org-name" className="block text-xs text-muted-foreground mb-1">Name</label>
            <input
              id="org-name"
              type="text"
              required
              value={form.name}
              onChange={handleNameChange}
              placeholder="My Organization"
              className="w-full px-3 py-1.5 text-sm border rounded-md bg-background focus:outline-none focus:ring-1 focus:ring-primary"
            />
          </div>
          <div>
            <label htmlFor="org-slug" className="block text-xs text-muted-foreground mb-1">Slug</label>
            <input
              id="org-slug"
              type="text"
              required
              value={form.slug}
              onChange={e => setForm(f => ({ ...f, slug: e.target.value }))}
              placeholder="my-organization"
              className="w-full px-3 py-1.5 text-sm border rounded-md bg-background focus:outline-none focus:ring-1 focus:ring-primary"
            />
          </div>
          <div className="flex gap-2 pt-1">
            <button
              type="submit"
              disabled={submitting}
              className="px-3 py-1.5 rounded-md bg-primary text-primary-foreground text-sm font-medium hover:bg-primary/90 disabled:opacity-50 transition-colors"
            >
              {submitting ? 'Creating…' : 'Create'}
            </button>
            <button
              type="button"
              onClick={() => { setCreating(false); setForm({ name: '', slug: '' }); setError(null) }}
              className="px-3 py-1.5 rounded-md text-sm text-muted-foreground hover:bg-accent transition-colors"
            >
              Cancel
            </button>
          </div>
        </form>
      )}

      {loading ? (
        <p className="text-sm text-muted-foreground">Loading…</p>
      ) : orgs.length === 0 ? (
        <p className="text-sm text-muted-foreground">No organizations yet. Create one to get started.</p>
      ) : (
        <ul className="space-y-2">
          {orgs.map(org => (
            <li key={org.id} className="p-4 border rounded-lg bg-card hover:bg-accent/50 transition-colors">
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 rounded-md bg-primary/10 flex items-center justify-center shrink-0">
                  <Building2 className="w-4 h-4 text-primary" />
                </div>
                <div className="min-w-0">
                  <p className="text-sm font-medium truncate">{org.name}</p>
                  <p className="text-xs text-muted-foreground font-mono">{org.slug}</p>
                </div>
                <span className="ml-auto text-xs text-muted-foreground capitalize">{org.plan}</span>
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
