'use client'

import { useEffect, useState } from 'react'
import { api } from '@/lib/api-client'
import { formatDate } from '@/lib/utils'
import type { AccessRequest } from '@/types/api'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import ApprovalActions from './approval-actions'

const TABS = ['all', 'pending', 'approved', 'rejected'] as const
type Tab = typeof TABS[number]

const statusVariant = (s: string) => {
  if (s === 'approved') return 'default'
  if (s === 'rejected') return 'destructive'
  return 'secondary'
}

export default function ApprovalList() {
  const [requests, setRequests] = useState<AccessRequest[]>([])
  const [tab, setTab] = useState<Tab>('all')
  const [loading, setLoading] = useState(true)
  const [actionsFor, setActionsFor] = useState<AccessRequest | null>(null)

  async function load() {
    setLoading(true)
    try {
      const res = await api.requests.list(tab === 'all' ? undefined : tab)
      setRequests(res.requests ?? [])
    } catch (e) {
      console.error('Failed to load requests', e)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [tab]) // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <div className="space-y-4">
      <div>
        <h2 className="text-lg font-semibold">Access Requests</h2>
        <p className="text-sm text-muted-foreground">Manage human-in-the-loop approvals</p>
      </div>

      <div className="flex gap-1 border-b">
        {TABS.map((t) => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={`px-4 py-2 text-sm font-medium capitalize border-b-2 transition-colors ${
              tab === t ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'
            }`}
          >
            {t}
          </button>
        ))}
      </div>

      {loading ? (
        <div className="text-sm text-muted-foreground py-8 text-center">Loading…</div>
      ) : requests.length === 0 ? (
        <div className="text-sm text-muted-foreground py-8 text-center">No requests found.</div>
      ) : (
        <div className="rounded-lg border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Secret</TableHead>
                <TableHead>Reason</TableHead>
                <TableHead>Duration</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Created</TableHead>
                <TableHead className="w-24">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {requests.map((req) => (
                <TableRow key={req.id}>
                  <TableCell className="text-sm font-medium">
                    {req.secret_name ?? `${req.secret_id.slice(0, 8)}…`}
                  </TableCell>
                  <TableCell className="max-w-[200px] truncate text-sm">{req.reason}</TableCell>
                  <TableCell className="text-sm">
                    {req.requested_duration_minutes >= 60
                      ? `${Math.floor(req.requested_duration_minutes / 60)} hr${Math.floor(req.requested_duration_minutes / 60) > 1 ? 's' : ''}`
                      : `${req.requested_duration_minutes} min`}
                  </TableCell>
                  <TableCell>
                    <Badge variant={statusVariant(req.status)}>{req.status}</Badge>
                  </TableCell>
                  <TableCell className="text-xs text-muted-foreground">{formatDate(req.created_at)}</TableCell>
                  <TableCell>
                    {req.status === 'pending' && (
                      <Button variant="outline" size="sm" onClick={() => setActionsFor(req)}>
                        Review
                      </Button>
                    )}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      <ApprovalActions
        request={actionsFor}
        onClose={() => setActionsFor(null)}
        onSuccess={() => { setActionsFor(null); load() }}
      />
    </div>
  )
}
