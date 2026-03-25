'use server'

import { createSession, deleteSession } from '@/lib/session'
import { cookies } from 'next/headers'
import { z } from 'zod'

// @source https://github.com/vercel/ai-chatbot

const authFormSchema = z.object({
  email: z.email(),
  password: z.string().min(6),
})

const registerFormSchema = z.object({
  email: z.email(),
  password: z.string().min(6),
  confirmPassword: z.string().min(6),
  role: z.string().startsWith('ROLES'),
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
  if (!token.accessToken || !token.expiresAt) {
    throw new Error('Malformed auth response from server')
  }
  if (!token.accessToken || !token.expiresAt) {
    throw new Error('Malformed auth response from server')
  }

  const expiresAt = new Date(token.expiresAt)
  await createSession(token.accessToken, expiresAt)
}

export type RegisterActionState = {
  status:
    | 'idle'
    | 'in_progress'
    | 'success'
    | 'failed'
    | 'password_mismatch'
    | 'user_exists'
    | 'invalid_data'
}

export const register = async (
  _: RegisterActionState,
  formData: FormData,
): Promise<RegisterActionState> => {
  try {
    const email = formData.get('email') as string
    const password = formData.get('password') as string
    const confirmPassword = formData.get('confirm-password') as string
    const role = 'ROLES_CUSTOMER'

    if (password !== confirmPassword) return { status: 'password_mismatch' }
    const validatedData = registerFormSchema.parse({
      email,
      password,
      confirmPassword,
      role,
    })

    await signUp(
      validatedData.email,
      validatedData.password,
      validatedData.role,
    )

    return { status: 'success' }
  } catch (error) {
    if (error instanceof z.ZodError) {
      console.error(error)
      return { status: 'invalid_data' }
    }

    return { status: 'failed' }
  }
}

const signUp = async function (
  email: string,
  password: string,
  role: string,
): Promise<void> {
  const serverUrl = process.env.SERVER_URL
  if (!serverUrl) throw new Error('SERVER_URL is not defined')

  const response = await fetch(`${serverUrl}/auth/register`, {
    method: 'POST',
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ email, password, role }),
  })

  if (!response.ok) {
    const error = await response.json().catch(() => ({}))
    throw new Error(error.message ?? `Register failed: ${response.status}`)
  }

  // await response.json()
}

export async function logout() {
  try {
    await deleteSession()

    const cookieStore = await cookies()
    cookieStore.delete('default_address_id')

    return true
  } catch (error) {
    console.log(error)
    return false
  }
}
