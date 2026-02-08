"use client";

import { useFormStatus } from "react-dom";

interface ButtonProps {
    children: React.ReactNode
    isSuccessful: boolean
    variant?: 'primary' | 'secondary' | 'danger'; // Defined styles
}

// <SubmitButton isSuccessful={TODO}>Sign In<SubmitButton>
export function SubmitButton({ children, isSuccessful, variant = 'primary' }: ButtonProps) {
    const { pending } = useFormStatus();

    const baseStyles = "w-full rounded-xl px-4 py-3 font-semibold transition-all flex items-center justify-center gap-2 disabled:opacity-50";

    const variants = {
        primary: "bg-blue-600 text-white hover:bg-blue-700",
        secondary: "bg-gray-100 text-gray-900 hover:bg-gray-200",
        danger: "bg-red-600 text-white hover:bg-red-700",
    };

    return (
        <button aria-disabled={pending || isSuccessful}
            className={`${baseStyles} ${variants[variant]}`}
            disabled={pending || isSuccessful}
            type={pending ? "button" : "submit"}
        >
            {children}

            {(pending || isSuccessful) && (
                <span className="absolute right-4 animate-spin">
                    Loading...
                    {/* <LoaderIcon /> */}
                </span>
            )}

            <output aria-live="polite" className="sr-only">
                {pending || isSuccessful ? "Loading" : "Submit form"}
            </output>
        </button>
    );
}