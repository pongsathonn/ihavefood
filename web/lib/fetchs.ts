import {
  Coupon,
  CouponSchema,
  Customer,
  CustomerSchema,
  MerchantSchema,
  type Restaurant,
} from '@/lib/types'
import { cache } from 'react'
import 'server-only'
import { authentication, getCustomerId, getSession } from './session'

export const getCustomer = cache(async function (): Promise<Customer> {
  const { isAuth } = await authentication()
  if (!isAuth) {
    throw new Error('Unauthorized')
  }

  const customerId = await getCustomerId()
  if (!customerId) {
    throw new Error('customerId is required')
  }

  const serverUrl = process.env.SERVER_URL
  if (!serverUrl) throw new Error('SERVER_URL is not defined')
  const token = await getSession()
  const url = `${serverUrl}/api/customers/${customerId}`
  const res = await fetch(url, {
    method: 'GET',
    headers: {
      Accept: 'application/json',
      Cookie: `session=${token}`,
    },
  })
  if (!res.ok) {
    const errorText = await res.text()
    throw new Error(`Failed to get customer: ${errorText}`)
  }

  const result = CustomerSchema.safeParse(await res.json())
  if (!result.success) {
    throw new Error(`Invalid customer data: ${result.error.message}`)
  }

  return result.data
})

export const getDeliveryEstimate = cache(async function (
  customerId: string,
  customerAddrId: string,
  restaurantId: string,
): Promise<{
  distance: number
  deliveryFee: number
  eta: number
}> {
  const { isAuth } = await authentication()
  if (!isAuth) {
    throw new Error('Unauthorized')
  }

  const serverUrl = process.env.SERVER_URL
  if (!serverUrl) throw new Error('SERVER_URL is not defined')
  const token = await getSession()
  const url =
    `${serverUrl}/api/deliveries/delivery-estimate` +
    `?customer_id=${customerId}` +
    `&customer_address_id=${customerAddrId}` +
    `&merchant_id=${restaurantId}`
  const res = await fetch(url, {
    method: 'GET',
    headers: {
      Accept: 'application/json',
      Cookie: `session=${token}`,
    },
  })
  if (!res.ok) {
    const errorText = await res.text()
    throw new Error(`Failed to get delivery estimate: ${errorText}`)
  }

  const { distanceKm, deliveryFee, etaMinutes } = await res.json()

  return {
    distance: distanceKm,
    deliveryFee: deliveryFee,
    eta: etaMinutes,
  }
})

export const listCoupons = cache(async function (): Promise<Coupon[]> {
  const { isAuth } = await authentication()
  if (!isAuth) {
    throw new Error('Unauthorized')
  }
  const serverUrl = process.env.SERVER_URL
  if (!serverUrl) throw new Error('SERVER_URL is not defined')
  const token = await getSession()
  const res = await fetch(`${serverUrl}/api/coupons`, {
    method: 'GET',
    headers: {
      Accept: 'application/json',
      Cookie: `session=${token}`,
    },
  })

  if (!res.ok) {
    const errorText = await res.text()
    throw new Error(`Failed to list coupons: ${errorText}`)
  }

  const jsonCoupons = (await res.json()).coupons

  const results = CouponSchema.array().safeParse(jsonCoupons)
  if (!results.success) {
    throw results.error
  }

  return results.data
})

export const getRestaurant = cache(async function (
  id: string,
): Promise<Restaurant> {
  const { isAuth } = await authentication()
  if (!isAuth) {
    throw new Error('Unauthorized')
  }
  const serverUrl = process.env.SERVER_URL
  if (!serverUrl) throw new Error('SERVER_URL is not defined')
  const token = await getSession()
  const res = await fetch(`${serverUrl}/api/merchants/${id}`, {
    method: 'GET',
    headers: {
      Accept: 'application/json',
      Cookie: `session=${token}`,
    },
  })

  if (!res.ok) {
    const errorText = await res.text()
    throw new Error(`Failed to get restaurant: ${errorText}`)
  }

  const result = MerchantSchema.safeParse(await res.json())
  if (!result.success) {
    throw result.error
  }

  return result.data
})

export const listRestaurants = cache(async function (): Promise<Restaurant[]> {
  const { isAuth } = await authentication()
  if (!isAuth) {
    throw new Error('Unauthorized')
  }
  const serverUrl = process.env.SERVER_URL
  if (!serverUrl) throw new Error('SERVER_URL is not defined')
  const token = await getSession()
  const res = await fetch(`${serverUrl}/api/merchants`, {
    method: 'GET',
    headers: {
      Accept: 'application/json',
      Cookie: `session=${token}`,
    },
  })

  if (!res.ok) {
    const errorText = await res.text()
    throw new Error(`Failed to list restaurants: ${errorText}`)
  }
  const jsonMerchants = (await res.json()).merchants
  const results = MerchantSchema.array().safeParse(jsonMerchants)
  if (!results.success) {
    throw results.error
  }

  return results.data
})

// Write operations belong to 'use server'
// export const updateAddress = cache(async function ({
//   customerId,
//   addressId,
//   update,
// }: {
//   customerId: string
//   addressId: string
//   update: Address
// }) {
//   const { isAuth } = await authentication()
//   if (!isAuth) {
//     throw new Error('Unauthorized')
//   }
//   const serverUrl = process.env.SERVER_URL
//   if (!serverUrl) throw new Error('SERVER_URL is not defined')
//   const token = getSession()
//   const res = await fetch(
//     `${serverUrl}/api/customers/${customerId}/addresses/${addressId}`,
//     {
//       method: 'PATCH',
//       headers: {
//         Accept: 'application/json',
//         Cookie: `session=${token}`,
//         'Content-Type': 'application/json',
//       },
//       body: JSON.stringify(update),
//     },
//   )

//   if (!res.ok) {
//     const errorText = await res.text()
//     throw new Error(`Failed to update customer address: ${errorText}`)
//   }

//   return res.json()
// })

// export const createPlaceOrder = cache(async function (order: PlaceOrder) {
//   const { isAuth } = await authentication()
//   if (!isAuth) {
//     throw new Error('Unauthorized')
//   }
//   const serverUrl = process.env.SERVER_URL
//   if (!serverUrl) throw new Error('SERVER_URL is not defined')
//   const token = getSession()
//   const res = await fetch(`${serverUrl}/api/orders/place_order`, {
//     method: 'POST',
//     headers: {
//       Accept: 'application/json',
//       Cookie: `session=${token}`,
//       'Content-Type': 'application/json',
//     },
//     body: JSON.stringify(order),
//   })

//   if (!res.ok) {
//     const errorText = await res.text()
//     throw new Error(`Failed to create order: ${errorText}`)
//   }
//   return res.json()
// })
