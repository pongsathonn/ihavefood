import { create } from 'zustand'
import { Restaurant } from '@/lib/types'

type RestaurantWithFee = Restaurant & { deliveryFee: number }

type RestaurantState = {
  restaurants: RestaurantWithFee[]
  setRestaurants: (list: RestaurantWithFee[]) => void
  getRestaurantById: (id: string) => RestaurantWithFee | undefined
}

export const useRestaurantWithFee = create<RestaurantState>((set, get) => ({
  restaurants: [],
  setRestaurants: (list) => set({ restaurants: list }),
  getRestaurantById: (id) =>
    get().restaurants.find((r) => r.restaurantId === id),
}))
