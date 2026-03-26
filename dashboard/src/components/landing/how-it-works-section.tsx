'use client'

import { motion } from 'framer-motion'
import { KeyRound, MessageSquare, CheckCircle2 } from 'lucide-react'

const steps = [
  {
    num: '01',
    icon: KeyRound,
    title: 'Store secrets',
    description:
      'Add credentials to your vault. Encrypted at rest with envelope encryption — the server never sees plaintext.',
  },
  {
    num: '02',
    icon: MessageSquare,
    title: 'AI requests access',
    description:
      'Your agent calls the MCP server or CLI, providing a reason for access. A notification is sent instantly.',
  },
  {
    num: '03',
    icon: CheckCircle2,
    title: 'Human approves',
    description:
      'One-click approve via Slack, Telegram, or email. The agent gets time-limited access. Full audit trail logged.',
  },
]

export function HowItWorksSection() {
  return (
    <section id="how-it-works" className="relative px-6 py-24 md:py-32">
      {/* Top divider */}
      <div
        className="absolute inset-x-0 top-0 h-px"
        style={{
          background: 'linear-gradient(90deg, transparent, rgba(255,255,255,0.08) 50%, transparent)',
        }}
      />

      <div className="mx-auto max-w-6xl">
        {/* Header */}
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
            How it works
          </p>
          <h2
            className="text-3xl font-bold tracking-tight text-white sm:text-4xl"
            style={{ letterSpacing: '-0.02em' }}
          >
            Three steps to secure access
          </h2>
          <p className="mx-auto mt-4 max-w-xl" style={{ color: 'rgba(255,255,255,0.5)' }}>
            Three steps between your AI agent and production credentials.
          </p>
        </motion.div>

        {/* Steps grid */}
        <div className="relative grid gap-12 sm:grid-cols-3 sm:gap-8">
          {/* Connecting dashed line (desktop only) */}
          <div className="pointer-events-none absolute left-[16.67%] right-[16.67%] top-[3.5rem] hidden sm:block">
            <svg className="w-full overflow-visible" style={{ height: '2px' }}>
              <line
                x1="0"
                y1="1"
                x2="100%"
                y2="1"
                stroke="oklch(0.5 0.22 265 / 25%)"
                strokeWidth="1.5"
                strokeDasharray="8 6"
              />
            </svg>
          </div>

          {steps.map((s, i) => (
            <motion.div
              key={s.num}
              initial={{ opacity: 0, y: 40 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true, margin: '-80px' }}
              transition={{ duration: 0.6, delay: i * 0.2 }}
              className="relative text-center"
            >
              {/* Large watermark number */}
              <div
                className="mx-auto mb-2 text-6xl font-black sm:text-7xl"
                style={{
                  color: 'transparent',
                  WebkitTextStroke: '1px rgba(255,255,255,0.08)',
                  lineHeight: 1,
                }}
              >
                {s.num}
              </div>

              {/* Glowing icon circle */}
              <div
                className="mx-auto mb-5 flex h-14 w-14 items-center justify-center rounded-full"
                style={{
                  background: 'oklch(0.5 0.22 265 / 12%)',
                  border: '1px solid oklch(0.5 0.22 265 / 25%)',
                  boxShadow: '0 0 30px -8px oklch(0.5 0.22 265 / 40%)',
                }}
              >
                <s.icon className="h-6 w-6" style={{ color: 'oklch(0.6 0.25 270)' }} />
              </div>

              <h3 className="mb-2 text-lg font-semibold text-white">{s.title}</h3>
              <p
                className="mx-auto max-w-xs text-sm leading-relaxed"
                style={{ color: 'rgba(255,255,255,0.45)' }}
              >
                {s.description}
              </p>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  )
}
