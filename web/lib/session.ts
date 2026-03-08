import { SecretManagerServiceClient } from '@google-cloud/secret-manager'
import { jwtVerify } from 'jose'
import { cookies } from 'next/headers'
import { cache } from 'react'
import 'server-only'

// https://nextjs.org/docs/app/guides/authentication

const client = new SecretManagerServiceClient()
let cachedKey: string | undefined = undefined

export const getCustomerId = cache(async (): Promise<string> => {
  const cookie = (await cookies()).get('session')?.value
  if (!cookie) throw new Error('No session')
  const session = await decrypt(cookie)
  if (!session?.sub) throw new Error('Customer ID not found in session token')
  return session.sub
})

export const createSession = async function (token: string, expiresAt: Date) {
  const cookieStore = await cookies()
  cookieStore.set('session', token, {
    httpOnly: true,
    secure: true,
    expires: expiresAt,
    sameSite: 'lax',
    path: '/',
  })
}

export const getSession = cache(async (): Promise<string> => {
  const cookieStore = await cookies()
  const session = cookieStore.get('session')
  if (!session) throw new Error('session not found in cookie')
  return session.value
})

export async function updateSession() {
  const session = (await cookies()).get('session')?.value
  const payload = await decrypt(session)

  if (!session || !payload) {
    return null
  }

  const expires = new Date(Date.now() + 7 * 24 * 60 * 60 * 1000)

  const cookieStore = await cookies()
  cookieStore.set('session', session, {
    httpOnly: true,
    secure: true,
    expires: expires,
    sameSite: 'lax',
    path: '/',
  })
}

export async function deleteSession() {
  const cookieStore = await cookies()
  cookieStore.delete('session')
}

export async function getSecretKey() {
  if (cachedKey) return cachedKey
  try {
    const [accessResponse] = await client.accessSecretVersion({
      name: `projects/${process.env.GCP_PROJECT_ID}/secrets/JWT_SIGNING_KEY/versions/latest`,
    })
    const data = accessResponse?.payload?.data
    if (!data) throw new Error('Secret payload is empty')
    cachedKey = data.toString('utf8')
    return cachedKey
  } catch (error) {
    console.error('Failed to fetch secret:', error)
    throw new Error('Could not retrieve signing key')
  }
}

export const authentication = cache(async () => {
  const cookie = (await cookies()).get('session')?.value
  if (!cookie) {
    return { isAuth: false }
  }

  try {
    const session = await decrypt(cookie)
    if (!session || !session.sub) {
      return { isAuth: false }
    }

    return { isAuth: true }
  } catch (error) {
    console.error('Session decryption failed:', error)
    return { isAuth: false }
  }
})

export async function decrypt(session: string | undefined = '') {
  const encodedKey = new TextEncoder().encode(await getSecretKey())
  try {
    const { payload } = await jwtVerify(session, encodedKey, {
      algorithms: ['HS256'],
    })
    return payload
  } catch (error) {
    console.error('Failed to verify session', error)
  }
}

// export async function encrypt(payload: SessionPayload) {
//     return new SignJWT(payload)
//         .setProtectedHeader({ alg: 'HS256' })
//         .setIssuedAt()
//         .setExpirationTime('7d')
//         .sign(encodedKey)
// }
