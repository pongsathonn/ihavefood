'use client'

import { useRouter } from 'next/navigation'
import SearchBar from './search-bar'
import { Button } from './ui/button'
import { Tooltip, TooltipContent, TooltipTrigger } from './ui/tooltip'
import Scooter02Icon from './scooter-02'
import { Separator } from './ui/separator'
import { Drawer, DrawerTrigger } from './ui/drawer'
import { UserIcon } from 'lucide-react'
import ProfileDrawer from './profile-drawer'
import DropdownMenuIcons from './dropdown'
import Link from 'next/link'

export default function Header() {
  const router = useRouter()
  return (
    <header
      id="main-header"
      className="grid grid-cols-4 items-center w-full h-[8vh] px-6 bg-slate-950"
    >
      <div className="col-span-1 flex justify-start">
        <Link href={'/'}>
          <h1 className="text-lg text-white drop-shadow-lg font-extrabold tracking-tight shrink-0">
            IHAVE<span className="text-amber-600">FOOD</span>
          </h1>
        </Link>
      </div>

      <div className="col-span-2 hidden md:flex justify-center items-center">
        <SearchBar />
      </div>

      <div className="col-span-1 hidden md:flex justify-end items-center">
        <div className="flex h-5 items-center gap-4 text-sm">
          <div>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant="ghost"
                  className="bg-transparent border-slate-600 text-slate-200 hover:bg-slate-800 hover:text-white cursor-pointer"
                  onClick={() => {
                    router.push('/dashboard/tracking')
                  }}
                >
                  {/* show badge */}

                  {/* TODO: add tooltip */}
                  <Scooter02Icon />
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                <p>Order Tracking</p>
              </TooltipContent>
            </Tooltip>
          </div>
          <Separator orientation="vertical" />
          <div>
            <Drawer direction="right">
              <DrawerTrigger asChild>
                <Button
                  variant="ghost"
                  className="bg-transparent border-slate-600 text-slate-200 hover:bg-slate-800 hover:text-white cursor-pointer"
                >
                  <UserIcon />
                </Button>
              </DrawerTrigger>
              <ProfileDrawer />
            </Drawer>
          </div>
          <Separator orientation="vertical" />
          <div>
            <DropdownMenuIcons />
          </div>
        </div>
      </div>
    </header>
  )
}
