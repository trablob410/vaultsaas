import { LandingNavbar } from './landing-navbar'
import { HeroSection } from './hero-section'
import { FeaturesSection } from './features-section'
import { HowItWorksSection } from './how-it-works-section'
import { PricingSection } from './pricing-section'
import { CtaBannerSection } from './cta-banner-section'
import { LandingFooter } from './landing-footer'

export function LandingPage() {
  return (
    <div className="relative min-h-screen text-white" style={{ background: '#000' }}>
      {/* Star dots background */}
      <div className="fixed inset-0 -z-10" style={{ background: '#000' }}>
        <div
          className="absolute inset-0"
          style={{
            backgroundImage: [
              'radial-gradient(1px 1px at 20% 30%, rgba(255,255,255,0.15) 0%, transparent 100%)',
              'radial-gradient(1px 1px at 80% 70%, rgba(255,255,255,0.1) 0%, transparent 100%)',
              'radial-gradient(1px 1px at 40% 80%, rgba(255,255,255,0.12) 0%, transparent 100%)',
              'radial-gradient(1px 1px at 60% 20%, rgba(255,255,255,0.08) 0%, transparent 100%)',
            ].join(', '),
            backgroundSize: '200px 200px, 150px 150px, 250px 250px, 180px 180px',
          }}
        />
      </div>

      <LandingNavbar />
      <main>
        <HeroSection />
        <FeaturesSection />
        <HowItWorksSection />
        <PricingSection />
        <CtaBannerSection />
      </main>
      <LandingFooter />
    </div>
  )
}
