'use client'

import { motion } from 'framer-motion'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { ArrowRight, ExternalLink } from 'lucide-react'

export function HeroSection() {
  return (
    <section className="relative flex min-h-screen items-center justify-center overflow-hidden px-4 pt-16">
      {/* Orbital ring — framer-motion continuous rotation */}
      <div className="pointer-events-none absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2">
        {/* Glow blur layer */}
        <motion.div
          animate={{ rotate: 360 }}
          transition={{ duration: 20, repeat: Infinity, ease: 'linear' }}
          style={{
            width: 540,
            height: 540,
            borderRadius: '50%',
            background:
              'conic-gradient(from 0deg, transparent 0%, oklch(0.5 0.2 265 / 20%) 25%, oklch(0.7 0.25 280 / 30%) 50%, oklch(0.5 0.2 265 / 20%) 75%, transparent 100%)',
            filter: 'blur(40px)',
            opacity: 0.5,
            position: 'absolute',
            top: -20,
            left: -20,
          }}
        />
        {/* Ring border layer */}
        <motion.div
          animate={{ rotate: 360 }}
          transition={{ duration: 20, repeat: Infinity, ease: 'linear' }}
          style={{
            width: 500,
            height: 500,
            borderRadius: '50%',
            border: '2px solid transparent',
            background:
              'conic-gradient(from 0deg, transparent 0%, oklch(0.5 0.2 265 / 60%) 25%, oklch(0.7 0.25 280 / 80%) 50%, oklch(0.5 0.2 265 / 60%) 75%, transparent 100%) border-box',
            WebkitMask: 'linear-gradient(#fff 0 0) padding-box, linear-gradient(#fff 0 0)',
            WebkitMaskComposite: 'xor',
            mask: 'linear-gradient(#fff 0 0) padding-box, linear-gradient(#fff 0 0)',
            maskComposite: 'exclude',
            opacity: 0.5,
          }}
        />
      </div>

      {/* Content */}
      <div className="relative z-10 mx-auto max-w-3xl text-center">
        {/* NEW badge pill */}
        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, ease: 'easeOut' }}
          className="mb-8 inline-flex items-center gap-2 rounded-full px-4 py-1.5 text-xs font-medium"
          style={{
            border: '1px solid rgba(255,255,255,0.1)',
            background: 'rgba(255,255,255,0.05)',
            color: 'rgba(255,255,255,0.7)',
          }}
        >
          <span
            className="rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wider text-white"
            style={{ background: 'oklch(0.5 0.22 265)' }}
          >
            New
          </span>
          MCP-native secret management
        </motion.div>

        {/* Headline */}
        <motion.h1
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, ease: 'easeOut', delay: 0.1 }}
          className="text-5xl font-bold text-white sm:text-6xl md:text-7xl"
          style={{ letterSpacing: '-0.03em', lineHeight: '1.05' }}
        >
          Secret vault for{' '}
          <span
            style={{
              background:
                'linear-gradient(90deg, oklch(0.5 0.22 265), oklch(0.7 0.25 280), oklch(0.6 0.25 270))',
              WebkitBackgroundClip: 'text',
              WebkitTextFillColor: 'transparent',
            }}
          >
            AI agents.
          </span>
        </motion.h1>

        {/* Subtitle */}
        <motion.p
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, ease: 'easeOut', delay: 0.3 }}
          className="mx-auto mt-6 max-w-xl text-lg sm:text-xl"
          style={{ color: 'rgba(255,255,255,0.5)' }}
        >
          Human-in-the-loop approval. MCP-native. Zero-trust by design.
        </motion.p>

        {/* CTAs */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, ease: 'easeOut', delay: 0.5 }}
          className="mt-10 flex flex-col items-center gap-4 sm:flex-row sm:justify-center"
        >
          <Link href="/login">
            <Button
              size="lg"
              className="gap-2 rounded-full px-8 text-white"
              style={{
                background: 'oklch(0.5 0.22 265)',
                boxShadow: '0 0 30px -4px oklch(0.5 0.22 265 / 50%)',
              }}
            >
              Get Started <ArrowRight className="h-4 w-4" />
            </Button>
          </Link>
          <a href="https://docs.valt.dev" target="_blank" rel="noopener noreferrer">
            <Button
              variant="outline"
              size="lg"
              className="gap-2 rounded-full px-8"
              style={{
                background: 'transparent',
                borderColor: 'rgba(255,255,255,0.15)',
                color: 'rgba(255,255,255,0.7)',
              }}
            >
              View Docs <ExternalLink className="h-4 w-4" />
            </Button>
          </a>
        </motion.div>

        {/* Social proof */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, ease: 'easeOut', delay: 0.7 }}
          className="mt-16"
        >
          <p
            className="text-xs font-medium uppercase tracking-widest"
            style={{ color: 'rgba(255,255,255,0.25)' }}
          >
            Used by teams building with
          </p>
          <div
            className="mt-3 flex flex-wrap items-center justify-center gap-x-6 gap-y-2 text-sm"
            style={{ color: 'rgba(255,255,255,0.3)' }}
          >
            {['Claude', 'GPT', 'Gemini', 'Cursor', 'Windsurf'].map((name) => (
              <span key={name}>{name}</span>
            ))}
          </div>
        </motion.div>
      </div>
    </section>
  )
}
