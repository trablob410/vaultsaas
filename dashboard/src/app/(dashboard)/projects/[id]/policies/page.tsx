import { redirect } from 'next/navigation'
import { getSession } from '@/lib/auth'
import PolicyTemplatesPage from '@/components/policies/policy-templates-page'

export default async function ProjectPoliciesPage({ params }: { params: Promise<{ id: string }> }) {
  const session = await getSession()
  if (!session) redirect('/login')
  const { id } = await params
  return <PolicyTemplatesPage projectId={id} currentUserId={session.id} />
}
