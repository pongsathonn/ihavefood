import { Button } from "@/components/ui/button"
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import {
    House,
    LogOutIcon,
    UserIcon,
    CreditCard,
    AlignJustifyIcon,
    Braces,
} from "lucide-react"

export default function DropdownMenuIcons() {
    return (
        <DropdownMenu>
            <DropdownMenuTrigger asChild>
                <Button variant="ghost" className="bg-transparent border-slate-600 text-slate-200 hover:bg-slate-800 hover:text-white" >
                    <AlignJustifyIcon />
                </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
                <DropdownMenuItem>
                    {/* 
                    <div className="pr-16 justify-center">
                        <IconButton
                            icon="profile"
                            showBadge={false}
                            onClick={() => {
                                setShowProfileMenu(true);
                            }}
                        >
                            Profile
                        </IconButton>
                        {showProfileMenu && <ProfileMenu />}
                    </div> */}
                    <UserIcon />
                    Profile
                </DropdownMenuItem>
                <DropdownMenuItem>
                    <House />
                    My Address
                </DropdownMenuItem>
                <DropdownMenuItem>
                    <CreditCard />
                    Payment
                </DropdownMenuItem>
                <DropdownMenuItem>
                    <Braces />
                    OpenAPI
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem variant="destructive">
                    <LogOutIcon />
                    Log out
                </DropdownMenuItem>
            </DropdownMenuContent>
        </DropdownMenu >
    )
}
