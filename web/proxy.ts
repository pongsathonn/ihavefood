import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'
import { authentication } from './lib/session'

export default async function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl
  const { isAuth } = await authentication()

  const isAuthPage = pathname === '/login' || pathname === '/register'
  if (isAuth && isAuthPage) {
    return NextResponse.redirect(new URL('/restaurants', request.url))
  }

  const isProtected =
    pathname.startsWith('/restaurants') || pathname.startsWith('/dashboard')

  if (isProtected && !isAuth) {
    return NextResponse.redirect(new URL('/login', request.url))
  }

  if (pathname === '/' && isAuth) {
    return NextResponse.redirect(new URL('/restaurants', request.url))
  }

  return NextResponse.next()
}
