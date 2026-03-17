'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { KeyRound, ClipboardCheck, ScrollText, Settings, LogOut, Shield, Bot, Building2, ChevronDown, ScanLine, Database, Zap } from 'lucide-react'
import { cn } from '@/lib/utils'

const navItems = [
  { href: '/secrets', label: 'Secrets', icon: KeyRound },
  { href: '/approvals', label: 'Approvals', icon: ClipboardCheck },
  { href: '/audit', label: 'Audit', icon: ScrollText },
  { href: '/agents', label: 'Agents', icon: Bot },
  { href: '/scans', label: 'Scanner', icon: ScanLine },
  { href: '/providers', label: 'Providers', icon: Database },
  { href: '/settings', label: 'Settings', icon: Settings },
]

async function handleLogout() {
  await fetch('/api/auth/logout', { method: 'POST' })
  window.location.href = '/login'
}

export default function Sidebar() {
  const pathname = usePathname()

  return (
    <aside className="flex flex-col w-60 border-r bg-card shrink-0">
      <div className="flex items-center gap-2 px-4 h-14 border-b">
        <div className="w-7 h-7 rounded-lg bg-primary/10 flex items-center justify-center">
          <Shield className="w-4 h-4 text-primary" />
        </div>
        <span className="font-semibold text-sm">Valt</span>
      </div>

      {/* Org context */}
      <div className="px-3 py-2 border-b">
        <button
          onClick={() => { window.location.href = '/orgs' }}
          className="w-full flex items-center gap-2 px-2 py-1.5 rounded text-xs text-muted-foreground hover:bg-accent hover:text-foreground transition-colors"
        >
          <Building2 className="w-3.5 h-3.5 shrink-0" />
          <span className="truncate font-medium">Organizations</span>
          <ChevronDown className="w-3 h-3 ml-auto shrink-0" />
        </button>
      </div>

      <nav className="flex-1 p-3 space-y-1">
        {navItems.map(({ href, label, icon: Icon }) => (
          <Link
            key={href}
            href={href}
            className={cn(
              'flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium transition-colors',
              pathname.startsWith(href)
                ? 'bg-primary/10 text-primary'
                : 'text-muted-foreground hover:bg-accent hover:text-foreground',
            )}
          >
            <Icon className="w-4 h-4 shrink-0" />
            {label}
          </Link>
        ))}
      </nav>

      <div className="p-3 border-t space-y-1">
        <Link
          href="/settings/upgrade"
          className="flex items-center gap-3 px-3 py-2 rounded-md text-sm text-muted-foreground hover:bg-accent hover:text-foreground transition-colors"
        >
          <Zap className="w-4 h-4 shrink-0" />
          Upgrade
        </Link>
        <button
          onClick={handleLogout}
          className="flex items-center gap-3 px-3 py-2 rounded-md text-sm text-muted-foreground hover:bg-accent hover:text-foreground transition-colors w-full"
        >
          <LogOut className="w-4 h-4 shrink-0" />
          Sign out
        </button>
      </div>
    </aside>
  )
}
