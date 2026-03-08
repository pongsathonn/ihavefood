'use client'

import { RestaurantWithEst, type Restaurant } from '@/lib/types'
import Image from 'next/image'
import { useRouter, useSearchParams } from 'next/navigation'
import { PaginationSimple } from './pagination'

export default function RestaurantList({
  restaurants,
}: {
  restaurants?: RestaurantWithEst[]
}) {
  const searchParams = useSearchParams()
  const pageParam = searchParams.get('page')
  const currentPage = pageParam ? parseInt(pageParam) : 1

  if (restaurants == undefined || restaurants.length == 0) {
    return <p className="text-gray-200 italic">Restaurants are undefined</p>
  }

  return (
    <div>
      <div className="grid grid-cols-1 sm:grid-cols-3 lg:grid-cols-4 gap-6 items-stretch">
        {!restaurants || restaurants.length === 0 ? (
          <p className="text-gray-500 italic">No restaurants available.</p>
        ) : (
          (() => {
            const pageSize = 8
            const start = (currentPage - 1) * pageSize
            const end = start + pageSize
            return restaurants.slice(start, end).map((restaurant) => {
              return (
                <RestaurantCard
                  key={restaurant.restaurantId}
                  restaurant={restaurant}
                />
              )
            })
          })()
        )}
      </div>

      <div className="pt-8">
        <PaginationSimple currentPage={currentPage} />
      </div>
    </div>
  )
}

function RestaurantCard({ restaurant }: { restaurant: RestaurantWithEst }) {
  const router = useRouter()

  const isClosed = restaurant.status !== 'STORE_STATUS_OPEN'

  const handleRestaurantSelect = (restaurant: Restaurant) => {
    // TODO: Pushing with id+name instead, for better SEO
    router.push(`/restaurants/${restaurant.restaurantId}`)
  }

  return (
    <div
      key={restaurant.restaurantId}
      onClick={() => !isClosed && handleRestaurantSelect(restaurant)}
      className={` flex flex-col h-full bg-gray-50 rounded-2xl shadow-sm overflow-hidden cursor-pointer hover:shadow-lg transition-all ${isClosed ? 'opacity-50 pointer-events-none' : ''} `}
    >
      <div className="relative aspect-video w-full shrink-0">
        <Image
          src={restaurant.imageInfo.url}
          alt={restaurant.restaurantName}
          fill
          className="object-cover"
        />
      </div>

      <div className="p-4 flex flex-col grow">
        <h3 className="text-lg font-semibold text-gray-800 line-clamp-1 mb-1">
          {restaurant.restaurantName}
        </h3>

        <div className="flex" style={{ display: isClosed ? 'none' : 'block' }}>
          <p className="text-sm text-gray-500 flex items-center gap-2 whitespace-nowrap">
            <span>
              Delivery: <span className="text-red-700">฿20.00</span>
            </span>
            <span className="text-gray-300">|</span>
            <span className="truncate">
              {restaurant.distance} km ({restaurant.eta} min)
            </span>
          </p>
        </div>

        <p className="text-sm text-gray-500 pt-2">
          Status: {isClosed ? 'CLOSED' : 'OPEN'}
        </p>
      </div>
    </div>
  )
}
