'use client'

import { useRouter } from 'next/navigation'
import { LoginActionState, login } from '../actions'
import { SubmitButton } from '@/components/submit-button'
import { useActionState, useEffect, useState } from 'react'
import Form from "next/form";
import Link from "next/link";

export default function Page() {
    const router = useRouter()
    const [isSuccessful, setIsSuccessful] = useState(false);

    // login return state -> formAction(updater) -> state
    const [state, signinAction] = useActionState<LoginActionState, FormData>(
        login, { status: "idle", }
    );

    useEffect(() => {
        if (state.status === "failed") {
            console.error('Authentication failed: Invalid credentials.');
            // toast.error("Invalid username or password");
        } else if (state.status === "invalid_data") {
            console.warn('Form validation failed.');
            // toast.warn("Please check your input fields");
        } else if (state.status === "success") {
            console.log('Login successful, redirecting...');
            setIsSuccessful(true);
            router.refresh();
            router.push('/');
        }
    }, [state.status]);

    return (
        <section className="flex items-center justify-center mt-20 md:mt-10 p-10 bg-white rounded-3xl shadow-lg border border-gray-100 max-w-sm mx-auto" >
            <div className="w-full">
                <Form action={signinAction} className="space-y-4">
                    <h2 className="text-2xl font-bold text-gray-800 mb-6 text-center" > Sign In </h2>
                    < p className="text-sm text-gray-500 text-center" />
                    <div>
                        <label className="block text-sm font-medium text-gray-700" > Email </label>
                        < input name="email" type="email" autoComplete="email" required className="mt-1 block w-full px-3 py-2 bg-white border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500" />
                    </div>
                    < div >
                        <label className="block text-sm font-medium text-gray-700" > Password </label>
                        < input name="password" type="password" autoComplete="current-password" required className="mt-1 block w-full px-3 py-2 bg-white border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500" />
                    </div>
                    <SubmitButton isSuccessful={isSuccessful}>
                        Sign in
                    </SubmitButton>
                    <p className="text-sm mt-2 text-red-500 text-center" />
                </Form>

                < p className="text-sm mt-2 text-red-500 text-center" />
                <div className="mt-4 text-center text-sm" >
                    <p className="text-blue-500 hover:underline cursor-pointer" >
                        {"Don't have an account?"}
                        <Link href="/register">Sign Up</Link>
                    </p>
                </div>
            </div>
        </section >
    )
}

