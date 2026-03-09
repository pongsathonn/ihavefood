'use client'

import { register, RegisterActionState } from '@/app/(auth)/actions'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { useRouter } from 'next/navigation'
import { useActionState, useEffect } from 'react'

export function SignupForm({ ...props }: React.ComponentProps<typeof Card>) {
  const router = useRouter()

  const [state, signupAction] = useActionState<RegisterActionState, FormData>(
    register,
    { status: 'idle' },
  )

  useEffect(() => {
    if (state.status === 'user_exists') {
      console.error('Registration failed: User already exists.')
    } else if (state.status === 'invalid_data') {
      console.warn('Form validation failed.')
    } else if (state.status === 'failed') {
      console.error('Registration failed: Server error.')
    } else if (state.status === 'password_mismatch') {
      console.error('Registration failed: Password mismatch.')
    } else if (state.status === 'success') {
      console.log('Registration successful, redirecting...')
      router.push('/')
      router.refresh()
    }
  }, [state.status, router])

  return (
    <Card {...props}>
      <CardHeader>
        <CardTitle>Create an account</CardTitle>
        {/*<CardDescription>
          Enter your information below to create your account
        </CardDescription>*/}
      </CardHeader>
      <CardContent>
        <form action={signupAction}>
          <FieldGroup>
            <Field>
              <FieldLabel htmlFor="email">Email</FieldLabel>
              <Input
                id="email"
                type="email"
                autoComplete="email"
                placeholder="customer@example.com"
                required
              />
              {/*<FieldDescription>
                We&apos;ll use this to contact you. We will not share your email
                with anyone else.
              </FieldDescription>*/}
            </Field>
            <Field>
              <FieldLabel htmlFor="password">Password</FieldLabel>
              <Input
                id="password"
                name="passwordc"
                type="password"
                autoComplete="new-password"
                placeholder="&#9679;&#9679;&#9679;&#9679;&#9679;&#9679;&#9679;&#9679;&#9679;"
                required
              />
              {/*<FieldDescription>
                Must be at least 8 characters long.
              </FieldDescription>*/}
            </Field>
            <Field>
              <FieldLabel htmlFor="confirm-password">
                Confirm Password
              </FieldLabel>
              <Input
                id="confirm-password"
                name="confirmPassword"
                type="password"
                autoComplete="new-password"
                placeholder="&#9679;&#9679;&#9679;&#9679;&#9679;&#9679;&#9679;&#9679;&#9679;"
                required
              />
              {/*<FieldDescription>Please confirm your password.</FieldDescription>*/}
            </Field>
            <FieldGroup>
              <Field>
                <Button type="submit">Create Account</Button>
                {/*<Button variant="outline" type="button">
                  Sign up with Google
                </Button>*/}
                <FieldDescription className="px-6 text-center">
                  Already have an account? <a href="login">Sign in</a>
                </FieldDescription>
              </Field>
            </FieldGroup>
          </FieldGroup>
        </form>
      </CardContent>
    </Card>
  )
}
