'use client'

import { Button } from '@/components/ui/button'

export default function SignOutButton() {
  async function handleSignOut() {
    await fetch('/api/auth/logout', { method: 'POST' })
    window.location.href = '/login'
  }

  return (
    <Button variant="destructive" onClick={handleSignOut}>
      Sign out
    </Button>
  )
}
