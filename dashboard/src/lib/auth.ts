import { cookies } from 'next/headers'

export interface SessionUser {
  id: string
  email?: string
}

function decodeJWT(token: string): Record<string, unknown> | null {
  try {
    const payload = token.split('.')[1]
    return JSON.parse(Buffer.from(payload, 'base64url').toString())
  } catch {
    return null
  }
}

export async function getSession(): Promise<SessionUser | null> {
  const cookieStore = await cookies()
  const token = cookieStore.get('valt_access_token')?.value
  if (!token) return null
  const payload = decodeJWT(token)
  if (!payload || !payload.sub) return null
  const exp = payload.exp as number
  if (exp && Date.now() / 1000 > exp) return null
  return { id: payload.sub as string, email: payload.email as string | undefined }
}
