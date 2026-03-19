import { notFound } from 'next/navigation'
import { cookies } from 'next/headers'
import { formatDate } from '@/lib/utils'
import type { Secret } from '@/types/api'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { SecretPolicySection } from '@/components/secret-policy-section'
import type { CustomPolicy } from '@/components/policy-editor'

async function backendFetch<T>(path: string): Promise<T | null> {
  const cookieStore = await cookies()
  const token = cookieStore.get('valt_access_token')?.value
  if (!token) return null
  const backend = process.env.BACKEND_URL ?? 'http://localhost:8080'
  const res = await fetch(`${backend}/api/v1${path}`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: 'no-store',
  })
  if (!res.ok) return null
  return res.json() as Promise<T>
}

export default async function SecretDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  const [secret, policy] = await Promise.all([
    backendFetch<Secret>(`/secrets/${id}`),
    backendFetch<CustomPolicy>(`/secrets/${id}/policy`),
  ])
  if (!secret) notFound()

  return (
    <div className="max-w-2xl space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold">{secret.name}</h2>
        <Badge variant={secret.status === 'active' ? 'default' : 'outline'}>{secret.status}</Badge>
      </div>
      <Card>
        <CardHeader><CardTitle className="text-base">Details</CardTitle></CardHeader>
        <CardContent className="space-y-3 text-sm">
          {secret.description && (
            <div><span className="text-muted-foreground">Description: </span>{secret.description}</div>
          )}
          <div><span className="text-muted-foreground">Type: </span>
            <Badge variant="secondary">{secret.credential_type}</Badge>
          </div>
          {secret.source && (
            <div><span className="text-muted-foreground">Source: </span>{secret.source}</div>
          )}
          <div><span className="text-muted-foreground">Created: </span>{formatDate(secret.created_at)}</div>
          <div><span className="text-muted-foreground">Updated: </span>{formatDate(secret.updated_at)}</div>
        </CardContent>
      </Card>
      <SecretPolicySection secretId={id} initialPolicy={policy ?? {}} />
    </div>
  )
}
