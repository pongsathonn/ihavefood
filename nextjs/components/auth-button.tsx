"use client"

export function SigninButton() {
    return (
        <button type="submit" className="w-full bg-blue-500 hover:bg-blue-600 text-white font-bold py-2 px-4 rounded-full transition-colors duration-200" >
            Sign In
        </button>
    )
}

export function SignupButton() {
    return (
        <button type="submit" className="w-full bg-pink-500 hover:bg-pink-600 text-white font-bold py-2 px-4 rounded-full transition-colors duration-200" >
            Sign Up
        </button>
    )
}
