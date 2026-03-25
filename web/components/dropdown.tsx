'use client'

import { logout } from '@/app/(auth)/actions'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { AlignJustifyIcon, Braces, LogOutIcon } from 'lucide-react'
import { useTransition } from 'react'

export default function DropdownMenuIcons() {
  // useTransition
  // Provides a 'true/false' loading state to disable buttons or show spinners.
  // Safely handles the UI if the user navigates away before the action finishes.
  // Prevents 'race conditions' and bugs common with raw async event listeners.
  const [isPending, startTransition] = useTransition()

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          className="bg-transparent border-slate-600 text-slate-200 hover:bg-slate-800 hover:text-white cursor-pointer"
        >
          <AlignJustifyIcon />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        {/*<DropdownMenuItem className="cursor-pointer">
          <House />
          My Address
        </DropdownMenuItem>
        <DropdownMenuItem className="cursor-pointer">
          <CreditCard />
          Payment
        </DropdownMenuItem>*/}
        <DropdownMenuItem className="cursor-pointer">
          <Braces />
          OpenAPI
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem
          variant="destructive"
          className="cursor-pointer"
          onClick={(e) => {
            e.preventDefault()
            startTransition(async () => {
              await logout()
            })
          }}
        >
          <LogOutIcon />
          {isPending ? 'Logging out...' : 'Log Out'}
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
