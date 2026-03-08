import LoginForm from '@/components/login-form'
import { Suspense } from 'react'
import { GalleryVerticalEnd } from "lucide-react"
// import LoginForm from '@/app/ui/fake-login-form'

export default async function Page() {
    return (
        <div className="bg-muted flex min-h-svh flex-col items-center justify-center gap-6 p-6 md:p-10">
            <div className="flex w-full max-w-sm flex-col gap-6">
                <Suspense>
                    <LoginForm />
                </Suspense>
            </div>
        </div>
    )
}