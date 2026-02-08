import { redirect } from 'next/navigation';
import { verifySession } from '@/app/lib/session';
export default async function Page() {

    const res = await verifySession()
    if (!res.isAuth) {
        redirect('/login');
    }

    return (
        <>
            <div >
                HI FROM HOME
            </div>
        </>
    )
}

