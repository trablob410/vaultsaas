import { redirect } from 'next/navigation'
import { getSession } from '@/lib/auth'

// Minimal layout for onboarding wizard — no sidebar/header chrome.
export default async function OnboardingLayout({ children }: { children: React.ReactNode }) {
  const session = await getSession()
  if (!session) redirect('/login')

  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-4">
      <div className="w-full max-w-lg">
        {children}
      </div>
    </div>
  )
}
