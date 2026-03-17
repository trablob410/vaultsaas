import { Shield, Lock } from 'lucide-react'

const BACKEND_URL = process.env.BACKEND_URL ?? 'http://localhost:8080'

export default function LoginPage() {
  return (
    <div className="flex flex-col items-center gap-8">
      <div className="flex items-center gap-3">
        <div className="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center">
          <Shield className="w-5 h-5 text-primary" />
        </div>
        <div>
          <h1 className="text-xl font-semibold tracking-tight">Valt</h1>
          <p className="text-xs text-muted-foreground">AI Secret Vault</p>
        </div>
      </div>

      <div className="w-full rounded-2xl border bg-card p-8 shadow-lg">
        <div className="text-center mb-8">
          <Lock className="w-8 h-8 mx-auto mb-4 text-primary" />
          <h2 className="text-2xl font-bold">Welcome back</h2>
          <p className="text-muted-foreground mt-2 text-sm">
            Secure secret management for AI agents
          </p>
        </div>

        <a
          href={`${BACKEND_URL}/api/v1/auth/google`}
          className="flex items-center justify-center gap-3 w-full rounded-lg border bg-background hover:bg-accent transition-colors px-4 py-3 text-sm font-medium"
        >
          <svg className="w-5 h-5" viewBox="0 0 24 24">
            <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
            <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
            <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
            <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
          </svg>
          Sign in with Google
        </a>

        <p className="text-center text-xs text-muted-foreground mt-6">
          By signing in, you agree to our{' '}
          <span className="underline cursor-pointer">Terms of Service</span>
          {' '}and{' '}
          <span className="underline cursor-pointer">Privacy Policy</span>
        </p>
      </div>

      <p className="text-xs text-muted-foreground text-center">
        Human-in-the-loop approval for all AI agent secret access
      </p>
    </div>
  )
}
