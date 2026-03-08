import { authentication } from '@/lib/session'
import { redirect } from 'next/navigation'

export default async function Page() {
  const { isAuth } = await authentication()
  if (!isAuth) {
    redirect('/login')
  }

  redirect('/restaurants')
}
