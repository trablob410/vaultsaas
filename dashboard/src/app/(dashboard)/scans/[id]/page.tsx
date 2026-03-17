'use client'

import { useEffect, useState } from 'react'
import { useParams } from 'next/navigation'
import { api } from '@/lib/api-client'
import type { ScanFinding } from '@/types/api'
import { AlertCircle, ArrowLeft } from 'lucide-react'
import { cn } from '@/lib/utils'

function StatusBadge({ status }: { status: string }) {
  const styles =
    status === 'pending'
      ? 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300'
      : status === 'imported'
      ? 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300'
      : 'bg-muted text-muted-foreground'
  return (
    <span className={cn('px-2 py-0.5 rounded-full text-xs font-medium', styles)}>
      {status}
    </span>
  )
}

export default function ScanDetailPage() {
  const { id } = useParams<{ id: string }>()
  const [findings, setFindings] = useState<ScanFinding[]>([])
  const [loading, setLoading] = useState(true)
  const [projectId, setProjectId] = useState<string | null>(null)
  const [dismissing, setDismissing] = useState<string | null>(null)

  useEffect(() => {
    const pid = typeof window !== 'undefined' ? localStorage.getItem('valt_current_project') : null
    setProjectId(pid)
    if (!id) return
    api.scans.getFindings(id)
      .then(r => setFindings(r.findings ?? []))
      .catch(() => setFindings([]))
      .finally(() => setLoading(false))
  }, [id])

  async function handleDismiss(findingId: string) {
    setDismissing(findingId)
    try {
      await api.scans.dismissFinding(id, findingId)
      setFindings(prev => prev.map(f => f.id === findingId ? { ...f, status: 'dismissed' } : f))
    } catch {
      // best-effort
    } finally {
      setDismissing(null)
    }
  }

  if (!projectId) {
    return (
      <div className="flex flex-col items-center justify-center h-64 gap-3 text-muted-foreground">
        <AlertCircle className="w-8 h-8" />
        <p>No project selected. <a href="/projects" className="text-primary underline">Select a project</a> first.</p>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <a href="/scans" className="flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground">
          <ArrowLeft className="w-4 h-4" />
          Back to Scans
        </a>
      </div>

      <div>
        <h1 className="text-2xl font-semibold">Scan Findings</h1>
        <p className="text-sm text-muted-foreground mt-1">Scan ID: <code className="font-mono">{id}</code></p>
      </div>

      {loading ? (
        <p className="text-sm text-muted-foreground">Loading findings…</p>
      ) : findings.length === 0 ? (
        <div className="flex flex-col items-center justify-center h-40 gap-2 text-muted-foreground border rounded-lg">
          <p className="text-sm">No findings for this scan.</p>
        </div>
      ) : (
        <div className="border rounded-lg overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-muted/50 border-b">
              <tr>
                <th className="text-left px-4 py-2 font-medium">File</th>
                <th className="text-left px-4 py-2 font-medium">Line</th>
                <th className="text-left px-4 py-2 font-medium">Pattern</th>
                <th className="text-left px-4 py-2 font-medium">Type</th>
                <th className="text-left px-4 py-2 font-medium">Preview</th>
                <th className="text-left px-4 py-2 font-medium">Status</th>
                <th className="px-4 py-2" aria-label="Actions" />
              </tr>
            </thead>
            <tbody className="divide-y">
              {findings.map(f => (
                <tr key={f.id} className="hover:bg-muted/30">
                  <td className="px-4 py-2 font-mono text-xs max-w-xs truncate" title={f.file_path}>
                    {f.file_path}
                  </td>
                  <td className="px-4 py-2 text-muted-foreground">{f.line_number}</td>
                  <td className="px-4 py-2">{f.pattern_name}</td>
                  <td className="px-4 py-2">{f.credential_type}</td>
                  <td className="px-4 py-2 font-mono text-xs text-muted-foreground max-w-xs truncate" title={f.redacted_preview}>
                    {f.redacted_preview}
                  </td>
                  <td className="px-4 py-2">
                    <StatusBadge status={f.status} />
                  </td>
                  <td className="px-4 py-2 text-right">
                    {f.status === 'pending' && (
                      <button
                        onClick={() => handleDismiss(f.id)}
                        disabled={dismissing === f.id}
                        className="text-xs text-muted-foreground hover:text-foreground disabled:opacity-50"
                      >
                        {dismissing === f.id ? 'Dismissing…' : 'Dismiss'}
                      </button>
                    )}
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
