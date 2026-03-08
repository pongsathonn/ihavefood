import { SearchIcon } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
  InputGroupInput,
} from '@/components/ui/input-group'
import {
  Command,
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command'
import { useState } from 'react'

export default function SearchBar() {
  const [open, setOpen] = useState(false)

  return (
    <div className="flex justify-center gap-4 w-full">
      <InputGroup className="w-full max-w-lg bg-background">
        <InputGroupInput
          placeholder="ค้นหาร้านอาหาร..."
          onClick={() => setOpen(true)}
        />
        <InputGroupAddon align="inline-end" className="pr-2 ">
          <InputGroupButton
            size="sm"
            variant="secondary"
            className="cursor-pointer"
          >
            <SearchIcon />
          </InputGroupButton>
        </InputGroupAddon>
      </InputGroup>
      <CommandDialog open={open} onOpenChange={setOpen}>
        <Command>
          <CommandInput placeholder="พิมพ์ชื่อร้านอาหารที่ต้องการค้นหา" />
          <CommandList>
            <CommandEmpty>ไม่พบร้านอาหารที่ต้องการ</CommandEmpty>
            <CommandGroup heading="ร้านแนะนำ">
              <CommandItem>อร่อยดีตามสั่ง</CommandItem>
              <CommandItem>กุ้งๆๆๆ</CommandItem>
            </CommandGroup>
          </CommandList>
        </Command>
      </CommandDialog>
    </div>
  )
}
