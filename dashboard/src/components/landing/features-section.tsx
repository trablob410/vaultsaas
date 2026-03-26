'use client'

import { motion } from 'framer-motion'
import { Shield, Lock } from 'lucide-react'

const features = [
  {
    badge: 'Approval Workflow',
    heading: 'Human approves, AI proceeds',
    description:
      'Multi-channel notifications via Slack, Telegram, and email. One-click approve or reject from anywhere.',
    tags: ['Slack', 'Telegram', 'Email', 'One-click'],
    visual: 'terminal' as const,
  },
  {
    badge: 'MCP Integration',
    heading: 'Built for AI agents',
    description:
      'MCP server and CLI built for agents. They request access with a reason, humans stay in control.',
    tags: ['MCP Server', 'CLI', 'Agent Tokens'],
    visual: 'flow' as const,
  },
  {
    badge: 'Enterprise Security',
    heading: 'Zero-trust by design',
    description:
      'AES-256-GCM envelope encryption. Role-based access control. Append-only hash-chain audit trail.',
    tags: ['AES-256', 'RBAC', 'Audit Trail', 'Encryption'],
    visual: 'shield' as const,
  },
]

function TerminalVisual() {
  return (
    <div
      className="w-full overflow-hidden rounded-xl font-mono text-sm"
      style={{
        background: 'rgba(255,255,255,0.03)',
        border: '1px solid rgba(255,255,255,0.08)',
        boxShadow: '0 0 60px -20px oklch(0.5 0.22 265 / 20%)',
      }}
    >
      <div
        className="flex items-center gap-1.5 px-4 py-2.5"
        style={{ borderBottom: '1px solid rgba(255,255,255,0.06)' }}
      >
        <span className="h-2.5 w-2.5 rounded-full" style={{ background: 'rgba(255,80,80,0.5)' }} />
        <span className="h-2.5 w-2.5 rounded-full" style={{ background: 'rgba(255,190,50,0.5)' }} />
        <span className="h-2.5 w-2.5 rounded-full" style={{ background: 'rgba(80,200,80,0.5)' }} />
        <span className="ml-2 text-xs" style={{ color: 'rgba(255,255,255,0.3)' }}>terminal</span>
      </div>
      <div className="space-y-1 p-4" style={{ color: 'rgba(255,255,255,0.6)' }}>
        <p><span style={{ color: '#4ade80' }}>$</span> valt secret get DATABASE_URL --reason &quot;deploy&quot;</p>
        <p style={{ color: 'rgba(250,204,21,0.8)' }}>⏳ Waiting for approval...</p>
        <p style={{ color: '#4ade80' }}>✓ Approved by admin@acme.com (Slack)</p>
        <p style={{ color: 'rgba(255,255,255,0.5)' }}>postgresql://prod:****@db.acme.com:5432/app</p>
      </div>
    </div>
  )
}

function FlowVisual() {
  return (
    <div
      className="flex w-full flex-col gap-4 rounded-xl p-8"
      style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.08)' }}
    >
      {['AI Agent requests access', 'Human reviews & approves', 'Credential issued securely'].map(
        (step, i) => (
          <div key={step} className="flex items-center gap-4">
            <div
              className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full text-sm font-bold text-white"
              style={{ background: 'oklch(0.5 0.22 265)' }}
            >
              {i + 1}
            </div>
            <span className="text-sm" style={{ color: 'rgba(255,255,255,0.6)' }}>{step}</span>
          </div>
        )
      )}
    </div>
  )
}

function ShieldVisual() {
  return (
    <div
      className="flex w-full items-center justify-center rounded-xl p-8"
      style={{
        background: 'rgba(255,255,255,0.03)',
        border: '1px solid rgba(255,255,255,0.08)',
        minHeight: '200px',
      }}
    >
      <div className="relative">
        <Shield className="h-20 w-20" style={{ color: 'oklch(0.5 0.22 265 / 30%)' }} />
        <Lock
          className="absolute left-1/2 top-1/2 h-8 w-8 -translate-x-1/2 -translate-y-1/2"
          style={{ color: 'oklch(0.6 0.25 270)' }}
        />
      </div>
    </div>
  )
}

const visuals = { terminal: TerminalVisual, flow: FlowVisual, shield: ShieldVisual }

export function FeaturesSection() {
  return (
    <section id="features" className="relative px-6 py-24 md:py-32">
      <div className="mx-auto max-w-6xl">
        {/* Section header */}
        <motion.div
          initial={{ opacity: 0, y: 40 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: '-100px' }}
          transition={{ duration: 0.6 }}
          className="mb-20 text-center"
        >
          <p
            className="mb-3 text-sm font-medium uppercase tracking-widest"
            style={{ color: 'oklch(0.6 0.25 270)' }}
          >
            Features
          </p>
          <h2
            className="text-3xl font-bold tracking-tight text-white sm:text-4xl"
            style={{ letterSpacing: '-0.02em' }}
          >
            Everything you need to secure AI agents
          </h2>
        </motion.div>

        {/* Alternating feature rows */}
        <div className="flex flex-col gap-24 md:gap-32">
          {features.map((f, i) => {
            const Visual = visuals[f.visual]
            const imageLeft = i % 2 === 0
            return (
              <div key={f.badge} className="grid items-start gap-16 md:grid-cols-2">
                {/* Visual — animates from the side it appears on */}
                <motion.div
                  initial={{ opacity: 0, x: imageLeft ? -40 : 40 }}
                  whileInView={{ opacity: 1, x: 0 }}
                  viewport={{ once: true, margin: '-100px' }}
                  transition={{ duration: 0.6, ease: 'easeOut' }}
                  className={imageLeft ? 'md:order-1' : 'md:order-2'}
                >
                  <Visual />
                </motion.div>

                {/* Text — animates from opposite side */}
                <motion.div
                  initial={{ opacity: 0, x: imageLeft ? 40 : -40 }}
                  whileInView={{ opacity: 1, x: 0 }}
                  viewport={{ once: true, margin: '-100px' }}
                  transition={{ duration: 0.6, ease: 'easeOut', delay: 0.1 }}
                  className={imageLeft ? 'md:order-2' : 'md:order-1'}
                >
                  <span
                    className="mb-4 inline-block rounded-full px-3 py-1 text-xs font-medium"
                    style={{
                      border: '1px solid rgba(255,255,255,0.1)',
                      background: 'rgba(255,255,255,0.05)',
                      color: 'oklch(0.7 0.15 280)',
                    }}
                  >
                    {f.badge}
                  </span>
                  <h3
                    className="mb-4 text-2xl font-bold text-white sm:text-3xl"
                    style={{ letterSpacing: '-0.02em' }}
                  >
                    {f.heading}
                  </h3>
                  <p className="mb-6 text-base leading-relaxed" style={{ color: 'rgba(255,255,255,0.5)' }}>
                    {f.description}
                  </p>
                  <div className="flex flex-wrap gap-2">
                    {f.tags.map((tag) => (
                      <span
                        key={tag}
                        className="rounded-full px-3 py-1 text-xs"
                        style={{ border: '1px solid rgba(255,255,255,0.15)', color: 'rgba(255,255,255,0.5)' }}
                      >
                        {tag}
                      </span>
                    ))}
                  </div>
                </motion.div>
              </div>
            )
          })}
        </div>
      </div>
    </section>
  )
}
