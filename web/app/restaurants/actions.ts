'use server'

import { getCustomer, getDeliveryFee, listRestaurants } from '@/lib/fetchs'
import { authentication, getCustomerId, getSession } from '@/lib/session'
import {
  Address,
  AddressSchema,
  Restaurant,
  RestaurantWithEst,
} from '@/lib/types'
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
  const restaurants = nearRestaurants.map(async (r) => {
    try {
      const fee = await getDeliveryFee(
        customer.customerId,
        defaultAddr,
        r.restaurantId,
      )
      const result = await attachDeliveryEstimate(r, fee)
      return result
    } catch (error) {
      console.error(
        `Failed to get fee for restaurant ${r.restaurantId}:`,
        error,
      )
      return attachDeliveryEstimate(r, 0)
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

export const attachDeliveryEstimate = async function (
  restaurant: Restaurant,
  deliveryFee: number,
): Promise<RestaurantWithEst> {
  const baseDistance = (deliveryFee - 10) / 1.6
  const randomVariation = Math.random() * 2 - 1

  const distance = Number(
    Math.max(0, Math.min(25, baseDistance + randomVariation)).toFixed(2),
  )

  const baseEta = Math.round((distance / 30) * 60)
  const etaVariation = Math.floor(Math.random() * 10) - 5
  const eta = Math.max(1, Math.min(60, baseEta + etaVariation))

  return {
    ...restaurant,
    distance,
    deliveryFee,
    eta,
  }
}
