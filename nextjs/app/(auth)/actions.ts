'use server'

import { z } from "zod";
import { cookies } from 'next/headers'
import { createSession, deleteSession } from '@/app/lib/session';


// @source https://github.com/vercel/ai-chatbot

const authFormSchema = z.object({
    email: z.email(),
    password: z.string().min(6),
});

export type LoginActionState = {
    status: "idle" | "in_progress" | "success" | "failed" | "invalid_data";
};

export const login = async (
    _: LoginActionState,
    formData: FormData
): Promise<LoginActionState> => {

    try {
        const serverUrl = process.env.SERVER_URL
        const validatedData = authFormSchema.parse({
            email: formData.get("email"),
            password: formData.get("password"),
        });

        const response = await fetch(`${serverUrl}/auth/login`, {
            method: "POST",
            headers: {
                "Accept": "application/json",
                "Content-Type": "application/json"
            },
            body: JSON.stringify({
                identifier: validatedData.email,
                password: validatedData.password,
            })
        });

        const token = await response.json();
        const unixTimestamp = parseInt(token.expiresIn);
        const expirationDate = new Date(unixTimestamp * 1000);
        await createSession(token.accessToken, expirationDate);
        return { status: "success" };

    } catch (error) {
        console.log("err", error);
        if (error instanceof z.ZodError) {
            return { status: "invalid_data" };
        }
        return { status: "failed" };
    }
}

export type RegisterActionState = {
    status:
    | "idle"
    | "in_progress"
    | "success"
    | "failed"
    | "user_exists"
    | "invalid_data";
};

export const register = async (
    _: RegisterActionState,
    formData: FormData
): Promise<RegisterActionState> => {
    try {
        const validatedData = authFormSchema.parse({
            email: formData.get("email"),
            password: formData.get("password"),
        });

        const [user] = await getUser(validatedData.email);
        // fetch getauthbyidentifier

        if (user) {
            return { status: "user_exists" } as RegisterActionState;
        }
        await createUser(validatedData.email, validatedData.password);
        await signIn("credentials", {
            email: validatedData.email,
            password: validatedData.password,
            redirect: false,
        });

        return { status: "success" };
    } catch (error) {
        if (error instanceof z.ZodError) {
            return { status: "invalid_data" };
        }

        return { status: "failed" };
    }
};

export async function logout() {
    try {
        await deleteSession();
        return true;
    } catch (error) {
        return false;
    }
}