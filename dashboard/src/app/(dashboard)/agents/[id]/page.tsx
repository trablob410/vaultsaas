'use client'

import { useEffect, useState } from 'react'
import { useParams } from 'next/navigation'
import { api } from '@/lib/api-client'
import type { AgentIdentity, AgentToken } from '@/types/api'
import { Key, Copy, AlertTriangle, Trash2, Bot, Globe, ShieldBan } from 'lucide-react'
import { cn } from '@/lib/utils'
import { ProxyRoutes } from './proxy-routes'
import { EndpointLimits } from './endpoint-limits'

export default function AgentDetailPage() {
  const { id } = useParams<{ id: string }>()
  const [agent, setAgent] = useState<AgentIdentity | null>(null)
  const [tokens, setTokens] = useState<AgentToken[]>([])
  const [issuingToken, setIssuingToken] = useState(false)
  const [newToken, setNewToken] = useState<string | null>(null)
  const [tokenForm, setTokenForm] = useState({ scopes: '', expires_in_seconds: '' })
  const [copied, setCopied] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [tab, setTab] = useState<'tokens' | 'proxy' | 'limits'>('tokens')

  useEffect(() => {
    api.agents.get(id).then(setAgent).catch(() => null)
    api.agents.listTokens(id).then(r => setTokens(r.tokens ?? [])).catch(() => null)
  }, [id])

  async function handleIssueToken(e: React.FormEvent) {
    e.preventDefault()
    setSubmitting(true)
    try {
      const body = {
        scopes: tokenForm.scopes ? tokenForm.scopes.split(',').map(s => s.trim()).filter(Boolean) : [],
        ...(tokenForm.expires_in_seconds ? { expires_in_seconds: parseInt(tokenForm.expires_in_seconds) } : {}),
      }
      const token = await api.agents.issueToken(id, body)
      setNewToken(token.token ?? null)
      setIssuingToken(false)
      setTokenForm({ scopes: '', expires_in_seconds: '' })
      api.agents.listTokens(id).then(r => setTokens(r.tokens ?? []))
    } finally {
      setSubmitting(false)
    }
  }

  function copyToken() {
    if (!newToken) return
    navigator.clipboard.writeText(newToken)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  async function handleRevoke(tokenId: string) {
    await api.agents.revokeToken(id, tokenId)
    api.agents.listTokens(id).then(r => setTokens(r.tokens ?? []))
  }

  if (!agent) return <p className="text-sm text-muted-foreground p-6">Loading agent…</p>

  return (
    <div className="space-y-6 max-w-3xl">
      <div className="flex items-center gap-2 text-sm text-muted-foreground">
        <a href="/agents" className="hover:underline">Agents</a><span>/</span>
        <span className="text-foreground font-medium">{agent.name}</span>
      </div>

      <div className="border rounded-lg p-5 space-y-3 bg-card">
        <div className="flex items-center gap-3">
          <Bot className="w-6 h-6 text-muted-foreground" />
          <div>
            <h1 className="text-xl font-semibold">{agent.name}</h1>
            {agent.description && <p className="text-sm text-muted-foreground">{agent.description}</p>}
          </div>
          <span className={cn('ml-auto px-2 py-0.5 rounded-full text-xs', agent.status === 'active' ? 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300' : 'bg-muted text-muted-foreground')}>
            {agent.status}
          </span>
        </div>
        <dl className="grid grid-cols-2 gap-x-6 gap-y-2 text-sm">
          <div><dt className="text-muted-foreground">Type</dt><dd className="font-mono">{agent.agent_type}</dd></div>
          <div><dt className="text-muted-foreground">Max Session TTL</dt><dd className="font-mono">{agent.max_session_ttl}s</dd></div>
          <div><dt className="text-muted-foreground">Auth Method</dt><dd className="font-mono">{agent.auth_method}</dd></div>
          <div><dt className="text-muted-foreground">Created</dt><dd>{new Date(agent.created_at).toLocaleString()}</dd></div>
          {agent.allowed_scopes.length > 0 && (
            <div className="col-span-2">
              <dt className="text-muted-foreground mb-1">Allowed Scopes</dt>
              <dd className="flex flex-wrap gap-1">
                {agent.allowed_scopes.map(s => <span key={s} className="px-2 py-0.5 rounded-full text-xs bg-secondary text-secondary-foreground font-mono">{s}</span>)}
              </dd>
            </div>
          )}
        </dl>
      </div>

      {newToken && (
        <div className="p-4 bg-amber-50 border border-amber-200 rounded-lg space-y-2 dark:bg-amber-950 dark:border-amber-800">
          <div className="flex items-center gap-2 text-amber-700 dark:text-amber-300 font-medium text-sm">
            <AlertTriangle className="w-4 h-4" />Save this token now — it will never be shown again
          </div>
          <div className="flex items-center gap-2">
            <code className="flex-1 text-xs bg-white dark:bg-black p-2 rounded border font-mono break-all">{newToken}</code>
            <button type="button" onClick={copyToken} className="flex items-center gap-1 px-2 py-1.5 text-xs border rounded hover:bg-muted whitespace-nowrap">
              {copied ? 'Copied!' : <Copy className="w-4 h-4" />}
            </button>
          </div>
          <button type="button" onClick={() => setNewToken(null)} className="text-xs text-muted-foreground hover:underline">Dismiss</button>
        </div>
      )}

      <div className="flex gap-1 border-b">
        {([
          { key: 'tokens' as const, label: 'Tokens', icon: Key },
          { key: 'proxy' as const, label: 'Proxy Routes', icon: Globe },
          { key: 'limits' as const, label: 'Rate Limits', icon: ShieldBan },
        ]).map(t => (
          <button key={t.key} type="button" onClick={() => setTab(t.key)} className={cn('flex items-center gap-1.5 px-4 py-2 text-sm border-b-2 -mb-px transition-colors', tab === t.key ? 'border-primary text-foreground font-medium' : 'border-transparent text-muted-foreground hover:text-foreground')}>
            <t.icon className="w-3.5 h-3.5" />{t.label}
          </button>
        ))}
      </div>

      {tab === 'proxy' && <ProxyRoutes agentId={id} />}
      {tab === 'limits' && <EndpointLimits agentId={id} />}

      {tab === 'tokens' && <div className="space-y-3">
        <div className="flex items-center justify-between">
          <h2 className="font-semibold flex items-center gap-2"><Key className="w-4 h-4" />Tokens</h2>
          {!issuingToken && (
            <button type="button" onClick={() => setIssuingToken(true)} className="flex items-center gap-1 px-3 py-1.5 text-sm bg-primary text-primary-foreground rounded hover:bg-primary/90">
              Issue New Token
            </button>
          )}
        </div>

        {issuingToken && (
          <form onSubmit={handleIssueToken} className="border rounded-lg p-4 space-y-3 bg-card">
            <h3 className="text-sm font-medium">Issue Token</h3>
            <div className="space-y-1">
              <label className="text-xs font-medium text-muted-foreground">Scopes (comma-separated)</label>
              <input value={tokenForm.scopes} onChange={e => setTokenForm(f => ({ ...f, scopes: e.target.value }))} placeholder="secrets:read, approvals:request" className="w-full border rounded px-3 py-1.5 text-sm bg-background font-mono" />
            </div>
            <div className="space-y-1">
              <label className="text-xs font-medium text-muted-foreground">Expires in (seconds, optional)</label>
              <input type="number" min={60} value={tokenForm.expires_in_seconds} onChange={e => setTokenForm(f => ({ ...f, expires_in_seconds: e.target.value }))} placeholder="86400" className="w-full border rounded px-3 py-1.5 text-sm bg-background" />
            </div>
            <div className="flex gap-2 justify-end">
              <button type="button" onClick={() => setIssuingToken(false)} className="px-3 py-1.5 text-sm border rounded hover:bg-muted">Cancel</button>
              <button type="submit" disabled={submitting} className="px-3 py-1.5 text-sm bg-primary text-primary-foreground rounded hover:bg-primary/90 disabled:opacity-50">
                {submitting ? 'Issuing…' : 'Issue Token'}
              </button>
            </div>
          </form>
        )}

        {tokens.length === 0 ? (
          <p className="text-sm text-muted-foreground py-4 text-center border rounded-lg">No tokens issued yet.</p>
        ) : (
          <div className="border rounded-lg overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-muted/50 border-b">
                <tr>
                  <th className="text-left px-4 py-2 font-medium">ID</th>
                  <th className="text-left px-4 py-2 font-medium">Scopes</th>
                  <th className="text-left px-4 py-2 font-medium">Expires</th>
                  <th className="text-left px-4 py-2 font-medium">Last Used</th>
                  <th className="text-left px-4 py-2 font-medium">Status</th>
                  <th className="px-4 py-2" />
                </tr>
              </thead>
              <tbody className="divide-y">
                {tokens.map(token => {
                  const isRevoked = token.revoked_at !== null
                  return (
                    <tr key={token.id} className="hover:bg-muted/30">
                      <td className="px-4 py-2 font-mono text-xs text-muted-foreground">{token.id.slice(0, 8)}…</td>
                      <td className="px-4 py-2"><div className="flex flex-wrap gap-1">{token.scopes.map(s => <span key={s} className="px-1.5 py-0.5 rounded text-xs bg-secondary text-secondary-foreground font-mono">{s}</span>)}</div></td>
                      <td className="px-4 py-2 text-xs text-muted-foreground">{token.expires_at ? new Date(token.expires_at).toLocaleDateString() : '—'}</td>
                      <td className="px-4 py-2 text-xs text-muted-foreground">{token.last_used_at ? new Date(token.last_used_at).toLocaleDateString() : '—'}</td>
                      <td className="px-4 py-2">
                        <span className={cn('px-2 py-0.5 rounded-full text-xs', isRevoked ? 'bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300' : 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300')}>
                          {isRevoked ? 'revoked' : 'active'}
                        </span>
                      </td>
                      <td className="px-4 py-2 text-right">
                        {!isRevoked && (
                          <button type="button" onClick={() => handleRevoke(token.id)} className="text-destructive hover:opacity-70" aria-label="Revoke token">
                            <Trash2 className="w-4 h-4" />
                          </button>
                        )}
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>}
    </div>
  )
}
