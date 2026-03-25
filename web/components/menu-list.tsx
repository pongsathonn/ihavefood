'use client'

import { Coupon, MenuItem, Restaurant } from '@/lib/types'
import { ChevronLeft } from 'lucide-react'
import Image from 'next/image'
import { useRouter } from 'next/navigation'
import { useState } from 'react'
import Cart from './cart'

export default function MenuList({
  coupons,
  restaurant,
  deliveryFee,
}: {
  coupons: Coupon[]
  restaurant: Restaurant
  deliveryFee: number
}) {
  const router = useRouter()

  if (!restaurant) {
    throw new Error('Restaurant not found')
  }

  // quantity tells how many item were added.
  const [cartItems, setCartItems] = useState<
    (MenuItem & { quantity: number })[]
  >([])

  const handleBackButton = () => {
    router.back()
  }

  const addMenuItem = (item: MenuItem) => {
    setCartItems((prev) => {
      const existItem = prev.find((v) => v.itemId === item.itemId)
      if (existItem) {
        return prev.map((v) =>
          v.itemId === item.itemId ? { ...v, quantity: v.quantity + 1 } : v,
        )
      }
      return [...prev, { ...item, quantity: 1 }]
    })
  }

  const removeMenuItem = (itemId: string) =>
    setCartItems((prev) => prev.filter((item) => item.itemId != itemId))

  return (
    <section className="bg-gray-50 p-6 rounded-3xl shadow-lg border border-gray-100">
      <div className="flex items-center mb-4">
        <button onClick={handleBackButton} className="cursor-pointer">
          <ChevronLeft />
        </button>
        <h2 className="text-2xl font-bold text-gray-800 ml-6">
          {restaurant != undefined && restaurant.restaurantName}
        </h2>
      </div>

      <div className="flex flex-col md:flex-row gap-6 ">
        <div className="flex-1">
          <MenuCard menu={restaurant.menu} onAddMenuItem={addMenuItem} />
        </div>
        <div className="w-full md:w-80 lg:w-96 shrink-0">
          <Cart
            restaurantId={restaurant.restaurantId}
            deliveryFee={deliveryFee}
            coupons={coupons}
            cartItems={cartItems}
            onRemoveMenuItem={removeMenuItem}
          />
        </div>
      </div>
    </section>
  )
}

function MenuCard({
  menu,
  onAddMenuItem,
}: {
  menu: MenuItem[]
  onAddMenuItem: (item: MenuItem) => void
}) {
  const MenuCard = ({ item }: { item: MenuItem }) => {
    return (
      <div className="flex items-center justify-between p-4 bg-white rounded-2xl border border-gray-100 shadow-sm">
        <div className="flex items-center space-x-4">
          <div className="w-16 h-16 rounded-lg overflow-hidden shrink-0">
            {item.imageInfo?.url ? (
              <Image
                src={item.imageInfo.url}
                alt={item.foodName}
                width={64}
                height={64}
                className="w-full h-full object-cover"
              />
            ) : (
              <div className="w-full h-full bg-gray-200 animate-pulse" />
            )}
          </div>

          <div>
            <h4 className="text-md font-semibold text-gray-800">
              {item.foodName}
            </h4>
            <span className="text-sm text-gray-500">
              ฿{item.price.toFixed(2)}
            </span>
          </div>
        </div>
        <button
          className="add-to-cart-btn bg-gray-900 text-white px-4 py-2 rounded-xl text-sm font-bold hover:bg-black transition cursor-pointer"
          data-item={JSON.stringify(item)}
          onClick={() => onAddMenuItem(item)}
        >
          Add
        </button>
      </div>
    )
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-1 gap-4 p-4 rounded-2xl ">
      {menu.length > 0 &&
        (() => {
          if (!menu || menu.length === 0) {
            return (
              <p className="text-gray-500 italic">
                This restaurant has no menu items.
              </p>
            )
          }
          return menu.map((item, key) => <MenuCard key={key} item={item} />)
        })()}
    </div>
  )
}
