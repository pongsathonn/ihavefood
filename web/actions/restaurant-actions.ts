'use server'

import { getCustomer, getDeliveryEstimate, listRestaurants } from '@/lib/fetchs'
import { RestaurantWithEst } from '@/lib/types'

export type NewAddress = {
    addressName: string
    subDistrict: string
    district: string
    province: string
    postalCode: string
}

export const getNearbyRestaurants = async function (
    defaultAddr: string,
): Promise<RestaurantWithEst[]> {
    // Assume listRestaurants returns nearby restaurants
    // listRestaurants tends to be called with defaultAddr
    const nearRestaurants = await listRestaurants()

    const customer = await getCustomer()
    if (!customer) throw new Error('customer not found')

    // TODO: fix n+1 problem
    // Move estimation logic into the batch 'listRestaurants'
    const restaurants = nearRestaurants.map(async (r) => {
        try {
            const { distance, deliveryFee, eta } = await getDeliveryEstimate(
                customer.customerId,
                defaultAddr,
                r.restaurantId,
            )

            return { ...r, distance, deliveryFee, eta }
        } catch (error) {
            console.error(
                `Failed to get fee for restaurant ${r.restaurantId}:`,
                error,
            )
            return { ...r, distance: 0, deliveryFee: 0, eta: 0 }
        }
    })

    return await Promise.all(restaurants)
}