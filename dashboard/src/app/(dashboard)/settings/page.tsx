import { cookies } from 'next/headers'
import { redirect } from 'next/navigation'
import { getSession } from '@/lib/auth'
import type { User } from '@/types/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Separator } from '@/components/ui/separator'
import SignOutButton from './sign-out-button'

async function getUser(): Promise<User | null> {
  const cookieStore = await cookies()
  const token = cookieStore.get('valt_access_token')?.value
  if (!token) return null
  const session = await getSession()
  if (!session) return null
  // Return minimal info from JWT; full profile would require a /me endpoint
  return { id: session.id, email: session.email ?? '', region_code: '', auth_provider: 'google' }
}

export default async function SettingsPage() {
  const user = await getUser()
  if (!user) redirect('/login')

  return (
    <div className="max-w-lg space-y-4">
      <h2 className="text-lg font-semibold">Settings</h2>

      <Card>
        <CardHeader><CardTitle className="text-base">Profile</CardTitle></CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center gap-4">
            <Avatar className="w-12 h-12">
              <AvatarImage src={undefined} />
              <AvatarFallback>{user.email?.[0]?.toUpperCase() ?? 'U'}</AvatarFallback>
            </Avatar>
            <div>
              <p className="font-medium text-sm">{user.email}</p>
              <p className="text-xs text-muted-foreground capitalize">{user.auth_provider} account</p>
            </div>
          </div>

          <Separator />

          <div className="text-sm space-y-2">
            <div className="flex justify-between">
              <span className="text-muted-foreground">User ID</span>
              <span className="font-mono text-xs">{user.id.slice(0, 16)}…</span>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader><CardTitle className="text-base">Account</CardTitle></CardHeader>
        <CardContent>
          <SignOutButton />
        </CardContent>
      </Card>
    </div>
  )
}
