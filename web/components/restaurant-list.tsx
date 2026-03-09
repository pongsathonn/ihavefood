'use client'

import { getNearbyRestaurants } from '@/app/restaurants/actions'
import { RestaurantWithEst } from '@/lib/types'
import Image from 'next/image'
import { useRouter, useSearchParams } from 'next/navigation'
import { useEffect, useState } from 'react'
import { PaginationSimple } from './pagination'
import { Skeleton } from './ui/skeleton'

export default function RestaurantList({
  customerAddrId,
}: {
  customerAddrId: string
}) {
  const searchParams = useSearchParams()
  const pageParam = searchParams.get('page')
  const currentPage = pageParam ? parseInt(pageParam) : 1

  const [restaurants, setRestaurants] = useState<RestaurantWithEst[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchRestaurants = async () => {
      setIsLoading(true)
      setError(null)
      try {
        const data = await getNearbyRestaurants(customerAddrId)
        setRestaurants(data)
      } catch (err) {
        console.error(err)
        setError('Failed to load nearby restaurants.')
      } finally {
        setIsLoading(false)
      }
    }
    if (customerAddrId) fetchRestaurants()
  }, [customerAddrId])

  if (isLoading) {
    return (
      <div className="grid grid-cols-1 sm:grid-cols-3 lg:grid-cols-4 gap-6">
        {[...Array(8)].map((_, i) => (
          <RestaurantCardSkeleton key={i} />
        ))}
      </div>
    )
  }

  if (error) {
    return <div className="text-red-500 text-center p-10">{error}</div>
  }

  if (restaurants.length === 0) {
    return (
      <p className="text-gray-500 italic text-center p-10">
        No restaurants available in your area.
      </p>
    )
  }

  const pageSize = 8
  const start = (currentPage - 1) * pageSize
  const displayedRestaurants = restaurants.slice(start, start + pageSize)

  return (
    <div>
      <div className="grid grid-cols-1 sm:grid-cols-3 lg:grid-cols-4 gap-6 items-stretch">
        {isLoading
          ? Array.from({ length: pageSize }).map((_, i) => (
              <RestaurantCardSkeleton key={i} />
            ))
          : displayedRestaurants.map((restaurant) => (
              <RestaurantCard
                key={restaurant.restaurantId}
                restaurant={restaurant}
              />
            ))}
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

  const handleRestaurantSelect = (restaurantId: string) => {
    // TODO: Pushing with id+name instead, for better SEO
    router.push(`/restaurants/${restaurantId}`)
  }

  return (
    <div
      key={restaurant.restaurantId}
      onClick={() =>
        !isClosed && handleRestaurantSelect(restaurant.restaurantId)
      }
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

function RestaurantCardSkeleton() {
  return (
    <div className="flex flex-col h-full bg-white rounded-2xl overflow-hidden border border-gray-200">
      <Skeleton className="aspect-video w-full rounded-none" />
      <div className="p-4 space-y-3 grow">
        <Skeleton className="h-5 w-3/4" />
        <Skeleton className="h-4 w-full mt-2" />
        <Skeleton className="h-4 w-2/4 mt-2" />
      </div>
    </div>
  )
}
