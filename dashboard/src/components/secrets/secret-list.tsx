'use client'

import { useEffect, useState } from 'react'
import { Plus, Pencil, Trash2, KeyRound } from 'lucide-react'
import { api } from '@/lib/api-client'
import { formatDate } from '@/lib/utils'
import type { Secret } from '@/types/api'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import SecretForm from './secret-form'

export default function SecretList() {
  const [secrets, setSecrets] = useState<Secret[]>([])
  const [loading, setLoading] = useState(true)
  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState<Secret | null>(null)

  async function load() {
    try {
      const res = await api.secrets.list()
      setSecrets(res.secrets ?? [])
    } catch (e) {
      console.error('Failed to load secrets', e)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  async function handleDelete(id: string) {
    if (!confirm('Delete this secret?')) return
    await api.secrets.delete(id)
    load()
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold">Secrets</h2>
          <p className="text-sm text-muted-foreground">{secrets.length} secrets stored</p>
        </div>
        <Button onClick={() => { setEditing(null); setFormOpen(true) }}>
          <Plus className="w-4 h-4 mr-2" /> New Secret
        </Button>
      </div>

      {loading ? (
        <div className="text-sm text-muted-foreground py-8 text-center">Loading…</div>
      ) : secrets.length === 0 ? (
        <div className="flex flex-col items-center py-16 gap-3 text-muted-foreground">
          <KeyRound className="w-10 h-10 opacity-30" />
          <p className="text-sm">No secrets yet. Create your first one.</p>
        </div>
      ) : (
        <div className="rounded-lg border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Created</TableHead>
                <TableHead className="w-24">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {secrets.map((s) => (
                <TableRow key={s.id}>
                  <TableCell className="font-medium">{s.name}</TableCell>
                  <TableCell>
                    <Badge variant="secondary">{s.credential_type}</Badge>
                  </TableCell>
                  <TableCell>
                    <Badge variant={s.status === 'active' ? 'default' : 'outline'}>{s.status}</Badge>
                  </TableCell>
                  <TableCell className="text-muted-foreground text-xs">{formatDate(s.created_at)}</TableCell>
                  <TableCell>
                    <div className="flex items-center gap-1">
                      <Button variant="ghost" size="icon" onClick={() => { setEditing(s); setFormOpen(true) }}>
                        <Pencil className="w-3.5 h-3.5" />
                      </Button>
                      <Button variant="ghost" size="icon" onClick={() => handleDelete(s.id)}>
                        <Trash2 className="w-3.5 h-3.5 text-destructive" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      <SecretForm
        open={formOpen}
        secret={editing}
        onClose={() => setFormOpen(false)}
        onSuccess={() => { setFormOpen(false); load() }}
      />
    </div>
  )
}
