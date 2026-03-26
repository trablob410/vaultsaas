import { NextRequest, NextResponse } from 'next/server'

// POST /api/auth/set-tokens
// Body: { access_token, refresh_token, expires_in? }
// Sets httpOnly cookies so tokens are not accessible via document.cookie (XSS protection)
export async function POST(req: NextRequest) {
  const body = await req.json().catch(() => null)
  if (!body?.access_token || !body?.refresh_token) {
    return NextResponse.json({ error: 'access_token and refresh_token required' }, { status: 400 })
  }

  const { access_token, refresh_token, expires_in } = body as {
    access_token: string
    refresh_token: string
    expires_in?: number
  }

  const res = NextResponse.json({ ok: true })
  res.cookies.set('valt_access_token', access_token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    path: '/',
    maxAge: expires_in ?? 900,
  })
  res.cookies.set('valt_refresh_token', refresh_token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    path: '/',
    maxAge: 7 * 24 * 60 * 60, // 7 days
  })
  return res
}
