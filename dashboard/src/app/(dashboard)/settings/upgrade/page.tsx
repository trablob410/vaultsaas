'use client'

interface UsageBarProps {
  label: string
  current: number
  limit: number
}

function UsageBar({ label, current, limit }: UsageBarProps) {
  const pct = Math.min((current / limit) * 100, 100)
  return (
    <div className="space-y-1">
      <div className="flex justify-between text-sm">
        <span>{label}</span>
        <span className="text-muted-foreground">{current} / {limit}</span>
      </div>
      <div className="h-1.5 rounded-full bg-muted overflow-hidden">
        <div
          className="h-full bg-primary rounded-full transition-all"
          style={{ width: `${pct}%` }}
        />
      </div>
    </div>
  )
}

export default function UpgradePage() {
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
          <span className="px-2 py-0.5 rounded-full text-xs bg-secondary text-secondary-foreground">Free</span>
        </div>
        <p className="text-sm text-muted-foreground">
          Free tier includes 50 secrets, 3 agents, and 1,000 API requests per day.
        </p>
      </div>

      {/* Usage summary */}
      <div className="border rounded-lg p-4 space-y-3">
        <h2 className="font-medium">Free Tier Limits</h2>
        <UsageBar label="Secrets" current={0} limit={50} />
        <UsageBar label="Agents" current={0} limit={3} />
        <UsageBar label="API Requests (today)" current={0} limit={1000} />
      </div>

      {/* Upgrade CTA */}
      <div className="border rounded-lg p-6 bg-primary/5 space-y-3">
        <h2 className="font-semibold">Upgrade to Pro</h2>
        <ul className="text-sm text-muted-foreground space-y-1 list-disc list-inside">
          <li>Unlimited secrets and agents</li>
          <li>50,000 API requests per day</li>
          <li>Dynamic secret providers</li>
          <li>Priority support</li>
        </ul>
        <a
          href="mailto:hello@valt.dev?subject=Upgrade to Pro"
          className="inline-flex items-center px-4 py-2 bg-primary text-primary-foreground rounded-md text-sm font-medium hover:bg-primary/90"
        >
          Contact us to upgrade
        </a>
      </div>
    </div>
  )
}
