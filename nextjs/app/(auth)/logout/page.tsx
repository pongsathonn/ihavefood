'use client';

import { useActionState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { logout } from '../actions';
import { SubmitButton } from '@/components/submit-button';

export default function Page() {
    const router = useRouter();

    const [state, signoutAction, isPending] = useActionState(logout, false);

    useEffect(() => {
        if (state === true) {
            router.push('/login');
            router.refresh();
        }
    }, [state, router]);

    return (
        <form action={signoutAction}>
            <SubmitButton variant="secondary" isSuccessful={isPending}>
                {isPending ? "Signing Out..." : "Sign Out"}
            </SubmitButton>

            {state && <div className="text-green-500">Successfully signed out</div>}
        </form>
    );
}