'use client'

import { useEffect, useState } from 'react'
import { api } from '@/lib/api-client'
import { formatDate } from '@/lib/utils'
import type { AuditLog } from '@/types/api'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { ChevronLeft, ChevronRight } from 'lucide-react'

const PAGE_SIZE = 20

export default function AuditTable() {
  const [logs, setLogs] = useState<AuditLog[]>([])
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(true)
  const [hasMore, setHasMore] = useState(false)

  async function load(p: number) {
    setLoading(true)
    try {
      const res = await api.audit.list({ page: p, limit: PAGE_SIZE })
      setLogs(res.logs ?? [])
      setHasMore((res.logs ?? []).length === PAGE_SIZE)
    } catch (e) {
      console.error('Failed to load audit logs', e)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load(page) }, [page])

  return (
    <div className="space-y-4">
      <div>
        <h2 className="text-lg font-semibold">Audit Logs</h2>
        <p className="text-sm text-muted-foreground">Immutable record of all actions</p>
      </div>

      {loading ? (
        <div className="text-sm text-muted-foreground py-8 text-center">Loading…</div>
      ) : (
        <div className="rounded-lg border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Timestamp</TableHead>
                <TableHead>Action</TableHead>
                <TableHead>Resource</TableHead>
                <TableHead>Actor</TableHead>
                <TableHead>IP</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {logs.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={5} className="text-center text-muted-foreground py-8">
                    No audit logs found.
                  </TableCell>
                </TableRow>
              ) : logs.map((log) => (
                <TableRow key={log.id}>
                  <TableCell className="text-xs text-muted-foreground whitespace-nowrap">
                    {formatDate(log.created_at)}
                  </TableCell>
                  <TableCell>
                    <Badge variant="secondary" className="font-mono text-xs">{log.action}</Badge>
                  </TableCell>
                  <TableCell className="text-sm">
                    <span className="text-muted-foreground">{log.resource_type}/</span>
                    <span className="font-mono text-xs">{log.resource_id?.slice(0, 8)}…</span>
                  </TableCell>
                  <TableCell className="font-mono text-xs">{log.user_id?.slice(0, 8)}…</TableCell>
                  <TableCell className="text-xs text-muted-foreground">{log.ip_address}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      <div className="flex items-center justify-between text-sm text-muted-foreground">
        <span>Page {page}</span>
        <div className="flex gap-1">
          <Button variant="outline" size="icon" disabled={page === 1} onClick={() => setPage(p => p - 1)}>
            <ChevronLeft className="w-4 h-4" />
          </Button>
          <Button variant="outline" size="icon" disabled={!hasMore} onClick={() => setPage(p => p + 1)}>
            <ChevronRight className="w-4 h-4" />
          </Button>
        </div>
      </div>
    </div>
  )
}
