import { SearchIcon } from "lucide-react"
import {
    InputGroup,
    InputGroupAddon,
    InputGroupButton,
    InputGroupInput,
} from "@/components/ui/input-group"

const SearchBar = () => (
    <InputGroup className="w-full max-w-sm bg-background">
        <InputGroupInput placeholder="ค้นหาร้านอาหาร..." />
        <InputGroupAddon align="inline-end" className="pr-2">
            <InputGroupButton size="sm" variant="secondary">
                <SearchIcon />
            </InputGroupButton>
        </InputGroupAddon>
    </InputGroup>
)

export default SearchBar