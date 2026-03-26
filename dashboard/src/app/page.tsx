import { redirect } from 'next/navigation'
import { getSession } from '@/lib/auth'
import { LandingPage } from '@/components/landing/landing-page'

export default async function Home() {
  const session = await getSession()
  if (session) redirect('/secrets')
  return <LandingPage />
}
