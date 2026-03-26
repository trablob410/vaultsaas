'use client'

import { motion } from 'framer-motion'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { ArrowRight } from 'lucide-react'

export function CtaBannerSection() {
  return (
    <section className="relative px-6 py-24 md:py-32">
      {/* Top divider */}
      <div
        className="absolute inset-x-0 top-0 h-px"
        style={{ background: 'linear-gradient(90deg, transparent, rgba(255,255,255,0.08) 50%, transparent)' }}
      />

      <div className="relative mx-auto max-w-4xl text-center">
        {/* Ambient glow blob */}
        <div
          className="pointer-events-none absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2"
          style={{
            width: 400,
            height: 400,
            borderRadius: '50%',
            background:
              'radial-gradient(circle, oklch(0.5 0.22 265 / 15%) 0%, oklch(0.7 0.25 280 / 8%) 50%, transparent 70%)',
            filter: 'blur(40px)',
          }}
        />

        {/* Heading */}
        <motion.h2
          initial={{ opacity: 0, y: 40 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: '-100px' }}
          transition={{ duration: 0.6, ease: 'easeOut' }}
          className="relative text-3xl font-bold tracking-tight text-white md:text-5xl"
          style={{ letterSpacing: '-0.02em' }}
        >
          Ready to secure your AI agents?
        </motion.h2>

        {/* Subtext */}
        <motion.p
          initial={{ opacity: 0, y: 40 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: '-100px' }}
          transition={{ duration: 0.6, ease: 'easeOut', delay: 0.15 }}
          className="relative mx-auto mt-5 max-w-lg"
          style={{ color: 'rgba(255,255,255,0.5)' }}
        >
          Get started in minutes. No credit card required.
        </motion.p>

        {/* CTA button */}
        <motion.div
          initial={{ opacity: 0, y: 40 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: '-100px' }}
          transition={{ duration: 0.6, ease: 'easeOut', delay: 0.3 }}
          className="relative mt-10"
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
              Get Started Free <ArrowRight className="h-4 w-4" />
            </Button>
          </Link>
        </motion.div>
      </div>
    </section>
  )
}
