import 'server-only';
import { cookies } from 'next/headers';
import { SignJWT, jwtVerify } from 'jose';
// import { SessionPayload, Role } from '@/app/lib/definitions';
import { redirect } from 'next/navigation'
import { cache } from 'react';


// https://nextjs.org/docs/app/guides/authentication

const { SecretManagerServiceClient } = require('@google-cloud/secret-manager');
const client = new SecretManagerServiceClient();
let cachedKey: string | undefined = undefined;

export async function createSession(token: string, expiresAt: Date) {
    const cookieStore = await cookies()
    cookieStore.set('session', token, {
        httpOnly: true,
        secure: true,
        expires: expiresAt,
        sameSite: 'lax',
        path: '/',
    })
}

export async function updateSession() {
    const session = (await cookies()).get('session')?.value
    const payload = await decrypt(session)

    if (!session || !payload) {
        return null
    }

    const expires = new Date(Date.now() + 7 * 24 * 60 * 60 * 1000)

    const cookieStore = await cookies()
    cookieStore.set('session', session, {
        httpOnly: true,
        secure: true,
        expires: expires,
        sameSite: 'lax',
        path: '/',
    })
}

export async function deleteSession() {
    const cookieStore = await cookies()
    cookieStore.delete('session')
}

// change logic follow Next.js document
export const verifySession = cache(async () => {
    const session = (await cookies()).get('session')?.value

    console.log('verify session', session);


    const payload = await decrypt(session);
    const userId = payload?.sub;
    const role = payload?.role;

    console.log("verify session after", payload, userId, role);
    if (!userId) {
        redirect('/login')
    }

    return { isAuth: true, userId }
})


export async function getSecretKey() {
    if (cachedKey) return cachedKey;
    try {
        const [accessResponse] = await client.accessSecretVersion({
            name: `projects/${process.env.GCP_PROJECT_ID}/secrets/JWT_SIGNING_KEY/versions/latest`,
        });

        cachedKey = accessResponse.payload.data.toString('utf8');
        return cachedKey;
    } catch (error) {
        console.error("Failed to fetch secret:", error);
        throw new Error("Could not retrieve signing key");
    }
}

// export async function encrypt(payload: SessionPayload) {
//     return new SignJWT(payload)
//         .setProtectedHeader({ alg: 'HS256' })
//         .setIssuedAt()
//         .setExpirationTime('7d')
//         .sign(encodedKey)
// }

export async function decrypt(session: string | undefined = '') {
    const encodedKey = new TextEncoder().encode(await getSecretKey())
    try {
        const { payload } = await jwtVerify(session, encodedKey, {
            algorithms: ['HS256'],
        })
        return payload
    } catch (error) {
        console.log('Failed to verify session')
    }
}