"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useSession } from "next-auth/react";
import { useActionState, useEffect, useState } from "react";
import { SubmitButton } from "@/components/submit-button";
// import { toast } from "@/components/toast";
import { type RegisterActionState, register } from "../actions";
import Form from "next/form";

export default function Page() {
    const router = useRouter();

    const [email, setEmail] = useState("");
    const [isSuccessful, setIsSuccessful] = useState(false);

    const [state, formAction] = useActionState<RegisterActionState, FormData>(
        register,
        {
            status: "idle",
        }
    );

    const { update: updateSession } = useSession();

    // biome-ignore lint/correctness/useExhaustiveDependencies: router and updateSession are stable refs
    useEffect(() => {
        if (state.status === "user_exists") {
            // toast({ type: "error", description: "Account already exists!" });
        } else if (state.status === "failed") {
            // toast({ type: "error", description: "Failed to create account!" });
        } else if (state.status === "invalid_data") {
            // toast({
            //     type: "error",
            //     description: "Failed validating your submission!",
            // });
        } else if (state.status === "success") {
            // toast({ type: "success", description: "Account created successfully!" });

            setIsSuccessful(true);
            updateSession();
            router.refresh();
        }
    }, [state.status]);

    const handleSubmit = (formData: FormData) => {
        setEmail(formData.get("email") as string);
        formAction(formData);
    };

    return (

        // <AuthForm action={handleSubmit} defaultEmail={email}>
        //     <SubmitButton isSuccessful={isSuccessful}>Sign Up</SubmitButton>
        //     <p className="mt-4 text-center text-gray-600 text-sm dark:text-zinc-400">
        //         {"Already have an account? "}
        //             Sign in
        //         </Link>
        //         {" instead."}
        //     </p>
        // </AuthForm>

        <Form action={signupAction}>
            <h2 className="text-2xl font-bold text-gray-800 mb-6 text-center" >
                Sign Up
            </h2>
            < div >
                <label className="block text-sm font-medium text-gray-700" > Email </label>
                < input name="email" type="email" autoComplete="email" className="mt-1 block w-full px-3 py-2 bg-white border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500" />
                <p className="text-sm mt-2 text-red-500 w-full text-left" />
            </div>
            < div >
                <label className="block text-sm font-medium text-gray-700" > Password </label>
                < input name="password" type="password" autoComplete="new-password" className="mt-1 block w-full px-3 py-2 bg-white border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500" />
                <p className="text-sm mt-2 text-red-500 text-center" />
            </div>
            < div >
                <label className="block text-sm font-medium text-gray-700" > Confirm Password </label>
                < input name="confirm-password" type="password" autoComplete="new-password" className="mt-1 block w-full px-3 py-2 bg-white border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500" />
                <p className="text-sm mt-2 text-red-500 text-center" />
            </div>
            < div >
                <label className="block text-sm font-medium text-gray-700" > I am a...  </label>
                < select name="role" className="mt-1 block w-full px-3 py-2 bg-white border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500" >
                    <option value="customer" > Customer </option>
                    < option value="rider" > Rider </option>
                </select>
            </div>
            <SubmitButton isSuccessful={TODO}>
                Sign Up
            </SubmitButton>

            < p className="text-blue-500 hover:underline cursor-pointer hidden" >
                Already have an account ?
                <Link> </Link>
            </p>
        </Form>



        //         <Link
        //             className="font-semibold text-gray-800 hover:underline dark:text-zinc-200"
        //             href="/login"
        //         >


    );
}