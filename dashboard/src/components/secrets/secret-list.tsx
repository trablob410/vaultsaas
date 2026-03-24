'use client'

import { useEffect, useRef, useState } from 'react'
import { Plus, Pencil, Trash2, KeyRound, Shield, ShieldAlert } from 'lucide-react'
import { ApiError, api } from '@/lib/api-client'
import { formatDate } from '@/lib/utils'
import type { Secret } from '@/types/api'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import SecretForm from './secret-form'
import { SecretPolicyBindingDialog } from './secret-policy-binding-dialog'

export default function SecretList() {
  const [secrets, setSecrets] = useState<Secret[]>([])
  const [projectId, setProjectId] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState<Secret | null>(null)

  const [policyDialogOpen, setPolicyDialogOpen] = useState(false)
  const [policySecretId, setPolicySecretId] = useState<string | null>(null)
  const [createdSecretForPolicyId, setCreatedSecretForPolicyId] = useState<string | null>(null)
  const [policyBoundBySecretId, setPolicyBoundBySecretId] = useState<Record<string, boolean | null>>({})
  const loadSeqRef = useRef(0)

  async function loadBindingIndicators(list: Secret[], loadSeq: number) {
    if (list.length === 0) {
      setPolicyBoundBySecretId({})
      return
    }
    setPolicyBoundBySecretId(Object.fromEntries(list.map((secret) => [secret.id, null])))
    await Promise.all(
      list.map(async (secret) => {
        let isBound: boolean | null = null
        try {
          const binding = await api.policies.getBinding(secret.id)
          isBound = Boolean(binding.template?.id)
        } catch (err) {
          if (err instanceof ApiError && err.status === 404) isBound = false
        }
        if (loadSeqRef.current !== loadSeq || isBound === null) return
        setPolicyBoundBySecretId((prev) => {
          if (!(secret.id in prev)) return prev
          return { ...prev, [secret.id]: isBound }
        })
      })
    )
  }

  async function load() {
    const loadSeq = ++loadSeqRef.current
    try {
      const res = await api.secrets.list()
      const list = res.secrets ?? []
      if (loadSeqRef.current !== loadSeq) return
      setSecrets(list)
      void loadBindingIndicators(list, loadSeq)
    } catch (e) {
      if (loadSeqRef.current !== loadSeq) return
      console.error('Failed to load secrets', e)
    } finally {
      if (loadSeqRef.current !== loadSeq) return
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  useEffect(() => {
    const pid = typeof window !== 'undefined' ? localStorage.getItem('valt_current_project') : null
    setProjectId(pid)
  }, [])

  async function handleDelete(id: string) {
    if (!confirm('Delete this secret?')) return
    await api.secrets.delete(id)
    load()
  }

  function openPolicyDialog(secretId: string) {
    setPolicySecretId(secretId)
    setPolicyDialogOpen(true)
  }

  const policySecretName = secrets.find((s) => s.id === policySecretId)?.name

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

      {createdSecretForPolicyId && (
        <div className="flex items-center justify-between rounded-md border border-amber-300 bg-amber-50 px-3 py-2 text-sm dark:bg-amber-950/30">
          <div className="flex items-center gap-2 text-amber-900 dark:text-amber-300">
            <ShieldAlert className="w-4 h-4" />
            <p>Secret created. Apply a policy now.</p>
          </div>
          <div className="flex items-center gap-2">
            <Button
              size="sm"
              variant="outline"
              onClick={() => {
                openPolicyDialog(createdSecretForPolicyId)
                setCreatedSecretForPolicyId(null)
              }}
            >
              <Shield className="w-3.5 h-3.5 mr-1" /> Apply policy now
            </Button>
            <Button size="sm" variant="ghost" onClick={() => setCreatedSecretForPolicyId(null)}>
              Later
            </Button>
          </div>
        </div>
      )}

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
                <TableHead className="w-32">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {secrets.map((s) => (
                <TableRow key={s.id}>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <span className="font-medium">{s.name}</span>
                      {policyBoundBySecretId[s.id] === false && (
                        <Badge
                          variant="outline"
                          title="Secret has no access policy bound."
                          className="gap-1 border-amber-300 text-amber-700"
                        >
                          <ShieldAlert className="w-3 h-3" />
                          No policy
                        </Badge>
                      )}
                    </div>
                  </TableCell>
                  <TableCell>
                    <Badge variant="secondary">{s.credential_type}</Badge>
                  </TableCell>
                  <TableCell>
                    <Badge variant={s.status === 'active' ? 'default' : 'outline'}>{s.status}</Badge>
                  </TableCell>
                  <TableCell className="text-muted-foreground text-xs">{formatDate(s.created_at)}</TableCell>
                  <TableCell>
                    <div className="flex items-center gap-1">
                      <Button variant="ghost" size="icon" title="Edit Secret" onClick={() => { setEditing(s); setFormOpen(true) }}>
                        <Pencil className="w-3.5 h-3.5" />
                      </Button>
                      <Button variant="ghost" size="sm" title="Manage Policies" onClick={() => openPolicyDialog(s.id)}>
                        <Shield className="w-3.5 h-3.5 mr-1" /> Manage Policies
                      </Button>
                      <Button variant="ghost" size="icon" title="Delete Secret" onClick={() => handleDelete(s.id)}>
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
        projectId={projectId}
        onClose={() => setFormOpen(false)}
        onSuccess={() => { setFormOpen(false); load() }}
        onCreated={(id) => setCreatedSecretForPolicyId(id)}
      />

      <SecretPolicyBindingDialog
        open={policyDialogOpen}
        secretId={policySecretId}
        secretName={policySecretName}
        onClose={() => {
          setPolicyDialogOpen(false)
          setPolicySecretId(null)
        }}
        onSaved={() => {
          setPolicyDialogOpen(false)
          setPolicySecretId(null)
          load()
        }}
      />
    </div>
  )
}
