'use client'

import { authenticate, LoginActionState } from '@/app/(auth)/actions'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'
import { useRouter } from 'next/navigation'
import { useActionState, useEffect } from 'react'

export default function SignInForm({
  className,
  ...props
}: React.ComponentProps<'div'>) {
  const router = useRouter()

  // authenticate return state -> formAction(updater) -> state
  const [state, signinAction] = useActionState<LoginActionState, FormData>(
    authenticate,
    { status: 'idle' },
  )

  useEffect(() => {
    if (state.status === 'failed') {
      console.error('Authentication failed: Invalid credentials.')
      // toast.error("Invalid username or password");
    } else if (state.status === 'invalid_data') {
      console.warn('Form validation failed.')
      // toast.warn("Please check your input fields");
    } else if (state.status === 'success') {
      console.log('Login successful, redirecting...')
      router.push('/')
      router.refresh()
    }
  }, [state.status, router])

  return (
    <div className={cn('flex flex-col gap-6', className)} {...props}>
      <Card>
        <CardHeader className="text-center">
          <CardTitle className="text-xl">Welcome</CardTitle>
        </CardHeader>
        <CardContent>
          <form action={signinAction}>
            <FieldGroup>
              <Field>
                <FieldLabel htmlFor="email">Email</FieldLabel>
                <Input
                  name="email"
                  type="email"
                  autoComplete="username"
                  placeholder="customer@example.com"
                  required
                />
              </Field>
              <Field>
                <div className="flex items-center">
                  <FieldLabel htmlFor="password">Password</FieldLabel>
                  <a
                    href="#TODO-impl-forgot-password"
                    className="ml-auto text-sm underline-offset-4 hover:underline"
                  >
                    Forgot your password?
                  </a>
                </div>
                <Input
                  name="password"
                  type="password"
                  autoComplete="current-password"
                  placeholder="&#9679;&#9679;&#9679;&#9679;&#9679;&#9679;&#9679;&#9679;&#9679;"
                  required
                />
              </Field>
              <Field>
                <Button type="submit">Login</Button>
                <FieldDescription className="text-center">
                  Don&apos;t have an account? <a href="register">Sign up</a>
                </FieldDescription>
              </Field>
            </FieldGroup>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}

// signupForm.addEventListener('submit', async (e) => {

//     e.preventDefault();

//     emailRegisterMsg.textContent = "";
//     passRegisterMsg.textContent = "";
//     passConfirmMsg.textContent = "";

//     const auth = signupEmailInput.value;
//     const pass = signupPasswordInput.value;
//     const passConfirm = signupPasswordConfirmInput.value;
//     const roleInput = signupRoleSelect.value;

//     if (pass != passConfirm) {
//         passConfirmMsg.textContent = 'Passwords do not match';
//         passConfirmMsg.className = `text-sm mt-2 text-center text-red-500`;
//         return;
//     }

//     let role = "";
//     if (roleInput === "customer") {
//         role = "ROLES_CUSTOMER"
//     } else if (roleInput === "rider") {
//         role = "ROLES_RIDER"
//     }

//     try {
//         await register(auth, pass, role);
//         showModalRegister();

//         signupForm.classList.add('hidden');
//         signinForm.classList.remove('hidden');
//         showSigninLink.classList.add('hidden');
//         showSignupLink.classList.remove('hidden');
//         authStatus.textContent = '';

//     } catch (error) {

//         const cleanMsg = error.message.split(/:(.+)/)[1]?.trim() || "";
//         const errors = cleanMsg.match(/(Email[^,]*|Password.*)/gi) || [];

//         // console.error(error.stack);

//         if (error.status == 409) {
//             emailRegisterMsg.textContent = "email already exists";
//             emailRegisterMsg.className = "text-sm mt-2 text-center text-red-500";
//             return;
//         }

//         if (errors.length === 0) {
//             registerMsg.textContent = "Signup failed. Please try again.";
//             registerMsg.className = "text-sm mt-2 text-center text-red-500";
//             return;
//         }

//         errors.forEach(err => {
//             if (err.toLowerCase().startsWith("email")) {
//                 emailRegisterMsg.textContent = err;
//                 emailRegisterMsg.className = `text-sm mt-2 text-center text-red-500`;
//             } else if (err.toLowerCase().startsWith("password")) {
//                 passRegisterMsg.textContent = err;
//                 passRegisterMsg.className = `text-sm mt-2 text-center text-red-500`;
//             }
//         });

//         return;
//     };
// });

///////////////////////////////////////////////////////

// signinForm.addEventListener('submit', async (e) => {

//     e.preventDefault();

//     const auth = signinEmailInput.value;
//     const pass = signinPasswordInput.value;

//     try {
//         const loginResponse = await login(auth, pass);
//         const accessToken = loginResponse.accessToken;
//         sessionStorage.setItem("token", accessToken);
//         const payload = decodeJwt(accessToken);
//         const customerId = payload.sub;

//         if (!payload || !customerId) {
//             console.error('JWT payload is missing the required ID field.');
//             return;
//         }

//         const customer = await fetchCustomer(customerId);
//         sessionStorage.setItem("customer_id", customer.customerId);

//         // sessionStorage.setItem("customer_addresses", JSON.stringify(customer.addresses));

//         addresses = customer.addresses || [];
//         if (addresses.length > 0) addresses[0].isDefault = true;

//         isUserAuthenticated = true;

//         profileButton.classList.remove('hidden');
//         trackButton.classList.remove('hidden');
//         apiButton.classList.remove('hidden');
//         showSection('user-profile');
//         renderAddresses();
//         renderRestaurants();
//         showSection('restaurant-list');

//     } catch (error) {
//         if (error.status === 401) {
//             authStatus.textContent = "Login or password is invalid.";
//             authStatus.className = `text-sm mt-2 text-center text-red-500`;
//             return;
//         }
//         // console.error(error.stack);
//         authStatus.textContent = "Login failed. Please try again.";
//         authStatus.className = `text-sm mt-2 text-center text-red-500`;
//         return;
//     }
// });
