import { RestaurantWithEst } from '@/lib/types'
import { create } from 'zustand'

type RestaurantState = {
  restaurants: RestaurantWithEst[]
  setRestaurants: (list: RestaurantWithEst[]) => void
  getRestaurantById: (id: string) => RestaurantWithEst | undefined
}


export const useRestaurantWithEst = create<RestaurantState>((set, get) => ({
  restaurants: [],
  setRestaurants: (list) => set({ restaurants: list }),
  getRestaurantById: (id) =>
    get().restaurants.find((r) => r.restaurantId === id),
}))
