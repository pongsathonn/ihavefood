'use server'

import { createSession, deleteSession } from '@/lib/session'
import { z } from 'zod'

// @source https://github.com/vercel/ai-chatbot

const authFormSchema = z.object({
  email: z.email(),
  password: z.string().min(6),
})

export type LoginActionState = {
  status: 'idle' | 'in_progress' | 'success' | 'failed' | 'invalid_data'
}

export const authenticate = async (
  _: LoginActionState,
  formData: FormData,
): Promise<LoginActionState> => {
  try {
    const { email, password } = authFormSchema.parse({
      email: formData.get('email'),
      password: formData.get('password'),
    })
    await signIn(email, password)
    return { status: 'success' }
  } catch (error) {
    if (error instanceof z.ZodError) {
      return { status: 'invalid_data' }
    }
    console.log('authenticate err', error)
    return { status: 'failed' }
  }
}

const signIn = async function (email: string, password: string): Promise<void> {
  if (!email || !password) {
    throw new Error('Email and password are required')
  }

  const serverUrl = process.env.SERVER_URL
  if (!serverUrl) throw new Error('SERVER_URL is not defined')
  const response = await fetch(`${serverUrl}/auth/login`, {
    method: 'POST',
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ identifier: email, password }),
  })

  if (!response.ok) {
    const error = await response.json().catch(() => ({}))
    throw new Error(
      error.message ?? `Authentication failed: ${response.status}`,
    )
  }

  const token = await response.json()

  if (!token.accessToken || !token.expiresIn) {
    throw new Error('Malformed auth response from server')
  }

  const expiresAt = Number(token.expiresIn)
  if (isNaN(expiresAt)) throw new Error('Invalid token expiration')

  await createSession(token.accessToken, new Date(expiresAt * 1000))
}

// export type RegisterActionState = {
//   status:
//     | 'idle'
//     | 'in_progress'
//     | 'success'
//     | 'failed'
//     | 'user_exists'
//     | 'invalid_data'
// }

// export const register = async (
//   _: RegisterActionState,
//   formData: FormData,
// ): Promise<RegisterActionState> => {
//   try {
//     const validatedData = authFormSchema.parse({
//       email: formData.get('email'),
//       password: formData.get('password'),
//     })

//     const [user] = await getUser(validatedData.email)
//     // fetch getauthbyidentifier

//     if (user) {
//       return { status: 'user_exists' } as RegisterActionState
//     }
//     await createUser(validatedData.email, validatedData.password)
//     await signUp('credentials', {
//       email: validatedData.email,
//       password: validatedData.password,
//       redirect: false,
//     })

//     return { status: 'success' }
//   } catch (error) {
//     if (error instanceof z.ZodError) {
//       return { status: 'invalid_data' }
//     }

//     return { status: 'failed' }
//   }
// }

export async function logout() {
  try {
    await deleteSession()
    return true
  } catch (error) {
    return false
  }
}
