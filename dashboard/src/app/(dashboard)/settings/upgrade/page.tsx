'use client'

import { useEffect, useState } from 'react'
import { api } from '@/lib/api-client'

interface UsageBarProps {
  label: string
  current: number
  limit: number
}

function UsageBar({ label, current, limit }: UsageBarProps) {
  const isUnlimited = limit < 0
  const pct = isUnlimited ? 0 : Math.min((current / limit) * 100, 100)
  return (
    <div className="space-y-1">
      <div className="flex justify-between text-sm">
        <span>{label}</span>
        <span className="text-muted-foreground">
          {current} / {isUnlimited ? '∞' : limit.toLocaleString()}
        </span>
      </div>
      <div className="h-1.5 rounded-full bg-muted overflow-hidden">
        <div
          className="h-full bg-primary rounded-full transition-all"
          style={{ width: isUnlimited ? '0%' : `${pct}%` }}
        />
      </div>
    </div>
  )
}

const PLAN_LABELS: Record<string, string> = {
  free: 'Free',
  pro: 'Pro',
  team: 'Team',
}

const PLAN_DESCRIPTIONS: Record<string, string> = {
  free: 'Free tier includes 50 secrets, 3 agents, and 1,000 API requests per day.',
  pro: 'Pro plan includes 500 secrets, 25 agents, and 10,000 API requests per day.',
  team: 'Team plan includes unlimited secrets, unlimited agents, and 50,000 API requests per day.',
}

interface UsageData {
  plan: string
  usage: Record<string, number>
  limits: Record<string, number>
}

export default function UpgradePage() {
  const [data, setData] = useState<UsageData | null>(null)
  const [loading, setLoading] = useState(true)
  const [upgrading, setUpgrading] = useState(false)

  useEffect(() => {
    async function load() {
      try {
        const orgs = await api.orgs.list()
        if (orgs.orgs.length > 0) {
          const usage = await api.usage.get(orgs.orgs[0].id)
          setData(usage)
        }
      } catch {
        // fail silently — show static fallback
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [])

  const plan = data?.plan ?? 'free'
  const usage = data?.usage ?? { secrets_count: 0, agents_count: 0, requests_today: 0 }
  const limits = data?.limits ?? { secrets_count: 50, agents_count: 3, requests_today: 1000 }

  async function handleUpgrade(targetPlan: string) {
    setUpgrading(true)
    try {
      const { url } = await api.billing.createCheckout({
        plan: targetPlan,
        success_url: `${window.location.origin}/settings/upgrade?success=1`,
        cancel_url: `${window.location.origin}/settings/upgrade`,
      })
      window.location.href = url
    } catch {
      setUpgrading(false)
    }
  }

  async function handleManageBilling() {
    try {
      const { url } = await api.billing.createPortal({
        return_url: `${window.location.origin}/settings/upgrade`,
      })
      window.location.href = url
    } catch {
      // portal unavailable
    }
  }

  if (loading) {
    return <div className="text-sm text-muted-foreground p-4">Loading usage...</div>
  }

  return (
    <div className="space-y-6 max-w-2xl">
      <div>
        <h1 className="text-2xl font-semibold">Plan &amp; Usage</h1>
        <p className="text-sm text-muted-foreground mt-1">Manage your subscription and monitor usage.</p>
      </div>

      {/* Current Plan */}
      <div className="border rounded-lg p-4 space-y-2">
        <div className="flex items-center justify-between">
          <h2 className="font-medium">Current Plan</h2>
          <span className="px-2 py-0.5 rounded-full text-xs bg-secondary text-secondary-foreground">
            {PLAN_LABELS[plan] ?? plan}
          </span>
        </div>
        <p className="text-sm text-muted-foreground">
          {PLAN_DESCRIPTIONS[plan] ?? 'Custom plan.'}
        </p>
        {plan !== 'free' && (
          <button
            onClick={handleManageBilling}
            className="text-sm text-primary hover:underline"
          >
            Manage billing
          </button>
        )}
      </div>

      {/* Usage summary */}
      <div className="border rounded-lg p-4 space-y-3">
        <h2 className="font-medium">Usage</h2>
        <UsageBar label="Secrets" current={usage.secrets_count} limit={limits.secrets_count} />
        <UsageBar label="Agents" current={usage.agents_count} limit={limits.agents_count} />
        <UsageBar label="API Requests (today)" current={usage.requests_today} limit={limits.requests_today} />
      </div>

      {/* Upgrade CTA — only show if on free plan */}
      {plan === 'free' && (
        <>
          <div className="border rounded-lg p-6 bg-primary/5 space-y-3">
            <h2 className="font-semibold">Upgrade to Pro — $19/user/mo</h2>
            <ul className="text-sm text-muted-foreground space-y-1 list-disc list-inside">
              <li>500 secrets, 25 agents</li>
              <li>10,000 API requests per day</li>
              <li>Dynamic secret providers</li>
              <li>Priority support</li>
            </ul>
            <button
              onClick={() => handleUpgrade('pro')}
              disabled={upgrading}
              className="inline-flex items-center px-4 py-2 bg-primary text-primary-foreground rounded-md text-sm font-medium hover:bg-primary/90 disabled:opacity-50"
            >
              {upgrading ? 'Redirecting...' : 'Upgrade to Pro'}
            </button>
          </div>

          <div className="border rounded-lg p-6 space-y-3">
            <h2 className="font-semibold">Team — $39/user/mo</h2>
            <ul className="text-sm text-muted-foreground space-y-1 list-disc list-inside">
              <li>Unlimited secrets and agents</li>
              <li>50,000 API requests per day</li>
              <li>SSO &amp; advanced audit</li>
              <li>Dedicated support</li>
            </ul>
            <button
              onClick={() => handleUpgrade('team')}
              disabled={upgrading}
              className="inline-flex items-center px-4 py-2 border border-primary text-primary rounded-md text-sm font-medium hover:bg-primary/5 disabled:opacity-50"
            >
              {upgrading ? 'Redirecting...' : 'Upgrade to Team'}
            </button>
          </div>
        </>
      )}
    </div>
  )
}
