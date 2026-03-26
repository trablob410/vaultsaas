'use client'

import { useEffect, useState } from 'react'
import { api } from '@/lib/api-client'
import type { EndpointLimit } from '@/types/api'
import { Plus, Trash2, ShieldBan } from 'lucide-react'
import { cn } from '@/lib/utils'

interface EndpointLimitsProps {
  agentId: string
}

export function EndpointLimits({ agentId }: EndpointLimitsProps) {
  const [limits, setLimits] = useState<EndpointLimit[]>([])
  const [adding, setAdding] = useState(false)
  const [form, setForm] = useState({
    host_pattern: '',
    path_pattern: '/*',
    rpm: '60',
    blocked: false,
  })

  useEffect(() => { loadLimits() }, [agentId])

  function loadLimits() {
    api.endpointLimits.list(agentId).then(setLimits).catch(() => setLimits([]))
  }

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    await api.endpointLimits.create({
      agent_id: agentId,
      host_pattern: form.host_pattern,
      path_pattern: form.path_pattern,
      rpm: parseInt(form.rpm) || 60,
      blocked: form.blocked,
    })
    setAdding(false)
    setForm({ host_pattern: '', path_pattern: '/*', rpm: '60', blocked: false })
    loadLimits()
  }

  async function handleDelete(id: string) {
    await api.endpointLimits.delete(id)
    loadLimits()
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="font-semibold flex items-center gap-2">
          <ShieldBan className="w-4 h-4" />Endpoint Rate Limits
        </h2>
        {!adding && (
          <button type="button" onClick={() => setAdding(true)} className="flex items-center gap-1 px-3 py-1.5 text-sm bg-primary text-primary-foreground rounded hover:bg-primary/90">
            <Plus className="w-3 h-3" />Add Limit
          </button>
        )}
      </div>

      {adding && (
        <form onSubmit={handleCreate} className="border rounded-lg p-4 space-y-3 bg-card">
          <h3 className="text-sm font-medium">Add Endpoint Limit</h3>
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1">
              <label className="text-xs font-medium text-muted-foreground">Host Pattern</label>
              <input value={form.host_pattern} onChange={e => setForm(f => ({ ...f, host_pattern: e.target.value }))} placeholder="api.openai.com" className="w-full border rounded px-3 py-1.5 text-sm bg-background font-mono" required />
            </div>
            <div className="space-y-1">
              <label className="text-xs font-medium text-muted-foreground">Path Pattern</label>
              <input value={form.path_pattern} onChange={e => setForm(f => ({ ...f, path_pattern: e.target.value }))} placeholder="/*" className="w-full border rounded px-3 py-1.5 text-sm bg-background font-mono" />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1">
              <label className="text-xs font-medium text-muted-foreground">Requests/min</label>
              <input type="number" min={1} value={form.rpm} onChange={e => setForm(f => ({ ...f, rpm: e.target.value }))} className="w-full border rounded px-3 py-1.5 text-sm bg-background" disabled={form.blocked} />
            </div>
            <div className="flex items-end gap-2 pb-1">
              <label className="flex items-center gap-2 text-sm cursor-pointer">
                <input type="checkbox" checked={form.blocked} onChange={e => setForm(f => ({ ...f, blocked: e.target.checked }))} className="rounded" />
                Block endpoint entirely
              </label>
            </div>
          </div>
          <div className="flex gap-2 justify-end">
            <button type="button" onClick={() => setAdding(false)} className="px-3 py-1.5 text-sm border rounded hover:bg-muted">Cancel</button>
            <button type="submit" className="px-3 py-1.5 text-sm bg-primary text-primary-foreground rounded hover:bg-primary/90">Create Limit</button>
          </div>
        </form>
      )}

      {limits.length === 0 && !adding ? (
        <p className="text-sm text-muted-foreground py-4 text-center border rounded-lg">No endpoint limits configured. All endpoints use global rate limits.</p>
      ) : (
        <div className="border rounded-lg overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-muted/50 border-b">
              <tr>
                <th className="text-left px-4 py-2 font-medium">Endpoint</th>
                <th className="text-left px-4 py-2 font-medium">Limit</th>
                <th className="px-4 py-2" />
              </tr>
            </thead>
            <tbody className="divide-y">
              {limits.map(limit => (
                <tr key={limit.id} className="hover:bg-muted/30">
                  <td className="px-4 py-2">
                    <code className="font-mono text-xs">{limit.host_pattern} {limit.path_pattern}</code>
                  </td>
                  <td className="px-4 py-2">
                    {limit.blocked ? (
                      <span className="px-2 py-0.5 rounded-full text-xs bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300 font-medium">BLOCKED</span>
                    ) : (
                      <span className="text-xs font-mono">{limit.rpm} rpm</span>
                    )}
                  </td>
                  <td className="px-4 py-2 text-right">
                    <button type="button" onClick={() => handleDelete(limit.id)} className="text-destructive hover:opacity-70" aria-label="Delete limit">
                      <Trash2 className="w-4 h-4" />
                    </button>
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
