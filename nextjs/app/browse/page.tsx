'use client'

import { useState } from "react";
import { Restaurant } from "@/app/lib/definitions";
import { Separator } from "@/components/ui/separator"
import { Button } from "@/components/ui/button"
import DropdownMenuIcons from "@/components/dropdown";
import FoodCategory from "@/components/food-card";
import PromotionCard from "@/components/promotion-card";
import SearchBar from "@/components/search-bar";
import { useRouter } from "next/navigation";

const models = [
    {
        name: "v0-1.5-sm",
        description: "Everyday tasks and UI generation.",
        // image:
        //     "https://images.unsplash.com/photo-1650804068570-7fb2e3dbf888?q=80&w=640&auto=format&fit=crop",
        credit: "Valeria Reverdo on Unsplash",
    },
    {
        name: "v0-1.5-lg",
        description: "Advanced thinking or reasoning.",
        // image:
        //     "https://images.unsplash.com/photo-1610280777472-54133d004c8c?q=80&w=640&auto=format&fit=crop",
        credit: "Michael Oeser on Unsplash",
    },
    {
        name: "v0-2.0-mini",
        description: "Open Source model for everyone.",
        // image:
        //     "https://images.unsplash.com/photo-1602146057681-08560aee8cde?q=80&w=640&auto=format&fit=crop",
        credit: "Cherry Laithang on Unsplash",
    },
]

export default function Page() {
    const [isOrder, setIsOrder] = useState(false);
    const router = useRouter();
    const [currentRestaurant, setCurrentRestaurant] = useState<Restaurant | undefined>();


    const [showProfileMenu, setShowProfileMenu] = useState(false);
    const [showOrderTracking, setShowOrderTracking] = useState(false);

    // currentPage?
    // const pageStatue = 'BrowsePage' | 'ProfileMenu' | 'OrderTracking';
    // useState<>()

    // // useActionState ?
    // const [deliveryStatus, setDeliveryStatus] = useState();

    // const restaurants = await fetchRestaurant();

    //     if (addresses.length === 0) {
    //         restaurantsContainer.innerHTML = `< div class= "col-span-full flex items-center justify-center min-h-[300px] text-gray-400 text-2xl" >
    //                     Please update your delivery address to see nearby restaurants.
    //                 </div>`;
    //         return;
    //     }

    // TODO: get current restaurant from context?

    const handleTrack = () => {
        setIsOrder(true);
        setShowOrderTracking(true);
        router.push("/dashboard/tracking")
    };

    return (
        <div className="h-[80vh] bg-gradient-to-r from-slate-900 to-slate-700 rounded-b-3xl">
            <header id="main-header" className="grid grid-cols-4 items-center w-full h-[8vh] px-6 bg-slate-950">
                <div className="col-span-1 flex justify-start">
                    <h1 className="text-lg text-white drop-shadow-lg font-extrabold tracking-tight shrink-0">
                        IHAVE<span className="text-amber-600">FOOD</span>
                    </h1>
                </div>
                <div className="col-span-2 hidden md:flex justify-center items-center">
                    <SearchBar />
                </div>
                <div className="col-span-1 hidden md:flex justify-end items-center">

                    <div className="flex h-5 items-center gap-4 text-sm">
                        <div >
                            <Button
                                variant="ghost"
                                className="bg-transparent border-slate-600 text-slate-200 hover:bg-slate-800 hover:text-white"
                                onClick={handleTrack}
                            >
                                {/* show badge */}
                                Track
                            </Button>
                        </div>
                        <Separator orientation="vertical" />
                        <div>
                            <DropdownMenuIcons />
                        </div>
                    </div>


                </div>
            </header >

            <div className="flex justify-center p-4 overflow-x-hidden">
                <PromotionCard />
            </div>

            < div className="sm:pl-10 sm:pr-10" >
                <FoodCategory
                    restaurants={restaurants}
                    setCurrentRestaurant={setCurrentRestaurant}
                    currentRestaurant={currentRestaurant}
                />
            </div >
        </div >
    )
}


