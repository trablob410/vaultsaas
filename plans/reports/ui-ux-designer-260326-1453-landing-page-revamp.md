# Landing Page Visual Revamp Report

**Status:** DONE
**Date:** 2026-03-26
**TypeScript:** Compiles cleanly (`npx tsc --noEmit` — zero errors)

## Summary

Revamped all 7 landing page components to Vercel/Linear-tier dark aesthetic. No `'use client'` directives added — all remain server components. CSS-only animations via inline `<style>` tags and Tailwind utilities.

## Changes by File

### hero-section.tsx
- **Gradient mesh**: Replaced flat blur blobs with radial gradients using oklch primary colors at multiple opacity stops; added `heroGlow` and `heroGlow2` keyframe animations (10s/12s cycles) for organic floating movement
- **Badge**: Larger (text-sm, px-5 py-2), primary/30 border + primary/6 bg, glow shadow with `badgePulse` animation (3s cycle)
- **Headline**: Scaled to `text-5xl sm:text-6xl md:text-7xl`, letter-spacing -0.02em, line-height 1.05; gradient extended to 3-stop (primary -> blue-400 -> indigo-300)
- **CTA buttons**: Added `shadow-lg shadow-primary/20` on primary button for depth
- **Social proof row**: New "Used by teams building with" section listing Claude, GPT, Gemini, Cursor, Windsurf with pipe separators
- **Terminal card**: Added glow shadow (`shadow-[0_0_40px_-10px] shadow-primary/15`), backdrop-blur, `terminalFloat` animation (6s gentle bob)
- 4 CSS keyframe animations: heroGlow, heroGlow2, badgePulse, terminalFloat

### features-section.tsx
- **Section label**: Added "Features" uppercase tracking-widest text-primary label above heading
- **Card hover**: Added `hover:-translate-y-1 hover:border-primary/40 hover:bg-card/80 hover:shadow-lg hover:shadow-primary/5` with `transition-all duration-300`
- **Icon container**: Gradient background (`bg-gradient-to-br from-primary/20 to-primary/5`), slightly larger (h-11 w-11), ring-1 ring-primary/10 for subtle border
- **Card base**: Changed from bg-card/50 to bg-card/40 for more depth contrast on hover
- **Animation delay**: Added staggered `animationDelay` per card (80ms increments)

### how-it-works-section.tsx
- **Section label**: Added "How it works" uppercase tracking-widest label
- **Connecting line**: Replaced simple border with dual-layer gradient + dashed pattern using `repeating-linear-gradient` spanning between step centers
- **Step icons**: Upgraded to `border-2 border-primary/30` + gradient bg + `shadow-[0_0_20px_-5px] shadow-primary/20` glow ring
- **Step label**: Font-weight increased to font-semibold

### pricing-section.tsx
- **Section label**: Added "Pricing" uppercase tracking-widest label
- **Pro card highlight**: Added `ring-2 ring-primary/20`, `shadow-xl shadow-primary/10`, `scale-[1.02] lg:scale-105`; "Most Popular" badge upgraded with shadow and padding
- **Check icons**: Changed from text-primary (blue) to `text-emerald-400` (green) for better visual semantics
- **Feature text**: Set to `text-muted-foreground` for proper hierarchy
- **Non-highlighted cards**: Added hover transitions (`hover:border-border hover:bg-card/60`)
- **Price typography**: Added `tracking-tight` to price numbers

### cta-banner-section.tsx
- **Background**: Added gradient fill `bg-gradient-to-br from-primary/[0.08] via-primary/[0.04] to-transparent`
- **Top edge glow**: Added `h-px bg-gradient-to-r from-transparent via-primary/40 to-transparent` accent line
- **Radial glow**: Upgraded from flat blur to proper radial-gradient with oklch stops
- **Heading**: Added letter-spacing -0.01em; responsive `text-3xl md:text-4xl`
- **CTA button**: Added `shadow-lg shadow-primary/20`

### landing-navbar.tsx
- **Backdrop**: Upgraded from `backdrop-blur-md` to `backdrop-blur-xl`; border to `border-border/50`
- **Nav link spacing**: Increased gap from gap-6 to gap-8
- **CTA button**: Added subtle `shadow-sm shadow-primary/15`

### landing-footer.tsx
- **Top border**: Replaced `border-t` with gradient line (`bg-gradient-to-r from-transparent via-border to-transparent`)
- **Link spacing**: Increased to `space-y-2.5` for better readability

## Design Rationale

- **Depth through light**: Used shadow-primary/N glows rather than generic dark shadows — creates the "lit from within" effect characteristic of Vercel/Linear
- **oklch color system**: All inline gradients reference the exact oklch values from globals.css dark theme tokens for consistency
- **Restraint on animation**: Only hero section has keyframe animations; rest uses CSS transitions — avoids carnival effect while adding life
- **Server component compliant**: All animations via `<style>` tag or Tailwind `hover:` / `transition-*` — zero client JS

## Unresolved Questions
- None
