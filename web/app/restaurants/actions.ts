'use server'

import { getCustomer, getDeliveryEstimate, listRestaurants } from '@/lib/fetchs'
import { authentication, getCustomerId, getSession } from '@/lib/session'
import { Address, AddressSchema, RestaurantWithEst } from '@/lib/types'
import { cookies } from 'next/headers'

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

export const createCustomerAddress = async function (
  newAddress: NewAddress,
): Promise<Address> {
  const { isAuth } = await authentication()
  if (!isAuth) {
    throw new Error('Unauthorized')
  }

  const customerId = await getCustomerId()
  const serverUrl = process.env.SERVER_URL
  if (!serverUrl) throw new Error('SERVER_URL is not defined')
  const token = await getSession()
  const res = await fetch(`${serverUrl}/api/customers/${customerId}/address`, {
    method: 'POST',
    headers: {
      Accept: 'application/json',
      Cookie: `session=${token}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      address: {
        address_name: newAddress.addressName,
        sub_district: newAddress.subDistrict,
        district: newAddress.district,
        province: newAddress.province,
        postal_code: newAddress.postalCode,
      },
    }),
  })
  if (!res.ok) {
    const errorText = await res.text()
    throw new Error(`Failed to create customer address: ${errorText}`)
  }
  const createdAddr = AddressSchema.safeParse(await res.json()).data
  if (!createdAddr) {
    throw new Error('Failed to parse Custoemr address')
  }

  ;(await cookies()).set('default_address_id', createdAddr.addressId)

  return createdAddr
}
