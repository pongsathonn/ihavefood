'use client'

import { useRouter } from 'next/navigation'
import {
  ChevronLeft,
  Mail,
  Phone,
  User,
  Facebook,
  Instagram,
  MessageCircle,
  MapPin,
  Plus,
  LogOut,
} from 'lucide-react'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Separator } from '@/components/ui/separator'

import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
  DrawerTrigger,
} from '@/components/ui/drawer'

export default function ProfileDrawer() {
  const router = useRouter()
  return (
    <DrawerContent className="h-screen top-0 right-0 left-auto mt-0 max-w-none border-l bg-white data-[vaul-drawer-direction=right]:sm:max-w-lg">
      <DrawerHeader>
        <DrawerTitle>Account Settings</DrawerTitle>
        <DrawerDescription>
          Manage your profile and saved addresses
        </DrawerDescription>
      </DrawerHeader>
      <div className="no-scrollbar overflow-y-auto flex-1">
        <section className="w-full p-4 md:p-6 space-y-6">
          {/* Header Area - Following Track Dashboard Style */}
          <div className="flex items-center gap-4">
            {/* <Button
                            variant="outline"
                            size="icon"
                            onClick={() => router.back()}
                            className="rounded-full border-slate-200 hover:bg-slate-100"
                        >
                            <ChevronLeft className="h-5 w-5 text-slate-600" />
                        </Button> */}
          </div>

          {/* Profile Card */}
          <Card className="rounded-3xl border-none shadow-md overflow-hidden">
            <CardContent className="pt-8 pb-6">
              <div className="flex flex-col items-center">
                <Avatar className="h-24 w-24 border-4 border-slate-50 mb-4 shadow-sm">
                  <AvatarImage src="https://github.com/shadcn.png" />
                  <AvatarFallback className="bg-slate-200 text-slate-600 text-xl">
                    FC
                  </AvatarFallback>
                </Avatar>
                <h3 className="text-xl font-bold text-slate-900">
                  FooCustomer
                </h3>
                <p className="text-sm text-slate-500">Premium Member</p>
              </div>

              <div className="mt-8 space-y-6">
                {/* Info Section */}
                <div className="space-y-4">
                  <div className="flex justify-between items-center px-2">
                    <h4 className="font-bold text-slate-900">
                      Personal Information
                    </h4>
                    <Button
                      variant="link"
                      size="sm"
                      className="text-blue-600 h-auto p-0"
                    >
                      Edit
                    </Button>
                  </div>

                  <div className="grid gap-4 bg-slate-50/50 p-4 rounded-2xl border border-slate-100">
                    <div className="flex items-center gap-3">
                      <User className="h-4 w-4 text-slate-400" />
                      <span className="text-sm text-slate-500 w-24">
                        Customer ID
                      </span>
                      <span className="text-sm font-medium text-slate-900">
                        #TODO-12345
                      </span>
                    </div>
                    <div className="flex items-center gap-3">
                      <Mail className="h-4 w-4 text-slate-400" />
                      <span className="text-sm text-slate-500 w-24">Email</span>
                      <span className="text-sm font-medium text-slate-900">
                        TODO@example.comasodijasd
                      </span>
                    </div>
                    <div className="flex items-center gap-3">
                      <Phone className="h-4 w-4 text-slate-400" />
                      <span className="text-sm text-slate-500 w-24">Phone</span>
                      <Input
                        defaultValue="TODO"
                        className="h-9 bg-white border-slate-200 rounded-lg focus-visible:ring-slate-200"
                      />
                    </div>
                  </div>

                  <div className="flex justify-end gap-2 px-2">
                    <Button
                      variant="ghost"
                      size="sm"
                      className="rounded-xl text-slate-500"
                    >
                      Cancel
                    </Button>
                    <Button
                      size="sm"
                      className="rounded-xl bg-slate-900 hover:bg-slate-800 px-6"
                    >
                      Save
                    </Button>
                  </div>
                </div>

                <Separator className="bg-slate-100" />

                {/* Social Section */}
                <div className="space-y-4 px-2">
                  <h4 className="font-bold text-slate-900">Social</h4>
                  <div className="space-y-3">
                    <div className="flex items-center justify-between text-sm">
                      <div className="flex items-center gap-3">
                        <Facebook className="h-4 w-4 text-blue-600" />
                        <span className="text-slate-600">Facebook</span>
                      </div>
                      <span className="font-medium text-slate-900">TODO</span>
                    </div>
                    <div className="flex items-center justify-between text-sm">
                      <div className="flex items-center gap-3">
                        <Instagram className="h-4 w-4 text-pink-600" />
                        <span className="text-slate-600">Instagram</span>
                      </div>
                      <span className="font-medium text-slate-900">TODO</span>
                    </div>
                    <div className="flex items-center justify-between text-sm">
                      <div className="flex items-center gap-3">
                        <MessageCircle className="h-4 w-4 text-green-500" />
                        <span className="text-slate-600">LINE</span>
                      </div>
                      <span className="font-medium text-slate-900">TODO</span>
                    </div>
                  </div>
                </div>
              </div>
            </CardContent>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-lg font-bold">My Address</CardTitle>
              <Button
                variant="ghost"
                size="sm"
                className="text-blue-600 font-bold hover:bg-blue-50"
              >
                <Plus className="h-4 w-4 mr-1" /> Add New
              </Button>
            </CardHeader>
            <CardContent className="pb-6">
              <div className="p-8 border-2 border-dashed border-slate-100 rounded-2xl flex flex-col items-center justify-center text-slate-400">
                <MapPin className="h-8 w-8 mb-2 opacity-20" />
                <p className="text-xs">No addresses saved yet</p>
              </div>
            </CardContent>
          </Card>
        </section>
      </div>
      {/* <DrawerFooter>
                <Button>Submit</Button>
                <DrawerClose asChild>
                    <Button variant="outline">Cancel</Button>
                </DrawerClose>
            </DrawerFooter> */}
    </DrawerContent>
  )
}