export const restaurants: Restaurant[] = [
    {
        restaurantId: "550e8400-e29b-41d4-a716-446655440000",
        restaurantName: "อร่อยดีตามสั่ง",
        imageInfo: { url: "https://images.unsplash.com/photo-1555939594-58d7cb561ad1", type: "image/jpg" },
        status: "STORE_STATUS_OPEN",
        menu: [
            { itemId: "7216694e-7f61-460b-8081-306915f4a47d", foodName: "ข้าวผัดหมู", price: 50, imageInfo: { url: "https://images.unsplash.com/photo-1603133872878-684f208fb84b", type: "image/jpg" } },
            { itemId: "8472937e-6f82-491c-b192-417026f5b58e", foodName: "กระเพราหมูไข่ดาว", price: 50, imageInfo: { url: "https://images.unsplash.com/photo-1512058560366-cd2427ff06b3", type: "image/jpg" } },
            { itemId: "c928456a-1d2e-4f3b-9a8c-528137f6d79a", foodName: "สุกี้แห้งทะเล", price: 60, imageInfo: { url: "https://images.unsplash.com/photo-1562607311-20921703cc2c", type: "image/jpg" } }
        ],
        address: {
            addressId: "f47ac10b-58cc-4372-a567-0e02b2c3d479",
            addressName: "123 ถนนนิมมานเหมินท์",
            subDistrict: "สุเทพ",
            district: "เมืองเชียงใหม่",
            province: "เชียงใหม่",
            postalCode: "50200"
        },
        phone: "0981234567",
        email: "aoroydee@example.com"
    },
    {
        restaurantId: "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
        restaurantName: "ป้าแดง ก๋วยเตี๋ยว",
        imageInfo: { url: "https://images.unsplash.com/photo-1582878826629-29b7ad1cdc43", type: "image/jpg" },
        status: "STORE_STATUS_OPEN",
        menu: [
            { itemId: "10928374-abcd-4321-bcde-90817263544d", foodName: "ก๋วยเตี๋ยวหมูน้ำใส", price: 40, imageInfo: { url: "https://images.unsplash.com/photo-1617093727343-374698b1b08d", type: "image/jpg" } },
            { itemId: "21039485-bcde-5432-cdef-01928374655e", foodName: "ก๋วยเตี๋ยวต้มยำ", price: 45, imageInfo: { url: "https://images.unsplash.com/photo-1552611052-33e04de081de", type: "image/jpg" } },
            { itemId: "32140596-cdef-6543-defa-12039485766f", foodName: "เส้นเล็กน้ำตก", price: 45, imageInfo: { url: "https://images.unsplash.com/photo-1569718212165-3a8278d5f624", type: "image/jpg" } }
        ],
        address: {
            addressId: "a1b2c3d4-e5f6-4a5b-bc6d-7e8f9a0b1c2d",
            addressName: "45 ถนนราชภาคินัย",
            subDistrict: "ช้างคลาน",
            district: "เมืองเชียงใหม่",
            province: "เชียงใหม่",
            postalCode: "50100"
        },
        phone: "0982345678",
        email: "padang_noodle@example.com"
    },
    {
        restaurantId: "78901234-5678-4321-8765-432109876543",
        restaurantName: "ไก่จอมพลัง",
        imageInfo: { url: "https://images.unsplash.com/photo-1626082927389-6cd097cdc6ec", type: "image/jpg" },
        status: "STORE_STATUS_OPEN",
        menu: [
            { itemId: "56789012-3456-4789-9012-345678901234", foodName: "ข้าวมันไก่ต้ม", price: 50, imageInfo: { url: "https://images.unsplash.com/photo-1626074353765-517a681e40be", type: "image/jpg" } },
            { itemId: "67890123-4567-4890-0123-456789012345", foodName: "ข้าวมันไก่ทอด", price: 60, imageInfo: { url: "https://images.unsplash.com/photo-1562967914-608f82629710", type: "image/jpg" } }
        ],
        address: {
            addressId: "b2c3d4e5-f6a7-4b8c-9d0e-1f2a3b4c5d6e",
            addressName: "12 ถนนวัวลาย",
            subDistrict: "หายยา",
            district: "เมืองเชียงใหม่",
            province: "เชียงใหม่",
            postalCode: "50100"
        },
        phone: "0984567890",
        email: "kaijomphalang@example.com"
    },
    {
        restaurantId: "1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p",
        restaurantName: "ตำ ตำ ตำ",
        imageInfo: { url: "https://images.unsplash.com/photo-1547592166-23ac45744acd", type: "image/jpg" },
        status: "STORE_STATUS_CLOSED",
        menu: [
            { itemId: "9i0j1k2l-3m4n-5o6p-7q8r-9s0t1u2v3w4x", foodName: "ส้มตำไทย", price: 50, imageInfo: { url: "https://images.unsplash.com/photo-1511690656952-34342bb7c2f2", type: "image/jpg" } },
            { itemId: "0t1u2v3w-4x5y-6z7a-8b9c-0d1e2f3g4h5i", foodName: "ไก่ย่าง", price: 60, imageInfo: { url: "https://images.unsplash.com/photo-1598515214211-89d3c73ae83b", type: "image/jpg" } }
        ],
        address: {
            addressId: "c3d4e5f6-a7b8-4c9d-0e1f-2a3b4c5d6e7f",
            addressName: "78 ถนนเจริญประเทศ",
            subDistrict: "ช้างม่อย",
            district: "เมืองเชียงใหม่",
            province: "เชียงใหม่",
            postalCode: "50000"
        },
        phone: "0983456789",
        email: "tumtum@example.com"
    },
    {
        restaurantId: "1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p",
        restaurantName: "ตำ ตำ ตำ",
        imageInfo: { url: "https://images.unsplash.com/photo-1547592166-23ac45744acd", type: "image/jpg" },
        status: "STORE_STATUS_CLOSED",
        menu: [
            { itemId: "9i0j1k2l-3m4n-5o6p-7q8r-9s0t1u2v3w4x", foodName: "ส้มตำไทย", price: 50, imageInfo: { url: "https://images.unsplash.com/photo-1511690656952-34342bb7c2f2", type: "image/jpg" } },
            { itemId: "0t1u2v3w-4x5y-6z7a-8b9c-0d1e2f3g4h5i", foodName: "ไก่ย่าง", price: 60, imageInfo: { url: "https://images.unsplash.com/photo-1598515214211-89d3c73ae83b", type: "image/jpg" } }
        ],
        address: {
            addressId: "c3d4e5f6-a7b8-4c9d-0e1f-2a3b4c5d6e7f",
            addressName: "78 ถนนเจริญประเทศ",
            subDistrict: "ช้างม่อย",
            district: "เมืองเชียงใหม่",
            province: "เชียงใหม่",
            postalCode: "50000"
        },
        phone: "0983456789",
        email: "tumtum@example.com"
    },
    {
        restaurantId: "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
        restaurantName: "ป้าแดง ก๋วยเตี๋ยว",
        imageInfo: { url: "https://images.unsplash.com/photo-1582878826629-29b7ad1cdc43", type: "image/jpg" },
        status: "STORE_STATUS_OPEN",
        menu: [
            { itemId: "10928374-abcd-4321-bcde-90817263544d", foodName: "ก๋วยเตี๋ยวหมูน้ำใส", price: 40, imageInfo: { url: "https://images.unsplash.com/photo-1617093727343-374698b1b08d", type: "image/jpg" } },
            { itemId: "21039485-bcde-5432-cdef-01928374655e", foodName: "ก๋วยเตี๋ยวต้มยำ", price: 45, imageInfo: { url: "https://images.unsplash.com/photo-1552611052-33e04de081de", type: "image/jpg" } },
            { itemId: "32140596-cdef-6543-defa-12039485766f", foodName: "เส้นเล็กน้ำตก", price: 45, imageInfo: { url: "https://images.unsplash.com/photo-1569718212165-3a8278d5f624", type: "image/jpg" } }
        ],
        address: {
            addressId: "a1b2c3d4-e5f6-4a5b-bc6d-7e8f9a0b1c2d",
            addressName: "45 ถนนราชภาคินัย",
            subDistrict: "ช้างคลาน",
            district: "เมืองเชียงใหม่",
            province: "เชียงใหม่",
            postalCode: "50100"
        },
        phone: "0982345678",
        email: "padang_noodle@example.com"
    },
    {
        restaurantId: "78901234-5678-4321-8765-432109876543",
        restaurantName: "ไก่จอมพลัง",
        imageInfo: { url: "https://images.unsplash.com/photo-1626082927389-6cd097cdc6ec", type: "image/jpg" },
        status: "STORE_STATUS_OPEN",
        menu: [
            { itemId: "56789012-3456-4789-9012-345678901234", foodName: "ข้าวมันไก่ต้ม", price: 50, imageInfo: { url: "https://images.unsplash.com/photo-1626074353765-517a681e40be", type: "image/jpg" } },
            { itemId: "67890123-4567-4890-0123-456789012345", foodName: "ข้าวมันไก่ทอด", price: 60, imageInfo: { url: "https://images.unsplash.com/photo-1562967914-608f82629710", type: "image/jpg" } }
        ],
        address: {
            addressId: "b2c3d4e5-f6a7-4b8c-9d0e-1f2a3b4c5d6e",
            addressName: "12 ถนนวัวลาย",
            subDistrict: "หายยา",
            district: "เมืองเชียงใหม่",
            province: "เชียงใหม่",
            postalCode: "50100"
        },
        phone: "0984567890",
        email: "kaijomphalang@example.com"
    },
];