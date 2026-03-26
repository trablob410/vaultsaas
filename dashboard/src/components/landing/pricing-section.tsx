'use client'

import { motion } from 'framer-motion'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { Check } from 'lucide-react'

interface Tier {
  name: string
  price: string
  period: string
  description: string
  features: string[]
  cta: string
  href: string
  highlighted?: boolean
}

const tiers: Tier[] = [
  {
    name: 'Free',
    price: '$0',
    period: 'forever',
    description: 'For individuals exploring AI agent workflows.',
    features: ['50 secrets', '3 agents', '1,000 requests / day', 'Email notifications', 'Community support'],
    cta: 'Get Started',
    href: '/login',
  },
  {
    name: 'Pro',
    price: '$19',
    period: 'per user / month',
    description: 'For teams shipping AI-powered products.',
    features: [
      '500 secrets',
      '25 agents',
      '10,000 requests / day',
      'Slack & Telegram notifications',
      'Dynamic secrets',
      'Custom policies',
      'Priority support',
    ],
    cta: 'Upgrade to Pro',
    href: '/login',
    highlighted: true,
  },
  {
    name: 'Team',
    price: '$39',
    period: 'per user / month',
    description: 'For organizations with compliance needs.',
    features: [
      'Unlimited secrets & agents',
      '50,000 requests / day',
      'SSO & SAML',
      'Advanced audit trail',
      'Webhook integrations',
      'Dedicated support',
      'SLA guarantee',
    ],
    cta: 'Contact Sales',
    href: '/login',
  },
]

const cardVariants = {
  hidden: { opacity: 0, y: 40 },
  show: { opacity: 1, y: 0 },
}

export function PricingSection() {
  return (
    <section id="pricing" className="relative px-6 py-24 md:py-32">
      {/* Top divider */}
      <div
        className="absolute inset-x-0 top-0 h-px"
        style={{ background: 'linear-gradient(90deg, transparent, rgba(255,255,255,0.08) 50%, transparent)' }}
      />

      <div className="mx-auto max-w-6xl">
        {/* Header */}
        <motion.div
          initial={{ opacity: 0, y: 40 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: '-100px' }}
          transition={{ duration: 0.6 }}
          className="mb-16 text-center"
        >
          <p className="mb-3 text-sm font-medium uppercase tracking-widest" style={{ color: 'oklch(0.6 0.25 270)' }}>
            Pricing
          </p>
          <h2 className="text-3xl font-bold tracking-tight text-white sm:text-4xl" style={{ letterSpacing: '-0.02em' }}>
            Simple, transparent pricing
          </h2>
          <p className="mx-auto mt-4 max-w-xl" style={{ color: 'rgba(255,255,255,0.5)' }}>
            Start free. Scale as your AI fleet grows.
          </p>
        </motion.div>

        {/* Cards */}
        <div className="grid items-start gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {tiers.map((tier, i) => (
            <motion.div
              key={tier.name}
              variants={cardVariants}
              initial="hidden"
              whileInView="show"
              viewport={{ once: true, margin: '-80px' }}
              transition={{ duration: 0.5, ease: 'easeOut', delay: i * 0.12 }}
              whileHover={{ y: -4 }}
              className="relative flex flex-col rounded-xl p-6"
              style={{
                background: tier.highlighted ? 'rgba(255,255,255,0.04)' : 'rgba(255,255,255,0.02)',
                border: tier.highlighted ? 'none' : '1px solid rgba(255,255,255,0.06)',
                transform: tier.highlighted ? 'scale(1.03)' : undefined,
              }}
            >
              {/* Glowing ring border for Pro */}
              {tier.highlighted && (
                <div
                  className="pointer-events-none absolute -inset-px rounded-xl"
                  style={{
                    background:
                      'linear-gradient(130deg, oklch(0.5 0.22 265 / 60%), oklch(0.7 0.25 280 / 30%), oklch(0.6 0.25 270 / 40%), oklch(0.5 0.22 265 / 60%))',
                    backgroundSize: '300% 300%',
                    animation: 'proBorderRotate 6s ease infinite',
                    WebkitMask: 'linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0)',
                    WebkitMaskComposite: 'xor',
                    mask: 'linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0)',
                    maskComposite: 'exclude',
                    padding: '1px',
                    borderRadius: '0.75rem',
                  }}
                />
              )}

              {tier.highlighted && (
                <span
                  className="absolute -top-3.5 left-1/2 -translate-x-1/2 rounded-full px-4 py-1 text-xs font-semibold text-white"
                  style={{
                    background: 'oklch(0.5 0.22 265)',
                    boxShadow: '0 0 20px -4px oklch(0.5 0.22 265 / 50%)',
                  }}
                >
                  Most Popular
                </span>
              )}

              <div className="relative mb-6">
                <h3 className="text-lg font-semibold text-white">{tier.name}</h3>
                <div className="mt-3 flex items-baseline gap-1">
                  <span className="text-4xl font-bold tracking-tight text-white">{tier.price}</span>
                  <span className="text-sm" style={{ color: 'rgba(255,255,255,0.4)' }}>{tier.period}</span>
                </div>
                <p className="mt-2 text-sm" style={{ color: 'rgba(255,255,255,0.45)' }}>{tier.description}</p>
              </div>

              <ul className="relative mb-8 flex-1 space-y-3">
                {tier.features.map((feature) => (
                  <li key={feature} className="flex items-start gap-2.5 text-sm">
                    <Check className="mt-0.5 h-4 w-4 shrink-0" style={{ color: '#4ade80' }} />
                    <span style={{ color: 'rgba(255,255,255,0.55)' }}>{feature}</span>
                  </li>
                ))}
              </ul>

              <Link href={tier.href} className="relative">
                <Button
                  className="w-full rounded-lg text-white"
                  style={
                    tier.highlighted
                      ? { background: 'oklch(0.5 0.22 265)', boxShadow: '0 0 20px -4px oklch(0.5 0.22 265 / 40%)' }
                      : {
                          background: 'transparent',
                          border: '1px solid rgba(255,255,255,0.15)',
                          color: 'rgba(255,255,255,0.7)',
                        }
                  }
                >
                  {tier.cta}
                </Button>
              </Link>
            </motion.div>
          ))}
        </div>
      </div>

      <style>{`
        @keyframes proBorderRotate {
          0% { background-position: 0% 50%; }
          50% { background-position: 100% 50%; }
          100% { background-position: 0% 50%; }
        }
      `}</style>
    </section>
  )
}
