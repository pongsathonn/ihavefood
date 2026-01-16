'use client'

import { FormEvent } from 'react'
import { useRouter } from 'next/navigation'
import { SigninButton, SignupButton, } from '@/components/auth-button'

export default function LoginPage() {
    const router = useRouter()

    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault()

        const formData = new FormData(event.currentTarget)
        const serverUrl = process.env.NEXT_PUBLIC_SERVER_URL
        const response = await fetch(`${serverUrl}/auth/login`, {
            method: "POST",
            headers: {
                "Accept": "application/json",
                "Content-Type": "application/json"
            },
            body: JSON.stringify({
                identifier: formData.get('email'),
                password: formData.get('password'),
                role: "CUSTOMER",
            })
        });

        if (!response.ok) {
            const errorData = await response.json();
            const err = new Error(errorData.message);
            // err.status = response.status;
            throw err;
        }

        console.log(await response.json());

        router.push('/')
        // return await response.json();
    }

    return (
        <section className="flex items-center justify-center p-6 bg-white rounded-3xl shadow-lg border border-gray-100 max-w-sm mx-auto">
            <div className="w-full">

                <form onSubmit={handleSubmit} className="space-y-4">

                    <h2 className="text-2xl font-bold text-gray-800 mb-6 text-center" > Sign In </h2>
                    <p className="text-sm text-gray-500 text-center" />
                    <div>
                        <label className="block text-sm font-medium text-gray-700" > Email </label>
                        <input name="email" type="email" autoComplete="email" required className="mt-1 block w-full px-3 py-2 bg-white border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500" />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700" > Password </label>
                        <input name="password" type="password" autoComplete="current-password" required className="mt-1 block w-full px-3 py-2 bg-white border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500" />
                    </div>

                    <SigninButton />


                    <p className="text-sm mt-2 text-red-500 text-center" />
                </form>

                <form className="space-y-4 hidden">
                    <h2 className="text-2xl font-bold text-gray-800 mb-6 text-center" >
                        Sign Up
                    </h2>
                    <div>
                        <label className="block text-sm font-medium text-gray-700" >
                            Email
                        </label>
                        <input
                            type="email"
                            autoComplete="email"
                            className="mt-1 block w-full px-3 py-2 bg-white border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                        />
                        <p className="text-sm mt-2 text-red-500 w-full text-left" />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700" >
                            Password
                        </label>
                        <input
                            type="password"
                            autoComplete="new-password"
                            className="mt-1 block w-full px-3 py-2 bg-white border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                        />
                        <p className="text-sm mt-2 text-red-500 text-center" />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700" >
                            Confirm Password
                        </label>
                        <input
                            type="password"
                            autoComplete="new-password"
                            className="mt-1 block w-full px-3 py-2 bg-white border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                        />
                        <p className="text-sm mt-2 text-red-500 text-center" />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700" >
                            I am a...
                        </label>
                        <select className="mt-1 block w-full px-3 py-2 bg-white border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500" >
                            <option value="customer">Customer</option>
                            <option value="rider">Rider</option>
                        </select>
                    </div>
                    <SignupButton />
                </form>
                <p className="text-sm mt-2 text-red-500 text-center" />
                <div className="mt-4 text-center text-sm">
                    <span className="text-blue-500 hover:underline cursor-pointer" >
                        {"Don't have an account? Sign Up"}
                    </span>
                    <span className="text-blue-500 hover:underline cursor-pointer hidden" >
                        Already have an account? Sign In
                    </span>
                </div>
            </div>
        </section >
    )
}
