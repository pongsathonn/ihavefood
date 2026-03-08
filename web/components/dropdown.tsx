import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  AlignJustifyIcon,
  Braces,
  CreditCard,
  House,
  LogOutIcon,
} from 'lucide-react'

export default function DropdownMenuIcons() {
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
        <DropdownMenuItem className="cursor-pointer">
          <House />
          My Address
        </DropdownMenuItem>
        <DropdownMenuItem className="cursor-pointer">
          <CreditCard />
          Payment
        </DropdownMenuItem>
        <DropdownMenuItem className="cursor-pointer">
          <Braces />
          OpenAPI
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem variant="destructive" className="cursor-pointer">
          <LogOutIcon />
          Log out
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
